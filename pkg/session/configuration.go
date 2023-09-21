package session

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

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
		adDomain string
	)

	ip = r.Asset.Address
	port = strconv.Itoa(r.Asset.ProtocolPort(rdp))
	username = r.Account.Username
	password = r.Account.Secret

	conf := guacd.NewConfiguration()
	conf.Protocol = rdp
	conf.SetParameter(guacd.Hostname, ip)
	conf.SetParameter(guacd.Port, port)

	if r.Platform != nil {
		if rdpSetting, ok := r.Platform.GetProtocolSetting(rdp); ok {
			if rdpSetting.Setting.AdDomain != "" {
				adDomain = rdpSetting.Setting.AdDomain
			}
		}
	}
	/*
		AD Domain 的处理调整为
		1、如果账号 username 格式是 domain\username 则需要转换为 username@domain，且覆盖平台的 AD 域设置。
		2、其他格式的账号，如果平台中设置了 AD 域则使用平台中的设置，否则使用不设置
	*/

	parts := strings.Split(username, `\`)
	if len(parts) == 2 {
		username = fmt.Sprintf("%s@%s", parts[1], parts[0])
		adDomain = parts[0]
	}

	conf.SetParameter(guacd.RDPUsername, username)
	conf.SetParameter(guacd.RDPPassword, password)
	if adDomain != "" {
		conf.SetParameter(guacd.RDPDomain, adDomain)
	}

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

	// 平台中的设置
	rdpSecurityValue := SecurityAny
	if r.Platform != nil {
		if rdpSettings, ok := r.Platform.GetProtocolSetting(rdp); ok {
			if ValidateSecurityValue(rdpSettings.Setting.Security) {
				rdpSecurityValue = rdpSettings.Setting.Security
			}
			if rdpSettings.Setting.Console {
				conf.SetParameter(guacd.RDPConsole, BoolTrue)
			}
		}
	}
	conf.SetParameter(guacd.RDPSecurity, rdpSecurityValue)
	conf.SetParameter(guacd.RDPIgnoreCert, BoolTrue)

	// 设置客户端名称，任务管理器--用户---客户端名称显示
	conf.SetParameter(guacd.RDPClientName, "JumpServer-Lion")

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
