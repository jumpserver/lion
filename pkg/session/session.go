package session

import (
	"lion/pkg/common"
	"lion/pkg/guacd"
	"lion/pkg/jms-sdk-go/model"
)

type TunnelSession struct {
	ID             string                `json:"id"`
	Protocol       string                `json:"protocol"`
	Created        common.UTCTime        `json:"-"`
	Asset          *model.Asset          `json:"asset"`
	Account        *model.Account        `json:"-"`
	User           *model.User           `json:"user"`
	Platform       *model.Platform       `json:"platform"`
	RemoteApp      *model.Applet         `json:"remote_app"`
	Permission     *model.Permission     `json:"permission"`
	Domain         *model.Domain         `json:"-"`
	Gateway        *model.Gateway        `json:"-"`
	TerminalConfig *model.TerminalConfig `json:"-"`
	ExpireInfo     model.ExpireInfo      `json:"expire_info"`
	ActionPerm     *ActionPermission     `json:"action_permission"`
	DisplayAccount *model.Account        `json:"system_user"`

	AppletOpts *model.AppletOption `json:"-"`
	AuthInfo   *model.ConnectToken `json:"-"`

	ConnectedCallback        func() error          `json:"-"`
	ConnectedSuccessCallback func() error          `json:"-"`
	ConnectedFailedCallback  func(err error) error `json:"-"`
	DisConnectedCallback     func() error          `json:"-"`

	FinishReplayCallback func(guacd.ClientInformation) error `json:"-"`

	ReleaseAppletAccount func() error `json:"-"`
}

const (
	vnc = "vnc"
	rdp = "rdp"
)

func (s TunnelSession) GuaConfiguration() guacd.Configuration {
	switch s.Protocol {
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
		Account:        s.Account,
		Platform:       s.Platform,
		TerminalConfig: s.TerminalConfig,
		ActionsPerm:    s.ActionPerm,
	}
	return conf.GetGuacdConfiguration()
}

func (s TunnelSession) configurationRDP() guacd.Configuration {
	if s.AppletOpts != nil {
		return s.configurationRemoteAppRDP()
	}
	rdpConf := RDPConfiguration{
		SessionId:      s.ID,
		Created:        s.Created,
		User:           s.User,
		Asset:          s.Asset,
		Account:        s.Account,
		Platform:       s.Platform,
		TerminalConfig: s.TerminalConfig,
		ActionsPerm:    s.ActionPerm,
	}
	return rdpConf.GetGuacdConfiguration()
}

func (s TunnelSession) configurationRemoteAppRDP() guacd.Configuration {
	appletOpt := s.AppletOpts
	rdpConf := RDPConfiguration{
		SessionId:      s.ID,
		Created:        s.Created,
		User:           s.User,
		Asset:          &appletOpt.Host,
		Account:        &appletOpt.Account,
		TerminalConfig: s.TerminalConfig,
		ActionsPerm:    s.ActionPerm,
	}
	conf := rdpConf.GetGuacdConfiguration()
	remoteAPP := appletOpt.RemoteAppOption
	// 设置 remote app 参数
	{
		conf.SetParameter(guacd.RDPRemoteApp, remoteAPP.Name)
		conf.SetParameter(guacd.RDPRemoteAppDir, "")
		conf.SetParameter(guacd.RDPRemoteAppArgs, remoteAPP.CmdLine)
	}
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

func ValidateSecurityValue(security string) bool {
	switch security {
	case SecurityAny,
		SecurityNla,
		SecurityNlaExt,
		SecurityTls,
		SecurityVmConnect,
		SecurityRdp:
		return true
	}
	return false
}
