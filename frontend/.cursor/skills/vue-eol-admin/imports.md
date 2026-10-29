# Import 路径别名

| 别名          | 实际路径                 | 用途             |
| ------------- | ------------------------ | ---------------- |
| `#/`          | `apps/web/src/`          | 当前应用内部引用 |
| `@vben/`      | `packages/` 下的各个包   | 框架公共能力     |
| `@vben-core/` | `packages/@core/` 下的包 | 核心底层能力     |

```typescript
// 正确
import { getUserListApi } from '#/api/system/user';
import { useAccess } from '@vben/access';

// 错误：跨层级相对路径
import { getUserListApi } from '../../../api/system/user';
```

## 规则

- 应用内模块一律用 `#/` 别名
- 框架能力用 `@vben/` 或 `@vben-core/`
- 同目录内可用 `./` 相对路径
- 类型导入用 `import type`
