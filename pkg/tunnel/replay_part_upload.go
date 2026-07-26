package tunnel

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"lion/pkg/config"
	"lion/pkg/guacd"
	"lion/pkg/logger"

	"github.com/jumpserver-dev/sdk-go/common"
	"github.com/jumpserver-dev/sdk-go/model"
	"github.com/jumpserver-dev/sdk-go/service"
	"github.com/jumpserver-dev/sdk-go/service/videoworker"
	"github.com/jumpserver-dev/sdk-go/storage"
)

/*
	原始录像的 part 数据格式

data/sessions/e32248ce-2dc8-43c8-b37e-a61d5ee32176
├── e32248ce-2dc8-43c8-b37e-a61d5ee32176.0.part
├── e32248ce-2dc8-43c8-b37e-a61d5ee32176.0.part.meta
└── e32248ce-2dc8-43c8-b37e-a61d5ee32176.json

upload
├── e32248ce-2dc8-43c8-b37e-a61d5ee32176.replay.json
├── e32248ce-2dc8-43c8-b37e-a61d5ee32176.0.part.gz
*/

const ReplayType = "guacamole"

type SessionReplayMeta struct {
	model.Session
	DateEnd    common.UTCTime `json:"date_end,omitempty"`
	ReplayType string         `json:"type,omitempty"`

	PartMetas []PartFileMeta `json:"files,omitempty"`
}

type PartFileMeta struct {
	Name string `json:"name"`
	PartMeta
}

type PartUploader struct {
	SessionId string
	RootPath  string
	ApiClient *service.JMService
	TermCfg   *model.TerminalConfig

	replayMeta SessionReplayMeta
	partFiles  []os.DirEntry

	Info guacd.ClientInformation
}

func (p *PartUploader) preCheckSessionMeta() error {
	metaPath := filepath.Join(p.RootPath, p.SessionId+".json")
	if _, err := os.Stat(metaPath); err != nil {
		logger.Errorf("PartUploader %s get meta file error: %v", p.SessionId, err)
		return err
	}
	metaBuf, err := os.ReadFile(metaPath)
	if err != nil {
		logger.Errorf("PartUploader %s read meta file error: %v", p.SessionId, err)
		return err
	}
	if err1 := json.Unmarshal(metaBuf, &p.replayMeta); err1 != nil {
		logger.Errorf("PartUploader %s unmarshal meta file error: %v", p.SessionId, err1)
		return err1
	}
	if p.replayMeta.DateStart.IsZero() {
		return errors.New("session replay meta has no start time")
	}
	if p.replayMeta.DateStart == p.replayMeta.DateEnd {
		// 未结束的录像, 计算结束时间，并上传到 core api 作为会话结束时间
		endTime := GetMaxModTime(p.partFiles)
		if endTime.IsZero() {
			return errors.New("cannot determine unfinished session end time")
		}
		p.replayMeta.DateEnd = common.NewUTCTime(endTime)
		// api finish time
		if p.ApiClient == nil {
			return errors.New("API client is nil")
		}
		if _, err1 := p.ApiClient.SessionFinished(p.SessionId, p.replayMeta.DateEnd); err1 != nil {
			logger.Errorf("PartUploader %s finish session error: %v", p.SessionId, err1)
			return err1
		}
		// write meta file
		metaBuf, err = json.Marshal(p.replayMeta)
		if err != nil {
			return err
		}
		if err1 := writeFileAtomically(metaPath, metaBuf, 0o600); err1 != nil {
			logger.Errorf("PartUploader %s write meta file error: %v", p.SessionId, err1)
			return err1
		}
	}
	if p.replayMeta.DateEnd.Before(p.replayMeta.DateStart.Time) {
		return errors.New("session replay end time is before start time")
	}
	p.replayMeta.ReplayType = ReplayType
	return nil
}

func GetMaxModTime(parts []os.DirEntry) time.Time {
	var t time.Time
	for i := range parts {
		partFile := parts[i]
		partFileInfo, err := partFile.Info()
		if err != nil {
			logger.Errorf("PartUploader get part file %s info error: %v", partFile.Name(), err)
			continue
		}
		modTime := partFileInfo.ModTime()
		if t.Before(modTime) {
			t = modTime
		}
	}
	return t
}

func (p *PartUploader) Start() {
	/*
		1、创建 upload 目录
		2、将所有的 part 文件压缩成gz文件，并移动到 upload 目录
		3、生成新的 meta 文件
		4、上传
	*/
	if p.TermCfg == nil {
		logger.Errorf("PartUploader %s terminal config is nil", p.SessionId)
		return
	}
	if p.ApiClient == nil {
		logger.Errorf("PartUploader %s API client is nil", p.SessionId)
		return
	}
	uploadPath, err := p.prepareUpload()
	if err != nil {
		logger.Errorf("PartUploader %s prepare upload failed: %v", p.SessionId, err)
		return
	}
	p.uploadToStorage(uploadPath)
}

func (p *PartUploader) prepareUpload() (string, error) {
	if err := p.CollectionPartFiles(); err != nil {
		return "", err
	}
	if len(p.partFiles) == 0 {
		return "", errors.New("no part file")
	}
	if err := p.preCheckSessionMeta(); err != nil {
		return "", err
	}
	// Build a clean staging directory first. Raw parts remain untouched, so
	// compression or process failures can be retried safely on the next boot.
	uploadPath := filepath.Join(p.RootPath, "upload")
	stagingPath := filepath.Join(p.RootPath, "upload.tmp")
	if err := os.RemoveAll(stagingPath); err != nil {
		return "", fmt.Errorf("clean staging directory: %w", err)
	}
	if err := os.MkdirAll(stagingPath, 0o750); err != nil {
		return "", fmt.Errorf("create staging directory: %w", err)
	}
	p.replayMeta.PartMetas = p.replayMeta.PartMetas[:0]
	// 2、将所有的 part 文件压缩移动到 upload 目录
	for i := range p.partFiles {
		partFile := p.partFiles[i]
		partFilePath := filepath.Join(p.RootPath, partFile.Name())
		partGzFilename := partFile.Name() + ".gz"
		uploadFilePath := filepath.Join(stagingPath, partGzFilename)

		partMeta, err := loadPartMeta(partFilePath)
		if err != nil {
			logger.Warnf("PartUploader %s skip unusable part file %s: %v",
				p.SessionId, partFile.Name(), err)
			continue
		}
		if err := common.CompressToGzipFile(partFilePath, uploadFilePath); err != nil {
			_ = os.RemoveAll(stagingPath)
			return "", fmt.Errorf("compress part file %s: %w", partFile.Name(), err)
		}

		// 3、生成新的 meta 文件
		partFileMeta := PartFileMeta{Name: partGzFilename, PartMeta: partMeta}
		p.replayMeta.PartMetas = append(p.replayMeta.PartMetas, partFileMeta)
	}
	if len(p.replayMeta.PartMetas) == 0 {
		_ = os.RemoveAll(stagingPath)
		return "", errors.New("no usable part file")
	}
	// upload 写入 replayMeta json
	replayMetaBuf, err := json.Marshal(p.replayMeta)
	if err != nil {
		_ = os.RemoveAll(stagingPath)
		return "", fmt.Errorf("marshal replay meta: %w", err)
	}
	if err = os.WriteFile(filepath.Join(stagingPath, p.SessionId+".replay.json"), replayMetaBuf, 0o600); err != nil {
		_ = os.RemoveAll(stagingPath)
		return "", fmt.Errorf("write replay meta file: %w", err)
	}
	if err = os.RemoveAll(uploadPath); err != nil {
		return "", fmt.Errorf("clean upload directory: %w", err)
	}
	if err = os.Rename(stagingPath, uploadPath); err != nil {
		return "", fmt.Errorf("publish staging directory: %w", err)
	}
	return uploadPath, nil
}

func collectReplayPartFiles(rootPath, sessionID string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}
	type indexedPart struct {
		index int
		entry os.DirEntry
	}
	indexedParts := make([]indexedPart, 0, 5)
	prefix := sessionID + "."
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, PartSuffix) {
			continue
		}
		indexText := strings.TrimSuffix(strings.TrimPrefix(name, prefix), PartSuffix)
		index, err1 := strconv.Atoi(indexText)
		if err1 != nil || index < 0 || strconv.Itoa(index) != indexText {
			continue
		}
		indexedParts = append(indexedParts, indexedPart{index: index, entry: entry})
	}
	sort.Slice(indexedParts, func(i, j int) bool {
		return indexedParts[i].index < indexedParts[j].index
	})
	partFiles := make([]os.DirEntry, 0, len(indexedParts))
	for _, part := range indexedParts {
		partFiles = append(partFiles, part.entry)
	}
	return partFiles, nil
}

func (p *PartUploader) CollectionPartFiles() error {
	partFiles, err := collectReplayPartFiles(p.RootPath, p.SessionId)
	if err != nil {
		logger.Errorf("PartUploader %s read dir %s error: %v", p.SessionId, p.RootPath, err)
		return err
	}
	p.partFiles = partFiles
	return nil
}

func (p *PartUploader) GetStorage() storage.ReplayStorage {
	return storage.NewReplayStorage(p.ApiClient, p.TermCfg.ReplayStorage)
}

const recordDirTimeFormat = "2006-01-02"

func (p *PartUploader) uploadToStorage(uploadPath string) {
	// check whether to use ENABLE_VIDEO_WORKER
	if videoWorkerClient := NewWorkerClient(*config.GlobalConfig); videoWorkerClient != nil {
		taskCfg := videoworker.TaskConfig{
			Width:   p.Info.OptimalScreenWidth,
			Height:  p.Info.OptimalScreenHeight,
			Bitrate: 1,
		}
		taskId, err := videoWorkerClient.CreateReplaySessionTask(p.SessionId, uploadPath, &taskCfg)
		if err == nil {
			logger.Infof("Create replay session VideoWorker task success, task id: %s", taskId)
			if err = os.RemoveAll(p.RootPath); err != nil {
				logger.Errorf("PartUploader %s remove root path %s error: %v", p.SessionId, p.RootPath, err)
			}
			return
		}
		// videoWorkerClient failed then try to use self storage to upload
		logger.Errorf("Create replay session task error: %v, try to use self storage", err)
	}

	// 上传到存储
	uploadFiles, err := os.ReadDir(uploadPath)
	if err != nil {
		logger.Errorf("PartUploader %s read upload dir %s error: %v", p.SessionId, uploadPath, err)
		return
	}
	//defaultStorage := storage.ServerStorage{StorageType: "server", JmsService: p.apiClient}
	p.RecordLifecycleLog(model.ReplayUploadStart, model.EmptyLifecycleLog)
	replayStorage := p.GetStorage()
	storageType := replayStorage.TypeName()
	dateRoot := p.replayMeta.DateStart.Format(recordDirTimeFormat)
	targetRoot := strings.Join([]string{dateRoot, p.SessionId}, "/")
	logger.Infof("PartUploader %s upload replay files: %v, type: %s", p.SessionId, uploadFiles, storageType)
	totalSize := int64(0)
	for _, uploadFile := range uploadFiles {
		if uploadFile.IsDir() {
			continue
		}
		fileInfo, err := uploadFile.Info()
		if err != nil {
			logger.Errorf("PartUploader %s get file info %s error: %v", p.SessionId, uploadFile.Name(), err)
			reason := model.SessionLifecycleLog{Reason: err.Error()}
			p.RecordLifecycleLog(model.ReplayUploadFailure, reason)
			return
		}
		totalSize += fileInfo.Size()
		uploadFilePath := filepath.Join(uploadPath, uploadFile.Name())
		targetFile := strings.Join([]string{targetRoot, uploadFile.Name()}, "/")
		if err1 := replayStorage.Upload(uploadFilePath, targetFile); err1 != nil {
			logger.Errorf("PartUploader %s upload file %s error: %v", p.SessionId, uploadFilePath, err1)
			reason := model.SessionLifecycleLog{Reason: err1.Error()}
			p.RecordLifecycleLog(model.ReplayUploadFailure, reason)
			return
		}
		logger.Debugf("PartUploader %s upload file %s success", p.SessionId, uploadFilePath)
	}
	if _, err = p.ApiClient.FinishReplyWithSize(p.SessionId, totalSize); err != nil {
		logger.Errorf("PartUploader %s finish replay error: %v", p.SessionId, err)
		reason := model.SessionLifecycleLog{Reason: err.Error()}
		p.RecordLifecycleLog(model.ReplayUploadFailure, reason)
		return
	}

	p.RecordLifecycleLog(model.ReplayUploadSuccess, model.EmptyLifecycleLog)
	logger.Infof("PartUploader %s upload replay success", p.SessionId)
	if err = os.RemoveAll(p.RootPath); err != nil {
		logger.Errorf("PartUploader %s remove root path %s error: %v", p.SessionId, p.RootPath, err)
		return
	}
	logger.Infof("PartUploader %s remove root path %s success", p.SessionId, p.RootPath)

}

func (p *PartUploader) RecordLifecycleLog(event model.LifecycleEvent, logObj model.SessionLifecycleLog) {
	if err := p.ApiClient.RecordSessionLifecycleLog(p.SessionId, event, logObj); err != nil {
		logger.Errorf("Record session %s lifecycle %s log err: %s", p.SessionId, event, err)
	}
}

func ReadInstruction(r *bufio.Reader) (guacd.Instruction, error) {
	var ret strings.Builder
	for {
		msg, err := r.ReadString(guacd.ByteSemicolonDelimiter)
		if err != nil && msg == "" {
			return guacd.Instruction{}, err
		}
		ret.WriteString(msg)
		if retInstruction, err1 := guacd.ParseInstructionString(ret.String()); err1 == nil {
			return retInstruction, nil
		} else {
			logger.Infof("ReadInstruction err:  %v\n", err1.Error())
		}
	}
}

func LoadPartMetaByFile(partFile string) (PartMeta, error) {
	var partMeta PartMeta
	info, err := os.Stat(partFile)
	if err != nil {
		logger.Errorf("LoadPartMetaByFile stat %s error: %v", partFile, err)
		return partMeta, err
	}
	partMeta.Size = info.Size()
	startTime, endTime, syncCount, err := scanPartReplayTime(partFile)
	if err != nil {
		logger.Errorf("LoadPartMetaByFile %s load replay time error: %v", partFile, err)
		return partMeta, err
	}
	partMeta.StartTime = startTime
	partMeta.EndTime = endTime
	partMeta.Duration = endTime - startTime
	partMeta.SyncCount = syncCount
	if partMeta.Duration < 0 {
		return PartMeta{}, fmt.Errorf("replay sync time moved backwards")
	}
	return partMeta, nil
}

var errNoReplaySync = errors.New("replay part contains no sync instruction")

func scanPartReplayTime(partFile string) (startTime int64, endTime int64, syncCount int64, err error) {
	fd, err := os.Open(partFile)
	if err != nil {
		return 0, 0, 0, err
	}
	defer fd.Close()
	reader := bufio.NewReader(fd)
	for {
		inst, err1 := ReadInstruction(reader)
		if err1 != nil {
			break
		}
		if inst.Opcode != "sync" {
			continue
		}
		if len(inst.Args) > 0 {
			syncMill, err2 := strconv.ParseInt(inst.Args[0], 10, 64)
			if err2 != nil {
				continue
			}
			endTime = syncMill
			if syncCount == 0 {
				startTime = syncMill
			}
			syncCount++
		}
	}
	if syncCount == 0 {
		return 0, 0, 0, errNoReplaySync
	}
	return startTime, endTime, syncCount, nil
}

func LoadPartReplayTime(partFile string) (startTime int64, endTime int64, err error) {
	startTime, endTime, _, err = scanPartReplayTime(partFile)
	return startTime, endTime, err
}

func loadPartMeta(partFile string) (PartMeta, error) {
	var meta PartMeta
	info, err := os.Stat(partFile)
	if err != nil {
		return meta, err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return meta, fmt.Errorf("replay part is not a non-empty regular file")
	}
	metaPath := partFile + MetaSuffix
	if buf, readErr := os.ReadFile(metaPath); readErr == nil {
		if jsonErr := json.Unmarshal(buf, &meta); jsonErr == nil {
			// sync_count was added after the original format. Old metadata is
			// accepted when it already contains a meaningful time range.
			hasSync := meta.SyncCount > 0 ||
				(meta.StartTime != 0 && meta.EndTime >= meta.StartTime)
			if hasSync && meta.EndTime >= meta.StartTime && meta.Size == info.Size() {
				meta.Duration = meta.EndTime - meta.StartTime
				return meta, nil
			}
		}
	}
	meta, err = LoadPartMetaByFile(partFile)
	if err != nil {
		return meta, err
	}
	metaBuf, marshalErr := json.Marshal(meta)
	if marshalErr != nil {
		return meta, marshalErr
	}
	if writeErr := writeFileAtomically(metaPath, metaBuf, 0o600); writeErr != nil {
		logger.Warnf("Write recovered replay meta file %s failed: %v", metaPath, writeErr)
	}
	return meta, nil
}

func NewWorkerClient(cfg config.Config) *videoworker.WorkClient {
	if !cfg.EnableVideoWorker {
		return nil
	}
	workerURL := cfg.VideoWorkerHost
	var key model.AccessKey
	if err := key.LoadFromFile(cfg.AccessKeyFilePath); err != nil {
		logger.Errorf("Create video worker client failed: loading access key err %s", err)
		return nil
	}
	workClient := videoworker.NewClient(workerURL, key, cfg.IgnoreVerifyCerts)
	if workClient == nil {
		logger.Errorf("Create video worker client failed: worker url %s", workerURL)
		return nil
	}
	return workClient
}
