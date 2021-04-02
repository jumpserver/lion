package service

import (
	"fmt"
)

func (s *JMService) Upload(sessionID, gZipFile string) error {
	var res map[string]interface{}
	Url := fmt.Sprintf(SessionReplayURL, sessionID)
	return s.authClient.UploadFile(Url, gZipFile, &res)
}

func (s *JMService) FinishReply(sid string) error {
	data := map[string]bool{"has_replay": true}
	Url := fmt.Sprintf(SessionDetailURL, sid)
	_, err := s.authClient.Patch(Url, data, nil)
	return err
}
