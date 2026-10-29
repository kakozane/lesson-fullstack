import type {
  RouteLocationNormalizedLoaded,
  RouteMeta,
  RouteRecordName,
} from 'vue-router';

import { ref } from 'vue';
import { useRouter } from 'vue-router';

const cachedViews = ref<string[]>([]);

function isCacheName(
  name: RouteRecordName | null | undefined,
): name is string {
  return typeof name === 'string';
}

export function useKeepAlive() {
  const router = useRouter();

  /**
   * 添加缓存
   */
  function addCache(name: RouteRecordName | null | undefined) {
    if (!isCacheName(name) || cachedViews.value.includes(name)) {
      return;
    }
    cachedViews.value = [...cachedViews.value, name];
  }

  /**
   * 移除缓存
   */
  function removeCache(name: RouteRecordName | null | undefined) {
    if (!isCacheName(name)) {
      return;
    }
    cachedViews.value = cachedViews.value.filter((n) => n !== name);
  }

  /**
   * 判断「离开当前页去 targetName」时是否应保留缓存
   * 优先级：同路由换参 > keepAliveFrom（白名单）> keepAliveExclude（黑名单）> 默认保留
   */
  function shouldKeepCache(
    routeMeta: RouteMeta,
    targetName: RouteRecordName | null | undefined,
    fromName: RouteRecordName | null | undefined,
  ) {
    const { keepAliveFrom, keepAliveExclude } = routeMeta;

    // 同路由不同参数（如 Orders tab 切换 ?type=all → ?type=pending）→ 保留缓存
    if (targetName === fromName) {
      return true;
    }

    // 白名单模式：只有去指定页面时才缓存（如 List 仅去 Detail 时保留）
    if (keepAliveFrom && keepAliveFrom.length > 0) {
      return isCacheName(targetName) && keepAliveFrom.includes(targetName);
    }

    // 黑名单模式：去指定页面时清除，其余情况保留
    if (keepAliveExclude && keepAliveExclude.length > 0) {
      return !isCacheName(targetName) || !keepAliveExclude.includes(targetName);
    }

    // 未配置白/黑名单 → 离开即缓存
    return true;
  }

  /**
   * 注册全局路由守卫，动态维护 cachedViews（驱动 App.vue 的 keep-alive :include）
   *
   * 离开缓存页：根据目标页决定 addCache / removeCache
   * 进入缓存页：若配置了 keepAliveFrom，来源不在白名单则 removeCache（防止复用旧实例）
   */
  function setupRouteGuard() {
    router.beforeEach((to, from) => {
      // --- 离开缓存页面时：决定是否把 from 加入 include 列表 ---
      if (from.meta.keepAlive) {
        if (shouldKeepCache(from.meta, to.name, from.name)) {
          addCache(from.name);
        } else {
          removeCache(from.name);
        }
      }

      // --- 进入缓存页面时：非白名单来源进入则强制清缓存 ---
      // 同路由换参（tab 切换）跳过；例：Home → List 需销毁旧实例
      if (to.meta.keepAlive) {
        const { keepAliveFrom } = to.meta;
        if (keepAliveFrom && keepAliveFrom.length > 0) {
          if (
            to.name !== from.name &&
            isCacheName(from.name) &&
            !keepAliveFrom.includes(from.name)
          ) {
            removeCache(to.name);
          }
        }
      }
    });
  }

  /**
   * 生成 KeepAlive component 的 key
   * - keepAliveKey: 'fullPath' → 同路由不同参数各自独立缓存
   * - 默认 → 不设 key，按组件 name 缓存（一个组件一份实例）
   */
  function getCacheKey(routeObj: RouteLocationNormalizedLoaded) {
    if (routeObj.meta.keepAliveKey === 'fullPath') {
      return routeObj.fullPath;
    }
    return undefined;
  }

  return {
    cachedViews,
    addCache,
    removeCache,
    setupRouteGuard,
    getCacheKey,
  };
}
