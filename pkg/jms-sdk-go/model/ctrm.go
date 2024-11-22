package model

type Ctrm struct {
	Success   bool   `json:"success"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      string `json:"data"`
	Id        int    `json:"id"`
	FaildType string `json:"faildType"`
}
