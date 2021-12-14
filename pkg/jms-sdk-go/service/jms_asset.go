package service

import (
	"fmt"

	"lion/pkg/jms-sdk-go/model"
)

func (s *JMService) GetAssetById(assetId string) (asset model.Asset, err error) {
	url := fmt.Sprintf(AssetDetailURL, assetId)
	_, err = s.authClient.Get(url, &asset)
	return
}

func (s *JMService) GetAssetPlatform(assetId string) (platform model.Platform, err error) {
	url := fmt.Sprintf(AssetPlatFormURL, assetId)
	_, err = s.authClient.Get(url, &platform)
	return
}

func (s *JMService) GetDomainGateways(domainId string) (domain model.Domain, err error) {
	Url := fmt.Sprintf(DomainDetailWithGateways, domainId)
	_, err = s.authClient.Get(Url, &domain)
	return
}

func (s *JMService) GetSystemUserById(systemUserId string) (sysUser model.SystemUser, err error) {
	url := fmt.Sprintf(SystemUserDetailURL, systemUserId)
	_, err = s.authClient.Get(url, &sysUser)
	return
}

func (s *JMService) GetApplicationSysUserAuthInfo(sysId, appId, userId,
	username string) (info model.SystemUserAuthInfo, err error) {
	reqURL := fmt.Sprintf(SystemUserAppAuthURL, sysId, appId)
	return s.getSysUserAuthInfo(reqURL, userId, username)
}

func (s *JMService) GetAssetSysUserAuthInfo(sysId, assetId, userId,
	username string) (info model.SystemUserAuthInfo, err error) {
	reqURL := fmt.Sprintf(SystemUserAssetAuthURL, sysId, assetId)
	return s.getSysUserAuthInfo(reqURL, userId, username)
}

func (s *JMService) getSysUserAuthInfo(authURL, userId,
	username string) (info model.SystemUserAuthInfo, err error) {
	params := make(map[string]string)
	if username != "" {
		params["username"] = username
	}
	if userId != "" {
		params["user_id"] = userId
	}
	_, err = s.authClient.Get(authURL, &info, params)
	return
}
