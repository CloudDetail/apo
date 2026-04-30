# Api To Module Map

## Scope
本文档按 API 家族而不是单文件目录来标注模块归属，避免把共享 route group 或外部代理误挂到错误模块。

## Business API Ownership

| API Family | Canonical Ownership | Frontend Wrappers | Backend Sources | Notes |
|---|---|---|---|---|
| `/api/user/*` | `identity-access-and-organization` | `frontend/src/core/api/user.js` | `backend/pkg/api/user`, `backend/pkg/services/user` | 登录、登出、刷新、用户管理。 |
| `/api/permission/*` | `identity-access-and-organization` | `frontend/src/core/api/permission.js` | `backend/pkg/api/permission`, `backend/pkg/services/permission` | 菜单、feature、router 权限。 |
| `/api/role/*` | `identity-access-and-organization` | `frontend/src/core/api/role.ts` | `backend/pkg/api/role`, `backend/pkg/services/role` | 角色 CRUD 与角色查询。 |
| `/api/team/*` | `identity-access-and-organization` | `frontend/src/core/api/team.ts` | `backend/pkg/api/team`, `backend/pkg/services/team` | 团队与团队用户管理。 |
| `/api/data/*`, `/api/v2/data/*` | `data-scope-governance` | `frontend/src/core/api/dataGroup.ts` | `backend/pkg/api/data`, `backend/pkg/services/data` | 数据范围和授权治理。 |
| `/api/service/list`, `/api/service/instances`, `/api/service/instance/*`, `/api/service/endpoint/*`, `/api/service/entry/endpoints`, `/api/service/relation`, `/api/service/topology`, `/api/service/error/*`, `/api/service/log/*`, `/api/service/trace/*`, `/api/service/sql/metrics`, `/api/service/redcharts` | `service-observability` | `frontend/src/core/api/service.js`, `serviceInfo.js`, `frontend/src/pro/api/service.js` | `backend/pkg/api/service`, `backend/pkg/services/service` | service 主观察链路。 |
| `/api/service/endpoints`, `/api/service/servicesAlert`, `/api/service/moreUrl`, `/api/service/getThreshold`, `/api/service/setThreshold`, `/api/service/ryglight`, `/api/service/monitor/status` | `dashboards-and-health-monitoring` | `frontend/src/core/api/service.js`, `frontend/src/pro/api/serviceHealthy.js` | `backend/pkg/api/serviceoverview`, `backend/pkg/services/serviceoverview` | 这些接口更偏聚合大盘与健康态。 |
| `/api/log/*` | `logs-exploration-and-parsing` | `frontend/src/core/api/logs.js` | `backend/pkg/api/log`, `backend/pkg/services/log` | 日志查询、规则、日志表管理。 |
| `/api/trace/*` | `trace-exploration-and-profiling` | `frontend/src/core/api/trace.js` | `backend/pkg/api/trace`, `backend/pkg/services/trace` | trace 查询与 profiling。 |
| `/api/alerts/rule*`, `/api/alerts/alertmanager/receiver*`, `/api/alerts/event/list`, `/api/alerts/event/detail`, `/api/alerts/events/classify`, `/api/alerts/slient*`, `/api/alerts/resolve`, `/api/alerts/filter/*`, `/api/alerts/events/report/*` | `alert-rules-and-incident-events` | `frontend/src/core/api/alerts.js`, `frontend/src/pro/api/report.ts` | `backend/pkg/api/alerts`, `backend/pkg/services/alerts` | 事件报告接口虽然被报告页面消费，但其存储与获取逻辑属于告警事件域。 |
| `/api/alerts/alert/events`, `/api/alerts/event/impact`, `/api/alerts/descendant/*`, `/api/alerts/anomaly-span/list`, `/api/alerts/detect/*` | `automated-diagnosis-and-workflows` | `frontend/src/pro/api/alertAnalyze.js` | `backend/pkg/api/alertanalyze`, `backend/pkg/services/alertanalyze` | 自动化分析和异常贡献度。 |
| `/api/alertinput/*` | `alert-sources-and-enrichment` | `frontend/src/core/api/alertInput.ts` | `backend/pkg/api/alertinput` | 告警源接入与 enrich。 |
| `/api/integration/*` | `data-ingestion-and-oneagent` | `frontend/src/core/api/integration.ts` | `backend/pkg/api/integration`, `backend/pkg/services/integration` | 集群、安装命令与安装配置。 |
| `/api/dataplane/*` | `dataplane-and-topology-governance` | `frontend/src/pro/api/dataplane.ts` | `backend/pkg/api/dataplane`, `backend/pkg/services/dataplane` | dataplane、服务名规则和拓扑治理。 |
| `/api/custom/*` | `conversational-diagnosis-and-reporting` | `frontend/src/pro/api/custom.ts` | `backend/pkg/api/custom`, `backend/pkg/services/custom` | 自定义查询与会话型展示的本仓库接口。 |
| `/api/config/*`, `/api/license/*`, `/api/app/logo`, `/api/app/upload`, `/api/app/deployMode` | `platform-administration` | `frontend/src/core/api/config.js`, `frontend/src/pro/api/license.ts`, `frontend/src/core/api/app.js` | `backend/pkg/api/config`, `license`, `app` | 平台设置、license、logo 上传与 deploy mode 读取。 |

## Shared Or Supporting API Ownership

| API Family | Recommended Ownership | Frontend Usage | Backend Sources | Notes |
|---|---|---|---|---|
| `/api/metric/*` | Shared infrastructure | dashboard、service、诊断等页面间接使用 | `backend/pkg/api/metric`, `backend/pkg/services/metric` | 支撑型指标查询接口。 |
| `/api/network/*` | Shared infrastructure | dashboard/健康/拓扑相关页面 | `backend/pkg/api/network`, `backend/pkg/services/network` | 支撑型网络分段接口。 |
| `/api/k8s/*` | Shared infrastructure | 供其他页面拉 namespace/pod 元数据 | `backend/pkg/api/k8s`, `backend/pkg/services/k8s` | 共享 K8s 元数据接口。 |
| `/api/health` | Shared infrastructure | 运维和部署探活 | `backend/pkg/api/health` | 不对应独立业务模块。 |

## External Or Sidecar API Ownership

| API Family | Recommended Ownership | Frontend Wrappers | Current Status |
|---|---|---|---|
| `/ai/api/*` | `external-service-boundaries` | `frontend/src/pro/api/chatAI.ts`, `conversation.ts`, `modelSetting.ts` | 当前仓库未见主后端注册，视为外部 AI 边界。 |
| `/rag/api/kb/*` | `external-service-boundaries` | `frontend/src/pro/api/knowledgeBase.ts`, `knowledgeBaseClient.ts` | 当前仓库未见主后端注册；开发态由 Vite `/rag` 代理转发到外部 rag-management，覆盖知识库管理、文档导入状态、上传前文档切片预览、已上传文档切片预览和原文件下载。 |
| `/originx/api/*` | `external-service-boundaries` | `frontend/src/pro/api/cause.ts` | 当前仓库未见主后端注册。 |
| `/dify/console/api/*` | `external-service-boundaries` | `frontend/src/core/api/workflows.ts` | workflow console 旁路接口。 |
| `/alert/api/v1/*` | `external-service-boundaries` | `frontend/src/core/api/alerts.js` 中历史调用 | 当前仓库未见主后端注册。 |

## Unmatched Frontend API Paths

| Frontend Path | Wrapper File | Current Status | Recommended Interpretation |
|---|---|---|---|
| `/api/permission/roles` | `frontend/src/core/api/permission.js` | 后端注册的是 `/api/role/roles` | 更像历史 wrapper 残留；当前未看到调用。 |

## Coverage Check

- 已注册的 backend route group 中，没有发现完全无归属的业务 API 家族。
- 未归入业务模块的 `/api/metric`、`/api/network`、`/api/k8s`、`/api/health` 已显式收口到共享基础设施。
- 真正的缺口主要来自 `/api/permission/roles` 这类历史 wrapper 漂移，以及明确的外部服务边界。
