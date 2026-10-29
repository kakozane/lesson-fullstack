# 样式编写规范

**优先 Tailwind CSS 工具类，复杂样式用 scoped style。**

```vue
<!-- 推荐 -->
<div class="flex items-center gap-2 rounded-lg bg-white p-4 shadow">
  <span class="text-lg font-bold text-gray-800">标题</span>
</div>

<!-- Element Plus 覆盖用 :deep() -->
<style scoped>
:deep(.el-table__header) {
  background-color: #f5f7fa;
}
</style>
```

## 规则

- 布局、间距、颜色、字体优先 Tailwind class
- 仅当 Tailwind 无法表达或需覆盖第三方组件样式时用 `<style scoped>`
- 覆盖 Element Plus 内部类名必须用 `:deep()`
- 避免全局样式污染
