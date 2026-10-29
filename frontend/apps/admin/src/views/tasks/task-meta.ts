import type { Task, TaskPriority, TaskStatus } from '#/api/tasks';

export const statuses: { value: TaskStatus; label: string; hint: string }[] = [
  { value: 'todo', label: '待处理', hint: '整理想法，明确下一步' },
  { value: 'doing', label: '进行中', hint: '专注推进，及时协作' },
  { value: 'done', label: '已完成', hint: '记录成果，让进度可见' },
];
export const priorities: {
  value: TaskPriority;
  label: string;
  type: 'info' | 'warning' | 'danger';
}[] = [
  { value: 'high', label: '高优先级', type: 'danger' },
  { value: 'medium', label: '中优先级', type: 'warning' },
  { value: 'low', label: '低优先级', type: 'info' },
];
export function isOverdue(task: Task, today = new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10)) {
  return task.status !== 'done' && !!task.dueDate && task.dueDate < today;
}
