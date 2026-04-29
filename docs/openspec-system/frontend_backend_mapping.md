# Frontend Backend Mapping

## Scope
本文档把每个主要业务模块的前端入口、前端 API wrapper、后端 route group/service 和主要数据依赖放到同一张表里，作为后续阅读模块详细文档前的统一桥接层。

## Business Module Mapping

| Module | Frontend Routes And Pages | Frontend API Wrappers | Backend Routes And Services | Data / External Dependencies | Notes |
|---|---|---|---|---|---|
| `service-observability` | `/service`, `/service/info`; `frontend/src/oss/views/service/*`, `serviceInfo/*` | `frontend/src/core/api/service.js`, `serviceInfo.js`, `frontend/src/pro/api/service.js` | `/api/service/*`; `backend/pkg/api/service`, `backend/pkg/api/serviceoverview`, `backend/pkg/services/service`, `serviceoverview` | ClickHouse, Prometheus, Kubernetes, PolarisAnalyzer, database | service detail、entry endpoint、依赖关系和健康摘要都从这一链路进入。 |
| `logs-exploration-and-parsing` | `/logs/fault-site`, `/logs/full`; `frontend/src/oss/views/logs/*` | `frontend/src/core/api/logs.js` | `/api/log/*`; `backend/pkg/api/log`, `backend/pkg/services/log` | ClickHouse, database, Kubernetes, Prometheus | 前端日志查询、上下文和解析规则管理均映射到 `/api/log`。 |
| `trace-exploration-and-profiling` | `/trace/fault-site`, `/trace/full`; `frontend/src/oss/views/trace/*` | `frontend/src/core/api/trace.js` | `/api/trace/*`; `backend/pkg/api/trace`, `backend/pkg/services/trace` | Jaeger, ClickHouse, Prometheus, Kubernetes | flame graph 和 on/off CPU 也属于该模块，但部分接口不走 Auth。 |
| `dashboards-and-health-monitoring` | `/system-dashboard`, `/basic-dashboard`, `/application-dashboard`, `/middleware-dashboard`, `/mysql-dashboard`, `/healthy-service` | `frontend/src/core/api/service.js`, `frontend/src/pro/api/serviceHealthy.js`，部分页面间接复用 metrics/network 请求 | `/api/service/ryglight`, `/api/service/monitor/status`, `/api/metric/*`, `/api/network/*`; `backend/pkg/services/metric`, `network`, `serviceoverview` | Prometheus, DeepFlow ClickHouse, ClickHouse | 该模块自身没有单一前端 API wrapper 文件，更多依赖 service/metric/network 三组支撑接口。 |
| `alert-rules-and-incident-events` | `/alerts/rule`, `/alerts/notify`, `/alerts/events`, `/alerts/events/detail/:alertId/:eventId` | `frontend/src/core/api/alerts.js` | `/api/alerts/rule*`, `/api/alerts/event*`, `/api/alerts/slient*`, `/api/alerts/resolve`, `/api/alerts/events/report/*`; `backend/pkg/api/alerts`, `backend/pkg/services/alerts` | database, ClickHouse, Prometheus, alert receivers, Dify-derived report content | 事件报告的存储与读取接口在该模块，但最终报告消费页面归属会话式诊断模块。 |
| `alert-sources-and-enrichment` | `/integration/alerts`; `frontend/src/core/views/IntegrationCenter/AlertsIntegration/*` | `frontend/src/core/api/alertInput.ts` | `/api/alertinput/*`; `backend/pkg/api/alertinput` | database, schema metadata, cluster metadata, Dify-related provider setup | 负责告警源、schema、cluster 与 enrich 管理。 |
| `automated-diagnosis-and-workflows` | `/workflows`, `/alert-analyze`, `/alert-analyze/info` | `frontend/src/core/api/workflows.ts`, `frontend/src/pro/api/alertAnalyze.js` | `/api/alerts/event/impact`, `/api/alerts/descendant/*`, `/api/alerts/anomaly-span/list`, `/api/alerts/detect/*`; `backend/pkg/api/alertanalyze`, `backend/pkg/services/alertanalyze` | Dify, `/dify/console/api`, ClickHouse, Prometheus, PolarisAnalyzer | `workflows` 登录与 token 刷新是外部 Dify console 边界，不等于本仓库后端实现。 |
| `conversational-diagnosis-and-reporting` | `/cause/report/:traceId/:spanId`, `/cause/error-report/:traceId`, `/alert-analyze/chatAI/:service/:endpoint`, `/report`, `/agent` | `frontend/src/pro/api/chatAI.ts`, `conversation.ts`, `cause.ts`, `custom.ts`, `modelSetting.ts`, `report.ts` | `/api/custom/*`，以及依赖 `alert-rules-and-incident-events` 提供的 `/api/alerts/events/report/*`; 其余大量调用落向外部 `/ai/api/*`、`/originx/api/*` | `/ai/api`, `/originx/api`, database, Dify outputs | 这是前后端映射里外部边界最重的模块，必须显式标注旁路服务。 |
| `ops-knowledge-base-management` | `/knowledge-bases`; `frontend/src/pro/views/knowledgeBases/*`; 顶层菜单 `knowledgeBases` | `frontend/src/pro/api/knowledgeBase.ts`, `knowledgeBaseClient.ts` | 当前 APO Go 后端无知识库管理 API route；仅通过 permission/menu/router 初始化提供菜单元数据；管理接口通过外部 `/rag/api/kb/*` 消费 rag-management，包括文档列表、上传、导入状态、原文件下载和删除 | rag-management, `/rag/api/kb/*` | 只负责 APO 侧知识库管理页面、菜单入口和 RAG 管理接口消费；AgentSRE 仍只读消费 RAG 检索。 |
| `data-ingestion-and-oneagent` | `/integration/data`, `/integration/data/settings`, `/probe-management`, `/probe-management/settings`；`frontend/src/core/views/IntegrationCenter/DataIntegration/*` | `frontend/src/core/api/integration.ts`, `frontend/src/core/api/appConfig.js` | `/api/integration/*`; `backend/pkg/api/integration`, `backend/pkg/services/integration` | database, `backend/static/integration-tmpl`, OneAgent install assets, external app-config service | `probe-management` 当前复用 `DataIntegration` 页面；`PRODUCT_MODE=syncause` 时菜单/feature/router 初始化会切到这条 minimal 接入链路。 |
| `dataplane-and-topology-governance` | `/dataplane`; `frontend/src/pro/views/agentDataplane/*` | `frontend/src/pro/api/dataplane.ts` | `/api/dataplane/*`; `backend/pkg/api/dataplane`, `backend/pkg/services/dataplane`, `backend/pkg/repository/dataplane` | ClickHouse, Prometheus, database | `/api/dataplane/capabilities` 已实现，但语义是已配置 dataplane capability 摘要；历史 allowlist 仍残留 `/agent-dataplane`。 |
| `identity-access-and-organization` | `/login`, `/user`, `/system/user-manage`, `/system/menu-manage`, `/system/role-manage`, `/system/team` | `frontend/src/core/api/user.js`, `permission.js`, `role.ts`, `team.ts` | `/api/user/*`, `/api/permission/*`, `/api/role/*`, `/api/team/*`; `backend/pkg/services/user`, `permission`, `role`, `team` | database, cache, Dify user sync | `getAllRoleList()` 调用 `/api/permission/roles`，但后端主路由只注册 `/api/role/roles`。 |
| `data-scope-governance` | `/system/data-group`; `frontend/src/core/views/DataGroup/*` | `frontend/src/core/api/dataGroup.ts` | `/api/data/*`, `/api/v2/data/*`; `backend/pkg/api/data`, `backend/pkg/services/data` | database, Prometheus, ClickHouse, Kubernetes | 当前前端主流程已偏向 `/api/v2/data/group*`，旧版 wrapper 仍保留。 |
| `platform-administration` | `/config`, `/system/platform-manage`, `/demo/service-config` | `frontend/src/core/api/config.js`, `frontend/src/core/api/app.js`, `frontend/src/pro/api/license.ts` | `/api/config/*`, `/api/license/*`, `/api/app/logo`, `/api/app/upload`, `/api/app/deployMode`; `backend/pkg/services/config`, `license`, `app` | database, Prometheus | `getDeployModeApi()` 已有后端实现，但它只是读取 `PRODUCT_MODE`；管理端 branding UI 仍未开放。 |

## Source Ownership Overlay
为了把“前后端映射”写准确，当前还必须补一层“源码归属”说明：

| Module | Frontend Ownership | Backend Ownership | Mapping Implication |
|---|---|---|---|
| `service-observability` | 前端以 `upstream-common` 为主 | 后端以共享包为主 | 页面主体接近基线，但运行时依赖已分叉的 app shell、request 与 middleware。 |
| `logs-exploration-and-parsing` | 前端以 `upstream-common` 为主 | 后端以共享包为主 | 业务链清晰，但时间范围与 license 门控属于共享层分叉。 |
| `trace-exploration-and-profiling` | 前端以 `upstream-common` 为主 | 后端以共享包为主 | 页面主体共享，后端 route group 已叠加商业仓门控。 |
| `dashboards-and-health-monitoring` | `mixed` | 共享包 + 本地扩展 | 同时包含共享 dashboard、共享路径变体和 Pro 页面。 |
| `alert-rules-and-incident-events` | `mixed` | 共享包上的 `commercial-variant` | 后端 `alerts` 包已新增事件报告能力。 |
| `alert-sources-and-enrichment` | 前端更接近 `upstream-common` | 共享包 + 门控 | 接入治理主体共享，但已纳入商业仓统一门控。 |
| `automated-diagnosis-and-workflows` | `mixed` | 存在 `commercial-only` 后端包 | `alertanalyze` 是商业仓独有后端扩展，不能被写成共享后端能力。 |
| `conversational-diagnosis-and-reporting` | 以 `commercial-only` 为主 | 存在 `commercial-only` 后端包 | 前端 Pro 与后端 `custom` 包形成更清晰的商业链路。 |
| `ops-knowledge-base-management` | 以 `commercial-only` 为主，顶层菜单图标触及共享前端路径 | 无本仓知识库管理 API；菜单权限元数据在共享后端候选初始化文件中 | 页面和 wrapper 在 Pro；知识库管理能力归属外部 rag-management，菜单可见性由本仓权限模型控制。 |
| `data-ingestion-and-oneagent` | `mixed` | 共享包上的本地扩展 | 前端同时连接本仓 `/api/integration/*` 与外部 `/ai/api/v1/app-config/*`。 |
| `dataplane-and-topology-governance` | 前端以 `commercial-only` 为主 | 共享包 + 本地扩展 | 前端在 Pro，后端 `dataplane` 共享包也被商业仓显著扩展。 |
| `identity-access-and-organization` | `mixed` | 共享包 + 门控扩展 | 登录壳层与后端权限链都已偏离上游纯基线。 |
| `data-scope-governance` | `mixed` | 共享包 + 门控扩展 | DataGroup 主体共享，但前后端可见性控制都已变化。 |
| `platform-administration` | `mixed` 且商业特征最强 | 存在 `commercial-only` 包 | 前端 `/config` 共享，`license`、`app`、branding 是商业扩展。 |

更完整的归属规则见 [source_governance.md](source_governance.md)。

## Shared Infrastructure Mapping

| Shared Area | Frontend Side | Backend Side | Why It Matters |
|---|---|---|---|
| App shell and route render | `index.js`, `App.js`, `DefaultLayout`, `AppSidebar`, `AppContent` | `backend/pkg/services/permission/service_user_configs.go`, menu/router init data | 前端页面是否可见，与后端菜单配置直接耦合。 |
| Request and auth | `frontend/src/core/utils/request.js` | `AuthMiddleware`, `/api/user/login`, `/api/user/refresh` | token 注入、刷新和登录态恢复贯穿全部业务模块。 |
| License gating | `LicenseContext.tsx`, `DefaultLayout.js` | `LicMiddleware`, `/api/license/*` | 控制 pro 能力与平台告警横幅。 |
| Data scope context | `DataGroupContext.tsx`, Redux time range | `backend/pkg/services/data/*` | 决定多类查询的可见范围和过滤条件。 |

## Missing Or Weak Frontend-Backend Links

| Path Or Area | Current Status | Impact |
|---|---|---|
| `/api/app/deployMode` | 已在主后端注册，直接返回 `PRODUCT_MODE` | `platform-administration` 的 deploy mode 读取链路已闭环，但仍缺少可视化配置模型。 |
| `/api/dataplane/capabilities` | 已在主后端注册，按已配置 dataplane 聚合 capability | `dataplane-and-topology-governance` 的能力卡片更偏配置态摘要，不等于运行时健康。 |
| `/api/permission/roles` | wrapper 存在，主后端注册的是 `/api/role/roles` | `identity-access-and-organization` 有潜在历史 wrapper 漂移。 |
| `/ai/api/*`, `/rag/api/kb/*`, `/originx/api/*`, `/dify/console/api/*`, `/alert/api/v1/*` | 由前端直接调用或通过代理调用 | 必须持续标为外部边界，而不能并入 `backend/` 内部实现。 |
