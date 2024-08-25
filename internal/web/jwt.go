package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

// 长短 token
type jwtHandler struct {
	accessTokenKey  []byte
	refreshTokenKey []byte
}

func newJwtHandler() jwtHandler {
	return jwtHandler{
		accessTokenKey:  []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"),
		refreshTokenKey: []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvfx"),
	}
}

// UserClaims 是自定义的 claims， 用于编码成 JWT
type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

func (h jwtHandler) setLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.setAccessToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = h.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return nil
}

func (h jwtHandler) setAccessToken(ctx *gin.Context, uid int64, ssid string) error {
	// 用 JWT 设置登陆态, 生成一个 JWT token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedToken, err := token.SignedString(h.accessTokenKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误，生成 access-token 失败",
		})
		return err
	}
	ctx.Header("X-Access-Token", signedToken)
	return nil
}

func (h jwtHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := &RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signedToken, err := token.SignedString(h.refreshTokenKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误，生成 refresh-token失败",
		})
	}
	ctx.Header("X-Refresh-Token", signedToken)
	return nil
}

func ExtractToken(ctx *gin.Context) string {
	token := ctx.GetHeader("Authorization")
	segs := strings.Split(token, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}
