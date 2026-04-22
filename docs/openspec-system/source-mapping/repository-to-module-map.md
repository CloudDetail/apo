# Repository To Module Map

## Scope
本文档把当前仓库源码映射到建议的文档模块。映射目标不是机械地按文件夹归类，而是按业务域、共享基础设施和横切关注点建立稳定文档边界。

## Mapping Principles

### Principle 1
同一业务域即使横跨 `frontend/src/core`、`frontend/src/oss`、`frontend/src/pro`、`backend/pkg/api`、`backend/pkg/services`、`backend/pkg/repository`，也应映射到同一个模块文档。

### Principle 2
共享基础设施单独成章，不混入某一个业务模块。

### Principle 3
权限、license、数据范围、外部边界这类横切主题单独维护。

### Principle 4
对当前仓库未完整包含实现的外部依赖，只记录边界，不把外部服务纳入本仓库模块内部实现。

## Business Module Mapping

| Proposed Module | Primary Source Areas | Why This Mapping Is Reasonable |
|---|---|---|
| `service-observability` | `frontend/src/oss/views/service`, `frontend/src/oss/views/serviceInfo`, `frontend/src/core/api/service.js`, `frontend/src/core/api/serviceInfo.js`, `frontend/src/pro/api/service.js`, `backend/pkg/api/service`, `backend/pkg/api/serviceoverview`, `backend/pkg/services/service`, `backend/pkg/services/serviceoverview` | 这些代码共同围绕 service 目录、依赖关系、拓扑、error instance、entry endpoint 和 service detail 展开，应视为一个业务域。 |
| `logs-exploration-and-parsing` | `frontend/src/oss/views/logs`, `frontend/src/core/api/logs.js`, `backend/pkg/api/log`, `backend/pkg/services/log`, `backend/pkg/repository/database/dao_logtableinfo.go`, `dao_othertableinfo.go` | 这一组同时覆盖日志查询与解析配置，比单写“日志查询”更贴近真实实现边界。 |
| `trace-exploration-and-profiling` | `frontend/src/oss/views/trace`, `frontend/src/core/api/trace.js`, `backend/pkg/api/trace`, `backend/pkg/services/trace`, `backend/pkg/repository/jaeger` | 页面和后端 route group 都围绕 trace 及 profiling 相关能力集中组织。 |
| `dashboards-and-health-monitoring` | `frontend/src/oss/views/dashboard`, `frontend/src/pro/views/MysqlDashboard`, `frontend/src/pro/views/serviceHealthy`, `backend/pkg/api/metric`, `backend/pkg/api/network`, `backend/pkg/services/metric`, `backend/pkg/services/network`, `backend/pkg/api/service` 中 `ryglight` 与 `monitor/status` 相关能力 | dashboard 和健康监控都提供聚合观察视图，适合统一归档。 |
| `alert-rules-and-incident-events` | `frontend/src/oss/views/alerts`, `frontend/src/oss/views/alertEvents`, `frontend/src/core/api/alerts.js`, `frontend/src/pro/api/report.ts`, `backend/pkg/api/alerts`, `backend/pkg/services/alerts` | 该模块聚焦告警规则、通知、事件、silence、resolve、事件报告等“事件治理”能力。 |
| `alert-sources-and-enrichment` | `frontend/src/core/views/IntegrationCenter/AlertsIntegration`, `frontend/src/core/api/alertInput.ts`, `backend/pkg/api/alertinput`, `backend/pkg/services/integration/alert`, `backend/pkg/model/integration/alert` | 告警来源接入、schema、cluster、标签 enrich 与告警事件治理是不同子系统，应独立成模块。 |
| `automated-diagnosis-and-workflows` | `frontend/src/oss/views/workflows`, `frontend/src/pro/views/alertAnalyze`, `frontend/src/pro/views/serviceAnalyze`, `frontend/src/core/api/workflows.ts`, `frontend/src/pro/api/alertAnalyze.js`, `backend/pkg/api/alertanalyze`, `backend/pkg/services/alertanalyze`, `backend/pkg/repository/dify` | 这组代码围绕自动化诊断、workflow 执行、异常贡献度分析与缺陷检测，边界比“全部 AI 能力”更清晰。 |
| `conversational-diagnosis-and-reporting` | `frontend/src/pro/views/causeReport`, `frontend/src/pro/views/report`, `frontend/src/pro/views/chatAI`, `frontend/src/pro/views/agent`, `frontend/src/pro/api/chatAI.ts`, `frontend/src/pro/api/conversation.ts`, `frontend/src/pro/api/cause.ts`, `frontend/src/pro/api/report.ts`, `frontend/src/pro/api/custom.ts`, `frontend/src/pro/api/modelSetting.ts`, `backend/pkg/api/custom`, `backend/pkg/services/custom`, `backend/pkg/api/alerts` 中 report 相关接口 | 这一组以会话、证据展示、报告消费和自定义数据视图为核心，与自动化分析链路应分开记录。 |
| `data-ingestion-and-oneagent` | `frontend/src/core/views/IntegrationCenter/DataIntegration`, `frontend/src/pro/views/probeInstaller`, `frontend/src/core/api/integration.ts`, `frontend/src/core/api/appConfig.js`, `backend/pkg/api/integration`, `backend/pkg/services/integration`, `backend/static/integration-tmpl`, `one-agent/README.md`, `.gitmodules` | 当前主仓库内可见的 OneAgent 主要通过接入流程、安装配置和模板体现，因此与数据接入合并更合理。 |
| `dataplane-and-topology-governance` | `frontend/src/pro/views/agentDataplane`, `frontend/src/pro/api/dataplane.ts`, `backend/pkg/api/dataplane`, `backend/pkg/services/dataplane`, `backend/pkg/repository/dataplane`, `backend/pkg/repository/database/dao_dataplane.go`, `dao_service_name_rule*.go`, `dao_custom_service_topology.go` | 该模块同时管理数据平面、服务名治理与自定义拓扑，是明确的平台治理能力。 |
| `identity-access-and-organization` | `frontend/src/core/views/Login`, `frontend/src/core/views/UserPage`, `frontend/src/core/views/UserManage`, `frontend/src/core/views/MenuManage`, `frontend/src/core/views/RoleManage`, `frontend/src/core/views/Team`, `frontend/src/core/api/user.js`, `frontend/src/core/api/permission.js`, `frontend/src/core/api/role.ts`, `frontend/src/core/api/team.ts`, `backend/pkg/api/user`, `backend/pkg/api/permission`, `backend/pkg/api/role`, `backend/pkg/api/team`, `backend/pkg/services/user`, `backend/pkg/services/permission`, `backend/pkg/services/role`, `backend/pkg/services/team` | 用户、角色、菜单、团队和权限共同定义身份与组织边界，应以单一治理模块呈现。 |
| `data-scope-governance` | `frontend/src/core/views/DataGroup`, `frontend/src/core/contexts/DataGroupContext.tsx`, `frontend/src/core/api/dataGroup.ts`, `backend/pkg/api/data`, `backend/pkg/services/data`, `backend/pkg/model/datagroup`, `backend/pkg/repository/database/dao_data_group.go`, `dao_data_group_new.go`, `dao_data_scope.go`, `dao_auth_data_group.go` | DataGroup 与数据 scope 是共享治理能力，不宜只是系统管理子页。 |
| `platform-administration` | `frontend/src/oss/views/config`, `frontend/src/pro/views/platformManage`, `frontend/src/pro/views/LicenseManage`, `frontend/src/pro/api/license.ts`, `frontend/src/core/api/config.js`, `frontend/src/core/api/app.js`, `frontend/src/core/contexts/DeployModeContext.js`, `backend/pkg/api/config`, `backend/pkg/api/license`, `backend/pkg/api/app`, `backend/pkg/services/config`, `backend/pkg/services/license`, `backend/pkg/services/app`, `backend/pkg/repository/database/license`, `backend/pkg/repository/database/app` | TTL 配置、license、logo、deploy mode、branding 都属于平台管理而非观测主域，应合并记录。 |

## Shared Infrastructure Mapping

| Shared Infrastructure | Primary Source Areas | Why This Mapping Is Reasonable |
|---|---|---|
| `frontend-app-shell` | `frontend/src/index.js`, `frontend/src/App.js`, `frontend/src/routes.js`, `frontend/src/oss/routes.js`, `frontend/src/pro/routes.js`, `frontend/src/core/layout`, `frontend/src/core/components/AppSidebar.js`, `frontend/src/core/components/AppContent.js` | 这些代码共同定义前端如何启动、如何组装布局、如何渲染主页面。 |
| `frontend-state-request-and-i18n` | `frontend/src/core/utils/request.js`, `frontend/src/core/contexts`, `frontend/src/core/store`, `frontend/src/i18n.js`, `frontend/public/locales` | 请求、状态、国际化、通知机制是全站共享设施。 |
| `backend-http-kernel` | `backend/main.go`, `backend/pkg/router`, `backend/pkg/core`, `backend/pkg/middleware` | 定义后端启动、路由注册、错误处理、中间件执行链。 |
| `backend-persistence-and-connectors` | `backend/pkg/repository/database`, `backend/pkg/repository/clickhouse`, `backend/pkg/repository/prometheus`, `backend/pkg/repository/jaeger`, `backend/pkg/repository/kubernetes`, `backend/pkg/repository/dify`, `backend/pkg/repository/cache`, `backend/pkg/repository/polarisanalyzer` | 这是后端连接数据库和外部观测系统的共用数据访问层。 |
| `configuration-logging-and-bootstrap` | `backend/config`, `backend/pkg/logger`, `backend/Dockerfile` | 配置、日志、运行时打包是系统基础能力。 |
| `deployment-assets-and-static-templates` | `backend/static`, `backend/sqlscripts`, `frontend/public`, `one-agent/README.md` | 这些资产参与安装、初始化、预定义模板与资源分发。 |

## Cross-Cutting Mapping

| Cross-Cutting Concern | Primary Source Areas | Why This Mapping Is Reasonable |
|---|---|---|
| `auth-and-session` | `frontend/src/core/utils/request.js`, `frontend/src/core/views/Login`, `backend/pkg/api/user`, `backend/pkg/middleware/auth_middleware.go`, `backend/pkg/util/jwt` | 登录态、token 注入、refresh、认证中间件共同构成会话机制。 |
| `permission-menu-router-model` | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/components/AppSidebar.js`, `backend/pkg/services/permission`, `backend/pkg/repository/database/init_feature.go`, `init_router.go`, `init_menu_item.go`, `init_feature_mapping.go`, `init_router_page.go` | 导航、路由、权限、插页文案的底层控制数据都在这里汇合。 |
| `license-and-edition-gating` | `frontend/src/routes.js`, `frontend/src/pro/contexts/LicenseContext.tsx`, `backend/pkg/middleware/lic_middleware.go`, `backend/pkg/services/license` | edition 与 license 决定能力开放范围，是典型横切主题。 |
| `data-group-and-time-range-context` | `frontend/src/core/contexts/DataGroupContext.tsx`, `frontend/src/core/store/reducers/timeRangeReducer.js`, `backend/pkg/services/data` | 数据范围与时间范围共同影响多个观测模块查询结果。 |
| `observability-data-flow` | `one-agent`, `backend/pkg/repository/*`, `frontend/src/oss/views/*`, `frontend/src/pro/views/*` 中观测消费视图 | 这一主题跨采集、存储、查询、展示四层，不属于单一目录。 |
| `external-service-boundaries` | `frontend/src/pro/api/*`, `frontend/src/core/api/workflows.ts`, `frontend/vite.config.mjs`, `backend/config/apo.yml` 中 Dify/Jaeger 等配置 | 该主题负责区分仓库实现与外部依赖边界。 |
| `i18n-and-branding` | `frontend/src/i18n.js`, `frontend/public/locales`, `frontend/src/core/contexts/DeployModeContext.js`, `backend/pkg/services/app` | 国际化与品牌切换跨多个页面与平台管理逻辑。 |

## External Or Partial Boundaries

| Boundary | Observed Source Areas | Current Interpretation |
|---|---|---|
| `/ai/api` | `frontend/src/pro/api/modelSetting.ts`, `chatAI`, `agent`, `conversation` 相关代码 | 当前仓库前端明确依赖，主后端未完整呈现其实现。 |
| `/originx/api` | `frontend/src/pro/api/cause.ts` | 当前仓库可视为外部诊断查询边界。 |
| `/dify/console/api` | `frontend/src/core/api/workflows.ts`, `request.js` | 当前仓库中作为外部 console 或旁路服务边界处理更安全。 |
| `/alert/api/v1` | `frontend/src/core/api/alerts.js` | 与 `/api/alerts/*` 并存，需后续继续确认。 |

## Files That Need Joint Reading
后续写详细文档时，以下文件组合应始终联读：

| Topic | Joint Reading Set |
|---|---|
| Frontend navigation | `frontend/src/routes.js` + `frontend/src/core/components/AppSidebar.js` + `frontend/src/core/contexts/UserContext.js` |
| Permission metadata | `backend/pkg/services/permission/service_user_configs.go` + `backend/pkg/repository/database/init_*.go` |
| Backend API chain | `backend/pkg/router/router_api.go` + `backend/pkg/api/<module>` + `backend/pkg/services/<module>` |
| Ingestion and install | `frontend/src/core/views/IntegrationCenter/*` + `backend/pkg/api/integration` + `backend/pkg/services/integration` + `backend/static/integration-tmpl` |
| AI domain | `frontend/src/pro/views/chatAI` + `frontend/src/pro/views/agent` + `frontend/src/pro/api/*` + `backend/pkg/api/alertanalyze` + `backend/pkg/repository/dify` |

## Documentation Implication
这份映射表将作为后续所有模块文档的落点依据。除非后续在代码中发现新的证据，否则不建议再按“纯目录名”重构模块边界。优先保持业务域稳定，再在模块内部记录目录分布。
