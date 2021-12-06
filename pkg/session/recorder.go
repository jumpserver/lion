package session

import (
	"time"

	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
	"lion/pkg/logger"
	"lion/pkg/storage"
)

type CommandRecorder struct {
	sessionID string
	storage   storage.CommandStorage

	queue  chan *model.Command
	closed chan struct{}

	jmsService *service.JMService
}

func (c *CommandRecorder) Record(command *model.Command) {
	c.queue <- command
}

func (c *CommandRecorder) End() {
	select {
	case <-c.closed:
		return
	default:
	}
	close(c.closed)
}

func (c *CommandRecorder) record() {
	cmdList := make([]*model.Command, 0, 10)
	notificationList := make([]*model.Command, 0, 10)
	maxRetry := 0
	logger.Infof("Session %s: Command recorder start", c.sessionID)
	defer logger.Infof("Session %s: Command recorder close", c.sessionID)
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()
	for {
		select {
		case <-c.closed:
			if len(cmdList) == 0 {
				return
			}
		case p, ok := <-c.queue:
			if !ok {
				return
			}
			if p.RiskLevel == model.DangerLevel {
				notificationList = append(notificationList, p)
			}
			cmdList = append(cmdList, p)
			if len(cmdList) < 5 {
				continue
			}
		case <-tick.C:
			if len(cmdList) == 0 {
				continue
			}
		}
		err := c.storage.BulkSave(cmdList)
		if err == nil {
			cmdList = cmdList[:0]
			maxRetry = 0
			continue
		}
		if err != nil {
			logger.Errorf("Session %s: command bulk save err: %s", c.sessionID, err)
		}

		if maxRetry > 5 {
			cmdList = cmdList[1:]
		}
		maxRetry++
	}
}
