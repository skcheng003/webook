package web

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/skcheng003/webook/internal/domain"
	"github.com/skcheng003/webook/internal/service"
	svcmocks "github.com/skcheng003/webook/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncrypt(t *testing.T) {
	password := "hello#world123"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		reqBody    string
		expectCode int
		expectBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "senkie003@gmail.com",
					Password: "hello#world123",
				}).Return(nil)
				return userSvc, nil
			},
			reqBody: `
{
	"email": "senkie003@gmail.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			expectCode: http.StatusOK,
			expectBody: "注册成功",
		},
		{
			name: "参数不对，Bind失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil
			},
			reqBody: `
{
	"email": "senkie003@gmail.com",
	"passwd": "hello#world123"
}
`,
			expectCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil
			},
			reqBody: `
{
	"email": "senkie003#gmail.cc",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			expectCode: http.StatusOK,
			expectBody: "邮箱格式错误",
		},
		{
			name: "两次输入密码不一致",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil
			},
			reqBody: `
{
	"email": "senkie003@gmail.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world124"
}
`,
			expectCode: http.StatusOK,
			expectBody: "两次输入密码不一致",
		},
		{
			name: "密码必须包含数字、特殊字符，并且长度不能小于8位",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil
			},
			reqBody: `
{
	"email": "senkie003@gmail.com",
	"password": "hello",
	"confirmPassword": "hello"
}
`,
			expectCode: http.StatusOK,
			expectBody: "密码必须包含数字、特殊字符，并且长度不能小于8位",
		},
		{
			name: "邮箱地址冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "senkie003@gmail.com",
					Password: "hello#world123",
				}).Return(service.ErrUserDuplicateEmail)
				return userSvc, nil
			},
			reqBody: `
{
	"email": "senkie003@gmail.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			expectCode: http.StatusOK,
			expectBody: "邮箱地址冲突",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "senkie003@gmail.com",
					Password: "hello#world123",
				}).Return(errors.New("service调用错误"))
				return userSvc, nil
			},
			reqBody: `
{
	"email": "senkie003@gmail.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			expectCode: http.StatusOK,
			expectBody: "系统错误",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 注册路由
			userSvc, codeSvc := tc.mock(ctrl)
			h := NewUserHandler(userSvc, codeSvc, nil)
			h.RegisterRoutes(server)
			// 构造请求
			req, err := http.NewRequest(http.MethodPost, "/users/signup",
				bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			t.Log(resp)

			// HTTP 请求进入 Gin 框架的入口，响应写回到 resq 中
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.expectCode, resp.Code)
			assert.Equal(t, tc.expectBody, resp.Body.String())
		})
	}
}

func TestMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usersvc := svcmocks.NewMockUserService(ctrl)
	usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))

	err := usersvc.SignUp(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}
