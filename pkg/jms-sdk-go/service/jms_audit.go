package service

import (
	"fmt"

	"lion/pkg/jms-sdk-go/model"
)

func (s *JMService) CreateFileOperationLog(data model.FTPLog) (err error) {
	_, err = s.authClient.Post(FTPLogListURL, data, nil)
	return
}

func (s *JMService) PushSessionCommand(commands []*model.Command) (err error) {
	_, err = s.authClient.Post(SessionCommandURL, commands, nil)
	return
}

func (s *JMService) UploadFTPFile(fid, file string) error {
	var res map[string]interface{}
	url := fmt.Sprintf(FTPLogFileURL, fid)
	return s.authClient.PostFileWithFields(url, file, nil, &res)
}

func (s *JMService) FinishFTPFile(fid string) error {
	data := map[string]bool{"has_file": true}
	url := fmt.Sprintf(FTPLogUpdateURL, fid)
	_, err := s.authClient.Patch(url, data, nil)
	return err
}
