package utils

import (
	"context"
	"github.com/gin-gonic/gin"
	"healthcheckProject/internal/api/middlewares"
	"healthcheckProject/internal/entity"
	"log"
)

func GetUser(ctx context.Context) *entity.UserShort {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		log.Fatal("not gin context")
	}

	value, exists := ginCtx.Get(middlewares.CurrentUserKey)
	if !exists {
		return nil
	}

	userShort, ok := value.(entity.UserShort)
	if !ok {
		return nil
	}

	return &userShort
}
