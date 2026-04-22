# Dashboards And Health Monitoring Frontend

## Pages

| Page | File | Description |
|---|---|---|
| `/system-dashboard` | `frontend/src/oss/views/dashboard/SystemDashboard.jsx` | 全局资源 dashboard 容器。 |
| `/basic-dashboard` | `frontend/src/oss/views/dashboard/BasicDashboard.jsx` | 基础设施 dashboard 容器。 |
| `/application-dashboard` | `frontend/src/oss/views/dashboard/ApplicationDashboard.jsx` | 应用指标 dashboard 容器。 |
| `/middleware-dashboard` | `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx` | 中间件 dashboard 容器，额外展示文档链接提示。 |
| `/mysql-dashboard` | `frontend/src/pro/views/MysqlDashboard/index.jsx` | MySQL dashboard 容器。 |
| `/healthy-service` | `frontend/src/pro/views/serviceHealthy/index.js` | 服务健康状态总览页。 |

## Routes

| Route | Route File Evidence | Route Metadata | Notes |
|---|---|---|---|
| `/system-dashboard` | `frontend/src/oss/routes.js` | 无 `hideSystemTimeRangePicker` | 走系统时间范围。 |
| `/basic-dashboard` | `frontend/src/oss/routes.js` | 无 `hideSystemTimeRangePicker` | 同上。 |
| `/application-dashboard` | `frontend/src/oss/routes.js` | 无 `hideSystemTimeRangePicker` | 同上。 |
| `/middleware-dashboard` | `frontend/src/oss/routes.js` | 无 `hideSystemTimeRangePicker` | 同上。 |
| `/mysql-dashboard` | `frontend/src/pro/routes.js` | 无特殊 metadata | Pro 页面。 |
| `/healthy-service` | `frontend/src/pro/routes.js` | `showDataGroup: true` | 健康页显式使用数据组。 |

## Menus

| Item | Evidence | Notes |
|---|---|---|
| `system` | `backend/pkg/repository/database/init_menu_item.go` + `init_feature_mapping.go` | 映射到 `/system-dashboard`。 |
| `basic` | 同上 | 映射到 `/basic-dashboard`。 |
| `application` | 同上 | 映射到 `/application-dashboard`。 |
| `middleware` | 同上 | 映射到 `/middleware-dashboard`。 |
| `mysql` | 同上 | 映射到 `/mysql-dashboard`。 |
| `healthy` | 同上 | 映射到 `/healthy-service`。 |

实现细节：

- dashboard 页面真正依赖的是 `UserContext` 拉下来的 `menuItems`。
- `IframeDashboard` 会在 `menuItems` 里按 `dashboardKey` 找 `item.router.page.url`。
- 所以菜单配置不只是导航用途，也承担 dashboard 页面 URL 配置源。

## Key Components

| Component | File | Responsibility |
|---|---|---|
| `IframeDashboard` | `frontend/src/core/components/Dashboard/IframeDashboard.jsx` | 统一渲染 dashboard iframe，并自动拼时间范围与主题参数。 |
| `SystemDashboard` | `frontend/src/oss/views/dashboard/SystemDashboard.jsx` | 用 `dashboardKey="system"` 交给 `IframeDashboard`。 |
| `BasicDashboard` | `frontend/src/oss/views/dashboard/BasicDashboard.jsx` | 用 `dashboardKey="basic"`。 |
| `ApplicationDashboard` | `frontend/src/oss/views/dashboard/ApplicationDashboard.jsx` | 用 `dashboardKey="application"`。 |
| `MiddlewareDashboard` | `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx` | 用 `dashboardKey="middleware"`，同时展示说明性 `Alert`。 |
| `MysqlDashboard` | `frontend/src/pro/views/MysqlDashboard/index.jsx` | 试图直接指定 MySQL dashboard URL。 |
| `ServiceHealthyPage` | `frontend/src/pro/views/serviceHealthy/index.js` | 渲染健康状态表格与搜索输入。 |

## State Management

### Dashboard Pages

| State Container | Evidence | Notes |
|---|---|---|
| `userReducer.menuItems` | `IframeDashboard.jsx` | dashboard URL 来自 `menuItems.find(item.key === dashboardKey)`。 |
| `timeRangeReducer` | `selectProcessedTimeRange`, `state.timeRange` | dashboard URL 会拼 `from`、`to`。 |
| `settingReducer.theme` | `IframeDashboard.jsx` | dashboard URL 会追加 `theme`。 |
| Local State | `const [src, setSrc] = useState(...)` | 由 `IframeDashboard` 计算最终 iframe 地址。 |

### Health Page

| State Container | Evidence | Notes |
|---|---|---|
| `timeRangeReducer` | `selectProcessedTimeRange` in `serviceHealthy/index.js` | 查询窗口来自共享时间范围。 |
| `dataGroupReducer` | `const { dataGroupId } = useSelector(...)` | 用于数据组过滤。 |
| Local State | `data`, `serachServiceName`, `serachEndpointName`, `serachNamespace`, `loading` | 当前实际页面只渲染了 `serviceName` 输入。 |
| Debounced Fetch | `useDebounce(...)` | 时间范围与数据组变化 300ms 后自动请求。 |

## API Usage

| Frontend Action | Wrapper Or Mechanism | Backend API Or Target | Notes |
|---|---|---|---|
| 打开四个 OSS dashboard | `IframeDashboard dashboardKey=...` | `router.page.url` from `menuItems` | 不走模块专属业务 API。 |
| 打开 MySQL dashboard | `IframeDashboard src=...` | 期望目标 `grafana/d/0D6dTg3Zk/mysql-e68c87-e6a087?orgId=1` | 当前传参名与组件定义不一致。 |
| 查询服务健康状态 | `getServiceHealthyApi` | `/api/service/ryglight` | 当前健康页唯一直接消费的模块 API。 |
| 获取 dashboard URL 元数据 | `UserContext -> getUserPermissionApi` | `/api/permission/config` | 属于共享权限基础设施，但对 dashboard 页面是必需依赖。 |

## User Interaction Flow

### Dashboard Rendering Flow

1. 用户进入某个 dashboard 路由。
2. 页面只渲染一个 `IframeDashboard`。
3. `IframeDashboard` 从 `menuItems` 中找到对应 `dashboardKey` 的 `router.page.url`。
4. 组件根据当前时间范围追加 `from`、`to` 参数。
5. 组件根据当前主题追加 `theme` 参数。
6. 最终渲染 iframe。

### Service Health Flow

1. 用户进入 `/healthy-service`。
2. 页面从 `timeRangeReducer` 和 `dataGroupReducer` 读取时间范围与数据组。
3. 页面调用 `/api/service/ryglight`。
4. 后端返回 `serviceList`，每条 service 包含 `percentScore`、`status` 和 `scoreDetail`。
5. 前端表格展示 service 名称、健康灯颜色、百分制健康分和展开式评分详情。

## Permissions

### Route And Menu Layer

- 所有页面都通过 `AuthRouter` 包装。
- 但 `AuthRouter` 中实际调用 `getRouterPermissionApi` 的校验逻辑当前被注释，所以前端 guard 只保留壳层。
- 页面是否能在导航里出现，更依赖后端 `/api/permission/config` 下发的 `menuItems`。

### Data Scope Layer

- `/healthy-service` 路由显式标记 `showDataGroup: true`。
- 健康页会从 `dataGroupReducer` 读取 `dataGroupId`，但当前请求参数名是 `dataGroupId`，后端期待 `groupId`。
- 四个 OSS dashboard 和 MySQL dashboard 页面当前未表现出独立的数据组交互。

### UI-Level Constraints

- dashboard 页面内容本质上是外部 Grafana / Jaeger URL，因此部分访问体验和权限结果依赖外部可视化页面本身。

## Notes

1. `MiddlewareDashboard` 是唯一有额外文档提示和外链说明的 dashboard 页面。
2. `IframeDashboard` 里有较多 DOM `MutationObserver` 和 iframe 内部元素隐藏逻辑，说明它不仅负责嵌入，还尝试裁剪 Grafana UI。
3. `MysqlDashboard` 与 `pro/components/DataViewCom/MetricsDashboard.jsx` 传的是 `src` 而不是 `srcProp`，当前实现存在可疑接线。
4. `/api/service/monitor/status` 在当前主前端代码中没有发现直接调用。
