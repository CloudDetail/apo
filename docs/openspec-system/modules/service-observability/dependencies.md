# Service Observability Dependencies

## Upstream Dependencies

| Dependency | Evidence | Why It Matters |
|---|---|---|
| Permission Menu Config | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/api/permission.js` | 决定用户是否能在侧边栏看到 `/service`。 |
| Data Group Context | `frontend/src/core/contexts/DataGroupContext.tsx`, `frontend/src/core/store/reducers/dataGroupReducer.js` | 决定服务页面查询使用哪个 `groupId`。 |
| Data Source Filter Tree | `frontend/src/core/components/Filter/DataSourceFilter.tsx`, `frontend/src/core/api/dataGroup.ts` | 服务筛选器的数据来源属于数据范围治理模块。 |
| Shared Time Range | `frontend/src/core/store/reducers/timeRangeReducer.js` | 所有服务查询都依赖统一的微秒时间范围。 |
| Request Wrapper | `frontend/src/core/utils/request.js` | 承担 token、语言头和错误处理。 |
| Backend Permission Service | `backend/pkg/services/data/service_check_permission.go` | 所有 handler 的可见性校验依赖它。 |
| Backend Scope Filter | `backend/pkg/services/common/group_filter.go` | 决定查询最终只落到哪些 cluster/namespace/service。 |

## Downstream Dependencies

| Consumer | Evidence | Dependency Direction |
|---|---|---|
| Logs Module | `frontend/src/oss/views/serviceInfo/component/infoUni/TimeLine.jsx` | 服务详情时间线会打开 `#/logs/fault-site`。 |
| Trace Module | `frontend/src/oss/views/serviceInfo/component/infoUni/TimeLine.jsx` | 服务详情时间线会打开 `#/trace/fault-site`。 |
| Dashboard / Health Module | `frontend/src/oss/views/serviceInfo/component/infoUni/MetricsDashboard.jsx` | 实例面板会为 Grafana dashboard 生成变量。 |
| Automated Diagnosis / Alert Analyze | `frontend/src/pro/api/service.js` 与 `backend/pkg/api/service/*` | 其他模块复用部分服务关系与告警事件接口，但不构成本模块主入口。 |

## Shared Components

| Component | File | Usage |
|---|---|---|
| `BasicTable` | `frontend/src/core/components/Table/basicTable` | 服务列表、endpoint 弹窗、实例、错误实例、依赖排行、SQL 等多个表格。 |
| `StatusInfo` | `frontend/src/core/components/StatusInfo` | 渲染基础设施、网络、K8s、告警等状态。 |
| `TempCell` | `frontend/src/core/components/Table/TempCell` | 渲染 RED/logs 数值与趋势。 |
| `ChartTempCell` | `frontend/src/core/components/Chart/ChartTempCell` | 渲染列表页小型折线图。 |
| `Topology` | `frontend/src/core/components/ReactFlow/Topology` | 直接拓扑和全拓扑弹窗。 |
| `IframeDashboard` | `frontend/src/core/components/Dashboard/IframeDashboard` | 详情页 dashboard。 |
| `LoadingSpinner` | `frontend/src/core/components/Spinner` | 各面板加载态。 |

## Shared Utilities

| Utility | File | Usage |
|---|---|---|
| `getStep` | `frontend/src/core/utils/step` | 计算 RED 和时序请求的步长。 |
| `convertTime` | `frontend/src/core/utils/time` | 渲染时间和格式化 tooltip。 |
| `DelaySourceTimeUnit` | `frontend/src/constants.js` | 把 `self` / `dependency` / `unknown` 映射成文案。 |
| `TimeLineTypeApiMap` | `frontend/src/constants.js` | 在时间线组件里统一选择日志/Trace/错误日志 API。 |
| `notify` | `frontend/src/core/utils/notify` | 下游节点不可追踪时提示用户。 |
| `MarkTopologyNodeInGroup` | `backend/pkg/services/common/group_filter.go` | 为拓扑节点补 `OutOfGroup`。 |
| `GetPQLFilterByGroupID` | `backend/pkg/services/common/group_filter.go` | 把数据组树转换为查询过滤。 |

## External Services

| External Service | Evidence | Role |
|---|---|---|
| `Prometheus` | `backend/pkg/services/serviceoverview/*`, `backend/pkg/services/service/*` | RED、日志计数、实例指标、monitor 状态的主要数据源。 |
| `ClickHouse` | `backend/pkg/services/service/*` | 拓扑、入口祖先、故障现场日志、K8s 事件、部分实例数据源。 |
| `PolarisAnalyzer` | `backend/pkg/services/service/service_get_polaris_infer.go`, `service_get_descendant_relevance.go` | 相似度排序和 Polaris 诊断。 |
| `Grafana` | `frontend/src/oss/views/serviceInfo/component/infoUni/MetricsDashboard.jsx` | 详情页 dashboard iframe。 |
| `Kubernetes` | 通过实例与 K8s 事件聚合链路间接依赖 | 用于补充 namespace、pod、节点以及 K8s 事件语义。 |

## Dependency Notes

1. 本模块前端最关键的上游依赖不是“服务自身下拉字典”，而是共享的 `DataGroup` 过滤树。
2. 本模块后端最关键的上游依赖不是某一张业务表，而是 `Prometheus`、`ClickHouse` 和 `database` 权限范围查询的组合。
3. 详情页时间线与二级跳转说明该模块是 logs/trace 模块的重要上游入口，而不是完全自闭环模块。
