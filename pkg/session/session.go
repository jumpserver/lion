package session

import (
	"strings"

	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

type TunnelSession struct {
	ID             string                `json:"id"`
	Created        common.UTCTime        `json:"-"`
	Asset          *model.Asset          `json:"asset"`
	SystemUser     *model.SystemUser     `json:"system_user"`
	User           *model.User           `json:"user"`
	Platform       *model.Platform       `json:"platform"`
	RemoteApp      *model.RemoteAPP      `json:"remote_app"`
	Permission     *model.Permission     `json:"permission"`
	Domain         *model.Domain         `json:"domain"`
	TerminalConfig *model.TerminalConfig `json:"-"`

	ConnectedCallback    func() error `json:"-"`
	DisConnectedCallback func() error `json:"-"`
}

const (
	vnc = "vnc"
	rdp = "rdp"
)

func (s TunnelSession) GuaConfiguration() guacd.Configuration {
	switch strings.ToLower(s.SystemUser.Protocol) {
	case vnc:
		return s.configurationVNC()
	default:
		return s.configurationRDP()
	}
}

func (s TunnelSession) configurationVNC() guacd.Configuration {
	conf := VNCConfiguration{
		SessionId:      s.ID,
		User:           s.User,
		Asset:          s.Asset,
		SystemUser:     s.SystemUser,
		Platform:       s.Platform,
		Permission:     s.Permission,
		TerminalConfig: s.TerminalConfig,
	}
	return conf.GetGuacdConfiguration()
}

func (s TunnelSession) configurationRDP() guacd.Configuration {
	rdpConf := RDPConfiguration{
		SessionId:      s.ID,
		User:           s.User,
		Asset:          s.Asset,
		SystemUser:     s.SystemUser,
		Platform:       s.Platform,
		Permission:     s.Permission,
		TerminalConfig: s.TerminalConfig,
	}
	if s.RemoteApp != nil {
		remoteConf := RemoteAPPConfiguration{
			RDPConfiguration: rdpConf,
			RemoteApp:        s.RemoteApp,
		}
		return remoteConf.GetGuacdConfiguration()
	}
	return rdpConf.GetGuacdConfiguration()
}

const (
	SecurityAny       = "any"
	SecurityNla       = "nla"
	SecurityNlaExt    = "nla-ext"
	SecurityTls       = "tls"
	SecurityVmConnect = "vmconnect"
	SecurityRdp       = "rdp"
)
