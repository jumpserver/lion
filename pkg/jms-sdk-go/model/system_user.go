package model

import (
	"fmt"
)

const LoginModeManual = "manual"

const (
	AllAction      = "all"
	ConnectAction  = "connect"
	UploadAction   = "upload_file"
	DownloadAction = "download_file"
)

type SystemUser struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Username             string   `json:"username"`
	Priority             int      `json:"priority"`
	Protocol             string   `json:"protocol"`
	AdDomain             string   `json:"ad_domain"`
	Comment              string   `json:"comment"`
	LoginMode            string   `json:"login_mode"`
	Password             string   `json:"-"`
	PrivateKey           string   `json:"-"`
	Actions              []string `json:"actions"`
	SftpRoot             string   `json:"sftp_root"`
	OrgId                string   `json:"org_id"`
	OrgName              string   `json:"org_name"`
	UsernameSameWithUser bool     `json:"username_same_with_user"`
	Token                string   `json:"-"`
}

func (s *SystemUser) String() string {
	return fmt.Sprintf("%s(%s)", s.Name, s.Username)
}

type SystemUserAuthInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Username   string `json:"username"`
	Protocol   string `json:"protocol"`
	LoginMode  string `json:"login_mode"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	AdDomain   string `json:"ad_domain"`
	Token      string `json:"token"`
	OrgId      string `json:"org_id"`
	OrgName    string `json:"org_name"`
	PublicKey  string `json:"public_key"`

	UsernameSameWithUser bool `json:"username_same_with_user"`
}

func (s *SystemUserAuthInfo) String() string {
	return fmt.Sprintf("%s(%s)", s.Name, s.Username)
}
