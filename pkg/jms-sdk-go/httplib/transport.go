package httplib

import (
	"crypto/tls"
	"net/http"
)

// 创建 Transport 支持使用不安全的 https

func NewTransport(insecure bool) http.RoundTripper {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return transport
}
