package common

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewJSONTime(t *testing.T) {
	jsonT := NewUTCTime( time.Now())
	var s struct {
		T UTCTime
	}
	s.T = jsonT
	j, _ := json.Marshal(s)
	t.Logf("%s\n", j)
	var s2 struct {
		T UTCTime
	}
	err := json.Unmarshal(j, &s2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", s2.T)

}
