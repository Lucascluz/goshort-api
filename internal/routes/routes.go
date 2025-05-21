package routes
import (
	"goshort-api/internal/handlers"
	"github.com/gin-gonic/gin"

)

func SetupRouter(
	urlHandler *handlers.URLHandler,
	authHandler *handlers.AuthHandler,
	passwordResetHandler *handlers.PasswordResetHandler,
	jwtMiddleware gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	// Public routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)

		// Password reset
		auth.POST("/password-reset/request", passwordResetHandler.RequestReset)  	// sends otp
		auth.GET("/password-reset/confirm", passwordResetHandler.ConfirmReset)    	// verify otp 
		auth.POST("/password-reset/submit", passwordResetHandler.SubmitNewPassword) // sets new password
	}

	// Protected routes
	authGroup := r.Group("/").Use(jwtMiddleware)
	{
		authGroup.POST("/auth/logout", authHandler.Logout)

		// URL management
		authGroup.POST("/shorten", urlHandler.ShortenURL)
		authGroup.GET("/list", urlHandler.ListURLs)
	}

	// Public redirect route
	r.GET("/:shortKey", urlHandler.RedirectURL)

	return r
}
