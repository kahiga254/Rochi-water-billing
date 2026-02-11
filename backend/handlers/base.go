package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse returns a successful API response
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse returns a 201 Created response
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse returns an error API response
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// BadRequest returns a 400 Bad Request response
func BadRequest(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusBadRequest, message, err)
}

// NotFound returns a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message, nil)
}

// InternalServerError returns a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusInternalServerError, message, err)
}

// Unauthorized returns a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// Forbidden returns a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message, nil)
}

// ValidationError returns validation errors
func ValidationError(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: "Validation failed",
		Data:    errors,
		Error:   "validation_error",
	})
}
