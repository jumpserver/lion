package service

import (
	"lion/pkg/common"
	"lion/pkg/jms-sdk-go/model"
)

func (s *JMService) TerminalHeartBeat(sIds []string) (res []model.TerminalTask, err error) {
	data := model.HeartbeatData{
		SessionOnlineIds: sIds,
		CpuUsed:          common.CpuLoad1Usage(),
		MemoryUsed:       common.MemoryUsagePercent(),
		DiskUsed:         common.DiskUsagePercent(),
		SessionOnline:    len(sIds),
	}
	_, err = s.authClient.Post(TerminalHeartBeatURL, data, &res)
	return
}
