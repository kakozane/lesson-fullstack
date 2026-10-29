# Vue 组件编写规范

`<script setup>` 内代码按以下顺序组织：

```vue
<script lang="ts" setup>
// 1. 类型导入
import type { UserInfo } from '#/api/system/user';

// 2. Vue / 第三方库导入
import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';

// 3. 内部模块导入
import { getUserListApi } from '#/api/system/user';

// 4. Props / Emits
interface Props {
  userId?: number;
}
const props = withDefaults(defineProps<Props>(), { userId: 0 });
const emit = defineEmits<{ success: [] }>();

// 5. 响应式数据
const loading = ref(false);
const dataList = ref<UserInfo[]>([]);

// 6. 计算属性
const isEmpty = computed(() => dataList.value.length === 0);

// 7. 方法
async function loadData() {
  loading.value = true;
  try {
    dataList.value = await getUserListApi({ page: 1, pageSize: 10 });
  } finally {
    loading.value = false;
  }
}

// 8. 生命周期
onMounted(() => loadData());
</script>

<template>
  <!-- 模板内容 -->
</template>
```

## 要点

- 使用 `lang="ts"` + `setup`
- Element Plus 组件按需导入，不要全局注册
- 异步请求用 `loading` + `try/finally` 模式
- 需要页面级组件名时用 `defineOptions({ name: 'UserListPage' })`（PascalCase）
- 类型优先从 API 模块导入（`import type`）
