package helpers

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prashant/golang-jwt-project/database"
	"github.com/prashant/golang-jwt-project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.RegisteredClaims
}

var userCollection *mongo.Collection = database.OpenCollection("user")
var revokedTokenCollection *mongo.Collection = database.OpenCollection("revocations")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, userType string, userId string) (signedToken string, sigendRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		User_type:  userType,
		Uid:        userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	refreshClaims := &SignedDetails{
		Uid:       userId,
		User_type: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}

	token, err1 := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err2 := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err1 != nil || err2 != nil {
		log.Printf("Failed to generate refresh token: %v", err)
		return "", "", err1
	}

	return token, refreshToken, nil
}

func UpdateAllTokens(token string, refreshToken string, userId string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: token})
	updateObj = append(updateObj, bson.E{Key: "refreshToken", Value: refreshToken})

	Updated_at := time.Now().UTC()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		&opt,
	)

	if err != nil {
		log.Fatalf("failed to update tokens %v", err)
		return
	}
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "token is invalid"
		return
	}

	context, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	var revokedToken models.Revocation
	err = revokedTokenCollection.FindOne(context, bson.M{"token": signedToken}).Decode(&revokedToken)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			msg = "failed to check token in database: " + err.Error()
			return
		}
	}

	if revokedToken != (models.Revocation{}) {
		msg = "token is revoked"
		return
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now().UTC()) {
		msg = "token is expired"
		return
	}

	return claims, msg
}
