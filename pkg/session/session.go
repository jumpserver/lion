package session

import (
	"strings"

	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

type Session struct {
	ID         string
	Created    common.UTCTime
	Asset      *model.Asset
	SystemUser *model.SystemUser
	User       *model.User


	ConnectedCallback    func() error
	DisConnectedCallback func() error
}

const (
	vnc = "vnc"
	rdp = "rdp"
)

const (
	BoolDisable = "false"
	BoolEnable  = "true"
)

func (s Session) GuaConfiguration() guacd.Configuration {
	switch strings.ToLower(s.SystemUser.Protocol) {
	case vnc:
		return s.ConfigurationVNC()
	case rdp:
		return s.ConfigurationRDP()
	default:
	}
	return guacd.Configuration{}
}

func (s Session) ConfigurationVNC() guacd.Configuration {
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

	return conf
}
func (s Session) ConfigurationRDP() guacd.Configuration {
	var (
		username string
		password string
		ip       string
		port     string
	)
	conf := guacd.NewConfiguration()
	conf.SetParameter(guacd.RDPUsername, username)
	conf.SetParameter(guacd.RDPPassword, password)
	conf.SetParameter(guacd.RDPHostname, ip)
	conf.SetParameter(guacd.RDPPort, port)
	conf.SetParameter(guacd.RDPSecurity, "any")
	conf.SetParameter(guacd.RDPIgnoreCert, BoolEnable)

	conf.SetParameter(guacd.RDPResizeMethod, "reconnect")
	conf.SetParameter(guacd.RDPDisableGlyphCaching, BoolEnable)

	conf.SetParameter(guacd.RDPEnableDrive, BoolEnable)
	conf.SetParameter(guacd.RDPCreateDrivePath, BoolEnable)

	conf.SetParameter(guacd.RDPDrivePath, config.GlobalConfig.DrivePath)
	conf.SetParameter(guacd.RDPDriveName, "Jumpserver")
	return conf
}
