package common

import (
	"testing"
)

func TestRandomStr(t *testing.T) {
	str := RandomStr(10)
	t.Log(str)
}
