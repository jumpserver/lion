package videoworker

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"lion/pkg/jms-sdk-go/httplib"
	"lion/pkg/jms-sdk-go/model"
)

const (
	orgHeaderKey   = "X-JMS-ORG"
	orgHeaderValue = "ROOT"
)

func NewClient(baseUrl string, key model.AccessKey, Insecure bool) *Client {
	opts := make([]httplib.Opt, 0, 2)
	if Insecure {
		opts = append(opts, httplib.WithInsecure())
	}
	client, err := httplib.NewClient(baseUrl, 30*time.Second, opts...)
	if err != nil {
		return nil
	}
	sign := ProfileAuth{
		KeyID:    key.ID,
		SecretID: key.Secret,
	}
	client.SetAuthSign(&sign)
	client.SetHeader(orgHeaderKey, orgHeaderValue)
	return &Client{BaseURL: baseUrl, client: client}
}

type Client struct {
	BaseURL string
	client  *httplib.Client
}

func (s *Client) CreateReplayTask(sessionId string, file string, meta ReplayMeta) (model.Task, error) {
	var res model.Task
	fileReplayURL := fmt.Sprintf(ReplayFileURL, sessionId)
	fieldsMap := StructToMapString(meta)
	err := s.client.PostFileWithFields(fileReplayURL, file, fieldsMap, &res)
	if err != nil {
		return res, err
	}
	return res, nil
}

type ReplayMeta struct {
	SessionId     string `json:"session_id"`
	ComponentType string `json:"component_type"`
	FileType      string `json:"file_type"`
	SessionDate   string `json:"session_date"` //  格式是 "2006-01-02"
	MaxFrame      int    `json:"max_frame"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Bitrate       int    `json:"bitrate"` // 1 或者 2
}

const tagName = "json"

// 任意 struct 的 json 标签转换为 map[string]string, 用于 post form, 如果非 struct 对象则返回 nil

func StructToMapString(m interface{}) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string)

	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct { // Non-structural return error
		return nil
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fi := t.Field(i)
		if tagValue := fi.Tag.Get(tagName); tagValue != "" {
			interValue := v.Field(i).Interface()
			fieldValue := ""
			switch interValue1 := interValue.(type) {
			case string:
				fieldValue = interValue1
			case int:
				fieldValue = strconv.Itoa(interValue1)
			case int32:
				fieldValue = strconv.FormatInt(int64(interValue1), 10)
			case int64:
				fieldValue = strconv.FormatInt(interValue1, 10)
			case float64:
				fieldValue = strconv.FormatFloat(interValue1, 'f', -1, 64)
			case bool:
				fieldValue = strconv.FormatBool(interValue1)
			default:
				fieldValue = fmt.Sprintf("%v", interValue1)
			}
			// 如果值为空或者为0则不传递
			if fieldValue == "" || fieldValue == "0" {
				continue
			}
			out[tagValue] = fieldValue
		}
	}
	return out

}

const (
	ReplayFileURL = "/api/v2/replay/sessions/%s/task/"
)
