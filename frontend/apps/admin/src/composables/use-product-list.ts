import type { Product, ProductQuery } from '../api/products';

import { onScopeDispose, reactive, ref } from 'vue';

import { getProducts } from '../api/products';

export function useProductList() {
  const filters = reactive<Pick<ProductQuery, 'category' | 'keyword' | 'status'>>({
    keyword: '',
    category: '',
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
      const result = await getProducts({ ...query });
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
    Object.assign(query, filters, { keyword: filters.keyword.trim(), page: 1 });
    return load();
  }
  function reset() {
    Object.assign(filters, { keyword: '', category: '', status: '' });
    return search();
  }
  function changePage(page: number) {
    query.page = page;
    return load();
  }
  function changeSize(pageSize: number) {
    Object.assign(query, { pageSize, page: 1 });
    return load();
  }
  onScopeDispose(() => {
    requestId++;
  });
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
