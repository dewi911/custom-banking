package rest

import (
	"custom-banking/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

type Auth struct {
	userService UserService
}

func NewAuth(userService UserService) *Auth {
	return &Auth{userService}
}

func (a *Auth) InjectRouters(ginEngine *gin.Engine) {
	auth := ginEngine.Group("/auth")
	{
		auth.POST("/sing-up", a.singUp)
		auth.POST("/sing-in", a.singIn)
	}
}

func (a *Auth) singUp(ctx *gin.Context) {
	var inp models.SingUpInput

	if err := ctx.ShouldBindJSON(&inp); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation request body error": err.Error()})
		return
	}

	err := a.userService.SingUp(ctx, inp)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation request body error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (a *Auth) singIn(ctx *gin.Context) {
	var inp models.SingInInput
	if err := ctx.ShouldBindJSON(&inp); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation request body error": err.Error()})
		return
	}

	//todo token acces

	accessToken, refreshToken, err := a.userService.SingIn(ctx, inp)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"user not found": err.Error()})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"sing in error": err.Error()})
		return
	}

	//todo set cockie

	ctx.JSON(http.StatusOK, gin.H{"accessToken": accessToken, "refreshToken": refreshToken})
}

func (a *Auth) refresh(ctx *gin.Context) {
	cookie, err := ctx.Cookie("refresh-token")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"refresh token error": err.Error()})
		return
	}

	accessToken, refreshToken, err := a.userService.RefreshTokens(ctx, cookie)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"refresh token error": err.Error()})
		return
	}

	ctx.SetCookie("refresh-token", accessToken, 60*60*24, "/auth", "localhost", false, true)
	ctx.JSON(http.StatusOK, gin.H{"accessToken": accessToken, "refreshToken": refreshToken})
}
