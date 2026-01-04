package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response format
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Wrap wraps a handler function to return standard response format
func Wrap(h func(*gin.Context) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := h(c)
		if err != nil {
			c.JSON(http.StatusOK, Response{
				Code:    -1,
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, Response{
			Code:    0,
			Message: "success",
			Data:    data,
		})
	}
}

// Success returns a success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error returns an error response
func Error(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    -1,
		Message: message,
	})
}
