package session

import (
	"strings"

	"lion/pkg/common"
	"lion/pkg/guacd"
	"lion/pkg/jms-sdk-go/model"
)

type TunnelSession struct {
	ID             string                    `json:"id"`
	Created        common.UTCTime            `json:"-"`
	Asset          *model.Asset              `json:"asset"`
	SystemUser     *model.SystemUserAuthInfo `json:"-"`
	User           *model.User               `json:"user"`
	Platform       *model.Platform           `json:"platform"`
	RemoteApp      *model.RemoteAPP          `json:"remote_app"`
	Permission     *model.Permission         `json:"permission"`
	Domain         *model.Domain             `json:"-"`
	TerminalConfig *model.TerminalConfig     `json:"-"`
	ExpireInfo     *model.ExpireInfo         `json:"expire_info"`
	ActionPerm     *ActionPermission         `json:"action_permission"`

	DisplaySystemUser *model.SystemUser `json:"system_user"`

	ConnectedCallback        func() error          `json:"-"`
	ConnectedSuccessCallback func() error          `json:"-"`
	ConnectedFailedCallback  func(err error) error `json:"-"`
	DisConnectedCallback     func() error          `json:"-"`
	FinishReplayCallback     func() error          `json:"-"`
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
		Created:        s.Created,
		User:           s.User,
		Asset:          s.Asset,
		SystemUser:     s.SystemUser,
		Platform:       s.Platform,
		TerminalConfig: s.TerminalConfig,
		ActionsPerm:    s.ActionPerm,
	}
	return conf.GetGuacdConfiguration()
}

func (s TunnelSession) configurationRDP() guacd.Configuration {
	rdpConf := RDPConfiguration{
		SessionId:      s.ID,
		Created:        s.Created,
		User:           s.User,
		Asset:          s.Asset,
		SystemUser:     s.SystemUser,
		Platform:       s.Platform,
		TerminalConfig: s.TerminalConfig,
		ActionsPerm:    s.ActionPerm,
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
