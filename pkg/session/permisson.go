package session

import (
	"lion/pkg/config"

	"github.com/jumpserver-dev/sdk-go/model"
)

type ActionPermission struct {
	EnableConnect bool `json:"enable_connect"`

	EnableCopy  bool `json:"enable_copy"`
	EnablePaste bool `json:"enable_paste"`

	EnableUpload   bool `json:"enable_upload"`
	EnableDownload bool `json:"enable_download"`
	EnableShare    bool `json:"enable_share"`

	ClipboardPolicy model.ClipboardPolicy `json:"clipboard_policy"`
}

func NewActionPermission(perm *model.Permission, connectType string, connectOptions model.ConnectOptions) *ActionPermission {
	action := ActionPermission{
		EnableConnect:  perm.EnableConnect(),
		EnableCopy:     perm.EnableCopy(),
		EnablePaste:    perm.EnablePaste(),
		EnableUpload:   perm.EnableUpload(),
		EnableDownload: perm.EnableDownload(),
		EnableShare:    perm.EnableShare(),
		ClipboardPolicy: model.ClipboardPolicy{
			FileUpload:   true,
			FileDownload: true,
			TextCopy:     true,
			TextPaste:    true,
		},
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
	action.applyClipboardPolicy(connectOptions)
	return &action
}

func (a *ActionPermission) applyClipboardPolicy(connectOptions model.ConnectOptions) {
	if connectOptions.ClipboardPolicy == nil {
		return
	}
	policy := *connectOptions.ClipboardPolicy
	a.ClipboardPolicy = policy
	// SFTP drive transfer.
	a.EnableUpload = a.EnableUpload && policy.FileUpload
	a.EnableDownload = a.EnableDownload && policy.FileDownload
	// Clipboard channel (text + file share one guacd channel per direction), so
	// keep it open when either text or file transfer is allowed in that direction.
	// Per-stream text/file enforcement is done by the clipboard policy filter.
	// copy: server -> client (text via TextCopy, file via FileDownload)
	a.EnableCopy = a.EnableCopy && (policy.TextCopy || policy.FileDownload)
	// paste: client -> server (text via TextPaste, file via FileUpload)
	a.EnablePaste = a.EnablePaste && (policy.TextPaste || policy.FileUpload)
}
