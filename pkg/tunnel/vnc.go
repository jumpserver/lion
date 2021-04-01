package tunnel

import (
	"github.com/gin-gonic/gin"
	"guacamole-client-go/pkg/guacd"
	"io/ioutil"
	"os"
	"strings"
)

func GetVNCConfiguration(ctx *gin.Context) guacd.Configuration {
	configuration := guacd.NewConfiguration()
	configuration.Protocol = "vnc"
	// todo 临时获取测试的服务器
	testFd, _ := os.Open("vnc_server.out")
	defer testFd.Close()
	data, _ := ioutil.ReadAll(testFd)
	rdpData := strings.Split(string(data), ",")
	username := rdpData[0]
	password := rdpData[1]
	ip := rdpData[2]
	port := rdpData[3]
	configuration.SetParameter(guacd.VNCUsername, username)
	configuration.SetParameter(guacd.VNCPassword, password)
	configuration.SetParameter(guacd.VNCHostname, ip)
	configuration.SetParameter(guacd.VNCPort, port)
	//configuration.SetParameter(guacd.RDPSecurity, "any")
	//configuration.SetParameter(guacd.RDPIgnoreCert, BoolEnable)

	//configuration.SetParameter(guacd.RDPResizeMethod, "reconnect")
	//configuration.SetParameter(guacd.RDPDisableGlyphCaching, BoolEnable)

	//configuration.SetParameter(guacd.RDPEnableDrive, BoolEnable)
	//configuration.SetParameter(guacd.RDPCreateDrivePath, BoolEnable)

	//configuration.SetParameter(guacd.RDPDrivePath, config.GlobalConfig.DrivePath)
	//configuration.SetParameter(guacd.RDPDriveName, "Jumpserver")

	return configuration
}
