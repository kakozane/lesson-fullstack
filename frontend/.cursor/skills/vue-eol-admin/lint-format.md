# Lint & 格式化工具链

| 工具      | 配置包                   | 作用                         |
| --------- | ------------------------ | ---------------------------- |
| ESLint    | `@vben/eslint-config`    | JS/TS/Vue/JSON/YAML 静态检查 |
| OxLint    | `@vben/oxlint-config`    | 高性能 lint                  |
| OxFmt     | `@vben/oxfmt-config`     | 代码格式化（替代 Prettier）  |
| Stylelint | `@vben/stylelint-config` | CSS/SCSS 样式检查            |
| Lefthook  | 根目录 `lefthook.yml`    | 提交前自动 lint              |

## OxFmt 配置

| 选项          | 值    |
| ------------- | ----- |
| printWidth    | 80    |
| semi          | true  |
| singleQuote   | true  |
| trailingComma | all   |
| proseWrap     | never |

## 命令

```bash
pnpm lint      # lint 检查
pnpm format    # 代码格式化
```

## 开发注意

- 提交前 Lefthook 会自动跑 lint，失败需修复后重新提交
- 格式化统一走 `pnpm format`，不要手动混用其他 formatter
- Vue template 的 HTML 换行/缩进由 OxFmt 负责；ESLint 已关闭 `html-closing-bracket-newline`、`multiline-html-element-content-newline` 等与 formatter 冲突的规则
