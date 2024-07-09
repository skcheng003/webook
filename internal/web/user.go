package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/service"
	"net/http"
	"time"
)

const (
	nicknameSize = 16
	bioSize      = 256
)

var ErrUserDuplicateEmail = service.ErrUserDuplicateEmail
var ErrUserNoFound = service.ErrUserNoFound
var ErrInvalidUserOrPassword = service.ErrInvalidUserOrPassword

// UserHandler is the handler of user
type UserHandler struct {
	svc              *service.UserService
	codeSvc          *service.SMSCodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	birthRegexExp    *regexp.Regexp
}

// UserClaims is the custom claims struct that will be encoded to a JWT.
type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

func NewUserHandler(svc *service.UserService, codeSvc *service.SMSCodeService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
		birthdayRegexPattern = `\d{4}-\d{2}-\d{2}`
	)

	return &UserHandler{
		svc:              svc,
		codeSvc:          codeSvc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		birthRegexExp:    regexp.MustCompile(birthdayRegexPattern, regexp.None),
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms/code/verify", u.VerifyLoginSMSCode)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {

	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := u.emailRegexExp.MatchString(req.Email)

	if err != nil {
		ctx.String(http.StatusOK, "system error")
		return
	}

	if !isEmail {
		ctx.String(http.StatusOK, "email address format wrong")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "different password")
		return
	}

	isPassword, err := u.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "system error")
		return
	}

	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符，并且长度不能小于8位")
		return
	}

	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if errors.Is(err, ErrUserDuplicateEmail) {
		ctx.String(http.StatusOK, "conflict email address")
		return
	}

	ctx.String(http.StatusOK, "Sign up successful!")
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
		ctx.String(http.StatusOK, "username or password wrong")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "system error")
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
	sess.Save()
	ctx.String(http.StatusOK, "Log in successful!")
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
		ctx.String(http.StatusOK, "username or password wrong")
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	if err := u.setJWTToken(ctx, user.Id); err != nil {
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登陆成功！",
	})
	return
}

func (u *UserHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	// 用 JWT 设置登陆态, 生成一个 JWT token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误，生成token失败",
		})
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birth    string `json:"birth"`
		Bio      string `json:"bio"`
	}

	c, _ := ctx.Get("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "system error, get claims failed")
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
	c, _ := ctx.Get("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "system error, get claims failed")
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

	user, _ := u.svc.FindOrCreate(ctx, req.Phone)

	u.setJWTToken(ctx, user.Id)

	ctx.JSON(http.StatusOK, Result{
		Msg: "校验验证码通过",
	})
}
