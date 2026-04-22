# Logs Exploration And Parsing Frontend

## Pages

| Page | File | Description |
|---|---|---|
| `/logs/fault-site` | `frontend/src/oss/views/logs/FaultSiteLogsPage.tsx` | 用 `LogsTraceFilterProvider` 包装 `FaultSiteLogs`，强调故障现场筛选与单次日志回放。 |
| `/logs/full` | `frontend/src/oss/views/logs/FullLogsPage.jsx` | 用 `LogsProvider` 包装 `FullLogs`，强调全量检索、字段分析、规则配置与外部表管理。 |

## Routes

| Route | Route File Evidence | Route Metadata | Notes |
|---|---|---|---|
| `/logs/fault-site` | `frontend/src/oss/routes.js` | `hideSystemTimeRangePicker: true`, `showDataGroup: true` | 页面自己内嵌 `LogsTraceFilter` 与时间选择器。 |
| `/logs/full` | `frontend/src/oss/routes.js` | 无 `showDataGroup`，无 `hideSystemTimeRangePicker` | 使用系统级时间范围，不展示独立数据组面板。 |

## Menus

| Item | Evidence | Notes |
|---|---|---|
| `/logs/fault-site` | `name: <TranslationCom text="faultLogsName" ... />` | 菜单文案来自 i18n，是否展示仍取决于 `/api/permission/config` 返回的菜单配置。 |
| `/logs/full` | `name: <TranslationCom text="allLogsName" ... />` | 同样走统一菜单权限体系。 |

当前代码中没有本模块专属的菜单组装逻辑；侧边栏实际展示由 `frontend/src/core/contexts/UserContext.js` 拉取的 `menuItems` 决定。

## Key Components

### Fault Site View

| Component | File | Responsibility |
|---|---|---|
| `LogsTraceFilter` | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | 组合 `DataSourceFilter`、`Trace ID` 输入框与 `DateTimeRangePickerCom`。 |
| `FaultSiteLogs` | `frontend/src/oss/views/logs/FaultSiteLogs.jsx` | 负责分页查询、当前条目切换、日志内容请求。 |
| `LogContent` | `frontend/src/oss/views/logs/component/LogContent.jsx` | 负责故障现场日志正文展示。 |

### Full Logs View

| Component | File | Responsibility |
|---|---|---|
| `FullLogs` | `frontend/src/oss/views/logs/component/FullLogs/index.jsx` | 三栏布局装配，协调筛选、字段、结果区。 |
| `RawLogQuery` | `frontend/src/oss/views/logs/component/FullLogs/component/SerarchBar/RawLogQuery/index.jsx` | 查询语句输入与触发搜索。 |
| `IndexList` | `frontend/src/oss/views/logs/component/FullLogs/component/IndexList/index.jsx` | 展示字段树与显隐开关。 |
| `IndexCollapseItem` | `frontend/src/oss/views/logs/component/FullLogs/component/IndexList/component/IndexCollapseItem.jsx` | 读取字段值分布，并把选中的值追加到查询语句。 |
| `LogQueryResult` | `frontend/src/oss/views/logs/component/FullLogs/component/LogQueryResult/index.jsx` | 结果列表、分页、上下文弹窗。 |
| `ContextModal` | `frontend/src/oss/views/logs/component/FullLogs/component/LogQueryResult/ContextModal/index.jsx` | 查看单条日志前后上下文。 |
| `LogRuleList` | `frontend/src/oss/views/logs/component/FullLogs/component/Sider/LogRuleList/index.jsx` | 解析规则树、编辑、删除。 |
| `DataSourceTree` | `frontend/src/oss/views/logs/component/FullLogs/component/Sider/DataSourceTree/index.jsx` | 外部表树、切换、删除。 |
| `ConfigLogRuleModal` | `frontend/src/oss/views/logs/component/FullLogs/component/ConfigLogRuleModal/index.jsx` | 新增或编辑解析规则。 |
| `ConfigTableModal` | `frontend/src/oss/views/logs/component/FullLogs/component/ConfigTableModal/index.jsx` | 登记外部日志表。 |

## State Management

### Fault Site

| State Container | Evidence | Notes |
|---|---|---|
| `LogsTraceFilterContext` | `frontend/src/oss/contexts/LogsTraceFilterContext.tsx` | 保存 `clusterIds`、`services`、`instance`、`traceId`、`namespaces`、`startTime`、`endTime` 等筛选条件。 |
| `dataGroupReducer` | `useSelector((state) => state.dataGroupReducer)` in `FaultSiteLogs.jsx` | 把 `dataGroupId` 注入 `/api/log/fault/pagelist`。 |
| `useDebounce` | `FaultSiteLogs.jsx` | 筛选条件变化 300ms 后触发重新查询。 |

`FaultSiteLogs` 会从选中的 instance 对象里解构 `containerId`、`node`、`pid`、`id` 并拼进请求参数，因此故障现场查询不是单纯的字符串过滤，而是对实例对象做二次展开。

### Full Logs

| State Container | Evidence | Notes |
|---|---|---|
| `LogsContext` | `frontend/src/core/contexts/LogsContext.js` | 暴露 `fetchData`、`getLogTableInfo`、`getFieldIndexData` 以及更新方法。 |
| `logsReducer` | `frontend/src/core/store/reducers/logsReducer.js` | 保存 `tableInfo`、`logRules`、`instances`、`logs`、`pagination`、`defaultFields`、`hiddenFields`、`displayFields`、`query`、`searchValue`。 |
| `timeRangeReducer` | `selectProcessedTimeRange` | 全量日志查询依赖系统统一时间范围。 |

几个关键状态约定：

- `displayFields` 以 `tableName + "_" + type` 为 key，说明同一页面会在解析日志表和外部表之间保留独立显示配置。
- `tableInfo.type === 'logLibrary'` 表示内置解析日志表。
- `tableInfo.type === 'database'` 表示外部表。
- `tableInfo.timeField` 只在外部表模式下被设置，前端据此隐藏上下文按钮。

## API Usage

| Frontend Action | Wrapper | Backend API | Notes |
|---|---|---|---|
| 故障现场分页列表 | `getLogPageListApi` | `/api/log/fault/pagelist` | 请求中带 `groupId`。 |
| 故障现场日志内容 | `getLogContentApi` | `/api/log/fault/content` | 当前 source 会传到 `sourceFrom`。 |
| 全量日志查询 | `getFullLogApi` | `/api/log/query` | 与图表并行请求。 |
| 全量日志直方图 | `getFullLogChartApi` | `/api/log/chart` | 结果用于 `Histogram` 与大偏移回算。 |
| 字段值分布分析 | `getLogIndexApi` | `/api/log/index` | `IndexCollapseItem` 懒加载。 |
| 查看上下文 | `getLogContextApi` | `/api/log/context` | 仅内置解析日志表显示入口。 |
| 读取规则树与外部表树 | `getLogTableInfoAPi` | `/api/log/table` | 页面初始化即调用。 |
| 读取解析规则 | `getLogRuleApi` | `/api/log/rule/get` | 编辑规则时使用。 |
| 自动补 route rule | `getLogRuleServiceRouteRuleApi` | `/api/log/rule/service` | service 选择变化时触发。 |
| 新增解析规则 | `addLogRuleApi` | `/api/log/rule/add` | 保存后刷新规则树。 |
| 更新解析规则 | `updateLogRuleApi` | `/api/log/rule/update` | 保存后刷新规则树。 |
| 删除解析规则 | `deleteLogRuleApi` | `/api/log/rule/delete` | 侧边栏树节点删除。 |
| 浏览外部表 | `getLogOtherTableListApi` | `/api/log/other` | 弹窗打开时加载。 |
| 获取外部表字段 | `getLogOtherTableInfoApi` | `/api/log/other/table` | 选择表后加载列定义。 |
| 登记外部表 | `addLogOtherTableApi` | `/api/log/other/add` | 保存后刷新侧边栏树。 |
| 删除外部表登记 | `deleteLogOtherTableApi` | `/api/log/other/delete` | 外部表树节点删除。 |
| 读取 service 列表 | `getServiceListApi` | `/api/service/list` | 解析规则配置页附属依赖。 |

## User Interaction Flow

### `/logs/fault-site`

1. 页面加载后由 `LogsTraceFilterProvider` 提供筛选状态。
2. 用户通过 `DataSourceFilter` 选择 cluster、namespace、service、instance，并可输入 `Trace ID`。
3. `FaultSiteLogs` 监听筛选状态、时间范围与 `dataGroupId`，在条件完整后触发 `/api/log/fault/pagelist`。
4. 左侧标签页按故障时间点展示分页结果。
5. 切换条目后调用 `/api/log/fault/content`，右侧展示日志正文与 source 切换。

### `/logs/full`

1. 页面初始化先调用 `/api/log/table`，默认选中第一条解析规则对应的日志表。
2. 用户输入查询语句或切换时间范围，`LogsContext.fetchData()` 并行拉取 `/api/log/query` 和 `/api/log/chart`。
3. 左侧字段树按 `defaultFields` / `hiddenFields` 展示，并在展开时调用 `/api/log/index` 做字段值分布分析。
4. 点击某个字段值会把 `` `field` = 'value' `` 追加到查询语句中。
5. 点击单条日志的 `View Context` 会打开 `ContextModal` 并调用 `/api/log/context`。
6. 侧边栏可以在解析日志表与外部表之间切换，也可以打开规则配置和外部表接入弹窗。

## Permissions

### Route And Menu Layer

- 两个页面都通过 `AuthRouter(...)` 包装。
- 但 `frontend/src/core/components/AuthRouter.tsx` 中实际调用 `getRouterPermissionApi` 的 `useEffect` 已被注释，所以当前前端路由守卫并不真正阻断访问。
- 侧边栏菜单仍依赖 `UserContext` 从 `/api/permission/config` 拉取的 `menuItems`。

### Data Scope Layer

- `/logs/fault-site` 在路由定义中标记了 `showDataGroup: true`，页面内部也把 `dataGroupId` 传给 `/api/log/fault/pagelist`。
- `/logs/full` 路由定义中没有 `showDataGroup`，当前前端请求参数也不传 `groupId`。
- 现有实现显示，本模块两个页面的数据范围模型并不一致。

### UI-Level Constraints

- 上下文按钮只有在 `!tableInfo.timeField` 时才展示，因此外部表模式默认不开放上下文查询。
- 解析规则与外部表管理入口在界面上对所有进入 `/logs/full` 的用户可见，当前前端未观察到细粒度按钮级权限控制。

## Notes

1. `LogsTraceFilter` 是日志与 trace 页面复用组件；本模块只使用其 `type="logs"` 分支。
2. `RawLogQuery` 里还保留了被注释掉的 `FullTextSearch` 与刷新按钮，说明查询区经历过旧实现演化。
3. `ConfigTableModal` 当前只让用户选择表和时间字段，没有收集 `instance`、`cluster`、`logField`。
