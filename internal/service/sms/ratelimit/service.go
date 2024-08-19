package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"github.com/skcheng003/webook/internal/service/sms"
	"github.com/skcheng003/webook/pkg/ratelimit"
)

// 装饰器模式
// 开闭原则：对修改闭合，对扩展开放
// 非侵入式：不修改已有代码

const key = "tencent_sms"

// 暂时不对外
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
		// 系统错误
		// 可以限流：保守策略，下游很弱
		// 可以不限：容错策略，下游很强，业务可用性要求很高
		return fmt.Errorf("短信服务判断是否限流异常 %w", err)
	}
	if limited {
		return errLimited
	}
	return s.delegate.Send(ctx, tplId, args, numbers...)
}
