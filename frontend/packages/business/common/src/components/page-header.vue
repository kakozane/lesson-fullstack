<script setup lang="ts">
/**
 * 跨端共享组件样例
 * 用法：import { PageHeader } from '@vben/common';
 * 两端（admin / teacher）均可直接使用。
 */
defineProps<{
  subtitle?: string;
  title: string;
}>();
</script>

<template>
  <div class="mb-4">
    <h1 class="text-xl font-semibold">{{ title }}</h1>
    <p v-if="subtitle" class="mt-1 text-sm text-gray-500">{{ subtitle }}</p>
  </div>
</template>
