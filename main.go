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

	"lion/pkg/config"
	"lion/pkg/logger"
	"lion/pkg/middleware"
	"lion/pkg/proxy"
	"lion/pkg/session"
	"lion/pkg/storage"
	"lion/pkg/tunnel"

	"github.com/jumpserver-dev/sdk-go/common"
	"github.com/jumpserver-dev/sdk-go/model"
	"github.com/jumpserver-dev/sdk-go/service"
	"github.com/jumpserver-dev/sdk-go/service/panda"
	"github.com/jumpserver-dev/sdk-go/service/videoworker"
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
	pandaClient := NewPandaClient(*config.GlobalConfig)
	bootstrap(jmsService)
	tunnelService := tunnel.GuacamoleTunnelServer{
		Cache: &tunnel.GuaTunnelCacheManager{
			GuaTunnelCache: NewGuaTunnelCache(),
		},
		JmsService: jmsService,
		SessionService: &session.Server{JmsService: jmsService,
			VideoWorkerClient: videoWorkerClient,
			PandaClient:       pandaClient},
	}
	eng := registerRouter(jmsService, &tunnelService)
	go runHeartTask(jmsService, tunnelService.Cache)
	go runCleanDriverDisk(tunnelService.Cache)
	go runTokenCheck(jmsService, tunnelService.Cache)
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
				if connection := tunnelCache.GetBySessionId(task.Args); connection != nil {
					if err1 := connection.HandleTask(&task); err1 != nil {
						logger.Errorf("Ws client HandleTask failed: %s", err1)
						continue
					}
					if err = jmsService.FinishTask(task.ID); err != nil {
						logger.Errorf("Ws client FinishTask failed: %s", err)
					}
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

func runTokenCheck(jmsService *service.JMService, tunnelCache *tunnel.GuaTunnelCacheManager) {
	for {
		time.Sleep(5 * time.Minute)
		connections := tunnelCache.GetActiveConnections()
		tokens := make(map[string]model.TokenCheckStatus, len(connections))
		for _, s := range connections {
			tokenId := s.Sess.AuthInfo.Id
			ret, ok := tokens[tokenId]
			if ok {
				handleTokenCheck(s, &ret)
				continue
			}
			ret, err := jmsService.CheckTokenStatus(tokenId)
			if err != nil && ret.Code == "" {
				logger.Errorf("Check token status failed: %s", err)
				continue
			}
			tokens[tokenId] = ret
			handleTokenCheck(s, &ret)
		}
	}
}

func handleTokenCheck(session *tunnel.Connection, tokenStatus *model.TokenCheckStatus) {
	var task model.TerminalTask
	switch tokenStatus.Code {
	case model.CodePermOk:
		task = model.TerminalTask{
			Name: model.TaskPermValid,
			Args: tokenStatus.Detail,
		}
	default:
		task = model.TerminalTask{
			Name: model.TaskPermExpired,
			Args: tokenStatus.Detail,
		}
	}
	if err := session.HandleTask(&task); err != nil {
		logger.Errorf("Handle token check task failed: %s", err)
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
			status["uptime"] = time.Since(now).Minutes()
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

	{
		shareGroup := lionGroup.Group("/share")
		shareGroup.Use(middleware.JmsCookieAuth(jmsService))
		shareGroup.Any("/:id", func(ctx *gin.Context) {
			ctx.File("./ui/dist/index.html")
		})
	}

	// token 使用 lion 自带认证

	{
		tokenGroup := lionGroup.Group("/token")
		tokenGroup.Use(middleware.SessionAuth(jmsService))
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

		wsGroup.Group("/share").Use(
			middleware.JmsCookieAuth(jmsService)).GET("/", tunnelService.Share)

		wsGroup.Group("/token").Use(
			middleware.SessionAuth(jmsService)).GET("/", tunnelService.Connect)

	}

	{
		apiGroup := lionGroup.Group("/api")
		apiGroup.Use(middleware.JmsCookieAuth(jmsService))
		apiGroup.GET("/tunnels/:tid/streams/:index/:filename", tunnelService.DownloadFile)
		apiGroup.POST("/tunnels/:tid/streams/:index/:filename", tunnelService.UploadFile)
		apiGroup.POST("/share/", tunnelService.CreateShare)
		apiGroup.POST("/share/remove/", tunnelService.DeleteShare)
		apiGroup.POST("/share/:id/", tunnelService.GetShare)
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
	updateEncryptConfigValue(jmsService)
	replayDir := config.GlobalConfig.RecordPath
	ftpFilePath := config.GlobalConfig.FTPFilePath
	sessionDir := config.GlobalConfig.SessionFolderPath
	allRemainFiles := scanRemainReplay(jmsService, replayDir)
	go uploadRemainReplay(jmsService, allRemainFiles)
	go uploadRemainFTPFile(jmsService, ftpFilePath)
	go uploadRemainSessionPartReplay(jmsService, sessionDir)
}

func updateEncryptConfigValue(jmsService *service.JMService) {
	cfg := config.GlobalConfig
	encryptKey := cfg.SecretEncryptKey
	if encryptKey != "" {
		redisPassword := cfg.RedisPassword
		ret, err := jmsService.GetEncryptedConfigValue(encryptKey, redisPassword)
		if err != nil {
			logger.Error("Get encrypted config value failed: " + err.Error())
			return
		}
		if ret.Value != "" {
			cfg.UpdateRedisPassword(ret.Value)
		} else {
			logger.Error("Get encrypted config value failed: empty value")
		}
	}
}
func uploadRemainFTPFile(jmsService *service.JMService, fileStoreDir string) {
	err := config.EnsureDirExist(fileStoreDir)
	if err != nil {
		logger.Debugf("upload failed FTP file err: %s", err.Error())
		return
	}

	terminalConf, _ := jmsService.GetTerminalConfig()
	ftpFileStorage := storage.NewFTPFileStorage(jmsService, terminalConf.ReplayStorage)

	allRemainFiles := make(map[string]string)
	_ = filepath.Walk(fileStoreDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		var fid string
		filename := info.Name()
		if len(filename) == 36 {
			fid = filename
		}
		if fid != "" {
			allRemainFiles[path] = fid
		}
		return nil
	})

	if len(allRemainFiles) == 0 {
		logger.Info("No remain ftp files")
		return
	}
	logger.Infof("Start upload remain %d ftp files 10 min later", len(allRemainFiles))
	time.Sleep(10 * time.Minute)

	for path, fid := range allRemainFiles {
		dateTarget, _ := filepath.Rel(fileStoreDir, path)
		target := strings.Join([]string{proxy.FTPTargetPrefix, dateTarget}, "/")
		logger.Infof("Upload FTP file: %s, type: %s", path, ftpFileStorage.TypeName())
		if err = ftpFileStorage.Upload(path, target); err != nil {
			logger.Errorf("Upload remain FTP file %s failed: %s", path, err)
			continue
		}
		if err := jmsService.FinishFTPFile(fid); err != nil {
			logger.Errorf("Notify FTP file %s upload failed: %s", fid, err)
			continue
		}
		_ = os.Remove(path)
		logger.Infof("Upload remain FTP file %s success", path)
	}
	logger.Debug("Upload remain replay done")
}

func uploadRemainReplay(jmsService *service.JMService, remainFiles map[string]string) {
	if len(remainFiles) == 0 {
		logger.Info("No remain replay files")
		return
	}
	logger.Infof("Start upload remain %d replay files 10 min later", len(remainFiles))
	time.Sleep(10 * time.Minute)

	var replayStorage storage.ReplayStorage
	terminalConf, _ := jmsService.GetTerminalConfig()
	replayStorage = storage.NewReplayStorage(jmsService, terminalConf.ReplayStorage)
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
		targetName := strings.Join([]string{replayDateDirName,
			sid + session.ReplayFileNameSuffix}, "/")
		err = replayStorage.Upload(absGzPath, targetName)
		if err != nil {
			logger.Errorf("Upload replay failed: %s", err)
			reason := model.SessionReplayErrUploadFailed
			if _, err1 := jmsService.SessionReplayFailed(sid, reason); err1 != nil {
				logger.Errorf("Update Session[%s] status %s failed: %s", sid, reason, err1)
			}
			continue
		}

		absGzFileInfo, err := os.Stat(absGzPath)
		if err != nil {
			logger.Errorf("Get file info %s failed: %s", absGzPath, err)
			continue
		}

		logger.Infof("Upload remain session replay %s success", absGzPath)
		// 上传成功删除文件
		_ = os.Remove(absGzPath)
		if _, err = jmsService.FinishReplyWithSize(sid, absGzFileInfo.Size()); err != nil {
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
		if common.IsUUID(sidFilename) {
			sid = sidFilename
		}
		if sid != "" {
			allRemainFiles[sid] = path
			if _, err = jmsService.SessionFinished(sid, common.NewUTCTime(info.ModTime())); err != nil {
				logger.Errorf("Session[%s] finished err: %s", sid, err)
			}
		}
		return nil
	})
	return allRemainFiles
}

func uploadRemainSessionPartReplay(jmsService *service.JMService, sessionDir string) {
	sessions, err := os.ReadDir(sessionDir)
	if err != nil {
		logger.Errorf("Read session dir failed: %s", err)
		return
	}
	if len(sessions) == 0 {
		logger.Info("No remain replay files")
		return
	}
	logger.Infof("Start upload remain %d session replay files 10 min later", len(sessions))
	time.Sleep(10 * time.Minute)

	terminalConf, _ := jmsService.GetTerminalConfig()
	for _, sessionEntry := range sessions {
		sessionId := sessionEntry.Name()
		if !common.IsUUID(sessionId) {
			continue
		}
		sessionRootPath := filepath.Join(sessionDir, sessionId)
		uploader := tunnel.PartUploader{
			RootPath:  sessionRootPath,
			SessionId: sessionId,
			ApiClient: jmsService,
			TermCfg:   &terminalConf,
		}
		uploader.Start()
		logger.Infof("Upload remain session part replay %s finish", sessionId)
	}
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
			string(model.Lion),
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
	return workClient
}

func NewPandaClient(cfg config.Config) *panda.Client {
	pandaHost := cfg.PandaHost
	var key model.AccessKey
	if err := key.LoadFromFile(cfg.AccessKeyFilePath); err != nil {
		logger.Errorf("Create panda client failed: loading access key err %s", err)
		return nil
	}
	return panda.NewClient(pandaHost, key, cfg.IgnoreVerifyCerts)
}
