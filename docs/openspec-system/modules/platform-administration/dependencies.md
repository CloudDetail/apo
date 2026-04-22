# Platform Administration Dependencies

## Upstream Dependencies

| Dependency | Evidence | Why It Matters |
|---|---|---|
| `request.js` | `core/api/config.js`, `core/api/app.js`, `pro/api/license.ts` | 统一认证请求。 |
| `LicenseContext` | `DefaultLayout.js`, `LicenseManage/index.tsx` | 全局平台 license 状态。 |
| `DeployModeContext` | `Login.jsx`, `DefaultLayout.js` | 全局品牌资源选择。 |
| ClickHouse Repo | `services/config` | TTL 能力基础。 |
| Prometheus Repo | `services/license` | 计算节点用量。 |

## Downstream Dependencies

| Consumer | Evidence | Dependency Relationship |
|---|---|---|
| 登录页 | `Login.jsx` | 使用 `DeployModeContext` 渲染品牌。 |
| 默认布局 | `DefaultLayout.js` | 使用 `LicenseContext` 决定 banner。 |
| 其他页面 | 整个 app shell | 通过全局布局间接受到 license 状态影响。 |

## Shared Components

| Shared Component | File | Role In This Module |
|---|---|---|
| `BasicCard` | `ConfigTTL`, `LicenseInfoCard`, `PlatformManage` | 平台管理 UI 容器。 |
| `notify` | 多个页面 | 上传、错误、刷新提示。 |

## Shared Utilities

| Utility | File | Role |
|---|---|---|
| `prepareTTLInfo` | `service_setttl.go` | 从 ClickHouse `CREATE TABLE` 语句中抽取 TTL。 |
| `sanitizeSVG` | `services/app/logo.go` | SVG 安全清洗。 |
| `checkLoop` | `services/license` | 后台持续检查 license 状态。 |

## External Services And Runtime Inputs

| Service | Evidence | Current Role |
|---|---|---|
| ClickHouse | `services/config` | TTL 读写目标。 |
| Prometheus | `services/license` | seat 使用量统计。 |
| `PRODUCT_MODE` 运行时环境变量 | `backend/pkg/api/app/func_getdeploymode.go` | `/api/app/deployMode` 直接读取该值，用来驱动 APO / Syncause 品牌切换。 |

## Notes

- 本模块本质上是平台级横切能力，因此它的状态会影响登录页、布局和管理员页面，不只是单一配置页。
- 当前实际闭环能力包括 TTL、license 和 deploy mode 读取；branding 可视化配置 UI 仍未开放。
