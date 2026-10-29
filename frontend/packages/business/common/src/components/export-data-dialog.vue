<script lang="ts" setup>
import type { ExportDataApi } from '../api/export-data';

import { computed, onBeforeUnmount, ref, watch } from 'vue';

import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElMessage,
  ElProgress,
  ElRadio,
  ElRadioGroup,
} from 'element-plus';

import {
  downloadExportFileApi,
  exportDataApi,
  getExportTaskProgressApi,
} from '../api/export-data';

type ExportScope = 'all' | 'selected';

interface Props {
  /** 导出文件名，不含扩展名 */
  fileName: string;
  filter?: Record<string, any>;
  modelType: ExportDataApi.ModelType;
  selectedCount?: number;
  selectedIds?: number[];
  totalCount?: number;
}

defineOptions({
  name: 'ExportDialog',
});

const props = withDefaults(defineProps<Props>(), {
  filter: () => ({}),
  selectedCount: 0,
  selectedIds: () => [],
  totalCount: 0,
});

const emit = defineEmits<{
  success: [];
}>();

const visible = defineModel<boolean>({ default: false });

const POLL_INTERVAL_MS = 1500;
const MAX_POLL_COUNT = 120;
const STABLE_PROGRESS_THRESHOLD = 3;

const scope = ref<ExportScope>('selected');
const format = ref<ExportDataApi.ExportFormat>('excel');
const exporting = ref(false);
const exportProgress = ref(0);

let pollTimer: ReturnType<typeof setTimeout> | undefined;
let pollCount = 0;
let stableProgressCount = 0;
let lastUpdatedAt = '';

const downloadFileName = computed(() => {
  const extension = format.value === 'csv' ? 'csv' : 'xlsx';
  return `${props.fileName}.${extension}`;
});

function updateVisible(value: boolean) {
  visible.value = value;
}

function clearPollTimer() {
  if (pollTimer) {
    clearTimeout(pollTimer);
    pollTimer = undefined;
  }
}

function resetState() {
  clearPollTimer();
  pollCount = 0;
  stableProgressCount = 0;
  lastUpdatedAt = '';
  exportProgress.value = 0;
  exporting.value = false;
  scope.value = props.selectedCount > 0 ? 'selected' : 'all';
  format.value = 'excel';
}

function parseCount(value?: string) {
  const count = Number(value ?? 0);
  return Number.isFinite(count) ? count : 0;
}

function isSuccessStatus(status?: string) {
  // 任务表状态是 success，Redis 进度完成时写的是 completed
  return status === 'success' || status === 'completed';
}

function isTerminalStatus(status?: string) {
  return isSuccessStatus(status) || status === 'failed';
}

function scheduleNextPoll(taskId: string) {
  pollTimer = setTimeout(() => {
    void pollExportProgress(taskId);
  }, POLL_INTERVAL_MS);
}

async function handleDownload(taskId: string) {
  await downloadExportFileApi(taskId, downloadFileName.value);
  exportProgress.value = 100;
  ElMessage.success('导出成功');
  updateVisible(false);
  emit('success');
}

async function pollExportProgress(taskId: string) {
  pollCount += 1;

  if (pollCount > MAX_POLL_COUNT) {
    exporting.value = false;
    ElMessage.error('导出任务处理超时，请稍后重试');
    return;
  }

  try {
    const progress = await getExportTaskProgressApi(taskId);
    const processedCount = parseCount(progress.processed);
    const failedCount = parseCount(progress.failed);
    const status = progress.status;

    if (processedCount > 0 || failedCount > 0) {
      exportProgress.value = Math.min(
        95,
        Math.max(exportProgress.value, processedCount + failedCount),
      );
    } else if (exportProgress.value < 30) {
      exportProgress.value += 5;
    }

    if (status === 'failed') {
      exporting.value = false;
      ElMessage.error(progress.message || '导出失败，请稍后重试');
      return;
    }

    if (isSuccessStatus(status)) {
      exportProgress.value = 90;
      await handleDownload(taskId);
      exporting.value = false;
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
        exportProgress.value = 90;
        await handleDownload(taskId);
        exporting.value = false;
        return;
      }
    }

    if (isTerminalStatus(status)) {
      if (isSuccessStatus(status)) {
        exportProgress.value = 90;
        await handleDownload(taskId);
      } else {
        ElMessage.error(progress.message || '导出失败，请稍后重试');
      }
      exporting.value = false;
      return;
    }

    scheduleNextPoll(taskId);
  } catch (error) {
    if (pollCount >= STABLE_PROGRESS_THRESHOLD) {
      exporting.value = false;
      const requestError = error as { response?: unknown };
      const message =
        error instanceof Error && !requestError.response
          ? error.message
          : '导出任务处理失败，请稍后重试';
      ElMessage.error(message || '导出任务处理失败，请稍后重试');
      return;
    }

    scheduleNextPoll(taskId);
  }
}

async function handleExport() {
  if (scope.value === 'selected' && props.selectedCount === 0) {
    ElMessage.warning('请先选择要导出的数据');
    return;
  }

  exporting.value = true;
  exportProgress.value = 10;

  try {
    const result = await exportDataApi({
      all_data: scope.value === 'all',
      file_name: props.fileName,
      filter: props.filter,
      format: format.value,
      ids: scope.value === 'selected' ? props.selectedIds : undefined,
      model_type: props.modelType,
    });

    exportProgress.value = 20;
    await pollExportProgress(result.task_id);
  } catch (error) {
    exporting.value = false;
    exportProgress.value = 0;
    if (
      error instanceof Error &&
      error.message.includes('setupExportRequest')
    ) {
      ElMessage.error(error.message);
    }
  }
}

watch(visible, (value) => {
  if (value) {
    resetState();
  } else {
    clearPollTimer();
  }
});

onBeforeUnmount(() => {
  clearPollTimer();
});
</script>

<template>
  <ElDialog
    :close-on-click-modal="!exporting"
    :close-on-press-escape="!exporting"
    :model-value="visible"
    :show-close="!exporting"
    title="导出数据"
    width="520px"
    @update:model-value="updateVisible"
  >
    <ElForm label-width="100px">
      <ElFormItem label="操作范围">
        <ElRadioGroup v-model="scope" :disabled="exporting">
          <ElRadio value="selected">
            选中的 {{ selectedCount }} 条数据
          </ElRadio>
          <ElRadio value="all">全部数据（共 {{ totalCount }} 条）</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem label="导出格式">
        <ElRadioGroup v-model="format" :disabled="exporting">
          <ElRadio value="excel">XLSX</ElRadio>
          <ElRadio value="csv">CSV</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem v-if="exporting" label="导出进度">
        <div class="flex w-full flex-col gap-3">
          <ElProgress :percentage="exportProgress" :stroke-width="16" />
          <span class="text-sm text-[var(--el-text-color-secondary)]">
            正在导出{{ fileName }}，请稍候...
          </span>
        </div>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton :disabled="exporting" @click="updateVisible(false)">
        取消
      </ElButton>
      <ElButton :loading="exporting" type="primary" @click="handleExport">
        导出
      </ElButton>
    </template>
  </ElDialog>
</template>
