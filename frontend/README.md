# Lesson 06 前端

Vue3 + TypeScript + Vben + Element Plus 多包项目，包含两个独立应用，共用 packages 和 internal。

| 应用 | 用途 | 启动 | 地址 |
|---|---|---|---|
| apps/admin | 完整参考版本 | pnpm dev:admin | http://localhost:6111 |
| apps/starter | 课堂初始版本，商品管理含静态样式 | pnpm dev:starter | http://localhost:6112 |

```sh
# 在本目录安装依赖
pnpm install --frozen-lockfile

# 初始版检查与构建
pnpm typecheck:starter
pnpm build:starter

# 完整版检查与构建
pnpm -F @vben/admin run typecheck
pnpm build
```

启动开发服务后会自动打开浏览器。admin 固定使用 6111 端口，starter 固定使用 6112 端口；端口被占用时会报错，不自动切换。

`pnpm dev` 默认启动完整版。Go API 需要同时运行，两个应用都将 `/api` 代理到 http://127.0.0.1:8080，未启用 Mock。两应用共用同一数据库，分别使用独立的应用缓存命名空间。

课堂修改 apps/starter/src/views/products/index.vue 或 apps/starter/src/views/tasks/index.vue；API 文件已经给出。完整实现留在 apps/admin/src 的对应目录。

完整启动说明见 [项目 README](../README.md)，初始版实现步骤见 [任务看板讲义](../docs/01-任务看板实战.md)或 [商品管理讲义](../docs/03-商品管理实战.md)。
