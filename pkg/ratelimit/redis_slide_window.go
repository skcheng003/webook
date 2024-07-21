package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var slideWindow string

type RedisSlidingWindowLimiter struct {
	cmd redis.Cmdable
	// 阈值
	rate int
	// 窗口大小
	interval time.Duration
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, rate int, interval time.Duration) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		rate:     rate,
		interval: interval,
	}
}

func (r RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, slideWindow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}
