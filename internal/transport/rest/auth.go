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

func (a *Auth) singUp(c *gin.Context) {
	var inp models.SingUpInput

	if err := c.ShouldBindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation request body error": err.Error()})
		return
	}

	err := a.userService.SingUp(c, inp)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation request body error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (a *Auth) singIn(c *gin.Context) {
	var inp models.SingInInput
	if err := c.ShouldBindJSON(&inp); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation request body error": err.Error()})
		return
	}

	//todo token acces

	accessToken, refreshToken, err := a.userService.SingIn(c, inp)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"user not found": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"sing in error": err.Error()})
		return
	}

	//todo set cockie

	c.JSON(http.StatusOK, gin.H{"accessToken": accessToken, "refreshToken": refreshToken})
}
