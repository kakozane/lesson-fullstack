package middleware

import (
	"log"
	"runtime/debug"

	"lesson06/backend/internal/dto/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if value := recover(); value != nil {
				log.Printf("request panic: %v\n%s", value, debug.Stack())
				if !c.Writer.Written() {
					response.Error(c, 500, "服务暂时不可用，请稍后重试")
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
