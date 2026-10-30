<script setup lang="ts">
import type { FormRules } from 'element-plus';

import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElSelect,
  ElSwitch,
} from 'element-plus';

// 校验规则已准备，实战时接入 model、prop 和提交校验。
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
  categoryId: [
    {
      type: 'integer',
      min: 1,
      required: true,
      message: '请选择分类',
      trigger: 'change',
    },
  ],
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
</script>

<template>
  <!-- 实战接入：弹窗开关、表单草稿、校验、回填与提交。仅预览样式时可暂将 false 改为 true。 -->
  <ElDialog
    :model-value="false"
    title="新增商品"
    width="600px"
    class="product-dialog"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <ElForm :rules="rules" label-width="90px">
      <ElFormItem label="商品名称">
        <ElInput
          maxlength="80"
          show-word-limit
          placeholder="例如：无线蓝牙鼠标"
        />
      </ElFormItem>
      <ElFormItem label="商品编码">
        <ElInput maxlength="32" placeholder="例如：CLASS-001" />
      </ElFormItem>
      <ElFormItem label="商品分类">
        <ElSelect placeholder="请选择分类" class="w-full" />
      </ElFormItem>
      <ElFormItem label="销售价格">
        <ElInputNumber :precision="2" /><span class="text-muted-foreground ml-2"
          >元</span
        >
      </ElFormItem>
      <ElFormItem label="商品库存">
        <ElInputNumber /><span class="text-muted-foreground ml-2">件</span>
      </ElFormItem>
      <ElFormItem label="上架状态"><ElSwitch /></ElFormItem>
      <ElFormItem label="商品描述">
        <ElInput type="textarea" :rows="3" maxlength="500" show-word-limit />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton>取消</ElButton>
      <ElButton type="primary">保存商品</ElButton>
    </template>
  </ElDialog>
</template>
