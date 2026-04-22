# Architecture Notes

## Scope
本文档记录当前仓库在整编后仍然必须保留的系统级说明、边界决策和不一致项。它不替代模块文档，而是作为全局补充说明。

## Note 1: Navigation Is Backend-Driven
前端虽然在 `src/routes.js`、`src/oss/routes.js`、`src/pro/routes.js` 中声明了本地 route，但实际菜单结构、默认跳转路径和页面可见性还依赖后端 `/api/permission/config` 返回的 `menuItem` 与 `routes`。

这导致两个结果：

- 只读前端 route 文件无法完整推导可见功能。
- 只读后端 `validRouters` 初始化数据也无法完整推导页面组件实现。
- `PRODUCT_MODE` 还会影响 menu/router/feature 初始化结果，因此不同产品模式下可见导航树本身就不同。

因此，所有模块文档都必须同时保留“页面入口”和“权限元数据入口”两个视角。

## Note 2: License And Edition Gating Is Horizontal
`frontend/src/pro/contexts/LicenseContext.tsx`、`frontend/src/core/layout/DefaultLayout.js` 与 `backend/pkg/middleware/lic_middleware.go` 共同形成了 license 横切能力。

这意味着：

- `/pro/` 目录中的页面不能被简单写成“商业版天然可访问”。
- 某些 backend route group 即使前端存在页面，也会因为 `LicMiddleware` 拦截而表现为受限能力。

## Note 3: Data Scope And Time Range Are Shared Query Contexts
`DataGroupContext` 与 Redux 中的 time range reducer 会同时影响 service、logs、trace、dashboard、健康监控、告警分析等多个业务模块。

因此：

- DataGroup 不是单个系统管理页面的小功能，而是跨业务域治理能力。
- 时间范围不是某个页面内部状态，而是观测域共有上下文。

## Note 4: Backend Layering Is Stable, But Route Ownership Is Mixed
后端总体遵循：

```text
router
  -> api handler
  -> service
  -> repository
  -> database / observability backend / external service
```

但路由归属并不总是“一个 route group 等于一个模块”：

- `/api/alerts` 同时承载告警治理和自动化诊断接口。
- `/api/service` 同时承载 service detail 和健康/聚合观察相关接口。
- `/api/app` 当前包含 `/upload`、`/logo`、`/deployMode`，但 deploy mode 只是环境变量读取接口。

这也是为什么整编文档必须按业务能力而不是按 route group 直接切割。

## Note 5: External Boundary Must Stay Explicit
当前代码里以下接口不能被写成本仓库后端已完整实现：

| Boundary | Evidence | Why It Must Stay Explicit |
|---|---|---|
| `/ai/api/*` | `frontend/src/pro/api/chatAI.ts`, `conversation.ts`, `modelSetting.ts` | 当前仓库前端明确依赖，但 `backend/pkg/router/router_api.go` 未注册对应主路由。 |
| `/originx/api/*` | `frontend/src/pro/api/cause.ts` | 作为外部诊断查询边界更安全。 |
| `/dify/console/api/*` | `frontend/src/core/api/workflows.ts` | 当前仓库中主要表现为旁路登录/登出边界。 |
| `/alert/api/v1/*` | `frontend/src/core/api/alerts.js` 部分调用 | 与 `/api/alerts/*` 并存，当前主后端未见等价注册。 |

## Note 6: Shared Utilities Should Not Be Attached To Business Modules
以下源代码区域更适合作为共享基础设施或支撑层处理，而不是直接归入某个业务模块：

| Source Area | Recommended Ownership |
|---|---|
| `frontend/src/core/utils/request.js` | `frontend-state-request-and-i18n` |
| `backend/pkg/services/common/*` | 共享服务支撑层 |
| `backend/pkg/services/utils/filter.go` | 共享工具 |
| `backend/pkg/api/health` | 共享运行时支撑接口 |
| `backend/pkg/api/k8s` | 共享观测元数据接口 |
| `backend/pkg/api/metric` | 共享指标查询接口 |
| `backend/pkg/api/network` | 共享网络/拓扑支撑接口 |

## Note 7: Source Ownership Is Separate From Module Ownership
当前项目至少有两层需要同时记录的结构：

- 业务模块边界：谁负责什么业务能力。
- 源码归属边界：这部分代码来自上游共享基线、共享路径变体，还是商业仓独有实现。

当前已确认：

- 前端 `frontend/src/core`、`frontend/src/oss` 不能简单视为商业仓本地代码。
- 后端没有清晰的 `pro/` 目录，而是以共享包上的商业扩展和商业独有包的形式出现。

因此，模块文档如果只写“模块边界”，而不写“源码归属”，后续维护时就很容易把改动落在错误仓库或错误包层级。

## Note 8: Non-Commercial Paths Also Drift
共享路径与共享包并不等于“与上游完全一致”：

- 前端 `App.js`、`routes.js`、`AppSidebar.js`、`AuthRouter.tsx`、`request.js` 等共享层代码都已经漂移。
- `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx` 也存在与上游不同的实现。
- 后端 `router_api.go`、`middleware.go`、`alerts`、`dataplane`、`integration`、`serviceoverview` 等共享包同样已经被商业仓改写。

这意味着：

- “共享路径变体”和“Pro 专属功能”必须分开记录。
- 非商业路径代码同样需要在文档中标注它们是否已偏离上游。

## Inconsistency Log

| Area | Issue | Current Handling |
|---|---|---|
| Frontend route vs backend router metadata | 数据库初始化已切到 `/dataplane`、`/probe-management`，但权限白名单仍保留 `/agent-dataplane`、`probe-installer` 历史字符串 | 在统一文档中采用前端实际可见路径为主，同时注明剩余历史别名只存在于 allowlist。 |
| Frontend API vs backend registration | `/api/app/deployMode`、`/api/dataplane/capabilities` 已注册；真正仍有漂移的是 `/api/permission/roles` | 在全局 API 索引中把前两者转为已实现接口，只把 `/api/permission/roles` 继续标为 wrapper 漂移。 |
| Frontend page file vs active route | `frontend/src/pro/views/probeInstaller/index.tsx` 存在，但 `proRoutes` 实际复用 `IntegrationCenter/DataIntegration` | 在页面清单中标记为未接入主路由的历史页面。 |
| Client-side route guard | `AuthRouter` 权限检查逻辑注释掉 | 在所有权限描述里都不依赖前端守卫作为唯一证据。 |

## Consolidation Rules
为了避免后续文档再次重复或边界漂移，建议统一遵守：

1. 模块文档只解释本模块的页面、API、service、repository 和业务流。
2. 共享壳层、权限模型、license、数据范围、外部边界统一引用根目录索引或专题文档。
3. 不把 `/ai/api`、`/originx/api`、`/dify/console/api` 等旁路依赖误写为 `backend/` 内部实现。
4. 不按 `core`、`oss`、`pro` 目录名直接决定模块边界。
5. 涉及前后端源码时，额外补充源码归属说明，统一引用 [source_governance.md](source_governance.md)。
