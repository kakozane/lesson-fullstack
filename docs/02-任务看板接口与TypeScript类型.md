# 任务看板接口与 TypeScript 类型

任务看板包含全量查询、新增、编辑、删除和拖拽移动五个接口。所有写入都携带看板版本 `version`，成功后返回完整看板快照，前端用它更新任务列表和版本。

本文对照当前 Gin 路由、DTO、业务层、排序事务和 Vben 请求封装。示例是一个从 1 条任务开始的小型看板，用于说明字段、排序和版本变化；不代表线上当前数据。在线站点共用演示数据，写入练习优先在本地进行。

## 1. 请求约定与接口总览

| 项目 | 约定 |
|---|---|
| 本地 API 前缀 | `http://127.0.0.1:8080/api` |
| 线上 API 前缀 | `https://test.retniw.cc/api` |
| 前端请求前缀 | `/api`，开发由 Vite 代理，线上由 Nginx 代理 |
| 登录凭证 | `Authorization: Bearer <accessToken>` |
| 请求体 | JSON；POST、PUT、PATCH 和本模块 DELETE 均使用 JSON 请求体 |
| 成功 | HTTP `200`，`code: 0`，`message: ""` |
| 写入返回 | 完整 `TaskBoard`，不是单条任务或 null |

通过 `POST /api/auth/login` 登录，请求 `{"username":"admin","password":"123456"}`，从响应的 `data.accessToken` 获取令牌。演示密码可由 `ADMIN_PASSWORD` 修改。所有任务接口都需要登录，会话有效期 8 小时，后端重启后需要重新登录。

| 接口名称 | 方法 | 完整路径 | 当前前端函数 | 请求类型 | 成功的 data 类型 |
|---|---|---|---|---|---|
| 获取完整看板 | GET | `/api/tasks` | `getTaskBoard()` | 无 | `TaskBoard` |
| 新建任务 | POST | `/api/tasks` | `saveTask(data, version)` | `TaskSaveInput` | `TaskBoard` |
| 编辑任务 | PUT | `/api/tasks/:id` | `saveTask(data, version, id)` | `TaskSaveInput` | `TaskBoard` |
| 删除任务 | DELETE | `/api/tasks/:id` | `removeTask(id, version)` | `TaskDeleteInput` | `TaskBoard` |
| 移动与排序任务 | PATCH | `/api/tasks/:id/move` | `moveTask(move, version)` | `TaskMoveInput` | `TaskBoard` |

当前是单看板实现，没有看板 ID、独立任务详情、统计、分页或服务端筛选接口。最多 200 条任务，前端拿到全量数据后计算统计、搜索、优先级筛选和逾期状态。

### 原始响应与前端收到的值

原始 HTTP 响应可以表示为：

```ts
export interface ApiResponse<T> {
  code: number;
  data: T | null;
  message: string;
}
```

`requestClient` 成功时已经解包 `data`，因此 `await getTaskBoard()` 得到 `TaskBoard`，直接访问 `board.items` 和 `board.version`，不是 `board.data.items`。错误 HTTP 状态会抛出异常。

JSON 写入拒绝未知字段和拼接的多个 JSON 值，请求体上限 64 KiB。`id` 放在 URL，`updatedAt` 和内部 `position` 不放进请求体。版本、ID 和锚点使用整数数值，不用数字字符串；前端 `number` 类型本身不检查整数与范围。

## 2. 字段、状态与版本

### 任务输入和返回字段

| 字段 | TypeScript 类型 | 请求要求 | 含义与约束 |
|---|---|---|---|
| `id` | `number` | 仅响应；操作时放 URL | 服务端生成的正整数 |
| `title` | `string` | 必填 | 去首尾空白后 1～80 个字符 |
| `description` | `string` | 可省略；前端统一传字符串 | 去首尾空白后最多 2000 字，空值为 `""` |
| `assignee` | `string` | 可省略；前端统一传字符串 | 去首尾空白后最多 30 字，`""` 表示待分配；是姓名文本，不是用户 ID |
| `priority` | `'low' \| 'medium' \| 'high'` | 必填 | 低、中、高优先级 |
| `status` | `'todo' \| 'doing' \| 'done'` | 必填 | 待处理、进行中、已完成 |
| `dueDate` | `string` | 可省略；前端统一传字符串 | `""` 表示不设置，否则必须为真实日历日期 `YYYY-MM-DD` |
| `updatedAt` | `string` | 仅响应 | `YYYY-MM-DD HH:mm:ss`，当前数据库连接使用 Asia/Shanghai 时区 |

新建和编辑共用 `TaskInput`。描述、负责人、截止日期省略时，后端按空字符串处理；编辑时也会清空，不会保留旧值。清空字段推荐显式提交 `""`，不要把 `null` 作为前端表单空值约定。

`dueDate` 不带时分秒，不传 `2026-10-30T00:00:00Z`。非法日期、额外空白或不符合格式的值会被拒绝。可以设置过去的日期，是否逾期由前端结合当前日期与任务状态判断。

### 看板返回字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `items` | `Task[]` | 全部任务，空看板为 `[]`；没有分页 total 字段 |
| `version` | `number` | 全看板共用的正整数版本，不是某一条任务的版本 |

服务器按状态和列内位置组织返回结果；每一列中的顺序由内部 `position`、`id` 决定。前端应固定按 `todo → doing → done` 展示三列，并保留 `items` 在各列中的相对顺序，不要重新按 ID 排序。内部 `position` 不返回给前端，也不接受前端直接修改。

### version 的使用规则

1. 先 GET 获取当前看板和 `version`。
2. 每次新增、编辑、删除或移动，把当前版本一起提交。
3. 服务端在事务内锁定看板版本行，检查请求版本，再执行写入。
4. 成功后版本加 1，并返回完整看板；前端整体替换本地状态。
5. 版本过期返回 `409`，本次写入不生效，前端刷新后让用户确认是否重试。

不要预先把本地版本加 1，也不要把版本写死为 1。即使编辑内容不变，或一次合法移动没有改变最终顺序，当前实现仍会递增版本。并发操作、其他浏览器页签和其他访问者都可能改变同一个版本。

## 3. 获取完整看板

**接口名称：获取全部任务和看板版本**  
**方法与路径：`GET /api/tasks`**  
**参数：无 Query 参数、无请求体。**

成功响应示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "title": "梳理需求",
        "description": "明确交付标准",
        "assignee": "林悦",
        "priority": "high",
        "status": "todo",
        "dueDate": "2026-10-30",
        "updatedAt": "2026-10-29 10:00:00"
      }
    ],
    "version": 7
  },
  "message": ""
}
```

`items` 和 `version` 来自同一事务快照。前端搜索与筛选在这份数据上完成，传 `keyword`、`page` 等 Query 参数不会实现服务端筛选或分页。

接口不会长连接推送其他用户的变化，需要主动刷新，或者在写入冲突后重新查询。首次读取失败时，不应使用初始化的 `version: 0` 发起写入。

## 4. 新建任务

**接口名称：新增任务并返回完整看板**  
**方法与路径：`POST /api/tasks`**  
**参数位置：JSON 请求体，`TaskInput` 的字段加上 `version`。**

| 参数组 | 规则 |
|---|---|
| `title`、`description`、`assignee`、`priority`、`status`、`dueDate` | 见第 2 节任务字段 |
| `version` | 必填，正整数，必须等于当前服务端版本 |

沿用前面的 `version: 7`：

```json
{
  "title": "实现登录页面",
  "description": "接入登录接口",
  "assignee": "陈晨",
  "priority": "medium",
  "status": "todo",
  "dueDate": "2026-10-31",
  "version": 7
}
```

新任务追加到指定状态列末尾，返回示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "title": "梳理需求",
        "description": "明确交付标准",
        "assignee": "林悦",
        "priority": "high",
        "status": "todo",
        "dueDate": "2026-10-30",
        "updatedAt": "2026-10-29 10:00:00"
      },
      {
        "id": 2,
        "title": "实现登录页面",
        "description": "接入登录接口",
        "assignee": "陈晨",
        "priority": "medium",
        "status": "todo",
        "dueDate": "2026-10-31",
        "updatedAt": "2026-10-29 10:01:00"
      }
    ],
    "version": 8
  },
  "message": ""
}
```

返回的是两条任务组成的完整快照。不能把 `data` 当成新任务对象再追加到旧数组，否则会造成结构错误或重复数据。

字段不合法返回 `400`；版本冲突或看板已有 200 条任务时返回 `409`。

## 5. 编辑任务

**接口名称：完整更新任务业务字段**  
**方法与路径：`PUT /api/tasks/:id`**

| 参数位置 | 参数 | 类型与要求 |
|---|---|---|
| Path | `id` | 待编辑任务的正整数 ID |
| Body | `TaskInput` 的全部字段 | 与新增相同，完整提交业务字段 |
| Body | `version` | 当前看板版本，正整数 |

例如 `PUT /api/tasks/2`：

```json
{
  "title": "实现登录页面",
  "description": "",
  "assignee": "",
  "priority": "high",
  "status": "doing",
  "dueDate": "",
  "version": 8
}
```

清空描述、负责人和截止日期，并把任务移动到进行中。编辑时状态不变会保留原列内位置；状态改变则追加到目标列末尾。

成功响应示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 2,
        "title": "实现登录页面",
        "description": "",
        "assignee": "",
        "priority": "high",
        "status": "doing",
        "dueDate": "",
        "updatedAt": "2026-10-29 10:02:00"
      },
      {
        "id": 1,
        "title": "梳理需求",
        "description": "明确交付标准",
        "assignee": "林悦",
        "priority": "high",
        "status": "todo",
        "dueDate": "2026-10-30",
        "updatedAt": "2026-10-29 10:00:00"
      }
    ],
    "version": 9
  },
  "message": ""
}
```

非法字段返回 `400`，任务不存在返回 `404`，版本过期返回 `409`。不能把 `Partial<TaskInput>` 当成编辑请求；这是完整更新，不是只修改非空字段。

## 6. 移动任务与列内排序

**接口名称：移动任务到目标状态列或调整列内顺序**  
**方法与路径：`PATCH /api/tasks/:id/move`**

| 参数位置 | 参数 | 类型 | 规则 |
|---|---|---|---|
| Path | `id` | `number` | 被移动任务的正整数 ID |
| Body | `status` | `TaskStatus` | 目标列：todo、doing 或 done |
| Body | `beforeId` | `number` | 目标任务 ID；0 表示列末尾；不能为负数或当前任务自己的 ID |
| Body | `version` | `number` | 当前看板版本，正整数 |

`beforeId` 表示“插入到哪个任务之前”，不是数组下标，也不是前一个任务 ID。非 0 时，该任务必须存在于目标列中。前端应显式传入；后端省略时按 0 处理。

例如 `PATCH /api/tasks/1/move`：

```json
{
  "status": "doing",
  "beforeId": 2,
  "version": 9
}
```

表示把任务 1 移到进行中列，放在任务 2 前面。成功响应：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "title": "梳理需求",
        "description": "明确交付标准",
        "assignee": "林悦",
        "priority": "high",
        "status": "doing",
        "dueDate": "2026-10-30",
        "updatedAt": "2026-10-29 10:03:00"
      },
      {
        "id": 2,
        "title": "实现登录页面",
        "description": "",
        "assignee": "",
        "priority": "high",
        "status": "doing",
        "dueDate": "",
        "updatedAt": "2026-10-29 10:02:00"
      }
    ],
    "version": 10
  },
  "message": ""
}
```

| 交互 | 参数含义 |
|---|---|
| 同列排序 | `status` 不变，`beforeId` 指定该列中的另一个任务 |
| 移到列末尾 | `beforeId: 0` |
| 移到空列 | 使用目标 `status` 和 `beforeId: 0` |
| 跨列插入 | `status` 为新列，`beforeId` 为新列内的任务 ID |

服务端在事务中更新目标列顺序；不接受整列 ID 数组或 `position` 字段。非法移动参数、锚点不在目标列返回 `400`，被移动任务不存在返回 `404`，旧版本返回 `409`。

当前页面在筛选状态下禁止拖拽，避免可见子集与完整顺序不一致；这是前端交互约束，服务端没有接收“是否筛选”的参数。

## 7. 删除任务

**接口名称：删除任务并返回完整看板**  
**方法与路径：`DELETE /api/tasks/:id`**

| 参数位置 | 参数 | 类型与要求 |
|---|---|---|
| Path | `id` | 被删除任务的正整数 ID |
| Body | `version` | 必填，当前看板版本，正整数 |

例如 `DELETE /api/tasks/2`，必须发送 JSON 请求体：

```json
{
  "version": 10
}
```

成功响应：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "title": "梳理需求",
        "description": "明确交付标准",
        "assignee": "林悦",
        "priority": "high",
        "status": "doing",
        "dueDate": "2026-10-30",
        "updatedAt": "2026-10-29 10:03:00"
      }
    ],
    "version": 11
  },
  "message": ""
}
```

删除永久生效，成功后用返回的快照替换本地看板。缺少请求体、缺少 version 或 version 小于 1 返回 `400`。当前版本下删除不存在的任务返回 `404`；过期版本优先触发 `409`。

这与商品删除不同：商品 DELETE 无请求体，任务 DELETE 需要 `{ version }`。当前 RequestClient 的第二个参数是请求配置，正确写法是 `delete(url, { data: { version } })`，不是 `delete(url, { version })`。

## 8. 完整 TypeScript 类型与请求封装

对应文件：`frontend/apps/admin/src/api/tasks.ts`。下面保持现有函数名与调用方式，额外给传输请求体命名，便于教学和复用。`TaskMove` 是前端事件数据，包含 URL 所需的 `id`；真正发送的 `TaskMoveInput` 不包含 `id`，但必须包含 `version`。

```ts
import { requestClient } from '#/api/request';

export type TaskStatus = 'todo' | 'doing' | 'done';
export type TaskPriority = 'low' | 'medium' | 'high';

export interface TaskInput {
  title: string;
  description: string;
  assignee: string;
  priority: TaskPriority;
  status: TaskStatus;
  /** 空字符串，或 YYYY-MM-DD 日期。 */
  dueDate: string;
}

export interface Task extends TaskInput {
  id: number;
  updatedAt: string;
}

export interface TaskBoard {
  items: Task[];
  version: number;
}

export interface TaskSaveInput extends TaskInput {
  version: number;
}

export interface TaskDeleteInput {
  version: number;
}

export interface TaskMoveInput {
  status: TaskStatus;
  beforeId: number;
  version: number;
}

/** 前端拖拽事件使用；id 会进入 URL，version 在发请求时补充。 */
export interface TaskMove {
  id: number;
  status: TaskStatus;
  beforeId: number;
}

export const getTaskBoard = (): Promise<TaskBoard> =>
  requestClient.get<TaskBoard>('/tasks');

export const saveTask = (
  data: TaskInput,
  version: number,
  id?: number,
): Promise<TaskBoard> => {
  const body: TaskSaveInput = { ...data, version };
  return id
    ? requestClient.put<TaskBoard>(`/tasks/${id}`, body)
    : requestClient.post<TaskBoard>('/tasks', body);
};

export const removeTask = (
  id: number,
  version: number,
): Promise<TaskBoard> => {
  const data: TaskDeleteInput = { version };
  return requestClient.delete<TaskBoard>(`/tasks/${id}`, { data });
};

export const moveTask = (
  { id, status, beforeId }: TaskMove,
  version: number,
): Promise<TaskBoard> => {
  const data: TaskMoveInput = { status, beforeId, version };
  return requestClient.request<TaskBoard>(`/tasks/${id}/move`, {
    method: 'PATCH',
    data,
  });
};
```

新建时不传 `saveTask` 的第三个参数；编辑时传正整数 ID，不要用 `0` 表示编辑目标。类型中的 `number` 对应当前 Go JSON 数值；ID、version 在本课程数据规模下可安全表示，若将来需要超过 JavaScript 安全整数范围，必须同时调整前后端传输契约。

## 9. 前端状态更新与失败处理

本项目使用 `frontend/apps/admin/src/composables/use-task-board.ts` 统一处理这些接口：

- 首次查询成功后保存完整 `TaskBoard`。
- 一次只允许一个写入，写入期间禁用重复提交。
- 新增、编辑、删除成功后，直接使用服务端返回的完整快照。
- 拖拽先保存原快照，立即更新卡片位置；接口成功后以服务端快照为准。
- 写入失败时恢复本地快照，再重新 GET 同步服务端；重新读取也失败时，禁用写入直到刷新成功。

超时并不一定意味着服务端没有完成写入，所以失败后需要重读。收到 409 时保留编辑草稿，读取最新版本，让用户核对后再次提交，不要只换一个新 version 就自动覆盖他人的修改。

## 10. 错误响应与源码位置

版本冲突响应示例：

```json
{
  "code": 409,
  "data": null,
  "message": "看板已被其他操作更新，请刷新后重试"
}
```

| HTTP / code | 原因 | 当前典型 message |
|---|---|---|
| 400 | 标题为空或过长 | `任务标题需为 1～80 个字符` |
| 400 | 描述或负责人过长 | `描述最多 2000 字，负责人最多 30 字` |
| 400 | 保存或删除的版本小于 1 | `看板版本不合法` |
| 400 | 状态或优先级非法 | `任务状态不合法` / `任务优先级不合法` |
| 400 | 日期非法 | `截止日期格式应为 YYYY-MM-DD` |
| 400 | 移动版本、目标状态、beforeId 不合法 | `移动参数不合法` |
| 400 | 非 0 锚点不在目标列 | `目标位置不合法，请刷新后重试` |
| 400 | JSON 类型错误、未知字段或请求体超限 | `请求参数格式错误` |
| 401 | 缺少令牌或会话失效 | `登录已过期，请重新登录` |
| 404 | 当前版本下目标任务不存在 | `任务不存在或已删除` |
| 409 | 版本冲突 | `看板已被其他操作更新，请刷新后重试` |
| 409 | 新建时看板已有 200 条任务 | `本课堂看板最多支持 200 条任务` |
| 500 | 数据库或服务异常 | `服务暂时不可用，请稍后重试` |

非法 Path ID 也返回 400。当前任务和商品复用 `httpx.ID`，其消息仍为 `商品 ID 不合法`，即使请求的是任务接口也如此；前端应按 HTTP 状态码处理，不依赖该字符串判断所属模块。

字段校验通过后，事务会先检查版本再查目标任务，因此“版本过期且任务已删除”的请求可能先返回 409，而不是 404。

| 实现文件 | 对应内容 |
|---|---|
| `backend/router/admin/router.go` | 五个任务路由 |
| `backend/internal/task/admin/dto/task.go` | TaskInput、写入 DTO、Task 与 Board |
| `backend/internal/task/admin/handler/task.go` | Path/JSON 绑定和完整看板响应 |
| `backend/internal/task/admin/service/task.go` | 字段、日期、状态和错误映射 |
| `backend/internal/task/repository/task_repository.go` | 全看板版本锁、事务、锚点与排序 |
| `frontend/apps/admin/src/api/tasks.ts` | 类型与请求函数 |
| `frontend/apps/admin/src/composables/use-task-board.ts` | 防重提交、乐观更新、回滚与重新同步 |

