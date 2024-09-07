//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/skcheng003/webook/internal/repository"
	"github.com/skcheng003/webook/internal/repository/cache"
	"github.com/skcheng003/webook/internal/repository/dao"
	"github.com/skcheng003/webook/internal/service"
	"github.com/skcheng003/webook/internal/web"
	jwt2 "github.com/skcheng003/webook/internal/web/jwt"
	"github.com/skcheng003/webook/ioc"
)

func initWebServer() *gin.Engine {
	wire.Build(
		// 第三方组件
		ioc.InitRedis, ioc.InitDB,

		dao.NewGORMUserDAO,

		cache.NewRedisUserCache,
		cache.NewRedisCodeCache,

		repository.NewUserRepository,
		repository.NewCachedCodeRepository,

		// 基于内存实现的短信服务
		ioc.InitSMSService,

		service.NewUserService,
		service.NewSMSCodeService,

		web.NewUserHandler,
		jwt2.NewRedisJWTHandler,

		ioc.InitMiddleWares,
		ioc.InitGinServer,
	)
	return new(gin.Engine)
}
