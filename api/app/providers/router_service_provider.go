package providers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var routeRegistrations []func(*gin.Engine)

func RegisterRouteGroup(registration func(*gin.Engine)) {
	routeRegistrations = append(routeRegistrations, registration)
}

func RegisterRoutes(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://app.baselaragoproject.test"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	for _, registration := range routeRegistrations {
		registration(router)
	}
}
