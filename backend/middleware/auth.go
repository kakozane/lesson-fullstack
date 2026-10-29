package middleware

import (
	"strings"

	"lesson06/backend/internal/auth/service"
	"lesson06/backend/internal/dto/response"

	"github.com/gin-gonic/gin"
)

func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") || !auth.Valid(strings.TrimPrefix(header, "Bearer ")) {
			response.Error(c, 401, "登录已过期，请重新登录")
			return
		}
		c.Next()
	}
}
