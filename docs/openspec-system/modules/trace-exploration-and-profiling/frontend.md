# Trace Exploration And Profiling Frontend

## Pages

| Page | File | Description |
|---|---|---|
| `/trace/fault-site` | `frontend/src/oss/views/trace/FaultSiteTracePage.tsx` | 用共享筛选上下文包装 `FaultSiteTrace`，承载故障现场链路检索。 |
| `/trace/full` | `frontend/src/oss/views/trace/FullTrace.jsx` | 直接展示 `/jaeger/search` iframe。 |

## Routes

| Route | Route File Evidence | Route Metadata | Notes |
|---|---|---|---|
| `/trace/fault-site` | `frontend/src/oss/routes.js` | `hideSystemTimeRangePicker: true`, `showDataGroup: true` | 页面内自行渲染 `LogsTraceFilter`。 |
| `/trace/full` | `frontend/src/oss/routes.js` | `hideSystemTimeRangePicker: true` | 页面不使用系统时间范围选择器，也不显示数据组面板。 |

## Menus

| Item | Evidence | Notes |
|---|---|---|
| `/trace/fault-site` | `name: <TranslationCom text="faultSiteTraces" ... />` in `routes.js` | 实际是否显示取决于 `/api/permission/config` 返回的菜单权限。 |
| `/trace/full` | `name: <TranslationCom text="allTrace" ... />` in `routes.js` | 同样依赖统一权限菜单模型。 |

补充证据：

- `frontend/src/_nav.js` 仍保留静态菜单 `故障现场链路` 和 `全量链路`，但当前主菜单体系更依赖 `UserContext` 从后端拉取的 `menuItems`。

## Key Components

| Component | File | Responsibility |
|---|---|---|
| `FaultSiteTrace` | `frontend/src/oss/views/trace/FaultSiteTrace.jsx` | 故障现场 Trace 主页面，负责查询、分页、打开 Jaeger 弹窗、跳转日志页。 |
| `LogsTraceFilter` | `frontend/src/oss/components/Filter/LogsTraceFilter.jsx` | 提供 cluster、namespace、service、instance、`Trace ID`、duration、fault type、时间范围。 |
| `JaegerIframeModal` | `frontend/src/oss/views/trace/component/JaegerIframeModal.jsx` | 通过 `/jaeger/trace/<traceId>?uiEmbed=v0` 展示单条 Trace。 |
| `TraceErrorType` | `frontend/src/oss/views/trace/component/TraceErrorType.jsx` | 把 `slow`、`error`、`normal`、`slowAndError` 渲染为状态标签。 |
| `BasicTable` | `src/core/components/Table/basicTable` | 渲染故障现场 Trace 表格。 |
| `FullTrace` | `frontend/src/oss/views/trace/FullTrace.jsx` | 直接承载 Jaeger 搜索页。 |

## State Management

### `/trace/fault-site`

| State Container | Evidence | Notes |
|---|---|---|
| `LogsTraceFilterContext` | `useLogsTraceFilterContext(...)` in `FaultSiteTrace.jsx` | 保存 `clusterIds`、`services`、`instance`、`traceId`、`namespaces`、`startTime`、`endTime`、`minDuration`、`maxDuration`、`faultTypeList`、`endpoint`。 |
| `dataGroupReducer` | `useSelector((state) => state.dataGroupReducer)` | `groupId` 被写入 `/api/trace/pagelist`。 |
| Local State | `useState(...)` in `FaultSiteTrace.jsx` | 保存 `tracePageList`、`pageIndex`、`pageSize`、`total`、`modalVisible`、`selectTraceId`。 |
| Debounced Fetch | `useDebounce(...)` | 筛选条件变化 300ms 后触发查询。 |

几个实现细节：

- 当前前端没有单独的 `traceReducer` 或 `TraceContext`。
- `instance` 在上下文里是对象数组，页面会把其中的 `containerId`、`node`、`pid`、`id` 拆出来组装请求。
- `traceId` 会同时用作查询条件和打开 Jaeger 弹窗后的默认目标。

### `/trace/full`

`/trace/full` 没有额外状态容器，页面只是一个纯 iframe 容器。

## API Usage

| Frontend Action | Wrapper Or Mechanism | Backend API Or Target | Notes |
|---|---|---|---|
| 故障现场 Trace 分页查询 | `getTracePageListApi` | `/api/trace/pagelist` | 当前前端唯一直接消费的 trace API。 |
| 查看单条 Trace 详情 | `JaegerIframeModal` iframe | `/jaeger/trace/<traceId>?uiEmbed=v0` | 不经过 `/api/trace/info`。 |
| 查看全量 Trace 搜索 | `FullTrace` iframe | `/jaeger/search` | 直接嵌入 Jaeger 搜索页。 |
| 从 Trace 跳到故障现场日志 | `window.open(targetUrl)` | `#/logs/fault-site?...` | 由 `FaultSiteTrace` 表格操作列触发。 |
| 从 service 时间线跳到 fault-site trace | `window.open(fullUrl, '_blank')` in `TimeLine.jsx` | `#/trace/fault-site?...` | 会额外设置 `sessionStorage.openJaegerModalAfterLoad`。 |

## User Interaction Flow

### `/trace/fault-site`

1. 页面加载后，`FaultSiteTracePage` 提供 `LogsTraceFilterProvider`。
2. 用户通过 `LogsTraceFilter type="trace"` 选择 cluster、namespace、service、instance、`Trace ID`、duration、fault type 和时间范围。
3. `FaultSiteTrace` 组装查询参数并调用 `/api/trace/pagelist`。
4. 返回的分页结果渲染到 `BasicTable`。
5. 点击 `traceId` 打开 `JaegerIframeModal`。
6. 点击操作列按钮，可把当前 Trace 跳转到 `#/logs/fault-site` 做日志联查。

### `/trace/full`

1. 用户进入页面。
2. 页面直接嵌入 `/jaeger/search`。
3. 后续交互主要发生在 Jaeger 自身界面，而不是当前前端组件内部。

### Cross-Module Entry

1. 在 service 详情时间线点击 trace 项时，会打开 `#/trace/fault-site`。
2. 如果时间线中的条目本身包含 `traceId`，代码会先设置 `sessionStorage.openJaegerModalAfterLoad = 'true'`。
3. `FaultSiteTrace` 在首轮数据加载后检测这个标记，并自动打开 Jaeger 弹窗。

## Permissions

### Route And Menu Layer

- 两个页面都通过 `AuthRouter(...)` 包装。
- 但 `frontend/src/core/components/AuthRouter.tsx` 内真正执行 `getRouterPermissionApi` 的 `useEffect` 目前被注释，所以客户端路由守卫没有实际生效。
- 菜单显隐仍由 `UserContext` 从 `/api/permission/config` 拉取的 `menuItems` 控制。

### Data Scope Layer

- `/trace/fault-site` 路由显式标记 `showDataGroup: true`，页面请求也会传 `groupId`。
- `/trace/full` 没有 `showDataGroup`，当前也没有通过前端请求与数据组联动。
- 这意味着 fault-site trace 和 full trace 的数据范围模型并不一致。

### UI-Level Constraints

- `/trace/full` 的实际内容是 Jaeger iframe，因此 UI 权限边界部分取决于 Jaeger 接入与代理方式，而不是只取决于当前 React 组件。

## Notes

1. `frontend/src/oss/views/trace/TraceMoreFilters.jsx` 当前未观察到引用，不能作为已接线功能描述。
2. `traceTableMock` 被用于 `traceTableMock && <BasicTable ... />` 这种渲染开关，显示存在残留调试代码。
3. 当前前端没有消费 `/api/trace/pagelist/filters`、`/api/trace/pagelist/filter/value`、`/api/trace/info`、profiling/flame graph 接口。
