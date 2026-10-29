<script setup lang="ts">
import type { Task, TaskMove, TaskStatus } from '#/api/tasks';
import type { Sortable } from '@vben/hooks';
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useSortable } from '@vben/hooks';
import { ElButton, ElMessage } from 'element-plus';
import { statuses } from '../task-meta';
import TaskCard from './task-card.vue';

const props = defineProps<{
  items: Task[];
  disabled: boolean;
  dragDisabled: boolean;
}>();
const emit = defineEmits<{
  move: [TaskMove];
  edit: [Task];
  remove: [Task];
  copy: [Task];
  add: [TaskStatus];
  dragging: [boolean];
}>();
const columns = ref<HTMLElement[]>([]);
const sortables: Sortable[] = [];
let disposed = false;
watch(
  () => props.disabled || props.dragDisabled,
  (disabled) => {
    sortables.forEach((sortable) => sortable.option('disabled', disabled));
  },
);
onMounted(async () => {
  await nextTick();
  try {
    for (const container of columns.value) {
      const { initializeSortable } = useSortable(container, {
        group: 'lesson-task-board',
        handle: '.drag-handle',
        draggable: '.task-card',
        animation: 180,
        ghostClass: 'task-ghost',
        chosenClass: 'task-chosen',
        disabled: props.disabled || props.dragDisabled,
        onStart: () => emit('dragging', true),
        onEnd(event) {
          const { item, from, to, oldDraggableIndex, newDraggableIndex } =
            event;
          const status = to.dataset.status as TaskStatus;
          const beforeId = Number(
            (item.nextElementSibling as HTMLElement | null)?.dataset.id || 0,
          );
          const id = Number(item.dataset.id);
          // Sortable 已移动真实 DOM；先还原，再让 Vue 根据响应式数据完成渲染。
          item.remove();
          from.insertBefore(
            item,
            from.querySelectorAll(':scope > .task-card')[
              oldDraggableIndex ?? 0
            ] ?? null,
          );
          emit('dragging', false);
          if (from === to && oldDraggableIndex === newDraggableIndex) return;
          emit('move', { id, status, beforeId });
        },
      });
      const sortable = await initializeSortable();
      if (disposed) {
        sortable.destroy();
        return;
      }
      sortable.option('disabled', props.disabled || props.dragDisabled);
      sortables.push(sortable);
    }
  } catch {
    ElMessage.error('拖拽初始化失败，可通过编辑修改任务状态');
  }
});
onBeforeUnmount(() => {
  disposed = true;
  sortables.forEach((sortable) => sortable.destroy());
});
</script>

<template>
  <div class="task-board-grid">
    <section
      v-for="(column, index) in statuses"
      :key="column.value"
      class="board-column"
      :aria-label="`${column.label}任务列`"
    >
      <header class="column-heading">
        <div>
          <h2 class="flex items-center gap-2 text-sm font-semibold">
            <span class="column-dot" :data-status="column.value"></span
            >{{ column.label
            }}<span class="column-count">{{
              items.filter((task) => task.status === column.value).length
            }}</span>
          </h2>
          <p class="text-muted-foreground mt-1 text-xs">{{ column.hint }}</p>
        </div>
        <ElButton
          text
          :disabled="disabled"
          :aria-label="`在${column.label}新建任务`"
          @click="emit('add', column.value)"
          >＋</ElButton
        >
      </header>
      <div
        :ref="
          (el) => {
            if (el) columns[index] = el as HTMLElement;
          }
        "
        class="column-list"
        :data-status="column.value"
      >
        <TaskCard
          v-for="task in items.filter((task) => task.status === column.value)"
          :key="task.id"
          :task="task"
          :disabled="disabled"
          :drag-disabled="dragDisabled"
          @edit="emit('edit', $event)"
          @remove="emit('remove', $event)"
          @copy="emit('copy', $event)"
        />
      </div>
      <p
        v-if="!items.some((task) => task.status === column.value)"
        class="empty-column text-muted-foreground"
      >
        暂无任务，可点击 ＋ 新建
      </p>
    </section>
  </div>
</template>

<style scoped>
.task-board-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: start;
  gap: 18px;
}
.board-column {
  position: relative;
  min-width: 0;
  padding: 14px;
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
  background: hsl(var(--muted) / 0.45);
}
.column-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 14px;
}
.column-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--el-color-info);
}
.column-dot[data-status='doing'] {
  background: var(--el-color-primary);
}
.column-dot[data-status='done'] {
  background: var(--el-color-success);
}
.column-count {
  padding: 1px 7px;
  border-radius: 5px;
  background: hsl(var(--background));
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}
.column-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 150px;
}
.empty-column {
  position: absolute;
  top: 120px;
  left: 0;
  right: 0;
  text-align: center;
  pointer-events: none;
  font-size: 12px;
}
:deep(.task-ghost) {
  opacity: 0.3;
  border: 1px dashed var(--el-color-primary);
}
:deep(.task-chosen) {
  box-shadow: 0 8px 24px rgb(0 0 0 / 10%);
}
@media (max-width: 1000px) {
  .task-board-grid {
    grid-template-columns: 1fr;
  }
}
</style>
