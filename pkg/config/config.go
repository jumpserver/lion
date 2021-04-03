package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

type Config struct {
	Root              string
	DrivePath         string
	AccessKeyFilePath string

	Name           string `mapstructure:"NAME"`
	CoreHost       string `mapstructure:"CORE_HOST"`
	BootstrapToken string `mapstructure:"BOOTSTRAP_TOKEN"`
	BindHost       string `mapstructure:"BIND_HOST"`
	HTTPPort       string `mapstructure:"HTTPD_PORT"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`

	GuaHost string `mapstructure:"GUA_HOST"`
	GuaPort string `mapstructure:"GUA_PORT"`
}

func Setup() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
	var conf = getDefaultConfig()
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatal(err)
	}
	GlobalConfig = &conf

}

func getDefaultConfig() Config {
	defaultName := getDefaultName()
	rootPath := getPwdDirPath()
	dataFolderPath := filepath.Join(rootPath, "data")
	driveFolderPath := filepath.Join(dataFolderPath, "drive")
	keyFolderPath := filepath.Join(dataFolderPath, "key")
	accessKeyFilePath := filepath.Join(keyFolderPath, ".access_key")

	folders := []string{dataFolderPath, driveFolderPath, keyFolderPath}
	for i := range folders {
		if err := EnsureDirExist(folders[i]); err != nil {
			log.Fatalf("Create folder failed: %s", err.Error())
		}
	}
	return Config{
		Name:              defaultName,
		Root:              rootPath,
		DrivePath:         driveFolderPath,
		AccessKeyFilePath: accessKeyFilePath,
		CoreHost:          "http://127.0.0.1:8080",
		BootstrapToken:    "",
		BindHost:          "0.0.0.0",
		HTTPPort:          "8081",
		LogLevel:          "INFO",
		GuaHost:           "127.0.0.1",
		GuaPort:           "4822",
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
