package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skcheng003/webook/internal/web"
	"net/http"
	"strings"
	"time"
)

var AccessTokenKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")
var RefreshTokenKey = []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvfx")

type RedisJWTHandler struct {
	redisCmd     redis.Cmdable
	rtExpiration time.Duration
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return RedisJWTHandler{
		redisCmd:     cmd,
		rtExpiration: time.Hour * 24 * 7,
	}
}

func (h RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.SetAccessToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = h.SetRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return nil
}

func (h RedisJWTHandler) SetAccessToken(ctx *gin.Context, uid int64, ssid string) error {
	// 用 JWT 设置登陆态, 生成一个 JWT token
	claims := UserClaims{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedToken, err := token.SignedString(AccessTokenKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.Result{
			Code: 5,
			Msg:  "系统错误，生成 access-token 失败",
		})
		return err
	}
	ctx.Header("X-Access-Token", signedToken)
	return nil
}

func (h RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := &RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedToken, err := token.SignedString(RefreshTokenKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.Result{
			Code: 5,
			Msg:  "系统错误，生成 refresh-token失败",
		})
	}
	ctx.Header("X-Refresh-Token", signedToken)
	return nil
}

func (h RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	logout, err := h.redisCmd.Exists(ctx, h.key(ssid)).Result()
	if err != nil {
		return err
	}
	if logout > 0 {
		return errors.New("用户已经退出登录")
	}
	return nil
}

func (h RedisJWTHandler) ClearSession(ctx *gin.Context) error {
	ctx.Header("X-Access-Token", "")
	ctx.Header("X-Refresh-Token", "")
	uc := ctx.MustGet("userClaims").(UserClaims)
	return h.redisCmd.Set(ctx, h.key(uc.Ssid), "", h.rtExpiration).Err()
}

func (h RedisJWTHandler) key(ssid string) string {
	return fmt.Sprintf("user:ssid:%s", ssid)
}

func (h RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	token := ctx.GetHeader("Authorization")
	segs := strings.Split(token, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}
