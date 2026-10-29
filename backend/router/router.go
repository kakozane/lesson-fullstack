package router

import (
	"lesson06/backend/bootstrap"
	"lesson06/backend/internal/dto/response"
	"lesson06/backend/middleware"
	"lesson06/backend/router/admin"

	"github.com/gin-gonic/gin"
)

func SetupRouter(app *bootstrap.App) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.HandleMethodNotAllowed = true
	r.Use(middleware.Recovery())
	r.NoRoute(func(c *gin.Context) { response.Error(c, 404, "接口不存在") })
	r.NoMethod(func(c *gin.Context) { response.Error(c, 405, "请求方法不支持") })
	r.GET("/api/health", func(c *gin.Context) {
		if err := app.SQLDB.PingContext(c.Request.Context()); err != nil {
			response.Error(c, 503, "数据库不可用")
			return
		}
		response.Success(c, gin.H{"status": "ok"}, "")
	})
	admin.Register(r.Group("/api"), app)
	return r
}
