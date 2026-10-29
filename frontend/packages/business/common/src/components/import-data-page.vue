<script lang="ts" setup>
import type { UploadFile, UploadInstance } from 'element-plus';

import type { ImportDataApi } from '../api/import-data';

import { computed, nextTick, onBeforeUnmount, ref } from 'vue';
import { useRoute } from 'vue-router';

import { IconifyIcon } from '@vben/icons';

import {
  ElAlert,
  ElButton,
  ElCard,
  ElMessage,
  ElProgress,
  ElUpload,
} from 'element-plus';

import {
  downloadImportFailureReportApi,
  downloadImportTemplateApi,
  getImportTaskProgressApi,
  importDataApi,
} from '../api/import-data';
import { usePageBack } from '../composables/use-page-back';

interface ImportRouteMeta {
  importModelType?: string;
  templateFileName?: string;
  title?: string;
}

interface Props {
  /** 失败报告默认文件名 */
  failureReportFileName?: string;
  /** 与后端 model_type 一致；不传时读取路由 meta.importModelType */
  modelType?: string;
  /** 模板下载默认文件名 */
  templateFileName?: string;
  /** 页面标题，如「商品导入」 */
  title?: string;
}

defineOptions({
  name: 'ImportDataPage',
});

const props = withDefaults(defineProps<Props>(), {
  failureReportFileName: '导入失败报告.xlsx',
  modelType: '',
  templateFileName: '',
  title: '',
});

const POLL_INTERVAL_MS = 1500;
const MAX_POLL_COUNT = 120;
const STABLE_PROGRESS_THRESHOLD = 3;

const { handleBack } = usePageBack();
const route = useRoute();

const uploadRef = ref<UploadInstance>();
const selectedFile = ref<UploadFile>();
const downloadingTemplate = ref(false);
const importing = ref(false);
const downloadingReport = ref(false);
const importProgress = ref(0);
const importResult = ref<{
  failedCount: number;
  message: string;
  processedCount: number;
  status: 'failed' | 'success';
  taskId: string;
}>();

let pollTimer: ReturnType<typeof setTimeout> | undefined;
let pollCount = 0;
let stableProgressCount = 0;
let lastUpdatedAt = '';

const routeMeta = computed(() => route.meta as ImportRouteMeta);

const modelType = computed(
  () => props.modelType || routeMeta.value.importModelType || '',
);

const pageTitle = computed(
  () => props.title || routeMeta.value.title || '数据导入',
);

const templateFileName = computed(
  () =>
    props.templateFileName ||
    routeMeta.value.templateFileName ||
    '导入模板.xlsx',
);

const isImporting = computed(() => importing.value);

function clearPollTimer() {
  if (pollTimer) {
    clearTimeout(pollTimer);
    pollTimer = undefined;
  }
}

function resetImportState() {
  clearPollTimer();
  pollCount = 0;
  stableProgressCount = 0;
  lastUpdatedAt = '';
  importProgress.value = 0;
  importResult.value = undefined;
}

function parseCount(value?: string) {
  const count = Number(value ?? 0);
  return Number.isFinite(count) ? count : 0;
}

function isTerminalStatus(status?: ImportDataApi.TaskProgress['status']) {
  return status === 'success' || status === 'failed' || status === 'completed';
}

function buildResultMessage(
  processedCount: number,
  failedCount: number,
  status: 'failed' | 'success',
) {
  if (status === 'failed') {
    if (failedCount > 0) {
      return `导入完成，成功 ${processedCount} 行，失败 ${failedCount} 行`;
    }
    return '导入失败，请检查文件格式后重试';
  }

  return `导入成功，共导入 ${processedCount} 行`;
}

function finishImport(
  taskId: string,
  processedCount: number,
  failedCount: number,
  status: 'failed' | 'success',
  message?: string,
) {
  clearPollTimer();
  importing.value = false;
  importProgress.value = 100;
  importResult.value = {
    failedCount,
    message: message ?? buildResultMessage(processedCount, failedCount, status),
    processedCount,
    status,
    taskId,
  };
}

function scheduleNextPoll(taskId: string) {
  pollTimer = setTimeout(() => {
    void pollImportProgress(taskId);
  }, POLL_INTERVAL_MS);
}

async function pollImportProgress(taskId: string) {
  pollCount += 1;

  if (pollCount > MAX_POLL_COUNT) {
    importing.value = false;
    importResult.value = {
      failedCount: 0,
      message: '导入任务处理超时，请稍后重试',
      processedCount: 0,
      status: 'failed',
      taskId,
    };
    return;
  }

  try {
    const progress = await getImportTaskProgressApi(taskId);
    const processedCount = parseCount(progress.processed);
    const failedCount = parseCount(progress.failed);
    const status = progress.status;
    const message = progress.message;

    if (processedCount > 0 || failedCount > 0) {
      importProgress.value = Math.min(
        95,
        Math.max(importProgress.value, processedCount + failedCount),
      );
    } else if (importProgress.value < 30) {
      importProgress.value += 5;
    }

    if (isTerminalStatus(status)) {
      finishImport(
        taskId,
        processedCount,
        failedCount,
        failedCount > 0 || status === 'failed' ? 'failed' : 'success',
        message,
      );
      return;
    }

    if (progress.updated_at) {
      if (progress.updated_at === lastUpdatedAt) {
        stableProgressCount += 1;
      } else {
        stableProgressCount = 0;
        lastUpdatedAt = progress.updated_at;
      }

      if (stableProgressCount >= STABLE_PROGRESS_THRESHOLD) {
        finishImport(
          taskId,
          processedCount,
          failedCount,
          failedCount > 0 ? 'failed' : 'success',
          message,
        );
        return;
      }
    }

    scheduleNextPoll(taskId);
  } catch {
    if (pollCount >= STABLE_PROGRESS_THRESHOLD) {
      importing.value = false;
      importResult.value = {
        failedCount: 0,
        message: '导入任务处理失败，请检查文件格式后重试',
        processedCount: 0,
        status: 'failed',
        taskId,
      };
      return;
    }

    scheduleNextPoll(taskId);
  }
}

function ensureModelType() {
  if (modelType.value) {
    return true;
  }
  ElMessage.error('未配置导入模型类型');
  return false;
}

function showDownloadError(error: unknown) {
  const requestError = error as { response?: unknown };
  if (error instanceof Error && !requestError.response) {
    ElMessage.error(error.message || '下载失败');
  }
}

async function handleDownloadTemplate() {
  if (!ensureModelType()) {
    return;
  }

  downloadingTemplate.value = true;
  try {
    await downloadImportTemplateApi(modelType.value, templateFileName.value);
  } catch (error) {
    showDownloadError(error);
  } finally {
    downloadingTemplate.value = false;
  }
}

function handleFileChange(file: UploadFile) {
  selectedFile.value = file;
  resetImportState();
  void nextTick(() => {
    uploadRef.value?.clearFiles();
  });
}

function isExcelFile(fileName: string) {
  return /\.(?:xls|xlsx)$/i.test(fileName);
}

async function handleImport() {
  const rawFile = selectedFile.value?.raw;
  if (!rawFile) {
    ElMessage.warning('请先选择上传文件');
    return;
  }

  if (!isExcelFile(rawFile.name)) {
    ElMessage.warning('请上传 XLS 或 XLSX 文件');
    return;
  }

  if (!ensureModelType()) {
    return;
  }

  resetImportState();
  importing.value = true;
  importProgress.value = 10;

  try {
    const result = await importDataApi(rawFile, modelType.value);
    importProgress.value = 20;
    await pollImportProgress(result.task_id);
  } catch (error) {
    importing.value = false;
    importProgress.value = 0;
    if (
      error instanceof Error &&
      error.message.includes('setupExportRequest')
    ) {
      ElMessage.error(error.message);
    }
  }
}

async function handleDownloadFailureReport() {
  if (!importResult.value?.taskId) {
    return;
  }

  downloadingReport.value = true;
  try {
    await downloadImportFailureReportApi(
      importResult.value.taskId,
      props.failureReportFileName,
    );
  } catch (error) {
    showDownloadError(error);
  } finally {
    downloadingReport.value = false;
  }
}

onBeforeUnmount(() => {
  clearPollTimer();
});
</script>

<template>
  <div class="p-5">
    <ElButton
      class="mb-4 !h-auto !px-0"
      link
      type="primary"
      @click="handleBack()"
    >
      <IconifyIcon class="mr-1 size-4" icon="lucide:arrow-left" />
      返回
    </ElButton>

    <ElCard>
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-base font-medium">{{ pageTitle }}</span>
        </div>
      </template>

      <div class="flex min-h-[360px] flex-col gap-8">
        <section class="flex flex-col gap-4">
          <div
            class="rounded border-l-4 border-[var(--el-color-primary)] bg-[var(--el-fill-color-light)] px-4 py-3.5 text-sm font-medium leading-normal text-[var(--el-text-color-primary)]"
          >
            第一步：下载文件模板，并按照规则填写
          </div>
          <div class="flex flex-wrap items-center gap-3 pl-1">
            <ElButton
              :loading="downloadingTemplate"
              type="primary"
              @click="handleDownloadTemplate"
            >
              下载导入模板
            </ElButton>
            <span class="text-sm text-[var(--el-color-warning)]">
              请先下载文件模板再上传文件！
            </span>
          </div>
        </section>

        <section class="flex flex-col gap-4">
          <div
            class="rounded border-l-4 border-[var(--el-color-primary)] bg-[var(--el-fill-color-light)] px-4 py-3.5 text-sm font-medium leading-normal text-[var(--el-text-color-primary)]"
          >
            第二步：上传文件模板
          </div>
          <div class="flex flex-wrap items-center gap-3 pl-1">
            <ElUpload
              ref="uploadRef"
              :auto-upload="false"
              :disabled="isImporting"
              :limit="1"
              :on-change="handleFileChange"
              :show-file-list="false"
              accept=".xls,.xlsx"
              class="inline-flex"
            >
              <ElButton :disabled="isImporting" type="primary">
                选择上传文件
              </ElButton>
            </ElUpload>
            <ElButton
              :disabled="!selectedFile || isImporting"
              :loading="isImporting"
              plain
              type="primary"
              @click="handleImport"
            >
              导入(XLS/XLSX)
            </ElButton>
            <span
              v-if="selectedFile"
              class="text-sm text-[var(--el-text-color-secondary)]"
            >
              已选择：{{ selectedFile.name }}
            </span>
          </div>
        </section>

        <section v-if="isImporting" class="flex flex-col gap-4">
          <div
            class="rounded border-l-4 border-[var(--el-color-primary)] bg-[var(--el-fill-color-light)] px-4 py-3.5 text-sm font-medium leading-normal text-[var(--el-text-color-primary)]"
          >
            第三步：导入处理中
          </div>
          <div class="flex max-w-[640px] flex-col gap-3 pl-1">
            <ElProgress :percentage="importProgress" :stroke-width="16" />
            <span class="text-sm text-[var(--el-text-color-secondary)]">
              正在导入数据，请稍候...
            </span>
          </div>
        </section>

        <section v-if="importResult" class="flex flex-col gap-4">
          <div
            class="rounded border-l-4 border-[var(--el-color-primary)] bg-[var(--el-fill-color-light)] px-4 py-3.5 text-sm font-medium leading-normal text-[var(--el-text-color-primary)]"
          >
            导入结果
          </div>
          <div class="flex max-w-[720px] flex-col gap-4 pl-1">
            <ElAlert
              :closable="false"
              :title="importResult.message"
              :type="importResult.status === 'success' ? 'success' : 'error'"
              show-icon
            />
            <div
              v-if="
                importResult.status === 'failed' && importResult.failedCount > 0
              "
            >
              <ElButton
                :loading="downloadingReport"
                type="primary"
                @click="handleDownloadFailureReport"
              >
                下载导入失败 Excel
              </ElButton>
            </div>
          </div>
        </section>
      </div>
    </ElCard>
  </div>
</template>
