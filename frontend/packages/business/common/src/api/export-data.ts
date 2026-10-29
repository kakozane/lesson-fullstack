import type { DataTaskApi } from './data-request';

import {
  downloadBlobResponse,
  getDataRequestClient,
  getDataTaskProgressApi,
} from './data-request';

export namespace ExportDataApi {
  export type ModelType = DataTaskApi.ModelType;
  export type TaskStatus = DataTaskApi.TaskStatus;
  export type TaskProgress = DataTaskApi.TaskProgress;

  export type ExportFormat = 'csv' | 'excel';

  export interface ExportParams {
    all_data?: boolean;
    file_name?: string;
    filter?: Record<string, any>;
    format?: ExportFormat;
    ids?: number[];
    model_type: ModelType;
  }

  export type ExportResult = DataTaskApi.TaskResult;
}

/**
 * 创建导出任务
 * POST /v1/admin/data/export
 */
export function exportDataApi(params: ExportDataApi.ExportParams) {
  return getDataRequestClient().post<ExportDataApi.ExportResult>(
    '/v1/admin/data/export',
    params,
  );
}

/**
 * 获取导出任务进度
 * GET /v1/admin/data/tasks/:id/progress?type=export
 */
export function getExportTaskProgressApi(taskId: string) {
  return getDataTaskProgressApi(taskId, 'export');
}

/**
 * 下载导出文件
 * GET /v1/admin/data/tasks/:id/download
 */
export function downloadExportFileApi(
  taskId: string,
  defaultFileName = 'export.xlsx',
) {
  return downloadBlobResponse(
    `/v1/admin/data/tasks/${taskId}/download`,
    defaultFileName,
  );
}
