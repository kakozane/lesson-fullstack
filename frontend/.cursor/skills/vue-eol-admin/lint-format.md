# Lint 与格式化

- Oxlint：JS、TS 和 Vue 脚本静态检查，配置为根目录 `oxlint.config.ts` 和共享包 `@vben/oxlint-config`。
- Oxfmt：统一格式化，配置为根目录 `oxfmt.config.ts` 和共享包 `@vben/oxfmt-config`。
- Stylelint：独立检查 CSS、SCSS 和 Vue 样式。
- Vue 模板和类型检查仍由 vue-tsc 负责，不等同于 Oxlint 的脚本检查。

在 frontend 目录执行：

```bash
pnpm lint
pnpm lint:fix
pnpm format
pnpm format:check
pnpm lint:style
pnpm check:type
```

安装 VS Code 的 Oxc 扩展后，保存时使用 Oxfmt 格式化，并应用 Oxlint 修复。工作区禁用 ESLint。共享 Oxlint 配置使用原生规则，不加载 ESLint JavaScript 插件。Lefthook 配置在提交前格式化暂存文件并运行 Oxlint；需要 Git 仓库和已安装的钩子才会执行。
