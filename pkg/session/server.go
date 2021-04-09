package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
	"guacamole-client-go/pkg/storage"
)

const (
	TypeRDP       = "rdp"
	TypeVNC       = "vnc"
	TypeRemoteApp = "remoteapp"
)

const loginFrom = "WT"

var (
	ErrAPIService          = errors.New("connect API core err")
	ErrUnSupportedType     = errors.New("unsupported type")
	ErrUnSupportedProtocol = errors.New("unsupported protocol")
)

type Server struct {
	JmsService       *service.JMService
}

func (s *Server) CreatByToken(ctx *gin.Context, token string) (TunnelSession, error) {
	tokenUser, err := s.JmsService.GetTokenAsset(token)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	user, err := s.JmsService.GetUserById(tokenUser.UserID)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	targetType := TypeRDP
	targetId := tokenUser.AssetID
	systemUserId := tokenUser.SystemUserID
	return s.Creat(ctx, user, targetType, targetId, systemUserId)
}

func (s *Server) Creat(ctx *gin.Context, user *model.User, targetType, targetId, systemUserId string) (sess TunnelSession, err error) {
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
		sess, err = s.CreateRDPAndVNCSession(user, &asset, &sysUser)
		if err != nil {
			return TunnelSession{}, err
		}
	case TypeRemoteApp:
		remoteApp, err := s.JmsService.GetRemoteApp(targetId)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		sess, err = s.CreateRemoteSession(user, &remoteApp, &sysUser)
		if err != nil {
			return TunnelSession{}, err
		}
	default:
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrUnSupportedType, targetType)
	}
	jmsSession := model.Session{
		ID:           sess.ID,
		User:         sess.User.String(),
		Asset:        sess.Asset.Hostname,
		SystemUser:   sess.SystemUser.String(),
		LoginFrom:    loginFrom,
		RemoteAddr:   ctx.ClientIP(),
		Protocol:     sess.SystemUser.Protocol,
		DateStart:    sess.Created,
		OrgID:        sess.Asset.OrgID,
		UserID:       sess.User.ID,
		AssetID:      sess.Asset.ID,
		SystemUserID: sess.SystemUser.ID,
	}
	sess.ConnectedCallback = s.RegisterConnectedCallback(jmsSession)
	sess.ConnectedSuccessCallback = s.RegisterConnectedSuccessCallback(jmsSession)
	sess.ConnectedFailedCallback = s.RegisterConnectedFailedCallback(jmsSession)
	sess.DisConnectedCallback = s.RegisterDisConnectedCallback(jmsSession)
	sess.FinishReplayCallback = s.RegisterFinishReplayCallback(sess)
	return
}

func (s *Server) CreateRDPAndVNCSession(user *model.User, asset *model.Asset, systemUser *model.SystemUser) (TunnelSession, error) {
	platform, err := s.JmsService.GetAssetPlatform(asset.ID)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	sysUserAuth, err := s.JmsService.GetSystemUserAuthById(systemUser.ID, asset.ID)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	systemUser.Password = sysUserAuth.Password
	systemUser.PrivateKey = sysUserAuth.PrivateKey
	systemUser.Token = sysUserAuth.Token
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

func (s *Server) RegisterConnectedCallback(sess model.Session) func() error {
	return func() error {
		return s.JmsService.CreateSession(sess)
	}
}

func (s *Server) RegisterConnectedSuccessCallback(sess model.Session) func() error {
	return func() error {
		return s.JmsService.SessionSuccess(sess.ID)
	}
}

func (s *Server) RegisterConnectedFailedCallback(sess model.Session) func(err error) error {
	return func(err error) error {
		return s.JmsService.SessionFailed(sess.ID, err)
	}
}

func (s *Server) RegisterDisConnectedCallback(sess model.Session) func() error {
	return func() error {
		return s.JmsService.SessionDisconnect(sess.ID)
	}
}

const ReplayFileNameSuffix = ".replay.gz"

func (s *Server) RegisterFinishReplayCallback(tunnel TunnelSession) func() error {
	return func() error {
		replayConfig := tunnel.TerminalConfig.ReplayStorage
		if replayConfig["type"] == "null" {
			fmt.Println("ReplayStorage 为 null，无存储")
			return nil
		}
		recordDirPath := filepath.Join(config.GlobalConfig.RecordPath,
			tunnel.Created.Format(recordDirTimeFormat))
		originReplayFilePath := filepath.Join(recordDirPath, tunnel.ID)
		dstReplayFilePath := originReplayFilePath + ReplayFileNameSuffix
		fi, err := os.Stat(originReplayFilePath)
		if err != nil {
			return err
		}
		if fi.Size() < 1024 {
			fmt.Println("录像文件小于1024字节，可判断连接失败，未能产生有效的录像文件")
			_ = os.Remove(originReplayFilePath)
			return err
		}

		// 压缩文件
		err = common.CompressToGzipFile(originReplayFilePath, dstReplayFilePath)
		if err != nil {
			return err
		}
		// 压缩完成则删除源文件
		defer os.Remove(originReplayFilePath)
		if replayStorage := storage.NewReplayStorage(replayConfig); replayStorage != nil {
			err = replayStorage.Upload(dstReplayFilePath, tunnel.Created.Format(recordDirTimeFormat))
		} else {
			err = s.JmsService.Upload(tunnel.ID, dstReplayFilePath)
		}
		// 上传文件
		if err != nil {
			return err
		}
		// 上传成功，删除压缩文件
		defer os.Remove(dstReplayFilePath)
		// 通知core上传完成
		err = s.JmsService.FinishReply(tunnel.ID)
		return err
	}
}

func (s *Server) AuditFileOperation(fileLog model.FTPLog) error {
	return s.JmsService.CreateFileOperationLog(fileLog)
}

func ValidReplayDirname(dirname string) bool {
	_, err := time.Parse(recordDirTimeFormat, dirname)
	return err == nil
}
