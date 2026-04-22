# Logs Exploration And Parsing Flows

## Main Flow 1

### Fault Site Log Replay

1. 用户进入 `/logs/fault-site`，页面由 `FaultSiteLogsPage` 装配 `LogsTraceFilterProvider`。
2. `LogsTraceFilter` 负责收集 cluster、namespace、service、instance、`Trace ID` 和时间范围。
3. `FaultSiteLogs` 从 `dataGroupReducer` 读取 `dataGroupId`，并把筛选条件组装成 `/api/log/fault/pagelist` 请求。
4. 后端 `GetFaultLogPageList` 先做 `CheckGroupPermission`。
5. `service_get_fault_log_page_list.go` 把请求转成 `clickhouse.FaultLogQuery`；若前端未传 service 且 `groupId > 0`，会用数据组范围补齐可见 service。
6. 前端按返回结果渲染故障时间点分页标签。
7. 选中某一条后，前端调用 `/api/log/fault/content`。
8. 后端先查应用日志，查不到时 fallback 到 `raw_logs`，最终返回 `sources` 与 `logContents`。

## Main Flow 2

### Full Logs Query And Analysis

1. 用户进入 `/logs/full`，`LogsProvider` 首先调用 `/api/log/table`。
2. 后端返回 `parses` 与 `instances`，前端默认选中第一条解析规则对应的日志表。
3. 当时间范围、查询语句或所选表变化时，`LogsContext.fetchData()` 并行调用 `/api/log/query` 与 `/api/log/chart`。
4. 后端 `QueryLog` 在 `offset > 10000` 时，会借助 histogram 回算偏移区间，然后再执行日志查询。
5. 返回结果被拆成 `logs`、`defaultFields`、`hiddenFields`、`histograms`、`pagination`。
6. 左侧 `IndexList` 在字段展开时调用 `/api/log/index` 拉取字段值分布。
7. 点击某个字段值会把对应条件追加到查询语句，触发新一轮查询。
8. 对于内置解析日志表，点击 `View Context` 会调用 `/api/log/context` 打开上下文弹窗。

## Main Flow 3

### Parse Rule Lifecycle

1. 用户在 `/logs/full` 左侧 `LogRuleList` 打开 `ConfigLogRuleModal`。
2. 若是编辑已有规则，前端先调用 `/api/log/rule/get` 读取规则详情。
3. 弹窗还会调用 `/api/service/list` 拉取 service 列表；选择 service 后再调用 `/api/log/rule/service` 自动补 route rule。
4. 保存时调用 `/api/log/rule/add` 或 `/api/log/rule/update`。
5. 后端会先创建或更新 ClickHouse 日志表结构。
6. 然后读取 Kubernetes 中的 Vector 配置，修改 `route_logs`、`parse_<name>`、`to_<name>` 节点。
7. 最后把解析规则元数据写入或更新到 `logtableinfo`。
8. 删除规则时，链路顺序变为删除 Vector 配置、删除 ClickHouse 表、删除 `logtableinfo` 记录。

## Main Flow 4

### External Table Onboarding

1. 用户在 `/logs/full` 左侧 `DataSourceTree` 打开 `ConfigTableModal`。
2. 前端调用 `/api/log/other` 获取可登记的数据库与表列表。
3. 用户选择目标表后，前端调用 `/api/log/other/table` 获取列定义。
4. 用户选择一个 `timeField` 并提交。
5. 前端调用 `/api/log/other/add`，当前仅提交 `dataBase`、`tableName`、`timeField`。
6. 后端把这些信息写入 `otherlogtable`，并通过默认值补齐未传的 `cluster`、`instance`。
7. 侧边栏外部表树刷新后，用户可以直接切到该表做 `/api/log/query` 查询。
8. 外部表模式下，前端把 `isExternal=true` 传给查询接口，后端把所有字段都视为 `HiddenFields`/`LogFields`。

## Edge Cases

1. 当 `/logs/fault-site` 只传了 `groupId` 而未传 service 时，后端会按数据组权限自动补 service 范围。
2. 当故障现场内容查询不到应用日志时，会回退到 `raw_logs` 做全表检索。
3. 当全量日志查询页码太深导致 `offset > 10000` 时，后端会先调用 histogram 结果回算偏移位置。
4. 当用户切换到外部表模式时，前端会隐藏上下文按钮，因为当前上下文查询只对内置解析日志表开放。
5. 当无日志命中时，`/api/log/query` 返回的是带 `error` 字段的成功响应体，而不是直接 HTTP 错误。

## Failure Paths

1. 请求参数绑定失败时，handler 会直接返回 `ParamBindError`。
2. `GetFaultLogPageList` 权限检查失败时，会返回空列表响应。
3. 解析规则新增/更新/删除跨 ClickHouse、Vector ConfigMap、数据库三段执行；如果中间某一步失败，当前代码未见显式回滚。
4. 外部表登记如果数据库中已存在相同 `(database, tablename, instance)`，service 会把 `table already exists` 写入响应体。
5. `GetServiceRoute` 依赖最近 7 天 Prometheus 活跃实例，如果实例列表为空，自动补 route rule 的结果可能不完整。

## Permission Constraints

1. 当前只在 `POST /api/log/fault/pagelist` 明确看到 `CheckGroupPermission`。
2. `POST /api/log/fault/content` 的 handler 注释明确指出当前未做 service 级数据权限检查。
3. `/logs/full` 当前前后端证据中都未看到 `groupId` 参与查询、图表、字段分析、规则管理或外部表管理。
4. 前端的 `AuthRouter` 存在但当前实际校验逻辑被注释，因此真正的访问控制主要依赖后端中间件和菜单配置。
