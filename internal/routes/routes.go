package routes

import (
	"github.com/gin-gonic/gin"
	"goshort-api/internal/handlers"
)

func SetupRouter(
	urlHandler *handlers.URLHandler,
	authHandler *handlers.AuthHandler,
	jwtMiddleware gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	// Public routes
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// Protected routes
	authGroup := r.Group("/").Use(jwtMiddleware)
	{
		authGroup.POST("/shorten", urlHandler.ShortenURL)
		authGroup.GET("/list", urlHandler.ListURLs)
	}

	// Public redirect
	r.GET("/:shortKey", urlHandler.RedirectURL)

	return r
}