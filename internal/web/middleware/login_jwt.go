package middleware

import (
	"encoding/gob"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/skcheng003/webook/internal/web"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
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

		tokenHeader := ctx.GetHeader("Authorization")
		// token 为空，没登陆
		if tokenHeader == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		strs := strings.Split(tokenHeader, " ")
		if len(strs) != 2 {
			// 没登录，或者有人瞎搞
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenStr := strs[1]
		claims := &web.UserClaims{}
		// ParseWithClaims 里面一定要传指针
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
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

		if claims.UserAgent != ctx.Request.UserAgent() {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		now := time.Now()
		// 过期时间小于一分钟时，刷新 jwt
		if claims.ExpiresAt.Sub(now) < time.Minute {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
			// 生成一个新的 token
			tokenStr, err = token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
			if err != nil {
				// TODO: need log module
				log.Println("regenerate jwt token failed: ", err)
			}
			ctx.Header("x-jwt-token", tokenStr)
		}
		// 把解析后的 claim 放在 context 里面，方便其他路由函数获取
		ctx.Set("claims", claims)
	}
}
