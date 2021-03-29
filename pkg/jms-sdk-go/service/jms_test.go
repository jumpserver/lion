package service

import (
	"guacamole-client-go/pkg/jms-sdk-go/httplib"
	"testing"
)

func setup() *JMService {
	auth := httplib.SigAuth{
		KeyID:    "25bfc52e-48de-4c0c-9b1c-44c79aeb238a",
		SecretID: "e275a9d9-2c9b-4823-be41-a9012d5cd0c3",
	}
	jms, err := NewAuthJMService(JMSAuthSign(&auth),
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
