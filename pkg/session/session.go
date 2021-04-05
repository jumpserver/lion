package session

import (
	"strconv"
	"strings"

	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

type ConnectSession struct {
	ID         string            `json:"id"`
	Created    common.UTCTime    `json:"-"`
	Asset      *model.Asset      `json:"asset"`
	SystemUser *model.SystemUser `json:"system_user"`
	User       *model.User       `json:"user"`
	Platform   *model.Platform   `json:"platform"`

	Domain *model.Domain `json:"domain"`

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

func (s ConnectSession) GuaConfiguration() guacd.Configuration {
	switch strings.ToLower(s.SystemUser.Protocol) {
	case vnc:
		return s.configurationVNC()
	case rdp:
		return s.configurationRDP()
	default:
	}
	return guacd.Configuration{}
}

func (s ConnectSession) configurationVNC() guacd.Configuration {
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
func (s ConnectSession) configurationRDP() guacd.Configuration {
	var (
		username string
		password string
		ip       string
		port     string
	)
	ip = s.Asset.IP
	port = strconv.Itoa(s.Asset.ProtocolPort(s.SystemUser.Protocol))
	username = s.SystemUser.Username
	password = s.SystemUser.Password

	conf := guacd.NewConfiguration()
	conf.Protocol = rdp
	conf.SetParameter(guacd.RDPHostname, ip)
	conf.SetParameter(guacd.RDPPort, port)

	conf.SetParameter(guacd.RDPUsername, username)
	conf.SetParameter(guacd.RDPPassword, password)

	if s.SystemUser.AdDomain != "" {
		conf.SetParameter(guacd.RDPDomain, s.SystemUser.AdDomain)
	}

	conf.SetParameter(guacd.RDPSecurity, SecurityAny)
	conf.SetParameter(guacd.RDPIgnoreCert, BoolEnable)

	conf.SetParameter(guacd.RDPResizeMethod, "reconnect")
	conf.SetParameter(guacd.RDPDisableGlyphCaching, BoolEnable)

	conf.SetParameter(guacd.RDPEnableDrive, BoolEnable)
	conf.SetParameter(guacd.RDPCreateDrivePath, BoolEnable)

	conf.SetParameter(guacd.RDPDrivePath, config.GlobalConfig.DrivePath)
	conf.SetParameter(guacd.RDPDriveName, "Jumpserver")
	return conf
}

const (
	SecurityAny       = "any"
	SecurityNla       = "nla"
	SecurityNlaExt    = "nla-ext"
	SecurityTls       = "tls"
	SecurityVmConnect = "vmconnect"
	SecurityRdp       = "rdp"
)
