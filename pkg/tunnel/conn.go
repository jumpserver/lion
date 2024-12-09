package tunnel

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"lion/pkg/common"
	"lion/pkg/guacd"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/logger"
	"lion/pkg/session"
)

const (
	INTERNALDATAOPCODE = ""
	PINGOPCODE         = "ping"
)

var _ sort.Interface = Connections{}

type Connections []*Connection

func (c Connections) Len() int {
	return len(c)
}
func (c Connections) Less(i, j int) bool {
	iCreated := c[i].Sess.Created.UTC()
	jCreated := c[j].Sess.Created.UTC()
	return iCreated.Before(jCreated)
}

func (c Connections) Swap(i, j int) {
	c[j], c[i] = c[i], c[j]
}

type Connection struct {
	Sess        *session.TunnelSession
	guacdTunnel *guacd.Tunnel
	Service     *session.Server

	guacdAddr string

	ws *websocket.Conn

	wsLock    sync.Mutex
	guacdLock sync.Mutex

	outputFilter *OutputStreamInterceptingFilter

	inputFilter *InputStreamInterceptingFilter

	done chan struct{}

	traceLock sync.Mutex
	traceMap  map[*guacd.Tunnel]struct{}

	lockedStatus atomic.Bool
	operatorUser atomic.Value

	recordStatus atomic.Bool
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
		newInstruction := &instruction
		if t.inputFilter != nil {
			newInstruction = t.inputFilter.Filter(newInstruction)
			if newInstruction == nil {
				continue
			}
		}
		if t.outputFilter != nil {
			newInstruction = t.outputFilter.Filter(newInstruction)
			if newInstruction == nil {
				continue
			}
		}
		return newInstruction, nil
	}

}

func (t *Connection) Run(ctx *gin.Context) (err error) {
	defer t.releaseMonitorTunnel()
	// 需要发送 uuid 返回给 guacamole tunnel
	err = t.SendWsMessage(guacd.NewInstruction(
		INTERNALDATAOPCODE, t.guacdTunnel.UUID))
	if err != nil {
		logger.Error("Run err: ", err)
		return err
	}

	parser := t.Service.GetFilterParser(t.Sess)
	userInputMessageChan := make(chan *session.Message, 1)
	defer func() {
		parser.Close()
	}()
	// 处理数据流
	parser.ParseStream(userInputMessageChan)
	// 记录命令
	cmdChan := parser.CommandRecordChan()
	go t.recordCommand(cmdChan)

	meta := session.MetaMessage{
		UserId:  t.Sess.User.ID,
		User:    t.Sess.User.String(),
		Created: common.NewNowUTCTime().String(),
	}
	exit := make(chan error, 2)
	activeChan := make(chan struct{})
	noNopTime := time.Now()
	maxNopTimeout := time.Minute * 5
	var requiredErr guacd.Instruction
	go func(t *Connection) {
		for {
			instruction, err := t.readTunnelInstruction()
			if err != nil {
				logger.Errorf("Session[%s] guacamole server read err: %+v", t, err)
				exit <- err
				break
			}

			switch instruction.Opcode {
			case guacd.InstructionServerDisconnect,
				guacd.InstructionServerError:
				logger.Infof("Session[%s] receive guacamole server disconnect: %s", t, instruction.String())
			case guacd.InstructionStreamingAck:
				select {
				case activeChan <- struct{}{}:
				default:
				}
			}

			switch instruction.Opcode {
			case guacd.InstructionClientNop:
				if time.Now().Sub(noNopTime) > maxNopTimeout {
					logger.Errorf("Session[%s] guacamole server nop timeout", t)
					if requiredErr.Opcode != "" {
						logger.Errorf("Session[%s] send guacamole server required err: %s", t,
							requiredErr.String())
						_ = t.writeWsMessage([]byte(requiredErr.String()))
						requiredErr = guacd.Instruction{}
						continue
					}

				}
			case guacd.InstructionRequired:
				msg := fmt.Sprintf("required: %s", strings.Join(instruction.Args, ","))
				logger.Infof("Session[%s] receive guacamole server required: %s", t, msg)
				requiredErr = guacd.NewInstruction(guacd.InstructionServerError, msg)
				logger.Errorf("Session[%s] send guacamole server required err: %s", t,
					requiredErr.String())
				_ = t.writeWsMessage([]byte(requiredErr.String()))
				requiredErr = guacd.Instruction{}
				continue
			default:
				noNopTime = time.Now()
			}

			if err = t.writeWsMessage([]byte(instruction.String())); err != nil {
				logger.Errorf("Session[%s] send web client err: %+v", t, err)
				exit <- err
				break
			}
		}
		_ = t.ws.Close()
	}(t)

	go func(t *Connection) {
		for {
			_, message, err1 := t.ws.ReadMessage()
			if err1 != nil {
				if websocket.IsCloseError(err1, websocket.CloseNoStatusReceived) {
					logger.Warnf("Session[%s] web client read err: %+v", t, err1)
				} else {
					logger.Errorf("Session[%s] web client read err: %+v", t, err1)
				}
				exit <- err1
				break
			}

			if ret, err2 := guacd.ParseInstructionString(string(message)); err2 == nil {
				if ret.Opcode == INTERNALDATAOPCODE && len(ret.Args) >= 2 && ret.Args[0] == PINGOPCODE {
					if err3 := t.SendWsMessage(guacd.NewInstruction(INTERNALDATAOPCODE, PINGOPCODE)); err3 != nil {
						logger.Errorf("Session[%s] unable to send 'ping' response for WebSocket tunnel: %+v",
							t, err3)
					}
					continue
				}
				if t.lockedStatus.Load() {
					switch ret.Opcode {
					case guacd.InstructionClientSync,
						guacd.InstructionClientNop,
						guacd.InstructionStreamingAck:
					default:
						select {
						case activeChan <- struct{}{}:
						default:
						}
						logger.Infof("Session[%s] in locked status drop receive web client message opcode[%s]",
							t, ret.Opcode)
						continue
					}
					_, err4 := t.writeTunnelMessage(message)
					if err4 != nil {
						logger.Errorf("Session[%s] guacamole server write err: %+v", t, err2)
						exit <- err4
						break
					}
					logger.Debugf("Session[%s] send guacamole server message when locked status", t)
					continue
				}

				switch ret.Opcode {
				case guacd.InstructionKey:
					userInputMessageChan <- &session.Message{
						Opcode: ret.Opcode, Body: ret.Args,
						Meta: meta}
				default:

				}

				switch ret.Opcode {
				case guacd.InstructionClientSync,
					guacd.InstructionClientNop,
					guacd.InstructionStreamingAck:
				case guacd.InstructionClientDisconnect:
					logger.Infof("Session[%s] receive web client disconnect opcode", t)
				default:
					select {
					case activeChan <- struct{}{}:
					default:
					}
				}
			} else {
				logger.Errorf("Session[%s] parse instruction err %s", t, err)
			}
			_, err = t.writeTunnelMessage(message)
			if err != nil {
				logger.Errorf("Session[%s] guacamole server write err: %+v", t, err)
				exit <- err
				break
			}
		}
	}(t)
	maxIndexTime := t.Sess.TerminalConfig.MaxIdleTime
	maxSessionTimeInt := t.Sess.TerminalConfig.MaxSessionTime
	maxSessionDuration := time.Duration(maxSessionTimeInt) * time.Hour
	maxSessionTime := time.Now().Add(maxSessionDuration)
	maxIdleMinutes := time.Duration(maxIndexTime) * time.Minute
	activeDetectTicker := time.NewTicker(time.Minute)
	defer activeDetectTicker.Stop()
	latestActive := time.Now()
	for {
		select {
		case err = <-exit:
			logger.Infof("Session[%s] Connection exit %+v", t, err)
			if !t.recordStatus.Load() {
				reason := model.SessionLifecycleLog{Reason: string(model.ReasonErrConnectDisconnect)}
				t.Service.RecordLifecycleLog(t.Sess.ID, model.AssetConnectFinished, reason)
			}
			return err
		case <-ctx.Request.Context().Done():
			_ = t.ws.Close()
			_ = t.guacdTunnel.Close()
			reason := model.SessionLifecycleLog{Reason: string(model.ReasonErrConnectDisconnect)}
			t.Service.RecordLifecycleLog(t.Sess.ID, model.AssetConnectFinished, reason)
			logger.Errorf("Session[%s] request ctx done", t)
			return nil
		case <-activeChan:
			latestActive = time.Now()
		case detectTime := <-activeDetectTicker.C:
			if detectTime.After(maxSessionTime) {
				errSession := NewJMSMaxSessionTimeError(t.Sess.TerminalConfig.MaxSessionTime)
				_ = t.SendWsMessage(errSession.Instruction())
				logger.Errorf("Session[%s] terminated by max session time %d hour",
					t, maxSessionTimeInt)
				reason := model.SessionLifecycleLog{Reason: string(model.ReasonErrMaxSessionTimeout)}
				t.Service.RecordLifecycleLog(t.Sess.ID, model.AssetConnectFinished, reason)
				return nil
			}
			if detectTime.After(latestActive.Add(maxIdleMinutes)) {
				errIdle := NewJMSIdleTimeOutError(maxIndexTime)
				_ = t.SendWsMessage(errIdle.Instruction())
				logger.Errorf("Session[%s] terminated by %d min timeout",
					t, maxIndexTime)
				reason := model.SessionLifecycleLog{Reason: string(model.ReasonErrIdleDisconnect)}
				t.Service.RecordLifecycleLog(t.Sess.ID, model.AssetConnectFinished, reason)
				return nil
			}
			if t.IsPermissionExpired(detectTime) {
				_ = t.SendWsMessage(ErrPermissionExpired.Instruction())
				logger.Errorf("Session[%s] permission has expired", t)
				reason := model.SessionLifecycleLog{Reason: string(model.ReasonErrPermissionExpired)}
				t.Service.RecordLifecycleLog(t.Sess.ID, model.AssetConnectFinished, reason)
				return nil
			}
		}
	}

}

func (t *Connection) HandleTask(task *model.TerminalTask) error {
	switch task.Name {
	case model.TaskUnlockSession:
		t.lockedStatus.Store(false)
		t.operatorUser.Store(task.Kwargs.CreatedByUser)
		data := map[string]interface{}{
			"user": task.Kwargs.CreatedByUser,
		}
		p, _ := json.Marshal(data)
		ins := NewJmsEventInstruction("session_resume", string(p))
		_ = t.SendWsMessage(ins)
	case model.TaskLockSession:
		t.lockedStatus.Store(true)
		t.operatorUser.Store(task.Kwargs.CreatedByUser)
		data := map[string]interface{}{
			"user": task.Kwargs.CreatedByUser,
		}
		p, _ := json.Marshal(data)
		ins := NewJmsEventInstruction("session_pause", string(p))
		_ = t.SendWsMessage(ins)
	case model.TaskKillSession:
		t.recordStatus.Store(true)
		username := task.Kwargs.TerminatedBy
		ins := NewJMSGuacamoleError(1005, username)
		_ = t.SendWsMessage(ins.Instruction())
		reason := model.SessionLifecycleLog{Reason: string(model.ReasonErrAdminTerminate)}
		t.Service.RecordLifecycleLog(t.Sess.ID, model.AssetConnectFinished, reason)
	default:
		return fmt.Errorf("unknown task %s", task.Name)
	}
	logger.Infof("Session[%s] handle task %s", t, task.Name)
	return nil
}

func (t *Connection) String() string {
	return t.Sess.ID
}

func (t *Connection) IsPermissionExpired(now time.Time) bool {
	return t.Sess.ExpireInfo.IsExpired(now)
}

func (t *Connection) CloneMonitorTunnel() (*guacd.Tunnel, error) {
	info := guacd.NewClientInformation()
	conf := guacd.NewConfiguration()
	conf.ConnectionID = t.guacdTunnel.UUID
	guacdAddr := t.guacdAddr
	monitorTunnel, err := guacd.NewTunnel(guacdAddr, conf, info)
	if err != nil {
		return nil, err
	}
	t.traceMonitorTunnel(monitorTunnel)
	return monitorTunnel, err
}

func (t *Connection) traceMonitorTunnel(monitorTunnel *guacd.Tunnel) {
	t.traceLock.Lock()
	defer t.traceLock.Unlock()
	if t.traceMap == nil {
		t.traceMap = make(map[*guacd.Tunnel]struct{})
	}
	t.traceMap[monitorTunnel] = struct{}{}
}

func (t *Connection) releaseMonitorTunnel() {
	t.traceLock.Lock()
	defer t.traceLock.Unlock()
	if t.traceMap == nil {
		return
	}
	for tunneler := range t.traceMap {
		_ = tunneler.Close()
	}
}

func (t *Connection) unTraceMonitorTunnel(monitorTunnel *guacd.Tunnel) {
	t.traceLock.Lock()
	defer t.traceLock.Unlock()
	if t.traceMap == nil {
		return
	}
	delete(t.traceMap, monitorTunnel)
}

func (t *Connection) recordCommand(cmdRecordChan chan *session.ExecutedCommand) {
	// 命令记录
	cmdRecorder := t.Service.GetCommandRecorder(t.Sess)
	for item := range cmdRecordChan {
		if item.Command == "" {
			continue
		}
		cmd := t.generateCommandResult(item)
		cmdRecorder.Record(cmd)
	}
	// 关闭命令记录
	cmdRecorder.End()
}

// generateCommandResult 生成命令结果
func (t *Connection) generateCommandResult(item *session.ExecutedCommand) *model.Command {
	var (
		input     string
		output    string
		riskLevel int64
		user      string
	)
	user = item.User.User
	if len(item.Command) > 128 {
		input = item.Command[:128]
	} else {
		input = item.Command
	}
	switch item.RiskLevel {
	case model.HighRiskFlag:
		riskLevel = model.DangerLevel
	default:
		riskLevel = model.NormalLevel
	}
	return t.Service.GenerateCommandItem(t.Sess, user, input, output, riskLevel, item.CreatedDate)
}
