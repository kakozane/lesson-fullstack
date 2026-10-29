package httpx

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"lesson06/backend/internal/dto/response"

	"github.com/gin-gonic/gin"
)

// DecodeJSON 保留大小限制和严格 JSON 校验，拒绝拼接对象与未知字段。
func DecodeJSON(c *gin.Context, value any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		response.Error(c, 400, "请求参数格式错误")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		response.Error(c, 400, "请求只能包含一个 JSON 对象")
		return false
	}
	return true
}
func ID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		response.Error(c, 400, "商品 ID 不合法")
		return 0, false
	}
	return id, true
}
