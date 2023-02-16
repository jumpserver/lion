package model

import (
	"lion/pkg/common"
)

type Session struct {
	ID string `json:"id"`
	// "%s(%s)" Name Username
	User       string         `json:"user"`
	Asset      string         `json:"asset"`
	Account    string         `json:"account"`
	LoginFrom  string         `json:"login_from"`
	RemoteAddr string         `json:"remote_addr"`
	Protocol   string         `json:"protocol"`
	DateStart  common.UTCTime `json:"date_start"`
	OrgID      string         `json:"org_id"`
	UserID     string         `json:"user_id"`
	AssetID    string         `json:"asset_id"`
	AccountID  string         `json:"account_id"`
}
