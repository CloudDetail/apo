# Routing Overview

## Scope
本文档整合当前前端主路由装配方式、动态菜单与权限耦合方式，以及路由层面的主要漂移风险。

## Frontend Routing Assembly

| Layer | Source Files | Description |
|---|---|---|
| Root router | `frontend/src/App.js` | 使用 `HashRouter`，显式注册 `/login`，其余路径进入 `DefaultLayout`。 |
| Base routes | `frontend/src/routes.js` | 系统管理、接入中心、数据范围等基础路由。 |
| OSS routes | `frontend/src/oss/routes.js` | service、logs、trace、dashboard、alerts、workflows。 |
| Pro routes | `frontend/src/pro/routes.js` | alert analyze、cause report、report、dataplane、platform manage、agent。 |
| Runtime route render | `frontend/src/core/components/AppContent.js` | 根据 `src/routes.js` 提供的本地 route 表渲染页面组件。 |
| Menu source | `frontend/src/core/contexts/UserContext.js` + `/api/permission/config` | 后端返回 `menuItem` 与 `routes`，侧边栏和默认跳转依赖这组数据。 |

## Route Resolution Model

```text
App.js
  -> /login -> Login page
  -> * -> DefaultLayout
       -> AppSidebar reads backend-driven menuItems
       -> AppContent renders local routes[]
       -> "/" redirects to first backend-provided menu router
```

这意味着当前系统的页面可达性并不是纯前端静态配置，而是三部分同时决定：

1. `src/routes.js` 是否本地注册了对应 route。
2. `/api/permission/config` 是否返回对应菜单和路由。
3. 中间件、license、数据范围和外部代理是否允许后续 API 成功。

## Route Families

| Route Family | Representative Paths | Primary Module |
|---|---|---|
| Auth and user center | `/login`, `/user`, `/system/user-manage`, `/system/menu-manage`, `/system/role-manage`, `/system/team` | `identity-access-and-organization` |
| Data governance | `/system/data-group` | `data-scope-governance` |
| Integration and onboarding | `/integration/alerts`, `/integration/data`, `/integration/data/settings`, `/probe-management`, `/probe-management/settings` | `alert-sources-and-enrichment`, `data-ingestion-and-oneagent` |
| Service and diagnostics | `/service`, `/service/info`, `/logs/*`, `/trace/*`, `/alert-analyze*`, `/cause/*`, `/report`, `/agent` | `service-observability`, `logs-exploration-and-parsing`, `trace-exploration-and-profiling`, `automated-diagnosis-and-workflows`, `conversational-diagnosis-and-reporting` |
| Dashboards and alerts | `/system-dashboard`, `/basic-dashboard`, `/application-dashboard`, `/middleware-dashboard`, `/mysql-dashboard`, `/healthy-service`, `/alerts/*`, `/workflows` | `dashboards-and-health-monitoring`, `alert-rules-and-incident-events`, `automated-diagnosis-and-workflows` |
| Platform governance | `/config`, `/system/platform-manage`, `/dataplane` | `platform-administration`, `dataplane-and-topology-governance` |

## Dynamic Navigation And Permission Notes

| Observation | Evidence | Documentation Impact |
|---|---|---|
| 菜单来源于后端，不是本地常量 | `UserContext.js` 调 `/api/permission/config`，`AppSidebar.js` 渲染返回结果 | 路由文档必须同时写本地 route 和后端 permission metadata。 |
| 默认首页由后端菜单第一项决定 | `AppContent.js` 的 `getDefaultTo()` | 不能把 `/` 简单写成固定首页。 |
| `AuthRouter` 客户端权限校验目前未生效 | `frontend/src/core/components/AuthRouter.tsx` 中 `useEffect` 被整段注释 | 当前更多依赖后端权限配置和 API 中间件，而不是前端阻断。 |

## Build-Time Route Divergence
`frontend/src/routes.js` 当前存在构建分支：

- `VITE_APP_VERSION === 'pro'` 时，只拼接 `baseRoutes + proRoutes`
- 否则拼接 `baseRoutes + ossRoutes + proRoutes`

这意味着：

- 当前代码层面并不能直接推断“pro 构建一定包含 OSS 页面”。
- 但已生成的模块文档与后端权限元数据又默认整个系统存在完整观察域。
- 因此这部分需要在后续版本核对真实构建参数和产品发布方式。

## Route Drift Notes

| Drift Or Mismatch | Observed Evidence | Current Interpretation |
|---|---|---|
| `/dataplane` 与权限元数据中的 `/agent-dataplane` 并存 | 已有模块文档与后端初始化数据扫描结果 | 很可能存在历史命名漂移，业务能力仍归属 `dataplane-and-topology-governance`。 |
| `/probe-management` 路由名称指向 ProbeInstaller，但页面组件实际复用 `IntegrationCenter/DataIntegration` | `frontend/src/pro/routes.js` | 当前应把它视为接入中心的别名入口，而不是独立业务模块。 |
| `/system/platform-manage` 未经过 `AuthRouter` 包装 | `frontend/src/pro/routes.js` | 由于 `AuthRouter` 当前本身近似空壳，影响有限，但仍属于路由层不一致。 |
| `/demo/service-config` 指向 `frontend/src/core/views/Config/index.tsx`，而真正平台配置页面是 `/config` -> `frontend/src/oss/views/config/index.jsx` | `frontend/src/routes.js` 与 `frontend/src/oss/routes.js` | 当前应把 `/demo/service-config` 标为非主链入口或历史页面。 |

## Canonical Reading Path
如果要继续追踪路由到模块和 API 的落点，请按以下顺序阅读：

1. [route_to_module_map.md](route_to_module_map.md)
2. [page_inventory.md](page_inventory.md)
3. [frontend_backend_mapping.md](frontend_backend_mapping.md)
