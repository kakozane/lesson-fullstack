# Lesson 06 Go API

Gin 1.12 + GORM 1.31 + MySQL，参考 aiagent_course 的模块分层和手工依赖装配。启动、数据库与接口说明见项目根 README.md 和 [商品接口文档](../docs/05-商品管理接口与TypeScript类型.md)。

```sh
# 首次在新机器使用时
cp .env.example .env
# 填写数据库密码后
go run .

# 检查
go test -race ./...
go vet ./...
```

Go 会自动读取当前工作目录的 `.env`，请在 `backend` 目录执行 `go run .`，无需先 `source .env`。系统环境变量优先；没有 `.env` 时可使用系统环境变量，文件无法读取或格式错误时会停止启动。配置解析使用 [godotenv](https://github.com/joho/godotenv)，原有密码与数据库名校验保持不变。`run.sh` 仅作为可选的启动入口。

启动只初始化配置指定的课程库。已有 .env 时保留本地配置；新机器从 .env.example 创建自己的配置。分享项目时排除真实 .env。

## 调用链

`router/admin → internal/product/admin/handler → service → repository → MySQL`

- main 只负责配置、应用初始化、HTTP 服务生命周期。
- bootstrap 管理连接、建库、建表和首次种子初始化，migration 只保存假数据 SQL，说明见 [数据库初始化与迁移](bootstrap/README.md)。
- bootstrap 完成数据库初始化，再创建 repository / service / handler。
- handler 处理 Gin 参数和响应；service 校验业务、转换 DTO；repository 通过 GORM 访问数据。
- entity 表示表结构，DTO 表示接口契约；两者不互相替代。
- middleware 负责会话校验和 panic 恢复；response 保留 `{code,data,message}`。

接口统一使用 `/api` 前缀，配置由 `.env` 和环境变量提供。商品表由显式迁移管理，不使用 AutoMigrate 改动现有表。初始化迁移保留 database/sql，运行时商品仓储使用 GORM。

课堂侧重点是前端，后端分层可按上述调用链对照源码阅读。

## 任务看板

任务模块位于 internal/task，沿用相同分层。bootstrap/task_migration.go 在商品迁移之后初始化任务表与看板版本，种子版本为 2。所有任务写入校验 version 并在事务中加锁，单看板最多 200 条。接口说明见 [任务看板接口文档](../docs/03-任务看板接口与TypeScript类型.md)。

```sh
# 项目根目录，确保服务已启动
python3 backend/tests/task_smoke.py
API_URL=http://localhost:5666/api python3 backend/tests/task_smoke.py
```

## 分享与首次启动

数据库初始化 Go 文件位于 `bootstrap/`，假数据 SQL 位于 `migration/`。新机器准备 MySQL、配置 `.env` 后运行 `go run .`，自动创建表和演示数据，无需导入数据库备份。重复启动不会覆盖已修改数据或补回已删除记录。运行时业务实体和仓储仍位于 `internal/` 各业务模块。

假数据 SQL 为 `migration/001_products_seed.sql`（36 条商品）和 `migration/002_tasks_seed.sql`（9 条任务）。Go 将文件嵌入程序，首次启动自动执行；无需再次手动导入。新机器也会自动创建配置指定的数据库，前提是 MySQL 已运行且账号有建库建表权限。

`router/admin/router.go` 按登录、鉴权、任务和商品分组，每条接口附有中文用途说明；外层统一使用 `/api` 前缀。

打包时保留 `.env.example`、迁移源码和 `go.mod/go.sum`；排除 `.env`、编译产物，以及前端的 `node_modules`、`dist`。手动压缩不会自动遵循 `.gitignore`。
