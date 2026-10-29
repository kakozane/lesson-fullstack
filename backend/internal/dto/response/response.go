package response

import (
	"errors"
	"log"

	"lesson06/backend/internal/pkg/apperr"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

func Success(c *gin.Context, data any, message string) {
	c.Header("Cache-Control", "no-store")
	c.JSON(200, Response{Data: data, Message: message})
}
func Error(c *gin.Context, code int, message string) {
	c.Header("Cache-Control", "no-store")
	c.AbortWithStatusJSON(code, Response{Code: code, Message: message})
}
func Fail(c *gin.Context, err error) {
	var business *apperr.Error
	if errors.As(err, &business) {
		Error(c, business.Code, business.Message)
		return
	}
	log.Printf("request failed: %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	Error(c, 500, "服务暂时不可用，请稍后重试")
}
