package panda

import (
	"testing"

	"lion/pkg/jms-sdk-go/model"
)

func TestNewClient(t *testing.T) {
	key := model.AccessKey{
		ID:     "8298f537-12c2-4c7c-aac1-5330b5f46a46",
		Secret: "6323d771-e714-44ca-a1e8-d4771978913d",
		//Secret: "6323d771-e714-44ca-a1e8-d4771978913ds",
	}
	client := NewClient("http://localhost:9001", key, false)
	if client == nil {
		t.Error("client is nil")
	}
	token := "asdasdsdaasd"
	ret, err := client.CreateContainer(token, model.VirtualAppOption{})
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", ret)
	err = client.ReleaseContainer("asdasdasd")
}
