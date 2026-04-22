# Dataplane And Topology Governance Flows

## Main Flow 1

### DataPlane Configuration Flow

1. 用户进入 `/dataplane`，默认打开 `dataplane` tab。
2. `DataplaneList` 调用 `/api/dataplane/type/list` 渲染可配置的数据平面类型。
3. 用户点击某个数据平面类型卡片，页面打开 `ConfigDrawer`。
4. `DataplaneConfigForm` 根据 `paramSpec.children` 渲染参数输入框。
5. 用户点击 “Test” 时，前端调用 `/api/dataplane/checkDataSource`，传 `dataSource`、`attributes`、`enabledCapabilities`。
6. 后端验证配置，返回成功或错误消息。
7. 用户点击 “Connect” 时，前端调用 `/api/dataplane/create` 或 `/api/dataplane/update`。
8. 后端校验 `typ` 是否存在、参数是否满足模板约束；若 `capability` 为空则自动补全全部能力。
9. 成功后前端刷新 `/api/dataplane/list`，右侧 `ConfiguredSources` 展示最新配置。

## Main Flow 2

### Service Matcher Rule Flow

1. 用户切到 `/dataplane?tab=serviceMatcher`。
2. `ServiceMatcherPage` 使用当前时间范围调用 `/api/dataplane/servicesWithInstances`。
3. 后端只返回来源于当前已配置数据源或规则归因的服务，并把实例补上 `source`/`rule` 来源。
4. 用户展开某个 service 行，前端调用 `/api/dataplane/servicename/listRule?serviceName=<service>`。
5. 用户打开 `ConfigModal` 新建或编辑规则。
6. `ConditionsFormList` 调用 `/api/dataplane/appinfo/tags` 和 `/api/dataplane/appinfo/tags/values` 辅助填写条件。
7. 用户点击 “Test” 时，前端先校验表单，再调用 `/api/dataplane/servicename/checkRule`。
8. 后端遍历 `GetToResolveApps` 返回的待归因实例，按条件即时匹配，返回命中的实例列表。
9. 用户点击保存，前端调用 `/api/dataplane/servicename/upsertRule`。
10. 后端写入 `service_name_rule` 与 `service_name_rule_condition`，前端刷新规则列表和 service 匹配结果。

## Main Flow 3

### Agent Readiness And Status Flow

1. Agent 页面或相关组件调用 `/api/dataplane/data/check`。
2. 后端从 Prometheus 判断最近 7 天是否存在最小 agent cluster，并从 DB 判断是否已配置数据平面。
3. 返回结果中的 `agentReady`、`dataPlaneExist`、`clusters`、`dataSource` 被前端用于展示前置状态。
4. Agent header 中的 `DataStatus` 再调用 `/api/dataplane/datasources` 读取数据源数量与服务数量。
5. 前端同时调用 `/api/dataplane/capabilities` 构造能力状态和配置跳转入口。
6. 后端从 DB 读取已配置 `data_plane` 和 `data_plane_type`，按 telemetry type 聚合 `enabled` 状态与来源列表。
7. 用户点击某个 badge 或 capability 设置按钮时，会打开 `/#/dataplane?tab=...`。

## Main Flow 4

### Topology Merge Flow

1. 某前端或内部调用请求 `/api/dataplane/topology`。
2. 后端 `GetServiceTopology` 先从 ClickHouse 查询实时拓扑边。
3. 后端再从 DB 读取 `custom_service_topology` 静态边。
4. 对于未过期的静态边，后端补建 `ServiceToplogyNode` 并加入父子关系。
5. 最终返回包含实时节点与静态节点的统一拓扑结果。

## Edge Cases

| Case | Implemented Behavior |
|---|---|
| `SHOW_APO_DATA_PLANE_TYPE` 未开启 | 类型列表和已配置列表都过滤 `apo`。 |
| 数据平面未传 capability | 自动使用 `DataPlaneType.CapabilitySpec.Obj`。 |
| `prometheus` 类型接入 | 后端强制把 `promStorage` 改写为 `"prom"`。 |
| 无任何数据平面 | `/api/dataplane/data/check` 返回 `dataPlaneExist=false`。 |
| 规则条件为空 | `CheckServiceNameRule` 中 `checkMatch` 会直接返回 false。 |
| 静态拓扑边已过期 | `GetServiceTopology` 不会合并该边。 |
| 删除规则 | `DeleteServiceNameRule` 会先删条件，再删主规则。 |
| 删除带 cluster 关系的数据平面 | 后端事务内会同时删 `cluster2data_plane`。 |

## Failure Paths

| Path | Failure Behavior |
|---|---|
| `/api/dataplane/checkDataSource` 测试失败 | 前端 `DataplaneConfigForm` 展示错误 `Alert`。 |
| 创建数据平面时 type 不存在 | 后端返回 `DataPlaneTypeNotExistError`。 |
| 参数不合法 | 后端返回 `DataPlaneParamInvalidError`。 |
| 删除不存在的数据平面 | 后端返回 `DataPlaneNotExistError`。 |
| 查询服务或实例失败 | 对应接口返回 error；前端当前主要是 `console.log`，缺少明确错误 UI。 |
| 实时拓扑查询失败 | `/api/dataplane/topology` 返回 `Msg: query realtime topology failed: ...`。 |
| DB 查询静态拓扑失败 | `/api/dataplane/topology` 返回 `Msg: query custom topology failed: ...`。 |
| `/api/dataplane/capabilities` 查询失败 | 前端相关能力状态展示可能为空，或退回到缺少 capability 详情的状态。 |

## Permission Constraints

| Layer | Constraint |
|---|---|
| Frontend Route | `/dataplane` 经 `AuthRouter` 包装，但当前不会真正拦截。 |
| Backend API | `/api/dataplane` 整组使用 `LicMiddleware()`。 |
| Historical Router Permission | `CheckRouterPermission` 对 `/agent-dataplane`、`agent`、`probe-installer` 有硬编码放行。 |

## To Be Confirmed

- `/api/dataplane/capabilities` 当前表达的是“已配置 capability 摘要”还是“运行时能力可用性”，需要确认。
- `customtopology` 虽然在后端已完整暴露，但前端主页面当前没有明显入口，是否属于隐式能力需要确认。
- 当前权限模型仍依赖 `/agent-dataplane`，是否存在前端迁移后未同步权限元数据的问题，需要确认。
