package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"lion/pkg/common"
	"lion/pkg/config"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
	"lion/pkg/logger"
	"lion/pkg/storage"
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
	ErrPermissionDeny      = errors.New("permission deny")
)

type Server struct {
	JmsService *service.JMService
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
	return s.Create(ctx, user, targetType, targetId, systemUserId)
}

func (s *Server) Create(ctx *gin.Context, user *model.User, targetType, targetId, systemUserId string) (sess TunnelSession, err error) {
	sysUser, err := s.JmsService.GetSystemUserById(systemUserId)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	switch sysUser.Protocol {
	case rdp, vnc:
	default:
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrUnSupportedProtocol, sysUser.Protocol)
	}
	var (
		sessionAssetName string
		assetInfo        *model.Asset
		appInfo          *model.RemoteAPP

		systemUserAuth *model.SystemUserAuthInfo

		validatePermission func() (model.ExpireInfo, error)
	)
	switch targetType {
	case TypeRDP, TypeVNC:
		asset, err := s.JmsService.GetAssetById(targetId)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		validatePermission = func() (model.ExpireInfo, error) {
			return s.JmsService.ValidateAssetPermission(user.ID, asset.ID, sysUser.ID)
		}
		sysUserAuthInfo, err := s.JmsService.GetAssetSysUserAuthInfo(sysUser.ID, asset.ID, user.ID, user.Username)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		assetInfo = &asset
		systemUserAuth = &sysUserAuthInfo
		sessionAssetName = asset.String()
	case TypeRemoteApp:
		remoteApp, err := s.JmsService.GetRemoteApp(targetId)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		asset, err := s.JmsService.GetAssetById(remoteApp.AssetId)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		validatePermission = func() (model.ExpireInfo, error) {
			return s.JmsService.ValidateRemoteAppPermission(user.ID, remoteApp.ID, sysUser.ID)
		}
		sysUserAuthInfo, err := s.JmsService.GetApplicationSysUserAuthInfo(sysUser.ID, remoteApp.ID, user.ID, user.Username)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		appInfo = &remoteApp
		assetInfo = &asset
		systemUserAuth = &sysUserAuthInfo
		sessionAssetName = remoteApp.Name
	default:
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrUnSupportedType, targetType)
	}
	expireInfo, err := validatePermission()
	if err != nil {
		logger.Error(err)
	}
	if !expireInfo.EnableConnect() {
		return TunnelSession{}, fmt.Errorf("%w: connect deny", ErrPermissionDeny)
	}
	sess, err = s.CreateRDPAndVNCSession(user, assetInfo, &sysUser)
	if err != nil {
		return TunnelSession{}, err
	}
	sess.RemoteApp = appInfo
	sess.ExpireInfo = &expireInfo
	sess.Permission = &expireInfo.Permission
	sess.SystemUser = systemUserAuth
	sess.ActionPerm = NewActionPermission(&expireInfo.Permission, targetType)
	jmsSession := model.Session{
		ID:           sess.ID,
		User:         sess.User.String(),
		Asset:        sessionAssetName,
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
		Platform:       &platform,
		Domain:         assetDomain,
		TerminalConfig: &terminal,

		DisplaySystemUser: systemUser,
	}
	return newSession, nil
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
		storageType := replayConfig["TYPE"]
		if storageType == "null" {
			logger.Error("录像存储设置为 null，无存储")
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
			logger.Error("录像文件小于1024字节，可判断连接失败，未能产生有效的录像文件")
			_ = os.Remove(originReplayFilePath)
			return s.JmsService.SessionFailed(tunnel.ID, err)
		}

		// 压缩文件
		err = common.CompressToGzipFile(originReplayFilePath, dstReplayFilePath)
		if err != nil {
			logger.Error("压缩文件失败: ", err)
			return err
		}
		// 压缩完成则删除源文件
		defer os.Remove(originReplayFilePath)
		logger.Infof("Upload record file: %s, type: %s", dstReplayFilePath, storageType)
		if replayStorage := storage.NewReplayStorage(replayConfig); replayStorage != nil {
			targetName := strings.Join([]string{tunnel.Created.Format(recordDirTimeFormat),
				tunnel.ID + ReplayFileNameSuffix}, "/")
			err = replayStorage.Upload(dstReplayFilePath, targetName)
		} else {
			err = s.JmsService.Upload(tunnel.ID, dstReplayFilePath)
		}
		// 上传文件
		if err != nil {
			logger.Error("Upload replay failed: ", err.Error())
			return err
		}
		// 上传成功，删除压缩文件
		defer os.Remove(dstReplayFilePath)
		// 通知core上传完成
		err = s.JmsService.FinishReply(tunnel.ID)
		return err
	}
}

func (s *Server) GetFilterParser(tunnel *TunnelSession) ParseEngine {
	winParser := Parser{
		id:         tunnel.ID,
		jmsService: s.JmsService,
	}
	winParser.initial()
	return &winParser
}

func (s *Server) GetCommandRecorder(tunnel *TunnelSession) *CommandRecorder {
	cmdR := CommandRecorder{
		sessionID:  tunnel.ID,
		storage:    storage.NewCommandStorage(s.JmsService, tunnel.TerminalConfig),
		queue:      make(chan *model.Command, 10),
		closed:     make(chan struct{}),
		jmsService: s.JmsService,
	}
	go cmdR.record()
	return &cmdR
}

func (s *Server) GenerateCommandItem(tunnel *TunnelSession, user, input, output string,
	riskLevel int64, createdDate time.Time) *model.Command {
	var (
		server string
	)
	if tunnel.RemoteApp != nil {
		server = tunnel.RemoteApp.Name
	} else {
		server = tunnel.Asset.String()
	}
	return &model.Command{
		SessionID:   tunnel.ID,
		OrgID:       tunnel.SystemUser.OrgId,
		Server:      server,
		User:        user,
		SystemUser:  tunnel.SystemUser.String(),
		Input:       input,
		Output:      output,
		Timestamp:   createdDate.Unix(),
		RiskLevel:   riskLevel,
		Protocol:    tunnel.SystemUser.Protocol,
		DateCreated: createdDate.UTC(),
	}
}

func (s *Server) AuditFileOperation(fileLog model.FTPLog) {
	if err := s.JmsService.CreateFileOperationLog(fileLog); err != nil {
		logger.Errorf("Audit file operation err: %s", err)
	}
}

func ValidReplayDirname(dirname string) bool {
	_, err := time.Parse(recordDirTimeFormat, dirname)
	return err == nil
}
