package session

import (
	"path/filepath"
	"strconv"

	"lion/pkg/common"
	"lion/pkg/config"
	"lion/pkg/guacd"
	"lion/pkg/jms-sdk-go/model"
)

type ConnectionConfiguration interface {
	GetGuacdConfiguration() guacd.Configuration
}

var _ ConnectionConfiguration = RDPConfiguration{}
var _ ConnectionConfiguration = VNCConfiguration{}
var _ ConnectionConfiguration = RemoteAPPConfiguration{}

type RDPConfiguration struct {
	SessionId      string
	Created        common.UTCTime
	User           *model.User
	Asset          *model.Asset
	Account        *model.Account
	Platform       *model.Platform
	TerminalConfig *model.TerminalConfig
	ActionsPerm    *ActionPermission
}

func (r RDPConfiguration) GetGuacdConfiguration() guacd.Configuration {
	var (
		username string
		password string
		ip       string
		port     string
	)
	ip = r.Asset.Address
	port = strconv.Itoa(r.Asset.ProtocolPort("rdp"))
	username = r.Account.Username
	password = r.Account.Secret

	conf := guacd.NewConfiguration()
	conf.Protocol = rdp
	conf.SetParameter(guacd.Hostname, ip)
	conf.SetParameter(guacd.Port, port)

	conf.SetParameter(guacd.RDPUsername, username)
	conf.SetParameter(guacd.RDPPassword, password)

	// todo: 账户 域账号
	//if r.SystemUser.AdDomain != "" {
	//	conf.SetParameter(guacd.RDPDomain, r.SystemUser.AdDomain)
	//}
	conf.SetParameter(guacd.RDPSecurity, SecurityAny)
	conf.SetParameter(guacd.RDPIgnoreCert, BoolTrue)

	// 设置 录像路径
	if r.TerminalConfig.ReplayStorage.TypeName != "null" {
		recordDirPath := filepath.Join(config.GlobalConfig.RecordPath,
			r.Created.Format(recordDirTimeFormat))
		conf.SetParameter(guacd.RecordingPath, recordDirPath)
		conf.SetParameter(guacd.CreateRecordingPath, BoolTrue)
		conf.SetParameter(guacd.RecordingName, r.SessionId)
	}

	// display 相关
	{
		for key, value := range RDPDisplay.GetDisplayParams() {
			conf.SetParameter(key, value)
		}
		for key, value := range RDPBuiltIn {
			conf.SetParameter(key, value)
		}
		// reconnect 会造成创建多个录像文件
		conf.SetParameter(guacd.RDPResizeMethod, "display-update")
	}

	// 设置 挂载目录 上传下载
	{
		drivePath := filepath.Join(config.GlobalConfig.DrivePath, r.User.ID)
		enableDrive := ConvertBoolToString(r.ActionsPerm.EnableDownload || r.ActionsPerm.EnableUpload)
		disableDownload := ConvertBoolToString(!r.ActionsPerm.EnableDownload)
		disableUpload := ConvertBoolToString(!r.ActionsPerm.EnableUpload)
		conf.SetParameter(guacd.RDPDrivePath, drivePath)
		conf.SetParameter(guacd.RDPCreateDrivePath, BoolTrue)
		conf.SetParameter(guacd.RDPEnableDrive, enableDrive)
		conf.SetParameter(guacd.RDPDriveName, "JumpServer")
		conf.SetParameter(guacd.RDPDisableDownload, disableDownload)
		conf.SetParameter(guacd.RDPDisableUpload, disableUpload)
	}

	// 粘贴复制
	{
		disableCopy := ConvertBoolToString(!r.ActionsPerm.EnableCopy)
		disablePaste := ConvertBoolToString(!r.ActionsPerm.EnablePaste)
		conf.SetParameter(guacd.DisableCopy, disableCopy)
		conf.SetParameter(guacd.DisablePaste, disablePaste)
	}

	// platform meta 数据 todo: 优化
	//{
	//	for k, v := range ConvertMetaToParams(r.Platform.MetaData) {
	//		conf.SetParameter(k, v)
	//	}
	//}

	return conf
}

type VNCConfiguration struct {
	SessionId      string
	Created        common.UTCTime
	User           *model.User
	Asset          *model.Asset          `json:"asset"`
	Account        *model.Account        `json:"system_user"`
	Platform       *model.Platform       `json:"platform"`
	TerminalConfig *model.TerminalConfig `json:"terminal_config"`
	ActionsPerm    *ActionPermission
}

const recordDirTimeFormat = "2006-01-02"

func (r VNCConfiguration) GetGuacdConfiguration() guacd.Configuration {
	conf := guacd.NewConfiguration()
	var (
		username string
		password string
		ip       string
		port     string
	)
	ip = r.Asset.Address
	port = strconv.Itoa(r.Asset.ProtocolPort("vnc"))
	username = r.Account.Username
	password = r.Account.Secret
	conf.Protocol = vnc
	conf.SetParameter(guacd.Hostname, ip)
	conf.SetParameter(guacd.Port, port)

	{
		conf.SetParameter(guacd.VNCUsername, username)
		conf.SetParameter(guacd.VNCPassword, password)
		conf.SetParameter(guacd.VNCAutoretry, "3")
	}
	// 设置存储
	replayCfg := r.TerminalConfig.ReplayStorage
	if replayCfg.TypeName != "null" {
		recordDirPath := filepath.Join(config.GlobalConfig.RecordPath, r.Created.Format(recordDirTimeFormat))
		conf.SetParameter(guacd.RecordingPath, recordDirPath)
		conf.SetParameter(guacd.CreateRecordingPath, BoolTrue)
		conf.SetParameter(guacd.RecordingName, r.SessionId)
	}
	{
		for key, value := range VNCDisplay.GetDisplayParams() {
			conf.SetParameter(key, value)
		}
	}

	// 粘贴复制
	{
		disableCopy := ConvertBoolToString(!r.ActionsPerm.EnableCopy)
		disablePaste := ConvertBoolToString(!r.ActionsPerm.EnablePaste)
		conf.SetParameter(guacd.DisableCopy, disableCopy)
		conf.SetParameter(guacd.DisablePaste, disablePaste)
	}
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

const (
	BoolFalse = "false"
	BoolTrue  = "true"
)

func ConvertBoolToString(b bool) string {
	if b {
		return BoolTrue
	}
	return BoolFalse
}
