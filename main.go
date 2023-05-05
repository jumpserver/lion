package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	ginCookie "github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"lion/pkg/common"
	"lion/pkg/config"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
	"lion/pkg/jms-sdk-go/service/videoworker"
	"lion/pkg/logger"
	"lion/pkg/middleware"
	"lion/pkg/session"
	"lion/pkg/storage"
	"lion/pkg/tunnel"
)

var (
	Version    = "Unknown"
	Buildstamp = ""
	Githash    = ""
	Goversion  = ""
)
var (
	infoFlag   = false
	configPath = ""
)

func init() {
	// default config.yml
	flag.StringVar(&configPath, "f", "config.yml", "config.yml path")
	flag.BoolVar(&infoFlag, "V", false, "version info")
}

func main() {
	flag.Parse()
	if infoFlag {
		fmt.Printf("Version:             %s\n", Version)
		fmt.Printf("Git Commit Hash:     %s\n", Githash)
		fmt.Printf("UTC Build Time :     %s\n", Buildstamp)
		fmt.Printf("Go Version:          %s\n", Goversion)
		return
	}
	config.Setup(configPath)
	logger.SetupLogger(config.GlobalConfig)
	jmsService := MustJMService()
	videoWorkerClient := NewWorkerClient(*config.GlobalConfig)
	bootstrap(jmsService)
	tunnelService := tunnel.GuacamoleTunnelServer{
		Cache: &tunnel.GuaTunnelCacheManager{
			GuaTunnelCache: NewGuaTunnelCache(),
		},
		SessCache: &tunnel.SessionCache{
			Sessions: make(map[string]*session.TunnelSession),
		},
		JmsService: jmsService,
		SessionService: &session.Server{JmsService: jmsService,
			VideoWorkerClient: videoWorkerClient},
	}
	eng := registerRouter(jmsService, &tunnelService)
	go runHeartTask(jmsService, tunnelService.Cache)
	go runCleanDriverDisk(tunnelService.Cache)
	addr := net.JoinHostPort(config.GlobalConfig.BindHost, config.GlobalConfig.HTTPPort)
	fmt.Printf("Lion Version %s, more see https://www.jumpserver.org\n", Version)
	logger.Infof("listen on: %s", addr)
	logger.Fatal(http.ListenAndServe(addr, eng))
}

func NewGuaTunnelCache() tunnel.GuaTunnelCache {
	cfg := config.GlobalConfig
	switch strings.ToLower(config.GlobalConfig.ShareRoomType) {
	case config.ShareTypeRedis:
		existFile := func(path string) string {
			if info, err2 := os.Stat(path); err2 == nil && !info.IsDir() {
				return path
			}
			return ""
		}
		sslCaPath := filepath.Join(cfg.CertsFolderPath, "redis_ca.crt")
		sslCertPath := filepath.Join(cfg.CertsFolderPath, "redis_client.crt")
		sslKeyPath := filepath.Join(cfg.CertsFolderPath, "redis_client.key")
		return tunnel.NewGuaTunnelRedisCache(tunnel.Config{
			Addr: net.JoinHostPort(cfg.RedisHost,
				strconv.Itoa(cfg.RedisPort)),
			Password: cfg.RedisPassword,
			DBIndex:  cfg.RedisDBIndex,

			SentinelsHost:    cfg.RedisSentinelHosts,
			SentinelPassword: cfg.RedisSentinelPassword,

			UseSSL:  cfg.RedisUseSSL,
			SSLCa:   existFile(sslCaPath),
			SSLCert: existFile(sslCertPath),
			SSLKey:  existFile(sslKeyPath),
		})
	default:
		return tunnel.NewLocalTunnelLocalCache()
	}
}

func runHeartTask(jmsService *service.JMService, tunnelCache *tunnel.GuaTunnelCacheManager) {
	// default 30s
	ws, err := jmsService.GetWsClient()
	if err != nil {
		logger.Errorf("Start ws heart beat failed: %s", err)
		time.Sleep(10 * time.Second)
		go runHeartTask(jmsService, tunnelCache)
		return
	}
	logger.Info("Start ws heart beat success")
	done := make(chan struct{}, 2)
	go func() {
		defer close(done)
		for {
			msgType, message, err2 := ws.ReadMessage()
			if err2 != nil {
				logger.Errorf("Ws client read err: %s", err2)
				return
			}
			switch msgType {
			case websocket.PingMessage,
				websocket.PongMessage:
				logger.Debug("Ws client ping/pong Message")
				continue
			case websocket.CloseMessage:
				logger.Debug("Ws client close Message")
				return
			}
			var tasks []model.TerminalTask
			if err = json.Unmarshal(message, &tasks); err != nil {
				logger.Errorf("Ws client Unmarshal failed: %s", err)
				continue
			}
			for i := range tasks {
				task := tasks[i]
				switch task.Name {
				case model.TaskKillSession:
					if connection := tunnelCache.GetBySessionId(task.Args); connection != nil {
						connection.Terminate(task.Kwargs.TerminatedBy)
						if err = jmsService.FinishTask(task.ID); err != nil {
							logger.Error(err)
						}
					}
				default:
				}
			}
		}
	}()

	beatTicker := time.NewTicker(time.Second * 30)
	defer beatTicker.Stop()
	if err1 := ws.WriteJSON(GetStatusData(tunnelCache)); err1 != nil {
		logger.Errorf("Ws heart beat data failed: %s", err1)
	}
	for {
		select {
		case <-done:
			logger.Error("Ws heart beat closed, try reconnect after 10s")
			time.Sleep(10 * time.Second)
			go runHeartTask(jmsService, tunnelCache)
			return
		case <-beatTicker.C:
			if err1 := ws.WriteJSON(GetStatusData(tunnelCache)); err1 != nil {
				logger.Errorf("Ws client write stat data failed: %s", err1)
				continue
			}
		}
	}
}

func runCleanDriverDisk(tunnelCache *tunnel.GuaTunnelCacheManager) {
	scheduleTime := config.GlobalConfig.CleanDriveScheduleTime
	if scheduleTime < 1 {
		logger.Info("Clean driver folder schedule task stop")
		return
	}
	logger.Info("Clean driver folder schedule task start")
	cleanDriveTicker := time.NewTicker(time.Duration(scheduleTime) * time.Hour)
	defer cleanDriveTicker.Stop()
	drivePath := config.GlobalConfig.DrivePath
	for range cleanDriveTicker.C {
		folders, err := os.ReadDir(drivePath)
		if err != nil {
			logger.Error(err)
			continue
		}
		currentOnlineUserIds := tunnelCache.RangeActiveUserIds()
		for i := range folders {
			if _, ok := currentOnlineUserIds[folders[i].Name()]; ok {
				continue
			}
			logger.Debugf("Remove drive folder %s", folders[i].Name())
			if err = os.RemoveAll(filepath.Join(drivePath, folders[i].Name())); err != nil {
				logger.Errorf("Remove drive folder %s err: %s", folders[i].Name(), err)
			}
		}
	}
}

func registerRouter(jmsService *service.JMService, tunnelService *tunnel.GuacamoleTunnelServer) *gin.Engine {
	if config.GlobalConfig.LogLevel != "DEBUG" {
		gin.SetMode(gin.ReleaseMode)
	}
	eng := gin.New()
	eng.Use(gin.Recovery())
	eng.Use(gin.Logger())

	lionGroup := eng.Group("/lion")
	cookieStore := ginCookie.NewStore([]byte(common.RandomStr(32)))
	lionGroup.Use(middleware.GinSessionAuth(cookieStore))
	now := time.Now()
	// vue的设置
	{
		lionGroup.Static("/assets", "./ui/dist/assets")
		lionGroup.StaticFile("/favicon.ico", "./ui/dist/favicon.ico")

		lionGroup.GET("/health/", func(ctx *gin.Context) {
			status := make(map[string]interface{})
			status["timestamp"] = time.Now().UTC()
			status["uptime"] = time.Now().Sub(now).Minutes()
			ctx.JSON(http.StatusOK, status)
		})
	}
	{
		connectGroup := lionGroup.Group("/connect")
		connectGroup.Use(middleware.JmsCookieAuth(jmsService))
		connectGroup.GET("/", func(ctx *gin.Context) {
			ctx.File("./ui/dist/index.html")
		})
	}
	{
		monitorGroup := lionGroup.Group("/monitor")
		monitorGroup.Use(middleware.JmsCookieAuth(jmsService))
		monitorGroup.Any("", func(ctx *gin.Context) {
			ctx.File("./ui/dist/index.html")
		})
	}
	// token 使用 lion 自带认证

	{
		tokenGroup := lionGroup.Group("/token")
		tokenGroup.Use(middleware.SessionAuth(jmsService))
		tokenGroup.POST("/session", tunnelService.TokenSession)
		tokenGroup.DELETE("/sessions/:sid/", tunnelService.DeleteSession)
		tokenTunnels := tokenGroup.Group("/tunnels")
		tokenTunnels.GET("/:tid/streams/:index/:filename", tunnelService.DownloadFile)
		tokenTunnels.POST("/:tid/streams/:index/:filename", tunnelService.UploadFile)
	}

	// ws的设置
	wsGroup := lionGroup.Group("/ws")
	{
		wsGroup.Group("/connect").Use(
			middleware.JmsCookieAuth(jmsService)).GET("/", tunnelService.Connect)
		wsGroup.Group("/monitor").Use(
			middleware.JmsCookieAuth(jmsService)).GET("/", tunnelService.Monitor)

		wsGroup.Group("/token").Use(
			middleware.SessionAuth(jmsService)).GET("/", tunnelService.Connect)
	}

	{
		apiGroup := lionGroup.Group("/api")
		apiGroup.Use(middleware.JmsCookieAuth(jmsService))
		apiGroup.POST("/session", tunnelService.CreateSession)
		apiGroup.DELETE("/sessions/:sid/", tunnelService.DeleteSession)
		apiGroup.GET("/tunnels/:tid/streams/:index/:filename", tunnelService.DownloadFile)
		apiGroup.POST("/tunnels/:tid/streams/:index/:filename", tunnelService.UploadFile)
	}

	pprofRouter := eng.Group("/debug/pprof")
	{
		pprofRouter.Use(middleware.HTTPMiddleDebugAuth())
		pprofRouter.GET("/", gin.WrapF(pprof.Index))
		pprofRouter.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
		pprofRouter.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofRouter.GET("/profile", gin.WrapF(pprof.Profile))
		pprofRouter.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofRouter.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofRouter.GET("/trace", gin.WrapF(pprof.Trace))
		pprofRouter.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
		pprofRouter.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
		pprofRouter.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
		pprofRouter.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
		pprofRouter.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
	}
	return eng
}

func bootstrap(jmsService *service.JMService) {
	replayDir := config.GlobalConfig.RecordPath
	allRemainFiles := scanRemainReplay(jmsService, replayDir)
	go uploadRemainReplay(jmsService, allRemainFiles)
}

func uploadRemainReplay(jmsService *service.JMService, remainFiles map[string]string) {
	var replayStorage storage.ReplayStorage
	terminalConf, _ := jmsService.GetTerminalConfig()
	replayStorage = storage.NewReplayStorage(terminalConf.ReplayStorage)
	for sid, path := range remainFiles {
		absGzPath := path
		replayDateDirName := filepath.Base(filepath.Dir(path))
		if !session.ValidReplayDirname(replayDateDirName) {
			logger.Error("Invalid replay folder ", replayDateDirName)
			continue
		}
		if !strings.HasSuffix(path, session.ReplayFileNameSuffix) {
			absGzPath = path + session.ReplayFileNameSuffix
			if err := common.CompressToGzipFile(path, absGzPath); err != nil {
				logger.Errorf("Compress to gzip file %s err: %s", path, err)
				continue
			}
			_ = os.Remove(path)
		}
		var err error
		replayCfg := terminalConf.ReplayStorage
		storageType := replayCfg.TypeName
		if storageType == "null" {
			storageType = "server"
		}
		logger.Infof("Upload record file: %s, type: %s", absGzPath, storageType)
		if replayStorage != nil {
			targetName := strings.Join([]string{replayDateDirName,
				sid + session.ReplayFileNameSuffix}, "/")
			err = replayStorage.Upload(absGzPath, targetName)
		} else {
			err = jmsService.Upload(sid, absGzPath)
		}
		if err != nil {
			logger.Errorf("Upload replay failed: %s", err)
			continue
		}
		logger.Infof("Upload remain session replay %s success", absGzPath)
		// 上传成功删除文件
		_ = os.Remove(absGzPath)
		if err = jmsService.FinishReply(sid); err != nil {
			logger.Errorf("Finish reply to core api failed: %s", err)
		}
	}
}

func scanRemainReplay(jmsService *service.JMService, replayDir string) map[string]string {
	allRemainFiles := make(map[string]string)
	_ = filepath.Walk(replayDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		sidFilename := info.Name()
		var sid string
		sidFilename = strings.TrimSuffix(sidFilename, session.ReplayFileNameSuffix)
		if common.ValidUUIDString(sidFilename) {
			sid = sidFilename
		}
		if sid != "" {
			allRemainFiles[sid] = path
			if err = jmsService.SessionFinished(sid, common.NewUTCTime(info.ModTime())); err != nil {
				logger.Errorf("Session[%s] finished err: %s", sid, err)
			}
		}
		return nil
	})
	return allRemainFiles
}

func GetStatusData(tunnelCache *tunnel.GuaTunnelCacheManager) interface{} {
	sids := tunnelCache.RangeActiveSessionIds()
	payload := model.HeartbeatData{
		SessionOnlineIds: sids,
		CpuUsed:          common.CpuLoad1Usage(),
		MemoryUsed:       common.MemoryUsagePercent(),
		DiskUsed:         common.DiskUsagePercent(),
		SessionOnline:    len(sids),
	}
	return map[string]interface{}{
		"type":    "status",
		"payload": payload,
	}
}

func MustJMService() *service.JMService {
	key := MustLoadValidAccessKey()
	jmsService, err := service.NewAuthJMService(service.JMSCoreHost(
		config.GlobalConfig.CoreHost), service.JMSTimeOut(30*time.Second),
		service.JMSAccessKey(key.ID, key.Secret),
	)
	if err != nil {
		logger.Fatal("创建JMS Service 失败 " + err.Error())
		os.Exit(1)
	}
	return jmsService
}

func MustLoadValidAccessKey() model.AccessKey {
	conf := config.GlobalConfig
	var key model.AccessKey
	if err := key.LoadFromFile(conf.AccessKeyFilePath); err != nil {
		return MustRegisterTerminalAccount()
	}
	// 校验accessKey
	return MustValidKey(key)
}

func MustRegisterTerminalAccount() (key model.AccessKey) {
	conf := config.GlobalConfig
	for i := 0; i < 10; i++ {
		terminal, err := service.RegisterTerminalAccount(conf.CoreHost,
			conf.Name, conf.BootstrapToken)
		if err != nil {
			logger.Error(err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		key.ID = terminal.ServiceAccount.AccessKey.ID
		key.Secret = terminal.ServiceAccount.AccessKey.Secret
		if err := key.SaveToFile(conf.AccessKeyFilePath); err != nil {
			logger.Error("保存key失败: " + err.Error())
		}
		return key
	}
	logger.Error("注册终端失败退出")
	os.Exit(1)
	return
}

func MustValidKey(key model.AccessKey) model.AccessKey {
	conf := config.GlobalConfig
	for i := 0; i < 10; i++ {
		if err := service.ValidAccessKey(conf.CoreHost, key); err != nil {
			switch {
			case errors.Is(err, service.ErrUnauthorized):
				return MustRegisterTerminalAccount()
			default:
				logger.Error("校验 access key failed: " + err.Error())
			}
			time.Sleep(5 * time.Second)
			continue
		}
		return key
	}
	logger.Error("校验 access key failed退出")
	os.Exit(1)
	return key
}

func NewWorkerClient(cfg config.Config) *videoworker.Client {
	if !cfg.EnableVideoWorker {
		return nil
	}
	workerURL := cfg.VideoWorkerHost
	var key model.AccessKey
	if err := key.LoadFromFile(cfg.AccessKeyFilePath); err != nil {
		logger.Errorf("Create video worker client failed: loading access key err %s", err)
		return nil
	}
	workClient := videoworker.NewClient(workerURL, key, cfg.IgnoreVerifyCerts)
	if workClient == nil {
		logger.Errorf("Create video worker client failed: worker url %s", workerURL)
		return nil
	}
	go KeepWsConnect(workClient)
	return workClient
}

func KeepWsConnect(s *videoworker.Client) {
	if err := s.Login(); err != nil {
		logger.Errorf("Worker Ws client login failed: %s, try next 10s", err)
		time.Sleep(10 * time.Second)
		go KeepWsConnect(s)
		return
	}
	wsCon, err := s.GetWsClient()
	if err != nil {
		time.Sleep(10 * time.Second)
		go KeepWsConnect(s)
		return
	}
	defer wsCon.Close()
	logger.Info("Start worker ws client beat success")
	done := make(chan struct{}, 2)
	go func() {
		defer close(done)
		for {
			msgType, message, err2 := wsCon.ReadMessage()
			if err2 != nil {
				logger.Errorf("Worker Ws client read err: %s", err2)
				return
			}
			switch msgType {
			case websocket.PingMessage,
				websocket.PongMessage:
				logger.Debug("Worker Ws client ping/pong Message")
				continue
			case websocket.CloseMessage:
				logger.Debug("Worker Ws client close Message")
				return
			}
			logger.Debugf("Worker Ws client read message: %s", message)
		}
	}()
	beatTicker := time.NewTicker(time.Second * 30)
	defer beatTicker.Stop()
	pingEvent := map[string]string{
		"event": "ping",
	}
	for {
		select {
		case <-done:
			logger.Error("Worker Ws heart beat closed, try reconnect after 10s")
			time.Sleep(10 * time.Second)
			go KeepWsConnect(s)
			return
		case <-beatTicker.C:
			if err1 := wsCon.WriteJSON(pingEvent); err1 != nil {
				logger.Errorf("Ws client write stat data failed: %s", err1)
				continue
			}
		}
	}
}
