# 数据库初始化与迁移

本目录保存数据库连接、建库、建表与依赖装配的 Go 代码；假数据 SQL 单独保存在同级 `migration/` 目录。迁移使用 Go + 显式 SQL，在后端启动时执行，无需先手动导入数据库备份。

**假数据保存在两个独立 SQL 文件中**：`migration/001_products_seed.sql`（36 条商品）、`migration/002_tasks_seed.sql`（9 条任务）。Go 使用 `go:embed` 将它们编译进程序，在首次初始化时执行；发布编译后的程序时，不依赖启动目录查找 SQL 文件。

## 文件职责

以下路径相对于 `backend/`：

| 文件 | 职责 |
|---|---|
| `bootstrap/database.go` | 连接 MySQL，创建配置指定的数据库，初始化连接池与 GORM |
| `bootstrap/app.go` | 按商品、任务、分类迁移的顺序初始化数据库，再装配业务依赖 |
| `bootstrap/migrations.go` | 创建商品表与迁移记录表，版本 1 首次插入 36 条演示商品 |
| `bootstrap/task_migration.go` | 创建任务表与看板版本表，版本 2 首次插入 9 条演示任务 |
| `bootstrap/category_migration.go` | 版本 3 创建分类表，将旧商品中文分类转换为分类 ID |
| `migration/001_products_seed.sql` | 商品假数据，价格单位为分 |
| `migration/002_tasks_seed.sql` | 任务假数据，截止日期以执行当天为基准 |
| `bootstrap/migrations_test.go` | 独立临时 MySQL 库验证首次初始化及重复启动 |

`migration/` 只保存 SQL。`main.go` 使用 `go:embed` 嵌入 SQL 内容并传给 `bootstrap.InitializeApp`，后者负责初始化与依赖装配。业务实体和仓储仍位于 `internal/product`、`internal/task` 下。

启动调用顺序：

```text
main → config.Load → bootstrap.InitializeApp（传入两份 SQL 内容）
                       ├─ LaunchDatabase：连接与建库
                       ├─ initialize：商品表、版本 1 种子
                       ├─ initializeTasks：任务表、版本 2 种子
                       ├─ initializeCategories：分类表、版本 3 分类迁移
                       └─ 创建 repository / service / handler
```

## 在新机器运行

在 `backend` 目录操作：

```sh
# 首次运行时复制；已有 .env 时直接检查配置，不要覆盖
cp .env.example .env
# 编辑 .env，填写本机 MySQL 地址、账号、密码及数据库名
go run .
```

Go 会自动读取当前工作目录的 `.env`，请在 `backend` 目录执行 `go run .`，无需先 `source .env`。系统环境变量优先；没有 `.env` 时可使用系统环境变量，文件无法读取或格式错误时会停止启动。配置解析使用 [godotenv](https://github.com/joho/godotenv)，原有密码与数据库名校验保持不变。`run.sh` 仅作为可选的启动入口。

先启动 MySQL。账号需能创建配置指定的数据库和表，并读写课程数据。数据库名仅使用小写字母、数字和下划线；默认是 `lesson06_shop`。Go 服务启动成功后会自动准备下列表：

| 表 | 内容 |
|---|---|
| `products` | 商品名称、编码、分类、整数分价格、库存与上下架状态 |
| `categories` | 商品分类字典，商品通过 category_id 关联 |
| `tasks` | 任务内容、状态、优先级、负责人、截止日期与列内位置 |
| `task_board_meta` | 当前看板版本，任务写入用它检查并发冲突 |
| `lesson_migrations` | 已完成的迁移版本，当前为 1、2、3 |

已有 MySQL 服务可以直接使用，不必为分享项目导出本机业务数据。首次创建的示例内容由迁移文件提供。

这不是 Go 语言自带的行为，而是本项目 `LaunchDatabase`、`initialize`、`initializeTasks` 和 `initializeCategories` 主动执行建库、建表与种子 SQL 的结果。程序不会安装或启动 MySQL；MySQL 未启动、地址或凭据错误、账号权限不足时，初始化会失败并返回错误。

## 如何使用假数据 SQL

正常运行课程项目时，只需启动 Go 服务，**不需要再在 Navicat 手动执行这两个 SQL**。它们不包含建库、建表、清空数据或迁移版本登记语句，版本控制与事务由 Go 负责。

如需在其他练习库手动导入，先准备与本项目一致的表结构，并确认对应的商品表或任务表为空，再在 Navicat 选中该练习库执行相应 SQL。商品编码唯一，重复导入商品会报冲突；任务没有同类唯一约束，重复导入会生成重复任务。手动导入的练习库不要随后交给本项目进行首次种子初始化，因为没有迁移版本记录，会再次尝试插入。

SQL 文件随源码分享。修改文件后需要重新运行 `go run` 或重新编译；已有迁移版本的数据库不会因此重新导入。课堂需要全新演示数据时，在 `.env` 中换一个新的 `DB_NAME` 后重新启动即可。

## 重复启动会发生什么

每次启动会执行 `CREATE DATABASE/TABLE IF NOT EXISTS`，已有数据库和表不重建；`task_board_meta` 仅在版本行不存在时初始化为 1。

每组种子数据与对应的 `lesson_migrations` 版本登记在同一事务中写入。版本已存在时跳过种子插入，因此：

- 修改过的商品、任务和任务日期不会被覆盖。
- 删除的演示数据不会因重启被补回。
- 已递增的看板版本不会被重置。
- 种子插入失败会回滚该组种子及版本记录，下次启动可以重试。

MySQL 的建库、建表操作不属于上述种子事务，不能将其理解为全部 DDL 都能回滚。当前使用显式 SQL，不通过 GORM AutoMigrate 自动变更已有表。

需要重新练习时，可在 `.env` 中选择一个新的课程数据库名。保留旧库与版本记录，避免删除版本记录后与已有唯一编码发生冲突。

## 添加后续迁移

后续表结构变更放在 `bootstrap/` 中，并在 `InitializeApp` 中明确调用顺序；假数据 SQL 放在 `migration/` 中，由 `main.go` 嵌入并传入。文件不会根据名称自动扫描执行。

已经共享的迁移文件应保留其含义。后续字段变更使用新版本和明确 SQL，考虑已存在数据、重复执行及 DDL 失败后的恢复方式；历史种子值不应直接依赖未来可能变化的业务选项。

## 验证

普通测试与构建：

```sh
go test -race ./...
go vet ./...
go build ./...
```

真实 MySQL 测试默认跳过；确认本地 MySQL 可用后，在 `backend` 目录执行：

```sh
set -a
. ./.env
set +a
RUN_MYSQL_MIGRATION_TEST=1 go test -race ./bootstrap -run TestRunMySQL -count=1 -v
```

测试只读取连接地址与凭据，忽略 `.env` 中的 `DB_NAME`。它创建 `lesson06_migration_test_*` 临时库，验证 36 条商品、9 条任务、3 个分类和三条版本记录，再修改、删除数据并重复初始化，检查数据和看板版本是否保留，并验证旧分类迁移、分类筛选和单个商品上下架。正常结束会清理本次临时库；测试账号需要建库和删库权限。

## 分享文件

分享 `migration` 中的 SQL、`bootstrap` 和其他 Go 源码、`go.mod`、`go.sum`、`run.sh` 和 `.env.example`。项目的 `.gitignore` 已忽略 `backend/.env`，但手动压缩目录不会自动应用 Git 忽略规则，应自行排除真实 `.env`、编译产物和本地测试输出。

### 分类迁移（版本 3）

`category_migration.go` 创建 `categories` 表，将旧商品分类名称转换成 `products.category_id`，保留已有商品和更新时间。重复启动不会重新填充或覆盖分类。历史 `category` 列保留用于旧库升级兼容，业务不再读写它；`migration/001_products_seed.sql` 已使用数字分类 ID，新环境直接写入 ID。
