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

func (s *JMService) GetConnectTokenInfo(tokenId string, expireNow bool) (resp model.ConnectToken, err error) {
	data := map[string]interface{}{
		"id":         tokenId,
		"expire_now": expireNow,
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
	_, err = s.authClient.Post(SuperConnectAppletHostAccountReleaseURL, data, nil)
	return
}

func (s *JMService) GetConnectTokenVirtualAppOption(tokenId string) (resp model.VirtualApp, err error) {
	data := map[string]string{
		"id": tokenId,
	}
	_, err = s.authClient.Post(SuperConnectTokenVirtualAppOptionURL, data, &resp)
	return
}

func (s *JMService) CheckTokenStatus(tokenId string) (res model.TokenCheckStatus, err error) {
	reqURL := fmt.Sprintf(SuperConnectTokenCheckURL, tokenId)
	_, err = s.authClient.Get(reqURL, &res)
	return
}
