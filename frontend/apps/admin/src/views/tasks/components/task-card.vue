<script setup lang="ts">
import type { Task } from '#/api/tasks';
import { computed } from 'vue';
import { ElButton, ElTag, ElTooltip } from 'element-plus';
import { formatDate, formatDateTime } from '@vben/utils';
import { isOverdue, priorities } from '../task-meta';

const props = defineProps<{
  task: Task;
  disabled: boolean;
  dragDisabled: boolean;
}>();
defineEmits<{ edit: [Task]; remove: [Task]; copy: [Task] }>();
const priority = computed(
  () => priorities.find((item) => item.value === props.task.priority)!,
);
</script>

<template>
  <article class="task-card" :data-id="task.id" :data-priority="task.priority">
    <div class="flex items-center justify-between gap-2">
      <ElTag :type="priority.type" size="small" effect="light">{{
        priority.label
      }}</ElTag>
      <div class="flex items-center gap-2">
        <span class="text-muted-foreground text-xs">#{{ task.id }}</span>
        <ElTooltip
          :content="
            dragDisabled
              ? '清空筛选后可拖拽，也可在编辑中修改状态'
              : '拖动调整顺序或移入其他列'
          "
        >
          <span
            ><button
              type="button"
              class="drag-handle"
              :disabled="disabled || dragDisabled"
              :aria-label="`拖动任务：${task.title}`"
            >
              ⠿
            </button></span
          >
        </ElTooltip>
      </div>
    </div>
    <button
      class="task-title"
      type="button"
      :disabled="disabled"
      @click="$emit('edit', task)"
    >
      {{ task.title }}
    </button>
    <p class="task-description text-muted-foreground">
      {{ task.description || '暂无任务描述' }}
    </p>
    <div class="mb-4 flex flex-wrap items-center justify-between gap-2 text-xs">
      <span class="flex items-center gap-2"
        ><span class="task-avatar">{{ task.assignee.slice(0, 1) || '待' }}</span
        >{{ task.assignee || '待分配' }}</span
      >
      <span
        :class="isOverdue(task) ? 'text-destructive' : 'text-muted-foreground'"
        >{{
          task.dueDate
            ? formatDate(task.dueDate, 'MM-DD') + ' 截止'
            : '未设截止日期'
        }}<span v-if="isOverdue(task)"> · 已逾期</span></span
      >
    </div>
    <div class="task-actions">
      <ElTooltip :content="`更新于 ${formatDateTime(task.updatedAt)}`"
        ><ElButton
          link
          type="primary"
          :disabled="disabled"
          @click="$emit('edit', task)"
          >编辑</ElButton
        ></ElTooltip
      >
      <ElButton link :disabled="disabled" @click="$emit('copy', task)"
        >复制摘要</ElButton
      >
      <ElButton
        link
        type="danger"
        :disabled="disabled"
        @click="$emit('remove', task)"
        >删除</ElButton
      >
    </div>
  </article>
</template>

<style scoped>
.task-card {
  padding: 16px;
  border: 1px solid hsl(var(--border));
  border-left: 3px solid var(--el-color-info-light-5);
  border-radius: 10px;
  background: hsl(var(--card));
  box-shadow: 0 2px 5px rgb(0 0 0 / 3%);
}
.task-card[data-priority='high'] {
  border-left-color: var(--el-color-danger-light-3);
}
.task-card[data-priority='medium'] {
  border-left-color: var(--el-color-warning-light-3);
}
.task-title {
  display: block;
  margin: 12px 0 7px;
  width: 100%;
  text-align: left;
  font-size: 14px;
  line-height: 1.6;
  font-weight: 600;
  overflow-wrap: anywhere;
}
.task-title:hover {
  color: var(--el-color-primary);
}
.task-description {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 40px;
  margin-bottom: 14px;
  font-size: 12px;
  line-height: 20px;
  overflow-wrap: anywhere;
}
.task-avatar {
  display: inline-grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
.drag-handle {
  padding: 0 5px;
  font-size: 20px;
  color: hsl(var(--muted-foreground));
  cursor: grab;
  touch-action: none;
}
.drag-handle:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}
.task-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  border-top: 1px solid hsl(var(--border));
  padding-top: 10px;
}
.task-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}
</style>
