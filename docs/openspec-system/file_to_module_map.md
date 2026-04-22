# File To Module Map

## Scope
本文档把关键源码目录和文件映射到规范化模块，作为 `repository-to-module-map.md` 的收敛版总表。

## Business Module Mapping

| Source Area | Canonical Module | Notes |
|---|---|---|
| `frontend/src/oss/views/service`, `frontend/src/oss/views/serviceInfo`, `frontend/src/core/api/service.js`, `frontend/src/core/api/serviceInfo.js`, `frontend/src/pro/api/service.js`, `backend/pkg/api/service`, `backend/pkg/api/serviceoverview`, `backend/pkg/services/service`, `backend/pkg/services/serviceoverview` | `service-observability` | service 观察与 service detail 共享同一业务域。 |
| `frontend/src/oss/views/logs`, `frontend/src/core/api/logs.js`, `backend/pkg/api/log`, `backend/pkg/services/log`, `backend/pkg/repository/database/dao_logtableinfo.go`, `dao_othertableinfo.go` | `logs-exploration-and-parsing` | 日志查询与解析规则必须合并看。 |
| `frontend/src/oss/views/trace`, `frontend/src/core/api/trace.js`, `backend/pkg/api/trace`, `backend/pkg/services/trace`, `backend/pkg/repository/jaeger` | `trace-exploration-and-profiling` | trace 和 profiling 共用同一后端 route group。 |
| `frontend/src/oss/views/dashboard`, `frontend/src/pro/views/MysqlDashboard`, `frontend/src/pro/views/serviceHealthy`, `backend/pkg/api/metric`, `backend/pkg/api/network`, `backend/pkg/services/metric`, `backend/pkg/services/network` | `dashboards-and-health-monitoring` | 聚合视图与健康视图归于同一模块。 |
| `frontend/src/oss/views/alerts`, `frontend/src/oss/views/alertEvents`, `frontend/src/core/api/alerts.js`, `backend/pkg/api/alerts`, `backend/pkg/services/alerts` | `alert-rules-and-incident-events` | 包括规则、通知、事件与事件报告接口。 |
| `frontend/src/core/views/IntegrationCenter/AlertsIntegration`, `frontend/src/core/api/alertInput.ts`, `backend/pkg/api/alertinput` | `alert-sources-and-enrichment` | 告警接入前置配置入口。 |
| `frontend/src/oss/views/workflows`, `frontend/src/pro/views/alertAnalyze`, `frontend/src/pro/views/serviceAnalyze`, `frontend/src/core/api/workflows.ts`, `frontend/src/pro/api/alertAnalyze.js`, `backend/pkg/api/alertanalyze`, `backend/pkg/services/alertanalyze` | `automated-diagnosis-and-workflows` | workflow 与自动化分析在代码中高度耦合。 |
| `frontend/src/pro/views/causeReport`, `frontend/src/pro/views/chatAI`, `frontend/src/pro/views/report`, `frontend/src/pro/views/agent`, `frontend/src/pro/api/chatAI.ts`, `conversation.ts`, `cause.ts`, `custom.ts`, `modelSetting.ts`, `report.ts`, `backend/pkg/api/custom`, `backend/pkg/services/custom` | `conversational-diagnosis-and-reporting` | 报告消费、会话式诊断和自定义查询统一归档。 |
| `frontend/src/core/views/IntegrationCenter/DataIntegration`, `frontend/src/core/api/integration.ts`, `frontend/src/core/api/appConfig.js`, `backend/pkg/api/integration`, `backend/pkg/services/integration`, `backend/static/integration-tmpl`, `one-agent/README.md` | `data-ingestion-and-oneagent` | 当前主仓库只呈现安装与接入侧聚合入口。 |
| `frontend/src/pro/views/agentDataplane`, `frontend/src/pro/api/dataplane.ts`, `backend/pkg/api/dataplane`, `backend/pkg/services/dataplane`, `backend/pkg/repository/dataplane` | `dataplane-and-topology-governance` | 数据平面与服务拓扑治理统一归属。 |
| `frontend/src/core/views/Login`, `UserPage`, `UserManage`, `MenuManage`, `RoleManage`, `Team`, `frontend/src/core/api/user.js`, `permission.js`, `role.ts`, `team.ts`, `backend/pkg/api/user`, `permission`, `role`, `team`, `backend/pkg/services/user`, `permission`, `role`, `team` | `identity-access-and-organization` | 身份、权限、菜单、组织模型完整闭环。 |
| `frontend/src/core/views/DataGroup`, `frontend/src/core/contexts/DataGroupContext.tsx`, `frontend/src/core/api/dataGroup.ts`, `backend/pkg/api/data`, `backend/pkg/services/data`, `backend/pkg/model/datagroup` | `data-scope-governance` | DataGroup 是共享治理能力，不是单页功能。 |
| `frontend/src/oss/views/config`, `frontend/src/pro/views/platformManage`, `frontend/src/pro/views/LicenseManage`, `frontend/src/core/api/config.js`, `frontend/src/core/api/app.js`, `frontend/src/pro/api/license.ts`, `backend/pkg/api/config`, `license`, `app` | `platform-administration` | 平台管理统一覆盖 TTL、license、logo 与 deploy mode。 |

## Shared Infrastructure Mapping

| Source Area | Canonical Shared Ownership | Notes |
|---|---|---|
| `frontend/src/index.js`, `frontend/src/App.js`, `frontend/src/routes.js`, `frontend/src/oss/routes.js`, `frontend/src/pro/routes.js`, `frontend/src/core/layout`, `frontend/src/core/components/AppSidebar.js`, `AppContent.js` | `frontend-app-shell` | 所有模块都通过这组文件装配到运行时。 |
| `frontend/src/core/utils/request.js`, `frontend/src/core/contexts`, `frontend/src/core/store`, `frontend/src/i18n.js`, `frontend/public/locales` | `frontend-state-request-and-i18n` | 统一请求、状态与文案。 |
| `backend/main.go`, `backend/pkg/router`, `backend/pkg/core`, `backend/pkg/middleware` | `backend-http-kernel` | 后端启动与中间件总线。 |
| `backend/pkg/repository/database`, `clickhouse`, `prometheus`, `jaeger`, `kubernetes`, `dify`, `cache`, `dataplane`, `polarisanalyzer` | `backend-persistence-and-connectors` | 数据访问与外部系统连接层。 |
| `backend/config`, `backend/pkg/logger`, `backend/Dockerfile` | `configuration-logging-and-bootstrap` | 配置、日志、构建与启动。 |
| `backend/static`, `backend/sqlscripts`, `frontend/public`, `one-agent/README.md` | `deployment-assets-and-static-templates` | 安装模板和初始化资源。 |

## Cross-Cutting Mapping

| Source Area | Canonical Cross-Cutting Ownership | Notes |
|---|---|---|
| `frontend/src/core/views/Login`, `frontend/src/core/utils/request.js`, `backend/pkg/middleware/auth_middleware.go`, `backend/pkg/util/jwt` | `auth-and-session` | 会话与 token 机制不归单个业务模块。 |
| `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/components/AppSidebar.js`, `backend/pkg/services/permission/service_user_configs.go`, `backend/pkg/repository/database/init_feature.go`, `init_router.go`, `init_menu_item.go`, `init_feature_mapping.go`, `init_router_page.go` | `permission-menu-router-model` | 菜单与可访问路由由后端元数据驱动。 |
| `frontend/src/pro/contexts/LicenseContext.tsx`, `frontend/src/core/layout/DefaultLayout.js`, `backend/pkg/middleware/lic_middleware.go`, `backend/pkg/services/license` | `license-and-edition-gating` | 影响多个业务模块访问。 |
| `frontend/src/core/contexts/DataGroupContext.tsx`, Redux time range reducer, `backend/pkg/services/data` | `data-group-and-time-range-context` | 多模块共享查询上下文。 |
| `frontend/src/pro/api/*`, `frontend/src/core/api/workflows.ts`, `frontend/vite.config.mjs`, `backend/config/apo.yml` 中外部服务配置 | `external-service-boundaries` | 区分本仓库实现与旁路服务。 |

## Exceptional Or Unclear Files

| File Or Area | Current Status | Recommended Handling |
|---|---|---|
| `frontend/src/oss/views/camera/index.js` | 未看到主路由或其他页面引用 | 暂列为未接入主链的历史页面。 |
| `frontend/src/pro/views/probeInstaller/index.tsx` | 文件存在，但主路由未引用；`/probe-management` 当前复用 `IntegrationCenter/DataIntegration` | 归入 `data-ingestion-and-oneagent`，并标注为历史或待确认页面。 |
| `frontend/src/core/views/Config/index.tsx` | 通过 `/demo/service-config` 进入，而主平台配置入口是 `/config` | 归入 `platform-administration`，但标为非主链/演示入口。 |

## Source Ownership Exceptions
以下表格用于同时说明“模块归属”和“源码归属”：

| Source Area | Canonical Ownership | Source Classification | Notes |
|---|---|---|---|
| `frontend/src/pro/**/*` | 多个 Pro 业务模块 | `commercial-only` | 覆盖会话诊断、dataplane、license、平台管理、部分自动分析与健康视图。 |
| `frontend/public/locales/*/pro/*` | 多个 Pro 业务模块 | `commercial-only` | Pro 页面语言资源，仅商业仓存在。 |
| `frontend/src/App.js`, `frontend/src/index.js`, `frontend/src/routes.js`, `frontend/src/_nav.js` | `frontend-app-shell` | `commercial-variant` | 根层共享壳文件已偏离上游。 |
| `frontend/src/constants.js`, `frontend/src/i18n.js`, `frontend/src/core/utils/request.js` | `frontend-state-request-and-i18n` | `commercial-variant` | 共享请求层与国际化层已有本地分叉。 |
| `frontend/src/core/layout/DefaultLayout.js`, `frontend/src/core/components/AppHeader.js`, `frontend/src/core/components/AppSidebar.js` | `frontend-app-shell` + cross-cutting concerns | `commercial-variant` | 共享壳层同时承载 branding、license、菜单裁剪。 |
| `frontend/src/core/components/AuthRouter.tsx` | `permission-menu-router-model` | `commercial-variant` | 前端 route guard 当前被弱化。 |
| `frontend/src/core/api/app.js`, `frontend/src/core/contexts/DeployModeContext.js` | `platform-administration` | `commercial-only` | 共享路径中的商业新增文件。 |
| `frontend/src/core/api/appConfig.js`, `frontend/src/core/views/IntegrationCenter/DataIntegration/BotManagement.tsx`, `frontend/src/core/components/Integration/*` | `data-ingestion-and-oneagent` | `commercial-only` | 共享路径中的商业接入增强。 |
| `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx` | `dashboards-and-health-monitoring` | `commercial-variant` | `oss` 路径也已与上游不同。 |
| `backend/pkg/api/alertanalyze`, `backend/pkg/services/alertanalyze` | `automated-diagnosis-and-workflows` | `commercial-only` | 商业仓独有的自动分析后端包。 |
| `backend/pkg/api/license`, `backend/pkg/services/license`, `backend/pkg/model/license`, `backend/pkg/repository/database/license` | `platform-administration` | `commercial-only` | license 全链路。 |
| `backend/pkg/api/app`, `backend/pkg/services/app`, `backend/pkg/repository/database/app` | `platform-administration` | `commercial-only` | logo / app 资源管理链路。 |
| `backend/pkg/api/custom`, `backend/pkg/services/custom` | `conversational-diagnosis-and-reporting` | `commercial-only` | 自定义查询后端能力。 |
| `backend/pkg/router/router_api.go`, `backend/pkg/middleware/middleware.go`, `backend/pkg/middleware/lic_middleware.go` | `backend-http-kernel` + cross-cutting concerns | `mixed shared-package + commercial extension` | 后端共享内核在商业仓中承担 route 扩展与 license 门控。 |
| `backend/pkg/api/alerts`, `backend/pkg/services/alerts` | `alert-rules-and-incident-events` | `commercial-variant` | 共享 alerts 包已增加事件报告接口。 |
| `backend/pkg/api/dataplane`, `backend/pkg/services/dataplane`, `backend/pkg/model/*dataplane*` | `dataplane-and-topology-governance` | `mixed` | 共享 dataplane 包上叠加了更多商业治理接口。 |

完整规则见 [source_governance.md](source_governance.md)。
