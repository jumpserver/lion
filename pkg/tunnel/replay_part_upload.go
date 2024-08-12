package tunnel

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lion/pkg/common"
	"lion/pkg/guacd"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
	"lion/pkg/logger"
	"lion/pkg/storage"
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
		logger.Errorf("PartUploader %s unmarshal meta file error: %v", p.SessionId, err)
		return err1
	}
	if p.replayMeta.DateStart == p.replayMeta.DateEnd {
		// 未结束的录像, 计算结束时间，并上传到 core api 作为会话结束时间
		endTime := GetMaxModTime(p.partFiles)
		p.replayMeta.DateEnd = common.NewUTCTime(endTime)
		// api finish time
		if err1 := p.ApiClient.SessionFinished(p.SessionId, p.replayMeta.DateEnd); err1 != nil {
			logger.Errorf("PartUploader %s finish session error: %v", p.SessionId, err1)
			return err
		}
		// write meta file
		metaBuf, _ = json.Marshal(p.replayMeta)
		if err1 := os.WriteFile(metaPath, metaBuf, os.ModePerm); err1 != nil {
			logger.Errorf("PartUploader %s write meta file error: %v", p.SessionId, err1)
		}
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
	p.CollectionPartFiles()
	if err := p.preCheckSessionMeta(); err != nil {
		return
	}
	if len(p.partFiles) == 0 {
		logger.Errorf("PartUploader %s no part file", p.SessionId)
		return
	}
	// 1、创建 upload 目录
	uploadPath := filepath.Join(p.RootPath, "upload")
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		logger.Errorf("PartUploader %s create upload dir error: %v", p.SessionId, err)
		return
	}
	// 2、将所有的 part 文件压缩移动到 upload 目录
	for i := range p.partFiles {
		partFile := p.partFiles[i]
		partFilePath := filepath.Join(p.RootPath, partFile.Name())
		partGzFilename := partFile.Name() + ".gz"
		uploadFilePath := filepath.Join(uploadPath, partGzFilename)

		if err := common.CompressToGzipFile(partFilePath, uploadFilePath); err != nil {
			logger.Errorf("PartUploader %s compress part file %s error: %v", p.SessionId, partFile.Name(), err)
			return
		}

		// 3、生成新的 meta 文件

		partFileMeta := PartFileMeta{Name: partGzFilename}
		// 读取 {part}.meta 文件
		if buf, err := os.ReadFile(filepath.Join(p.RootPath, partFile.Name()+".meta")); err == nil {
			_ = json.Unmarshal(buf, &partFileMeta.PartMeta)
		} else {
			meta, err1 := LoadPartMetaByFile(partFilePath)
			if err1 != nil {
				logger.Errorf("PartUploader %s load part file %s meta error: %v", p.SessionId, partFile.Name(), err1)
				return
			}
			// 存储一份 meta 文件
			metaBuf, _ := json.Marshal(meta)
			_ = os.WriteFile(filepath.Join(uploadPath, partFile.Name()+".meta"), metaBuf, os.ModePerm)
			partFileMeta.PartMeta = meta
		}
		p.replayMeta.PartMetas = append(p.replayMeta.PartMetas, partFileMeta)
	}
	// upload 写入 replayMeta json
	replayMetaBuf, _ := json.Marshal(p.replayMeta)
	if err := os.WriteFile(filepath.Join(uploadPath, p.SessionId+".replay.json"), replayMetaBuf, os.ModePerm); err != nil {
		logger.Errorf("PartUploader %s write replay meta file error: %v", p.SessionId, err)
		return
	}
	// 4、上传 upload 目录下的所有文件到 存储
	p.uploadToStorage(uploadPath)
}

func (p *PartUploader) CollectionPartFiles() {
	entries, err := os.ReadDir(p.RootPath)
	if err != nil {
		logger.Errorf("PartUploader %s read dir %s error: %v", p.SessionId, p.RootPath, err)
		return
	}
	p.partFiles = make([]os.DirEntry, 0, 5)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".part") {
			p.partFiles = append(p.partFiles, entry)
		}
	}
}

func (p *PartUploader) GetStorage() storage.ReplayStorage {
	return storage.NewReplayStorage(p.ApiClient, p.TermCfg.ReplayStorage)
}

const recordDirTimeFormat = "2006-01-02"

func (p *PartUploader) uploadToStorage(uploadPath string) {
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
	for _, uploadFile := range uploadFiles {
		if uploadFile.IsDir() {
			continue
		}
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
	if err = p.ApiClient.FinishReply(p.SessionId); err != nil {
		logger.Errorf("PartUploader %s finish replay error: %v", p.SessionId, err)
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
	startTime, endTime, err := LoadPartReplayTime(partFile)
	if err != nil {
		logger.Errorf("LoadPartMetaByFile %s load replay time error: %v", partFile, err)
		return partMeta, err
	}
	partMeta.StartTime = startTime
	partMeta.EndTime = endTime
	partMeta.Duration = endTime - startTime
	return partMeta, nil
}

func LoadPartReplayTime(partFile string) (startTime int64, endTime int64, err error) {
	fd, err := os.Open(partFile)
	if err != nil {
		return 0, 0, err
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
			if startTime == 0 {
				startTime = syncMill
			}
		}
	}
	return startTime, endTime, nil
}
