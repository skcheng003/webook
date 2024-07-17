package ioc

import (
	"github.com/skcheng003/webook/internal/service/sms"
	"github.com/skcheng003/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
