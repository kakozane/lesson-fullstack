<script setup lang="ts">
import { computed } from 'vue';
import { AuthenticationLoginExpiredModal } from '@vben/common-ui';
import { BasicLayout, UserDropdown } from '@vben/layouts';
import { useAccessStore, useUserStore } from '@vben/stores';
import { useAuthStore } from '#/store';
import LoginForm from '#/views/_core/authentication/login.vue';
const accessStore = useAccessStore();
const userStore = useUserStore();
const authStore = useAuthStore();
const avatar = computed(() => userStore.userInfo?.avatar || '/logo.svg');
const logout = () => authStore.logout(false);
</script>
<template>
  <BasicLayout @clear-preferences-and-logout="logout">
    <template #user-dropdown
      ><UserDropdown
        :avatar="avatar"
        :menus="[]"
        :text="userStore.userInfo?.realName"
        description="课堂实战 · 初始版"
        tag-text="管理员"
        @logout="logout"
        @clear-preferences-and-logout="logout"
    /></template>
    <template #extra
      ><AuthenticationLoginExpiredModal
        v-model:open="accessStore.loginExpired"
        :avatar="avatar"
        ><LoginForm /></AuthenticationLoginExpiredModal
    ></template>
  </BasicLayout>
</template>
