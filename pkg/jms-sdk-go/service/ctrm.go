package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"lion/pkg/common"
	"lion/pkg/jms-sdk-go/httplib"
	"lion/pkg/jms-sdk-go/model"
	"time"
)

func CheckCtrmActive(Host string, sk string, account string, lastActiveTime string) (ctrm model.Ctrm, err error) {
	client, err := httplib.NewClient(Host, time.Second*30)
	if err != nil {
		return model.Ctrm{}, err
	}
	random := common.RandomStr(8)
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	token := md5.Sum([]byte(sk + random + timestamp))
	client.SetHeader("timestamp", timestamp)
	client.SetHeader("token", hex.EncodeToString(token[:]))
	client.SetHeader("random", random)
	data := map[string]interface{}{
		"data": []map[string]string{
			{
				"account":        account,
				"lastActiveTime": lastActiveTime,
			},
		},
	}
	_, err = client.Post("", data, &ctrm)
	return
}
