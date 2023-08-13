package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/service"
	"net/http"
)

const (
	emailRegexPattern   = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPatter = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

var ErrUserDuplicateEmail = service.ErrUserDuplicateEmail
var ErrUserNoFound = service.ErrUserNoFound
var ErrInvalidUserOrPassword = service.ErrInvalidUserOrPassword

type UserHandler struct {
	svc              *service.UserService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern   = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPatter = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
	)

	return &UserHandler{
		svc:              svc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPatter, regexp.None),
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
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
	sess.Save()
	ctx.String(http.StatusOK, "Log in successful!")
	return
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "hello profile")
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

	err := u.svc.Edit(ctx, domain.User{
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
