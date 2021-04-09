package common

import (
	"testing"
)

func TestCurrentLocalIP(t *testing.T) {
	ip := CurrentLocalIP()
	t.Logf("%+v", ip)
}
