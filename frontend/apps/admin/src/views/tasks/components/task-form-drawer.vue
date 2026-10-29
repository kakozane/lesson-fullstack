<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus';
import type { Task, TaskInput, TaskStatus } from '#/api/tasks';
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import {
  ElButton,
  ElDatePicker,
  ElDrawer,
  ElForm,
  ElFormItem,
  ElInput,
  ElMessageBox,
  ElOption,
  ElSelect,
} from 'element-plus';
import { priorities, statuses } from '../task-meta';

const props = defineProps<{
  modelValue: boolean;
  task?: Task;
  initialStatus: TaskStatus;
  save: (data: TaskInput, id?: number) => Promise<boolean>;
}>();
const emit = defineEmits<{ 'update:modelValue': [boolean] }>();
const formRef = ref<FormInstance>();
const submitting = ref(false);
const form = reactive<TaskInput>({
  title: '',
  description: '',
  assignee: '',
  priority: 'medium',
  status: 'todo',
  dueDate: '',
});
let initial = '';
const dirty = computed(
  () => props.modelValue && JSON.stringify(form) !== initial,
);
const rules: FormRules = {
  title: [
    {
      required: true,
      whitespace: true,
      message: '请输入任务标题',
      trigger: 'blur',
    },
    { max: 80, message: '标题最多 80 字', trigger: 'blur' },
  ],
  priority: [{ required: true, message: '请选择优先级', trigger: 'change' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
  assignee: [{ max: 30, message: '负责人最多 30 字', trigger: 'blur' }],
  description: [{ max: 2000, message: '描述最多 2000 字', trigger: 'blur' }],
};
watch(
  () => props.modelValue,
  async (open) => {
    if (!open) return;
    const t = props.task;
    Object.assign(
      form,
      t
        ? {
            title: t.title,
            description: t.description,
            assignee: t.assignee,
            priority: t.priority,
            status: t.status,
            dueDate: t.dueDate,
          }
        : {
            title: '',
            description: '',
            assignee: '',
            priority: 'medium',
            status: props.initialStatus,
            dueDate: '',
          },
    );
    initial = JSON.stringify(form);
    await nextTick();
    formRef.value?.clearValidate();
  },
);
async function canLeave() {
  if (submitting.value) return false;
  if (!dirty.value) return true;
  try {
    await ElMessageBox.confirm(
      '任务尚有未保存的修改，确定放弃吗？',
      '放弃修改',
      {
        confirmButtonText: '放弃修改',
        cancelButtonText: '继续编辑',
        type: 'warning',
      },
    );
    return true;
  } catch {
    return false;
  }
}
async function close() {
  if (await canLeave()) emit('update:modelValue', false);
}
onBeforeRouteLeave((to) => to.path.startsWith('/auth/') || canLeave());
function beforeUnload(event: BeforeUnloadEvent) {
  if (dirty.value || submitting.value) {
    event.preventDefault();
    event.returnValue = '';
  }
}
window.addEventListener('beforeunload', beforeUnload);
onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload));
async function submit() {
  if (submitting.value) return;
  submitting.value = true;
  try {
    if (!(await formRef.value?.validate().catch(() => false))) return;
    if (
      await props.save(
        {
          ...form,
          title: form.title.trim(),
          assignee: form.assignee.trim(),
          description: form.description.trim(),
          dueDate: form.dueDate || '',
        },
        props.task?.id,
      )
    )
      emit('update:modelValue', false);
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <ElDrawer
    :model-value="modelValue"
    :title="task ? '编辑任务' : '新建任务'"
    size="520px"
    class="task-drawer"
    :before-close="close"
    :close-on-click-modal="false"
    :close-on-press-escape="!submitting"
    :show-close="!submitting"
    destroy-on-close
  >
    <p class="text-muted-foreground mb-6 text-sm">
      明确目标、负责人和截止日期，让协作从一张任务卡开始。
    </p>
    <ElForm
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      :disabled="submitting"
      @submit.prevent="submit"
    >
      <ElFormItem label="任务标题" prop="title"
        ><ElInput
          v-model="form.title"
          placeholder="例如：完成首页交互设计"
          maxlength="80"
          show-word-limit
      /></ElFormItem>
      <div class="grid grid-cols-2 gap-4">
        <ElFormItem label="任务状态" prop="status"
          ><ElSelect v-model="form.status" class="w-full"
            ><ElOption
              v-for="item in statuses"
              :key="item.value"
              :label="item.label"
              :value="item.value" /></ElSelect
        ></ElFormItem>
        <ElFormItem label="优先级" prop="priority"
          ><ElSelect v-model="form.priority" class="w-full"
            ><ElOption
              v-for="item in priorities"
              :key="item.value"
              :label="item.label"
              :value="item.value" /></ElSelect
        ></ElFormItem>
      </div>
      <ElFormItem label="负责人" prop="assignee"
        ><ElInput
          v-model="form.assignee"
          placeholder="填写姓名，留空表示待分配"
          maxlength="30"
          clearable
      /></ElFormItem>
      <ElFormItem label="截止日期" prop="dueDate"
        ><ElDatePicker
          v-model="form.dueDate"
          type="date"
          value-format="YYYY-MM-DD"
          placeholder="选择截止日期（选填）"
          clearable
          style="width: 100%"
      /></ElFormItem>
      <ElFormItem label="任务描述" prop="description"
        ><ElInput
          v-model="form.description"
          type="textarea"
          :rows="7"
          placeholder="描述任务目标、交付内容或验收标准"
          maxlength="2000"
          show-word-limit
      /></ElFormItem>
    </ElForm>
    <template #footer
      ><ElButton :disabled="submitting" @click="close">取消</ElButton
      ><ElButton
        type="primary"
        :loading="submitting"
        :disabled="submitting"
        @click="submit"
        >{{ submitting ? '正在保存' : '保存任务' }}</ElButton
      ></template
    >
  </ElDrawer>
</template>

<style>
.task-drawer {
  max-width: 100vw;
}
</style>
