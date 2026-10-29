<script setup lang="ts">
import { PageHeader } from '@vben/common';
import {
  ElButton,
  ElCard,
  ElForm,
  ElFormItem,
  ElInput,
  ElPagination,
  ElSelect,
  ElTable,
  ElTableColumn,
  ElTag,
} from 'element-plus';

import ProductForm from './product-form.vue';
</script>

<template>
  <div class="product-page p-4">
    <PageHeader title="商品管理" subtitle="管理商品信息、库存与销售状态" />
    <ElCard shadow="never" class="mb-4">
      <!-- 实战接入：筛选值、分类选项、查询与重置事件。 -->
      <ElForm :inline="true" class="filter-form">
        <ElFormItem label="商品搜索">
          <ElInput
            clearable
            maxlength="80"
            placeholder="搜索商品名称 / 编码"
            style="width: 240px"
          />
        </ElFormItem>
        <ElFormItem label="商品分类">
          <ElSelect clearable placeholder="全部分类" style="width: 160px" />
        </ElFormItem>
        <ElFormItem label="销售状态">
          <ElSelect clearable placeholder="全部状态" style="width: 140px" />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary">查询</ElButton>
          <ElButton>重置</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
    <ElCard shadow="never">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <span class="text-base font-semibold">商品列表</span>
          <span class="text-muted-foreground ml-3 text-sm">共 — 件商品</span>
        </div>
        <div class="flex flex-wrap gap-2">
          <ElButton>刷新列表</ElButton>
          <ElButton type="primary">新增商品</ElButton>
        </div>
      </div>
      <div
        class="bg-accent mb-4 flex flex-wrap items-center gap-3 rounded-md px-4 py-3"
      >
        <span class="text-muted-foreground text-sm">已选 — 件（仅当前页）</span>
        <ElButton size="small" disabled>批量上架</ElButton>
        <ElButton size="small" disabled>批量下架</ElButton>
      </div>
      <!-- 实战接入：列表数据、加载状态、行字段和操作事件。 -->
      <ElTable row-key="id" stripe>
        <ElTableColumn type="selection" width="48" />
        <ElTableColumn label="商品信息" min-width="240">
          <template #default>
            <div class="py-1">
              <div class="font-medium">—</div>
              <div class="text-muted-foreground mt-1 text-xs">—</div>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="分类" min-width="110" />
        <ElTableColumn label="售价" min-width="110" align="right">
          <template #default><span class="font-medium">—</span></template>
        </ElTableColumn>
        <ElTableColumn label="库存" width="100" align="right" />
        <ElTableColumn label="状态" width="110" align="center">
          <template #default
            ><ElTag type="info" effect="light">—</ElTag></template
          >
        </ElTableColumn>
        <ElTableColumn label="更新时间" width="180" />
        <ElTableColumn label="操作" width="180" fixed="right">
          <template #default>
            <ElButton link type="primary">编辑</ElButton>
            <ElButton link type="primary">上下架</ElButton>
            <ElButton link type="danger">删除</ElButton>
          </template>
        </ElTableColumn>
        <template #empty
          ><div class="py-10 text-sm">暂无商品数据</div></template
        >
      </ElTable>
      <div class="pagination-wrap mt-5 flex justify-end">
        <ElPagination background layout="total, sizes, prev, pager, next" />
      </div>
    </ElCard>
    <ProductForm />
  </div>
</template>

<style scoped>
.filter-form :deep(.el-form-item) {
  margin-bottom: 8px;
  margin-top: 8px;
}
.pagination-wrap {
  overflow-x: auto;
}
</style>
