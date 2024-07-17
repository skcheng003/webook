package ioc

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/skcheng003/webook/internal/web"
	"github.com/skcheng003/webook/internal/web/middleware"
	ratelimit "github.com/skcheng003/webook/pkg/ginx/middlewares/ratelimits"
	"time"
)

func InitGinServer(middlewares []gin.HandlerFunc, handler *web.UserHandler) *gin.Engine {
	server := gin.Default()
	handler.RegisterRoutes(server)
	server.Use(middlewares...)
	return server
}

func InitMiddleWares(redisClient redis.Cmdable) []gin.HandlerFunc {
	store := memstore.NewStore([]byte("W6RUUWNs6W3OYUpxJMG3E4Nj9PStZZUS"), []byte("dUfHJuOWQSSoJNuoPir4fWhwTggzyVDR"))

	return []gin.HandlerFunc{
		middleware.NewCorsMiddlewareBuilder().Build(),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePath("/users/login_sms/code/send").
			IgnorePath("/users/login_sms").
			IgnorePath("/users/signup", "/users/login").Build(),
		sessions.Sessions("ssid", store),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}
