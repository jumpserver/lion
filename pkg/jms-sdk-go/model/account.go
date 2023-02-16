package model

import "fmt"

type BaseAccount struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Username   string     `json:"username"`
	Secret     string     `json:"secret"`
	SecretType LabelValue `json:"secret_type"`
}

func (a *BaseAccount) String() string {
	return fmt.Sprintf("%s(%s)", a.Name, a.Username)
}

func (a *BaseAccount) IsSSHKey() bool {
	return a.SecretType.Value == "ssh_key"
}

type Account struct {
	BaseAccount
	SuFrom *BaseAccount `json:"su_from"`
}
