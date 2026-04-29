# Route To Module Map

## Scope
本文档列出当前前端可见主路由与其模块归属，并显式写出页面文件和主要后端连接。

## Route Mapping

| Route | Page File | Canonical Module | Primary Backend Link | Notes |
|---|---|---|---|---|
| `/login` | `frontend/src/core/views/Login/Login.jsx` | `identity-access-and-organization` | `/api/user/login`, `/api/user/refresh` | 独立于 `DefaultLayout`。 |
| `/user` | `frontend/src/core/views/UserPage/index.jsx` | `identity-access-and-organization` | `/api/user/info`, `/api/user/update/self` | 用户中心。 |
| `/system/user-manage` | `frontend/src/core/views/UserManage/index.tsx` | `identity-access-and-organization` | `/api/user/list`, `/api/user/create`, `/api/user/reset`, `/api/user/remove` | 管理用户。 |
| `/system/menu-manage` | `frontend/src/core/views/MenuManage/index.tsx` | `identity-access-and-organization` | `/api/permission/feature`, `/api/permission/menu/configure` | 菜单配置。 |
| `/system/role-manage` | `frontend/src/core/views/RoleManage/index.tsx` | `identity-access-and-organization` | `/api/role/roles`, `/api/role/create`, `/api/role/update`, `/api/role/delete` | 角色管理。 |
| `/system/team` | `frontend/src/core/views/Team/index.tsx` | `identity-access-and-organization` | `/api/team`, `/api/team/create`, `/api/team/update`, `/api/team/delete` | 团队管理。 |
| `/integration/alerts` | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/index.tsx` | `alert-sources-and-enrichment` | `/api/alertinput/*` | 告警来源接入与 enrich。 |
| `/integration/data` | `frontend/src/core/views/IntegrationCenter/DataIntegration/index.tsx` | `data-ingestion-and-oneagent` | `/api/integration/*` | 数据接入主入口。 |
| `/integration/data/settings` | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/index.tsx` | `data-ingestion-and-oneagent` | `/api/integration/*`, 外部 app-config 相关接口 | 安装配置和参数设置。 |
| `/system/data-group` | `frontend/src/core/views/DataGroup/index.tsx` | `data-scope-governance` | `/api/data/*`, `/api/v2/data/*` | 数据范围治理。 |
| `/demo/service-config` | `frontend/src/core/views/Config/index.tsx` | `platform-administration` | `/api/config/*` | 非主链/演示入口。 |
| `/service` | `frontend/src/oss/views/service/index.js` | `service-observability` | `/api/service/list`, `/api/service/redcharts`, `/api/service/alert/events` | service 列表与聚合观察。 |
| `/service/info` | `frontend/src/oss/views/serviceInfo/index.js` | `service-observability` | `/api/service/entry/endpoints`, `/api/service/relation`, `/api/service/topology`, `/api/service/error/instance` | service detail。 |
| `/logs/fault-site` | `frontend/src/oss/views/logs/FaultSiteLogsPage.tsx` | `logs-exploration-and-parsing` | `/api/log/fault/pagelist`, `/api/log/fault/content`, `/api/log/context` | 故障现场日志。 |
| `/logs/full` | `frontend/src/oss/views/logs/FullLogsPage.jsx` | `logs-exploration-and-parsing` | `/api/log/query`, `/api/log/chart`, `/api/log/index` | 全量日志搜索。 |
| `/trace/fault-site` | `frontend/src/oss/views/trace/FaultSiteTracePage.tsx` | `trace-exploration-and-profiling` | `/api/trace/pagelist`, `/api/trace/info` | 故障现场 trace。 |
| `/trace/full` | `frontend/src/oss/views/trace/FullTrace.jsx` | `trace-exploration-and-profiling` | `/api/trace/pagelist`, `/api/trace/pagelist/filters`, `/api/trace/pagelist/filter/value` | 全量 trace。 |
| `/system-dashboard` | `frontend/src/oss/views/dashboard/SystemDashboard.jsx` | `dashboards-and-health-monitoring` | `/api/metric/*`, `/api/service/monitor/status` | 系统级大盘。 |
| `/basic-dashboard` | `frontend/src/oss/views/dashboard/BasicDashboard.jsx` | `dashboards-and-health-monitoring` | `/api/metric/*` | 基础设施大盘。 |
| `/application-dashboard` | `frontend/src/oss/views/dashboard/ApplicationDashboard.jsx` | `dashboards-and-health-monitoring` | `/api/metric/*`, `/api/network/*` | 应用大盘。 |
| `/middleware-dashboard` | `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx` | `dashboards-and-health-monitoring` | `/api/metric/*` | 中间件大盘。 |
| `/alerts/rule` | `frontend/src/oss/views/alerts/AlertsRule.jsx` | `alert-rules-and-incident-events` | `/api/alerts/rule*`, `/api/alerts/rule/groups`, `/api/alerts/rule/metrics` | 告警规则。 |
| `/alerts/notify` | `frontend/src/oss/views/alerts/AlertsNotify.jsx` | `alert-rules-and-incident-events` | `/api/alerts/alertmanager/receiver/*` | 通知渠道。 |
| `/config` | `frontend/src/oss/views/config/index.jsx` | `platform-administration` | `/api/config/getTTL`, `/api/config/setTTL`, `/api/config/setSingleTableTTL` | 当前主配置入口。 |
| `/alerts/events` | `frontend/src/oss/views/alertEvents/index.tsx` | `alert-rules-and-incident-events` | `/api/alerts/event/list`, `/api/alerts/filter/*`, `/api/alerts/slient*`, `/api/alerts/resolve` | 事件列表。 |
| `/alerts/events/detail/:alertId/:eventId` | `frontend/src/oss/views/alertEvents/detail/index.tsx` | `alert-rules-and-incident-events` | `/api/alerts/event/detail`, `/api/alerts/events/report/get` | 事件详情。 |
| `/workflows` | `frontend/src/oss/views/workflows/index.tsx` | `automated-diagnosis-and-workflows` | 外部 `/dify/console/api/*` | workflow 嵌入与跳转。 |
| `/alert-analyze` | `frontend/src/pro/views/alertAnalyze/index.jsx` | `automated-diagnosis-and-workflows` | `/api/alerts/event/impact`, `/api/alerts/alert/events` | 自动化告警分析总览。 |
| `/alert-analyze/info` | `frontend/src/pro/views/serviceAnalyze/index.jsx` | `automated-diagnosis-and-workflows` | `/api/alerts/descendant/*`, `/api/alerts/anomaly-span/list`, `/api/alerts/detect/*` | 服务级分析详情。 |
| `/cause/report/:traceId/:spanId` | `frontend/src/pro/views/causeReport/index.tsx` | `conversational-diagnosis-and-reporting` | 外部 `/originx/api/*`, `/ai/api/*` | 根因报告视图。 |
| `/cause/error-report/:traceId` | `frontend/src/pro/views/causeReport/index.tsx` | `conversational-diagnosis-and-reporting` | 外部 `/originx/api/*`, `/ai/api/*` | 错误报告视图。 |
| `/mysql-dashboard` | `frontend/src/pro/views/MysqlDashboard/index.jsx` | `dashboards-and-health-monitoring` | `/api/metric/*` | MySQL 仪表盘。 |
| `/healthy-service` | `frontend/src/pro/views/serviceHealthy/index.js` | `dashboards-and-health-monitoring` | `/api/service/ryglight`, `/api/service/monitor/status` | 服务健康总览。 |
| `/alert-analyze/chatAI/:service/:endpoint` | `frontend/src/pro/views/chatAI/index.tsx` | `conversational-diagnosis-and-reporting` | 外部 `/ai/api/*` | 聊天式分析页面。 |
| `/report` | `frontend/src/pro/views/report/index.tsx` | `conversational-diagnosis-and-reporting` | `/api/alerts/events/report/get`, 外部 `/ai/api/*` | 报告消费页面，底层数据依赖告警事件报告接口。 |
| `/dataplane` | `frontend/src/pro/views/agentDataplane/index.tsx` | `dataplane-and-topology-governance` | `/api/dataplane/*` | 数据平面治理入口。 |
| `/system/platform-manage` | `frontend/src/pro/views/platformManage/index.tsx` | `platform-administration` | `/api/license/current`, `/api/license/upload`, `/api/app/logo`, `/api/app/upload` | 平台管理主页。 |
| `/agent` | `frontend/src/pro/views/agent/index.tsx` | `conversational-diagnosis-and-reporting` | 外部 `/ai/api/*` | Agent 交互页面。 |
| `/knowledge-bases` | `frontend/src/pro/views/knowledgeBases/index.tsx` | `ops-knowledge-base-management` | 外部 `/rag/api/kb/*` | 知识库管理页面；已通过后端 `knowledgeBases` 顶层菜单元数据接入侧边栏。 |
| `/probe-management` | `frontend/src/core/views/IntegrationCenter/DataIntegration/index.tsx` | `data-ingestion-and-oneagent` | `/api/integration/*` | 路由名称指向 ProbeInstaller，但组件实际复用 DataIntegration。 |
| `/probe-management/settings` | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/index.tsx` | `data-ingestion-and-oneagent` | `/api/integration/*` | Probe 管理配置页。 |

## System-Level Routes

| Route | Ownership | Notes |
|---|---|---|
| `/` | Shared app shell | `AppContent` 会根据后端返回的首个菜单项跳转。 |
| `*` | Shared app shell | 由 `DefaultLayout` 承载所有非登录页面。 |

## Coverage Check

- 当前主路由文件中可见的所有业务路径都已归属到业务模块。
- 未发现“已注册在主路由、但完全无模块归属”的页面。
- 但存在若干页面文件不在主路由链路中，详见 [page_inventory.md](page_inventory.md)。
