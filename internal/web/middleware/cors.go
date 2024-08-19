package middleware

import (
	gincors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type CorsMiddlewareBuilder struct {
}

func NewCorsMiddlewareBuilder() *CorsMiddlewareBuilder {
	return &CorsMiddlewareBuilder{}
}

func (c *CorsMiddlewareBuilder) Build() gin.HandlerFunc {
	return gincors.New(gincors.Config{
		// AllowOrigins: []string{"https://localhost:3000"},
		// AllowMethods: []string{"POST", "GET"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"X-Access-Token", "X-Refresh-Token"},
		// 是否允许cookie
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "skcheng.com")
		},
		MaxAge: 12 * time.Hour,
	})
}
