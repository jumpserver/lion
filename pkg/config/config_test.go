package config

import (
	"testing"
)

func TestSetupByYml(t *testing.T) {
	var conf = getDefaultConfig()
	loadConfigFromFile("config_test.yml", &conf)
	t.Log(conf)
}

func TestSetupByEnv(t *testing.T) {
	var conf = getDefaultConfig()
	loadConfigFromEnv(&conf)
	t.Log(conf)
}
