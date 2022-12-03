package model

import (
	"fmt"
	"strings"
)

type Specific struct {
	// database
	DBName string `json:"db_name"`

	UseSSL           bool   `json:"use_ssl"`
	CaCert           string `json:"ca_cert"`
	ClientCert       string `json:"client_cert"`
	CertKey          string `json:"cert_key"`
	AllowInvalidCert bool   `json:"allow_invalid_cert"`

	// web
	Autofill         string `json:"autofill"`
	UsernameSelector string `json:"username_selector"`
	PasswordSelector string `json:"password_selector"`
	SubmitSelector   string `json:"submit_selector"`
}

type BasePlatform struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Protocol struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Port int    `json:"port"`
}

type Asset struct {
	ID        string       `json:"id"`
	Address   string       `json:"address"`
	Name      string       `json:"name"`
	OrgID     string       `json:"org_id"`
	Protocols []Protocol   `json:"protocols"`
	Specific  Specific     `json:"specific"`
	Platform  BasePlatform `json:"platform"`

	Domain   string `json:"domain"` // 是否需要走网域
	Comment  string `json:"comment"`
	OrgName  string `json:"org_name"`
	IsActive bool   `json:"is_active"` // 判断资产是否禁用
}

func (a *Asset) String() string {
	return fmt.Sprintf("%s(%s)", a.Name, a.Address)
}
func (a *Asset) ProtocolPort(protocol string) int {
	for _, item := range a.Protocols {
		protocolName := strings.ToLower(item.Name)
		protocol = strings.ToLower(protocol)
		if protocolName == protocol {
			return item.Port
		}
	}
	return 0
}

func (a *Asset) IsSupportProtocol(protocol string) bool {
	for _, item := range a.Protocols {
		protocolName := strings.ToLower(item.Name)
		protocol = strings.ToLower(protocol)
		if protocolName == protocol {
			return true
		}
	}
	return false
}

type Gateway struct {
	ID         string `json:"id"`
	Name       string `json:"Name"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}

type Domain struct {
	ID       string    `json:"id"`
	Gateways []Gateway `json:"gateways"`
	Name     string    `json:"name"`
}
