# 开发环境要求

| 工具 | 版本要求 | 说明 |
| --- | --- | --- |
| Node.js | ^22.18.0 \|\| ^24.0.0 | 推荐使用 nvm；`.node-version` / `.nvmrc` 锁定为 **22.22.0** |
| pnpm | >= 11.0.0 | 唯一包管理器（`preinstall` 会拦截 npm/yarn）；`packageManager` 锁定为 **pnpm@11.4.0** |
| Git | 最新稳定版 | 提交前由 lefthook 自动执行 lint / 类型检查 |
| IDE | VS Code（推荐） | 已配置 `.vscode/`，打开项目时会提示安装推荐扩展 |

## 详细文档

- `docs/03-项目规范.md` — 规范总览
- `docs/02-快速开始.md` — 完整安装步骤、环境变量、常见问题
- `apps/web/src/views/guide/convention/` — 项目规范页面
- `apps/web/src/views/guide/quickstart/` — 快速开始页面
