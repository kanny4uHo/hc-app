package middlewares

import (
	"github.com/gin-gonic/gin"
	"healthcheckProject/internal/entity"
	"log"
	"net/http"
	"strconv"
)

const CurrentUserKey = "current_user_key"
const XLoginHeader = "X-Login"
const XUserIDHeader = "X-User-Id"

func RequireUser(ctx *gin.Context) {
	login := ctx.GetHeader(XLoginHeader)
	if login == "" {
		log.Println("empty login")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userIDString := ctx.GetHeader(XUserIDHeader)
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Printf("invalid user id: %s\n", userIDString)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.Set(CurrentUserKey, entity.UserShort{
		ID:    userID,
		Login: login,
	})

	ctx.Next()
}
