package main

import (
	"guacamole-client-go/pkg/config"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"guacamole-client-go/pkg/guacd"
)

const (
	BoolDisable = "false"
	BoolEnable  = "true"
)

func GetRDPConfiguration(ctx *gin.Context) guacd.Configuration {
	configuration := guacd.NewConfiguration()
	configuration.Protocol = "rdp"
	// todo 临时获取测试的服务器
	testFd, _ := os.Open("rdp_server.out")
	defer testFd.Close()
	data, _ := ioutil.ReadAll(testFd)
	rdpData := strings.Split(string(data), ",")
	username := rdpData[0]
	password := rdpData[1]
	ip := rdpData[2]
	port := rdpData[3]
	configuration.SetParameter(guacd.RDPUsername, username)
	configuration.SetParameter(guacd.RDPPassword, password)
	configuration.SetParameter(guacd.RDPHostname, ip)
	configuration.SetParameter(guacd.RDPPort, port)
	configuration.SetParameter(guacd.RDPSecurity, "any")
	configuration.SetParameter(guacd.RDPIgnoreCert, BoolEnable)

	configuration.SetParameter(guacd.RDPResizeMethod, "reconnect")
	configuration.SetParameter(guacd.RDPDisableGlyphCaching, BoolEnable)

	configuration.SetParameter(guacd.RDPEnableDrive, BoolEnable)
	configuration.SetParameter(guacd.RDPCreateDrivePath, BoolEnable)

	configuration.SetParameter(guacd.RDPDrivePath, config.GlobalConfig.DrivePath)
	configuration.SetParameter(guacd.RDPDriveName, "Jumpserver")

	return configuration
}
