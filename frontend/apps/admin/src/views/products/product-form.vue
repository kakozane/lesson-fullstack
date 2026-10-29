<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus';
import type { Product, ProductInput } from '#/api/products';
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElMessageBox,
  ElOption,
  ElSelect,
  ElSwitch,
} from 'element-plus';
import { createProduct, updateProduct } from '#/api/products';

const props = defineProps<{
  modelValue: boolean;
  product?: Product;
  categories: string[];
}>();
const emit = defineEmits<{ 'update:modelValue': [boolean]; saved: [] }>();
const formRef = ref<FormInstance>();
const saving = ref(false);
const form = reactive({
  name: '',
  sku: '',
  category: '',
  price: 0.01,
  stock: 0,
  status: 0 as 0 | 1,
  description: '',
});
let initial = '';
const dirty = computed(
  () => props.modelValue && JSON.stringify(form) !== initial,
);
const rules: FormRules = {
  name: [
    {
      required: true,
      whitespace: true,
      message: '请输入商品名称',
      trigger: 'blur',
    },
    { max: 80, message: '最多 80 个字符', trigger: 'blur' },
  ],
  sku: [
    {
      required: true,
      pattern: /^[\dA-Za-z-]{1,32}$/,
      message: '请输入 1～32 位字母、数字或短横线',
      trigger: 'blur',
    },
  ],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
  price: [
    {
      required: true,
      type: 'number',
      min: 0.01,
      max: 999999.99,
      message: '价格范围为 0.01～999999.99 元',
      trigger: 'change',
    },
  ],
  stock: [
    {
      required: true,
      type: 'integer',
      min: 0,
      max: 999999,
      message: '库存需为 0～999999 的整数',
      trigger: 'change',
    },
  ],
};
watch(
  () => props.modelValue,
  async (open) => {
    if (!open) return;
    const p = props.product;
    Object.assign(
      form,
      p
        ? {
            name: p.name,
            sku: p.sku,
            category: p.category,
            price: p.priceCents / 100,
            stock: p.stock,
            status: p.status,
            description: p.description,
          }
        : {
            name: '',
            sku: '',
            category: '',
            price: 0.01,
            stock: 0,
            status: 0,
            description: '',
          },
    );
    initial = JSON.stringify(form);
    await nextTick();
    formRef.value?.clearValidate();
  },
);
async function canLeave() {
  if (saving.value) {
    ElMessage.warning('正在保存，请稍候');
    return false;
  }
  if (!dirty.value) return true;
  try {
    await ElMessageBox.confirm('尚有未保存的修改，确定放弃吗？', '放弃修改', {
      type: 'warning',
      confirmButtonText: '放弃修改',
      cancelButtonText: '继续编辑',
    });
    return true;
  } catch {
    return false;
  }
}
async function close() {
  if (await canLeave()) emit('update:modelValue', false);
}
onBeforeRouteLeave(canLeave);
function beforeUnload(event: BeforeUnloadEvent) {
  if (dirty.value || saving.value) {
    event.preventDefault();
    event.returnValue = '';
  }
}
window.addEventListener('beforeunload', beforeUnload);
onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload));
async function submit() {
  if (saving.value) return;
  saving.value = true;
  try {
    if (!(await formRef.value?.validate().catch(() => false))) return;
    const data: ProductInput = {
      name: form.name.trim(),
      sku: form.sku.trim(),
      category: form.category,
      priceCents: Math.round(form.price * 100),
      stock: form.stock,
      status: form.status,
      description: form.description.trim(),
    };
    if (props.product) await updateProduct(props.product.id, data);
    else await createProduct(data);
    ElMessage.success(props.product ? '商品已更新' : '商品已创建');
    emit('update:modelValue', false);
    emit('saved');
  } catch {
    // 保存失败时保留输入，由请求层显示错误信息。
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <ElDialog
    :model-value="modelValue"
    :title="product ? '编辑商品' : '新增商品'"
    width="600px"
    class="product-dialog"
    :before-close="close"
    :close-on-click-modal="false"
    :close-on-press-escape="!saving"
    :show-close="!saving"
    destroy-on-close
  >
    <ElForm
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="90px"
      :disabled="saving"
      @submit.prevent="submit"
    >
      <ElFormItem label="商品名称" prop="name"
        ><ElInput
          v-model="form.name"
          maxlength="80"
          show-word-limit
          placeholder="例如：无线蓝牙鼠标"
      /></ElFormItem>
      <ElFormItem label="商品编码" prop="sku"
        ><ElInput
          v-model="form.sku"
          maxlength="32"
          placeholder="例如：SP-0037，编码不可重复"
      /></ElFormItem>
      <ElFormItem label="商品分类" prop="category"
        ><ElSelect
          v-model="form.category"
          placeholder="请选择分类"
          class="w-full"
          ><ElOption
            v-for="item in categories"
            :key="item"
            :label="item"
            :value="item" /></ElSelect
      ></ElFormItem>
      <ElFormItem label="销售价格" prop="price"
        ><ElInputNumber
          v-model="form.price"
          :min="0.01"
          :max="999999.99"
          :precision="2"
          :step="1"
          controls-position="right"
          aria-label="销售价格"
        /><span class="text-muted-foreground ml-3">元</span></ElFormItem
      >
      <ElFormItem label="商品库存" prop="stock"
        ><ElInputNumber
          v-model="form.stock"
          :min="0"
          :max="999999"
          :precision="0"
          controls-position="right"
          aria-label="商品库存"
        /><span class="text-muted-foreground ml-3">件</span></ElFormItem
      >
      <ElFormItem label="上架状态"
        ><ElSwitch
          v-model="form.status"
          :active-value="1"
          :inactive-value="0"
          active-text="上架"
          inactive-text="下架"
      /></ElFormItem>
      <ElFormItem label="商品描述"
        ><ElInput
          v-model="form.description"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          placeholder="填写商品介绍（选填）"
      /></ElFormItem>
    </ElForm>
    <template #footer
      ><ElButton :disabled="saving" @click="close">取消</ElButton
      ><ElButton type="primary" :loading="saving" @click="submit">{{
        saving ? '正在保存' : '保存商品'
      }}</ElButton></template
    >
  </ElDialog>
</template>

<style scoped>
:global(.product-dialog) {
  max-width: calc(100vw - 32px);
}
</style>
