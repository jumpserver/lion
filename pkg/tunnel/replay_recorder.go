package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"lion/pkg/config"
	"lion/pkg/guacd"
	"lion/pkg/logger"
	"lion/pkg/session"
)

const (
	recordDirTimeFormat = "2006-01-02"
	maxSize             = 1024 * 1024 * 50
)

type ReplayRecorder struct {
	tunnelSession *session.TunnelSession
	RootPath      string
	SessionId     string
	guacdAddr     string
	conf          guacd.Configuration
	info          guacd.ClientInformation
	wg            sync.WaitGroup
	newChan       chan struct{}
	currentIndex  int
}

func (r *ReplayRecorder) run(ctx context.Context) {
	r.startRecordPartReplay(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Infof("replay conn %s done", r.SessionId)
			return
		case <-r.newChan:
			r.currentIndex++
			r.startRecordPartReplay(ctx)
		}
	}
}

func (r *ReplayRecorder) startRecordPartReplay(ctx context.Context) {
	r.wg.Add(1)
	go r.recordReplay(ctx, &r.wg)
}

func (r *ReplayRecorder) Start(ctx context.Context) {
	if r.tunnelSession.TerminalConfig.ReplayStorage.TypeName == "null" {
		logger.Infof("replay storage is null, not record")
		return
	}
	recordDirPath := filepath.Join(config.GlobalConfig.RecordPath,
		r.tunnelSession.Created.Format(recordDirTimeFormat))
	sessionReplayRootPath := filepath.Join(recordDirPath, r.tunnelSession.ID)
	_ = os.MkdirAll(sessionReplayRootPath, os.ModePerm)
	r.RootPath = sessionReplayRootPath
	go r.run(ctx)
}

func (r *ReplayRecorder) Stop() {
	r.wg.Wait()
	// todo: 上传分段的录像文件
	logger.Infof("Replay recorder %s stop", r.SessionId)
}

func (r *ReplayRecorder) recordReplay(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	path := filepath.Join(r.RootPath, strconv.Itoa(r.currentIndex))
	fd, err := os.Create(path)
	if err != nil {
		logger.Errorf("create replay file %s failed: %v", path, err)
		return
	}
	defer fd.Close()
	writer := bufio.NewWriter(fd)
	defer writer.Flush()
	joinTunnel, err1 := guacd.NewTunnel(r.guacdAddr, r.conf, r.info)
	if err1 != nil {
		logger.Errorf("Join replay tunnel %s failed: %v", r.SessionId, err1)
		return
	}
	defer joinTunnel.Close()
	totalWriteSize := 0
	disconnectInst := guacd.NewInstruction(guacd.InstructionClientDisconnect)
	waitExit := false
	for {
		inst, err2 := joinTunnel.ReadInstruction()
		if err2 != nil {
			if waitExit && (err1 == io.EOF) {
				logger.Infof("Join replay tunnel %s EOF", r.SessionId)
				return
			}
			logger.Warnf("Join tunnel %s read failed: %v", r.SessionId, err2)
			return
		}
		if inst.Opcode == INTERNALDATAOPCODE && len(inst.Args) >= 2 && inst.Args[0] == PINGOPCODE {
			if err3 := joinTunnel.WriteInstruction(guacd.NewInstruction(INTERNALDATAOPCODE, PINGOPCODE)); err3 != nil {
				fmt.Println("write internal data opcode ping failed: ", err3)
			}
			continue
		}
		select {
		case <-ctx.Done():
			if !waitExit {
				_ = joinTunnel.WriteInstructionAndFlush(disconnectInst)
				waitExit = true
				logger.Infof("recordReplay ctx %s done", r.SessionId)
			} else {
				logger.Infof("recordReplay ctx %s done, wait exit", r.SessionId)
			}
		default:

		}
		switch inst.Opcode {
		case guacd.InstructionClientSync:
			_ = joinTunnel.WriteInstructionAndFlush(inst)
		case guacd.InstructionClientNop:
			logger.Debugf("join replay nop")
			continue
		default:
		}
		wr, err3 := writer.WriteString(inst.String())
		if err3 != nil {
			logger.Errorf("write replay file %s failed: %v", path, err3)
		}
		totalWriteSize += wr
		if totalWriteSize > maxSize && !waitExit {
			logger.Infof("write size > 50M, create new replay file")
			_ = joinTunnel.WriteInstructionAndFlush(disconnectInst)
			waitExit = true
			r.newChan <- struct{}{}
		}
	}

}

func NewReplayConfiguration(conf *guacd.Configuration, connectionId string) guacd.Configuration {
	newCfg := conf.Clone()
	newCfg.ConnectionID = connectionId
	newCfg.SetParameter(guacd.READONLY, "true")
	return newCfg
}
