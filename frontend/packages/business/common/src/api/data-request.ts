import type { RequestClient } from '@vben/request';

import { downloadFileFromBlob } from '@vben/utils';

interface RawBlobResponse {
  data: Blob;
}

let requestClient: RequestClient | undefined;

/** 导入、导出任务共用的模型和进度结构 */
export namespace DataTaskApi {
  /** 与后端 model_type 一致，如 activity、association */
  export type ModelType = string;

  export type TaskStatus =
    | 'completed'
    | 'failed'
    | 'pending'
    | 'processing'
    | 'success';

  export type TaskKind = 'export' | 'import';

  export interface TaskResult {
    task_id: string;
  }

  /** Redis 进度字段都是字符串 */
  export interface TaskProgress {
    failed?: string;
    message?: string;
    processed?: string;
    status?: string | TaskStatus;
    updated_at?: string;
  }
}

/**
 * 注册带登录态的请求客户端。导出、导入共用。
 * 各应用在创建 requestClient 后调用 setupExportRequest 即可。
 */
export function setupDataRequest(client: RequestClient) {
  requestClient = client;
}

export function getDataRequestClient() {
  if (!requestClient) {
    throw new Error('数据接口未初始化，请先调用 setupExportRequest');
  }
  return requestClient;
}

export async function downloadBlobResponse(
  url: string,
  defaultFileName: string,
) {
  const response = await getDataRequestClient().download<RawBlobResponse>(url, {
    responseReturn: 'raw',
  });
  const blob = response.data;

  if (blob.type.includes('application/json')) {
    const text = await blob.text();
    const json = JSON.parse(text) as { msg?: string };
    throw new Error(json.msg || '下载失败');
  }

  downloadFileFromBlob({ fileName: defaultFileName, source: blob });
}

/**
 * 获取导入或导出任务进度
 * GET /v1/admin/data/tasks/:id/progress
 */
export function getDataTaskProgressApi(
  taskId: string,
  type: DataTaskApi.TaskKind,
) {
  return getDataRequestClient().get<DataTaskApi.TaskProgress>(
    `/v1/admin/data/tasks/${taskId}/progress`,
    {
      params: {
        type,
      },
    },
  );
}
