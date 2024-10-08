package web

import (
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/service"
	jwt2 "github.com/skcheng003/webook/internal/web/jwt"
	"net/http"
)

const (
	nicknameSize = 16
	bioSize      = 256
)

var ErrUserDuplicateEmail = service.ErrUserDuplicateEmail
var ErrUserNoFound = service.ErrUserNoFound
var ErrInvalidUserOrPassword = service.ErrInvalidUserOrPassword

// UserHandler 定义和用户有关的路由
type UserHandler struct {
	svc              service.UserService
	codeSvc          service.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	birthRegexExp    *regexp.Regexp
	jwt2.Handler
	cmd redis.Cmdable
}

func NewUserHandler(userSvc service.UserService, codeSvc service.CodeService, jwtHdl jwt2.Handler) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
		birthdayRegexPattern = `\d{4}-\d{2}-\d{2}`
	)

	return &UserHandler{
		svc:              userSvc,
		codeSvc:          codeSvc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		birthRegexExp:    regexp.MustCompile(birthdayRegexPattern, regexp.None),
		Handler:          jwtHdl,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.VerifyLoginSMSCode)
	ug.POST("/refresh_token", u.RefreshToken)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email" binding:"required"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required"`
	}
	var req SignUpReq
	// Bind 根据 Content-Type 解析数据到 req 里面，如果解析错误，返回 400
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 邮箱校验
	isEmail, err := u.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return
	}

	// 密码校验
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不一致")
		return
	}
	isPassword, err := u.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符，并且长度不能小于8位")
		return
	}

	// 调用 service 进行注册
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if errors.Is(err, ErrUserDuplicateEmail) {
		ctx.String(http.StatusOK, "邮箱地址冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, ErrUserNoFound) || errors.Is(err, ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	sess := sessions.Default(ctx)
	// 在session中放值
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		// Secure: true,
		// HttpOnly: true,
		MaxAge: 30 * 60,
	})
	_ = sess.Save()
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	// TODO: Log out function
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, ErrUserNoFound) || errors.Is(err, ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Msg: "设置登录token失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
	return
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birth    string `json:"birth"`
		Bio      string `json:"bio"`
	}
	signedToken := u.ExtractToken(ctx)
	var claims jwt2.UserClaims
	token, err := jwt.ParseWithClaims(signedToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwt2.AccessTokenKey, nil
	})
	if err != nil || token == nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 5,
			Msg:  "验证token失败",
		})
		return
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	if len(req.Nickname) > nicknameSize {
		ctx.String(http.StatusOK, "nickname too long")
		return
	}

	if len(req.Bio) > bioSize {
		ctx.String(http.StatusOK, "bio too long")
		return
	}

	isBirth, err := u.birthRegexExp.MatchString(req.Birth)

	if err != nil {
		ctx.String(http.StatusOK, "system error")
		return
	}

	if !isBirth {
		ctx.String(http.StatusOK, "birthday format wrong")
		return
	}

	err = u.svc.EditProfile(ctx, domain.User{
		Id:       claims.Uid,
		Nickname: req.Nickname,
		Birth:    req.Birth,
		Bio:      req.Bio,
	})

	if errors.Is(err, ErrUserNoFound) {
		ctx.String(http.StatusOK, "system error, user no found")
		return
	}

	ctx.String(http.StatusOK, "Update profile successful!")
	return
}

func (u *UserHandler) Profile(ctx *gin.Context) {

	type ProfileReq struct {
		Email string `form:"email"`
	}

	var req ProfileReq

	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "system error")
		return
	}

	user, err := u.svc.FindProfile(ctx, req.Email)

	if errors.Is(err, ErrUserNoFound) {
		ctx.String(http.StatusOK, "profile not exist")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "system error")
		return
	}

	ctx.String(http.StatusOK, "nickname: %s, birthday: %s, bio: %s", user.Nickname, user.Birth, user.Bio)
	return
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	signedToken := u.ExtractToken(ctx)
	var claims jwt2.UserClaims
	token, err := jwt.ParseWithClaims(signedToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwt2.AccessTokenKey, nil
	})
	if err != nil || token == nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 5,
			Msg:  "验证token失败",
		})
		return
	}
	user, err := u.svc.FindProfileJWT(ctx, claims.Uid)

	if errors.Is(err, ErrUserNoFound) {
		ctx.String(http.StatusOK, "profile not exist")
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.String(http.StatusOK, "nickname: %s, birthday: %s, bio: %s", user.Nickname, user.Birth, user.Bio)
	return
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		return
	}
	// TODO: 对手机号进行校验，使用正则表达式
	const biz = "user/login"
	err := u.codeSvc.Send(ctx, biz, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "发送成功",
	})
}

func (u *UserHandler) VerifyLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	const biz = "user/login"
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码有误",
		})
	}

	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 4,
		Msg:  "校验验证码通过",
	})
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := u.ClearSession(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "退出登录失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录成功",
	})
}

func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	refreshToken := u.ExtractToken(ctx)
	var rc jwt2.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return jwt2.RefreshTokenKey, nil
	})
	if err != nil || token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = u.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", rc.Ssid)).Err()
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 刷新 AccessToken
	err = u.SetAccessToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "刷新成功",
	})
}
