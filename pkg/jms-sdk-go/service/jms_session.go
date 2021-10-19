package service

import (
	"fmt"
	"lion/pkg/common"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/logger"
)

func (s *JMService) Upload(sessionID, gZipFile string) error {
	var res map[string]interface{}
	Url := fmt.Sprintf(SessionReplayURL, sessionID)
	return s.authClient.UploadFile(Url, gZipFile, &res)
}

func (s *JMService) UploadFTPFile(FTPLogId, gZipFile string) error {
	var res map[string]interface{}
	Url := fmt.Sprintf(FTPLogFileURL, FTPLogId)
	return s.authClient.UploadFile(Url, gZipFile, &res)
}

func (s *JMService) FinishFTPLogFileUpload(sid string) bool {
	var res map[string]interface{}
	data := map[string]bool{"has_file_record": true}
	Url := fmt.Sprintf(FTPLogUpdateURL, sid)
	_, err := s.authClient.Patch(Url, data, &res)
	if err != nil {
		logger.Error(err)
		return false
	}
	return true
}

func (s *JMService) FinishReply(sid string) error {
	data := map[string]bool{"has_replay": true}
	return s.sessionPatch(sid, data)
}

func (s *JMService) CreateSession(sess model.Session) error {
	_, err := s.authClient.Post(SessionListURL, sess, nil)
	return err
}

func (s *JMService) SessionSuccess(sid string) error {
	data := map[string]bool{
		"is_success": true,
	}
	return s.sessionPatch(sid, data)
}

func (s *JMService) SessionFailed(sid string, err error) error {
	data := map[string]bool{
		"is_success": false,
	}
	return s.sessionPatch(sid, data)
}
func (s *JMService) SessionDisconnect(sid string) error {
	data := map[string]interface{}{
		"is_finished": true,
		"date_end":    common.NewNowUTCTime(),
	}
	return s.sessionPatch(sid, data)
}

func (s *JMService) SessionFinished(sid string, time common.UTCTime) error {
	data := map[string]interface{}{
		"is_finished": true,
		"date_end":    time,
	}
	return s.sessionPatch(sid, data)
}

func (s *JMService) sessionPatch(sid string, data interface{}) error {
	Url := fmt.Sprintf(SessionDetailURL, sid)
	_, err := s.authClient.Patch(Url, data, nil)
	return err
}
