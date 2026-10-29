# 新增业务模块标准流程

以「订单管理」为例，三步完成：API → 页面 → 路由。

## Step 1: 定义接口

路径：`apps/web/src/api/order/index.ts`

```typescript
import { requestClient } from '#/api/request';

export interface OrderInfo {
  id: number;
  orderNo: string;
  status: 'cancelled' | 'completed' | 'pending' | 'shipped';
  amount: number;
  createdAt: string;
}

export interface OrderListParams {
  page: number;
  pageSize: number;
  status?: string;
  keyword?: string;
}

export function getOrderListApi(params: OrderListParams) {
  return requestClient.get('/order/list', { params });
}

export function getOrderDetailApi(id: number) {
  return requestClient.get<OrderInfo>(`/order/${id}`);
}

export function updateOrderStatusApi(id: number, status: string) {
  return requestClient.put(`/order/${id}/status`, { status });
}
```

多文件时可拆分子模块并在 `index.ts` 统一导出；在 `apps/web/src/api/index.ts` 按需 re-export。

## Step 2: 创建页面

路径：`apps/web/src/views/order/index.vue`

```vue
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import {
  ElButton,
  ElCard,
  ElMessage,
  ElTable,
  ElTableColumn,
  ElTag,
} from 'element-plus';
import { getOrderListApi, updateOrderStatusApi } from '#/api/order';

const loading = ref(false);
const orderList = ref([]);

async function loadOrders() {
  loading.value = true;
  try {
    const res = await getOrderListApi({ page: 1, pageSize: 20 });
    orderList.value = res.list;
  } finally {
    loading.value = false;
  }
}

async function handleShip(id: number) {
  await updateOrderStatusApi(id, 'shipped');
  ElMessage.success('发货成功');
  loadOrders();
}

onMounted(() => loadOrders());
</script>

<template>
  <div class="p-4">
    <ElCard>
      <ElTable v-loading="loading" :data="orderList">
        <ElTableColumn prop="orderNo" label="订单号" />
        <ElTableColumn prop="amount" label="金额" />
        <ElTableColumn prop="status" label="状态">
          <template #default="scope">
            <ElTag>{{ scope.row.status }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作">
          <template #default="scope">
            <ElButton size="small" @click="handleShip(scope.row.id)">
              发货
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </div>
</template>
```

子页面（详情/编辑）放同目录，如 `detail.vue`、`form.vue`。

## Step 3: 配置路由

路径：`apps/web/src/router/routes/modules/order.ts`

```typescript
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:shopping-cart',
      order: 20,
      title: '订单管理',
    },
    name: 'Order',
    path: '/order',
    children: [
      {
        name: 'OrderList',
        path: '/order/list',
        component: () => import('#/views/order/index.vue'),
        meta: { icon: 'lucide:list', title: '订单列表' },
      },
      {
        name: 'OrderDetail',
        path: '/order/detail/:id',
        component: () => import('#/views/order/detail.vue'),
        meta: {
          hideInMenu: true,
          title: '订单详情',
          activePath: '/order/list',
        },
      },
    ],
  },
];

export default routes;
```

保存后菜单自动出现，无需额外注册。

## 列表 + 筛选 + 详情模式

参考现有 demo 模块 `apps/web/src/router/routes/modules/demo.ts`：

- 列表页用 `hideChildrenInMenu: true` 折叠子路由
- 详情/编辑页设 `hideInMenu: true` + `activePath` 指向列表
- 路由跳转用 `router.push({ name: 'OrderDetail', params: { id } })`
