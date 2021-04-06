package session

import (
	"path/filepath"
	"strconv"

	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

type ConnectionConfiguration interface {
	GetGuacdConfiguration() guacd.Configuration
}

type RDPConfiguration struct {
	SessionId      string
	User           *model.User
	Asset          *model.Asset          `json:"asset"`
	SystemUser     *model.SystemUser     `json:"system_user"`
	Platform       *model.Platform       `json:"platform"`
	Permission     *model.Permission     `json:"permission"`
	TerminalConfig *model.TerminalConfig `json:"terminal_config"`
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

	// 设置 录像路径
	if r.TerminalConfig.ReplayStorage["type"] != "null" {
		// TODO: 加上录像创建日期
		conf.SetParameter(guacd.RecordingPath, config.GlobalConfig.RecordPath)
		conf.SetParameter(guacd.CreateRecordingPath, BoolTrue)
		conf.SetParameter(guacd.RecordingName, r.SessionId)
	}

	conf.SetParameter(guacd.RDPResizeMethod, "reconnect")
	conf.SetParameter(guacd.RDPDisableGlyphCaching, BoolTrue)

	// 设置 挂载目录 上传下载
	{
		drivePath := filepath.Join(config.GlobalConfig.DrivePath, r.User.ID)
		enableDrive := ConvertBoolToString(r.Permission.EnableDrive())
		disableDownload := ConvertBoolToString(!r.Permission.EnableDownload())
		disableUpload := ConvertBoolToString(!r.Permission.EnableUpload())
		conf.SetParameter(guacd.RDPDrivePath, drivePath)
		conf.SetParameter(guacd.RDPCreateDrivePath, BoolTrue)
		conf.SetParameter(guacd.RDPEnableDrive, enableDrive)
		conf.SetParameter(guacd.RDPDriveName, "JumpServer")
		conf.SetParameter(guacd.RDPDisableDownload, disableDownload)
		conf.SetParameter(guacd.RDPDisableUpload, disableUpload)
	}

	// 粘贴复制
	{
		disableCopy := ConvertBoolToString(!r.Permission.EnableCopy())
		disablePaste := ConvertBoolToString(!r.Permission.EnablePaste())
		conf.SetParameter(guacd.DisableCopy, disableCopy)
		conf.SetParameter(guacd.DisablePaste, disablePaste)
	}

	// platform meta 数据
	{
		for k, v := range ConvertMetaToParams(r.Platform.MetaData) {
			conf.SetParameter(k, v)
		}
	}

	return conf
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

	return conf
}

type RemoteAPPConfiguration struct {
	RDPConfiguration
	RemoteApp *model.RemoteAPP
}

func (r RemoteAPPConfiguration) GetGuacdConfiguration() guacd.Configuration {
	conf := r.RDPConfiguration.GetGuacdConfiguration()

	// 设置 remote app 参数
	{
		conf.SetParameter(guacd.RDPRemoteApp, r.RemoteApp.Parameters.Program)
		conf.SetParameter(guacd.RDPRemoteAppDir, r.RemoteApp.Parameters.WorkingDirectory)
		conf.SetParameter(guacd.RDPRemoteAppArgs, r.RemoteApp.Parameters.Parameters)
	}
	return conf
}

func ConvertMetaToParams(meta map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k, v := range meta {
		switch value := v.(type) {
		case string:
			res[k] = value
		case bool:
			res[k] = ConvertBoolToString(value)
		case int:
			res[k] = strconv.Itoa(value)
		case float64:
			res[k] = strconv.FormatFloat(value, 'E', -1, 64)
		}
	}

	return res
}
