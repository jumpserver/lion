package service

import (
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

func (s *JMService) CreateFileOperationLog(data model.FTPLog) (err error) {
	_, err = s.authClient.Post(FTPLogListURL, data, nil)
	return
}
