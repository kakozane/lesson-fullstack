import type { Task, TaskBoard, TaskInput, TaskMove } from '../api/tasks';
import { computed, onScopeDispose, ref } from 'vue';
import { getTaskBoard, moveTask, removeTask, saveTask } from '../api/tasks';

export function applyTaskMove(items: Task[], move: TaskMove): Task[] {
  const task = items.find((item) => item.id === move.id);
  if (!task) return items;
  const next = items.filter((item) => item.id !== move.id);
  const anchor = move.beforeId
    ? next.findIndex(
        (item) => item.id === move.beforeId && item.status === move.status,
      )
    : next.length;
  if (anchor < 0) return items;
  next.splice(anchor, 0, { ...task, status: move.status });
  return next;
}

export function useTaskBoard() {
  const board = ref<TaskBoard>({ items: [], version: 0 });
  const loading = ref(false);
  const pending = ref(false);
  const failed = ref(false);
  const ready = computed(() => board.value.version > 0 && !failed.value);
  let requestId = 0;
  let disposed = false;
  async function read() {
    const current = ++requestId;
    loading.value = true;
    try {
      const result = await getTaskBoard();
      if (current !== requestId || disposed) return;
      board.value = result;
      failed.value = false;
    } catch {
      if (current === requestId && !disposed) failed.value = true;
    } finally {
      if (current === requestId && !disposed) loading.value = false;
    }
  }
  async function refresh() {
    if (!pending.value) await read();
  }
  async function mutate(
    request: () => Promise<TaskBoard>,
    optimistic?: TaskMove,
  ) {
    if (pending.value || loading.value || !ready.value || disposed)
      return false;
    pending.value = true;
    const previous = board.value;
    if (optimistic)
      board.value = {
        ...previous,
        items: applyTaskMove(previous.items, optimistic),
      };
    try {
      const result = await request();
      if (!disposed) board.value = result;
      return true;
    } catch {
      if (!disposed) {
        board.value = previous;
        // 超时可能已经写入；恢复本地快照后重读服务端，失败则锁定写入直到刷新成功。
        await read();
      }
      return false;
    } finally {
      pending.value = false;
    }
  }
  function save(data: TaskInput, id?: number) {
    return mutate(() => saveTask(data, board.value.version, id));
  }
  function remove(id: number) {
    return mutate(() => removeTask(id, board.value.version));
  }
  function move(data: TaskMove) {
    return mutate(() => moveTask(data, board.value.version), data);
  }
  onScopeDispose(() => {
    disposed = true;
    requestId++;
  });
  return {
    board,
    loading,
    pending,
    failed,
    ready,
    refresh,
    save,
    remove,
    move,
  };
}
