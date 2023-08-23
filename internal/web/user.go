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

type UserHandler struct {
	svc              *service.UserService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	birthRegexExp    *regexp.Regexp
}

// UserClaims 用在JWT中
type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
		birthdayRegexPattern = `\d{4}-\d{2}-\d{2}`
	)

	return &UserHandler{
		svc:              svc,
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
		ctx.String(http.StatusOK, "system error")
		return
	}

	// 用 JWT 设置登陆态
	// 生成一个 JWT token

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:       user.Id,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "system error, generate token failed")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "Log in successful!")
	return
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		Birth    string `json:"birth"`
		Bio      string `json:"bio"`
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
		Email:    req.Email,
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
		ctx.String(http.StatusOK, "system error")
		return
	}

	ctx.String(http.StatusOK, "nickname: %s, birthday: %s, bio: %s", user.Nickname, user.Birth, user.Bio)
	return
}
