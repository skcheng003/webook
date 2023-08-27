package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/skcheng003/webook/config"
	"github.com/skcheng003/webook/internal/repository"
	"github.com/skcheng003/webook/internal/repository/dao"
	"github.com/skcheng003/webook/internal/service"
	"github.com/skcheng003/webook/internal/web"
	"github.com/skcheng003/webook/internal/web/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		// TODO: use logger replace panic
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(func(ctx *gin.Context) {
		println("This is a middleware")
	})

	/*
		redisClient := redis.NewClient(&redis.Options{
			Addr: config.Config.Redis.Addr,
		})
		// Using redis as a store for rate limiter
		server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
	*/
	server.Use(middleware.NewCorsMiddlewareBuilder().Build())

	// store, err := redis.NewStore(16, "tcp", config.Config.Redis.Addr, "",
	//	[]byte("W6RUUWNs6W3OYUpxJMG3E4Nj9PStZZUS"), []byte("dUfHJuOWQSSoJNuoPir4fWhwTggzyVDR"))

	store := memstore.NewStore([]byte("W6RUUWNs6W3OYUpxJMG3E4Nj9PStZZUS"), []byte("dUfHJuOWQSSoJNuoPir4fWhwTggzyVDR"))

	// myStore := sqlx_store.Store{}

	server.Use(sessions.Sessions("ssid", store))

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePath("/users/signup", "/users/login", "/hello").Build())

	return server
}

func main() {
	// initialize db

	db := initDB()
	// initialize user
	u := initUser(db)
	// initialize server
	server := initWebServer()
	// register routes
	u.RegisterRoutes(server)
	// start server

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello from the other side")
	})

	server.Run("localhost:8081")

	/*


		server2 := gin.Default()
		server2.GET("/hello", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, "hello k8s\n")
		})

		server2.Run("localhost:8080")
	*/
}
