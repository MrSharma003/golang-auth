package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prashant/golang-jwt-project/helpers"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientToken := ctx.Request.Header.Get("token")
		if clientToken == "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized access"})
			ctx.Abort() //Prevent further processing of the request
			return
		}

		claims, msg := helpers.ValidateToken(clientToken)
		if msg != "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			ctx.Abort()
			return
		}
		ctx.Set("email", claims.Email)
		ctx.Set("first_name", claims.First_name)
		ctx.Set("last_name", claims.Last_name)
		ctx.Set("uid", claims.Uid)
		ctx.Set("user_type", claims.User_type)
		ctx.Next()
	}
}
