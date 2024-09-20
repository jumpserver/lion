package model

import (
	"fmt"
	"strings"
)

type SecretInfo struct {
	CaCert     string `json:"ca_cert"`
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
}

type SpecInfo struct {
	// database
	DBName string `json:"db_name"`

	UseSSL           bool `json:"use_ssl"`
	AllowInvalidCert bool `json:"allow_invalid_cert"`

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
	ID         string       `json:"id"`
	Address    string       `json:"address"`
	Name       string       `json:"name"`
	OrgID      string       `json:"org_id"`
	Protocols  []Protocol   `json:"protocols"`
	SpecInfo   SpecInfo     `json:"spec_info"`
	SecretInfo SecretInfo   `json:"secret_info"`
	Platform   BasePlatform `json:"platform"`

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
	ID        string    `json:"id"`
	Name      string    `json:"Name"`
	Address   string    `json:"address"`
	Protocols Protocols `json:"protocols"`
	Account   Account   `json:"account"`
}

type Protocols []Protocol

func (p Protocols) GetProtocolPort(protocol string) int {
	for i := range p {
		if strings.EqualFold(p[i].Name, protocol) {
			return p[i].Port
		}
	}
	return 0
}
func (p Protocols) IsSupportProtocol(protocol string) bool {
	for _, item := range p {
		protocolName := strings.ToLower(item.Name)
		if protocolName == strings.ToLower(protocol) {
			return true
		}
	}
	return false
}

type Domain struct {
	ID       string    `json:"id"`
	Gateways []Gateway `json:"gateways"`
	Name     string    `json:"name"`
}
