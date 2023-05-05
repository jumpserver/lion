package videoworker

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

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
	sign := httplib.SigAuth{
		KeyID:    key.ID,
		SecretID: key.Secret,
	}
	client.SetAuthSign(&sign)
	client.SetHeader(orgHeaderKey, orgHeaderValue)
	wsDialer := createWsClientDialer(Insecure)
	wsDialer.Jar = client.Jar
	return &Client{BaseURL: baseUrl, client: client, wsDialer: wsDialer}
}

type Client struct {
	BaseURL  string
	sign     httplib.AuthSign
	client   *httplib.Client
	wsDialer *websocket.Dialer

	cacheToken map[string]interface{}
}

func (s *Client) Login() error {
	var res map[string]interface{}
	_, err := s.client.Get(ProfileURL, &res)
	if err != nil {
		return err
	}
	s.cacheToken = res
	return nil
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
			switch interValue.(type) {
			case string:
				fieldValue = interValue.(string)
			case int:
				fieldValue = strconv.Itoa(interValue.(int))
			case int32:
				fieldValue = strconv.FormatInt(int64(interValue.(int32)), 10)
			case int64:
				fieldValue = strconv.FormatInt(interValue.(int64), 10)
			case float64:
				fieldValue = strconv.FormatFloat(interValue.(float64), 'f', -1, 64)
			case bool:
				fieldValue = strconv.FormatBool(interValue.(bool))
			default:
				fieldValue = fmt.Sprintf("%v", interValue)
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
	ProfileURL    = "/api/v1/users/profile/"
	ReplayFileURL = "/api/v2/replay/sessions/%s/task/"
	wsURL         = "ws/events/"
)

func (s *Client) GetWsClient() (*websocket.Conn, error) {
	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return nil, err
	}
	scheme := "ws"
	switch u.Scheme {
	case "http":
		scheme = "ws"
	case "https":
		scheme = "wss"
	default:
		scheme = "ws"
	}
	wsReqURL := url.URL{Scheme: scheme, Host: u.Host, Path: wsURL}
	req, err := http.NewRequest(http.MethodGet, wsReqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	header := req.Header
	c, _, err := s.wsDialer.Dial(wsReqURL.String(), header)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func createWsClientDialer(insecure bool) *websocket.Dialer {
	var tlsCfg *tls.Config
	if insecure {
		tlsCfg = &tls.Config{InsecureSkipVerify: true}
	}
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  tlsCfg,
		Subprotocols:     []string{"JMS-Video-Worker"},
	}
	return dialer
}
