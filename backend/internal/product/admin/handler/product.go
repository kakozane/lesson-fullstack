package handler

import (
	"lesson06/backend/internal/dto/response"
	"lesson06/backend/internal/pkg/httpx"
	"lesson06/backend/internal/product/admin/dto"
	"lesson06/backend/internal/product/admin/service"

	"github.com/gin-gonic/gin"
)

type Handler struct{ products *service.ProductService }

func New(products *service.ProductService) *Handler { return &Handler{products: products} }
func (h *Handler) Categories(c *gin.Context)        { response.Success(c, service.Categories(), "") }
func (h *Handler) List(c *gin.Context) {
	q := dto.ListQuery{Page: 1, PageSize: 10}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, 400, "查询参数格式错误")
		return
	}
	result, err := h.products.List(c.Request.Context(), q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result, "")
}
func (h *Handler) Create(c *gin.Context) {
	var req dto.SaveRequest
	if !httpx.DecodeJSON(c, &req) {
		return
	}
	if err := h.products.Create(c.Request.Context(), req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil, "保存成功")
}
func (h *Handler) Update(c *gin.Context) {
	id, ok := httpx.ID(c)
	if !ok {
		return
	}
	var req dto.SaveRequest
	if !httpx.DecodeJSON(c, &req) {
		return
	}
	if err := h.products.Update(c.Request.Context(), id, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil, "保存成功")
}
func (h *Handler) Delete(c *gin.Context) {
	id, ok := httpx.ID(c)
	if !ok {
		return
	}
	if err := h.products.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}
func (h *Handler) BatchStatus(c *gin.Context) {
	var req dto.BatchStatusRequest
	if !httpx.DecodeJSON(c, &req) {
		return
	}
	if err := h.products.BatchStatus(c.Request.Context(), req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil, "批量更新成功")
}
