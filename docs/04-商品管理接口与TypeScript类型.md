# 商品管理接口与 TypeScript 类型

商品管理包含分类查询、列表筛选分页、新增、编辑、删除和批量上下架。本文以当前 Gin 路由、DTO、业务校验及 Vben 请求封装为准，说明请求参数、完整响应和前端类型写法。

示例中的 ID、时间和商品内容用于解释结构，不代表线上当前数据。在线演示站使用共享数据，新增和删除练习优先在本地进行。

## 1. 请求约定与接口总览

| 项目 | 约定 |
|---|---|
| 本地 API 前缀 | `http://127.0.0.1:8080/api` |
| 线上 API 前缀 | `https://test.retniw.cc/api` |
| 前端请求前缀 | `/api`，开发由 Vite 代理，线上由 Nginx 代理 |
| 登录凭证 | `Authorization: Bearer <accessToken>` |
| JSON 请求体 | `Content-Type: application/json` |
| 成功 | HTTP `200`，JSON 中 `code` 为 `0` |
| 失败 | HTTP `400 / 401 / 404 / 409 / 500` 等，`code` 与状态码一致 |

登录接口是 `POST /api/auth/login`，请求体为 `{"username":"admin","password":"123456"}`；从响应的 `data.accessToken` 获取令牌。这里的密码是课程演示配置，本地以 `ADMIN_PASSWORD` 为准。商品模块所有接口都需要登录。会话有效期为 8 小时，后端重启后失效。

| 接口名称 | 方法 | 完整路径 | 前端函数 | 成功的 data 类型 |
|---|---|---|---|---|
| 查询商品分类 | GET | `/api/categories` | `getCategories()` | `string[]` |
| 查询商品列表 | GET | `/api/products` | `getProducts(params)` | `ProductPage` |
| 新增商品 | POST | `/api/products` | `createProduct(data)` | `null` |
| 编辑商品 | PUT | `/api/products/:id` | `updateProduct(id, data)` | `null` |
| 删除商品 | DELETE | `/api/products/:id` | `deleteProduct(id)` | `null` |
| 批量上下架 | PATCH | `/api/products/status` | `batchStatus(ids, status)` | `null` |

`:id` 表示 URL 中的商品 ID，例如 `/api/products/42`。当前没有单独的商品详情接口，编辑表单直接使用列表行中的完整商品数据；单条上下架复用批量接口。

### 原始响应与 requestClient 的区别

服务器始终返回 `{ code, data, message }`。描述原始 HTTP 响应时可以使用：

```ts
export interface ApiResponse<T> {
  code: number;
  data: T | null;
  message: string;
}
```

项目的 `requestClient` 已配置成功码为 `0`，成功时解包 `data`，失败时抛出异常。因此 `await getProducts(params)` 得到的是 `ProductPage`，直接访问 `result.items`，不能再写 `result.data.items`。新增、编辑、删除、批量上下架成功得到 `null`，不是商品对象，也不是 HTTP 204。

JSON 写入接口只接受一个 JSON 值，并拒绝未知字段；请求体上限为 64 KiB。不要把包含 `id`、`updatedAt` 的整个列表行直接作为新增或编辑请求体，应只提取 `ProductInput` 字段。JSON 数值字段使用数值，不传数字字符串或小数形式的整数参数。

## 2. 商品字段与输入规则

| 字段 | TypeScript 类型 | 请求要求 | 含义与约束 |
|---|---|---|---|
| `id` | `number` | 仅响应；编辑、删除时放 URL | 服务端生成的正整数 |
| `name` | `string` | 必填 | 去首尾空白后 1～80 个字符 |
| `sku` | `string` | 必填 | 去首尾空白后 1～32 位，只允许 ASCII 字母、数字和 `-`；必须唯一 |
| `category` | `string` | 必填 | 必须等于分类接口返回的选项之一 |
| `priceCents` | `number` | 必填 | 整数分，范围 1～99999999，即 0.01～999999.99 元 |
| `stock` | `number` | 前端完整提交 | 整数，范围 0～999999；后端省略此字段时按 0 处理 |
| `status` | `0 \| 1` | 必填，不能为 null | `0` 下架，`1` 上架；省略会校验失败 |
| `description` | `string` | 可省略；前端统一传字符串 | 去首尾空白后最多 500 个字符，空值传 `""`，省略也按空字符串处理 |
| `updatedAt` | `string` | 仅响应 | `YYYY-MM-DD HH:mm:ss`，当前项目按 UTC+8 输出 |

`priceCents: 2990` 表示 29.90 元，不是 2990 元。前端显示可以使用 `(priceCents / 100).toFixed(2)`。TypeScript 的 `number` 不会自动限制为整数或检查上下限，表单仍需进行运行时校验。

名称、编码和描述会去除首尾空白；分类没有自动去空白，应直接使用分类选项。SKU 的唯一性还受数据库排序规则影响；项目默认排序规则不区分大小写，避免使用仅大小写不同的商品编码。

## 3. 查询商品分类

**接口名称：查询商品分类**  
**方法与路径：`GET /api/categories`**  
**参数：无 Query 参数、无请求体。**

成功响应：

```json
{
  "code": 0,
  "data": ["数码配件", "办公文具", "生活好物"],
  "message": ""
}
```

返回值是字符串数组，每个值既用于下拉展示，也用于新增、编辑和分类筛选。当前分类来自服务端固定选项，没有分类增删改接口。

前端函数返回 `Promise<string[]>`。常见失败为 `401`，表示登录凭证无效。

## 4. 查询商品列表

**接口名称：筛选并分页查询商品列表**  
**方法与路径：`GET /api/products`**  
**参数位置：URL Query，无请求体。**

| 参数 | 前端类型 | 是否必传 | 默认值 | 规则 |
|---|---|---|---|---|
| `keyword` | `string` | 否 | `""` | 名称或 SKU 包含查询，去首尾空白后最多 80 字；`%`、`_` 按普通字符匹配 |
| `category` | `string` | 否 | `""` | 精确匹配，空字符串表示全部分类 |
| `status` | `'' \| 0 \| 1` | 否 | `""` | 空字符串表示全部；0 下架，1 上架 |
| `page` | `number` | 否 | `1` | 整数，1～1000000 |
| `pageSize` | `number` | 否 | `10` | 整数，1～100；页面常用 10、20、50 |

前端传 `status: 0` 后，URL 中实际是 `status=0`。不要用 `if (status)` 判断是否带筛选条件，否则下架状态 `0` 会被漏掉。

请求示例：

```http
GET /api/products?keyword=鼠标&category=数码配件&status=1&page=1&pageSize=10
Authorization: Bearer <accessToken>
```

成功响应示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 42,
        "name": "无线鼠标",
        "sku": "CLASS-001",
        "category": "数码配件",
        "priceCents": 2990,
        "stock": 25,
        "status": 1,
        "description": "课堂演示商品",
        "updatedAt": "2026-10-29 10:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 10
  },
  "message": ""
}
```

| 返回字段 | 类型 | 含义 |
|---|---|---|
| `items` | `Product[]` | 当前页完整商品数据，无结果时为 `[]` |
| `total` | `number` | 满足筛选条件的总条数，不是当前页条数 |
| `page` | `number` | 服务端实际返回的页码 |
| `pageSize` | `number` | 每页条数 |

筛选条件之间使用 AND，商品默认按 `id` 倒序。超过最后一页的合法页码会被调整到最后一个有效页；无结果时 `items: []`、`total: 0`、`page: 1`。前端应使用响应中的 `page` 同步分页组件。

列表查询中的未知分类是普通精确匹配条件，通常得到空列表；新增和编辑中的未知分类则返回 400。

常见失败：分页范围或搜索词长度不合法返回 `400`；`status=2` 返回 `400`；无法解析分页 Query 返回 `400`，消息为 `查询参数格式错误`。

## 5. 新增商品

**接口名称：新增商品**  
**方法与路径：`POST /api/products`**  
**参数位置：JSON 请求体，类型为 `ProductInput`。**

```json
{
  "name": "无线鼠标",
  "sku": "CLASS-001",
  "category": "数码配件",
  "priceCents": 2990,
  "stock": 25,
  "status": 1,
  "description": "课堂演示商品"
}
```

字段要求见第 2 节，不传 `id` 和 `updatedAt`。成功响应：

```json
{
  "code": 0,
  "data": null,
  "message": "保存成功"
}
```

此接口不返回新商品 ID。成功后重新调用列表接口；需要展示新记录时可以查询其唯一 SKU。字段不合法返回 `400`；SKU 已存在返回 `409`，消息为 `商品编码已存在`。

## 6. 编辑商品

**接口名称：完整更新商品业务字段**  
**方法与路径：`PUT /api/products/:id`**

| 参数位置 | 参数 | 类型 | 要求 |
|---|---|---|---|
| Path | `id` | `number` | 正整数，例如 42 |
| Body | `ProductInput` 中的全部业务字段 | JSON 对象 | 与新增商品相同，不传 `id`、`updatedAt` |

例如 `PUT /api/products/42`：

```json
{
  "name": "无线鼠标 升级款",
  "sku": "CLASS-001",
  "category": "数码配件",
  "priceCents": 3990,
  "stock": 0,
  "status": 0,
  "description": ""
}
```

这表示把库存改为 0、设为下架并清空描述，合法零值不会被忽略。这里是完整更新，不是局部 PATCH；省略 `stock` 或 `description` 会按零值写入，不会保留旧值。

成功响应与新增相同：

```json
{
  "code": 0,
  "data": null,
  "message": "保存成功"
}
```

成功后重新查询列表。内容没有变化也可保存成功。非法 ID 或字段返回 `400`，商品不存在返回 `404`，SKU 与其他商品重复返回 `409`。

商品编辑没有版本控制，多人编辑同一商品时，后提交的内容可能覆盖前一次修改。

## 7. 删除商品

**接口名称：删除指定商品**  
**方法与路径：`DELETE /api/products/:id`**  
**参数：Path 中的正整数 `id`，无 Query 参数、无请求体。**

成功响应：

```json
{
  "code": 0,
  "data": null,
  "message": "删除成功"
}
```

删除是永久删除。商品不存在或重复删除返回 `404`，消息为 `商品不存在或已删除`。前端通常先二次确认，成功后重查列表，采用服务端返回的实际页码。

## 8. 单条与批量上下架

**接口名称：批量更新商品上下架状态**  
**方法与路径：`PATCH /api/products/status`**  
**参数位置：JSON 请求体。**

| 参数 | 类型 | 是否必填 | 约束 |
|---|---|---|---|
| `ids` | `number[]` | 是 | 1～100 个不重复的正整数 |
| `status` | `0 \| 1` | 是 | 0 下架，1 上架；不能省略或传 null |

```json
{
  "ids": [35, 36],
  "status": 0
}
```

单条上下架同样调用此接口，例如 `{"ids":[42],"status":1}`。成功响应：

```json
{
  "code": 0,
  "data": null,
  "message": "批量更新成功"
}
```

服务端在事务内检查并锁定全部目标商品后再更新。任意 ID 已不存在时返回 `409`，整批失败，不会部分更新。重复提交相同状态仍然成功。空数组、重复 ID、超出 100 个 ID 或非法状态返回 `400`。

## 9. 完整 TypeScript 类型与请求封装

对应文件：`frontend/apps/admin/src/api/products.ts`。下面保持现有函数名和调用方式，并补全写入接口的 `<null>` 返回类型以及批量请求类型。当前源码的部分写入函数没有显式泛型；补上泛型只改善类型约束，不改变后端响应。

`ProductQuery` 的所有属性在前端类型中必填，是为了表示完整筛选状态；这不代表 HTTP 接口要求每个 Query 参数都必须提交。

```ts
import { requestClient } from '#/api/request';

export type ProductStatus = 0 | 1;

export interface ProductInput {
  name: string;
  sku: string;
  category: string;
  /** 整数分，例如 2990 表示 29.90 元。 */
  priceCents: number;
  stock: number;
  status: ProductStatus;
  description: string;
}

export interface Product extends ProductInput {
  id: number;
  /** UTC+8，YYYY-MM-DD HH:mm:ss。 */
  updatedAt: string;
}

export interface ProductQuery {
  keyword: string;
  category: string;
  status: '' | ProductStatus;
  page: number;
  pageSize: number;
}

export interface ProductPage {
  items: Product[];
  total: number;
  page: number;
  pageSize: number;
}

export interface BatchStatusInput {
  ids: number[];
  status: ProductStatus;
}

export const getCategories = (): Promise<string[]> =>
  requestClient.get<string[]>('/categories');

export const getProducts = (params: ProductQuery): Promise<ProductPage> =>
  requestClient.get<ProductPage>('/products', { params });

export const createProduct = (data: ProductInput): Promise<null> =>
  requestClient.post<null>('/products', data);

export const updateProduct = (
  id: number,
  data: ProductInput,
): Promise<null> => requestClient.put<null>(`/products/${id}`, data);

export const deleteProduct = (id: number): Promise<null> =>
  requestClient.delete<null>(`/products/${id}`);

export const batchStatus = (
  ids: number[],
  status: ProductStatus,
): Promise<null> => {
  const data: BatchStatusInput = { ids, status };
  return requestClient.request<null>('/products/status', {
    method: 'PATCH',
    data,
  });
};
```

这里的泛型描述解包后的 `data`，不用再包 `ApiResponse<ProductPage>`。URL 只写 `/products`，不要重复拼接已在 `requestClient` 中配置的 `/api`。当前 Vben RequestClient 没有 `patch()` 快捷方法，PATCH 使用 `request()`。

页面使用顺序：查询分类 → 初始化完整筛选条件 → 查询列表 → 用户提交写入 → 等待成功 → 重查列表。查询请求放在 `{ params }`，POST/PUT 的 JSON 放第二个参数，通用 `request()` 的 JSON 放 `{ data }`。

## 10. 错误响应与源码位置

错误响应示例：

```json
{
  "code": 409,
  "data": null,
  "message": "商品编码已存在"
}
```

| HTTP / code | 常见原因 | 典型 message |
|---|---|---|
| 400 | 名称、SKU、价格、库存、状态等不合法 | `商品状态必须为 0 或 1` |
| 400 | Path ID 不是正整数 | `商品 ID 不合法` |
| 400 | JSON 字段错误、类型错误、未知字段或请求体超限 | `请求参数格式错误` |
| 401 | 未登录、令牌失效或后端重启 | `登录已过期，请重新登录` |
| 404 | 编辑或删除的商品不存在 | `商品不存在或已删除` |
| 409 | SKU 重复 | `商品编码已存在` |
| 409 | 批量操作中存在已删除商品 | `部分商品已删除，请刷新后重试` |
| 500 | 数据库或服务异常 | `服务暂时不可用，请稍后重试` |

界面可展示 `message`，业务分支优先依据 HTTP 状态码。请求层会抛出失败，页面不能在失败后显示“保存成功”或清空用户输入。

| 实现文件 | 对应内容 |
|---|---|
| `backend/router/admin/router.go` | 六个商品模块路由 |
| `backend/internal/product/admin/dto/product.go` | 请求和响应字段 |
| `backend/internal/product/admin/handler/product.go` | Query/JSON 绑定与成功响应 |
| `backend/internal/product/admin/service/product.go` | 字段范围、分类和错误映射 |
| `backend/internal/product/repository/product_repository.go` | 筛选分页、零值更新、事务批量状态 |
| `frontend/apps/admin/src/api/products.ts` | 页面使用的类型与请求函数 |
| `frontend/apps/admin/src/api/request.ts` | API 前缀、鉴权头、响应解包与错误提示 |

任务模块的版本与拖拽协议见 [任务看板接口与 TypeScript 类型](02-任务看板接口与TypeScript类型.md)。
