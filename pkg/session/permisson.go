package session

import (
	"lion/pkg/config"
	"lion/pkg/jms-sdk-go/model"
)

func RemoteAppPermission() *model.Permission {
	actions := make([]string, 0, 4)
	globConfig := config.GlobalConfig
	if !globConfig.DisableAllUpDownload && globConfig.EnableRemoteAppUpDownLoad {
		actions = append(actions, model.ActionDownload)
		actions = append(actions, model.ActionUpload)
	}
	if !globConfig.DisableAllCopyPaste && globConfig.EnableRemoteAPPCopyPaste {
		actions = append(actions, model.ActionCopy)
		actions = append(actions, model.ActionPaste)
	}
	return &model.Permission{Actions: actions}
}
