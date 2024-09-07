package middleware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	jwt2 "github.com/skcheng003/webook/internal/web/jwt"
	"net/http"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
	jwt2.Handler
}

func NewLoginJWTMiddlewareBuilder(hdl jwt2.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: hdl,
	}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePath(path ...string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path...)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	// 用 Go 的方式编码解码
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 注册和登陆不需要进行校验
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		signedToken := l.ExtractToken(ctx)
		claims := &jwt2.UserClaims{}
		token, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// token.Valid 会验证过期时间
		if token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		//  验证发送客户端
		if claims.UserAgent != ctx.Request.UserAgent() {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		// 查询当前 session 是否已经退出
		err = l.CheckSession(ctx, claims.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 把解析后的 claim 放在 context 里面，方便其他路由函数获取
		ctx.Set("userClaims", claims)
	}
}
