package model

import (
	"guacamole-client-go/pkg/common"
)

type Session struct {
	ID string `json:"id"`

	// "%s(%s)" Name Username
	User       string `json:"user"`
	Asset      string `json:"asset"`
	SystemUser string `json:"system_user"`

	RemoteAddr   string         `json:"remote_addr"`
	Protocol     string         `json:"protocol"`
	DateStart    common.UTCTime `json:"date_start"`
	DateEnd      common.UTCTime `json:"date_end"`
	OrgID        string         `json:"org_id"`
	UserID       string         `json:"user_id"`
	AssetID      string         `json:"asset_id"`
	SystemUserID string         `json:"system_user_id"`
}
