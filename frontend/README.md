# Lesson 06 前端

Vue3 + TypeScript + Vben + Element Plus 多包项目，包含两个独立应用，共用 packages 和 internal。

| 应用 | 用途 | 启动 | 地址 |
|---|---|---|---|
| apps/admin | 完整参考版本 | corepack pnpm dev:admin | http://localhost:5666 |
| apps/starter | 课堂初始版本，业务页仅标题 | corepack pnpm dev:starter | http://localhost:5667 |

```sh
# 在本目录安装依赖
corepack pnpm install --frozen-lockfile

# 初始版检查与构建
corepack pnpm typecheck:starter
corepack pnpm build:starter

# 完整版检查与构建
corepack pnpm -F @vben/admin run typecheck
corepack pnpm build
```

`corepack pnpm dev` 默认启动完整版。Go API 需要同时运行，两个应用都将 `/api` 代理到 http://127.0.0.1:8080，未启用 Mock。两应用共用同一数据库，分别使用独立的应用缓存命名空间。

课堂修改 apps/starter/src/views/products/index.vue 或 apps/starter/src/views/tasks/index.vue；API 文件已经给出。完整实现留在 apps/admin/src 的对应目录。

完整启动说明见 ../README.md，初始版操作见 ../docs/01-课堂初始版使用说明.md。
