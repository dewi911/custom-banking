package rest

import (
	"custom-banking/internal/models"
	"github.com/pkg/errors"

	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

func getUserIDFromContext(ctx *gin.Context) (int, error) {
	val, ok := ctx.Get("user-id")
	if !ok {
		return 0, errors.New("request context doesn't contains user id")
	}

	return val.(int), nil
}

func buildOrderingMessage(input string, supportedFields []string) (models.Orderings, error) {
	if input == "" {
		return nil, nil
	}

	orderings := make(models.Orderings)

	fieldsCondition := strings.Split(input, "|")
	for _, fieldCondition := range fieldsCondition {
		parts := strings.Split(fieldCondition, ":")
		if len(parts) != 2 {
			return nil, errors.New("wrong ordering filter format")
		}

		var isSupportedField bool
		for _, field := range supportedFields {
			if parts[0] == field {
				isSupportedField = true
			}
		}

		if !isSupportedField {
			return nil, fmt.Errorf("using unsupported ordering filter param [%s]", parts[0])
		}

		var direction string
		switch parts[1] {
		case "asc":
			direction = "asc"
		case "desc":
			direction = "desc"
		default:
			return nil, errors.New("unsupported filtering direction")
		}

		orderings[parts[0]] = direction
	}

	return orderings, nil
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

func GetUsernameFromContext(c *gin.Context) (string, error) {
	username, exists := c.Get(ctxUsernameKey)
	if !exists {
		return "", errors.New("username not found in context")
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", errors.New("username is of invalid type")
	}

	return usernameStr, nil
}
