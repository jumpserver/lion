package tunnel

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/session"
)

const (
	INTERNALDATAOPCODE = ""
	PINGOPCODE         = "ping"
)

type Connection struct {
	sess        *session.TunnelSession
	guacdTunnel *guacd.Tunnel

	ws *websocket.Conn

	wsLock    sync.Mutex
	guacdLock sync.Mutex

	outputFilter *OutputStreamInterceptingFilter

	inputFilter *InputStreamInterceptingFilter
}

func (t *Connection) SendWsMessage(msg guacd.Instruction) error {
	return t.writeWsMessage([]byte(msg.String()))
}

func (t *Connection) writeWsMessage(p []byte) error {
	t.wsLock.Lock()
	defer t.wsLock.Unlock()
	return t.ws.WriteMessage(websocket.TextMessage, p)
}

func (t *Connection) WriteTunnelMessage(msg guacd.Instruction) (err error) {
	_, err = t.writeTunnelMessage([]byte(msg.String()))
	return err
}

func (t *Connection) writeTunnelMessage(p []byte) (int, error) {
	t.guacdLock.Lock()
	defer t.guacdLock.Unlock()
	return t.guacdTunnel.WriteAndFlush(p)
}

func (t *Connection) readTunnelInstruction() (*guacd.Instruction, error) {
	for {
		instruction, err := t.guacdTunnel.ReadInstruction()
		if err != nil {
			return nil, err
		}
		newInstruction := t.inputFilter.Filter(&instruction)
		if newInstruction == nil {
			continue
		}
		newInstruction = t.outputFilter.Filter(newInstruction)
		if newInstruction == nil {
			continue
		}
		return newInstruction, nil
	}

}

func (t *Connection) Run(ctx *gin.Context) (err error) {
	// 需要发送 uuid 返回给 guacamole tunnel
	err = t.SendWsMessage(guacd.NewInstruction(
		INTERNALDATAOPCODE, t.guacdTunnel.UUID))
	if err != nil {
		log.Println("Run err: ", err)
		return err
	}
	exit := make(chan error, 2)
	activeChan := make(chan struct{})
	go func(t *Connection) {
		for {
			instruction, err := t.readTunnelInstruction()
			if err != nil {
				fmt.Printf("tunnel Read %v\n", err)
				exit <- err
				break
			}

			switch instruction.Opcode {
			case guacd.InstructionServerDisconnect,
				guacd.InstructionServerError,
				guacd.InstructionServerLog:
				fmt.Println(instruction.String())
			}

			if err = t.writeWsMessage([]byte(instruction.String())); err != nil {
				fmt.Printf(" ws WriteMessage %v\n", err)
				exit <- err
				break
			}
		}
		_ = t.ws.Close()
	}(t)

	go func(t *Connection) {
		for {
			_, message, err := t.ws.ReadMessage()
			if err != nil {
				fmt.Printf("ws ReadMessage %v\n", err)
				exit <- err
				break
			}

			if ret, err := guacd.ParseInstructionString(string(message)); err == nil {
				if ret.Opcode == INTERNALDATAOPCODE && len(ret.Args) >= 2 && ret.Args[0] == PINGOPCODE {
					if err := t.SendWsMessage(guacd.NewInstruction(INTERNALDATAOPCODE, PINGOPCODE)); err != nil {
						fmt.Printf("Unable to send 'ping' response for WebSocket tunnel. %s\n", err)
					}
					continue
				}
				if ret.Opcode == guacd.InstructionClientDisconnect {
					fmt.Println(ret.String())
				}
			}

			_, err = t.writeTunnelMessage(message)
			if err != nil {
				fmt.Printf("tunnel WriteAndFlush %v\n", err)
				exit <- err
				break
			}
			activeChan <- struct{}{}
		}
	}(t)
	activeDetectTicker := time.NewTicker(time.Minute)
	defer activeDetectTicker.Stop()
	latestActive := time.Now()
	for {
		select {
		case err = <-exit:
			log.Println("run exit: ", err.Error())
			return
		case <-ctx.Request.Context().Done():
			fmt.Println("Done")
			_ = t.ws.Close()
			_ = t.guacdTunnel.Close()
		case <-activeChan:
			latestActive = time.Now()
		case detectNow := <-activeDetectTicker.C:
			if latestActive.Add(time.Minute * 30).After(detectNow) {
				log.Println("Connection are terminated by timeout")
				return
			}
		}
	}

}
