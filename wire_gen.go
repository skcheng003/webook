// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/skcheng003/webook/internal/repository"
	"github.com/skcheng003/webook/internal/repository/cache"
	"github.com/skcheng003/webook/internal/repository/dao"
	"github.com/skcheng003/webook/internal/service"
	"github.com/skcheng003/webook/internal/web"
	"github.com/skcheng003/webook/ioc"
)

// Injectors from wire.go:

func initWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	v := ioc.InitMiddleWares(cmdable)
	db := ioc.InitDB()
	userDao := dao.NewGORMUserDAO(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository)
	smsService := ioc.InitSMSService()
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	codeService := service.NewSMSCodeService(smsService, codeRepository)
	userHandler := web.NewUserHandler(userService, codeService)
	engine := ioc.InitGinServer(v, userHandler)
	return engine
}