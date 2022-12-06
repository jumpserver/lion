package service

import (
	"testing"

	"lion/pkg/jms-sdk-go/httplib"
)

func setup() *JMService {
	auth := httplib.SigAuth{
		KeyID:    "25bfc52e-48de-4c0c-9b1c-44c79aeb238a",
		SecretID: "e275a9d9-2c9b-4823-be41-a9012d5cd0c3",
	}
	jms, err := NewAuthJMService(JMSAccessKey(auth.KeyID, auth.SecretID),
		JMSCoreHost("http://10.1.88.5:8080"))
	if err != nil {
		panic(err)
	}
	return jms
}

func TestJMService_GetProfile(t *testing.T) {
	jms := setup()
	user, err := jms.GetProfile()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", user)

}

func TestJMService_GetAssetById(t *testing.T) {
	jms := setup()
	assetIds := []string{
		"2e73f0e4-13ec-4f64-b03e-4ecbadab7233", // 有网域
		"bd87e0b9-9a94-48df-9fa1-4aab4c9f49a5", // 无网域
	}
	for i := range assetIds {
		asset, err := jms.GetAssetById(assetIds[i])
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v\n", asset)
	}

}

func TestJMService_GetDomainGateways(t *testing.T) {
	jms := setup()
	domains := []string{
		"aad81461-5f10-40f6-9064-ed6de855d0c7",
	}
	for i := range domains {
		asset, err := jms.GetDomainGateways(domains[i])
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v\n", asset)
	}
}
