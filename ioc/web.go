package ioc

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/skcheng003/webook/internal/web"
	jwt2 "github.com/skcheng003/webook/internal/web/jwt"
	"github.com/skcheng003/webook/internal/web/middleware"
)

func InitGinServer(middlewares []gin.HandlerFunc, handler *web.UserHandler) *gin.Engine {
	server := gin.Default()
	handler.RegisterRoutes(server)
	server.Use(middlewares...)
	return server
}

func InitMiddleWares(jwtHdl jwt2.Handler) []gin.HandlerFunc {
	store := memstore.NewStore([]byte("W6RUUWNs6W3OYUpxJMG3E4Nj9PStZZUS"), []byte("dUfHJuOWQSSoJNuoPir4fWhwTggzyVDR"))

	return []gin.HandlerFunc{
		middleware.NewCorsMiddlewareBuilder().Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePath("/users/login_sms/code/send").
			IgnorePath("/users/login_sms").
			IgnorePath("/users/signup", "/users/login").
			IgnorePath("/users/refresh_token").Build(),
		sessions.Sessions("ssid", store),
		// ratelimit.NewBuilder().Build(),
	}
}
