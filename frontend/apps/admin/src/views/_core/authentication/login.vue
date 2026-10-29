<script setup lang="ts">
import type { VbenFormSchema } from '@vben/common-ui';
import { AuthenticationLogin, z } from '@vben/common-ui';
import { useAuthStore } from '#/store';
const authStore = useAuthStore();
const formSchema: VbenFormSchema[] = [
  {
    component: 'VbenInput',
    componentProps: { placeholder: '请输入账号' },
    fieldName: 'username',
    label: '账号',
    rules: z.string().min(1, '请输入账号'),
  },
  {
    component: 'VbenInputPassword',
    componentProps: { placeholder: '请输入密码' },
    fieldName: 'password',
    label: '密码',
    rules: z.string().min(1, '请输入密码'),
  },
];
async function login(values: Record<string, any>) {
  if (authStore.loginLoading) return;
  try {
    await authStore.authLogin({
      username: values.username,
      password: values.password,
    });
  } catch {
    /* 请求层展示登录错误。 */
  }
}
</script>
<template>
  <AuthenticationLogin
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    title="欢迎回来"
    sub-title="登录商品管理工作台"
    :show-forget-password="false"
    :show-remember-me="false"
    @submit="login"
  />
</template>
