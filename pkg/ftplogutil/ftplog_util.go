package ftplogutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"lion/pkg/config"
	"lion/pkg/jms-sdk-go/model"
)

func GetFileCacheRootPath() string {
	conf := config.GlobalConfig
	return filepath.Join(conf.Root, "data", "files")
}

func GetFileCachePath(ftpLog *model.FTPLog) (string, error) {
	directory := filepath.Join(GetFileCacheRootPath(), fmt.Sprintf(`"%s"`, ftpLog.DataStart)[1:11])
	err := config.EnsureDirExist(directory)
	if err != nil {
		return directory, err
	}

	fileCacheDir := filepath.Join(directory, ftpLog.Id)
	return fileCacheDir, nil
}

func CacheFileLocally(ftpLog *model.FTPLog, reader io.Reader) (string, error) {
	path, err := GetFileCachePath(ftpLog)
	if err != nil {
		return path, err
	}

	localDst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return path, err
	}
	defer localDst.Close()
	go io.Copy(localDst, reader)
	SendNotifyFileReady(*ftpLog)
	return path, err
}
