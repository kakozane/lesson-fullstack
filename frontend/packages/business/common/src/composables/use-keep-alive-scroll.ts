import { nextTick, onActivated, onMounted } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';

/** 路由 scrollBehavior 和过渡会把滚动打回顶部，错开再写 */
const RESTORE_DELAYS = [0, 50, 300];

interface ScrollSnapshot {
  left: number;
  top: number;
}

/**
 * keep-alive 列表页滚动恢复。
 * 离开时记录 window 滚动位置；再次激活并完成数据请求后恢复。
 *
 * @example
 * useKeepAliveScroll(fetchList);
 */
export function useKeepAliveScroll(loader?: () => void | Promise<void>) {
  let snapshot: ScrollSnapshot = { left: 0, top: 0 };
  let hasSaved = false;

  function saveScroll() {
    snapshot = { left: window.scrollX, top: window.scrollY };
    hasSaved = true;
  }

  function applySaved() {
    window.scrollTo({
      left: snapshot.left,
      top: snapshot.top,
      behavior: 'instant',
    });
  }

  async function restoreScroll() {
    if (!hasSaved) return;
    await nextTick();
    await new Promise<void>((resolve) => {
      requestAnimationFrame(() => {
        requestAnimationFrame(() => resolve());
      });
    });
    for (const delay of RESTORE_DELAYS) {
      if (delay > 0) {
        await new Promise((resolve) => setTimeout(resolve, delay));
      }
      applySaved();
    }
  }

  onBeforeRouteLeave(() => {
    saveScroll();
  });

  onMounted(() => {
    void loader?.();
  });

  onActivated(() => {
    // 首次进入尚未离开过，不恢复
    if (!hasSaved) return;

    void (async () => {
      await loader?.();
      await restoreScroll();
    })();
  });

  return {
    saveScroll,
    restoreScroll,
  };
}
