package service

import (
	"fmt"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

func (s *JMService) GetRemoteApp(remoteAppId string) (remoteApp model.RemoteAPP, err error) {
	Url := fmt.Sprintf(RemoteAPPURL, remoteAppId)
	_, err = s.authClient.Get(Url, &remoteApp)
	return
}
