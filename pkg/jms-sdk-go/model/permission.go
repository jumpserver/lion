package model

type Permission struct {
	Actions []string `json:"actions"`
}

const (
	ActionALL            = "all"
	ActionConnect        = "connect"
	ActionUpload         = "upload_file"
	ActionDownload       = "download_file"
	ActionUploadDownLoad = "updownload"
	ActionCopy           = "clipboard_copy"
	ActionPaste          = "clipboard_paste"
	ActionCopyPaste      = "clipboard_copy_paste"
)
