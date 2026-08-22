package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hafidluqman50/maoi/src/http/handlers"
	"github.com/hafidluqman50/maoi/src/http/middleware"
	apiroutes "github.com/hafidluqman50/maoi/src/http/routes/api"
)

func NewRouter(corsOrigin string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Security())
	r.Use(middleware.CORS(corsOrigin))
	r.Use(gin.Recovery())

	r.GET("/health", handlers.HealthHandler)

	api := r.Group("/api/v1")
	apiroutes.RegisterRoutes(api)

	return r
}
