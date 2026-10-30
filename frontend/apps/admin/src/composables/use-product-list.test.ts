import { effectScope } from 'vue';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { getProducts } from '../api/products';
import { useProductList } from './use-product-list';
vi.mock('../api/products', () => ({ getProducts: vi.fn() }));
const scopes: ReturnType<typeof effectScope>[] = [];
function setup() {
  const scope = effectScope();
  scopes.push(scope);
  return scope.run(useProductList)!;
}
afterEach(() => {
  scopes.splice(0).forEach((scope) => scope.stop());
  vi.resetAllMocks();
});
const emptyPage = { items: [], total: 0, page: 1, pageSize: 10 };
describe('商品列表状态', () => {
  it('默认、清空和重置时省略分类参数，选中时传数字 ID', async () => {
    vi.mocked(getProducts).mockResolvedValue(emptyPage);
    const state = setup();
    await state.load();
    expect(vi.mocked(getProducts).mock.lastCall?.[0]).not.toHaveProperty('categoryId');
    state.filters.categoryId = 2;
    await state.search();
    expect(getProducts).toHaveBeenLastCalledWith(expect.objectContaining({ categoryId: 2 }));
    state.filters.categoryId = undefined;
    await state.search();
    expect(vi.mocked(getProducts).mock.lastCall?.[0]).not.toHaveProperty('categoryId');
    state.filters.categoryId = 3;
    await state.search();
    await state.reset();
    expect(vi.mocked(getProducts).mock.lastCall?.[0]).not.toHaveProperty('categoryId');
  });

  it('筛选保留下架值 0 并重置页码；重置清空全部条件', async () => {
    vi.mocked(getProducts).mockResolvedValue(emptyPage);
    const state = setup();
    state.query.page = 3;
    state.filters.status = 0;
    state.filters.keyword = ' 鼠标 ';
    await state.search();
    expect(getProducts).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 1, status: 0, keyword: '鼠标' }),
    );
    await state.reset();
    expect(state.query).toEqual({
      keyword: '',
      categoryId: undefined,
      status: '',
      page: 1,
      pageSize: 10,
    });
  });
  it('旧请求晚到不能覆盖新结果或提前结束 loading', async () => {
    let first!: (value: typeof emptyPage) => void;
    let second!: (value: typeof emptyPage) => void;
    vi.mocked(getProducts)
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            first = resolve;
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            second = resolve;
          }),
      );
    const state = setup();
    const old = state.load();
    const latest = state.search();
    first({ ...emptyPage, total: 99 });
    await old;
    expect(state.loading.value).toBe(true);
    expect(state.total.value).toBe(0);
    second({ ...emptyPage, total: 2 });
    await latest;
    expect(state.total.value).toBe(2);
    expect(state.loading.value).toBe(false);
  });
  it('最新请求完成后，旧响应仍不能覆盖；失败可重试', async () => {
    let first!: (value: typeof emptyPage) => void;
    vi.mocked(getProducts)
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            first = resolve;
          }),
      )
      .mockResolvedValueOnce({ ...emptyPage, total: 2 });
    const state = setup();
    const old = state.load();
    await state.search();
    first({ ...emptyPage, total: 99 });
    await old;
    expect(state.total.value).toBe(2);
    vi.mocked(getProducts).mockRejectedValueOnce(new Error('offline'));
    await state.load();
    expect(state.failed.value).toBe(true);
    expect(state.loading.value).toBe(false);
    vi.mocked(getProducts).mockResolvedValueOnce(emptyPage);
    await state.load();
    expect(state.failed.value).toBe(false);
  });
  it('每页数量变更重置页码并更新结果', async () => {
    vi.mocked(getProducts).mockResolvedValue({ ...emptyPage, pageSize: 20 });
    const state = setup();
    state.query.page = 3;
    await state.changeSize(20);
    expect(getProducts).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 1, pageSize: 20 }),
    );
    expect(state.query.page).toBe(1);
    expect(state.query.pageSize).toBe(20);
    expect(state.loading.value).toBe(false);
  });
});
