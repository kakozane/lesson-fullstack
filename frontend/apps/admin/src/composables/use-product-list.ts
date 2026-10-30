import type { Product, ProductQuery } from '../api/products';

import { reactive, ref } from 'vue';

import { getProducts } from '../api/products';

export function useProductList() {
  const filters = reactive<Pick<ProductQuery, 'categoryId' | 'keyword' | 'status'>>({
    keyword: '',
    categoryId: undefined,
    status: '',
  });
  const query = reactive<ProductQuery>({ ...filters, page: 1, pageSize: 10 });
  const items = ref<Product[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const failed = ref(false);
  let requestId = 0;

  async function load() {
    const current = ++requestId;
    loading.value = true;
    failed.value = false;
    items.value = [];
    try {
      const { categoryId, ...params } = query;
      const result = await getProducts({
        ...params,
        ...(categoryId === undefined ? {} : { categoryId }),
      });
      // 只让最新发出的请求更新页面，避免慢响应覆盖新筛选结果。
      if (current !== requestId) return;
      items.value = result.items;
      total.value = result.total;
      query.page = result.page;
    } catch {
      if (current === requestId) {
        failed.value = true;
        total.value = 0;
      }
      // 请求层统一展示错误，页面保留重试入口。
    } finally {
      if (current === requestId) loading.value = false;
    }
  }
  function search() {
    query.keyword = filters.keyword.trim();
    query.categoryId = filters.categoryId;
    query.status = filters.status;
    query.page = 1;
    return load();
  }
  function reset() {
    filters.keyword = '';
    filters.categoryId = undefined;
    filters.status = '';
    return search();
  }
  function changePage(page: number) {
    query.page = page;
    return load();
  }
  function changeSize(pageSize: number) {
    query.pageSize = pageSize;
    query.page = 1;
    return load();
  }
  return {
    filters,
    query,
    items,
    total,
    loading,
    failed,
    load,
    search,
    reset,
    changePage,
    changeSize,
  };
}
