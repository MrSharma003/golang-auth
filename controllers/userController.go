package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/prashant/golang-jwt-project/database"
	"github.com/prashant/golang-jwt-project/helpers"
	"github.com/prashant/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection("users")
var revokedTokenCollection *mongo.Collection = database.OpenCollection("revocations")
var validate = validator.New()

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(enteredPassword string, actualPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(actualPassword), []byte(enteredPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "Incorrect email or password!!!"
		check = false
	}

	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error})
			return
		}

		count_email, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Printf("Error occurred while checking the email: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while checking the email"})
			return
		}

		password, err := HashPassword(*user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		user.Password = &password

		count_phone, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Printf("Error occurred while checking the phone number: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while checking the phone"})
			return
		}

		if count_email > 0 || count_phone > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "this email or phone number already exist"})
			return
		}

		user.Created_at = time.Now().UTC()
		user.Updated_at = time.Now().UTC()
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refresh_token, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
		user.Token = &token
		user.Refresh_token = &refresh_token

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User item was not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Signin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		log.Println(user)

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		isPasswordValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if !isPasswordValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}

		token, refreshToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		// fetch updated user from db
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  token,
			"refresh_token": refreshToken,
		})
	}
}

func RefreshToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refreshToken := ctx.PostForm("refresh_token")

		claims, msg := helpers.ValidateToken(refreshToken)
		if msg != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		context, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()

		var foundUser models.User
		err := userCollection.FindOne(context, bson.M{"user_id": claims.Uid}).Decode(&foundUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		token, refreshToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		ctx.JSON(http.StatusOK, gin.H{
			"access_token":  token,
			"refresh_token": refreshToken,
		})

	}
}

func RevokeToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		context, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()

		userType := ctx.GetString("user_type")
		if userType != "ADMIN" {
			ctx.JSON(http.StatusForbidden, gin.H{
				"msg": "only Admin can revoke token",
			})
			return
		}

		var revokedToken models.Revocation

		if err := ctx.BindJSON(&revokedToken); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
			return
		}

		revokedToken.RevokedAt = time.Now().UTC()
		resultInsertionNumber, insertErr := revokedTokenCollection.InsertOne(context, revokedToken)

		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User item was not created"+insertErr.Error()})
			return
		}

		ctx.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func GetUserById() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ctx.JSON(http.StatusOK, gin.H{
			"msg": "you fetched user successfully",
		})
	}
}
