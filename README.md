# Lesson 06 · Vue3 + Go 商品管理与任务看板实战

基于 Vben Admin + Element Plus 的商品管理与任务看板课堂项目。后端使用 Gin + GORM，参考 aiagent_course 的业务模块分层。Go API 已提供，实战重点是前端业务实现。

现在有两个菜单：**商品管理** `/products`、**任务看板** `/tasks`。选择 [任务看板实战讲义](docs/01-任务看板实战.md)或 [商品管理实战讲义](docs/03-商品管理实战.md)，从初始版开始，按步骤完成代码、原理学习与结果验证。两条路线相互独立。

补充学习：[mise 全栈开发环境管理](docs/05-mise全栈开发环境管理.md)，介绍多语言版本固定、目录切换与项目任务，可独立阅读和练习。

## 学习资料怎么选

按下方启动说明准备环境，再进入 starter 后台。选择感兴趣的一条路线，无须先完成另一条。

| 学习路线 | 按顺序实战 | 随时查阅 |
|---|---|---|
| 任务看板 | [任务看板实战讲义](docs/01-任务看板实战.md) | [任务看板接口与 TS 类型](docs/02-任务看板接口与TypeScript类型.md) |
| 商品管理 | [商品管理实战讲义](docs/03-商品管理实战.md) | [商品管理接口与 TS 类型](docs/04-商品管理接口与TypeScript类型.md) |

05～07 是补充资料：mise 环境管理、线上部署、HTTPS 自动续期，可按需阅读。

## 初始版与完整版

前端提供独立初始应用 `frontend/apps/starter`，保留登录、布局和两个菜单，业务页只有标题；`frontend/apps/admin` 保留完整参考实现。

在 frontend 目录执行 `pnpm dev:starter`，访问 http://localhost:6112；完整版仍用 `pnpm dev:admin`，地址 http://localhost:6111。两者可同时运行，共用 Go 与数据库。具体实现步骤见对应的实战讲义。

## 启动

环境：Node.js 22.18+ 或 24、pnpm 11.4.0（以 frontend/package.json 为准）、Go 1.26、MySQL 8。

首次准备前端环境时，在已安装 Node.js 的终端安装 pnpm：

```sh
npm install -g pnpm@11.4.0
pnpm --version
```

后端终端：

```sh
# 在项目根目录打开终端
cd backend
# 首次拿到项目时：cp .env.example .env，并填写本地数据库密码。
# 已有 .env 时保留自己的配置，无须覆盖。
go run .
```

Go 会自动读取当前工作目录的 `.env`，因此请在 `backend` 目录执行 `go run .`，无需先 `source .env`。系统环境变量优先于文件配置；未设置的字段使用 `.env` 或代码默认值。没有 `.env` 时也可完全通过系统环境变量启动，原有密码和数据库名校验仍会执行；文件存在但无法读取或格式错误时，会明确报错并停止启动。引号、注释等语法由 [godotenv](https://github.com/joho/godotenv) 解析。

`run.sh` 保留为可选入口，仅切换目录并执行 `go run .`，配置读取统一交给 Go。

前端终端：

```sh
# 在项目根目录打开另一个终端
cd frontend
pnpm install --frozen-lockfile
pnpm dev
```

访问 http://localhost:6111 ，账号 `admin`，课堂默认密码 `123456`（由 backend/.env 的 ADMIN_PASSWORD 配置）。API：http://127.0.0.1:8080/api/health 。

请在 frontend 目录运行 pnpm，并使用项目约定的 pnpm 11.4.0。

启动后自动创建独立数据库 `lesson06_shop`、商品表、迁移记录和 36 条演示商品。只在首次初始化时插入演示数据，重启不会覆盖修改或补回删除的数据。不读取或修改其他业务库。新版增加独立任务表和看板版本表，迁移版本 2 首次插入 9 条任务，不重置现有商品。

## 前端代码检查与格式化

在 `frontend` 目录执行：

```sh
pnpm lint          # Oxlint 检查 JS、TS 和 Vue 脚本
pnpm lint:fix      # 自动修复可修复的问题
pnpm format        # Oxfmt 格式化
pnpm format:check  # 只检查格式，不修改文件
pnpm check:type    # Vue 模板与 TypeScript 类型检查
pnpm lint:style    # 独立的样式检查
```

两个应用和共享包共用 `oxlint.config.ts`、`oxfmt.config.ts`，具体规则位于 `internal/lint-configs`。已移除 ESLint 执行入口及配置包，Oxlint 使用原生规则；规则中的 `eslint/` 名称是 Oxlint 的兼容命名，并不调用 ESLint。

使用 VS Code 时，打开 `frontend` 目录，安装推荐的 Oxc 扩展即可在保存时格式化，工作区已禁用 ESLint。Oxlint 检查不能替代 `vue-tsc` 的模板与类型检查。Lefthook 也已改用 Oxfmt 和 Oxlint；它仅在 Git 仓库已安装对应钩子时运行。

## 功能

- 名称/编码搜索、分类与上下架筛选、重置、服务端分页。
- 新增/编辑共用弹窗、字段校验、重复编码提示、未保存修改提醒。
- 单条/批量上下架、删除确认；批量选择仅限当前页，刷新后清空。
- loading、提交防重复、请求失败重试、旧请求响应隔离。
- 金额使用整数分保存，页面显示元；删除末页数据后自动返回有效页。
- Vben 登录、菜单、标签页、主题和 Element Plus 组件风格。

任务看板功能：

- 待处理 / 进行中 / 已完成三列；跨列拖拽、列内排序持久化。
- 新建编辑抽屉、负责人、优先级、截止日期、逾期与进度统计。
- 关键词/负责人/编号与优先级筛选、复制摘要、删除确认。
- 复用 Element Plus、Vben useSortable、VueUse、日期工具；任务卡片、看板、抽屉按业务拆分。
- 乐观更新、失败恢复、版本冲突保护。单看板最多 200 条，筛选时禁止拖拽。

## 项目结构

```text
frontend/
  apps/admin/                完整参考应用
  apps/starter/              课堂初始应用，保留菜单与标题页
  packages/                  Vben 共享组件和能力包
  internal/                  构建、类型和样式配置
  scripts/                   框架工具脚本
backend/
  main.go                    加载配置、启动 Gin、优雅退出
  config/                    自动读取 .env、环境变量配置与启动校验
  bootstrap/                 数据库初始化、装配业务依赖
    app.go                   按顺序初始化商品和任务，再装配业务依赖
    database.go              MySQL/GORM 连接和连接池
    migrations.go            商品表、迁移记录与首次种子导入
    task_migration.go         任务表、看板版本与首次种子导入
    migrations_test.go       独立临时库的初始化回归测试
    README.md                新机器初始化、重复启动和分享说明
  migration/                 仅保存假数据 SQL
    001_products_seed.sql    36 条商品假数据，首次启动自动执行
    002_tasks_seed.sql       9 条任务假数据，首次启动自动执行
  router/admin/              后台路由注册
  middleware/                鉴权和统一异常恢复
  internal/product/          商品业务模块
    admin/dto/               查询、保存请求与响应结构
    admin/handler/           Gin 参数绑定与响应
    admin/service/           业务校验、实体转换和业务错误
    repository/              GORM 查询、更新和事务
    entity/                  商品表实体
  internal/task/             任务 dto、handler、service、repository、entity
  internal/auth/             登录 handler、service、dto
  internal/dto/response/     统一响应
  internal/pkg/              业务错误与 HTTP 参数工具
  .env.example               可公开的配置示例
  .env                       本地配置，禁止提交
  run.sh                     可选启动脚本，等同于在 backend 执行 go run .
docs/                        两条独立实战路线与配套参考资料

```

teacher、web、playground 和 backend-mock 已移除。保留 monorepo 是因为 admin 和 starter 共用 workspace 共享包，无需把 Vben 拆成另一套架构。原来的多端模板说明已替换为本项目说明。

## 检查

```sh
# backend 目录
 go test -race ./...
 go vet ./...
 go build ./...

# frontend 目录
 pnpm typecheck:starter
 pnpm build:starter
 pnpm -F @vben/admin run typecheck
 pnpm exec vitest run apps/admin/src/composables/use-product-list.test.ts apps/admin/src/composables/use-task-board.test.ts apps/admin/src/views/tasks/task-meta.test.ts
 pnpm build

```

完整版产物位于 `frontend/apps/admin/dist`，初始版产物位于 `frontend/apps/starter/dist`。开发时 Vite 将 `/api` 代理到 Go；部署静态产物时需要同源反向代理 `/api`，并把页面路由回退到 index.html。`vite preview` 仅预览静态资源，不能当作完整生产部署。

## 项目范围

这是学习与演示项目：固定管理员、单进程内存会话（8 小时过期，重启须重新登录）、不做注册和细粒度权限。Go 默认只监听 127.0.0.1。真实生产部署需要根据实际需求补充用户管理、登录限流、持久会话和并发编辑冲突检测；本项目的线上示例已配置 HTTPS 与独立数据库账号。商品保存策略是最后一次提交生效；任务看板已有全局版本冲突检测，但没有实时推送、多项目与成员权限。

数据库密码只保存在被忽略的 backend/.env 中；分享项目时保留 `.env.example`，排除真实 `.env`、node_modules 和 dist。手动压缩不自动遵循 `.gitignore`。新机器无需导入数据库备份，启动会执行 [数据库初始化与迁移](backend/bootstrap/README.md)。

## 线上演示与部署

完成版演示：[test.retniw.cc](https://test.retniw.cc)，账号 `admin` / `123456`。访问者共用演示数据，初始版仅供本地课堂实战。

发布流程：推送 GitHub `main` → Actions 检查并构建 → 推送阿里云 ACR → 服务器 Docker Compose 更新。配置与运维命令见 [线上部署与自动发布](docs/06-线上部署与自动发布.md)。

HTTPS 原理、首次配置和续期检查见 [HTTPS 证书申请与自动续期](docs/07-HTTPS证书申请与自动续期.md)。

模块接口详解：[商品管理与 TypeScript 类型](docs/04-商品管理接口与TypeScript类型.md)、[任务看板与 TypeScript 类型](docs/02-任务看板接口与TypeScript类型.md)。
