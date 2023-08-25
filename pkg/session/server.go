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
	"lion/pkg/guacd"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
	"lion/pkg/jms-sdk-go/service/videoworker"
	"lion/pkg/logger"
	"lion/pkg/storage"
)

const (
	TypeRDP       = "rdp"
	TypeVNC       = "vnc"
	TypeRemoteApp = "remoteapp"

	connectApplet = "applet"
)

const loginFrom = "WT"

var (
	ErrAPIService = errors.New("connect API core err")
	//ErrUnSupportedType     = errors.New("unsupported type")

	ErrUnSupportedProtocol = errors.New("unsupported protocol")
	ErrPermissionDeny      = errors.New("permission deny")
)

type Server struct {
	JmsService *service.JMService

	VideoWorkerClient *videoworker.Client
}

func (s *Server) CreatByToken(ctx *gin.Context, token string) (TunnelSession, error) {
	connectToken, err := s.JmsService.GetConnectTokenInfo(token)
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	cfg, err := s.JmsService.GetTerminalConfig()
	if err != nil {
		return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
	}
	if !connectToken.Actions.EnableConnect() {
		return TunnelSession{}, ErrPermissionDeny
	}
	opts := make([]TunnelOption, 0, 10)
	opts = append(opts, ConnectTokenAuthInfo(&connectToken))
	opts = append(opts, WithProtocol(connectToken.Protocol))
	opts = append(opts, WithUser(&connectToken.User))
	opts = append(opts, WithActions(connectToken.Actions))
	opts = append(opts, WithExpireInfo(connectToken.ExpireAt))
	opts = append(opts, WithAsset(&connectToken.Asset))
	opts = append(opts, WithAccount(&connectToken.Account))
	opts = append(opts, WithPlatform(&connectToken.Platform))
	opts = append(opts, WithGateway(connectToken.Gateway))
	opts = append(opts, WithTerminalConfig(&cfg))
	if connectToken.ConnectMethod.Type == connectApplet {
		appletOptions, err := s.JmsService.GetConnectTokenAppletOption(token)
		if err != nil {
			return TunnelSession{}, fmt.Errorf("%w: %s", ErrAPIService, err.Error())
		}
		appletOpt := &appletOptions
		opts = append(opts, WithAppletOption(appletOpt))
		logger.Infof("Connect applet(%s) use host(%s) account (%s)", connectToken.Asset.String(),
			appletOpt.Host.String(), appletOpt.Account.String())
		// 连接发布机，需要使用发布机的网关
		opts = append(opts, WithGateway(appletOptions.Gateway))
	}
	return s.Create(ctx, opts...)
}

func ConnectTokenAuthInfo(authInfo *model.ConnectToken) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.authInfo = authInfo
	}
}
func WithActions(actions model.Actions) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.Actions = actions
	}
}
func WithExpireInfo(expireInfo model.ExpireInfo) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.ExpireInfo = expireInfo
	}
}

func WithProtocol(protocol string) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.Protocol = protocol
	}
}

func WithAsset(asset *model.Asset) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.Asset = asset
	}
}

func WithAccount(account *model.Account) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.Account = account
	}
}

func WithPlatform(platform *model.Platform) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.Platform = platform
	}
}

func WithGateway(gateway *model.Gateway) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.Gateway = gateway
	}
}

func WithTerminalConfig(cfg *model.TerminalConfig) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.TerminalConfig = cfg
	}
}

func WithAppletOption(appletOpt *model.AppletOption) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.appletOpt = appletOpt
	}
}

func WithUser(user *model.User) TunnelOption {
	return func(tunnel *tunnelOption) {
		tunnel.User = user
	}
}

type tunnelOption struct {
	Protocol   string
	User       *model.User
	Asset      *model.Asset
	Account    *model.Account
	Platform   *model.Platform
	Domain     *model.Domain
	Gateway    *model.Gateway
	Actions    model.Actions
	ExpireInfo model.ExpireInfo

	authInfo       *model.ConnectToken
	TerminalConfig *model.TerminalConfig
	appletOpt      *model.AppletOption
}

type TunnelOption func(*tunnelOption)

func (s *Server) Create(ctx *gin.Context, opts ...TunnelOption) (sess TunnelSession, err error) {
	opt := &tunnelOption{}
	for _, setter := range opts {
		setter(opt)
	}
	targetType := TypeRDP
	sessionProtocol := opt.Protocol
	if opt.authInfo.ConnectMethod.Type == connectApplet {
		targetType = TypeRemoteApp
	} else {
		switch opt.Protocol {
		case TypeRDP:
			targetType = TypeRDP
		case TypeVNC:
			targetType = TypeVNC
		default:
			if opt.appletOpt == nil {
				return TunnelSession{}, fmt.Errorf("%w: %s", ErrUnSupportedProtocol, opt.Protocol)
			}
			targetType = TypeRemoteApp
		}
	}
	sessionAssetName := opt.Asset.String()
	sess, err = s.CreateRDPAndVNCSession(opt)
	if err != nil {
		return TunnelSession{}, err
	}
	perm := opt.Actions.Permission()
	sess.AppletOpts = opt.appletOpt
	sess.AuthInfo = opt.authInfo
	if opt.appletOpt != nil {
		sess.RemoteApp = &opt.appletOpt.Applet
	}
	sess.User = opt.User
	sess.ExpireInfo = opt.ExpireInfo
	sess.Permission = &perm
	sess.Account = opt.Account
	sess.ActionPerm = NewActionPermission(&perm, targetType)
	jmsSession := model.Session{
		ID:         sess.ID,
		User:       sess.User.String(),
		Asset:      sessionAssetName,
		Account:    sess.Account.String(),
		LoginFrom:  loginFrom,
		RemoteAddr: ctx.ClientIP(),
		Protocol:   sessionProtocol,
		DateStart:  sess.Created,
		OrgID:      sess.Asset.OrgID,
		UserID:     sess.User.ID,
		AssetID:    sess.Asset.ID,
		AccountID:  opt.Account.ID,
	}
	sess.ConnectedCallback = s.RegisterConnectedCallback(jmsSession)
	sess.ConnectedSuccessCallback = s.RegisterConnectedSuccessCallback(jmsSession)
	sess.ConnectedFailedCallback = s.RegisterConnectedFailedCallback(jmsSession)
	sess.DisConnectedCallback = s.RegisterDisConnectedCallback(jmsSession)
	sess.FinishReplayCallback = s.RegisterFinishReplayCallback(sess)
	sess.ReleaseAppletAccount = func() error {
		if opt.appletOpt == nil {
			return nil
		}
		return s.JmsService.ReleaseAppletAccount(opt.appletOpt.ID)
	}
	return
}

func (s *Server) CreateRDPAndVNCSession(opt *tunnelOption) (TunnelSession, error) {
	account := opt.Account
	newSession := TunnelSession{
		ID:             common.UUID(),
		Protocol:       opt.Protocol,
		Created:        common.NewNowUTCTime(),
		User:           opt.User,
		Asset:          opt.Asset,
		Platform:       opt.Platform,
		TerminalConfig: opt.TerminalConfig,
		Gateway:        opt.Gateway,

		DisplayAccount: &model.Account{
			BaseAccount: model.BaseAccount{
				Name:       account.Name,
				Username:   account.Username,
				Secret:     "",
				SecretType: account.SecretType}},
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

func (s *Server) UploadReplayToVideoWorker(tunnel TunnelSession, info guacd.ClientInformation) bool {
	recordDirPath := filepath.Join(config.GlobalConfig.RecordPath,
		tunnel.Created.Format(recordDirTimeFormat))
	originReplayFilePath := filepath.Join(recordDirPath, tunnel.ID)
	task, err := s.VideoWorkerClient.CreateReplayTask(tunnel.ID, originReplayFilePath,
		videoworker.ReplayMeta{
			SessionId:     tunnel.ID,
			ComponentType: "lion",
			FileType:      ".replay",
			SessionDate:   tunnel.Created.Format("2006-01-02"),
			Width:         info.OptimalScreenWidth,
			Height:        info.OptimalScreenHeight,
			Bitrate:       1,
		})
	if err != nil {
		logger.Errorf("video worker create replay task failed: %s", err)
		return false
	}
	logger.Infof("video worker create task success: %+v", task)
	return true
}

func (s *Server) RegisterFinishReplayCallback(tunnel TunnelSession) func(guacd.ClientInformation) error {
	return func(info guacd.ClientInformation) error {
		replayConfig := tunnel.TerminalConfig.ReplayStorage
		storageType := replayConfig.TypeName
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
		if s.VideoWorkerClient != nil && s.UploadReplayToVideoWorker(tunnel, info) {
			logger.Infof("Upload replay file to video worker: %s", originReplayFilePath)
			_ = os.Remove(originReplayFilePath)
			return nil
		}

		// 压缩文件
		err = common.CompressToGzipFile(originReplayFilePath, dstReplayFilePath)
		if err != nil {
			logger.Error("压缩文件失败: ", err)
			return err
		}
		// 压缩完成则删除源文件
		defer os.Remove(originReplayFilePath)
		defaultStorage := storage.ServerStorage{StorageType: "server", JmsService: s.JmsService}
		logger.Infof("Upload record file: %s, type: %s", dstReplayFilePath, storageType)
		if replayStorage := storage.NewReplayStorage(s.JmsService, replayConfig); replayStorage != nil {
			targetName := strings.Join([]string{tunnel.Created.Format(recordDirTimeFormat),
				tunnel.ID + ReplayFileNameSuffix}, "/")
			if err = replayStorage.Upload(dstReplayFilePath, targetName); err != nil {
				logger.Errorf("Upload replay failed: %s", err)
				logger.Errorf("Upload replay by type %s failed, try use default", storageType)
				err = defaultStorage.Upload(tunnel.ID, dstReplayFilePath)
			}
		} else {
			err = defaultStorage.Upload(tunnel.ID, dstReplayFilePath)
		}
		// 上传文件
		if err != nil {
			logger.Errorf("Upload replay failed: %s", err.Error())
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
	server := tunnel.Asset.String()
	return &model.Command{
		SessionID:   tunnel.ID,
		OrgID:       tunnel.Asset.OrgID,
		Server:      server,
		User:        user,
		Account:     tunnel.Account.String(),
		Input:       input,
		Output:      output,
		Timestamp:   createdDate.Unix(),
		RiskLevel:   riskLevel,
		Protocol:    tunnel.Protocol,
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
