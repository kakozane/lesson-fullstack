/**
 * 跨端共享 composable 样例
 * 用法：const formatDate = useFormatDate(); formatDate(Date.now());
 */

export function useFormatDate() {
  return (value: Date | number | string) =>
    new Date(value).toLocaleDateString('zh-CN');
}
