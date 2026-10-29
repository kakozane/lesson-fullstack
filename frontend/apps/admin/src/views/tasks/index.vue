<script setup lang="ts">
import type {
  Task,
  TaskInput,
  TaskMove,
  TaskPriority,
  TaskStatus,
} from '#/api/tasks';
import { computed, onMounted, ref } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { useClipboard } from '@vueuse/core';
import { PageHeader } from '@vben/common';
import {
  ElAlert,
  ElButton,
  ElCard,
  ElForm,
  ElFormItem,
  ElInput,
  ElMessage,
  ElMessageBox,
  ElOption,
  ElSelect,
  ElSwitch,
} from 'element-plus';
import { useTaskBoard } from '#/composables/use-task-board';
import TaskBoard from './components/task-board.vue';
import TaskFormDrawer from './components/task-form-drawer.vue';
import { isOverdue, priorities, statuses } from './task-meta';

const { board, loading, pending, failed, ready, refresh, save, remove, move } =
  useTaskBoard();
const keyword = ref('');
const priority = ref<'' | TaskPriority>('');
const overdueOnly = ref(false);
const dragging = ref(false);
const confirming = ref(false);
const disabled = computed(
  () => loading.value || pending.value || confirming.value || !ready.value,
);
const filtering = computed(
  () => !!keyword.value.trim() || !!priority.value || overdueOnly.value,
);
const visibleTasks = computed(() =>
  board.value.items.filter((task) => {
    const query = keyword.value.trim().toLocaleLowerCase();
    return (
      (!query ||
        `${task.title} ${task.assignee} #${task.id}`
          .toLocaleLowerCase()
          .includes(query)) &&
      (!priority.value || task.priority === priority.value) &&
      (!overdueOnly.value || isOverdue(task))
    );
  }),
);
const completed = computed(() => board.value.completed);
const doing = computed(() => board.value.doing);
const overdue = computed(() => board.value.overdue);
const open = ref(false);
const editing = ref<Task>();
const initialStatus = ref<TaskStatus>('todo');
const { copy } = useClipboard({ legacy: true });
onMounted(refresh);
function openForm(task?: Task, status: TaskStatus = 'todo') {
  if (disabled.value || dragging.value) return;
  editing.value = task;
  initialStatus.value = status;
  open.value = true;
}
function reset() {
  keyword.value = '';
  priority.value = '';
  overdueOnly.value = false;
}
async function saveForm(data: TaskInput, id?: number) {
  const success = await save(data, id);
  if (success) ElMessage.success(id ? '任务已更新' : '任务已创建');
  return success;
}
async function moveCard(data: TaskMove) {
  if (await move(data)) ElMessage.success('任务位置已保存');
}
async function deleteCard(task: Task) {
  if (disabled.value || dragging.value) return;
  confirming.value = true;
  try {
    await ElMessageBox.confirm(
      `确定删除「${task.title}」吗？删除后无法恢复。`,
      '删除任务',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    );
    if (await remove(task.id)) ElMessage.success('任务已删除');
  } catch {
    /* 取消确认，无需提示。 */
  } finally {
    confirming.value = false;
  }
}
async function copyCard(task: Task) {
  try {
    await copy(
      `#${task.id} ${task.title}\n状态：${statuses.find((item) => item.value === task.status)?.label}\n负责人：${task.assignee || '待分配'}\n截止日期：${task.dueDate || '未设置'}\n${task.description}`,
    );
    ElMessage.success('任务摘要已复制');
  } catch {
    ElMessage.warning('复制失败，请检查浏览器剪贴板权限');
  }
}
onBeforeRouteLeave((to) => {
  if (to.path.startsWith('/auth/')) return true;
  if (pending.value || dragging.value || confirming.value) {
    ElMessage.warning('操作处理中，请稍候');
    return false;
  }
  return true;
});
</script>

<template>
  <div class="task-page p-4">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-4">
      <PageHeader
        title="任务看板"
        subtitle="把计划变成行动，让每一步进展清晰可见"
      />
      <div class="flex gap-2">
        <ElButton :disabled="disabled || dragging" @click="refresh"
          >刷新看板</ElButton
        ><ElButton
          type="primary"
          :disabled="disabled || dragging"
          @click="openForm()"
          >＋ 新建任务</ElButton
        >
      </div>
    </div>
    <div class="mb-5 grid grid-cols-2 gap-4 xl:grid-cols-4">
      <ElCard shadow="never"
        ><div class="text-muted-foreground text-xs">全部任务</div>
        <div class="mt-2 text-3xl font-semibold">
          {{ board.total
          }}<span class="text-muted-foreground ml-2 text-xs font-normal"
            >项</span
          >
        </div></ElCard
      >
      <ElCard shadow="never"
        ><div class="text-muted-foreground text-xs">正在推进</div>
        <div class="text-primary mt-2 text-3xl font-semibold">
          {{ doing
          }}<span class="text-muted-foreground ml-2 text-xs font-normal"
            >项</span
          >
        </div></ElCard
      >
      <ElCard shadow="never"
        ><div class="text-muted-foreground text-xs">逾期待处理</div>
        <div
          class="mt-2 text-3xl font-semibold"
          :class="overdue ? 'text-destructive' : ''"
        >
          {{ overdue
          }}<span class="text-muted-foreground ml-2 text-xs font-normal"
            >项</span
          >
        </div></ElCard
      >
      <ElCard shadow="never"
        ><div class="text-muted-foreground text-xs">已完成</div>
        <div class="mt-2 text-3xl font-semibold">
          {{ completed
          }}<span class="text-muted-foreground ml-2 text-xs font-normal"
            >项</span
          >
        </div></ElCard
      >
    </div>
    <ElCard shadow="never" class="mb-5"
      ><ElForm
        inline
        :disabled="pending || loading || dragging || confirming"
        class="task-filters"
        @submit.prevent
      >
        <ElFormItem label="搜索任务"
          ><ElInput
            v-model="keyword"
            placeholder="标题、负责人或 #编号"
            clearable
            maxlength="80"
            style="width: 240px"
        /></ElFormItem>
        <ElFormItem label="优先级"
          ><ElSelect
            v-model="priority"
            clearable
            :value-on-clear="''"
            placeholder="全部优先级"
            style="width: 150px"
            ><ElOption
              v-for="item in priorities"
              :key="item.value"
              :label="item.label"
              :value="item.value" /></ElSelect
        ></ElFormItem>
        <ElFormItem
          ><ElSwitch
            v-model="overdueOnly"
            active-text="仅看逾期"
            aria-label="仅看逾期"
        /></ElFormItem>
        <ElFormItem><ElButton @click="reset">重置</ElButton></ElFormItem>
      </ElForm></ElCard
    >
    <ElAlert v-if="failed" type="error" :closable="false" class="mb-4" show-icon
      ><template #title
        >未能获取最新看板，请重新加载后再操作
        <ElButton
          link
          type="primary"
          :disabled="loading || pending"
          @click="refresh"
          >重新加载</ElButton
        ></template
      ></ElAlert
    >
    <div
      class="text-muted-foreground mb-3 flex flex-wrap items-center justify-between gap-2 text-xs"
      aria-live="polite"
    >
      <span>{{
        filtering
          ? `筛选结果 ${visibleTasks.length} / ${board.items.length} 项 · 清空筛选后可拖拽排序`
          : '按住卡片右上角手柄拖动，也可点击编辑修改状态'
      }}</span
      ><span>{{
        pending
          ? '正在保存，请稍候…'
          : loading
            ? '正在加载…'
            : '共三个阶段 · 按列管理进度'
      }}</span>
    </div>
    <div v-loading="loading">
      <TaskBoard
        :items="visibleTasks"
        :disabled="disabled"
        :drag-disabled="filtering"
        @move="moveCard"
        @edit="openForm($event)"
        @add="openForm(undefined, $event)"
        @remove="deleteCard"
        @copy="copyCard"
        @dragging="dragging = $event"
      />
    </div>
    <TaskFormDrawer
      v-model="open"
      :task="editing"
      :initial-status="initialStatus"
      :save="saveForm"
    />
  </div>
</template>

<style scoped>
.task-filters :deep(.el-form-item) {
  margin-bottom: 0;
  margin-right: 20px;
}
.task-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 0;
}
.task-page :deep(.el-card) {
  border-radius: 10px;
}
</style>
