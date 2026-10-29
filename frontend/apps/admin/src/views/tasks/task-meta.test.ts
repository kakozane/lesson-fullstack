import { expect, it } from 'vitest';
import type { Task } from '#/api/tasks';
import { isOverdue } from './task-meta';
it('截止日期按天判断，今天和已完成不标记逾期', () => {
  const task = { status: 'todo', dueDate: '2026-10-27' } as Task;
  expect(isOverdue(task, '2026-10-28')).toBe(true);
  expect(isOverdue({ ...task, dueDate: '2026-10-28' }, '2026-10-28')).toBe(
    false,
  );
  expect(isOverdue({ ...task, dueDate: '' }, '2026-10-28')).toBe(false);
  expect(isOverdue({ ...task, status: 'done' }, '2026-10-28')).toBe(false);
});
