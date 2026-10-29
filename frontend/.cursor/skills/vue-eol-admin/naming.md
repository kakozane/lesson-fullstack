# 命名规范

| 类型 | 规范 | 示例 |
| --- | --- | --- |
| 文件夹 | kebab-case | `user-manage/` |
| Vue 组件文件 | kebab-case 或 `index.vue` | `user-list.vue`、`index.vue` |
| TS/JS 文件 | kebab-case | `use-pagination.ts` |
| 路由 name | PascalCase | `UserManage` |
| 路由 path | kebab-case | `/system/user-manage` |
| 组件名 (defineOptions) | PascalCase | `UserListPage` |
| API 函数 | camelCase + `Api` 后缀 | `getUserListApi` |
| Composable | `use` 前缀 + camelCase | `usePagination` |
| 常量 | UPPER_SNAKE_CASE | `MAX_PAGE_SIZE` |

## 路由 meta 常用字段

- `title`：菜单/页面标题
- `icon`：图标，格式 `lucide:<name>`
- `order`：菜单排序（数字越小越靠前）
- `hideInMenu`：详情页等不在菜单显示
- `hideChildrenInMenu`：子路由折叠为单菜单项
- `activePath`：详情页高亮父级菜单路径
