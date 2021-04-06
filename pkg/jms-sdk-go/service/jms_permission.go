package service

import (
	"fmt"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

func (s *JMService) GetPermission(userId, assetId, systemUserId string) (perms model.Permission, err error) {
	Url := fmt.Sprintf(PermissionURL, userId, assetId, systemUserId)
	_, err = s.authClient.Get(Url, &perms)
	return
}
