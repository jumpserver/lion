package storage

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"guacamole-client-go/pkg/logger"
)

type OSSReplayStorage struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
}

func (o OSSReplayStorage) Upload(gZipFilePath, target string) (err error) {
	client, err := oss.New(o.Endpoint, o.AccessKey, o.SecretKey)
	if err != nil {
		logger.Errorf("OSS new err: %s", err)
		return
	}
	bucket, err := client.Bucket(o.Bucket)
	if err != nil {
		logger.Errorf("OSS create bucket err: %s", err)
		return err
	}
	return bucket.PutObjectFromFile(target, gZipFilePath)
}

func (o OSSReplayStorage) TypeName() string {
	return "oss"
}
