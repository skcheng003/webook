package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetAccessToken(ctx *gin.Context, uid int64, ssid string) error
	SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error
	CheckSession(ctx *gin.Context, ssid string) error
	ClearSession(ctx *gin.Context) error
	ExtractToken(ctx *gin.Context) string
}

type UserClaims struct {
	Uid  int64
	Ssid string
	jwt.RegisteredClaims
	UserAgent string
}
type RefreshClaims struct {
	Uid  int64
	Ssid string
	jwt.RegisteredClaims
}
