package panda

import (
	"net/http"

	"github.com/LeeEirc/httpsig"
)

const (
	signHeaderRequestTarget = "(request-target)"
	signHeaderDate          = "date"
	signAlgorithm           = "hmac-sha256"
)

type ProfileAuth struct {
	KeyID    string
	SecretID string
}

func (auth *ProfileAuth) Sign(r *http.Request) error {
	profileReq, err := http.NewRequest(http.MethodGet, UserProfileURL, nil)
	if err != nil {
		return err
	}
	headers := []string{signHeaderRequestTarget, signHeaderDate}
	signer, err := httpsig.NewRequestSigner(auth.KeyID, auth.SecretID, signAlgorithm)
	if err != nil {
		return err
	}
	err = signer.SignRequest(profileReq, headers, nil)
	if err != nil {
		return err
	}
	for k, v := range profileReq.Header {
		r.Header[k] = v
	}
	return nil
}

const UserProfileURL = "/api/v1/users/profile/"
