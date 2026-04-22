# Domain Map

## Scope
本文档给出当前项目的规范化模块边界。目标不是按目录机械切块，而是按业务能力、共享基础设施和横切关注点三类做稳定划分。

## Business Modules

### Observability Consumption Domains

| Module | Frontend Roots | Backend Roots | Why This Boundary Is Reasonable |
|---|---|---|---|
| `service-observability` | `frontend/src/oss/views/service`, `frontend/src/oss/views/serviceInfo`, `frontend/src/core/api/service.js`, `frontend/src/core/api/serviceInfo.js`, `frontend/src/pro/api/service.js` | `backend/pkg/api/service`, `backend/pkg/api/serviceoverview`, `backend/pkg/services/service`, `backend/pkg/services/serviceoverview` | 这些代码共同围绕 service 目录、service detail、entry endpoint、依赖拓扑、错误实例和 RED 指标展开。 |
| `logs-exploration-and-parsing` | `frontend/src/oss/views/logs`, `frontend/src/core/api/logs.js` | `backend/pkg/api/log`, `backend/pkg/services/log` | 日志查询与解析规则管理共享同一批页面与接口，不宜拆成两个文档模块。 |
| `trace-exploration-and-profiling` | `frontend/src/oss/views/trace`, `frontend/src/core/api/trace.js` | `backend/pkg/api/trace`, `backend/pkg/services/trace` | trace 查询、trace 详情和 profiling 能力在前后端都收敛到同一组 route 与 service。 |
| `dashboards-and-health-monitoring` | `frontend/src/oss/views/dashboard`, `frontend/src/pro/views/MysqlDashboard`, `frontend/src/pro/views/serviceHealthy` | `backend/pkg/api/metric`, `backend/pkg/api/network`, `/api/service` 中 dashboard/健康相关能力，`backend/pkg/services/metric`, `backend/pkg/services/network` | 聚合大盘和健康监控都属于“汇总观察视图”，与 service detail 或 trace detail 的钻取式页面不同。 |

### Alerting And Diagnosis Domains

| Module | Frontend Roots | Backend Roots | Why This Boundary Is Reasonable |
|---|---|---|---|
| `alert-rules-and-incident-events` | `frontend/src/oss/views/alerts`, `frontend/src/oss/views/alertEvents`, `frontend/src/core/api/alerts.js`, `frontend/src/pro/api/report.ts` | `backend/pkg/api/alerts`, `backend/pkg/services/alerts` | 这一模块关注告警规则、通知、事件、静默、恢复和事件报告读取，属于事件发生后的治理链路。 |
| `alert-sources-and-enrichment` | `frontend/src/core/views/IntegrationCenter/AlertsIntegration`, `frontend/src/core/api/alertInput.ts` | `backend/pkg/api/alertinput`, `backend/pkg/services/integration/alert` 风格的接入逻辑，`backend/pkg/model/integration/alert` 相关模型缺省并入 alert input/request-response 家族 | 这一模块关注告警如何被接入、建模和 enrich，与事件治理链路不同。 |
| `automated-diagnosis-and-workflows` | `frontend/src/oss/views/workflows`, `frontend/src/pro/views/alertAnalyze`, `frontend/src/pro/views/serviceAnalyze`, `frontend/src/core/api/workflows.ts`, `frontend/src/pro/api/alertAnalyze.js` | `backend/pkg/api/alertanalyze`, `/api/alerts` 中 impact、descendant、detect 相关接口，`backend/pkg/services/alertanalyze` | 该模块核心是自动化分析、workflow 和异常贡献度，不是聊天式交互。 |
| `conversational-diagnosis-and-reporting` | `frontend/src/pro/views/causeReport`, `frontend/src/pro/views/report`, `frontend/src/pro/views/chatAI`, `frontend/src/pro/views/agent`, `frontend/src/pro/api/chatAI.ts`, `conversation.ts`, `cause.ts`, `custom.ts`, `modelSetting.ts` | `backend/pkg/api/custom`, `backend/pkg/services/custom`，以及外部 `/ai/api`、`/originx/api` | 该模块以会话、报告、证据展示和自定义查询为核心，和自动化分析链路的交互模型明显不同。 |

### Governance And Administration Domains

| Module | Frontend Roots | Backend Roots | Why This Boundary Is Reasonable |
|---|---|---|---|
| `data-ingestion-and-oneagent` | `frontend/src/core/views/IntegrationCenter/DataIntegration`, `frontend/src/core/api/integration.ts`, `frontend/src/core/api/appConfig.js`, `frontend/src/pro/views/probeInstaller` | `backend/pkg/api/integration`, `backend/pkg/services/integration`, `backend/static/integration-tmpl`, `one-agent/README.md` | 当前仓库中可见的 OneAgent 能力主要体现在接入与安装模板，而不是完整 agent 源码，因此与数据接入合并更真实。 |
| `dataplane-and-topology-governance` | `frontend/src/pro/views/agentDataplane`, `frontend/src/pro/api/dataplane.ts` | `backend/pkg/api/dataplane`, `backend/pkg/services/dataplane`, `backend/pkg/repository/dataplane` | 数据平面、服务名规则、自定义拓扑和数据源能力暴露属于同一治理面。 |
| `identity-access-and-organization` | `frontend/src/core/views/Login`, `UserPage`, `UserManage`, `MenuManage`, `RoleManage`, `Team`, `frontend/src/core/api/user.js`, `permission.js`, `role.ts`, `team.ts` | `backend/pkg/api/user`, `permission`, `role`, `team`, `backend/pkg/services/user`, `permission`, `role`, `team` | 登录、用户、团队、角色、菜单、权限都围绕同一访问控制模型运行。 |
| `data-scope-governance` | `frontend/src/core/views/DataGroup`, `frontend/src/core/contexts/DataGroupContext.tsx`, `frontend/src/core/api/dataGroup.ts` | `backend/pkg/api/data`, `backend/pkg/services/data`, `backend/pkg/model/datagroup` | DataGroup 是多个观测模块共享的数据范围边界，不能被埋进单个系统管理页。 |
| `platform-administration` | `frontend/src/oss/views/config`, `frontend/src/pro/views/platformManage`, `frontend/src/pro/views/LicenseManage`, `frontend/src/core/api/config.js`, `frontend/src/core/api/app.js`, `frontend/src/pro/api/license.ts` | `backend/pkg/api/config`, `license`, `app`, `backend/pkg/services/config`, `license`, `app` | TTL、license、logo、deploy mode 都是平台级管理能力，不属于某个观测业务域。 |

## Shared Infrastructure Modules

| Shared Module | Primary Sources | Role |
|---|---|---|
| `frontend-app-shell` | `frontend/src/index.js`, `App.js`, `src/routes.js`, `src/core/layout`, `src/core/components/AppSidebar.js`, `AppContent.js` | 前端启动、布局、主路由渲染、默认跳转。 |
| `frontend-state-request-and-i18n` | `frontend/src/core/utils/request.js`, `frontend/src/core/contexts`, `frontend/src/core/store`, `frontend/src/i18n.js`, `frontend/public/locales` | 请求封装、状态、国际化和通知。 |
| `backend-http-kernel` | `backend/main.go`, `backend/pkg/router`, `backend/pkg/core`, `backend/pkg/middleware` | HTTP 启动、路由注册、中间件和统一响应。 |
| `backend-persistence-and-connectors` | `backend/pkg/repository/*` | 数据库、缓存、ClickHouse、Prometheus、Jaeger、Kubernetes、Dify 等连接层。 |
| `configuration-logging-and-bootstrap` | `backend/config`, `backend/pkg/logger`, `backend/Dockerfile` | 运行时配置、日志、启动与打包。 |
| `deployment-assets-and-static-templates` | `backend/static`, `backend/sqlscripts`, `frontend/public`, `one-agent/README.md` | 安装模板、初始化 SQL、静态资源与安装说明。 |

## Cross-Cutting Concerns

| Cross-Cutting Concern | Primary Sources | Why It Should Stay Separate |
|---|---|---|
| `auth-and-session` | `frontend/src/core/views/Login`, `frontend/src/core/utils/request.js`, `backend/pkg/api/user`, `backend/pkg/middleware/auth_middleware.go` | 登录态和 token 机制作用于全站，不属于某一个业务模块。 |
| `permission-menu-router-model` | `frontend/src/core/contexts/UserContext.js`, `AppSidebar.js`, `backend/pkg/services/permission`, `backend/pkg/repository/database/init_*.go` | 页面可见性、菜单结构、router 元数据和 feature 权限跨全部业务模块。 |
| `license-and-edition-gating` | `frontend/src/pro/contexts/LicenseContext.tsx`, `frontend/src/routes.js`, `backend/pkg/middleware/lic_middleware.go` | license 影响多个模块的前端与后端访问。 |
| `data-group-and-time-range-context` | `frontend/src/core/contexts/DataGroupContext.tsx`, Redux time range reducer, `backend/pkg/services/data` | 多个观测页面共享同一组数据范围和时间范围。 |
| `external-service-boundaries` | `frontend/src/pro/api/*`, `frontend/src/core/api/workflows.ts`, `frontend/vite.config.mjs`, `backend/config/apo.yml` | 用于严格区分“本仓库实现”和“旁路服务/外部系统”。 |

## Duplicate Module Check
当前 `modules/` 目录中没有发现“同一业务模块被用两个不同目录名重复记录”的情况。需要特别注意的不是目录重复，而是历史页面、旧路由或外部接口让边界看起来像重复：

- `alert-rules-and-incident-events` 与 `alert-sources-and-enrichment`：前者处理告警发生后的治理，后者处理告警接入前置建模。
- `automated-diagnosis-and-workflows` 与 `conversational-diagnosis-and-reporting`：前者偏自动化分析链路，后者偏聊天、报告与证据消费。
- `data-ingestion-and-oneagent` 与 `dataplane-and-topology-governance`：前者偏接入与安装，后者偏数据面和服务治理。
- `platform-administration` 中同时包含 `/config` 与 `/system/platform-manage`：这是同一平台管理域下的两个入口，不再拆成独立模块。

## Boundary Decisions
本次整编后的模块边界以“用户可感知能力”优先，而不是“目录名”或“技术栈”优先。这样可以确保：

- 同一业务能力跨 `frontend/src/core`、`frontend/src/oss`、`frontend/src/pro`、`backend/pkg/api`、`backend/pkg/services` 时，仍然只归属一个模块。
- 共享壳层、请求层、middleware、repository、初始化元数据等不会被错误挂到某个业务模块名下。
- 后续扩展文档时，可以直接从业务能力进入，再沿着前端页面、API wrapper、后端 route group 和 service 继续钻取。

## Source Ownership Overlay
模块边界与源码归属是两套正交维度。后续写模块文档时，除了回答“这个模块是什么”，还要回答“这部分源码由上游共享基线、共享变体还是商业独有实现构成”。

| Module | Ownership Overlay | Why It Matters |
|---|---|---|
| `service-observability` | 以前端 `upstream-common` 为主，后端以共享包为主 | 页面和 API 主体接近基线，但运行时仍依赖已分叉的壳层和后端中间件。 |
| `logs-exploration-and-parsing` | 前端以 `upstream-common` 为主，后端为共享包 | 时间范围、请求层和中间件门控已分叉，不能只写业务页。 |
| `trace-exploration-and-profiling` | 前端以 `upstream-common` 为主，后端为共享包 | trace 业务主体共享，但壳层、共享组件和后端门控有本地分叉。 |
| `dashboards-and-health-monitoring` | `mixed` | 同时包含共享 dashboard、共享路径变体和 Pro 页面。 |
| `alert-rules-and-incident-events` | `mixed` | 告警治理主体共享，但后端 alerts 包已新增事件报告能力。 |
| `alert-sources-and-enrichment` | 前端更接近 `upstream-common`，后端为共享包叠加门控 | 文档要区分共享接入逻辑与商业仓统一 license 门控。 |
| `automated-diagnosis-and-workflows` | `mixed`，后端存在明显 `commercial-only` 包 | 前端横跨共享与 Pro，后端 `alertanalyze` 是商业仓独有扩展。 |
| `conversational-diagnosis-and-reporting` | 以 `commercial-only` 为主 | 前端 Pro 与后端 `custom` 包都偏商业独有。 |
| `data-ingestion-and-oneagent` | `mixed` | 共享接入主线之上叠加了商业 Slack Bot 与外部 app-config 边界。 |
| `dataplane-and-topology-governance` | `mixed` 偏商业增强 | 前端在 Pro，后端 dataplane 共享包也被商业仓显著扩展。 |
| `identity-access-and-organization` | `mixed` | 用户/团队主体共享，但 Login、Sidebar、AuthRouter 和后端门控链路都已分叉。 |
| `data-scope-governance` | `mixed` | 前端共享基线之上增加本地控制，后端 data 包也挂了 license 门控。 |
| `platform-administration` | `mixed` 且商业特征最强 | 前端 `/config` 共享，前后端的 `license`、`app`、branding 都是商业增强。 |

更详细的源码归属规则见 [source_governance.md](source_governance.md)。
