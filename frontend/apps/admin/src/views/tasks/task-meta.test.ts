import { expect, it, vi } from 'vitest';
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

it('默认日期使用北京时间，UTC 午夜前也能识别跨日逾期', () => {
  vi.useFakeTimers();
  try {
    vi.setSystemTime(new Date('2026-10-28T16:00:00Z'));
    expect(isOverdue({ status: 'doing', dueDate: '2026-10-28' } as Task)).toBe(true);
    expect(isOverdue({ status: 'todo', dueDate: '2026-10-29' } as Task)).toBe(false);
  } finally {
    vi.useRealTimers();
  }
});
