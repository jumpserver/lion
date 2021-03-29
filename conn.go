package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"guacamole-client-go/pkg/guacd"
)

const (
	INTERNALDATAOPCODE = ""
	PINGOPCODE         = "ping"
)

type TunnelConn struct {
	guacdTunnel *guacd.Tunnel

	ws *websocket.Conn

	wsLock    sync.Mutex
	guacdLock sync.Mutex

	outputFilter *OutputStreamInterceptingFilter

	inputFilter *InputStreamInterceptingFilter
}

func (t *TunnelConn) SendWsMessage(msg guacd.Instruction) error {
	return t.writeWsMessage([]byte(msg.String()))
}

func (t *TunnelConn) writeWsMessage(p []byte) error {
	t.wsLock.Lock()
	defer t.wsLock.Unlock()
	return t.ws.WriteMessage(websocket.TextMessage, p)
}

func (t *TunnelConn) WriteTunnelMessage(msg guacd.Instruction) (err error) {
	_, err = t.writeTunnelMessage([]byte(msg.String()))
	return err
}

func (t *TunnelConn) writeTunnelMessage(p []byte) (int, error) {
	t.guacdLock.Lock()
	defer t.guacdLock.Unlock()
	return t.guacdTunnel.WriteAndFlush(p)
}

func (t *TunnelConn) readTunnelMessage() ([]byte, error) {
	for {
		instruction, err := t.guacdTunnel.ReadInstruction()
		if err != nil {
			return nil, err
		}
		newInstruction := t.inputFilter.Filter(&instruction)
		if newInstruction == nil {
			fmt.Println("inputFilter continue")
			continue
		}
		newInstruction = t.outputFilter.Filter(newInstruction)
		if newInstruction == nil {
			fmt.Println("continue")
			continue
		}
		return []byte(newInstruction.String()), nil
	}

}

func (t *TunnelConn) Run(ctx *gin.Context) (err error) {
	// 需要发送 uuid 返回给 guacamole tunnel
	err = t.SendWsMessage(guacd.NewInstruction(
		INTERNALDATAOPCODE, t.guacdTunnel.UUID))
	if err != nil {
		log.Println("Run err: ", err)
		return err
	}
	exit := make(chan error, 2)
	go func(t *TunnelConn) {
		for {
			instructionBytes, err := t.readTunnelMessage()
			if err != nil {
				fmt.Printf("tunnel Read %v\n", err)
				exit <- err
				break
			}

			if err = t.writeWsMessage(instructionBytes); err != nil {
				fmt.Printf(" ws WriteMessage %v\n", err)
				exit <- err
				break
			}
		}
		_ = t.ws.Close()
	}(t)

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

		}
		_, err = t.writeTunnelMessage(message)
		if err != nil {
			fmt.Printf("tunnel WriteAndFlush %v\n", err)
			exit <- err
			break
		}
	}

	select {
	case err = <-exit:
		log.Println("run exit: ", err.Error())
	case <-ctx.Request.Context().Done():
		_ = t.ws.Close()
		_ = t.guacdTunnel.Close()
	default:
	}
	log.Println("TunnelConn goroutines are terminated.")
	return
}
