package proxy

import (
    "io"
    "lion/pkg/common"
    "lion/pkg/config"
    "lion/pkg/jms-sdk-go/model"
    "lion/pkg/jms-sdk-go/service"
    "lion/pkg/logger"
    "lion/pkg/storage"
    "os"
    "path/filepath"
    "strings"
    "time"
)

const (
    dateTimeFormat = "2006-01-02"
)

type FTPFileRecorder struct {
    FTPLog     *model.FTPLog
    jmsService *service.JMService
    storage    storage.FTPFileStorage

    absFilePath     string
    Target          string
    TargetPrefix    string
    MaxStore        int64
    err             error

    file *os.File
}

type FTPFileInfo struct {
    TimeStamp  time.Time
}

func (r *FTPFileRecorder) PreRecord() (err error) {
    info := &FTPFileInfo{
        TimeStamp: time.Now(),
    }
    today := info.TimeStamp.UTC().Format(dateTimeFormat)
    ftpFileRootDir := config.GlobalConfig.FTPFilePath
    ftpFileDirPath := filepath.Join(ftpFileRootDir, today)
    err = config.EnsureDirExist(ftpFileDirPath)
    if err != nil {
        logger.Errorf("Create dir %s error: %s\n", ftpFileDirPath, err)
        return
    }
    absFilePath := filepath.Join(ftpFileDirPath, r.FTPLog.ID)
    storageTargetName := strings.Join([]string{today, r.FTPLog.ID}, "/")
    r.absFilePath = absFilePath
    r.Target = storageTargetName
    fd, err := os.OpenFile(r.absFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
    if err != nil {
        logger.Errorf("Create FTP file %s error: %s\n", r.absFilePath, err)
        return
    }
    r.file = fd
    return
}

func (r *FTPFileRecorder) Record(ftpLog *model.FTPLog, reader io.Reader) (err error) {
    if r.isNullStorage() {
        return
    }
    r.FTPLog = ftpLog
    err = r.PreRecord()
    if err != nil {
        return
    }
    io.Copy(r.file, reader)
    defer r.file.Close()
    reader.(io.Seeker).Seek(0, io.SeekStart)
    go func () {
        r.uploadFTPFile()
    }()
    return
}

func (r *FTPFileRecorder) RealTarget() string {
    return strings.Join([]string{r.TargetPrefix, r.Target}, "/")
}

func (r *FTPFileRecorder) isNullStorage() bool {
    return r.storage.TypeName() == "null" || r.err != nil
}

func (r *FTPFileRecorder) uploadFTPFile() {
    logger.Infof("FTP log %s: FTP File recorder is uploading", r.FTPLog.ID)
    if !common.FileExists(r.absFilePath) {
        logger.Info("FTP file not found, passed: ", r.absFilePath)
        return
    }
    stat, err := os.Stat(r.absFilePath)
    if err == nil {
        if stat.Size() == 0 {
            logger.Info("FTP file is empty, removed: ", r.absFilePath)
            _ = os.Remove(r.absFilePath)
            return
        } else if stat.Size() >= r.MaxStore * 1024 * 1024 {
            logger.Info("FTP file is exceeds the upper limit for saving files, removed: ", r.absFilePath)
            _ = os.Remove(r.absFilePath)
            return
        }
    }
    r.UploadFile(3)
}

func (r *FTPFileRecorder) UploadFile(maxRetry int) {
    if r.isNullStorage() {
        _ = os.Remove(r.absFilePath)
        return
    }
    for i := 0; i <= maxRetry; i++ {
        logger.Infof("Upload FTP file: %s, type: %s", r.absFilePath, r.storage.TypeName())
        err := r.storage.Upload(r.absFilePath, r.RealTarget())
        if err == nil {
            _ = os.Remove(r.absFilePath)
            if err := r.jmsService.FinishFTPFile(r.FTPLog.ID); err != nil {
                logger.Errorf("FTP file %s upload failed: %s", r.FTPLog.ID, err)
            }
            break
        }
        logger.Errorf("Upload FTP file err: %s", err)
        // 如果还是失败，上传 server 再传一次
        if i == maxRetry {
            if r.storage.TypeName() == "server" {
                break
            }
            logger.Errorf("Session[%s] using server storage retry upload", r.FTPLog.ID)
            r.storage = storage.ServerStorage{StorageType: "server", JmsService: r.jmsService}
            r.UploadFile(3)
            break
        }
    }
    r.file.Close()
}

func (r *FTPFileRecorder) SetFTPLog(ftpLog *model.FTPLog) {
    r.FTPLog = ftpLog
    err := r.PreRecord()
    if err != nil {
        return
    }
}

func (r *FTPFileRecorder) RecordWrite(p []byte) (err error) {
    if r.isNullStorage() {
        return
    }
    go func () {
        _, err = r.file.Write(p)
        if err != nil {
            logger.Errorf("Record write save err: %s", err)
        }
    }()
    return
}

func (r *FTPFileRecorder) Clear() (err error) {
    if r.isNullStorage() {
        return
    }
    err = os.Remove(r.absFilePath)
    return
}

func NewFTPFileRecord(jmsService *service.JMService, storage storage.FTPFileStorage, maxStore int64) (FTPFileRecorder, error) {
    recorder := FTPFileRecorder{
        jmsService:   jmsService,
        storage:      storage,
        TargetPrefix: "FTP_FILES",
        MaxStore:     maxStore,
    }
    return recorder, nil
}

func GetFTPFileRecorder(jmsService *service.JMService) FTPFileRecorder {
    terminalConfig, _ := jmsService.GetTerminalConfig()
    maxStore := terminalConfig.FTPFileMaxStore
    recorder, err := NewFTPFileRecord(jmsService, storage.NewFTPFileStorage(jmsService, terminalConfig.ReplayStorage), maxStore)
    if err != nil {
        logger.Error(err)
    }
    return recorder
}


