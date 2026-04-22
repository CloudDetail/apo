# Page Inventory

## Scope
本文档盘点当前前端“页面级入口”，优先覆盖主路由链上的页面，并单独列出非主链或疑似历史页面。

## Routed Pages

| Route | Page File | Module | Status |
|---|---|---|---|
| `/login` | `frontend/src/core/views/Login/Login.jsx` | `identity-access-and-organization` | Active |
| `/user` | `frontend/src/core/views/UserPage/index.jsx` | `identity-access-and-organization` | Active |
| `/system/user-manage` | `frontend/src/core/views/UserManage/index.tsx` | `identity-access-and-organization` | Active |
| `/system/menu-manage` | `frontend/src/core/views/MenuManage/index.tsx` | `identity-access-and-organization` | Active |
| `/system/role-manage` | `frontend/src/core/views/RoleManage/index.tsx` | `identity-access-and-organization` | Active |
| `/system/team` | `frontend/src/core/views/Team/index.tsx` | `identity-access-and-organization` | Active |
| `/integration/alerts` | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/index.tsx` | `alert-sources-and-enrichment` | Active |
| `/integration/data` | `frontend/src/core/views/IntegrationCenter/DataIntegration/index.tsx` | `data-ingestion-and-oneagent` | Active |
| `/integration/data/settings` | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/index.tsx` | `data-ingestion-and-oneagent` | Active |
| `/system/data-group` | `frontend/src/core/views/DataGroup/index.tsx` | `data-scope-governance` | Active |
| `/demo/service-config` | `frontend/src/core/views/Config/index.tsx` | `platform-administration` | Non-canonical route |
| `/service` | `frontend/src/oss/views/service/index.js` | `service-observability` | Active |
| `/service/info` | `frontend/src/oss/views/serviceInfo/index.js` | `service-observability` | Active |
| `/logs/fault-site` | `frontend/src/oss/views/logs/FaultSiteLogsPage.tsx` | `logs-exploration-and-parsing` | Active |
| `/logs/full` | `frontend/src/oss/views/logs/FullLogsPage.jsx` | `logs-exploration-and-parsing` | Active |
| `/trace/fault-site` | `frontend/src/oss/views/trace/FaultSiteTracePage.tsx` | `trace-exploration-and-profiling` | Active |
| `/trace/full` | `frontend/src/oss/views/trace/FullTrace.jsx` | `trace-exploration-and-profiling` | Active |
| `/system-dashboard` | `frontend/src/oss/views/dashboard/SystemDashboard.jsx` | `dashboards-and-health-monitoring` | Active |
| `/basic-dashboard` | `frontend/src/oss/views/dashboard/BasicDashboard.jsx` | `dashboards-and-health-monitoring` | Active |
| `/application-dashboard` | `frontend/src/oss/views/dashboard/ApplicationDashboard.jsx` | `dashboards-and-health-monitoring` | Active |
| `/middleware-dashboard` | `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx` | `dashboards-and-health-monitoring` | Active |
| `/alerts/rule` | `frontend/src/oss/views/alerts/AlertsRule.jsx` | `alert-rules-and-incident-events` | Active |
| `/alerts/notify` | `frontend/src/oss/views/alerts/AlertsNotify.jsx` | `alert-rules-and-incident-events` | Active |
| `/config` | `frontend/src/oss/views/config/index.jsx` | `platform-administration` | Active |
| `/alerts/events` | `frontend/src/oss/views/alertEvents/index.tsx` | `alert-rules-and-incident-events` | Active |
| `/alerts/events/detail/:alertId/:eventId` | `frontend/src/oss/views/alertEvents/detail/index.tsx` | `alert-rules-and-incident-events` | Active |
| `/workflows` | `frontend/src/oss/views/workflows/index.tsx` | `automated-diagnosis-and-workflows` | Active |
| `/alert-analyze` | `frontend/src/pro/views/alertAnalyze/index.jsx` | `automated-diagnosis-and-workflows` | Active |
| `/alert-analyze/info` | `frontend/src/pro/views/serviceAnalyze/index.jsx` | `automated-diagnosis-and-workflows` | Active |
| `/cause/report/:traceId/:spanId` | `frontend/src/pro/views/causeReport/index.tsx` | `conversational-diagnosis-and-reporting` | Active |
| `/cause/error-report/:traceId` | `frontend/src/pro/views/causeReport/index.tsx` | `conversational-diagnosis-and-reporting` | Active |
| `/mysql-dashboard` | `frontend/src/pro/views/MysqlDashboard/index.jsx` | `dashboards-and-health-monitoring` | Active |
| `/healthy-service` | `frontend/src/pro/views/serviceHealthy/index.js` | `dashboards-and-health-monitoring` | Active |
| `/alert-analyze/chatAI/:service/:endpoint` | `frontend/src/pro/views/chatAI/index.tsx` | `conversational-diagnosis-and-reporting` | Active |
| `/report` | `frontend/src/pro/views/report/index.tsx` | `conversational-diagnosis-and-reporting` | Active |
| `/dataplane` | `frontend/src/pro/views/agentDataplane/index.tsx` | `dataplane-and-topology-governance` | Active |
| `/system/platform-manage` | `frontend/src/pro/views/platformManage/index.tsx` | `platform-administration` | Active |
| `/agent` | `frontend/src/pro/views/agent/index.tsx` | `conversational-diagnosis-and-reporting` | Active |
| `/probe-management` | `frontend/src/core/views/IntegrationCenter/DataIntegration/index.tsx` | `data-ingestion-and-oneagent` | Active alias route |
| `/probe-management/settings` | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/index.tsx` | `data-ingestion-and-oneagent` | Active alias route |

## Nested Or Embedded Page Roots

| Page File | Module | Why It Is Not Listed As Top-Level Route |
|---|---|---|
| `frontend/src/pro/views/LicenseManage/index.tsx` | `platform-administration` | 当前由 `/system/platform-manage` 内部 `Collapse` 渲染，而不是单独路由。 |
| `frontend/src/oss/views/alertEvents/detail/index.tsx` | `alert-rules-and-incident-events` | 已有显式动态路由，但常作为 detail page 嵌套使用。 |

## Non-Routed Or Suspicious Page-Like Files

| Page File | Current Status | Recommended Ownership |
|---|---|---|
| `frontend/src/oss/views/camera/index.js` | 未发现路由引用或组件引用 | 暂列为未接入主链的历史页面，待确认是否废弃。 |
| `frontend/src/pro/views/probeInstaller/index.tsx` | 页面文件存在，但 `proRoutes` 未使用 | 归入 `data-ingestion-and-oneagent`，标记为历史或替代实现。 |
| `frontend/src/core/views/Config/index.tsx` | 仅通过 `/demo/service-config` 暴露，不是主配置入口 | 归入 `platform-administration`，标记为演示或旧入口。 |

## Coverage Check

- 主路由链中的页面全部已归属到业务模块。
- 没有发现“已在主路由注册、但未归属任何模块”的页面。
- 需要后续重点核查的是未接入主链的页面文件，以及同一业务能力存在两个 UI 入口实现的情况。
