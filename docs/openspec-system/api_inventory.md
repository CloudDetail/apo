# Api Inventory

## Scope
本文档盘点当前后端已注册 API 家族，以及前端显式依赖但不在主后端注册中的外部或缺口接口。

## Registered Backend API Families

| API Family | Representative Endpoints | Ownership | Backend Sources |
|---|---|---|---|
| `/api/user` | `/login`, `/logout`, `/refresh`, `/info`, `/list`, `/team`, `/create`, `/update/*`, `/remove`, `/reset` | `identity-access-and-organization` | `backend/pkg/api/user`, `backend/pkg/services/user` |
| `/api/permission` | `/config`, `/feature`, `/router`, `/sub/feature`, `/operation`, `/menu/configure` | `identity-access-and-organization` | `backend/pkg/api/permission`, `backend/pkg/services/permission` |
| `/api/role` | `/roles`, `/user`, `/operation`, `/create`, `/update`, `/delete` | `identity-access-and-organization` | `backend/pkg/api/role`, `backend/pkg/services/role` |
| `/api/team` | `GET /`, `/user`, `/create`, `/update`, `/delete`, `/operation`, `/user/operation` | `identity-access-and-organization` | `backend/pkg/api/team`, `backend/pkg/services/team` |
| `/api/data` | `/datasource`, `/group`, `/group/data`, `/sub/group`, `/user/group`, `/group/operation`, `/subs`, `/subs/operation` | `data-scope-governance` | `backend/pkg/api/data`, `backend/pkg/services/data` |
| `/api/v2/data` | `/group`, `/group/datasource/list`, `/group/detail`, `/group/add`, `/group/update`, `/group/delete`, `/group/filter`, `/group/datasource/refresh` | `data-scope-governance` | `backend/pkg/api/data`, `backend/pkg/services/data` |
| `/api/service` | `/list`, `/instances`, `/instance/list`, `/endpoint/list`, `/entry/endpoints`, `/relation`, `/topology`, `/error/instance`, `/log/*`, `/trace/*`, `/sql/metrics`, `/redcharts`, `/servicesAlert`, `/ryglight`, `/monitor/status` | `service-observability` + `dashboards-and-health-monitoring` | `backend/pkg/api/service`, `backend/pkg/api/serviceoverview` |
| `/api/log` | `/fault/pagelist`, `/fault/content`, `/context`, `/query`, `/chart`, `/index`, `/table`, `/rule/*`, `/other/*` | `logs-exploration-and-parsing` | `backend/pkg/api/log`, `backend/pkg/services/log` |
| `/api/trace` | `/onoffcpu`, `/flame`, `/flame/process`, `/pagelist`, `/pagelist/filters`, `/pagelist/filter/value`, `/info` | `trace-exploration-and-profiling` | `backend/pkg/api/trace`, `backend/pkg/services/trace` |
| `/api/alerts` | `/event/list`, `/event/detail`, `/events/classify`, `/events/report/get`, `/events/report/add`, `/inputs/alertmanager`, `/outputs/dingtalk/:uuid`, `/rule/*`, `/alertmanager/receiver/*`, `/slient*`, `/resolve`, `/filter/*`, 以及 impact/descendant/detect 系列 | `alert-rules-and-incident-events` + `automated-diagnosis-and-workflows` | `backend/pkg/api/alerts`, `backend/pkg/api/alertanalyze` |
| `/api/alertinput` | `/event/source`, `/event/json`, `/source/*`, `/cluster/*`, `/schema/*`, `/enrich/*` | `alert-sources-and-enrichment` | `backend/pkg/api/alertinput` |
| `/api/integration` | `/configuration`, `/vars/:variable`, `/cluster/*`, `/cluster/install/*`, `/adapter/update` | `data-ingestion-and-oneagent` | `backend/pkg/api/integration`, `backend/pkg/services/integration` |
| `/api/dataplane` | `/services`, `/redcharts`, `/endpoints`, `/instances`, `/servicename*`, `/topology`, `/checkDataSource`, `/customtopology/*`, `/type/list`, `/list`, `/create`, `/update`, `/delete`, `/data/check`, `/appinfo/tags*`, `/datasources`, `/servicesWithInstances`, `/capabilities` | `dataplane-and-topology-governance` | `backend/pkg/api/dataplane`, `backend/pkg/services/dataplane` |
| `/api/custom` | `/create`, `/update`, `/get`, `/delete` | `conversational-diagnosis-and-reporting` | `backend/pkg/api/custom`, `backend/pkg/services/custom` |
| `/api/config` | `/setTTL`, `/setSingleTableTTL`, `/getTTL` | `platform-administration` | `backend/pkg/api/config`, `backend/pkg/services/config` |
| `/api/license` | `/upload`, `GET /`, `/current` | `platform-administration` | `backend/pkg/api/license`, `backend/pkg/services/license` |
| `/api/app` | `/upload`, `/logo`, `/deployMode` | `platform-administration` | `backend/pkg/api/app`, `backend/pkg/services/app` |
| `/api/metric` | `/list`, `/query`, `/queryPods` | Shared infrastructure | `backend/pkg/api/metric`, `backend/pkg/services/metric` |
| `/api/network` | `/podmap`, `/segments` | Shared infrastructure | `backend/pkg/api/network`, `backend/pkg/services/network` |
| `/api/k8s` | `/namespaces`, `/namespace/info`, `/pods`, `/pod/info` | Shared infrastructure | `backend/pkg/api/k8s`, `backend/pkg/services/k8s` |
| `/api/health` | `GET /api/health` | Shared infrastructure | `backend/pkg/api/health` |

## Frontend-Visible External Or Missing API Families

| API Family | Observed Frontend Sources | Status |
|---|---|---|
| `/ai/api/*` | `frontend/src/pro/api/chatAI.ts`, `conversation.ts`, `modelSetting.ts` | 外部或旁路 AI 服务。 |
| `/originx/api/*` | `frontend/src/pro/api/cause.ts` | 外部根因分析边界。 |
| `/dify/console/api/*` | `frontend/src/core/api/workflows.ts` | 外部 workflow console 边界。 |
| `/alert/api/v1/*` | `frontend/src/core/api/alerts.js` | 历史或旁路告警接口边界。 |
| `/api/permission/roles` | `frontend/src/core/api/permission.js` | 前端 wrapper 存在，主后端注册是 `/api/role/roles`。 |

## Coverage Check

- 已注册的后端 API 家族都已归属到业务模块或共享基础设施。
- 未发现“在 `router_api.go` 中注册，但没有文档归属”的 route group。
- 仍需特别关注前端 wrapper 与后端注册不一致的 `/api/permission/roles`，以及四类外部 API 边界。
