package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
	"guacamole-client-go/pkg/logger"
	"guacamole-client-go/pkg/middleware"
	"guacamole-client-go/pkg/session"
	"guacamole-client-go/pkg/storage"
	"guacamole-client-go/pkg/tunnel"
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
	if configPath != "" {
		viper.AddConfigPath(configPath)
	}
	config.Setup()
	logger.SetupLogger(config.GlobalConfig)
	logger.Debug(config.GlobalConfig.DrivePath)
	jmsService := MustJMService()
	bootstrap(jmsService)
	tunnelService := tunnel.GuacamoleTunnelServer{
		Cache: &tunnel.GuaTunnelCache{
			Tunnels: make(map[string]*tunnel.Connection),
		},
		SessCache: &tunnel.SessionCache{
			Sessions: make(map[string]*session.TunnelSession),
		},
		JmsService:     jmsService,
		SessionService: &session.Server{JmsService: jmsService},
	}
	eng := registerRouter(jmsService, &tunnelService)
	go runHeartTask(jmsService, tunnelService.Cache)
	addr := net.JoinHostPort(config.GlobalConfig.BindHost, config.GlobalConfig.HTTPPort)
	logger.Infof("listen on: %s", addr)
	logger.Fatal(http.ListenAndServe(addr, eng))
}
func runHeartTask(jmsService *service.JMService, cache *tunnel.GuaTunnelCache) {
	beatTicker := time.NewTicker(time.Minute)
	defer beatTicker.Stop()
	for {
		select {
		case <-beatTicker.C:
			sids := cache.Range()
			tasks, err := jmsService.TerminalHeartBeat(sids)
			if err != nil {
				logger.Error(err)
				continue
			}
			for i := range tasks {
				task := tasks[i]
				switch task.Name {
				case model.TaskKillSession:
					if connection := cache.GetBySessionId(task.Args); connection != nil {
						connection.Terminate()
						if err = jmsService.FinishTask(task.ID); err != nil {
							logger.Error(err)
						}

					}
				default:
				}
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

	guacamoleGroup := eng.Group("/guacamole")
	// vue的设置
	{
		guacamoleGroup.Static("/assets", "./ui/guacamole/assets")
		guacamoleGroup.StaticFile("/", "./ui/guacamole/index.html")
	}

	// token 不需要认证
	tokenGroup := guacamoleGroup.Group("/token")
	{
		// TODO: 解决不认证，可能出现的安全问题
		tokenGroup.POST("/session", tunnelService.TokenSession)
		tokenGroup.GET("/tunnels/:tid/streams/:index/:filename", tunnelService.DownloadFile)
		tokenGroup.POST("/tunnels/:tid/streams/:index/:filename", tunnelService.UploadFile)
	}

	// ws的设置
	wsGroup := guacamoleGroup.Group("/ws")
	{
		wsGroup.Group("/connect").Use(
			middleware.SessionAuth(jmsService)).GET("/", tunnelService.Connect)

		wsGroup.Group("/token").GET("/", tunnelService.TokenConnect)
	}

	apiGroup := guacamoleGroup.Group("/api")
	apiGroup.Use(middleware.SessionAuth(jmsService))
	{
		apiGroup.POST("/session", tunnelService.CreateSession)
		apiGroup.GET("/tunnels/:tid/streams/:index/:filename", tunnelService.DownloadFile)
		apiGroup.POST("/tunnels/:tid/streams/:index/:filename", tunnelService.UploadFile)
	}

	pprofRouter := eng.Group("/debug/pprof")
	{
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
	allRemainFiles := scanRemainReplay(replayDir)
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
			logger.Error(replayDateDirName)
			continue
		}
		if !strings.HasSuffix(path, session.ReplayFileNameSuffix) {
			absGzPath = path + session.ReplayFileNameSuffix
			if err := common.CompressToGzipFile(path, absGzPath); err != nil {
				logger.Error(err)
				continue
			}
			_ = os.Remove(path)
		}
		var err error
		if replayStorage != nil {
			err = replayStorage.Upload(absGzPath, replayDateDirName)
		} else {
			err = jmsService.Upload(sid, absGzPath)
		}
		if err != nil {
			logger.Error(err)
			continue
		}
		// 上传成功删除文件
		_ = os.Remove(absGzPath)
		err = jmsService.FinishReply(sid)
	}
}

func scanRemainReplay(replayDir string) map[string]string {
	allRemainFiles := make(map[string]string)
	_ = filepath.Walk(replayDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		sidFilename := info.Name()
		var sid string
		if strings.HasSuffix(sidFilename, session.ReplayFileNameSuffix) {
			sidFilename = strings.TrimSuffix(sidFilename, session.ReplayFileNameSuffix)
		}
		if common.ValidUUIDString(sidFilename) {
			sid = sidFilename
		}
		if sid != "" {
			allRemainFiles[sid] = path
		}
		return nil
	})
	return allRemainFiles
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
