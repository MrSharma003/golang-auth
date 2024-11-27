package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prashant/golang-jwt-project/database"
	"github.com/prashant/golang-jwt-project/routes"
	"go.mongodb.org/mongo-driver/mongo"
)

var dbClient *mongo.Client

func main(){
	port:= os.Getenv("PORT")

	if port == ""{
		port = "8000"
	}

	dbClient = database.DBinstance()

	router:= gin.New()
	router.Use(gin.Logger())
	router.SetTrustedProxies(nil)

	router.GET("/api1", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Message coming from api1"})
	})

	routes.AuhtRoutes(router)
	routes.UserRoute(router)

	router.Run(":"+port)
}