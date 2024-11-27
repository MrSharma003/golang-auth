package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/prashant/golang-jwt-project/controllers"
)

func AuhtRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("users/singup", controller.Signup())
	incomingRoutes.POST("users/signin", controller.Signin())
	incomingRoutes.POST("users/refreshToken", controller.RefreshToken())
}
