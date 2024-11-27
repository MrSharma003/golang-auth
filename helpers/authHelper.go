package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// user role validation function
func CheckUserType(ctx *gin.Context, role string) (err error) {
	userType := ctx.GetString("user_type")
	err = nil

	if userType != role {
		err = errors.New("unauthorized role access to the resource")
		return err
	}

	return err
}

func MatchUserTypeToUid(ctx *gin.Context, userId string) (err error) {
	userType := ctx.GetString("user_type")
	uid := ctx.GetString("uid")

	if userType == "USER" && uid != userId {
		err = errors.New("unauthorized access to the resource")
		return err
	}

	if err := CheckUserType(ctx, userType); err != nil {
		return err
	}

	return err
}
