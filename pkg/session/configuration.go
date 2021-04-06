package session

import (
	"strconv"

	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

type ConnectionConfiguration interface {
	GetGuacdConfiguration() guacd.Configuration
}

type RDPConfiguration struct {
	Asset      *model.Asset      `json:"asset"`
	SystemUser *model.SystemUser `json:"system_user"`
	Platform   *model.Platform   `json:"platform"`
}

func (r RDPConfiguration) GetGuacdConfiguration() guacd.Configuration {
	var (
		username string
		password string
		ip       string
		port     string
	)
	ip = r.Asset.IP
	port = strconv.Itoa(r.Asset.ProtocolPort(r.SystemUser.Protocol))
	username = r.SystemUser.Username
	password = r.SystemUser.Password

	conf := guacd.NewConfiguration()
	conf.Protocol = rdp
	conf.SetParameter(guacd.RDPHostname, ip)
	conf.SetParameter(guacd.RDPPort, port)

	conf.SetParameter(guacd.RDPUsername, username)
	conf.SetParameter(guacd.RDPPassword, password)

	if r.SystemUser.AdDomain != "" {
		conf.SetParameter(guacd.RDPDomain, r.SystemUser.AdDomain)
	}

	conf.SetParameter(guacd.RDPSecurity, SecurityAny)
	conf.SetParameter(guacd.RDPIgnoreCert, BoolTrue)

	conf.SetParameter(guacd.RDPResizeMethod, "reconnect")
	conf.SetParameter(guacd.RDPDisableGlyphCaching, BoolTrue)

	conf.SetParameter(guacd.RDPEnableDrive, BoolTrue)
	conf.SetParameter(guacd.RDPCreateDrivePath, BoolTrue)

	conf.SetParameter(guacd.RDPDrivePath, config.GlobalConfig.DrivePath)
	conf.SetParameter(guacd.RDPDriveName, "Jumpserver")
	return guacd.Configuration{}
}

type VNCConfiguration struct {
}

func (r VNCConfiguration) GetGuacdConfiguration() guacd.Configuration {
	conf := guacd.NewConfiguration()
	var (
		username string
		password string
		ip       string
		port     string
	)

	conf.SetParameter(guacd.VNCUsername, username)
	conf.SetParameter(guacd.VNCPassword, password)
	conf.SetParameter(guacd.VNCHostname, ip)
	conf.SetParameter(guacd.VNCPort, port)

	return guacd.Configuration{}
}

type RemoteAPPConfiguration struct {
}

func (r RemoteAPPConfiguration) GetGuacdConfiguration() guacd.Configuration {

	return guacd.Configuration{}
}
