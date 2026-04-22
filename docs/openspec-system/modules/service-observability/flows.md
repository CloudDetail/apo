# Service Observability Flows

## Main Flow 1

### Service Overview Flow

1. 用户进入 `/service`。
2. `DataSourceFilter` 调用 `/api/v2/data/group/filter`，按当前 `groupId`、时间范围与 `category="apm"` 获取 cluster、namespace、service、endpoint 过滤树。
3. `ServiceTable` 调用 `/api/service/endpoints`，获取 service 维度的主表数据。
4. 主表按 service 展开最多 3 个 endpoint，并为每个 endpoint 展示 RED 指标与 `DelaySource`。
5. 行级的日志/基础设施/网络/K8s 状态通过 `/api/service/servicesAlert` 补齐。
6. 当前页分页中的 service/endpoint 组合通过 `/api/service/redcharts` 批量补齐折线图。
7. 用户可以直接点击 endpoint 进入 `/service/info`，或点击 `showMore` 打开 `EndpointTableModal` 查看该 service 的全部 endpoint。

这一条主链体现的是“从 service 目录进入单个 endpoint 详情”的入口能力。

## Main Flow 2

### Service Detail Diagnostic Flow

1. 用户进入 `/service/info?service-name=...&endpoint=...&groupId=...&clusterIds=...`。
2. 页面顶层先调用 `/api/service/topology` 获取当前 endpoint 的直接上下游拓扑。
3. `DependentTabs` 并行调用：
   - `/api/service/descendant/metrics` 生成下游依赖延时时序对比。
   - `/api/service/descendant/relevance` 生成下游依赖相似度排行。
4. `TopologyModal` 在用户需要更大范围拓扑时调用 `/api/service/relation`，并配合 `topologyReducer` 做 drill-down。
5. `InfoUni` 各折叠面板继续并行加载：
   - `/api/service/entry/endpoints`
   - `/api/service/alert/sample/events`
   - `/api/service/instances`
   - `/api/service/k8s/events/count`
   - `/api/service/error/instance`
   - `/api/service/log/metrics`
   - `/api/service/trace/metrics`
   - `/api/service/sql/metrics`
   - `/api/service/polaris/infer`
6. 日志、错误实例和 Trace 时间线继续调用对应故障现场 API，并允许用户跳转到 `#/logs/fault-site` 或 `#/trace/fault-site` 深入排查。
7. `InstanceInfo` 还会根据实例结果构造 namespace、pod、service 变量，供 `MetricsDashboard` 拼装 Grafana iframe URL。

这一条主链体现的是“从一个 endpoint 切到多视角诊断”的详情能力。

## Edge Cases

| Scenario | Observed Handling |
|---|---|
| 服务列表无数据 | `/service` 页面展示文档链接，分别引导排查“未监控”和“已监控但无数据”。 |
| 下游节点未被追踪 | `DependentTable` 点击此类节点时不跳转，只弹出提示。 |
| 拓扑节点超出数据组范围 | 后端通过 `OutOfGroup` 标记节点，前端保留可视化但不应等同于完全可用节点。 |
| 部分指标为空 | 多个面板用 `N/A`、空数组、空 map 或补零图处理。 |
| Polaris 无结果 | `PolarisMetricsInfo` 只在 `image` 或 `inferCause` 存在时展示内容。 |
| SQL 面板分页切换 | `SqlMetrics` 在 `pageIndex` 和 `pageSize` 变化后重新请求。 |

## Failure Paths

| Failure Path | Observed Behavior |
|---|---|
| 参数绑定失败 | handler 直接返回 `ParamBindError`。 |
| 数据组权限失败 | handler 返回空结构并走 `AbortWithPermissionError`。 |
| 单个详情面板请求失败 | 前端多数面板把本面板数据置空，并把状态设为 `unknown` 或维持默认值。 |
| Polaris 相似度排序失败 | `GetDescendantRelevance` 回退到未排序结果，并标记 `DistanceType = net_failed`。 |
| 图表补图失败 | `ChartsContext` 不会中断整个页面，只是局部图表缺失。 |

## Permission Constraints

| Constraint | Description |
|---|---|
| 菜单可见性 | `/service` 是否显示由 `/api/permission/config` 返回的菜单控制。 |
| 路由前端 guard | 两个页面都包在 `AuthRouter` 下，但当前 guard 逻辑被注释，实际前端阻断能力 `To Be Confirmed`。 |
| 数据组范围 | 主列表、详情页大部分查询都依赖 `groupId`，并由后端转成 Prometheus/拓扑范围过滤。 |
| 服务级校验 | 错误实例、日志时间线、Trace 时间线等接口额外按 service 做 `CheckServicesPermission`。 |
| 二级跳转 | 从详情页再跳到 logs/trace 页面时，会继续携带 `groupId` 和 `clusterIds`，形成跨模块上下文。 |
