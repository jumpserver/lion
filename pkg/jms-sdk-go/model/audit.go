package model

import (
	"lion/pkg/common"
)

type FTPLog struct {
	Id            string         `json:"id"`
	User          string         `json:"user"`
	Hostname      string         `json:"asset"`
	OrgID         string         `json:"org_id"`
	SystemUser    string         `json:"system_user"`
	RemoteAddr    string         `json:"remote_addr"`
	Operate       string         `json:"operate"`
	Path          string         `json:"filename"`
	DataStart     common.UTCTime `json:"data_start"`
	IsSuccess     bool           `json:"is_success"`
	HasFileRecord bool           `json:"has_file_record"`
}

const (
	OperateDownload = "Download"
	OperateUpload   = "Upload"
)
