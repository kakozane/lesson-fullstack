import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';

const STORAGE_PREFIX = 'list-page-size';

function resolvePageKey(pageKey?: string): string {
  if (pageKey) return pageKey;
  const route = useRoute();
  const name = route.name;
  if (typeof name === 'string' && name) return name;
  return route.path;
}

function readStoredSize(pageKey: string, fallback: number): number {
  try {
    const raw = localStorage.getItem(`${STORAGE_PREFIX}:${pageKey}`);
    const size = Number(raw);
    return Number.isFinite(size) && size > 0 ? size : fallback;
  } catch {
    return fallback;
  }
}

/**
 * 按页面记忆列表 pageSize。
 * 默认用当前路由 name 作为 key；同一页有多处分页时再传入 pageKey。
 * 用法：const pageSize = usePageSize();
 */
export function usePageSize(pageKey?: string, defaultSize = 10) {
  const key = resolvePageKey(pageKey);
  const pageSize = ref(readStoredSize(key, defaultSize));

  watch(pageSize, (size) => {
    try {
      localStorage.setItem(`${STORAGE_PREFIX}:${key}`, String(size));
    } catch {
      // 忽略配额不足 / 隐私模式等写入失败
    }
  });

  return pageSize;
}
