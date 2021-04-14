package config

import (
	"log"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func setupFromYmlFile() {
	f, err := os.Open("config_test.yml")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	viper.SetConfigType("yml")
	viper.AutomaticEnv()
	err = viper.ReadConfig(f)
	if err != nil {
		log.Fatal(err)
	}
}

func TestSetupByYml(t *testing.T) {
	setupFromYmlFile()
	var conf = getDefaultConfig()
	if err := viper.Unmarshal(&conf); err != nil {
		t.Fatalf("%s \n", err.Error())
	}
	t.Log(conf)
}

func TestSetupByEnv(t *testing.T) {
	viper.AutomaticEnv()
	loadEnvToViper()
	var conf = getDefaultConfig()
	if err := viper.Unmarshal(&conf); err != nil {
		t.Fatalf("%s \n", err.Error())
	}
	t.Logf("%+v\n",conf)
}
