package model

import "fmt"

type BaseAccount struct {
	Name       string `json:"name"`
	Username   string `json:"username"`
	Secret     string `json:"secret"`
	SecretType string `json:"secret_type"`
}

func (a *BaseAccount) String() string {
	return fmt.Sprintf("%s(%s)", a.Name, a.Username)
}

type Account struct {
	BaseAccount
	SuFrom *BaseAccount `json:"su_from"`
}
