package model

import (
	"lion/pkg/common"
)

type FTPLog struct {
	ID         string         `json:"id"`
	User       string         `json:"user"`
	Hostname   string         `json:"asset"`
	OrgID      string         `json:"org_id"`
	Account    string         `json:"account"`
	RemoteAddr string         `json:"remote_addr"`
	Operate    string         `json:"operate"`
	Path       string         `json:"filename"`
	DateStart  common.UTCTime `json:"date_start"`
	IsSuccess  bool           `json:"is_success"`
}

const (
	OperateDownload = "download"
	OperateUpload   = "upload"
)
