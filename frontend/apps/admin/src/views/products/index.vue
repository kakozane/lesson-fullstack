<script setup lang="ts">
import type { Category, Product } from '#/api/products';
import { onMounted, ref } from 'vue';
import { PageHeader } from '@vben/common';
import {
  ElAlert,
  ElButton,
  ElCard,
  ElForm,
  ElFormItem,
  ElInput,
  ElMessage,
  ElMessageBox,
  ElOption,
  ElPagination,
  ElSelect,
  ElTable,
  ElTableColumn,
  ElTag,
} from 'element-plus';
import { updateProductStatus, deleteProduct, getCategories } from '#/api/products';
import { useProductList } from '#/composables/use-product-list';
import ProductForm from './product-form.vue';

const {
  filters,
  query,
  items,
  total,
  loading,
  failed,
  load,
  search,
  reset,
  changePage,
  changeSize,
} = useProductList();
const categories = ref<Category[]>([]);
const categoriesFailed = ref(false);
const dialogVisible = ref(false);
const editing = ref<Product>();
const operating = ref(false);
async function loadCategories() {
  categoriesFailed.value = false;
  try {
    categories.value = await getCategories();
  } catch {
    categoriesFailed.value = true;
  }
}
onMounted(() => {
  void load();
  void loadCategories();
});
function openForm(product?: Product) {
  editing.value = product;
  dialogVisible.value = true;
}
async function saved() {
  await load();
}
async function changeStatus(status: 0 | 1, row: Product) {
  if (operating.value || loading.value) return;
  operating.value = true;
  try {
    await ElMessageBox.confirm(
      `确定${status ? '上架' : '下架'}商品「${row.name}」吗？`,
      '确认修改状态',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' },
    );
    await updateProductStatus(row.id, status);
    ElMessage.success('商品状态已更新');
    await saved();
  } catch {
    /* 用户取消或请求层已提示错误。 */
  } finally {
    operating.value = false;
  }
}
async function remove(row: Product) {
  if (operating.value || loading.value) return;
  operating.value = true;
  try {
    await ElMessageBox.confirm(
      `确定删除「${row.name}」吗？删除后无法恢复。`,
      '删除商品',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    );
    await deleteProduct(row.id);
    ElMessage.success('商品已删除');
    await saved();
  } catch {
    /* 用户取消或请求层已提示错误。 */
  } finally {
    operating.value = false;
  }
}

</script>

<template>
  <div class="product-page p-4">
    <PageHeader title="商品管理" subtitle="管理商品信息、库存与销售状态" />
    <ElCard shadow="never" class="mb-4">
      <ElForm
        :inline="true"
        :disabled="operating"
        class="filter-form"
        @submit.prevent="search"
      >
        <ElFormItem label="商品搜索"
          ><ElInput
            v-model="filters.keyword"
            clearable
            maxlength="80"
            placeholder="搜索商品名称 / 编码"
            style="width: 240px"
            @clear="search"
        /></ElFormItem>
        <ElFormItem label="商品分类"
          ><ElSelect
            v-model="filters.categoryId"
            clearable
            :value-on-clear="() => undefined"
            placeholder="全部分类"
            style="width: 160px"
            @change="search"
            ><ElOption
              v-for="item in categories"
              :key="item.id"
              :value="item.id"
              :label="item.name" /></ElSelect
        ></ElFormItem>
        <ElFormItem label="销售状态"
          ><ElSelect
            v-model="filters.status"
            clearable
            :value-on-clear="''"
            placeholder="全部状态"
            style="width: 140px"
            @change="search"
            ><ElOption label="上架中" :value="1" /><ElOption
              label="已下架"
              :value="0" /></ElSelect
        ></ElFormItem>
        <ElFormItem
          ><ElButton type="primary" native-type="submit">查询</ElButton
          ><ElButton @click="reset">重置</ElButton></ElFormItem
        >
      </ElForm>
    </ElCard>
    <ElAlert
      v-if="categoriesFailed"
      type="warning"
      :closable="false"
      class="mb-4"
      ><template #title
        >分类加载失败
        <ElButton link type="primary" @click="loadCategories"
          >重新加载</ElButton
        ></template
      ></ElAlert
    >
    <ElCard shadow="never">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <span class="text-base font-semibold">商品列表</span
          ><span class="text-muted-foreground ml-3 text-sm"
            >共 {{ total }} 件商品</span
          >
        </div>
        <div class="flex flex-wrap gap-2">
          <ElButton :disabled="operating || loading" @click="load"
            >刷新列表</ElButton
          ><ElButton
            type="primary"
            :disabled="operating || !categories.length"
            @click="openForm()"
            >新增商品</ElButton
          >
        </div>
      </div>
      <ElAlert
        v-if="failed"
        title="商品加载失败，请重试"
        type="error"
        show-icon
        :closable="false"
        class="mb-4"
        ><ElButton link type="primary" @click="load"
          >重新加载</ElButton
        ></ElAlert
      >
      <ElTable
        v-loading="loading"
        :data="items"
        row-key="id"
        stripe
      >
        <ElTableColumn label="商品信息" min-width="240"
          ><template #default="{ row }"
            ><div class="py-1">
              <div class="font-medium">{{ row.name }}</div>
              <div class="text-muted-foreground mt-1 text-xs">
                {{ row.sku }}
              </div>
            </div></template
          ></ElTableColumn
        >
        <ElTableColumn prop="categoryName" label="分类" min-width="110" />
        <ElTableColumn label="售价" min-width="110" align="right"
          ><template #default="{ row }"
            ><span class="font-medium"
              >¥ {{ (row.priceCents / 100).toFixed(2) }}</span
            ></template
          ></ElTableColumn
        >
        <ElTableColumn label="库存" width="100" align="right"
          ><template #default="{ row }"
            ><span :class="row.stock < 20 ? 'text-orange-500' : ''">{{
              row.stock
            }}</span></template
          >
          </ElTableColumn
        >
        <ElTableColumn label="状态" width="110" align="center"
          ><template #default="{ row }"
            ><ElTag
              :type="row.status === 1 ? 'success' : 'info'"
              effect="light"
              >{{ row.status === 1 ? '上架中' : '已下架' }}</ElTag
            ></template
          ></ElTableColumn
        >
        <ElTableColumn prop="updatedAt" label="更新时间" width="180" />
        <ElTableColumn label="操作" width="180" fixed="right"
          ><template #default="{ row }"
            ><ElButton
              link
              type="primary"
              :disabled="operating || !categories.length"
              @click="openForm(row)"
              >编辑</ElButton
            ><ElButton
              link
              type="primary"
              :disabled="operating"
              @click="changeStatus(row.status === 1 ? 0 : 1, row)"
              >{{ row.status === 1 ? '下架' : '上架' }}</ElButton
            ><ElButton
              link
              type="danger"
              :disabled="operating"
              @click="remove(row)"
              >删除</ElButton
            ></template
          ></ElTableColumn
        >
        <template #empty
          ><div class="py-10 text-sm">
            {{
              loading
                ? '正在加载商品…'
                : failed
                  ? '加载失败，请点击重新加载'
                  : '暂无符合条件的商品，请调整筛选或新增商品'
            }}
          </div></template
        >
      </ElTable>
      <div class="pagination-wrap mt-5 flex justify-end">
        <ElPagination
          :current-page="query.page"
          :page-size="query.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          :disabled="operating || loading"
          background
          layout="total, sizes, prev, pager, next"
          @size-change="changeSize"
          @current-change="changePage"
        />
      </div>
    </ElCard>
    <ProductForm
      v-model="dialogVisible"
      :product="editing"
      :categories="categories"
      @saved="saved"
    />
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
