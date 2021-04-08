package storage

import (
	"strings"
)

type ReplayStorage interface {
	Upload(gZipFile, target string) error
}

func NewReplayStorage(cf map[string]interface{}) ReplayStorage {
	tp, ok := cf["TYPE"]
	if !ok {
		tp = "server"
	}
	switch tp {
	case "azure":
		var accountName string
		var accountKey string
		var containerName string
		var endpointSuffix string
		if value, ok := cf["ENDPOINT_SUFFIX"].(string); ok {
			endpointSuffix = value
		}
		if value, ok := cf["ACCOUNT_NAME"].(string); ok {
			accountName = value
		}
		if value, ok := cf["ACCOUNT_KEY"].(string); ok {
			accountKey = value
		}
		if value, ok := cf["CONTAINER_NAME"].(string); ok {
			containerName = value
		}
		if endpointSuffix == "" {
			endpointSuffix = "core.chinacloudapi.cn"
		}
		return AzureReplayStorage{
			AccountName:    accountName,
			AccountKey:     accountKey,
			ContainerName:  containerName,
			EndpointSuffix: endpointSuffix,
		}
	case "oss":
		var endpoint string
		var bucket string
		var accessKey string
		var secretKey string

		if value, ok := cf["ENDPOINT"].(string); ok {
			endpoint = value
		}
		if value, ok := cf["BUCKET"].(string); ok {
			bucket = value
		}
		if value, ok := cf["ACCESS_KEY"].(string); ok {
			accessKey = value
		}
		if value, ok := cf["SECRET_KEY"].(string); ok {
			secretKey = value
		}
		return OSSReplayStorage{
			Endpoint:  endpoint,
			Bucket:    bucket,
			AccessKey: accessKey,
			SecretKey: secretKey,
		}
	case "s3", "swift":
		var region string
		var endpoint string
		var bucket string
		var accessKey string
		var secretKey string
		if value, ok := cf["BUCKET"].(string); ok {
			bucket = value
		}
		if value, ok := cf["ENDPOINT"].(string); ok {
			endpoint = value
		}
		if value, ok := cf["REGION"].(string); ok {
			region = value
		}
		if value, ok := cf["ACCESS_KEY"].(string); ok {
			accessKey = value
		}
		if value, ok := cf["SECRET_KEY"].(string); ok {
			secretKey = value
		}
		if region == "" && endpoint != "" {
			endpointArray := strings.Split(endpoint, ".")
			if len(endpointArray) >= 2 {
				region = endpointArray[1]
			}
		}
		if bucket == "" {
			bucket = "jumpserver"
		}
		return S3ReplayStorage{
			Bucket:    bucket,
			Region:    region,
			AccessKey: accessKey,
			SecretKey: secretKey,
			Endpoint:  endpoint,
		}
	default:
		return nil
	}
}
