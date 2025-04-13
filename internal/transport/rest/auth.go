package rest

import (
	"custom-banking/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type Auth struct {
	userService UserService
}

func NewAuth(userService UserService) *Auth {
	return &Auth{userService}
}

func (a *Auth) InjectRouters(ginEngine *gin.Engine, middlewares ...gin.HandlerFunc) {
	auth := ginEngine.Group("/auth")
	{
		auth.POST("/register", a.singUp)
		auth.POST("/login", a.login)
		auth.GET("/refresh", a.refresh)
	}

	user := ginEngine.Group("/user").Use(middlewares...)
	{
		user.POST("/:id/block", a.blockUser)
		user.POST("/:id/unblock", a.unblockUser)
	}
}

func (a *Auth) singUp(ctx *gin.Context) {
	var inp models.SingUpInput

	if err := ctx.ShouldBindJSON(&inp); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation request body error": err.Error()})
		return
	}

	if inp.Name == "" || inp.Surname == "" || inp.Email == "" || inp.Password == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: name, surname, email, and password are required",
		})
		return
	}

	if inp.Username == "" {
		inp.Username = inp.Email
	}

	err := a.userService.SingUp(ctx, inp)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validation request body error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (a *Auth) login(ctx *gin.Context) {
	var inp models.SingInInput
	if err := ctx.ShouldBindJSON(&inp); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("validation request body error", err))
		return
	}

	if inp.Email == "" && inp.Username == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Either email or username must be provided"})
		return
	}

	accessToken, refreshToken, err := a.userService.SingIn(ctx, inp)
	if err != nil {
		if errors.Is(err, errors.New("user with such credentials not found")) { //TODO REPLACE ERROR.new TO CONST
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid credentials"})
		return
	}

	ctx.SetCookie("refresh-token", refreshToken, 3600, "/auth", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"token": accessToken,
	})
}

func (a *Auth) refresh(ctx *gin.Context) {
	cookie, err := ctx.Cookie("refresh-token")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, NewBadRequestError("get cookie from request error", err))
		return
	}

	accessToken, refreshToken, err := a.userService.RefreshTokens(ctx, cookie)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError("refresh token error", err))
		return
	}

	ctx.SetCookie("refresh-token", refreshToken, 3600, "/auth", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"token": accessToken,
	})
}

func (a *Auth) blockUser(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"user not found": err.Error()})
	}

	blockUserID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"block user not found": err.Error()})
		return
	}

	err = a.userService.BlockUser(ctx, blockUserID, userID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"block user not found": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (a *Auth) unblockUser(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"block user not found": err.Error()})
		return
	}

	err = a.userService.UnblockUser(ctx, userID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"unblocking user error": err.Error()})

	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}
