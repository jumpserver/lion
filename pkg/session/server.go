package session

import (
	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
)

type Server struct {
	JmsService *service.JMService
}

func (s *Server) Creat(user *model.User, assetId, systemUserId string) (ConnectSession, error) {
	asset, err := s.JmsService.GetAssetById(assetId)
	if err != nil {
		return ConnectSession{}, err
	}
	sysUser, err := s.JmsService.GetSystemUserById(systemUserId)
	if err != nil {
		return ConnectSession{}, err
	}
	platform, err := s.JmsService.GetAssetPlatform(asset)
	if err != nil {
		return ConnectSession{}, err
	}
	var (
		assetDomain *model.Domain
	)
	if asset.Domain != "" {
		domain, err := s.JmsService.GetDomainGateways(asset.Domain)
		if err != nil {
			return ConnectSession{}, err
		}
		assetDomain = &domain
	}

	newSession := ConnectSession{
		ID:         common.UUID(),
		Created:    common.NewNowUTCTime(),
		User:       user,
		Asset:      &asset,
		SystemUser: &sysUser,
		Platform:   &platform,
		Domain:     assetDomain,
	}
	return newSession, nil
}

func (s *Server) GetSession(sid string) ConnectSession {
	return ConnectSession{}
}

func (s *Server) UpdateSession(sid string) {

}

func (s *Server) ValidateConnectionPerms(session *ConnectSession) error {

	return nil
}
