# Git 提交规范 (Conventional Commits)

## 格式

```
<type>(<scope>): <subject>
```

## 常用 type

| type       | 用途                   |
| ---------- | ---------------------- |
| `feat`     | 新功能                 |
| `fix`      | 修复 bug               |
| `docs`     | 文档变更               |
| `style`    | 代码格式（不影响逻辑） |
| `refactor` | 重构                   |
| `perf`     | 性能优化               |
| `test`     | 添加测试               |
| `chore`    | 构建/工具变动          |

## 示例

```
feat(order): 添加订单列表页面
fix(auth): 修复 token 刷新后重复请求问题
refactor(api): 统一接口返回类型定义
```

## 工具

- 提交前自动 lint：Lefthook（`lefthook.yml`）

## scope 建议

使用模块名：`order`、`auth`、`api`、`demo` 等，与变更范围对应。
