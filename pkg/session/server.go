package session

import (
	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
)

type Server struct {
	JmsService *service.JMService
}

func (s *Server) Creat(user model.User, assetId, systemUserId string) (Session, error) {
	asset, err := s.JmsService.GetAssetById(assetId)
	if err != nil {
		return Session{}, err
	}
	sysUser, err := s.JmsService.GetSystemUserById(systemUserId)
	if err != nil {
		return Session{}, err
	}
	newSession := Session{
		ID:         common.UUID(),
		Created:    common.NewNowUTCTime(),
		Asset:      asset,
		SystemUser: sysUser,
		User:       user,
	}
	return newSession, nil
}

func (s *Server) UpdateSession(sid string) {

}
