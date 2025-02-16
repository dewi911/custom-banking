package rest

import (
	"errors"
	"github.com/gin-gonic/gin"
)

func getUserIDFromContext(ctx *gin.Context) (int, error) {
	val, ok := ctx.Get("user-id")
	if !ok {
		return 0, errors.New("request context doesn't contains user id")
	}

	return val.(int), nil
}
