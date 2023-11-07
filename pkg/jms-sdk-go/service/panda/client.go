package panda

import (
	"time"

	"lion/pkg/jms-sdk-go/httplib"
	"lion/pkg/jms-sdk-go/model"
)

const (
	orgHeaderKey   = "X-JMS-ORG"
	orgHeaderValue = "ROOT"
)

func NewClient(baseUrl string, key model.AccessKey, insecure bool) *Client {
	opts := make([]httplib.Opt, 0, 2)
	if insecure {
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
	sign    httplib.AuthSign
	client  *httplib.Client
}

func (c *Client) CreateContainer(token string, virtualAPPOption model.VirtualAppOption) (ret model.VirtualAppContainer, err error) {
	data := map[string]interface{}{
		"token":       token,
		"virtual_app": virtualAPPOption,
	}
	var res Response
	_, err = c.client.Post(ContainerCreateURL, data, &res)
	return res.Data, err
}

func (c *Client) ReleaseContainer(id string) error {
	data := map[string]interface{}{
		"container_id": id,
	}
	_, err := c.client.Post(ContainerReleaseURL, data, nil)
	return err
}

const (
	ContainerCreateURL  = "/panda/container/create/"
	ContainerReleaseURL = "/panda/container/release/"
)

type Response struct {
	Success bool                      `json:"success"`
	Msg     string                    `json:"message"`
	Data    model.VirtualAppContainer `json:"data"`
}
