# Frontend Backend Linkage

## Scope
本文档用于建立整个系统的前后端连接视图。重点不是枚举全部页面细节，而是明确以下链路：

- 前端 route 与页面入口
- 前端 API client
- 后端 route group 与 handler
- 后端 service
- repository / 数据源 / 外部系统

## Global Linkage Model
当前仓库的主要前后端连接模型如下：

```text
frontend route
  -> view / component tree
  -> frontend API module
  -> backend HTTP route group
  -> service layer
  -> repository / database / external system
  -> response model
  -> frontend state / UI
```

这条链路在本项目中不是完全静态的，因为菜单、可访问路由、license 限制、数据组范围和 deploy mode 都会影响实际可见行为。

## Application Shell Linkage
应用壳层由以下实现共同组成：

| Layer | Observed Files | Role |
|---|---|---|
| App bootstrap | `frontend/src/index.js` | 装配 Redux、Toast、DeployMode、User、DataGroup、License 等上下文。 |
| Root router | `frontend/src/App.js` | 用 `HashRouter` 装配 `/login` 与主布局。 |
| Layout | `frontend/src/core/layout/DefaultLayout.js` | 承载侧边栏、Header、主内容区、license 状态提示。 |
| Sidebar | `frontend/src/core/components/AppSidebar.js` | 根据后端返回菜单结构生成导航。 |
| Content switch | `frontend/src/core/components/AppContent.js` | 根据路由表渲染页面组件。 |

这意味着系统文档中“前端入口”不能只写 `App.js`，还必须覆盖 provider、layout、sidebar、content 这几个装配层。

## Dynamic Navigation And Permission Linkage
本项目最关键的前后端耦合之一，是“导航与权限链路”。

### Observed Flow

```text
frontend UserProvider
  -> GET /api/permission/config
  -> backend permission service
  -> database feature/menu/router/router_insert_page mapping
  -> frontend receives menuItem + routes
  -> AppSidebar renders menu
  -> AppContent resolves default route and available pages
```

### Key Files

| Link Step | Files |
|---|---|
| Frontend permission fetch | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/api/permission.js` |
| Frontend menu render | `frontend/src/core/components/AppSidebar.js` |
| Frontend route render | `frontend/src/core/components/AppContent.js` |
| Backend route | `backend/pkg/router/router_api.go` under `/api/permission/config` |
| Backend service | `backend/pkg/services/permission/service_user_configs.go` |
| Backend metadata init | `backend/pkg/repository/database/init_feature.go`, `init_router.go`, `init_menu_item.go`, `init_feature_mapping.go`, `init_router_page.go` |

### Documentation Implication
后续每个业务模块文档都要回答两个问题：

1. 该模块页面如何进入导航体系。
2. 该模块页面可见性如何映射到后端 feature / router / menu 模型。

## License And Edition Linkage
当前项目还有第二条强耦合链路，即 edition 与 license 限制。

### Observed Flow

```text
frontend route assembly + LicenseProvider
  -> backend LicMiddleware
  -> backend license service
  -> feature access decision
```

### Key Files

| Link Step | Files |
|---|---|
| Frontend route combination | `frontend/src/routes.js`, `frontend/src/pro/routes.js` |
| Frontend license state | `frontend/src/pro/contexts/LicenseContext.tsx`, `frontend/src/core/layout/DefaultLayout.js` |
| Backend middleware | `backend/pkg/middleware/lic_middleware.go` |
| Backend license APIs | `backend/pkg/api/license/`, `backend/pkg/services/license/` |

### Documentation Implication
文档不能把 `pro/` 目录简单解释为“天然启用的商业模块”。实际访问能力还受 license 和后端 middleware 控制。

## Data Scope And Time Context Linkage
多个观测页面共享同一组数据范围与时间范围上下文。

| Concern | Frontend Side | Backend Side | Notes |
|---|---|---|---|
| Data group scope | `DataGroupProvider`, `frontend/src/core/api/dataGroup.ts` | `/api/data`, `/api/v2/data`, `backend/pkg/services/data/` | 会影响多个业务页面的数据可见范围。 |
| Time range | Redux `timeRangeReducer` 与 Header 时间选择器 | 多数查询 API 请求参数 | 不是单一模块内部状态，而是跨观测域横切上下文。 |
| User identity | `localStorage token`, `UserProvider` | `AuthMiddleware`, `/api/user/*` | 影响所有授权型接口。 |

## Business Module Linkage Map
以下表格给出当前主要业务域的前后端连接方式。

| Module | Frontend Entry | Frontend API | Backend Route Group / Service | Data / External Dependencies | Notes |
|---|---|---|---|---|---|
| `service-observability` | `/service`, `/service/info` | `core/api/service.js`, `core/api/serviceInfo.js`, `pro/api/service.js` | `/api/service`, `services/service`, `services/serviceoverview` | ClickHouse, Prometheus, Kubernetes, Polaris | 这一模块聚焦 service 目录、service detail、依赖拓扑、错误实例、RED 数据与服务相关告警摘要。 |
| `logs-exploration-and-parsing` | `/logs/fault-site`, `/logs/full` | `core/api/logs.js` | `/api/log`, `services/log` | ClickHouse, database, Kubernetes, Prometheus | 前后端映射清晰，前端入口与后端 `log` route group 基本一一对应。 |
| `trace-exploration-and-profiling` | `/trace/fault-site`, `/trace/full` | `core/api/trace.js` | `/api/trace`, `services/trace` | ClickHouse, Jaeger, Prometheus, Kubernetes | 前端 trace 页面与后端 trace route group 对应关系稳定，且 profiler/flame 相关接口也集中在该组。 |
| `dashboards-and-health-monitoring` | `/system-dashboard`, `/basic-dashboard`, `/application-dashboard`, `/middleware-dashboard`, `/mysql-dashboard`, `/healthy-service` | dashboard 组件调用的 `core/api/service.js`、`pro/api/serviceHealthy.js` 等 | `/api/service/*`, `/api/metric`, `/api/network`, `services/metric`, `services/network` | Prometheus, DeepFlow ClickHouse, ClickHouse | 该模块整合聚合大盘和健康态监控，两者都以“汇总视图”而非 drill-down 视图为核心。 |
| `alert-rules-and-incident-events` | `/alerts/rule`, `/alerts/notify`, `/alerts/events`, `/alerts/events/detail/:alertId/:eventId`, `/report` | `core/api/alerts.js`, `pro/api/report.ts` | `/api/alerts`, `services/alerts` | database, ClickHouse, Prometheus, receivers, Dify | 告警规则、通知、事件、静默、恢复和事件报告读取都由 `/api/alerts` 主 route group 驱动，适合独立成模块。 |
| `alert-sources-and-enrichment` | `/integration/alerts` | `core/api/alertInput.ts` | `/api/alertinput`, `services/integration/alert` | database, Dify, schema/enrich metadata | 该模块关注告警是如何被接入、归一和 enrich，而不是事件发生后的治理，因此从告警治理模块中拆出。 |
| `automated-diagnosis-and-workflows` | `/workflows`, `/alert-analyze`, `/alert-analyze/info` | `core/api/workflows.ts`, `pro/api/alertAnalyze.js` | `/api/alerts/event/impact`, `/api/alerts/descendant/*`, `/api/alerts/detect/*`, `services/alertanalyze`, `repository/dify` | Dify, `/dify/console/api`, ClickHouse, Prometheus, Polaris | 该模块以自动化分析链路为中心，覆盖 workflows 嵌入、异常贡献度分析、缺陷检测和 Dify 驱动流程。 |
| `conversational-diagnosis-and-reporting` | `/cause/report/:traceId/:spanId`, `/cause/error-report/:traceId`, `/report`, `/alert-analyze/chatAI/:service/:endpoint`, `/agent` | `pro/api/chatAI.ts`, `pro/api/conversation.ts`, `pro/api/cause.ts`, `pro/api/report.ts`, `pro/api/custom.ts`, `pro/api/modelSetting.ts` | `/api/alerts/events/report/*`, `/api/custom/*`，以及外部 `/ai/api`、`/originx/api` | `/ai/api`, `/originx/api`, database, Dify-derived outputs | 该模块更偏会话式诊断、证据展示和报告消费，与自动化分析模块在前端交互模型和后端边界上都不同。 |
| `data-ingestion-and-oneagent` | `/integration/data`, `/integration/data/settings`, `probeInstaller` 相关页面 | `core/api/integration.ts`, `core/api/appConfig.js` | `/api/integration`, `services/integration`，以及外部 `/ai/api/v1/app-config/*` 边界 | database, static templates, OneAgent install assets, external Slack app config | 当前仓库中可见的 OneAgent 主要通过安装模板、安装文档与接入链路暴露，因此与数据接入模块合并更贴近真实实现。 |
| `dataplane-and-topology-governance` | `/dataplane` | `pro/api/dataplane.ts` | `/api/dataplane`, `services/dataplane`, `repository/dataplane` | ClickHouse, Prometheus, database | 该模块聚焦数据平面、服务名治理、自定义拓扑与 datasource 能力暴露。 |
| `identity-access-and-organization` | `/login`, `/user`, `/system/user-manage`, `/system/menu-manage`, `/system/role-manage`, `/system/team` | `core/api/user.js`, `core/api/permission.js`, `core/api/role.ts`, `core/api/team.ts` | `/api/user`, `/api/permission`, `/api/role`, `/api/team` | database, cache | 身份、权限、菜单、团队和角色都围绕同一套访问控制模型展开。 |
| `data-scope-governance` | `/system/data-group` | `core/api/dataGroup.ts` | `/api/data`, `/api/v2/data`, `services/data` | database, Prometheus, ClickHouse, Kubernetes | 该模块直接控制数据可见范围，是多个业务模块共享的治理能力。 |
| `platform-administration` | `/config`, `/system/platform-manage`，以及全局 branding 行为 | `core/api/config.js`, `pro/api/license.ts`, `core/api/app.js` | `/api/config`, `/api/license`, `/api/app`, `services/config`, `services/license`, `services/app` | database, Prometheus | 将 TTL 配置、license、logo、deploy mode、branding 收敛到同一平台管理模块后，前后端边界更清晰。 |

## External Boundary Linkage
当前前端显式依赖多类非 `backend/pkg/router/router_api.go` 内部主路由注册的边界：

| External Boundary | Observed Frontend Usage | Current Status |
|---|---|---|
| `/ai/api` | chat、agent、model setting、conversation、node info 等 | 开发阶段由 Vite 代理到本地外部服务，当前仓库未完整呈现其后端实现。 |
| `/originx/api` | cause report 相关查询 | 当前仓库中未见对应主后端路由注册。 |
| `/dify/console/api` | workflows 登录、登出、refresh token | 当前仓库中未见对应主后端路由注册。 |
| `/alert/api/v1` | alert rule 相关调用 | 当前仓库中未见对应主后端路由注册。 |

因此，后续写模块 `contracts.md` 时，必须把这些接口标注为“外部边界”或“旁路服务”，不能写成 `backend` 内部子模块。

## Route Drift And Mapping Risk
当前路由还有明显漂移风险：

- 前端 `pro` 路由使用 `/dataplane`、`/probe-management`、`/agent`
- 后端初始化的 `router` 控制数据中出现 `/agent-dataplane`、`probe-installer`、`agent`

这说明“前端 route 文件”和“后端权限模型中的 router 数据”并不完全一致。后续所有模块文档都需要同时检查：

1. 前端 route 定义。
2. 后端 `router_api.go` 注册。
3. 后端数据库初始化的 `validRouters`。

## Documentation Rule For Later Rounds
后续每个模块详细文档都应固定包含以下四段：

1. `Frontend Entry`
   说明页面入口、路由、主组件、上下文依赖。
2. `API Contracts`
   说明前端 API client 与后端 route group 的对应关系。
3. `Backend Processing`
   说明 handler、service、repository 的流向。
4. `Boundary`
   说明其依赖的数据库、观测存储、外部服务及未在本仓库实现的部分。

只有这样，整套文档才能真正把 `frontend` 与 `backend` 连成一张系统图，而不是形成两份彼此孤立的说明书。
