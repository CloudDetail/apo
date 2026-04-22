# Logs Exploration And Parsing Source Map

## Scope
本文档只做证据提取，不提前扩写需求。目标是把 `logs-exploration-and-parsing` 模块在当前仓库中的真实代码落点固定下来，为后续详细文档提供可追溯依据。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/oss/views/logs` | `/logs/fault-site` 与 `/logs/full` 页面主目录。 |
| Frontend | `frontend/src/oss/views/logs/component/FullLogs` | 全量日志检索主界面、字段分析、结果列表、侧边栏配置能力。 |
| Frontend | `frontend/src/oss/contexts/LogsTraceFilterContext.tsx` | 故障现场日志筛选状态容器。 |
| Frontend | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | 故障现场日志筛选组件，复用在 trace 页面。 |
| Frontend | `frontend/src/core/contexts/LogsContext.js` | `/logs/full` 的查询、表信息、字段索引、分页状态容器。 |
| Frontend | `frontend/src/core/store/reducers/logsReducer.js` | `/logs/full` Redux 状态定义。 |
| Frontend | `frontend/src/core/api/logs.js` | 本模块所有前端 API wrapper。 |
| Frontend | `frontend/src/core/api/service.js` | `ConfigLogRuleModal` 额外依赖 `getServiceListApi`。 |
| Frontend | `frontend/src/core/store/reducers/timeRangeReducer.js` | `/logs/full` 共享系统时间范围。 |
| Frontend | `frontend/src/core/components/AuthRouter.tsx` | 两个页面都通过路由守卫包装，但当前实际检查逻辑被注释。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/log/*` 路由注册入口。 |
| Backend | `backend/pkg/api/log` | 日志模块 handler 层。 |
| Backend | `backend/pkg/services/log` | 故障现场日志、全量日志、上下文、字段索引、解析规则、外部表管理。 |
| Backend | `backend/pkg/services/log/vector` | 解析规则与 Vector ConfigMap 的配置变更逻辑。 |
| Backend | `backend/pkg/services/data` | 故障现场日志分页接口使用的数据组权限校验。 |
| Backend | `backend/pkg/model/request` | 日志模块请求结构。 |
| Backend | `backend/pkg/model/response` | 日志模块响应结构。 |
| Backend | `backend/pkg/repository/database` | `logtableinfo` 与 `otherlogtable` DAO。 |
| Backend | `backend/pkg/repository/clickhouse` | 实际日志查询、字段分析、上下文与外部表元信息来源。 |
| Backend | `backend/pkg/repository/kubernetes` | 解析规则新增/更新/删除时更新 Vector ConfigMap。 |
| Backend | `backend/pkg/repository/prometheus` | `GetServiceRoute` 通过最近 7 天活跃实例推断路由规则。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/oss/routes.js` | `/logs/fault-site`, `/logs/full` |
| Backend | `backend/pkg/router/router_api.go` | `/api/log/fault/pagelist`, `/api/log/fault/content`, `/api/log/context`, `/api/log/query`, `/api/log/chart`, `/api/log/index`, `/api/log/table`, `/api/log/rule/service`, `/api/log/rule/get`, `/api/log/rule/update`, `/api/log/rule/add`, `/api/log/rule/delete`, `/api/log/other`, `/api/log/other/table`, `/api/log/other/add`, `/api/log/other/delete` |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/oss/views/logs/FaultSiteLogsPage.tsx` | 用 `LogsTraceFilterProvider` 包装故障现场日志页。 |
| Route Page | `frontend/src/oss/views/logs/FullLogsPage.jsx` | 用 `LogsProvider` 包装全量日志页。 |
| Page Core | `frontend/src/oss/views/logs/FaultSiteLogs.jsx` | 故障现场日志主页面，负责分页列表与单条日志内容。 |
| Page Core | `frontend/src/oss/views/logs/component/FullLogs/index.jsx` | 全量日志三栏布局装配入口。 |
| Filter | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | 故障现场日志筛选器，内含 `DataSourceFilter` 与时间选择。 |
| Result | `frontend/src/oss/views/logs/component/FullLogs/component/LogQueryResult/index.jsx` | 日志列表、分页、上下文弹窗入口。 |
| Result | `frontend/src/oss/views/logs/component/FullLogs/component/LogQueryResult/ContextModal/index.jsx` | 日志上下文弹窗，调用 `/api/log/context`。 |
| Search | `frontend/src/oss/views/logs/component/FullLogs/component/SerarchBar/RawLogQuery/index.jsx` | 原始查询输入、搜索触发。 |
| Index Analysis | `frontend/src/oss/views/logs/component/FullLogs/component/IndexList/index.jsx` | 字段区分为基础字段与日志字段，支持批量显隐。 |
| Index Analysis | `frontend/src/oss/views/logs/component/FullLogs/component/IndexList/component/IndexCollapseItem.jsx` | 点击字段值后把过滤条件拼进查询。 |
| Rule Config | `frontend/src/oss/views/logs/component/FullLogs/component/ConfigLogRuleModal/index.jsx` | 新增/编辑解析规则。 |
| External Table Config | `frontend/src/oss/views/logs/component/FullLogs/component/ConfigTableModal/index.jsx` | 选择外部日志表并登记时间字段。 |
| Side Tree | `frontend/src/oss/views/logs/component/FullLogs/component/Sider/LogRuleList/index.jsx` | 解析规则树与删除入口。 |
| Side Tree | `frontend/src/oss/views/logs/component/FullLogs/component/Sider/DataSourceTree/index.jsx` | 外部表树与删除入口。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/logs.js` | `getLogPageListApi` | `/api/log/fault/pagelist` | 故障现场日志分页列表。 |
| `frontend/src/core/api/logs.js` | `getLogContentApi` | `/api/log/fault/content` | 单次故障现场日志内容。 |
| `frontend/src/core/api/logs.js` | `getFullLogApi` | `/api/log/query` | 全量日志查询。 |
| `frontend/src/core/api/logs.js` | `getFullLogChartApi` | `/api/log/chart` | 全量日志直方图。 |
| `frontend/src/core/api/logs.js` | `getLogIndexApi` | `/api/log/index` | 字段值分布分析。 |
| `frontend/src/core/api/logs.js` | `getLogContextApi` | `/api/log/context` | 单条日志上下文。 |
| `frontend/src/core/api/logs.js` | `getLogTableInfoAPi` | `/api/log/table` | 获取解析规则树与外部表树基础信息。 |
| `frontend/src/core/api/logs.js` | `getLogRuleApi` | `/api/log/rule/get` | 读取指定解析规则。 |
| `frontend/src/core/api/logs.js` | `addLogRuleApi` | `/api/log/rule/add` | 新增解析规则。 |
| `frontend/src/core/api/logs.js` | `updateLogRuleApi` | `/api/log/rule/update` | 更新解析规则。 |
| `frontend/src/core/api/logs.js` | `deleteLogRuleApi` | `/api/log/rule/delete` | 删除解析规则。 |
| `frontend/src/core/api/logs.js` | `getLogRuleServiceRouteRuleApi` | `/api/log/rule/service` | 根据 service 反查路由规则。 |
| `frontend/src/core/api/logs.js` | `getLogOtherTableListApi` | `/api/log/other` | 浏览可登记的外部日志表。 |
| `frontend/src/core/api/logs.js` | `getLogOtherTableInfoApi` | `/api/log/other/table` | 查询外部表列信息。 |
| `frontend/src/core/api/logs.js` | `addLogOtherTableApi` | `/api/log/other/add` | 把外部表登记到系统。 |
| `frontend/src/core/api/logs.js` | `deleteLogOtherTableApi` | `/api/log/other/delete` | 删除外部表登记。 |
| `frontend/src/core/api/service.js` | `getServiceListApi` | `/api/service/list` | 解析规则配置页用来选择 service。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router | `backend/pkg/router/router_api.go` | `/api/log` route group 同时挂载 `AuthMiddleware` 与 `LicMiddleware`。 |
| Handler Root | `backend/pkg/api/log/handler.go` | 构建 `logService`，并在启动时执行 `InitParseLogTable(core.EmptyCtx(), req)`。 |
| Handler | `backend/pkg/api/log/func_getfaultloglist.go` | `GetFaultLogPageList`，带 `CheckGroupPermission`。 |
| Handler | `backend/pkg/api/log/func_getfaultlogcontent.go` | `GetFaultLogContent`，代码中明确标注 `service` 未用于权限校验。 |
| Handler | `backend/pkg/api/log/func_querylog.go` | `QueryLog`，补默认分页、查询语句、`timestamp`、`content`。 |
| Handler | `backend/pkg/api/log/func_getlogchart.go` | `GetLogChart`，补默认查询语句、时间字段、日志字段。 |
| Handler | `backend/pkg/api/log/func_getlogindex.go` | `GetLogIndex`，补默认查询语句、时间字段、日志字段。 |
| Handler | `backend/pkg/api/log/func_querylogcontext.go` | `QueryLogContext`。 |
| Handler | `backend/pkg/api/log/func_getlogtableinfo.go` | `GetLogTableInfo`。 |
| Handler | `backend/pkg/api/log/func_getserviceroute.go` | `GetServiceRoute`。 |
| Handler | `backend/pkg/api/log/func_getlogparserule.go` | `GetLogParseRule`。 |
| Handler | `backend/pkg/api/log/func_addparserule.go` | `AddLogParseRule`。 |
| Handler | `backend/pkg/api/log/func_updatelogparserule.go` | `UpdateLogParseRule`。 |
| Handler | `backend/pkg/api/log/func_deleteparserule.go` | `DeleteLogParseRule`。 |
| Handler | `backend/pkg/api/log/func_othertable.go` | `OtherTable`。 |
| Handler | `backend/pkg/api/log/func_othertableinfo.go` | `OtherTableInfo`。 |
| Handler | `backend/pkg/api/log/func_addothertable.go` | `AddOtherTable`，会先执行 `req.FillerValue()`。 |
| Handler | `backend/pkg/api/log/func_deleteothertable.go` | `DeleteOtherTable`。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Fault Site Pagination | `backend/pkg/services/log/service_get_fault_log_page_list.go` | 故障现场分页列表，必要时按数据组自动补齐 service 范围。 |
| Fault Site Content | `backend/pkg/services/log/service_get_fault_log_contents.go` | 先查应用日志，查不到时 fallback 到 `raw_logs`。 |
| Full Query | `backend/pkg/services/log/service_query_log.go` | 全量日志查询，支持大偏移回算与内外部表字段拆分。 |
| Log Context | `backend/pkg/services/log/service_query_log_context.go` | 返回上下文前后日志。 |
| Histogram | `backend/pkg/services/log/service_get_log_chart.go` | 生成日志直方图并补齐空桶。 |
| Field Analysis | `backend/pkg/services/log/service_log_index.go` | 统计字段值分布与百分比。 |
| Table Tree | `backend/pkg/services/log/service_get_log_table_info.go` | 组合解析规则树和外部表树。 |
| Parse Rule Read | `backend/pkg/services/log/service_get_log_parse_rule.go` | 读取 `logtableinfo` 并反序列化规则信息。 |
| Parse Rule Add | `backend/pkg/services/log/service_add_log_parse_rule.go` | 创建 ClickHouse 表、更新 Vector ConfigMap、写入 `logtableinfo`。 |
| Parse Rule Update | `backend/pkg/services/log/service_update_log_parse_rule.go` | 更新 ClickHouse 表结构、更新 Vector ConfigMap、更新 `logtableinfo`。 |
| Parse Rule Delete | `backend/pkg/services/log/service_delete_log_parse_rule.go` | 删除 Vector 配置、删除 ClickHouse 表、删除 `logtableinfo`。 |
| Route Rule Guess | `backend/pkg/services/log/service_get_service_route.go` | 基于最近 7 天活跃实例名推断 `k8s.pod.name` 规则。 |
| External Table List | `backend/pkg/services/log/service_other_log.go` | 枚举 ClickHouse 可见外部日志表。 |
| External Table Columns | `backend/pkg/services/log/service_other_log_info.go` | 查询指定外部表列定义。 |
| External Table Add | `backend/pkg/services/log/service_add_other_table.go` | 只写入元数据，不创建真实表。 |
| External Table Delete | `backend/pkg/services/log/service_delete_other_table.go` | 删除元数据登记。 |
| Parse Table Bootstrap | `backend/pkg/services/log/service_log_table.go` | 默认 `raw_logs`/`all_logs` 解析表初始化与结构更新。 |
| Vector Config | `backend/pkg/services/log/vector/factory.go` | 操作 `route_logs`、`parse_<name>`、`to_<name>` 节点。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/log_req.go` | 故障现场日志分页与内容请求。 |
| Request | `backend/pkg/model/request/log_query.go` | 全量查询、字段索引、上下文请求。 |
| Request | `backend/pkg/model/request/log_parse.go` | 解析规则新增、更新、删除、读取请求。 |
| Request | `backend/pkg/model/request/log_other.go` | 外部表读取、登记、删除请求。 |
| Request | `backend/pkg/model/request/log_table.go` | 解析日志表结构定义与默认值填充。 |
| Response | `backend/pkg/model/response/log_resp.go` | 故障现场日志分页与内容响应。 |
| Response | `backend/pkg/model/response/log_query.go` | 全量日志查询与上下文响应。 |
| Response | `backend/pkg/model/response/log_chart.go` | 直方图响应。 |
| Response | `backend/pkg/model/response/log_index.go` | 字段索引响应。 |
| Response | `backend/pkg/model/response/log_parse.go` | 解析规则与 service 路由规则响应。 |
| Response | `backend/pkg/model/response/log_other.go` | 外部表列表与列信息响应。 |
| DAO | `backend/pkg/repository/database/dao_logtableinfo.go` | `logtableinfo` 表定义，保存解析规则元数据。 |
| DAO | `backend/pkg/repository/database/dao_othertableinfo.go` | `otherlogtable` 表定义，保存外部表接入元数据。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Data Scope UI | `frontend/src/core/components/Filter/DataSourceFilter.tsx` | `/logs/fault-site` 依赖共享数据组过滤树，而不是自建筛选数据源。 |
| Time Range | `frontend/src/core/store/reducers/timeRangeReducer.js` | `/logs/full` 完全依赖共享系统时间范围。 |
| Menu And Route Permission | `frontend/src/core/contexts/UserContext.js`, `frontend/src/core/components/AuthRouter.tsx`, `frontend/src/core/api/permission.js` | 菜单来自 `/api/permission/config`，但客户端路由权限检查当前被注释。 |
| Service Module API | `frontend/src/core/api/service.js` | 解析规则配置页要调用 `/api/service/list`。 |
| Request Utility | `frontend/src/core/utils/request` | 全部前端 wrapper 共用的请求入口。 |
| ClickHouse Repo | `backend/pkg/repository/clickhouse` | 故障现场、全量查询、字段索引、上下文、外部表信息都依赖它。 |
| Database Repo | `backend/pkg/repository/database` | 解析规则与外部表元数据都持久化到这里。 |
| Kubernetes Repo | `backend/pkg/repository/kubernetes` | 解析规则生命周期直接改 Vector ConfigMap。 |
| Prometheus Repo | `backend/pkg/repository/prometheus` | `GetServiceRoute` 要求最近 7 天活跃实例。 |
| Data Permission Service | `backend/pkg/services/data` | 仅故障现场分页接口显式使用了 `CheckGroupPermission`。 |
| Cross Module Link | `frontend/src/oss/views/serviceInfo/component/infoUni/ErrorInstance/ErrorInstanceInfo.jsx`, `frontend/src/oss/views/serviceInfo/component/infoUni/LogsInfo.jsx` | 其他模块会跳转到 `#/logs/fault-site`，使本模块成为排障落点。 |

## Evidence Notes

1. `/logs/fault-site` 与 `/logs/full` 的状态模型明显不同：前者依赖 `LogsTraceFilterContext` 与 `dataGroupId`，后者依赖 `LogsContext`、`logsReducer` 与系统时间范围。
2. `GetFaultLogContent` handler 中有显式 TODO，说明当前 `serviceName` 未参与数据权限校验。
3. `QueryLog`、`GetLogChart`、`GetLogIndex`、解析规则接口、外部表接口的 handler 代码中未观察到与 `groupId` 类似的数据范围校验。
4. `handler.New(...)` 在启动阶段调用 `InitParseLogTable(core.EmptyCtx(), req)`，意味着默认解析日志表不是纯前端初始化行为，而是服务启动时的后端自举逻辑。
5. `ConfigTableModal` 前端只收集 `dataBase`、`tableName`、`timeField`，但后端 `AddOtherTableRequest` 还支持 `cluster`、`instance`、`logField`，存在前后端信息收集不对称。
