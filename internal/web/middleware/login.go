package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePath(path ...string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path...)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 注册和登陆不需要进行校验
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		sess := sessions.Default(ctx)
		if sess.Get("userId") == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("updateTime")
		sess.Options(sessions.Options{
			MaxAge: 60,
		})
		now := time.Now()
		// 刚登陆，还没刷新过
		if updateTime == nil {
			sess.Set("updateTime", now)
			sess.Save()
			return
		}

		// 判断是否需要刷新
		timeVal, _ := updateTime.(time.Time)
		if now.Sub(timeVal) > time.Minute {
			sess.Set("updateTime", now)
			sess.Save()
			return
		}

	}
}
