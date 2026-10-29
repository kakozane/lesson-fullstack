import { requestClient } from '#/api/request';

export interface Category {
 id: number;
 name: string;
}

export interface ProductInput {
  name: string;
  sku: string;
  categoryId: number;
  priceCents: number;
  stock: number;
  status: 0 | 1;
  description: string;
}
export interface Product extends ProductInput {
  id: number;
  categoryName: string;
  updatedAt: string;
}
export interface ProductQuery {
  keyword: string;
  categoryId?: number;
  status: '' | 0 | 1;
  page: number;
  pageSize: number;
}
export interface ProductPage {
  items: Product[];
  total: number;
  page: number;
  pageSize: number;
}
export const getProducts = (params: ProductQuery) =>
  requestClient.get<ProductPage>('/products', { params });
export const getCategories = () => requestClient.get<Category[]>('/categories');
export const createProduct = (data: ProductInput) =>
  requestClient.post('/products', data);
export const updateProduct = (id: number, data: ProductInput) =>
  requestClient.put(`/products/${id}`, data);
export const deleteProduct = (id: number) =>
  requestClient.delete(`/products/${id}`);
export const batchStatus = (ids: number[], status: 0 | 1) =>
  requestClient.request('/products/status', {
    method: 'PATCH',
    data: { ids, status },
  });
