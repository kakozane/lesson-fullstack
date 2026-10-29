# 课堂初始版

本应用用于现场实战：登录、布局、商品管理与任务看板菜单已准备，两个业务页仅保留标题。

在 frontend 根目录执行：

```sh
corepack pnpm dev:starter
corepack pnpm typecheck:starter
corepack pnpm build:starter
```

开发地址 http://localhost:5667 ，完整参考应用在 apps/admin（http://localhost:5666）。两者共用 http://127.0.0.1:8080 的 Go API。

从 src/views/tasks/index.vue 或 src/views/products/index.vue 开始编写。接口函数和类型在 src/api/tasks.ts、src/api/products.ts，业务组件与 composables 留给课堂实现。

详细操作、课程安排和共享数据库边界见项目根 docs/08-课堂初始版使用说明.md。
