package ftplogutil

import "lion/pkg/jms-sdk-go/service"

func Initial(jmsService *service.JMService) {
	go ftpLogFileRecord(jmsService)
}
