import { defineOverridesPreferences } from '@vben/preferences';

/**
 * @description 项目配置文件
 * 只需要覆盖项目中的一部分配置，不需要的配置不用覆盖，会自动使用默认配置
 * !!! 更改配置后请清空缓存，否则可能不生效
 */
export const overridesPreferences = defineOverridesPreferences({
  // overrides
  app: {
    colorGrayMode: false,
    defaultHomePath: '/products',
    name: import.meta.env.VITE_APP_TITLE,
    // 是否显示复制偏好设置按钮
    enablePreferences: false,
    enableCheckUpdates: false,
    enableCopyPreferences: false,
  },
  logo: { source: '/logo.svg' },
  copyright: { enable: false },
  tabbar: { keepAlive: false },
  theme: {
    mode: 'light',
    radius: '0.25',
  },
  transition: {
    enable: true,
    loading: false,
    name: 'fade-slide',
    progress: true,
  },
  widget: {
    languageToggle: false,
    lockScreen: false,
    timezone: false,
    notification: false,
  },
});
