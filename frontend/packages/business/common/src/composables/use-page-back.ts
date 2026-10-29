import { useRoute, useRouter } from 'vue-router';

import { useTabs } from '@vben/hooks';
import { getTabKey } from '@vben/stores';

/**
 * 详情页返回：优先跳到 backPath / meta.activePath，并关闭当前详情 tab。
 * 用法：const { handleBack } = usePageBack();
 */
export function usePageBack() {
  const route = useRoute();
  const router = useRouter();
  const { closeCurrentTab, closeTabByKey } = useTabs();

  async function handleBack(backPath?: string) {
    const detailTabKey = getTabKey(route);
    const targetPath =
      backPath ?? (route.meta.activePath as string | undefined);

    if (targetPath) {
      await router.push(targetPath);
      await closeTabByKey(detailTabKey);
      return;
    }

    await closeCurrentTab();
  }

  return {
    handleBack,
  };
}
