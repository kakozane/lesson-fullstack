package handler

import (
	"strings"

	"lesson06/backend/internal/auth/dto"
	"lesson06/backend/internal/auth/service"
	"lesson06/backend/internal/dto/response"
	"lesson06/backend/internal/pkg/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct{ auth *service.AuthService }

func New(auth *service.AuthService) *Handler { return &Handler{auth: auth} }
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !httpx.DecodeJSON(c, &req) {
		return
	}
	result, err := h.auth.Login(req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result, "")
}
func (h *Handler) Logout(c *gin.Context) {
	h.auth.Logout(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	response.Success(c, nil, "")
}
func (h *Handler) Info(c *gin.Context) {
	response.Success(c, gin.H{"id": 1, "userId": 1, "username": "admin", "realName": "课程管理员", "roles": []string{"admin"}, "homePath": "/products", "avatar": "/logo.svg"}, "")
}
func (h *Handler) Codes(c *gin.Context) { response.Success(c, []string{"products:manage"}, "") }
