# mise：全栈开发者的多语言版本管理

前端需要 Node.js，后端需要 Go，另一个服务又依赖 Java。项目越多，越容易遇到“装过这个工具，但当前版本不对”的问题。

mise 可以安装并选择不同版本的开发工具，把项目所需版本写进配置文件。配合 Shell 激活，进入不同项目目录时就能选择对应工具；它还可以管理环境变量和常用任务。[mise 官方介绍](https://mise.jdx.dev/)

本讲义以 **Vue3 + Go** 为主线，介绍多语言环境的统一管理方法。示例采用 macOS / Linux 终端语法，安装与激活部分同时提供 Windows PowerShell 用法。

## 1. 为什么全栈项目适合使用 mise

考虑三个独立项目：

| 项目 | 前端或脚本环境 | 后端环境 |
|---|---|---|
| 项目 A：维护已有后台 | Node.js 22 | Go 1.25 |
| 项目 B：Vue + Go 实战 | Node.js 24 | Go 1.26 |
| 项目 C：Java 服务 | Node.js 24 | Java 21 |

这些版本是练习场景，不表示所有项目都应该升级到它们。实际选择应依据依赖和构建要求。

多语言版本管理要解决三个问题：

1. **同一台机器安装多个版本**，无需为切换项目反复卸载。
2. **项目明确声明使用哪个版本**，减少依赖个人记忆的配置。
3. **团队按同一份配置准备环境**，让开发与构建过程更容易复现。

mise 的核心是工具版本选择。Node 项目依赖仍由 pnpm/npm 管理，Go 依赖仍由 Go Modules 管理；MySQL 等服务可以继续使用 Docker。这些能力分别作用于不同层次。

```text
mise.toml                 → Node.js、Go 等工具的版本
package.json + pnpm-lock.yaml → 前端依赖与包管理器约定
go.mod + go.sum         → Go 模块依赖
Docker / Compose        → MySQL 等服务的运行环境
```

## 2. 安装与激活

### 2.1 安装 mise

macOS 已安装 Homebrew 时，可以使用：

```sh
brew install mise
mise --version
```

macOS / Linux 也可以使用官方安装脚本。两种安装方式选一种即可：

```sh
curl -fsSL https://mise.run | sh
~/.local/bin/mise --version
```

脚本安装的默认可执行文件位于 `~/.local/bin/mise`。在 PATH 尚未配置好时，可以先使用完整路径。Windows PowerShell 可用：

```powershell
winget install jdx.mise
```

安装后重新打开终端，再运行 `mise --version`。[官方安装说明](https://mise.jdx.dev/installing-mise.html)

### 2.2 让终端自动选择项目版本

安装完成后，还需要为日常使用的交互式 Shell 配置激活。将对应的一行加入启动文件，已有时不重复添加：

| Shell | 配置文件 | 添加的内容 |
|---|---|---|
| Zsh | `~/.zshrc` | `eval "$(mise activate zsh)"` |
| Bash | `~/.bashrc` | `eval "$(mise activate bash)"` |
| Fish | `~/.config/fish/config.fish` | `mise activate fish \| source` |
| PowerShell | `$PROFILE` 指向的文件 | `(&mise activate pwsh) \| Out-String \| Invoke-Expression` |

使用官方脚本安装、且启动时找不到 `mise` 的 Zsh 用户，可将对应行写成：

```sh
# 添加到 ~/.zshrc；替代上表中的 Zsh 行
# 若自定义了 ZDOTDIR，应编辑相应目录中的 .zshrc
eval "$(~/.local/bin/mise activate zsh)"
```

保存后新开一个终端，再检查：

```sh
mise --version
mise doctor
```

激活会在交互式终端更新环境。已经运行的服务不会因为切换目录就更换运行时，修改版本后应重新启动相关进程。[Shell 配置说明](https://mise.jdx.dev/shell-setup.html)、[activate 命令](https://mise.jdx.dev/cli/activate.html)

### 2.3 暂时不配置 Shell 也能使用

```sh
mise exec node@24 -- node --version
```

这条命令使用 Node.js 24 执行一次版本检查，必要时下载工具，不把版本选择写入项目配置，也不改变当前终端的环境。`--` 后面是实际执行的命令。脚本和 CI 中也可以采用这种明确的调用方式。[exec 命令](https://mise.jdx.dev/cli/exec)

## 3. 分清 install、use 和 exec

| 目的 | 命令示例 | 是否声明项目版本 |
|---|---|---|
| 只安装某个版本 | `mise install node@24` | 不修改项目版本声明 |
| 为项目选择并安装工具 | `mise use --path ./mise.toml node@24` | 写入指定配置 |
| 安装项目已有配置中的工具 | `mise install` | 沿用已有声明 |
| 临时使用某个版本 | `mise exec node@22 -- node --version` | 不修改项目版本声明 |
| 设置个人默认版本 | `mise use --global node@24` | 写入全局配置 |

**安装完成不等于当前终端已经使用它。** 当前命令使用哪个版本，还取决于项目配置、Shell 激活与 PATH。[开发工具管理](https://mise.jdx.dev/dev-tools/)

本讲义在创建项目配置时明确加 `--path ./mise.toml`，避免把修改写到上级目录已有配置。实际使用裸 `mise use` 时，它会根据当前目录和已有配置选择写入目标；在子目录运行不一定新建子目录配置。[use 命令](https://mise.jdx.dev/cli/use.html)

## 4. 第一个练习：为项目固定 Node.js 和 Go

### 4.1 创建独立练习目录

在自己的工作目录下执行：

```sh
mkdir mise-fullstack-demo
cd mise-fullstack-demo
mise use --path ./mise.toml --pin node@24 go@1.26
```

`--pin` 会把解析出的完整版本号写入 `mise.toml`。例如，配置可能是：

```toml
[tools]
node = "24.21.0"
go = "1.26.2"
```

这两个完整版本在本文核验环境中已安装；它们用于说明固定版本的写法，不代表各语言的最新版本。你的解析结果可能不同，以生成的文件为准。[use 的版本固定规则](https://mise.jdx.dev/cli/use.html)

### 4.2 检查实际生效结果

```sh
mise ls --current
mise which node
mise which go
mise exec -- node --version
mise exec -- go version
```

已激活 Shell 时，也可直接运行：

```sh
node --version
go version
```

`mise which` 用于查看 mise 选中的可执行文件；普通命令是否使用相同文件，还要看 Shell 环境。两种执行方式结果应一致。

### 4.3 范围版本与完整版本有什么区别

| 配置 | 含义 | 适用情形 |
|---|---|---|
| `node = "24"` | 允许使用 24 系列版本 | 接受该系列内版本变化 |
| `node = "24.21.0"` | 明确指定一个完整版本 | 课程、团队基线与可复现排查 |
| `node = "latest"` | 请求最新版本 | 独立尝试新环境 |

`"24"` 不意味着所有机器必然拿到同一个补丁版本，也不意味着每次进入目录都会重新下载最新版本。需要一致的团队基线时，应使用完整版本固定，或采用锁文件。

可选的锁文件工作流：

```sh
mise lock
mise install --locked
```

`mise.lock` 记录解析结果；使用时需要覆盖团队和 CI 的目标平台。本文基础练习采用 `--pin`，锁文件可以在进一步规范团队环境时加入。[锁文件说明](https://mise.jdx.dev/dev-tools/mise-lock.html)

## 5. 第二个练习：切换目录，切换 Node.js 版本

在一个新建的 `mise-switch-demo` 目录内准备两个独立项目：

```sh
mkdir mise-switch-demo
cd mise-switch-demo
mkdir app-a app-b

cd app-a
mise use --path ./mise.toml --pin node@22

cd ../app-b
mise use --path ./mise.toml --pin node@24
```

安装完成后，在**已激活 mise 的交互式终端**中，分次执行并观察：

```sh
cd ../app-a
node --version

cd ../app-b
node --version
```

预期分别输出 `v22...` 和 `v24...`。每次 `cd` 后等待提示符出现，再执行下一条命令，便于观察激活后的环境变化。

在脚本或一口气粘贴多条命令时，使用下面的写法更明确：

```sh
mise -C ../app-a exec -- node --version
mise -C ../app-b exec -- node --version
```

`-C` 指定工作目录，`exec` 按该目录的配置运行命令，不依赖提示符何时更新。[activate 命令](https://mise.jdx.dev/cli/activate.html)、[exec 命令](https://mise.jdx.dev/cli/exec)

## 6. 扩展到其他语言：以 Java 为例

以下命令用于各自的独立练习目录，按需要选择，不必为 Vue + Go 项目安装所有语言。

### Java

```sh
mise use --path ./mise.toml --pin java@temurin-21
mise exec -- java -version
mise exec -- javac -version
```

Java 需要同时关注版本与发行商，`temurin-21` 明确了这两项。mise 的 Java 支持还会为所选 JDK 设置 `JAVA_HOME`。[Java 支持](https://mise.jdx.dev/lang/java.html)

其他语言和 CLI 工具可在 [工具注册表](https://mise.jdx.dev/registry.html) 中查找。不同工具、平台的安装方式与系统依赖可能不同，确认支持范围后再接入项目。

## 7. 应用到 Vue3 + Go 课程项目

### 7.1 配置放在什么位置

对于前端和后端位于同一仓库的项目，可以从根目录统一声明工具：

```text
lesson-06/
├── mise.toml
├── frontend/
│   ├── package.json
│   ├── pnpm-lock.yaml
│   └── apps/
│       ├── admin/
│       └── starter/
└── backend/
    ├── go.mod
    └── run.sh
```

在普通目录配置继承规则下，子目录会读取上级配置，因此 frontend 和 backend 可以共用根目录的工具版本。如果子目录另有配置，会发生合并或覆盖，可用 `mise config ls` 查看实际读取来源。本文不需要启用 mise 的额外 monorepo 任务功能。[配置文件与继承](https://mise.jdx.dev/configuration.html)

当前 lesson-06 的源码约定为：

- 前端 `engines.node`：`^22.18.0 || ^24.0.0`。
- 前端 `packageManager`：`pnpm@11.4.0`。
- 后端 `go.mod`：`go 1.26`。

这来自课程项目文件，不是所有 Vue、Vben 或 Gin 项目的统一要求。Node.js 的 engines 是兼容范围，mise 固定的是实际使用版本，两者应保持一致。

可在项目根目录创建：

```sh
mise use --path ./mise.toml --pin node@24 go@1.26
```

接着检查生成的 `mise.toml`，确认工具版本符合项目约定，再将它纳入团队版本管理。本文只提供接入步骤，课程项目是否已接入应以根目录实际文件为准。

### 7.2 安装并使用 pnpm

mise 负责选择 Node.js、Go 版本，pnpm 负责安装前端依赖和运行 workspace 脚本。项目的 `frontend/package.json` 通过 `packageManager` 声明使用 `pnpm@11.4.0`。

选好 Node.js 版本后，安装 pnpm：

```sh
npm install -g pnpm@11.4.0
pnpm --version
```

然后进入项目的 frontend 目录：

```sh
cd frontend
pnpm install --frozen-lockfile
pnpm dev:starter
```

## 8. 进阶：把常用命令变成项目任务

在项目根目录的 `mise.toml` 中保留已有 `[tools]`，追加下面的任务。若文件已有同名任务，应修改原定义，不重复粘贴：

```toml
[tasks."front:dev"]
description = "启动课堂初始版前端"
dir = "{{config_root}}/frontend"
run = "pnpm dev:starter"

[tasks."front:check"]
description = "检查课堂前端类型"
dir = "{{config_root}}/frontend"
run = "pnpm typecheck:starter"

[tasks."front:build"]
description = "构建课堂前端"
dir = "{{config_root}}/frontend"
run = "pnpm build:starter"

[tasks."back:dev"]
description = "启动 Gin 后端"
dir = "{{config_root}}/backend"
run = "go run ."

[tasks."back:test"]
description = "运行 Go 测试"
dir = "{{config_root}}/backend"
run = "go test ./..."
```

`dir` 明确任务工作目录；`config_root` 对应配置所属根目录。这避免在项目根目录直接运行 pnpm 时，未读到 frontend 内的包管理器约定。后端任务直接执行 `go run .`，程序自动读取工作目录中的 `.env`。[TOML 任务](https://mise.jdx.dev/tasks/toml-tasks.html)、[任务配置参考](https://mise.jdx.dev/tasks/task-configuration.html)

在项目根目录查看与执行任务：

```sh
mise tasks ls
mise run front:check
mise run back:test
```

开发服务需要长期运行，因此分别开两个终端：

```sh
# 终端一
mise run back:dev
```

```sh
# 终端二
mise run front:dev
```

这些任务要求已安装前端依赖，且数据库和后端 `.env` 已准备好。`back:dev` 使用 Bash；原生 Windows 环境需要准备兼容 Shell，或在 WSL 中执行课程项目。安装 mise 本身不会自动完成这些应用配置。

## 9. 环境变量与配置共享

mise 还支持在配置中声明非敏感环境变量：

```toml
[env]
APP_ENV = "development"
```

这个变量只作为环境加载示例；应用需要读取它才会产生业务效果。

也可以加载项目自己的环境文件：

```toml
[env]
_.file = ".env"
```

以上是同一个 `[env]` 表的不同示例，需要组合时把字段合并到同一张表中。环境文件路径相对于配置根目录；`.env` 中的密码或令牌不应提交到仓库。课程后端已经由 Go 程序自动读取工作目录中的 `.env`（即 `backend/.env`），无需再为此在根目录重复配置。[环境变量说明](https://mise.jdx.dev/environments/)

团队共享文件可以这样区分：

| 文件 | 处理方式 |
|---|---|
| `mise.toml` | 共享工具版本与任务，纳入版本管理 |
| `mise.lock` | 采用锁文件工作流时一并提交 |
| `mise.local.toml` | 个人覆盖配置，加入 `.gitignore` |
| `.env` | 保留在本地，使用不含真实密钥的示例文件说明字段 |

常规情况下，项目配置覆盖个人全局默认，同目录的本地配置覆盖共享配置；特殊 override 配置还有额外优先级，遇到意外结果以 `mise config ls` 为排查起点。[配置优先级](https://mise.jdx.dev/configuration.html)

拿到他人的项目后，先阅读配置中的任务、环境加载与安装钩子。遇到未信任配置提示时，确认来源和内容后执行：

```sh
mise trust ./mise.toml
```

不同版本及安全设置的信任行为可能不同；当前部分执行命令会自动信任活动配置，因此不能把“没有弹出确认”理解成“配置不会执行代码”。[配置信任说明](https://mise.jdx.dev/cli/trust.html)

## 10. 常见问题排查

### 10.1 mise 选择了新版本，普通命令还是旧版本

在 Bash / Zsh 中检查：

```sh
mise ls --current
mise config ls
mise which node
type -a node
mise exec -- node --version
node --version
mise doctor
```

如果 `mise exec` 正确、普通 `node` 错误，优先检查 Shell 激活、别名与 PATH 顺序。nvm、其他版本管理器或系统安装的 Node.js 可能提供了更靠前的入口。先确认由哪套工具负责选择版本，再调整初始化配置。

已存在 `.tool-versions` 的项目可以使用其配置；`.nvmrc` 等语言专属文件需要开启相应识别设置。不要默认所有旧版本文件都会自动生效。[开发工具与版本文件](https://mise.jdx.dev/dev-tools/)

### 10.2 终端正确，IDE 运行或调试版本不对

IDE 的插件、语言服务、集成终端可能是不同进程，不能仅凭外部终端的输出判断它们的版本。

先在 IDE 集成终端检查版本，再检查语言插件的解释器或 SDK 路径。可以根据 `mise which` 的结果选择对应工具，或按编辑器集成文档使用 shims。固定到某个具体版本路径后，升级工具时也要检查 IDE 配置。[编辑器集成](https://mise.jdx.dev/ide-integration.html)

### 10.3 Go 的实际版本和 mise 声明不完全相同

Go 自身也有工具链选择机制。`GOTOOLCHAIN=auto` 时，`go` 命令可能根据 `go.mod` 或 `go.work` 的要求选择、下载另一套工具链：

```sh
mise which go
mise exec -- go version
mise exec -- go env GOTOOLCHAIN
```

需要区分“mise 选中的 go 入口”和“Go 最终执行的工具链”。严格固定时可以研究 `GOTOOLCHAIN=local`，但应先保证固定版本满足模块要求，否则构建会报版本不足。[Go 工具链说明](https://mise.jdx.dev/lang/go.html)

### 10.4 切换 Node.js 后找不到 pnpm

全局安装的命令可能与 Node.js 版本关联。切换版本后，先运行 `node --version` 和 `pnpm --version` 检查当前环境。如果找不到 pnpm，在当前 Node.js 环境重新执行 `npm install -g pnpm@11.4.0`，再进入 `frontend` 目录运行项目。

### 10.5 工具下载失败

先区分是版本不存在、网络访问失败，还是当前平台缺少对应安装包。可用 `mise ls-remote node` 查看版本，并阅读错误中实际请求的地址；不同工具可能访问不同的上游。下载失败时先解决明确的错误，不盲目重装整个开发环境。

## 11. 日常命令速查

| 需求 | 命令 |
|---|---|
| 查看 mise 版本 | `mise --version` |
| 检查环境 | `mise doctor` |
| 查看当前工具版本 | `mise ls --current` |
| 查看已安装的版本 | `mise ls --installed` |
| 查看可安装的 Node.js 版本 | `mise ls-remote node` |
| 查看配置来源 | `mise config ls` |
| 查看工具路径 | `mise which node` |
| 安装项目声明的工具 | `mise install` |
| 固定项目版本 | `mise use --path ./mise.toml --pin node@24` |
| 临时使用指定版本 | `mise exec node@22 -- node --version` |
| 查看任务 | `mise tasks ls` |
| 运行任务 | `mise run front:check` |
| 查看可更新工具 | `mise outdated` |
| 在配置允许范围内升级 | `mise upgrade node` |

精确固定版本不会因为普通升级就自动变成新的版本号。变更团队基线时，应在 `mise use --path ./mise.toml node@完整版本号` 中填入已验证的目标版本，再检查配置差异和项目构建结果。只重新选择 `node@24` 可能复用已安装的匹配版本，不能据此判断已升级到最新补丁。[升级命令](https://mise.jdx.dev/cli/upgrade.html)、[use 命令](https://mise.jdx.dev/cli/use.html)

## 12. 学习自查与练习

完成后应能回答：

1. `mise install node@24` 和 `mise use node@24` 有什么区别？
2. 为什么 `node = "24"` 不足以保证所有机器使用相同的补丁版本？
3. 为什么自动切换需要 Shell 配合，而 `mise exec` 可以直接用于脚本？
4. 为什么安装了 Node.js 和 Go，还需要 pnpm install、Go Modules 和数据库？
5. 本项目的 Node.js 版本与 pnpm 版本，分别由哪些配置约定？

练习任务：

- 在两个独立目录分别固定 Node.js 22 和 24，验证目录切换后的结果。
- 为一个 Vue + Go 项目添加 `mise.toml`，检查选择结果是否符合 `package.json` 与 `go.mod`。
- 定义前端检查与后端测试任务，从项目根目录执行它们。
- 将配置交给另一台机器，完成安装与版本核对，记录仍需单独配置的依赖和服务。

复习时可以从 [官方入门指南](https://mise.jdx.dev/getting-started) 开始；具体参数以本机 `mise <命令> --help` 和对应版本官方文档为准。
