package service

import (
	"fmt"

	"lion/pkg/jms-sdk-go/model"
)

func (s *JMService) GetTokenAsset(token string) (tokenUser model.TokenUser, err error) {
	Url := fmt.Sprintf(TokenAssetURL, token)
	_, err = s.authClient.Get(Url, &tokenUser)
	return
}

func (s *JMService) GetConnectTokenInfo(tokenId string) (resp model.ConnectToken, err error) {
	data := map[string]interface{}{
		"id":         tokenId,
		"expire_now": false,
	}
	_, err = s.authClient.Post(SuperConnectTokenSecretURL, data, &resp)
	return
}

func (s *JMService) GetConnectTokenAppletOption(tokenId string) (resp model.AppletOption, err error) {
	data := map[string]string{
		"id": tokenId,
	}
	_, err = s.authClient.Post(SuperConnectTokenAppletOptionURL, data, &resp)
	return
}

func (s *JMService) ReleaseAppletAccount(accountId string) (err error) {
	data := map[string]string{
		"id": accountId,
	}
	_, err = s.authClient.Delete(SuperConnectAppletHostAccountReleaseURL, data)
	return
}
