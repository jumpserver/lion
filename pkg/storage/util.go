package storage

import (
	"net/url"
	"strings"

	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
)

type StorageType interface {
	TypeName() string
}

type Storage interface {
	Upload(gZipFile, target string) error
	StorageType
}

type FTPFileStorage interface {
	Storage
}

type ReplayStorage interface {
	Storage
}

type CommandStorage interface {
	BulkSave(commands []*model.Command) error
	StorageType
}

func GetStorage(cfg model.ReplayConfig) Storage {
	switch cfg.TypeName {
	case "azure":
		var (
			accountName    string
			accountKey     string
			containerName  string
			endpointSuffix string
		)
		endpointSuffix = cfg.EndpointSuffix
		accountName = cfg.AccountName
		accountKey = cfg.AccountKey
		containerName = cfg.ContainerName

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
		var (
			endpoint  string
			bucket    string
			accessKey string
			secretKey string
		)

		endpoint = cfg.Endpoint
		bucket = cfg.Bucket
		accessKey = cfg.AccessKey
		secretKey = cfg.SecretKey

		return OSSReplayStorage{
			Endpoint:  endpoint,
			Bucket:    bucket,
			AccessKey: accessKey,
			SecretKey: secretKey,
		}
	case "s3", "swift", "cos":
		var (
			region    string
			endpoint  string
			bucket    string
			accessKey string
			secretKey string
		)
		bucket = cfg.Bucket
		endpoint = cfg.Endpoint
		region = cfg.Region
		accessKey = cfg.AccessKey
		secretKey = cfg.SecretKey

		if region == "" && endpoint != "" {
			region = ParseEndpointRegion(endpoint)
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
	case "obs":
		var (
			endpoint  string
			bucket    string
			accessKey string
			secretKey string
		)

		endpoint = cfg.Endpoint
		bucket = cfg.Bucket
		accessKey = cfg.AccessKey
		secretKey = cfg.SecretKey

		return OBSReplayStorage{
			Endpoint:  endpoint,
			Bucket:    bucket,
			AccessKey: accessKey,
			SecretKey: secretKey,
		}
	default:
		return nil
	}
}

func NewReplayStorage(jmsService *service.JMService, cfg model.ReplayConfig) ReplayStorage {
	replayStorage := GetStorage(cfg)
	if replayStorage == nil {
		replayStorage = ServerStorage{StorageType: "server", JmsService: jmsService}
	}
	return replayStorage
}

func NewFTPFileStorage(jmsService *service.JMService, cfg model.ReplayConfig) FTPFileStorage {
	ftpStorage := GetStorage(cfg)
	if ftpStorage == nil {
		ftpStorage = FTPServerStorage{StorageType: "server", JmsService: jmsService}
	}
	return ftpStorage
}

func NewCommandStorage(jmsService *service.JMService, conf *model.TerminalConfig) CommandStorage {
	cf := conf.CommandStorage
	tp, ok := cf["TYPE"]
	if !ok {
		tp = "server"
	}
	/*
		{
		'DOC_TYPE': 'command',
		  'HOSTS': ['http://172.16.10.122:9200'],
		  'INDEX': 'jumpserver',
		  'OTHER': {'IGNORE_VERIFY_CERTS': True},
		  'TYPE': 'es'
		}
	*/
	switch tp {
	case "es", "elasticsearch":
		var hosts = make([]string, len(cf["HOSTS"].([]interface{})))
		for i, item := range cf["HOSTS"].([]interface{}) {
			hosts[i] = item.(string)
		}
		var skipVerify bool
		index := cf["INDEX"].(string)
		docType := cf["DOC_TYPE"].(string)
		if otherMap, ok := cf["OTHER"].(map[string]interface{}); ok {
			if insecureSkipVerify, ok := otherMap["IGNORE_VERIFY_CERTS"]; ok {
				skipVerify = insecureSkipVerify.(bool)
			}
		}
		if index == "" {
			index = "jumpserver"
		}
		if docType == "" {
			docType = "_doc"
		}
		return ESCommandStorage{
			Hosts:              hosts,
			Index:              index,
			DocType:            docType,
			InsecureSkipVerify: skipVerify,
		}
	case "null":
		return NewNullStorage()
	default:
		return ServerStorage{StorageType: "server", JmsService: jmsService}
	}
}

func ParseEndpointRegion(s string) string {
	if strings.Contains(s, amazonawsSuffix) {
		return ParseAWSURLRegion(s)
	}
	endpoint, err := url.Parse(s)
	if err != nil {
		return ""
	}
	endpoints := strings.Split(endpoint.Hostname(), ".")
	if len(endpoints) >= 3 {
		return endpoints[len(endpoints)-3]
	}
	return ""
}

func ParseAWSURLRegion(s string) string {
	endpoint, err := url.Parse(s)
	if err != nil {
		return ""
	}
	s = endpoint.Hostname()
	s = strings.TrimSuffix(s, amazonawsCNSuffix)
	s = strings.TrimSuffix(s, amazonawsSuffix)
	regions := strings.Split(s, ".")
	return regions[len(regions)-1]
}

const (
	amazonawsCNSuffix = ".amazonaws.com.cn"
	amazonawsSuffix   = ".amazonaws.com"
)
