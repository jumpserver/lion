package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http/pprof"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
	"guacamole-client-go/pkg/logger"
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
	logger.Debug(config.GlobalConfig.DrivePath)
	eng := gin.New()
	eng.Use(gin.Recovery())
	eng.Use(gin.Logger())

	tunnelService := tunnel.GuacamoleTunnelServer{
		Tunnels:    map[string]*tunnel.TunnelConn{},
		JmsService: MustJMService(),
	}

	guacamoleGroup := eng.Group("/guacamole")
	{
		guacamoleGroup.GET("/ws", tunnelService.Connect)
	}
	guacamoleAPIGroup := guacamoleGroup.Group("/api")
	{
		guacamoleAPIGroup.POST("/session", tunnelService.CreateSession)
		guacamoleAPIGroup.GET("/tunnels/:tid/streams/:index/:filename", tunnelService.DownloadFile)
		guacamoleAPIGroup.POST("/tunnels/:tid/streams/:index/:filename", tunnelService.UploadFile)
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
	log.Fatal(eng.Run(":8081"))
}

func MustJMService() *service.JMService {
	key := MustLoadValidAccessKey()
	jmsService, err := service.NewAuthJMService(service.JMSCoreHost(
		config.GlobalConfig.CoreHost), service.JMSTimeOut(30*time.Second),
		service.JMSAccessKey(key.ID, key.Secret),
	)
	if err != nil {
		logger.Debug("创建JMS Service 失败 " + err.Error())
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
			logger.Debug(err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		key.ID = terminal.ServiceAccount.AccessKey.ID
		key.Secret = terminal.ServiceAccount.AccessKey.Secret
		if err := key.SaveToFile(conf.AccessKeyFilePath); err != nil {
			logger.Debug("保存key失败: " + err.Error())
		}
		return key
	}
	logger.Debug("注册终端失败退出")
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
				logger.Debug("校验 access key failed: " + err.Error())
			}
			time.Sleep(5 * time.Second)
			continue
		}
		return key
	}
	logger.Debug("校验 access key failed退出")
	os.Exit(1)
	return key
}
