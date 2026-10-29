import type { DataTaskApi } from './data-request';

import {
  downloadBlobResponse,
  getDataRequestClient,
  getDataTaskProgressApi,
} from './data-request';

export namespace ImportDataApi {
  export type ModelType = DataTaskApi.ModelType;
  export type TaskStatus = DataTaskApi.TaskStatus;
  export type TaskProgress = DataTaskApi.TaskProgress;
  export type ImportResult = DataTaskApi.TaskResult;
}

/**
 * 下载导入模板
 * GET /v1/admin/data/template/:type
 */
export function downloadImportTemplateApi(
  type: ImportDataApi.ModelType,
  fileName = '导入模板.xlsx',
) {
  return downloadBlobResponse(`/v1/admin/data/template/${type}`, fileName);
}

/**
 * 创建导入任务
 * POST /v1/admin/data/import
 */
export function importDataApi(file: File, modelType: ImportDataApi.ModelType) {
  return getDataRequestClient().upload<ImportDataApi.ImportResult>(
    '/v1/admin/data/import',
    {
      file,
      model_type: modelType,
    },
  );
}

/**
 * 获取导入任务进度
 * GET /v1/admin/data/tasks/:id/progress?type=import
 */
export function getImportTaskProgressApi(taskId: string) {
  return getDataTaskProgressApi(taskId, 'import');
}

/**
 * 下载导入失败报告
 * GET /v1/admin/data/tasks/:id/report
 */
export function downloadImportFailureReportApi(
  taskId: string,
  fileName = '导入失败报告.xlsx',
) {
  return downloadBlobResponse(
    `/v1/admin/data/tasks/${taskId}/report`,
    fileName,
  );
}
