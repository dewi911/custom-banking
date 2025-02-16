package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

const (
	ctxUserIDKey     = "user-id"
	ctxUserRoleIDKey = "user-role-id"
)

const AuthorizationHeaderName = "Authorization"

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		fields := logrus.Fields{
			"method":          c.Request.Method,
			"uri":             c.Request.RequestURI,
			"request-in-time": t.Format(time.RFC3339),
		}

		c.Next()

		duration := time.Since(t)
		fields["request-handling-duration"] = duration.Microseconds()

		logrus.WithFields(fields).Info()
	}
}

func (a *Auth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getTokenFromRequest(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"token is missing or invalid": err.Error()})
			return
		}

		userID, roleID, err := a.userService.ParseToken(c, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"wrong authorization token": err.Error()})
			return
		}

		checkBlockUser, err := a.userService.CheckBlockUser(c, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"user block check error user not found": err.Error()})
			return
		}

		if checkBlockUser {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"user is blocked": err.Error()})
			return
		}

		c.Set(ctxUserIDKey, userID)
		c.Set(ctxUserRoleIDKey, roleID)

		c.Next()
	}
}

func getTokenFromRequest(c *gin.Context) (string, error) {
	header := c.GetHeader(AuthorizationHeaderName)
	if header == "" {
		return "", errors.Wrap(nil, "empty authorization header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", errors.Wrap(nil, "invalid authorization header")
	}

	if len(headerParts[1]) == 0 {
		return "", errors.Wrap(nil, "token is empty")
	}

	return headerParts[1], nil
}
