package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Revocation struct {
	ID        primitive.ObjectID `bson:"id"`
	Token     string             `json:"token" validate:"required"`
	RevokedAt time.Time          `json:"revoked_at"`
	Email    string             `json:"email" validate:"required"`
}
