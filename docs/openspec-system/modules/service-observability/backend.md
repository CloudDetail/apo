# Service Observability Backend

## API Summary

### Overview APIs

| API Path | Handler | Service | Purpose |
|---|---|---|---|
| `/api/service/endpoints` | `serviceoverview.GetEndPointsData` | `serviceoverview.GetServicesEndPointData` | 返回服务列表及 top endpoint 摘要。 |
| `/api/service/servicesAlert` | `serviceoverview.GetServicesAlert` | `serviceoverview.GetServicesAlert` | 返回日志、基础设施、网络、K8s 状态摘要。 |
| `/api/service/moreUrl` | `serviceoverview.GetServiceMoreUrlList` | `serviceoverview.GetServiceMoreUrl` | 返回指定 service 的全量 endpoint。 |
| `/api/service/redcharts` | `service.GetServiceREDCharts` | `service.GetServiceREDCharts` | 批量补图。 |
| `/api/service/getThreshold` | `serviceoverview.GetThreshold` | `serviceoverview.GetThreshold` | 读取 threshold。当前前端主页面未接线。 |
| `/api/service/setThreshold` | `serviceoverview.SetThreshold` | `serviceoverview.SetThreshold` | 写 threshold。当前前端主页面未接线。 |

### Detail APIs

| API Path | Handler | Service | Purpose |
|---|---|---|---|
| `/api/service/topology` | `service.GetServiceEndpointTopology` | `service.GetServiceEndpointTopology` | 查询直接上下游拓扑。 |
| `/api/service/relation` | `service.GetServiceEndpointRelation` | `service.GetServiceEndpointRelation` | 查询完整下游关系和层级。 |
| `/api/service/descendant/metrics` | `service.GetDescendantMetrics` | `service.GetDescendantMetrics` | 下游依赖延时时序。 |
| `/api/service/descendant/relevance` | `service.GetDescendantRelevance` | `service.GetDescendantRelevance` | 下游依赖相似度与状态。 |
| `/api/service/entry/endpoints` | `service.GetServiceEntryEndpoints` | `service.GetServiceEntryEndpointsInGroup` + `serviceoverview` 组合 | 入口影响分析。 |
| `/api/service/instances` | `service.GetServiceInstance` | `service.GetInstancesNew` | 实例级 RED 和状态。 |
| `/api/service/error/instance` | `service.GetErrorInstance` | `service.GetErrorInstance` | 错误实例与传播链。 |
| `/api/service/errorinstance/logs` | `service.GetErrorInstanceLogs` | `service.GetErrorInstanceLogs` | 错误实例日志时间线。 |
| `/api/service/log/metrics` | `service.GetLogMetrics` | `service.GetLogMetrics` | 日志面板指标。 |
| `/api/service/log/logs` | `service.GetLogLogs` | `service.GetLogLogs` | 故障现场日志时间线。 |
| `/api/service/trace/metrics` | `service.GetTraceMetrics` | `service.GetTraceMetrics` | Trace 面板指标。 |
| `/api/service/trace/logs` | `service.GetTraceLogs` | `service.GetTraceLogs` | 故障现场 Trace 时间线。 |
| `/api/service/k8s/events/count` | `service.CountK8sEvents` | `service.CountK8sEvents` | K8s 事件统计。 |
| `/api/service/alert/sample/events` | `service.GetAlertEventsSample` | `service.GetAlertEventsSample` | 告警样本分组结果。 |
| `/api/service/sql/metrics` | `service.GetSQLMetrics` | `service.GetSQLMetrics` | SQL 操作维度指标。 |
| `/api/service/polaris/infer` | `service.GetPolarisInfer` | `service.GetPolarisInfer` | Polaris 分析结论。 |
| `/api/service/list` | `service.GetServiceList` | `service.GetServiceList` | 下拉服务列表。 |
| `/api/service/instance/list` | `service.GetServiceInstanceList` | `service.GetServiceInstanceList` | 旧实例列表接口。 |
| `/api/service/instance/options` | `service.GetServiceInstanceOptions` | `service.GetServiceInstanceOptions` | 实例 option 接口。 |
| `/api/service/endpoint/list` | `service.GetServiceEndPointList` | `service.GetServiceEndPointList` | endpoint option 接口。 |
| `/api/service/namespace/list` | `service.GetNamespaceList` | `service.GetServiceNamespaceList` | namespace option 接口。 |

## Controllers Or Routers

| File | Role |
|---|---|
| `backend/pkg/router/router_api.go` | 注册 `/api/service` route group，并在同一 group 内同时初始化 `serviceoverview.New(...)` 与 `service.New(...)`。 |
| `backend/pkg/api/serviceoverview/handler.go` | 列表侧 handler 工厂，依赖 `clickhouse`、`prometheus`、`database`、`kubernetes`、`data.Service`。 |
| `backend/pkg/api/service/handler.go` | 详情侧 handler 工厂，依赖 `clickhouse`、`prometheus`、`polarisanalyzer`、`database`、`kubernetes`、`serviceoverview.Service`、`data.Service`。 |

当前 handler 的共同模式比较稳定：

1. `ShouldBind` 或 `ShouldBindQuery` / `ShouldBindJSON` 做请求绑定。
2. 先做 `CheckGroupPermission` 或 `CheckServicesPermission`。
3. 调用 service 层。
4. 使用 `c.Payload(...)` 返回成功结果，或 `c.AbortWithError(...)` / `c.AbortWithPermissionError(...)` 返回失败。

## Services

### serviceoverview

`backend/pkg/services/serviceoverview` 负责 `/service` 主表和 `moreUrl` 弹窗的聚合。核心特点如下：

- `GetServicesEndPointData` 先通过 `GetPQLFilterByGroupID` 生成数据组过滤条件，再调用 `prometheus.FetchEndpointsData` 批量取 RED、delay source、namespace。
- 返回结果会按 service 聚合，并只保留 top 3 endpoint。
- `GetServicesAlert` 会进一步补齐日志、基础设施、网络、K8s 状态和 `AlertReason`。
- `GetServiceMoreUrl` 面向单个 service 返回全量 endpoint 列表。
- threshold 通过 `dbRepo.GetOrCreateThreshold` 读取，并参与排序或告警阈值判断。

### service

`backend/pkg/services/service` 负责 `/service/info` 的深度诊断能力。核心特点如下：

- `GetServiceEndpointTopology` 通过 `clickhouse.ListParentNodes` 和 `ListChildNodes` 生成直接拓扑，再用 `MarkTopologyNodeInGroup` 标记 `OutOfGroup`。
- `GetServiceEndpointRelation` 返回更完整的下游调用关系，并在需要时计算 `TopologyLevels`。
- `GetDescendantRelevance` 先从 ClickHouse 找下游节点，再调用 `polRepo.SortDescendantByRelevance` 做相似度排序，并补齐 RED/K8s/日志状态。
- `GetInstancesNew` 会优先尝试 `GetDataplaneServiceInstances`，失败时回退到 Prometheus 实例查询，再补齐 RED、日志和告警状态。
- `GetServiceEntryEndpoints` 先找 group 内最短祖先入口，再借助 `serviceoverview` 的 endpoint 和告警聚合逻辑生成入口影响结果。
- `GetServiceREDCharts` 支持 service 列表和 endpoint 弹窗的批量小图查询。

## Models And Schemas

| Type | File | Notes |
|---|---|---|
| Request Schema | `backend/pkg/model/request/service_req.go` | 定义了拓扑、依赖、实例、日志、Trace、SQL、threshold、列表等请求。 |
| Response Schema | `backend/pkg/model/response/service_resp.go` | 定义了 `ServiceEndPointsRes`、`ServiceDetail`、`InstancesRes`、`GetServiceEndpointTopologyResponse`、`GetAlertEventsSampleResponse` 等主要返回体。 |
| Domain Model | `backend/pkg/model/topology.go` | `TopologyNode`、`TopologyRelation`、`ServiceToplogyNode`。 |
| Domain Model | `backend/pkg/model/service_instance.go` | 实例去重与实例 ID 规则。 |
| DAO Model | `backend/pkg/repository/database/dao_threshold.go` | `threshold` 表。 |
| Internal Aggregation Model | `backend/pkg/services/serviceoverview/model_dao.go` | 服务总览聚合过程使用的内部结构。 |

## Database Mapping

| Area | Evidence | Notes |
|---|---|---|
| Threshold Persistence | `backend/pkg/repository/database/dao_threshold.go` | 唯一直接属于本模块的业务表是 `threshold`。 |
| Data Scope / Permission | `backend/pkg/repository/database/dao_auth_data_group.go` 以及 `GetDataGroupIDsByUserId`、`GetScopeIDsSelectedByGroupID` 等调用 | 本模块多处依赖数据组和权限表，但这些不是本模块自有业务表。 |
| Metrics Storage | `prometheus.Repo` | 绝大多数 RED、日志计数、实例指标来自时序库，不通过传统业务表存储。 |
| Topology / Fault Scene | `clickhouse.Repo` | 拓扑、入口祖先、故障现场日志、K8s 事件等来自 ClickHouse。 |

当前代码显示，本模块的数据并不是“单表 CRUD”型模块，而是“多数据源聚合查询”型模块。

## Validation And Rules

| Rule | Evidence | Notes |
|---|---|---|
| 时间范围校验 | `service_req.go` 中大量 `binding:"required,gtfield=StartTime"` | 多数接口要求合法的起止时间。 |
| `service` / `endpoint` 必填 | 拓扑、依赖、错误实例、日志、Trace 等请求结构 | 详情页所有核心查询都围绕 `service + endpoint`。 |
| 数据组权限 | `CheckGroupPermission` | 列表和多数详情查询依赖 `groupId`。 |
| 服务可见性权限 | `CheckServicesPermission` | 错误实例、日志时间线、Trace 时间线等更偏单 service 的接口按 service 校验。 |
| 拓扑组内标记 | `MarkTopologyNodeInGroup` | 对不在当前数据组范围内的 service 节点打 `OutOfGroup`。 |
| threshold 回退 | `dbRepo.GetOrCreateThreshold` | 若没有指定阈值，回退到 `global`，再不行就创建默认阈值。 |
| 相似度排序回退 | `GetDescendantRelevance` | 若 Polaris 排序失败，回退到未排序列表并标记 `DistanceType = net_failed`。 |

## Error Handling

| Pattern | Evidence | Notes |
|---|---|---|
| 统一参数错误 | `c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)` | 几乎所有 handler 都沿用。 |
| 权限失败返回空结构 | `c.AbortWithPermissionError(...)` | 返回空数组或空对象而不是直接崩溃。 |
| Service 层部分降级 | 例如 `GetDescendantRelevance`、`GetInstancesNew` | 某些聚合失败后会局部记录错误并继续返回部分结果。 |
| 前端友好空集合 | 多个 handler 显式构造 `[]` 或空 map | 避免前端收到 `null`。 |

## Notes

1. `/api/service` 不是纯粹的“服务详情 API”命名空间，而是同时承载“服务列表聚合 API”和“服务详情 API”的大 namespace。
2. `GetServiceEntryEndpoints` 不是简单单一查询，而是先从 `service` 服务侧找入口，再调用 `serviceoverview` 聚合 endpoint 和告警摘要。
3. 当前实现大量依赖共享 `prometheus.Repo`、`clickhouse.Repo` 与 `database.Repo`，说明本模块本质是聚合读模型，而不是独立持久化写模型。
4. `getThreshold` / `setThreshold` 已有后端，但从当前前端证据看还没有稳定接线到用户入口。
