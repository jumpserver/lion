package service

import (
	"lion/pkg/jms-sdk-go/model"
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

func (s *JMService) ValidateRemoteAppPermission(userId, remoteAppId, systemUserId string) (info model.ExpireInfo, err error) {
	params := map[string]string{
		"user_id":        userId,
		"application_id": remoteAppId,
		"system_user_id": systemUserId,
	}
	_, err = s.authClient.Get(ValidateApplicationPermissionURL, &info, params)
	return
}

const actionConnect = "connect"

func (s *JMService) ValidateAssetConnectPermission(userId, assetId, systemUserId string) (info model.ExpireInfo, err error) {
	params := map[string]string{
		"user_id":        userId,
		"asset_id":       assetId,
		"system_user_id": systemUserId,
		"action_name":    actionConnect,
	}
	_, err = s.authClient.Get(ValidateUserAssetPermissionURL, &info, params)
	return
}
