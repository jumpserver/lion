package session

import (
	"lion/pkg/config"
	"lion/pkg/jms-sdk-go/model"
)

type ActionPermission struct {
	EnableConnect bool `json:"enable_connect"`

	EnableCopy  bool `json:"enable_copy"`
	EnablePaste bool `json:"enable_paste"`

	EnableUpload   bool `json:"enable_upload"`
	EnableDownload bool `json:"enable_download"`
}

func NewActionPermission(perm *model.Permission, connectType string) *ActionPermission {
	action := ActionPermission{
		EnableConnect:  perm.EnableConnect(),
		EnableCopy:     perm.EnableCopy(),
		EnablePaste:    perm.EnablePaste(),
		EnableUpload:   perm.EnableUpload(),
		EnableDownload: perm.EnableDownload(),
	}
	globConfig := config.GlobalConfig
	switch connectType {
	case TypeRemoteApp:
		if globConfig.EnableRemoteAppUpDownLoad {
			action.EnableDownload = true
			action.EnableUpload = true
		}
		if globConfig.EnableRemoteAPPCopyPaste {
			action.EnablePaste = true
			action.EnableCopy = true
		}
	case TypeRDP, TypeVNC:
	}
	if globConfig.DisableAllUpDownload {
		action.EnableDownload = false
		action.EnableUpload = false
	}
	if globConfig.DisableAllCopyPaste {
		action.EnablePaste = false
		action.EnableCopy = false
	}
	return &action
}
