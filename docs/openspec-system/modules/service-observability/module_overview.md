# Service Observability Module Overview

## Module Goal
本模块是整个系统的主观测入口，围绕 `service` 与 `endpoint` 提供从总览到细查的完整链路。用户先在 `/service` 浏览服务与核心 endpoint 的 RED 摘要，再进入 `/service/info` 查看拓扑、依赖、入口影响、实例、日志、Trace、K8s、SQL 与 Polaris 诊断结果。

## Scope
当前代码证据支持以下范围：

- 服务总览页 `/service`
- 服务详情页 `/service/info`
- service 下全量 endpoint 展开弹窗
- 直接上下游拓扑与全量拓扑
- 下游依赖相似度排行与时序对比
- 入口影响分析
- 实例、错误实例、日志、Trace、K8s、SQL、Polaris 详情面板
- 服务列表与 endpoint 列表的小型 RED 图批量补图

当前代码不应纳入本模块主体范围的能力：

- `/healthy-service` 页面及其 `/api/service/ryglight`、`/api/service/monitor/status` 消费链路
- `/logs/*` 与 `/trace/*` 的完整检索页主体逻辑
- 告警规则与告警事件治理页面
- `alertAnalyze`、`serviceAnalyze` 等 AI/告警分析页面

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/oss/routes.js` | 注册 `/service` 与 `/service/info`。 |
| Page | `frontend/src/oss/views/service/index.js` | 服务总览页入口。 |
| Page | `frontend/src/oss/views/serviceInfo/index.js` | 服务详情页入口。 |
| Context | `frontend/src/oss/contexts/ServiceInfoContext.tsx` | 服务详情页内部折叠面板状态与 dashboard 变量。 |
| Shared Context | `frontend/src/core/contexts/ChartsContext.tsx` | 服务列表和 endpoint 弹窗批量折线图数据。 |
| Shared Context | `frontend/src/core/contexts/PropsContext.js` | 把 `serviceName`、`endpoint`、`clusterIds` 下发给详情子组件。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | 统一把本模块接口挂在 `/api/service` 下。 |
| Handler Package | `backend/pkg/api/serviceoverview` | 列表侧能力：`endpoints`、`servicesAlert`、`moreUrl`、threshold。 |
| Handler Package | `backend/pkg/api/service` | 详情侧能力：拓扑、依赖、实例、日志、Trace、K8s、SQL、入口影响、Polaris。 |
| Service Package | `backend/pkg/services/serviceoverview` | 服务总览聚合逻辑。 |
| Service Package | `backend/pkg/services/service` | 服务详情与诊断逻辑。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `service` | 用户在 `/service` 页面首先感知的观测主体，通常对应应用或服务。 |
| `endpoint` | 服务内部的具体入口或调用点，是列表展开与详情页分析的关键粒度。 |
| `RED` | 延时、错误率、吞吐在当前模块中是所有列表与详情的核心指标维度。 |
| `DelaySource` | 当前 endpoint 的主要延时来源，前端渲染为 `self`、`dependency`、`unknown`。 |
| `Topology` | 既有直接上下游拓扑，也有经 `relation` 扩展后的全量调用关系。 |
| `Descendant` | 指当前 endpoint 下游依赖节点，用于相似度排行与时序比较。 |
| `Entry Endpoint` | 指当前 service/endpoint 受到哪些入口流量路径影响。 |
| `Instance` | service 的实际运行实例，页面会补充节点名、IP、部署/重启时间和状态。 |
| `Error Instance` | 错误实例及其传播链，支持故障现场日志跳转。 |
| `DataGroup` | 决定可见 cluster/namespace/service 范围的共享治理上下文。 |

## Related Files

- `frontend/src/oss/views/service/ServiceTable.jsx`
- `frontend/src/oss/views/service/component/EndpointTableModal.jsx`
- `frontend/src/oss/views/serviceInfo/index.js`
- `frontend/src/oss/views/serviceInfo/component/dependent/DependentTable.jsx`
- `frontend/src/oss/views/serviceInfo/component/dependent/TopologyModal.jsx`
- `frontend/src/oss/views/serviceInfo/component/infoUni/EntryImpact.jsx`
- `frontend/src/oss/views/serviceInfo/component/infoUni/InstanceInfo.jsx`
- `frontend/src/oss/views/serviceInfo/component/infoUni/LogsInfo.jsx`
- `frontend/src/oss/views/serviceInfo/component/infoUni/TraceInfo.jsx`
- `frontend/src/oss/views/serviceInfo/component/infoUni/ErrorInstance/ErrorInstanceInfo.jsx`
- `frontend/src/core/api/service.js`
- `frontend/src/core/api/serviceInfo.js`
- `backend/pkg/api/serviceoverview/func_getendpointsdata.go`
- `backend/pkg/api/service/func_getserviceendpointtopology.go`
- `backend/pkg/api/service/func_getserviceentryendpoints.go`
- `backend/pkg/services/serviceoverview/service_get_endpoints_data.go`
- `backend/pkg/services/service/service_get_service_endpoint_relation.go`
- `backend/pkg/services/service/service_instance_new.go`
- `backend/pkg/model/request/service_req.go`
- `backend/pkg/model/response/service_resp.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `timeRangeReducer` | 所有查询参数都依赖统一时间范围。 |
| Frontend Shared | `DataSourceFilter` + `getDatasourceFilterApiV2` | `/service` 页面筛选器由数据组树提供，而不是本模块自造字典。 |
| Frontend Shared | `UserContext` + `AppSidebar` | `/service` 菜单不是静态硬编码，而是权限返回结果。 |
| Frontend Shared | `AuthRouter` | 页面被 guard 包装，但当前真正的前端路由校验逻辑被注释。 |
| Backend Shared | `data.Service` | 负责 `CheckGroupPermission` 和 `CheckServicesPermission`。 |
| Backend Shared | `common.GetPQLFilterByGroupID` | 把 `groupId` 转为 Prometheus 过滤条件。 |
| Repository | `prometheus.Repo` | 承担绝大多数 RED、日志计数、实例与 monitor 指标查询。 |
| Repository | `clickhouse.Repo` | 承担拓扑、入口关系、故障现场日志、K8s 事件、部分实例来源。 |
| Repository | `polarisanalyzer.Repo` | 承担下游相似度与 Polaris 分析。 |
| Repository | `database.Repo` | 承担 threshold、数据组、权限范围查询。 |
| Cross Module | `logs` / `trace` 模块 | 详情页时间线最终会跳转到 `#/logs/fault-site` 或 `#/trace/fault-site`。 |

## Risks And Unknowns

1. `frontend/src/core/api/service.js` 中的 `getServiceTableApi -> api/service/top3` 未在当前后端路由中出现，存在遗留接口漂移。
2. `frontend/src/oss/views/service/component/ThresholdConfigModal.jsx` 当前未接线，不能按已实现功能描述为“可配置阈值界面”。
3. `backend/pkg/services/serviceoverview/service_set_threshold.go` 的写入逻辑没有把 `serviceName` 与 `endPoint` 写入 `Threshold` 模型，且仅在 `level == global` 时赋值 `Level`，语义需要 `To Be Confirmed`。
4. `/api/service` route group 同时容纳了本模块和健康监控相关接口，文档必须显式切边，否则容易把 `serviceHealthy` 的能力误并入本模块。
