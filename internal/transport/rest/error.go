package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type errorResponse struct {
	Message string `json:"message"`
}

// newErrorResponse creates a standard error response and aborts the request
func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, errorResponse{
		Message: message,
	})
}

type statusRequest struct {
	Status string `json:"status" binding:"required"`
}

type CommonError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type BadRequestError struct {
	CommonError
	Fields map[string]string `json:"fields,omitempty"`
}

func NewBadRequestError(message string, err error) any {
	e := BadRequestError{
		CommonError: CommonError{
			Code:    http.StatusBadRequest,
			Message: message,
		},
	}

	fe, ok := err.(validator.ValidationErrors)
	if ok {
		errs := make(map[string]string)
		for _, fieldError := range fe {
			errs[fieldError.Field()] = fieldError.Error()
		}

		e.Fields = errs
	} else {
		e.Error = stringifyError(err)
	}

	return e
}

type InternalServerError struct {
	CommonError
}

func NewInternalServerError(message string, err error) InternalServerError {
	return InternalServerError{CommonError: CommonError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Error:   stringifyError(err),
	}}
}

type UnauthorizedError struct {
	CommonError
}

func NewUnauthorizedError(message string, err error) UnauthorizedError {
	return UnauthorizedError{CommonError: CommonError{
		Code:    http.StatusUnauthorized,
		Message: message,
		Error:   stringifyError(err),
	}}
}

type NotFoundError struct {
	CommonError
}

func NewNotFoundError(message string, err error) NotFoundError {
	return NotFoundError{CommonError: CommonError{
		Code:    http.StatusNotFound,
		Message: message,
		Error:   stringifyError(err),
	}}
}

type ForbiddenError struct {
	CommonError
}

func NewForbiddenError(message string, err error) ForbiddenError {
	return ForbiddenError{CommonError: CommonError{
		Code:    http.StatusForbidden,
		Message: message,
		Error:   stringifyError(err),
	}}
}

func stringifyError(err error) string {
	if err != nil {
		return err.Error()
	}

	return ""
}
