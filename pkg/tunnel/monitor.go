package tunnel

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"

	"lion/pkg/guacd"
	"lion/pkg/logger"
)

type MonitorCon struct {
	guacdTunnel Tunneler

	ws *websocket.Conn

	wsLock    sync.Mutex
	guacdLock sync.Mutex
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
				logger.Error(err)
				exit <- err
				break
			}
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
				exit <- err
				break
			}

			if ret, err := guacd.ParseInstructionString(string(message)); err == nil {
				if ret.Opcode == INTERNALDATAOPCODE && len(ret.Args) >= 2 && ret.Args[0] == PINGOPCODE {
					if err := t.SendWsMessage(guacd.NewInstruction(INTERNALDATAOPCODE, PINGOPCODE)); err != nil {
						logger.Error(err)
					}
					continue
				}
			} else {
				logger.Errorf("Parse instruction err %s", err)
			}
			_, err = t.writeTunnelMessage(message)
			if err != nil {
				logger.Errorf("Guacamole server write err: %+v", err)
				exit <- err
				break
			}
		}
		_ = t.guacdTunnel.Close()
	}(m)

	for {
		select {
		case err = <-exit:
			return err
		case <-ctx.Done():
			return nil
		}
	}
}
