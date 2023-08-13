package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/skcheng003/webook/internal/repository"
	"github.com/skcheng003/webook/internal/repository/dao"
	"github.com/skcheng003/webook/internal/service"
	"github.com/skcheng003/webook/internal/web"
	"github.com/skcheng003/webook/internal/web/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
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

	server.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"https://localhost:3000"},
		// AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "authorization"},
		// ExposeHeaders:    []string{"Content-Type", "authorization"},
		// 是否允许cookie
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "skcheng.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	store := cookie.NewStore([]byte("secret"))
	server.Use(sessions.Sessions("ssid", store))

	server.Use(middleware.NewLoginMiddlewareBuilder().
		IgnorePath("/users/signup", "/users/login").Build())

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
	server.Run("localhost:8080")
}
