package config

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/jumpserver-dev/sdk-go/common"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

type Config struct {
	Root              string
	DrivePath         string
	RecordPath        string
	FTPFilePath       string
	LogDirPath        string
	AccessKeyFilePath string
	CertsFolderPath   string
	SessionFolderPath string

	Name           string `mapstructure:"NAME"`
	CoreHost       string `mapstructure:"CORE_HOST"`
	BootstrapToken string `mapstructure:"BOOTSTRAP_TOKEN"`
	BindHost       string `mapstructure:"BIND_HOST"`
	HTTPPort       string `mapstructure:"HTTPD_PORT"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`

	GuacdAddrs string `mapstructure:"GUACD_ADDRS"`

	GuaHost                   string `mapstructure:"GUA_HOST"`
	GuaPort                   string `mapstructure:"GUA_PORT"`
	DisableAllCopyPaste       bool   `mapstructure:"JUMPSERVER_DISABLE_ALL_COPY_PASTE"`
	DisableAllUpDownload      bool   `mapstructure:"JUMPSERVER_DISABLE_ALL_UPLOAD_DOWNLOAD"`
	EnableRemoteAppUpDownLoad bool   `mapstructure:"JUMPSERVER_REMOTE_APP_UPLOAD_DOWNLOAD_ENABLE"`
	EnableRemoteAPPCopyPaste  bool   `mapstructure:"JUMPSERVER_REMOTE_APP_COPY_PASTE_ENABLE"`
	CleanDriveScheduleTime    int    `mapstructure:"JUMPSERVER_CLEAN_DRIVE_SCHEDULE_TIME"`

	ShareRoomType string `mapstructure:"SHARE_ROOM_TYPE"`
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     int    `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDBIndex  int    `mapstructure:"REDIS_DB_ROOM"`

	RedisSentinelPassword string `mapstructure:"REDIS_SENTINEL_PASSWORD"`
	RedisSentinelHosts    string `mapstructure:"REDIS_SENTINEL_HOSTS"`
	RedisUseSSL           bool   `mapstructure:"REDIS_USE_SSL"`

	EnableVideoWorker bool   `mapstructure:"ENABLE_VIDEO_WORKER"`
	VideoWorkerHost   string `mapstructure:"VIDEO_WORKER_HOST"`
	IgnoreVerifyCerts bool   `mapstructure:"IGNORE_VERIFY_CERTS"`
	PandaHost         string `mapstructure:"PANDA_HOST"`
	EnablePanda       bool   `mapstructure:"ENABLE_PANDA"`

	ReplayMaxSize    int    `mapstructure:"REPLAY_MAX_SIZE"`
	SecretEncryptKey string `mapstructure:"SECRET_ENCRYPT_KEY"`

	VncClipboardEncoding string `mapstructure:"VNC_CLIPBOARD_ENCODING"`
}

func (c *Config) UpdateRedisPassword(val string) {
	c.RedisPassword = val
}

func (c *Config) SelectGuacdAddr() string {
	if len(c.GuacdAddrs) == 0 {
		return net.JoinHostPort(c.GuaHost, c.GuaPort)
	}
	addresses := strings.Split(c.GuacdAddrs, ",")
	return addresses[rand.Intn(len(addresses))]
}

func Setup(configPath string) {
	var conf = getDefaultConfig()
	loadConfigFromEnv(&conf)
	loadConfigFromFile(configPath, &conf)
	GlobalConfig = &conf
	log.Printf("%+v\n", GlobalConfig)

}

func getDefaultConfig() Config {
	defaultName := getDefaultName()
	rootPath := getPwdDirPath()
	dataFolderPath := filepath.Join(rootPath, "data")
	driveFolderPath := filepath.Join(dataFolderPath, "drive")
	recordFolderPath := filepath.Join(dataFolderPath, "replays")
	sessionsPath := filepath.Join(dataFolderPath, "sessions")
	ftpFileFolderPath := filepath.Join(dataFolderPath, "ftp_files")
	LogDirPath := filepath.Join(dataFolderPath, "logs")
	keyFolderPath := filepath.Join(dataFolderPath, "keys")
	CertsFolderPath := filepath.Join(dataFolderPath, "certs")
	accessKeyFilePath := filepath.Join(keyFolderPath, ".access_key")

	folders := []string{dataFolderPath, driveFolderPath, recordFolderPath,
		keyFolderPath, LogDirPath, CertsFolderPath, sessionsPath}
	for i := range folders {
		if err := EnsureDirExist(folders[i]); err != nil {
			log.Fatalf("Create folder failed: %s", err.Error())
		}
	}
	return Config{
		Name:                      defaultName,
		Root:                      rootPath,
		RecordPath:                recordFolderPath,
		FTPFilePath:               ftpFileFolderPath,
		LogDirPath:                LogDirPath,
		DrivePath:                 driveFolderPath,
		CertsFolderPath:           CertsFolderPath,
		AccessKeyFilePath:         accessKeyFilePath,
		SessionFolderPath:         sessionsPath,
		CoreHost:                  "http://localhost:8080",
		BootstrapToken:            "",
		BindHost:                  "0.0.0.0",
		HTTPPort:                  "8081",
		LogLevel:                  "INFO",
		GuaHost:                   "127.0.0.1",
		GuaPort:                   "4822",
		DisableAllCopyPaste:       false,
		DisableAllUpDownload:      false,
		EnableRemoteAppUpDownLoad: false,
		EnableRemoteAPPCopyPaste:  false,
		CleanDriveScheduleTime:    1,
		PandaHost:                 "http://panda:9001",
		ReplayMaxSize:             defaultMaxSize,
	}

}

// 300MB
const defaultMaxSize = 1024 * 1024 * 300

func EnsureDirExist(path string) error {
	if !haveDir(path) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func have(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func haveDir(file string) bool {
	fi, err := os.Stat(file)
	return err == nil && fi.IsDir()
}

func getPwdDirPath() string {
	if rootPath, err := os.Getwd(); err == nil {
		return rootPath
	}
	return ""
}

func loadConfigFromEnv(conf *Config) {
	viper.AutomaticEnv() // 全局配置，用于其他 pkg 包可以用 viper 获取环境变量的值
	envViper := viper.New()
	for _, item := range os.Environ() {
		envItem := strings.SplitN(item, "=", 2)
		if len(envItem) == 2 {
			envViper.Set(envItem[0], viper.Get(envItem[0]))
		}
	}
	if err := envViper.Unmarshal(conf); err == nil {
		log.Println("Load config from env")
	}

}

func loadConfigFromFile(path string, conf *Config) {
	var err error
	if have(path) {
		fileViper := viper.New()
		fileViper.SetConfigFile(path)
		if err = fileViper.ReadInConfig(); err == nil {
			if err = fileViper.Unmarshal(conf); err == nil {
				log.Printf("Load config from %s success\n", path)
				return
			}
		}
	}
	if err != nil {
		log.Fatalf("Load config from %s failed: %s\n", path, err)
	}
}

const (
	prefixName = "[Lion]-"

	hostEnvKey = "SERVER_HOSTNAME"

	defaultNameMaxLen = 128
)

/*
SERVER_HOSTNAME: 环境变量名，可用于自定义默认注册名称的前缀
default name rule:
[Lion]-{SERVER_HOSTNAME}-{HOSTNAME}-{UUID}
 or
[Lion]-{HOSTNAME}-{UUID}
*/

func getDefaultName() string {
	hostname, _ := os.Hostname()
	hostname = fmt.Sprintf("%s-%s", hostname, common.RandomStr(7))
	if serverHostname, ok := os.LookupEnv(hostEnvKey); ok {
		hostname = fmt.Sprintf("%s-%s", serverHostname, hostname)
	}
	hostRune := []rune(prefixName + hostname)
	if len(hostRune) <= defaultNameMaxLen {
		return string(hostRune)
	}
	name := make([]rune, defaultNameMaxLen)
	index := defaultNameMaxLen / 2
	copy(name[:16], hostRune[:index])
	start := len(hostRune) - index
	copy(name[index:], hostRune[start:])
	return string(name)
}
