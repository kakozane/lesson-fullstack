# 课堂初始版

本应用用于现场实战：登录、布局、商品管理与任务看板菜单已准备，商品管理提供静态页面、表单样式与预置校验规则，不包含业务状态、接口调用和事件处理；任务看板保留标题页。

在 frontend 根目录执行：

```sh
pnpm dev:starter
pnpm typecheck:starter
pnpm build:starter
```

开发地址 http://localhost:6112 ，完整参考应用在 apps/admin（http://localhost:6111）。两者共用 http://127.0.0.1:8080 的 Go API。

从 src/views/tasks/index.vue 或 src/views/products/index.vue 开始编写。接口函数和类型在 src/api/tasks.ts、src/api/products.ts，商品表单已有静态模板，业务逻辑与 composables 留给实战实现。

学习步骤见项目根 docs/03-商品管理实战.md 和 docs/01-任务看板实战.md。
