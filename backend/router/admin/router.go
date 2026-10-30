package admin

import (
	"lesson06/backend/bootstrap"
	"lesson06/backend/middleware"

	"github.com/gin-gonic/gin"
)

// Register 注册后台接口，外层路由组已统一添加 /api 前缀。
// 登录接口公开访问；会话信息、商品管理和任务看板均需要 Bearer token。
func Register(g *gin.RouterGroup, app *bootstrap.App) {
	g.POST("/auth/login", app.Auth.Login) // 管理员登录，返回访问令牌。

	// 同一鉴权分组下的接口都会先验证会话，未登录或会话过期时返回 401。
	protected := g.Group("", middleware.Auth(app.Sessions))
	protected.POST("/auth/logout", app.Auth.Logout)      // 退出登录，使当前令牌失效。
	protected.GET("/auth/codes", app.Auth.Codes)         // 获取前端使用的权限标识。
	protected.GET("/user/info", app.Auth.Info)           // 获取当前管理员信息。
	protected.GET("/categories", app.Product.Categories) // 获取商品分类选项。

	// 任务看板：写入必须携带看板 version，成功后返回最新完整看板。
	tasks := protected.Group("/tasks")
	tasks.GET("", app.Task.List)            // 查询所有任务及当前看板版本。
	tasks.POST("", app.Task.Create)         // 创建任务，加入指定状态列。
	tasks.PUT("/:id", app.Task.Update)      // 编辑任务内容，也可修改状态。
	tasks.DELETE("/:id", app.Task.Delete)   // 删除指定任务，请求体携带 version。
	tasks.PATCH("/:id/move", app.Task.Move) // 跨列移动或列内排序，beforeId 为 0 时追加到列末尾。

	// 商品管理：列表采用服务端筛选分页，商品金额统一使用整数分。
	products := protected.Group("/products")
	products.GET("", app.Product.List)                      // 按关键词、分类和上下架状态分页查询。
	products.POST("", app.Product.Create)                   // 新增商品，商品编码必须唯一。
	products.PUT("/:id", app.Product.Update)                // 编辑商品，保留库存 0、下架 0 和空描述等合法值。
	products.DELETE("/:id", app.Product.Delete)             // 删除指定商品。
	products.PATCH("/:id/status", app.Product.UpdateStatus) // 更新单个商品的上下架状态。
}
