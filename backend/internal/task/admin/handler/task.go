package handler

import (
	"github.com/gin-gonic/gin"
	"lesson06/backend/internal/dto/response"
	"lesson06/backend/internal/pkg/httpx"
	"lesson06/backend/internal/task/admin/dto"
	"lesson06/backend/internal/task/admin/service"
)

type Handler struct{ tasks *service.TaskService }

func New(tasks *service.TaskService) *Handler { return &Handler{tasks: tasks} }
func send(c *gin.Context, b *dto.Board, err error) {
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, b, "")
}
func (h *Handler) List(c *gin.Context)   { b, e := h.tasks.List(c.Request.Context()); send(c, b, e) }
func (h *Handler) Create(c *gin.Context) { h.save(c, 0) }
func (h *Handler) Update(c *gin.Context) {
	id, ok := httpx.ID(c)
	if ok {
		h.save(c, id)
	}
}
func (h *Handler) save(c *gin.Context, id int64) {
	var p dto.SaveRequest
	if !httpx.DecodeJSON(c, &p) {
		return
	}
	b, e := h.tasks.Save(c.Request.Context(), id, p)
	send(c, b, e)
}
func (h *Handler) Delete(c *gin.Context) {
	id, ok := httpx.ID(c)
	if !ok {
		return
	}
	var p dto.DeleteRequest
	if !httpx.DecodeJSON(c, &p) {
		return
	}
	b, e := h.tasks.Delete(c.Request.Context(), id, p.Version)
	send(c, b, e)
}
func (h *Handler) Move(c *gin.Context) {
	id, ok := httpx.ID(c)
	if !ok {
		return
	}
	var p dto.MoveRequest
	if !httpx.DecodeJSON(c, &p) {
		return
	}
	b, e := h.tasks.Move(c.Request.Context(), id, p)
	send(c, b, e)
}
