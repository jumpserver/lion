package tunnel

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"lion/pkg/jms-sdk-go/model"

	"lion/pkg/guacd"
	"lion/pkg/logger"
)

type MonitorCon struct {
	Id          string
	guacdTunnel Tunneler

	ws *websocket.Conn

	wsLock    sync.Mutex
	guacdLock sync.Mutex

	Service *GuacamoleTunnelServer
	User    *model.User
}

func (m *MonitorCon) SendWsMessage(msg guacd.Instruction) error {
	return m.writeWsMessage([]byte(msg.String()))
}

func (m *MonitorCon) writeWsMessage(p []byte) error {
	m.wsLock.Lock()
	defer m.wsLock.Unlock()
	return m.ws.WriteMessage(websocket.TextMessage, p)
}

func (m *MonitorCon) WriteTunnelMessage(msg guacd.Instruction) (err error) {
	_, err = m.writeTunnelMessage([]byte(msg.String()))
	return err
}

func (m *MonitorCon) writeTunnelMessage(p []byte) (int, error) {
	m.guacdLock.Lock()
	defer m.guacdLock.Unlock()
	return m.guacdTunnel.WriteAndFlush(p)
}
func (m *MonitorCon) readTunnelInstruction() (*guacd.Instruction, error) {
	instruction, err := m.guacdTunnel.ReadInstruction()
	if err != nil {
		return nil, err
	}
	return &instruction, nil
}

func (m *MonitorCon) Run(ctx context.Context) (err error) {
	exit := make(chan error, 2)
	go func(t *MonitorCon) {
		for {
			instruction, err := t.readTunnelInstruction()
			if err != nil {
				_ = t.writeWsMessage([]byte(ErrDisconnect.String()))
				logger.Infof("Monitor[%s] guacd tunnel read err: %+v", t.Id, err)
				exit <- err
				break
			}
			logger.Debugf("Monitor[%s] guacd tunnel read: %s", t.Id, instruction.String())
			if err = t.writeWsMessage([]byte(instruction.String())); err != nil {
				logger.Error(err)
				exit <- err
				break
			}
		}
		_ = t.ws.Close()
	}(m)

	go func(t *MonitorCon) {
		for {
			_, message, err := t.ws.ReadMessage()
			if err != nil {
				logger.Infof("Monitor[%s] ws read err: %+v", t.Id, err)

				exit <- err
				break
			}
			logger.Debugf("Monitor[%s] ws read: %s", t.Id, message)

			if ret, err := guacd.ParseInstructionString(string(message)); err == nil {
				if ret.Opcode == INTERNALDATAOPCODE && len(ret.Args) >= 2 && ret.Args[0] == PINGOPCODE {
					if err := t.SendWsMessage(guacd.NewInstruction(INTERNALDATAOPCODE, PINGOPCODE)); err != nil {
						logger.Error(err)
					}
					continue
				}
			} else {
				logger.Errorf("Monitor[%s] parse instruction err %s", t.Id, err)
			}
			_, err = t.writeTunnelMessage(message)
			if err != nil {
				logger.Errorf("Monitor[%s] guacamole tunnel write err: %+v", t.Id, err)
				exit <- err
				break
			}
		}
		_ = t.guacdTunnel.Close()
	}(m)
	logObj := model.SessionLifecycleLog{User: m.User.String()}
	m.Service.RecordLifecycleLog(m.Id, model.AdminJoinMonitor, logObj)
	defer func() {
		m.Service.RecordLifecycleLog(m.Id, model.AdminExitMonitor, logObj)
	}()
	for {
		select {
		case err = <-exit:
			logger.Infof("Monitor[%s] exit: %+v", m.Id, err)
			return err
		case <-ctx.Done():
			logger.Info("Monitor[%s] done", m.Id)
			return nil
		}
	}
}
