import { effectScope } from 'vue';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Task, TaskBoard } from '../api/tasks';
import { getTaskBoard, moveTask, saveTask } from '../api/tasks';
import { applyTaskMove, useTaskBoard } from './use-task-board';
vi.mock('../api/tasks', () => ({
  getTaskBoard: vi.fn(),
  moveTask: vi.fn(),
  saveTask: vi.fn(),
  removeTask: vi.fn(),
}));
const items: Task[] = [1, 2, 3].map((id) => ({
  id,
  title: `任务${id}`,
  status: id === 3 ? 'doing' : 'todo',
  priority: 'medium',
  assignee: '',
  dueDate: '',
  description: '',
  updatedAt: '',
}));
const initial: TaskBoard = { items, version: 1 };
const scopes: ReturnType<typeof effectScope>[] = [];
function setup() {
  const scope = effectScope();
  scopes.push(scope);
  return scope.run(useTaskBoard)!;
}
afterEach(() => {
  scopes.splice(0).forEach((scope) => scope.stop());
  vi.resetAllMocks();
});
describe('任务看板状态', () => {
  it('跨列移动、列内排序和空列追加不修改原始快照', () => {
    const moved = applyTaskMove(items, { id: 1, status: 'doing', beforeId: 3 });
    expect(moved.filter((t) => t.status === 'doing').map((t) => t.id)).toEqual([
      1, 3,
    ]);
    expect(items[0]?.status).toBe('todo');
    expect(
      applyTaskMove(items, { id: 2, status: 'todo', beforeId: 1 }).map(
        (t) => t.id,
      ),
    ).toEqual([2, 1, 3]);
    expect(
      applyTaskMove(items, { id: 1, status: 'done', beforeId: 0 }).at(-1)
        ?.status,
    ).toBe('done');
    expect(applyTaskMove(items, { id: 1, status: 'done', beforeId: 3 })).toBe(
      items,
    );
  });
  it('乐观移动只允许一次写入，成功采用服务端快照', async () => {
    vi.mocked(getTaskBoard).mockResolvedValue(initial);
    let finish!: (board: TaskBoard) => void;
    vi.mocked(moveTask).mockImplementation(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    const state = setup();
    await state.refresh();
    const pending = state.move({ id: 1, status: 'doing', beforeId: 3 });
    expect(state.board.value.items.find((t) => t.id === 1)?.status).toBe(
      'doing',
    );
    expect(await state.move({ id: 2, status: 'done', beforeId: 0 })).toBe(
      false,
    );
    expect(moveTask).toHaveBeenCalledTimes(1);
    const saved = {
      items: applyTaskMove(items, { id: 1, status: 'doing', beforeId: 3 }),
      version: 2,
    };
    finish(saved);
    expect(await pending).toBe(true);
    expect(state.board.value).toEqual(saved);
    expect(state.pending.value).toBe(false);
  });
  it('失败恢复原顺序；重读也失败时锁定写入，刷新成功后解锁', async () => {
    vi.mocked(getTaskBoard)
      .mockResolvedValueOnce(initial)
      .mockRejectedValueOnce(new Error('offline'));
    vi.mocked(moveTask).mockRejectedValue(new Error('offline'));
    const state = setup();
    await state.refresh();
    expect(await state.move({ id: 1, status: 'done', beforeId: 0 })).toBe(
      false,
    );
    expect(state.board.value).toEqual(initial);
    expect(state.failed.value).toBe(true);
    expect(await state.save(items[0]!, 1)).toBe(false);
    expect(saveTask).not.toHaveBeenCalled();
    vi.mocked(getTaskBoard).mockResolvedValue(initial);
    await state.refresh();
    expect(state.ready.value).toBe(true);
  });
  it('版本冲突后同步最新看板，避免带旧版本再次覆盖', async () => {
    vi.mocked(getTaskBoard)
      .mockResolvedValueOnce(initial)
      .mockResolvedValueOnce({ items, version: 8 });
    vi.mocked(moveTask).mockRejectedValue({ response: { status: 409 } });
    const state = setup();
    await state.refresh();
    await state.move({ id: 1, status: 'done', beforeId: 0 });
    expect(state.board.value.version).toBe(8);
    expect(state.ready.value).toBe(true);
  });
  it('旧查询不能覆盖新结果或结束新请求的loading', async () => {
    let first!: (board: TaskBoard) => void;
    let second!: (board: TaskBoard) => void;
    vi.mocked(getTaskBoard)
      .mockImplementationOnce(
        () =>
          new Promise((r) => {
            first = r;
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((r) => {
            second = r;
          }),
      );
    const state = setup();
    const old = state.refresh();
    const latest = state.refresh();
    first(initial);
    await old;
    expect(state.loading.value).toBe(true);
    second({ items, version: 2 });
    await latest;
    expect(state.board.value.version).toBe(2);
  });
  it('卸载后忽略未完成请求', async () => {
    let finish!: (board: TaskBoard) => void;
    vi.mocked(getTaskBoard).mockImplementationOnce(
      () =>
        new Promise((r) => {
          finish = r;
        }),
    );
    const state = setup();
    const pending = state.refresh();
    scopes.at(-1)!.stop();
    finish(initial);
    await pending;
    expect(state.board.value.version).toBe(0);
  });
});
