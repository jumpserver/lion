package session

import (
	"testing"

	"github.com/jumpserver-dev/sdk-go/model"
	"github.com/jumpserver-dev/sdk-go/service/panda"
)

func TestPandaClientForProvider(t *testing.T) {
	defaultClient := &panda.Client{BaseURL: "http://default-panda:9001"}
	server := Server{
		PandaClient: defaultClient,
		PandaClientFactory: func(serviceURL string) *panda.Client {
			return &panda.Client{BaseURL: serviceURL}
		},
	}

	if got := server.pandaClientFor(nil); got != defaultClient {
		t.Fatal("legacy virtual app must use the default Panda client")
	}
	provider := &model.VirtualAppProvider{ServiceURL: "https://remote-panda.example"}
	if got := server.pandaClientFor(provider); got.BaseURL != provider.ServiceURL {
		t.Fatalf("provider Panda URL = %q, want %q", got.BaseURL, provider.ServiceURL)
	}
}
