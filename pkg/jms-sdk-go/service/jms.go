package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"guacamole-client-go/pkg/jms-sdk-go/httplib"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

var AccessKeyUnauthorized = errors.New("access key unauthorized")

var ConnectErr = errors.New("api connect err")

const minTimeOut = time.Second * 30

func NewAuthJMService(opts ...Option) (*JMService, error) {
	opt := option{
		CoreHost: "http://127.0.0.1:8080",
		TimeOut:  time.Minute,
	}
	for _, setter := range opts {
		setter(&opt)
	}
	if opt.TimeOut < minTimeOut {
		opt.TimeOut = minTimeOut
	}
	httpClient, err := httplib.NewClient(opt.CoreHost, opt.TimeOut)
	if err != nil {
		return nil, err
	}
	if opt.sign != nil {
		httpClient.SetAuthSign(opt.sign)
	}
	return &JMService{authClient: httpClient}, nil
}

type JMService struct {
	authClient *httplib.Client
}

func (s *JMService) GetUserById(userID string) (user *model.User, err error) {
	url := fmt.Sprintf(UserDetailURL, userID)
	_, err = s.authClient.Get(url, &user)
	return
}

func (s *JMService) GetProfile() (user *model.User, err error) {
	var res *http.Response
	res, err = s.authClient.Get(UserProfileURL, &user)
	if res == nil && err != nil {
		return nil, fmt.Errorf("%w:%s", ConnectErr, err.Error())
	}
	if res != nil && res.StatusCode == http.StatusUnauthorized {
		return user, AccessKeyUnauthorized
	}
	return user, err
}

func (s *JMService) GetTerminalConfig() (model.TerminalConfig, error) {
	var conf model.TerminalConfig
	_, err := s.authClient.Get(TerminalConfigURL, &conf)
	return conf, err
}

func (s *JMService) Upload(sessionID, gZipFile string) error {
	var res map[string]interface{}
	Url := fmt.Sprintf(SessionReplayURL, sessionID)
	return s.authClient.UploadFile(Url, gZipFile, &res)
}

func (s *JMService) FinishReply(sid string) error {
	var res map[string]interface{}
	data := map[string]bool{"has_replay": true}
	Url := fmt.Sprintf(SessionDetailURL, sid)
	_, err := s.authClient.Patch(Url, data, &res)
	return err
}

func (s *JMService) GetAssetById(assetID string) (asset model.Asset, err error) {
	url := fmt.Sprintf(AssetDetailURL, assetID)
	_, err = s.authClient.Get(url, &asset)
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
