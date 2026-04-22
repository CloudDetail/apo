# Service Observability Frontend

## Pages

| Page | File | Description |
|---|---|---|
| `/service` | `frontend/src/oss/views/service/index.js` | 服务总览页，展示 service 列表、top endpoint 摘要、日志/基础设施状态和批量趋势图。 |
| `/service/info` | `frontend/src/oss/views/serviceInfo/index.js` | 服务详情页，围绕单个 `service` + `endpoint` 展开拓扑、依赖和诊断面板。 |

## Routes

| Route | Route File | Notes |
|---|---|---|
| `/service` | `frontend/src/oss/routes.js` | `showDataGroup: true`，表示页面直接展示数据组选择能力。 |
| `/service/info` | `frontend/src/oss/routes.js` | `showDataGroup: 'view'`，详情页保留数据组视图上下文。 |

## Menus

| Menu Aspect | Evidence | Notes |
|---|---|---|
| Sidebar Item | `backend/pkg/repository/database/init_menu_item.go` | 菜单 key 为 `service`，路由为 `/service`。 |
| Menu To Feature Mapping | `backend/pkg/repository/database/init_feature_mapping.go` | `服务概览` 与菜单 `service` 绑定。 |
| Hidden Detail Route | `backend/pkg/repository/database/init_router.go` | `/service/info` 在 router 表中存在，但不是一级菜单。 |
| Frontend Menu Render | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/components/AppSidebar.js` | 侧边栏最终由 `/api/permission/config` 返回的 `menuItems` 渲染。 |

## Key Components

| Component | File | Role |
|---|---|---|
| `ServiceTable` | `frontend/src/oss/views/service/ServiceTable.jsx` | 服务列表主表、排序、筛选、跳转和小图补全。 |
| `EndpointTableModal` | `frontend/src/oss/views/service/component/EndpointTableModal.jsx` | 查看某个 service 的全量 endpoint。 |
| `ServiceInfo` | `frontend/src/oss/views/serviceInfo/index.js` | 详情页顶层容器，负责解析 query 参数和拉取直接拓扑。 |
| `DependentTabs` | `frontend/src/oss/views/serviceInfo/component/dependent/index.jsx` | 下游依赖分析区域。 |
| `TopologyModal` | `frontend/src/oss/views/serviceInfo/component/dependent/TopologyModal.jsx` | 基于 `relation` 接口展示更完整的下游拓扑。 |
| `InfoUni` | `frontend/src/oss/views/serviceInfo/component/infoUni/index.jsx` | 详情页折叠面板总入口。 |
| `EntryImpact` | `frontend/src/oss/views/serviceInfo/component/infoUni/EntryImpact.jsx` | 展示最短入口路径及其异常摘要。 |
| `InstanceInfo` | `frontend/src/oss/views/serviceInfo/component/infoUni/InstanceInfo.jsx` | 展示实例级 RED 和状态，并为 dashboard 构造变量。 |
| `AlertInfoTabs` | `frontend/src/oss/views/serviceInfo/component/infoUni/AlertInfo/AlertInfoTabs.jsx` | 展示告警样本。 |
| `ErrorInstanceInfo` | `frontend/src/oss/views/serviceInfo/component/infoUni/ErrorInstance/ErrorInstanceInfo.jsx` | 展示错误实例与错误传播。 |
| `LogsInfo` | `frontend/src/oss/views/serviceInfo/component/infoUni/LogsInfo.jsx` | 展示日志指标与故障现场日志时间线。 |
| `TraceInfo` | `frontend/src/oss/views/serviceInfo/component/infoUni/TraceInfo.jsx` | 展示 Trace 指标与故障现场 Trace 时间线。 |
| `K8sInfo` | `frontend/src/oss/views/serviceInfo/component/infoUni/K8sInfo.jsx` | 展示 K8s 事件统计。 |
| `SqlMetrics` | `frontend/src/oss/views/serviceInfo/component/infoUni/SqlMetrics.jsx` | 展示 SQL 操作维度指标。 |
| `PolarisMetricsInfo` | `frontend/src/oss/views/serviceInfo/component/infoUni/PolarisMetricsInfo/PolarisMetricsInfo.jsx` | 展示 Polaris 推理图和文本结论。 |
| `MetricsDashboard` | `frontend/src/oss/views/serviceInfo/component/infoUni/MetricsDashboard.jsx` | 通过 Grafana iframe 展示实例相关 dashboard。 |
| `Timeline` | `frontend/src/oss/views/serviceInfo/component/infoUni/TimeLine.jsx` | 统一封装错误日志、故障现场日志、故障现场 Trace 时间线与跳转。 |

## State Management

| State Area | Evidence | Notes |
|---|---|---|
| Time Range | `frontend/src/core/store/reducers/timeRangeReducer.js` | `/service` 用 `selectSecondsTimeRange`，部分详情面板用 `selectProcessedTimeRange` 或 `selectSecondsTimeRange`。 |
| Data Group | `frontend/src/core/store/reducers/dataGroupReducer.js` | 几乎所有服务观测请求都显式依赖 `dataGroupId`。 |
| Detail Local State | `frontend/src/oss/contexts/ServiceInfoContext.tsx` | 记录面板状态、当前展开面板、Grafana 变量。 |
| Query Propagation | `frontend/src/core/contexts/PropsContext.js` | 详情页把 `serviceName`、`endpoint`、`clusterIds` 透传给子组件。 |
| Chart Cache | `frontend/src/core/contexts/ChartsContext.tsx` | 统一批量请求列表小图数据。 |
| Topology Drill-Down | `frontend/src/core/store/reducers/topologyReducer.js` | `TopologyModal` 使用 redux 记忆 drill-down 路径与显示节点。 |
| Menu / Route Context | `frontend/src/core/contexts/UserContext.js` | 决定 `/service` 是否出现在菜单中。 |

## API Usage

| Frontend Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 加载 `/service` 主表 | `getServicesEndpointsApi` | `/api/service/endpoints` | 返回按 service 聚合后的 top endpoint。 |
| 加载 `/service` 行级状态摘要 | `getServicesAlertApi` | `/api/service/servicesAlert` | 返回日志、基础设施、网络、K8s 状态摘要。 |
| 加载 `/service` 和 endpoint 弹窗小图 | `getServiceChartsApi` | `/api/service/redcharts` | 由 `ChartsContext` 触发批量查询。 |
| 展开某个 service 的全部 endpoint | `getEndpointTableApi` | `/api/service/moreUrl` | endpoint 数量大于 3 时弹出。 |
| 加载详情页直接拓扑 | `getServiceTopologyApi` | `/api/service/topology` | `ServiceInfo` 页面初始化调用。 |
| 加载下游延时对比 | `getServiceDsecendantMetricsApi` | `/api/service/descendant/metrics` | `TimelapseLineChart` 使用。 |
| 加载下游相似度排行 | `getServiceDsecendantRelevanceApi` | `/api/service/descendant/relevance` | `DependentTable` 使用。 |
| 加载全拓扑 | `getServiceRelationApi` | `/api/service/relation` | `TopologyModal` 使用。 |
| 加载入口影响 | `getServiceEntryEndpoints` | `/api/service/entry/endpoints` | `EntryImpact` 使用。 |
| 加载实例列表 | `getServiceInstancesApi` | `/api/service/instances` | `InstanceInfo` 使用。 |
| 加载错误实例 | `getServiceErrorInstancesApi` | `/api/service/error/instance` | `ErrorInstanceInfo` 使用。 |
| 加载错误实例日志时间线 | `getServiceErrorInstancesLogsApi` | `/api/service/errorinstance/logs` | `Timeline` 在 `errorLogs` 模式下使用。 |
| 加载日志指标 | `getLogMetricsApi` | `/api/service/log/metrics` | `LogsInfo` 使用。 |
| 加载故障现场日志时间线 | `getServiceLogLogsApi` | `/api/service/log/logs` | `Timeline` 在 `logsInfo` 模式下使用。 |
| 加载 Trace 指标 | `getTraceMetricsApi` | `/api/service/trace/metrics` | `TraceInfo` 使用。 |
| 加载故障现场 Trace 时间线 | `getServiceTraceLogsApi` | `/api/service/trace/logs` | `Timeline` 在 `traceLogs` 模式下使用。 |
| 加载 K8s 事件 | `getK8sEventApi` | `/api/service/k8s/events/count` | `K8sInfo` 使用。 |
| 加载告警样本 | `getServiceAlertEventApi` | `/api/service/alert/sample/events` | `AlertInfoTabs` 使用。 |
| 加载 SQL 指标 | `getServiceSqlMetrics` | `/api/service/sql/metrics` | `SqlMetrics` 使用。 |
| 加载 Polaris 结论 | `getPolarisInferApi` | `/api/service/polaris/infer` | `PolarisMetricsInfo` 使用。 |
| 加载数据源筛选树 | `getDatasourceFilterApiV2` | `/api/v2/data/group/filter` | 来自共享数据组模块，不属于本模块后端。 |

## User Interaction Flow

### Flow A
用户进入 `/service` 后，`DataSourceFilter` 先根据 `groupId`、时间范围和 `category="apm"` 获取 cluster/namespace/service/endpoint 过滤树。随后 `ServiceTable` 调用 `/api/service/endpoints` 拉取主表，再通过 `/api/service/servicesAlert` 和 `/api/service/redcharts` 补齐状态摘要与小图。

### Flow B
用户点击 service 行中的某个 endpoint，前端跳转到 `/service/info?service-name=...&endpoint=...&groupId=...&clusterIds=...`。详情页先调用 `/api/service/topology` 渲染直接拓扑，再并行拉取依赖排行、入口影响、实例、告警样本、错误实例、日志、Trace、K8s、SQL 与 Polaris 面板。

### Flow C
用户在详情页继续钻取时，有三种典型动作：

- 在 `DependentTable` 中点击被监控的下游节点，再次跳转到新的 `/service/info`。
- 在 `Timeline` 中点击时间点，打开 `#/logs/fault-site` 或 `#/trace/fault-site`。
- 在 `TopologyModal` 中通过 redux 保存路径，逐层查看更深的下游结构。

## Permissions

| Aspect | Evidence | Notes |
|---|---|---|
| Route Wrapper | `frontend/src/oss/routes.js` + `frontend/src/core/components/AuthRouter.tsx` | 两个页面都被 `AuthRouter` 包装。 |
| Actual Frontend Route Check | `frontend/src/core/components/AuthRouter.tsx` | `checkRoutePermission` 逻辑目前被整段注释，前端侧实际并未主动阻断。 |
| Menu Visibility | `frontend/src/core/contexts/UserContext.js` + `frontend/src/core/components/AppSidebar.js` | 是否显示 `/service` 依赖 `/api/permission/config` 返回的菜单数据。 |
| Data Scope | 各页面组件普遍依赖 `dataGroupId` | 即使页面可见，数据查询仍受 `groupId` 影响。 |
| Detail Route Permission | `backend/pkg/repository/database/init_feature_mapping.go` | `/service/info` 作为 feature route 存在，但不是侧边栏菜单项。 |

## Notes

1. `ThresholdConfigModal` 当前未挂到 `ServiceTable` 或其他页面树上，不能视为已交付的用户入口。
2. `getServiceTableApi` 仍指向 `api/service/top3`，当前页面并未使用，且后端未见对应路由。
3. `MetricsDashboard` 不是通过本模块后端 API 拉数据，而是直接构造 Grafana iframe URL。
4. `DependentTable` 对 `isTraced === false` 的节点不会跳转，只会弹出提示，说明“拓扑可见”与“可深入追踪”在前端并不等价。
