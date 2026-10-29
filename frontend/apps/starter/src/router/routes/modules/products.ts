import type { RouteRecordRaw } from 'vue-router';
const routes: RouteRecordRaw[] = [
  {
    name: 'Products',
    path: '/products',
    component: () => import('#/views/products/index.vue'),
    meta: {
      title: '商品管理',
      icon: 'lucide:package',
      keepAlive: false,
      order: 1,
    },
  },
];
export default routes;
