package service

import (
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

func (s *JMService) GetPermission(userId, assetId, systemUserId string) (perms model.Permission, err error) {
	params := map[string]string{
		"user_id":        userId,
		"asset_id":       assetId,
		"system_user_id": systemUserId,
	}
	_, err = s.authClient.Get(PermissionURL, &perms, params)
	return
}

func (s *JMService) ValidateRemoteApp(userId, remoteAppId, systemUserId string) (bool, error) {
	params := map[string]string{
		"user_id":        userId,
		"application_id": remoteAppId,
		"system_user_id": systemUserId,
	}
	var res struct {
		Msg bool `json:"msg"`
	}
	_, err := s.authClient.Get(ValidateApplicationPermissionURL, &res, params)
	return res.Msg, err
}
