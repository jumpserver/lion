package tunnel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/jumpserver-dev/sdk-go/common"
	"github.com/jumpserver-dev/sdk-go/model"
	"github.com/jumpserver-dev/sdk-go/service"

	"lion/pkg/config"
	"lion/pkg/guacd"
	"lion/pkg/logger"
	"lion/pkg/session"
)

type ReplayRecorder struct {
	tunnelSession *session.TunnelSession
	SessionId     string
	guacdAddr     string
	conf          guacd.Configuration
	info          guacd.ClientInformation
	newPartChan   chan struct{}
	currentIndex  atomic.Int64
	MaxSize       int
	apiClient     *service.JMService

	RootPath string
	wg       sync.WaitGroup
	started  atomic.Bool
}

func (r *ReplayRecorder) run(ctx context.Context) {
	defer r.wg.Done()
	r.startRecordPartReplay(ctx, 0)
	nextIndex := 1
	for {
		// Prefer cancellation over a queued rollover. This prevents a new
		// recorder from being started while Stop is already waiting.
		if ctx.Err() != nil {
			logger.Infof("ReplayRecorder %s done", r.SessionId)
			return
		}
		select {
		case <-ctx.Done():
			logger.Infof("ReplayRecorder %s done", r.SessionId)
			return
		case <-r.newPartChan:
			if ctx.Err() != nil {
				return
			}
			r.startRecordPartReplay(ctx, nextIndex)
			nextIndex++
		}
	}
}

func (r *ReplayRecorder) startRecordPartReplay(ctx context.Context, index int) {
	r.currentIndex.Store(int64(index))
	r.wg.Add(1)
	go r.recordReplay(ctx, index)
}

func (r *ReplayRecorder) Start(ctx context.Context) {
	if r.tunnelSession == nil || r.tunnelSession.TerminalConfig == nil ||
		r.tunnelSession.ModelSession == nil {
		logger.Errorf("ReplayRecorder %s session metadata or terminal config is nil, not record", r.SessionId)
		return
	}
	if r.tunnelSession.TerminalConfig.ReplayStorage.TypeName == "null" {
		logger.Warnf("ReplayRecorder %s storage is null, not record", r.SessionId)
		return
	}
	if !r.started.CompareAndSwap(false, true) {
		logger.Warnf("ReplayRecorder %s already started", r.SessionId)
		return
	}
	rootPath := filepath.Join(config.GlobalConfig.SessionFolderPath, r.SessionId)
	if err := os.MkdirAll(rootPath, 0o750); err != nil {
		r.started.Store(false)
		logger.Errorf("ReplayRecorder %s create root path %s failed: %v", r.SessionId, rootPath, err)
		return
	}
	r.RootPath = rootPath
	if err := r.WriteSessionMeta(r.tunnelSession.Created); err != nil {
		r.started.Store(false)
		return
	}
	if r.newPartChan == nil {
		r.newPartChan = make(chan struct{}, 1)
	}
	r.wg.Add(1)
	go r.run(ctx)
}

func (r *ReplayRecorder) WriteSessionMeta(t common.UTCTime) error {
	var sessionData struct {
		model.Session
		DateEnd common.UTCTime `json:"date_end"`
	}
	sessionData.Session = *r.tunnelSession.ModelSession
	sessionData.DateEnd = t
	metaFilename := r.SessionId + ".json"
	metaFilePath := filepath.Join(r.RootPath, metaFilename)
	metaBuf, err := json.Marshal(sessionData)
	if err != nil {
		logger.Errorf("ReplayRecorder(%s) marshal session meta failed: %v", r.SessionId, err)
		return err
	}
	if err = writeFileAtomically(metaFilePath, metaBuf, 0o600); err != nil {
		logger.Errorf("ReplayRecorder(%s) Write session meta file %s failed: %v", r.SessionId, metaFilename, err)
		return err
	}
	logger.Infof("ReplayRecorder(%s) Write session meta file %s success", r.SessionId, metaFilename)
	return nil
}

func (r *ReplayRecorder) IsConnectFailed() bool {
	partFiles, err := collectReplayPartFiles(r.RootPath, r.SessionId)
	if err != nil {
		logger.Errorf("ReplayRecorder %s collect part files error: %v", r.SessionId, err)
		return true
	}
	for _, partFile := range partFiles {
		partFilePath := filepath.Join(r.RootPath, partFile.Name())
		if _, err = loadPartMeta(partFilePath); err == nil {
			return false
		}
	}
	return true
}

func (r *ReplayRecorder) CleanFailedPartFileReplay() {
	partFiles, err := collectReplayPartFiles(r.RootPath, r.SessionId)
	if err != nil {
		logger.Errorf("ReplayRecorder %s collect part files error: %v", r.SessionId, err)
		return
	}
	for _, partFile := range partFiles {
		partFilePath := filepath.Join(r.RootPath, partFile.Name())
		if _, err = loadPartMeta(partFilePath); err != nil {
			logger.Warnf("ReplayRecorder %s remove unusable part file %s: %v",
				r.SessionId, partFile.Name(), err)
			_ = os.Remove(partFilePath)
			_ = os.Remove(partFilePath + MetaSuffix)
		}
	}
}

func (r *ReplayRecorder) Stop() {
	if !r.started.CompareAndSwap(true, false) {
		return
	}
	r.wg.Wait()
	if err := r.WriteSessionMeta(common.NewNowUTCTime()); err != nil {
		logger.Errorf("ReplayRecorder %s update session meta failed: %v", r.SessionId, err)
		return
	}
	uploader := PartUploader{
		RootPath:  r.RootPath,
		SessionId: r.SessionId,
		ApiClient: r.apiClient,
		TermCfg:   r.tunnelSession.TerminalConfig,
		Info:      r.info,
	}

	r.CleanFailedPartFileReplay()
	// A replay is usable when at least one part contains a valid sync
	// instruction. File size alone incorrectly rejects short valid sessions.
	if r.IsConnectFailed() {
		logger.Warnf("ReplayRecorder %s connect failed, not upload replay parts", r.SessionId)
		if err := os.RemoveAll(r.RootPath); err != nil {
			logger.Errorf("ReplayRecorder %s remove root path %s error: %v", r.SessionId, r.RootPath, err)
		}
		return
	}
	go uploader.Start()
	logger.Infof("Replay recorder %s stop and uploading replay parts", r.SessionId)
}

func (r *ReplayRecorder) GetPartFilename() string {
	return r.GetPartFilenameByIndex(int(r.currentIndex.Load()))
}

func (r *ReplayRecorder) GetPartFilenameByIndex(index int) string {
	return fmt.Sprintf("%s.%d.part", r.SessionId, index)
}

type PartMeta struct {
	StartTime int64 `json:"start,omitempty"`
	EndTime   int64 `json:"end,omitempty"`
	Duration  int64 `json:"duration,omitempty"`
	Size      int64 `json:"size,omitempty"`
	SyncCount int64 `json:"sync_count,omitempty"`
}

const (
	PartSuffix = ".part"
	MetaSuffix = ".meta"
)

type replayTunnel interface {
	ReadInstruction() (guacd.Instruction, error)
	WriteInstructionAndFlush(guacd.Instruction) error
	Close() error
}

func (r *ReplayRecorder) recordReplay(ctx context.Context, index int) {
	defer r.wg.Done()
	if ctx.Err() != nil {
		return
	}
	joinTunnel, err1 := guacd.NewTunnelContext(ctx, r.guacdAddr, r.conf, r.info)
	if err1 != nil {
		if ctx.Err() != nil {
			return
		}
		logger.Errorf("Join replay tunnel %s failed: %v", r.SessionId, err1)
		return
	}
	defer joinTunnel.Close()
	partFilename := r.GetPartFilenameByIndex(index)
	partMetaFilename := partFilename + MetaSuffix
	partFilePath := filepath.Join(r.RootPath, partFilename)
	partMetaFilePath := filepath.Join(r.RootPath, partMetaFilename)
	partRecorder := PartRecorder{
		Id:           r.SessionId,
		MetaFilename: partMetaFilename,
		MetaFilePath: partMetaFilePath,
		PartFilename: partFilename,
		PartFilePath: partFilePath,
		MaxSize:      r.MaxSize,
		currentIndex: index,
		ExitSignal: func() {
			select {
			case r.newPartChan <- struct{}{}:
			case <-ctx.Done():
			}
		},
	}
	partRecorder.Start(ctx, joinTunnel)
}

func NewReplayConfiguration(conf *guacd.Configuration, connectionId string) guacd.Configuration {
	newCfg := conf.Clone()
	newCfg.ConnectionID = connectionId
	newCfg.SetParameter(guacd.READONLY, guacd.BoolTrue)
	return newCfg
}

type PartRecorder struct {
	Id           string
	MetaFilename string
	MetaFilePath string

	PartFilename string
	PartFilePath string

	MaxSize      int
	currentIndex int
	ExitSignal   func()

	StartTime int64
	EndTime   int64
	SyncCount int64
}

func (p *PartRecorder) String() string {
	return fmt.Sprintf("%s, part %d", p.Id, p.currentIndex)
}

func (p *PartRecorder) Start(ctx context.Context, joinTunnel replayTunnel) {
	fd, err := os.OpenFile(p.PartFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		logger.Errorf("PartRecorder create replay file %s failed: %v", p.PartFilePath, err)
		return
	}
	defer fd.Close()
	writer := bufio.NewWriter(fd)
	totalWrittenSize := 0
	disconnectInst := guacd.NewInstruction(guacd.InstructionClientDisconnect)
	waitExit := false
	watchDone := make(chan struct{})
	defer close(watchDone)
	go func() {
		select {
		case <-ctx.Done():
			_ = joinTunnel.Close()
		case <-watchDone:
		}
	}()
	for {
		if ctx.Err() != nil {
			break
		}
		inst, err2 := joinTunnel.ReadInstruction()
		if err2 != nil {
			if ctx.Err() != nil {
				break
			}
			if waitExit && (err2 == io.EOF) {
				logger.Infof("PartRecorder(%s) tunnel EOF", p)
				break
			}
			logger.Warnf("PartRecorder(%s) read failed: %v", p, err2)
			break
		}
		if inst.Opcode == INTERNALDATAOPCODE && len(inst.Args) >= 2 && inst.Args[0] == PINGOPCODE {
			if err3 := joinTunnel.WriteInstructionAndFlush(
				guacd.NewInstruction(INTERNALDATAOPCODE, PINGOPCODE)); err3 != nil {
				logger.Warnf("Join tunnel %s write ping failed: %v", p.Id, err3)
			}
			continue
		}
		switch inst.Opcode {
		case guacd.InstructionClientSync:
			_ = joinTunnel.WriteInstructionAndFlush(inst)
			if len(inst.Args) > 0 {
				syncTime, err3 := strconv.ParseInt(inst.Args[0], 10, 64)
				if err3 != nil {
					break
				}
				p.EndTime = syncTime
				if p.SyncCount == 0 {
					p.StartTime = syncTime
				}
				p.SyncCount++
			}
		case guacd.InstructionClientNop:
			logger.Debugf("PartRecorder(%s) receive nop", p)
			_ = joinTunnel.WriteInstructionAndFlush(inst)
			continue
		default:
		}
		wr, err3 := writer.WriteString(inst.String())
		if err3 != nil {
			logger.Errorf("PartRecorder(%s) write failed: %v", p, err3)
			break
		}
		totalWrittenSize += wr
		if p.MaxSize > 0 && totalWrittenSize >= p.MaxSize && !waitExit &&
			inst.Opcode != guacd.InstructionClientDisconnect {
			if err3 = joinTunnel.WriteInstructionAndFlush(disconnectInst); err3 != nil {
				logger.Warnf("PartRecorder(%s) send disconnect failed: %v", p, err3)
			}
			waitExit = true
			logger.Infof("PartRecorder(%s) finish, start new part", p)
			if p.ExitSignal != nil {
				p.ExitSignal()
			}
		}
		if inst.Opcode == guacd.InstructionClientDisconnect {
			logger.Infof("PartRecorder(%s) receive disconnect", p)
			break
		}
	}
	if err = writer.Flush(); err != nil {
		logger.Errorf("PartRecorder(%s) flush replay file failed: %v", p, err)
		return
	}
	if err = fd.Sync(); err != nil {
		logger.Errorf("PartRecorder(%s) sync replay file failed: %v", p, err)
		return
	}
	p.WritePartMeta(totalWrittenSize)
}

func (p *PartRecorder) WritePartMeta(size int) {
	meta := PartMeta{
		StartTime: p.StartTime,
		EndTime:   p.EndTime,
		Duration:  p.EndTime - p.StartTime,
		Size:      int64(size),
		SyncCount: p.SyncCount,
	}
	metaBuf, err := json.Marshal(meta)
	if err != nil {
		logger.Errorf("Marshal replay meta file %s failed: %v", p.MetaFilename, err)
		return
	}
	if err = writeFileAtomically(p.MetaFilePath, metaBuf, 0o600); err != nil {
		logger.Errorf("Write replay meta file %s failed: %v", p.MetaFilename, err)
	}
}

func writeFileAtomically(filename string, data []byte, perm os.FileMode) (err error) {
	tempFile, err := os.CreateTemp(filepath.Dir(filename), "."+filepath.Base(filename)+".tmp-")
	if err != nil {
		return err
	}
	tempName := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempName)
	}()
	if err = tempFile.Chmod(perm); err != nil {
		return err
	}
	written, err := tempFile.Write(data)
	if err != nil {
		return err
	}
	if written != len(data) {
		return io.ErrShortWrite
	}
	if err = tempFile.Sync(); err != nil {
		return err
	}
	if err = tempFile.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, filename)
}
