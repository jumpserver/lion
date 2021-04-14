package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

type Config struct {
	Root              string
	DrivePath         string
	RecordPath        string
	LogDirPath        string
	AccessKeyFilePath string

	Name           string `mapstructure:"NAME"`
	CoreHost       string `mapstructure:"CORE_HOST"`
	BootstrapToken string `mapstructure:"BOOTSTRAP_TOKEN"`
	BindHost       string `mapstructure:"BIND_HOST"`
	HTTPPort       string `mapstructure:"HTTPD_PORT"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`

	GuaHost                   string `mapstructure:"GUA_HOST"`
	GuaPort                   string `mapstructure:"GUA_PORT"`
	DisableAllCopyPaste       bool   `mapstructure:"JUMPSERVER_DISABLE_ALL_COPY_PASTE"`
	DisableAllUpDownload      bool   `mapstructure:"JUMPSERVER_DISABLE_ALL_UPLOAD_DOWNLOAD"`
	EnableRemoteAppUpDownLoad bool   `mapstructure:"JUMPSERVER_REMOTE_APP_UPLOAD_DOWNLOAD_ENABLE"`
	EnableRemoteAPPCopyPaste  bool   `mapstructure:"JUMPSERVER_REMOTE_APP_COPY_PASTE_ENABLE"`
}

func Setup() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	loadEnvToViper()
	log.Println("Load config from env")
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Load config from config.yml again")
	}
	var conf = getDefaultConfig()
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatal(err)
	}
	GlobalConfig = &conf
	log.Printf("%+v\n", GlobalConfig)

}

func getDefaultConfig() Config {
	defaultName := getDefaultName()
	rootPath := getPwdDirPath()
	dataFolderPath := filepath.Join(rootPath, "data")
	driveFolderPath := filepath.Join(dataFolderPath, "drive")
	recordFolderPath := filepath.Join(dataFolderPath, "record")
	LogDirPath := filepath.Join(dataFolderPath, "log")
	keyFolderPath := filepath.Join(dataFolderPath, "key")
	accessKeyFilePath := filepath.Join(keyFolderPath, ".access_key")

	folders := []string{dataFolderPath, driveFolderPath, keyFolderPath, LogDirPath}
	for i := range folders {
		if err := EnsureDirExist(folders[i]); err != nil {
			log.Fatalf("Create folder failed: %s", err.Error())
		}
	}
	return Config{
		Name:                      defaultName,
		Root:                      rootPath,
		RecordPath:                recordFolderPath,
		LogDirPath:                LogDirPath,
		DrivePath:                 driveFolderPath,
		AccessKeyFilePath:         accessKeyFilePath,
		CoreHost:                  "http://127.0.0.1:8080",
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
	}

}

func getPwdDirPath() string {
	if rootPath, err := os.Getwd(); err == nil {
		return rootPath
	}
	return ""
}

func have(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func haveDir(file string) bool {
	fi, err := os.Stat(file)
	return err == nil && fi.IsDir()
}

func EnsureDirExist(path string) error {
	if !haveDir(path) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

const prefixName = "[Guacamole]"

func getDefaultName() string {
	hostname, _ := os.Hostname()
	hostRune := []rune(prefixName + hostname)
	if len(hostRune) <= 32 {
		return string(hostRune)
	}
	name := make([]rune, 32)
	copy(name[:16], hostRune[:16])
	start := len(hostRune) - 16
	copy(name[16:], hostRune[start:])
	return string(name)
}

func loadEnvToViper() {
	for _, item := range os.Environ() {
		envItem := strings.Split(item, "=")
		if len(envItem) == 2 {
			viper.Set(envItem[0], envItem[1])
		}
	}
}
