package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response envelope
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    any `json:"data,omitempty"`
	Meta    any `json:"meta,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// ErrorBody follows RFC 7807 Problem Details
type ErrorBody struct {
	Type     string         `json:"type,omitempty"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Errors   []FieldFailure `json:"errors,omitempty"`
}

// FieldFailure represents a specific field validation error
type FieldFailure struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Success sends a standard success response
func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a standard success response with metadata
func SuccessWithMeta(c *gin.Context, status int, message string, data any, meta any) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Error sends a standard error response following RFC 7807
func Error(c *gin.Context, status int, message string, detail string, errs []FieldFailure) {
	c.JSON(status, Response{
		Success: false,
		Message: message,
		Error: &ErrorBody{
			Status:   status,
			Title:    http.StatusText(status),
			Detail:   detail,
			Instance: c.Request.URL.Path,
			Errors:   errs,
		},
	})
}

// InternalServerError sends a 500 status code with a standard error response
func InternalServerError(c *gin.Context, message string, detail string) {
	Error(c, http.StatusInternalServerError, message, detail, nil)
}

// BadRequest sends a 400 status code with a standard error response
func BadRequest(c *gin.Context, message string, detail string) {
	Error(c, http.StatusBadRequest, message, detail, nil)
}

// NotFound sends a 404 status code with a standard error response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, "The requested resource was not found", nil)
}
