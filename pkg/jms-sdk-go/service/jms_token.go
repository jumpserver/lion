package service

import (
	"fmt"

	"guacamole-client-go/pkg/jms-sdk-go/model"
)

func (s *JMService) GetTokenAsset(token string) (tokenUser model.TokenUser, err error) {
	Url := fmt.Sprintf(TokenAssetURL, token)
	_, err = s.authClient.Get(Url, &tokenUser)
	return
}
