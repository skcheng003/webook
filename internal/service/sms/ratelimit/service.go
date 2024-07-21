package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/skcheng003/webook/internal/service/sms"
	"github.com/skcheng003/webook/pkg/ratelimit"
)

const key = "tencent_sms"

var errLimited = errors.New("短信服务触发限流")

type RateLimitSMSService struct {
	delegate sms.Service
	limiter  ratelimit.Limiter
}

func NewRateLimitSMSService(delegate sms.Service, limiter ratelimit.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		delegate: delegate,
		limiter:  limiter,
	}
}

func (s *RateLimitSMSService) Send(ctx context.Context, tplId string,
	args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流异常 %w", err)
	}
	if limited {
		return errLimited
	}
	return s.delegate.Send(ctx, tplId, args, numbers...)
}
