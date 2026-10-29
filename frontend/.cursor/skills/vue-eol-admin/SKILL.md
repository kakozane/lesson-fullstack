---
name: vue-eol-admin
description: >-
  Enforces vue-eol-admin project conventions for naming, imports, Vue components, business modules, styling, and git commits. Use when adding features, creating pages/routes/APIs, writing Vue components, or contributing to apps/web in this monorepo.
---

# vue-eol-admin 项目规范

本 skill 汇总 `docs/03-项目规范.md` 中的约定。按任务类型查阅对应 reference。

## 快速决策

| 任务                              | 参考文件                             |
| --------------------------------- | ------------------------------------ |
| 开发环境（Node / pnpm / 接入）    | [dev-env.md](dev-env.md)             |
| 命名（文件/路由/API/composable）  | [naming.md](naming.md)               |
| 新增业务模块（API → 页面 → 路由） | [add-module.md](add-module.md)       |
| 编写 Vue 组件                     | [vue-component.md](vue-component.md) |
| Import 路径别名                   | [imports.md](imports.md)             |
| 样式（Tailwind / scoped）         | [styling.md](styling.md)             |
| Git commit 消息                   | [git-commit.md](git-commit.md)       |
| Lint / 格式化                     | [lint-format.md](lint-format.md)     |

## 项目路径

- Web 应用根目录：`apps/web/src/`（别名 `#/`）
- API：`apps/web/src/api/<module>/`
- 页面：`apps/web/src/views/<module>/`
- 路由：`apps/web/src/router/routes/modules/<module>.ts`
- 路由模块文件保存后菜单自动注册，无需额外步骤

## 新增模块检查清单

```
- [ ] API：类型 + camelCase + Api 后缀函数，放 apps/web/src/api/<module>/
- [ ] 页面：apps/web/src/views/<module>/，遵循 vue-component 结构
- [ ] 路由：apps/web/src/router/routes/modules/<module>.ts
- [ ] 命名：文件夹 kebab-case，路由 name PascalCase，path kebab-case
- [ ] Import：使用 #/ 别名，禁止跨层级相对路径
- [ ] 样式：优先 Tailwind，复杂样式用 scoped + :deep()
- [ ] 提交：Conventional Commits 格式
```

## 常用命令

```bash
pnpm lint      # lint 检查
pnpm format    # 代码格式化
```
