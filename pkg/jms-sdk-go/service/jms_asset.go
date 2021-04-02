package service

import (
	"fmt"

	"guacamole-client-go/pkg/jms-sdk-go/model"
)

func (s *JMService) GetAssetById(assetID string) (asset model.Asset, err error) {
	url := fmt.Sprintf(AssetDetailURL, assetID)
	_, err = s.authClient.Get(url, &asset)
	return
}

func (s *JMService) GetAssetPlatform(asset model.Asset) (platform model.Platform, err error) {
	url := fmt.Sprintf(AssetPlatFormURL, asset.ID)
	_, err = s.authClient.Get(url, &platform)
	return
}

func (s *JMService) GetDomainGateways(domainId string) (domain model.Domain, err error) {
	Url := fmt.Sprintf(DomainDetailWithGateways, domainId)
	_, err = s.authClient.Get(Url, &domain)
	return
}

func (s *JMService) GetSystemUserById(systemUserID string) (sysUser model.SystemUser, err error) {
	url := fmt.Sprintf(SystemUserDetailURL, systemUserID)
	_, err = s.authClient.Get(url, &sysUser)
	return
}

func (s *JMService) GetSystemUserAuthById(systemUserID, assetId string) (sysUser model.SystemUserAuthInfo, err error) {
	url := fmt.Sprintf(SystemUserAuthURL, systemUserID)
	if assetId != "" {
		url = fmt.Sprintf(SystemUserAssetAuthURL, systemUserID, assetId)
	}
	_, err = s.authClient.Get(url, &sysUser)
	return
}