# Service Observability Source Map

## Scope
本文档先做证据提取，不直接扩写业务结论。目标是把 `service-observability` 模块在当前仓库中的真实代码落点固定下来，作为后续模块文档的依据。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/oss/views/service` | `/service` 页面主目录，包含服务总览表和 endpoint 展开弹窗。 |
| Frontend | `frontend/src/oss/views/serviceInfo` | `/service/info` 页面主目录，包含拓扑、依赖、实例、日志、Trace、K8s、SQL、告警等详情面板。 |
| Frontend | `frontend/src/oss/contexts/ServiceInfoContext.tsx` | `serviceInfo` 页面内部状态容器，管理面板状态、展开项和 dashboard 变量。 |
| Frontend | `frontend/src/core/api/service.js` | 服务总览页和批量 RED 图表相关 API wrapper。 |
| Frontend | `frontend/src/core/api/serviceInfo.js` | 服务详情页相关 API wrapper。 |
| Frontend | `frontend/src/core/components/Filter/DataSourceFilter.tsx` | `/service` 页面筛选器，数据来源不是本模块 API，而是数据组过滤树。 |
| Frontend | `frontend/src/core/contexts/ChartsContext.tsx` | 批量 RED 折线图上下文，调用 `/api/service/redcharts`。 |
| Frontend | `frontend/src/core/contexts/PropsContext.js` | 在 `serviceInfo` 各子面板之间传递 `serviceName`、`endpoint`、`clusterIds`。 |
| Frontend | `frontend/src/core/store/reducers/timeRangeReducer.js` | 所有服务观测查询共享时间范围。 |
| Frontend | `frontend/src/core/store/reducers/topologyReducer.js` | 详情页全拓扑弹窗的 drill-down 状态。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/service/*` 路由注册入口。 |
| Backend | `backend/pkg/api/service` | 服务详情、拓扑、实例、日志、Trace、SQL、入口影响等 handler。 |
| Backend | `backend/pkg/api/serviceoverview` | 服务列表、`moreUrl`、告警摘要、threshold 等 handler。 |
| Backend | `backend/pkg/services/service` | 详情侧核心 service 实现。 |
| Backend | `backend/pkg/services/serviceoverview` | 列表侧核心 service 实现。 |
| Backend | `backend/pkg/services/data` | 数据组和服务可见性校验。 |
| Backend | `backend/pkg/services/common` | `GetPQLFilterByGroupID`、`MarkTopologyNodeInGroup` 等跨模块过滤逻辑。 |
| Backend | `backend/pkg/model/request` | 服务模块请求结构体。 |
| Backend | `backend/pkg/model/response` | 服务模块响应结构体。 |
| Backend | `backend/pkg/model` | `TopologyNode`、`ServiceInstance` 等领域结构。 |
| Backend | `backend/pkg/repository/database` | threshold 表与权限/数据组相关 DAO。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/oss/routes.js` | `/service`, `/service/info` |
| Backend | `backend/pkg/router/router_api.go` | `/api/service/endpoints`, `/api/service/servicesAlert`, `/api/service/moreUrl`, `/api/service/topology`, `/api/service/relation`, `/api/service/descendant/metrics`, `/api/service/descendant/relevance`, `/api/service/polaris/infer`, `/api/service/error/instance`, `/api/service/errorinstance/logs`, `/api/service/log/metrics`, `/api/service/log/logs`, `/api/service/trace/metrics`, `/api/service/trace/logs`, `/api/service/instances`, `/api/service/instance/list`, `/api/service/instance/options`, `/api/service/endpoint/list`, `/api/service/k8s/events/count`, `/api/service/alert/sample/events`, `/api/service/sql/metrics`, `/api/service/entry/endpoints`, `/api/service/redcharts` |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/oss/views/service/index.js` | `/service` 页面装配 `ChartsProvider` 和 `ServiceTable`。 |
| Route Page | `frontend/src/oss/views/serviceInfo/index.js` | `/service/info` 页面入口，读取 query 参数后拉取拓扑并装配详情上下文。 |
| Page Core | `frontend/src/oss/views/service/ServiceTable.jsx` | 服务列表主表、筛选、排序、跳转和批量图表拉取。 |
| Page Core | `frontend/src/oss/views/service/component/EndpointTableModal.jsx` | 某个 service 的完整 endpoint 列表弹窗。 |
| Page Core | `frontend/src/oss/views/serviceInfo/component/dependent/index.jsx` | 依赖分析区域，包含相似度排行和时序对比。 |
| Page Core | `frontend/src/oss/views/serviceInfo/component/dependent/DependentTable.jsx` | 下游依赖相似度排行。 |
| Page Core | `frontend/src/oss/views/serviceInfo/component/dependent/TimelapseLineChart.jsx` | 下游依赖延时对比图。 |
| Page Core | `frontend/src/oss/views/serviceInfo/component/dependent/TopologyModal.jsx` | 全拓扑弹窗。 |
| Page Core | `frontend/src/oss/views/serviceInfo/component/infoUni/index.jsx` | 详情页折叠面板总装配。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/EntryImpact.jsx` | 入口影响面板。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/AlertInfo/AlertInfoTabs.jsx` | 告警样本面板。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/InstanceInfo.jsx` | 实例观测面板。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/K8sInfo.jsx` | K8s 事件面板。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/ErrorInstance/ErrorInstanceInfo.jsx` | 错误实例面板。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/LogsInfo.jsx` | 日志指标与故障现场日志时间线。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/TraceInfo.jsx` | Trace 指标与故障现场 Trace 时间线。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/SqlMetrics.jsx` | SQL 指标面板。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/PolarisMetricsInfo/PolarisMetricsInfo.jsx` | Polaris 诊断图片和结论。 |
| Detail Panel | `frontend/src/oss/views/serviceInfo/component/infoUni/MetricsDashboard.jsx` | Grafana iframe dashboard。 |
| Shared Panel Helper | `frontend/src/oss/views/serviceInfo/component/infoUni/TimeLine.jsx` | 通过 `TimeLineTypeApiMap` 复用日志/Trace/错误日志接口。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/service.js` | `getServicesEndpointsApi` | `/api/service/endpoints` | `/service` 主表。 |
| `frontend/src/core/api/service.js` | `getServicesAlertApi` | `/api/service/servicesAlert` | 列表页每个 service 的日志/基础设施/网络/K8s 状态摘要。 |
| `frontend/src/core/api/service.js` | `getEndpointTableApi` | `/api/service/moreUrl` | endpoint 展开弹窗。 |
| `frontend/src/core/api/service.js` | `getServiceChartsApi` | `/api/service/redcharts` | 列表页和 endpoint 弹窗批量 RED 图。 |
| `frontend/src/core/api/service.js` | `getServiceRelationApi` | `/api/service/relation` | 全拓扑弹窗。 |
| `frontend/src/core/api/service.js` | `getServiceListApi` | `/api/service/list` | 当前模块主页面未直接调用，但属于服务观测数据源能力。 |
| `frontend/src/core/api/service.js` | `getServiceInstanceListApi` | `/api/service/instance/list` | 当前模块主页面未直接调用。 |
| `frontend/src/core/api/service.js` | `getServiceInstancOptionsListApi` | `/api/service/instance/options` | 当前模块主页面未直接调用。 |
| `frontend/src/core/api/service.js` | `getServiceTableApi` | `api/service/top3` | 代码中存在，但当前后端未见对应路由注册。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceTopologyApi` | `/api/service/topology` | 详情页顶部直接拓扑。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceDsecendantMetricsApi` | `/api/service/descendant/metrics` | 下游依赖时序对比。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceDsecendantRelevanceApi` | `/api/service/descendant/relevance` | 下游依赖相似度排行。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceInstancesApi` | `/api/service/instances` | 实例面板。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceErrorInstancesApi` | `/api/service/error/instance` | 错误实例面板。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceErrorInstancesLogsApi` | `/api/service/errorinstance/logs` | 错误日志时间线。 |
| `frontend/src/core/api/serviceInfo.js` | `getLogMetricsApi` | `/api/service/log/metrics` | 日志指标面板。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceLogLogsApi` | `/api/service/log/logs` | 故障现场日志时间线。 |
| `frontend/src/core/api/serviceInfo.js` | `getTraceMetricsApi` | `/api/service/trace/metrics` | Trace 指标面板。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceTraceLogsApi` | `/api/service/trace/logs` | 故障现场 Trace 时间线。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceAlertEventApi` | `/api/service/alert/sample/events` | 告警样本面板。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceSqlMetrics` | `/api/service/sql/metrics` | SQL 指标面板。 |
| `frontend/src/core/api/serviceInfo.js` | `getServiceEntryEndpoints` | `/api/service/entry/endpoints` | 入口影响面板。 |
| `frontend/src/core/api/serviceInfo.js` | `getK8sEventApi` | `/api/service/k8s/events/count` | K8s 事件面板。 |
| `frontend/src/core/api/serviceInfo.js` | `getPolarisInferApi` | `/api/service/polaris/infer` | Polaris 分析面板。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router | `backend/pkg/router/router_api.go` | 统一把服务观测接口挂在 `/api/service` 下，同时混合注册 `serviceoverview` 与 `service` 两组 handler。 |
| Handler | `backend/pkg/api/serviceoverview/handler.go` | 负责列表页、`moreUrl`、threshold、服务级状态摘要。 |
| Handler | `backend/pkg/api/service/handler.go` | 负责详情页拓扑、依赖、实例、日志、Trace、SQL、K8s、入口影响等能力。 |
| Handler Detail | `backend/pkg/api/serviceoverview/func_getendpointsdata.go` | `/api/service/endpoints`，先做 `CheckGroupPermission` 再转发到 `serviceoverview.GetServicesEndPointData`。 |
| Handler Detail | `backend/pkg/api/serviceoverview/func_getservicemoreurllist.go` | `/api/service/moreUrl`。 |
| Handler Detail | `backend/pkg/api/serviceoverview/func_getservicesalert.go` | `/api/service/servicesAlert`。 |
| Handler Detail | `backend/pkg/api/service/func_getserviceendpointtopology.go` | `/api/service/topology`。 |
| Handler Detail | `backend/pkg/api/service/func_getserviceendpointrelation.go` | `/api/service/relation`。 |
| Handler Detail | `backend/pkg/api/service/func_getdescendantmetrics.go` | `/api/service/descendant/metrics`。 |
| Handler Detail | `backend/pkg/api/service/func_getdescendantrelevance.go` | `/api/service/descendant/relevance`。 |
| Handler Detail | `backend/pkg/api/service/func_getserviceinstance.go` | `/api/service/instances`。 |
| Handler Detail | `backend/pkg/api/service/func_geterrorinstance.go` | `/api/service/error/instance`。 |
| Handler Detail | `backend/pkg/api/service/func_getlogmetrics.go` | `/api/service/log/metrics`。 |
| Handler Detail | `backend/pkg/api/service/func_getloglogs.go` | `/api/service/log/logs`。 |
| Handler Detail | `backend/pkg/api/service/func_gettracemetrics.go` | `/api/service/trace/metrics`。 |
| Handler Detail | `backend/pkg/api/service/func_gettracelogs.go` | `/api/service/trace/logs`。 |
| Handler Detail | `backend/pkg/api/service/func_countk8sevents.go` | `/api/service/k8s/events/count`。 |
| Handler Detail | `backend/pkg/api/service/func_getalerteventssample.go` | `/api/service/alert/sample/events`。 |
| Handler Detail | `backend/pkg/api/service/func_getsqlmetrics.go` | `/api/service/sql/metrics`。 |
| Handler Detail | `backend/pkg/api/service/func_getserviceentryendpoints.go` | `/api/service/entry/endpoints`，内部会组合调用 `serviceoverview`。 |
| Handler Detail | `backend/pkg/api/service/func_getserviceredcharts.go` | `/api/service/redcharts`。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Service List Aggregation | `backend/pkg/services/serviceoverview/service_get_endpoints_data.go`, `service_filldata.go`, `log_filldata.go`, `service_alert.go` | 负责 service 列表 top-N endpoint 聚合、RED、日志摘要、基础设施/网络/K8s 状态补齐。 |
| Full Endpoint Expansion | `backend/pkg/services/serviceoverview/service_moreurl.go` | `moreUrl` 返回指定 service 的全量 endpoint 列表。 |
| Threshold | `backend/pkg/services/serviceoverview/service_get_threshold.go`, `service_set_threshold.go` | 读取和写入 threshold。 |
| Topology | `backend/pkg/services/service/service_get_service_endpoint_topology.go`, `service_get_service_endpoint_relation.go` | 负责直接拓扑与全拓扑关系。 |
| Descendant Analysis | `backend/pkg/services/service/service_get_descendant_metrics.go`, `service_get_descendant_relevance.go` | 负责时序对比、相似度与告警状态补齐。 |
| Entry Impact | `backend/pkg/services/service/service_get_service_entry_endpoints.go` | 负责入口影响路径筛选。 |
| Instance Diagnostics | `backend/pkg/services/service/service_instance_new.go`, `service_filldata.go` | 负责实例列表、RED、日志与状态补齐。 |
| Error / Log / Trace Diagnostics | `backend/pkg/services/service/service_get_error_instance.go`, `service_get_errorinstance_logs.go`, `service_get_log_metrics.go`, `service_get_log_logs.go`, `service_get_trace_metrics.go`, `service_get_trace_logs.go` | 负责错误实例、日志、Trace 故障现场能力。 |
| SQL Diagnostics | `backend/pkg/services/service/service_get_sql_metrics.go` | 负责 SQL 维度 RED 与分页。 |
| Polaris | `backend/pkg/services/service/service_get_polaris_infer.go` | 调用 `polRepo.QueryPolarisInfer`。 |
| Batch Charts | `backend/pkg/services/service/service_get_red_charts.go` | 批量补全服务列表和 endpoint 弹窗的小图。 |
| Permission / Scope | `backend/pkg/services/data/service_check_permission.go`, `backend/pkg/services/common/group_filter.go` | 与 DataGroup 强耦合。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/service_req.go` | 该模块的大部分请求结构定义都在这里。 |
| Response | `backend/pkg/model/response/service_resp.go` | `ServiceEndPointsRes`、`GetServiceEndpointTopologyResponse`、`InstancesRes`、`GetAlertEventsSampleResponse` 等核心返回结构。 |
| Domain Model | `backend/pkg/model/topology.go` | `TopologyNode`、`TopologyRelation`。 |
| Domain Model | `backend/pkg/model/service_instance.go` | `ServiceInstance` 和实例去重逻辑。 |
| Domain Model | `backend/pkg/model/related_instance.go` | 相关实例聚合。 |
| Persistence | `backend/pkg/repository/database/dao_threshold.go` | `threshold` 表定义和读写。 |
| Internal DAO Model | `backend/pkg/services/serviceoverview/model_dao.go` | `serviceoverview` 服务内部聚合结构。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Data Group Filter | `frontend/src/core/components/Filter/DataSourceFilter.tsx` | `/service` 的 cluster/namespace/service/endpoint 过滤树来自 `/api/v2/data/group/filter`，不是本模块专属接口。 |
| Data Group API | `frontend/src/core/api/dataGroup.ts` | `getDatasourceFilterApiV2` 为服务页面提供筛选数据源。 |
| Menu And Route Permission | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/components/AppSidebar.js`, `frontend/src/core/components/AuthRouter.tsx`, `frontend/src/core/api/permission.js` | 菜单来自 `/api/permission/config`；路由检查函数存在，但 `AuthRouter` 内部实际检查被注释。 |
| Charts Batch Context | `frontend/src/core/contexts/ChartsContext.tsx` | 列表页和 endpoint 弹窗共享图表请求。 |
| Time Range | `frontend/src/core/store/reducers/timeRangeReducer.js` | 所有查询都以这里计算出的微秒/整秒微秒时间戳为准。 |
| Data Group State | `frontend/src/core/store/reducers/dataGroupReducer.js`, `frontend/src/core/contexts/DataGroupContext.tsx` | 页面是否能查询成功与 `dataGroupId` 强相关。 |
| Topology Drill-Down State | `frontend/src/core/store/reducers/topologyReducer.js` | 全拓扑弹窗的多级展开路径。 |
| Shared Backend Scope Filter | `backend/pkg/services/common/group_filter.go` | 把数据组范围转换成 Prometheus 过滤条件，并为拓扑节点标记 `OutOfGroup`。 |
| Shared Backend Permission Check | `backend/pkg/services/data/service_check_permission.go` | 负责 `CheckGroupPermission` 与 `CheckServicesPermission`。 |
| Connector Repositories | `backend/pkg/repository/prometheus`, `backend/pkg/repository/clickhouse`, `backend/pkg/repository/polarisanalyzer`, `backend/pkg/repository/database` | 模块核心查询分别落在指标、拓扑/故障现场、Polaris 和权限/threshold DAO 上。 |
| Grafana Dashboard | `frontend/src/oss/views/serviceInfo/component/infoUni/MetricsDashboard.jsx` | 详情页 dashboard 直接拼接 Grafana iframe URL。 |

## Evidence Notes

1. `frontend/src/core/api/service.js` 中仍保留 `getServiceTableApi -> api/service/top3`，但当前 `backend/pkg/router/router_api.go` 未观察到 `top3` 注册。
2. `frontend/src/oss/views/service/component/ThresholdConfigModal.jsx` 仅包含静态表单，当前未见在页面树中被引用，也未接入 `getThreshold`/`setThreshold`。
3. `/api/service/ryglight` 与 `/api/service/monitor/status` 同属 `/api/service` route group，但当前前端消费点主要在 `serviceHealthy`，更适合归入 `dashboards-and-health-monitoring`，不应混入本模块主体说明。
4. `backend/pkg/repository/database/init_menu_item.go` 仅把 `/service` 注册成菜单；`/service/info` 在 `init_router.go` 和 `init_feature_mapping.go` 中存在，但不是独立侧边栏菜单项。
