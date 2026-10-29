import type { RouteRecordRaw } from 'vue-router';
const routes: RouteRecordRaw[] = [
  {
    name: 'Tasks',
    path: '/tasks',
    component: () => import('#/views/tasks/index.vue'),
    meta: {
      title: '任务看板',
      icon: 'lucide:kanban',
      keepAlive: false,
      order: 2,
    },
  },
];
export default routes;
