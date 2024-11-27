package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/prashant/golang-jwt-project/controllers"
	"github.com/prashant/golang-jwt-project/middleware"
)

func UserRoute(inccomingRoutes *gin.Engine){
	inccomingRoutes.Use(middleware.Authenticate())
	inccomingRoutes.GET("/users/user", controllers.GetUserById())
	inccomingRoutes.POST("users/revokeToken", controllers.RevokeToken()) //token revocation should be done by Admin only
}