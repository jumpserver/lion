package session

import (
	"errors"
	"fmt"

	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
)

const (
	TypeRDP       = "rdp"
	TypeVNC       = "vnc"
	TypeRemoteApp = "remoteapp"
)

var (
	ErrAPIService          = errors.New("connect API core err")
	ErrUnSupportedType     = errors.New("unsupported type")
	ErrUnSupportedProtocol = errors.New("unsupported protocol")
)

type Server struct {
	JmsService *service.JMService
}

func (s *Server) Creat(user *model.User, targetType, targetId, systemUserId string) (TunnelSession, error) {
	sysUser, err := s.JmsService.GetSystemUserById(systemUserId)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	switch sysUser.Protocol {
	case rdp, vnc:
	default:
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrUnSupportedProtocol, sysUser.Protocol)
	}
	switch targetType {
	case TypeRDP, TypeVNC:
		asset, err := s.JmsService.GetAssetById(targetId)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		sysUserAuth, err := s.JmsService.GetSystemUserAuthById(systemUserId, asset.ID)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		sysUser.Password = sysUserAuth.Password
		sysUser.PrivateKey = sysUserAuth.PrivateKey
		sysUser.Token = sysUserAuth.Token
		return s.CreateRDPAndVNCSession(user, &asset, &sysUser)

	case TypeRemoteApp:
		remoteApp, err := s.JmsService.GetRemoteApp(targetId)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		return s.CreateRemoteSession(user, &remoteApp, &sysUser)
	default:
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrUnSupportedType, targetType)
	}

}

func (s *Server) CreateRDPAndVNCSession(user *model.User, asset *model.Asset, systemUser *model.SystemUser) (TunnelSession, error) {
	platform, err := s.JmsService.GetAssetPlatform(asset.ID)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	permission, err := s.JmsService.GetPermission(user.ID, asset.ID, systemUser.ID)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	terminal, err := s.JmsService.GetTerminalConfig()
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	var (
		assetDomain *model.Domain
	)
	if asset.Domain != "" {
		domain, err := s.JmsService.GetDomainGateways(asset.Domain)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		assetDomain = &domain
	}

	newSession := TunnelSession{
		ID:             common.UUID(),
		Created:        common.NewNowUTCTime(),
		User:           user,
		Asset:          asset,
		SystemUser:     systemUser,
		Platform:       &platform,
		Domain:         assetDomain,
		Permission:     &permission,
		TerminalConfig: &terminal,
	}
	return newSession, nil
}

func (s *Server) CreateRemoteSession(user *model.User, remoteApp *model.RemoteAPP,
	systemUser *model.SystemUser) (TunnelSession, error) {
	asset, err := s.JmsService.GetAssetById(remoteApp.AssetId)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	sess, err := s.CreateRDPAndVNCSession(user, &asset, systemUser)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	sess.RemoteApp = remoteApp
	sess.Permission = RemoteAppPermission()
	return sess, nil
}
