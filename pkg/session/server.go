package session

import (
	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
)

type Server struct {
	JmsService *service.JMService
}

func (s *Server) Creat(user *model.User, assetId, systemUserId string) (Session, error) {
	asset, err := s.JmsService.GetAssetById(assetId)
	if err != nil {
		return Session{}, err
	}
	sysUser, err := s.JmsService.GetSystemUserById(systemUserId)
	if err != nil {
		return Session{}, err
	}
	platform, err := s.JmsService.GetAssetPlatform(asset)
	if err != nil {
		return Session{}, err
	}
	var (
		assetDomain *model.Domain
	)
	if asset.Domain != "" {
		domain, err := s.JmsService.GetDomainGateways(asset.Domain)
		if err != nil {
			return Session{}, err
		}
		assetDomain = &domain
	}

	newSession := Session{
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

func (s *Server) GetSession(sid string) Session {
	return Session{}
}

func (s *Server) UpdateSession(sid string) {

}

func (s *Server) ValidateConnectionPerms(session *Session) error {

	return nil
}
