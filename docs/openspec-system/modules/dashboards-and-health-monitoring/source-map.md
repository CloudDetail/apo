# Dashboards And Health Monitoring Source Map

## Scope
当前模块在代码中实际覆盖两类能力：

- 多个 Grafana / Jaeger 风格的大盘入口，包括 `/system-dashboard`、`/basic-dashboard`、`/application-dashboard`、`/middleware-dashboard`、`/mysql-dashboard`
- 服务健康状态页 `/healthy-service`，以及其后端支撑接口 `/api/service/ryglight`、`/api/service/monitor/status`

与多数业务模块不同，本模块有一半能力是“配置驱动 iframe 大盘”，而不是页面自己拼业务数据。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/oss/views/dashboard` | OSS 四个 dashboard 页面。 |
| Frontend | `frontend/src/pro/views/MysqlDashboard` | Pro 侧 MySQL dashboard 页面。 |
| Frontend | `frontend/src/pro/views/serviceHealthy` | Pro 侧服务健康状态页。 |
| Frontend | `frontend/src/core/components/Dashboard/IframeDashboard.jsx` | 所有 dashboard 页面共用的 iframe 载体。 |
| Frontend | `frontend/src/pro/api/serviceHealthy.js` | 健康状态页 API wrapper。 |
| Frontend | `frontend/src/core/contexts/UserContext.js` | `IframeDashboard` 依赖 `menuItems` 获取 dashboard URL。 |
| Frontend | `frontend/src/core/store/reducers/timeRangeReducer.js` | dashboard URL 拼接时间范围；健康页也使用共享时间范围。 |
| Backend | `backend/pkg/router/router_api.go` | 健康状态相关接口挂在 `/api/service`。 |
| Backend | `backend/pkg/api/serviceoverview` | `GetRYGLight`、`GetMonitorStatus` handler 所在目录。 |
| Backend | `backend/pkg/services/serviceoverview` | 服务健康评分与监控状态计算逻辑。 |
| Backend | `backend/pkg/model/request/service_req.go` | `GetRygLightRequest`、`GetMonitorStatusRequest`。 |
| Backend | `backend/pkg/model/response/service_resp.go` | `ServiceRYGLightRes`、`RYGResult`、`GetMonitorStatusResponse`。 |
| Backend | `backend/pkg/repository/database/init_router.go` | dashboard / healthy 路由初始化。 |
| Backend | `backend/pkg/repository/database/init_menu_item.go` | dashboard / healthy 菜单初始化。 |
| Backend | `backend/pkg/repository/database/init_feature_mapping.go` | dashboard / healthy 的 feature 到 menu mapping。 |
| Backend | `backend/pkg/repository/database/init_router_page.go` | dashboard 对应 Grafana / Jaeger 页面 URL 的初始化。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/oss/routes.js` | `/system-dashboard`, `/basic-dashboard`, `/application-dashboard`, `/middleware-dashboard` |
| Frontend | `frontend/src/pro/routes.js` | `/mysql-dashboard`, `/healthy-service` |
| Backend | `backend/pkg/router/router_api.go` | `/api/service/ryglight`, `/api/service/monitor/status` |

## Page Files

| Type | File | Notes |
|---|---|---|
| Dashboard Page | `frontend/src/oss/views/dashboard/SystemDashboard.jsx` | 使用 `IframeDashboard dashboardKey="system"`。 |
| Dashboard Page | `frontend/src/oss/views/dashboard/BasicDashboard.jsx` | 使用 `IframeDashboard dashboardKey="basic"`。 |
| Dashboard Page | `frontend/src/oss/views/dashboard/ApplicationDashboard.jsx` | 使用 `IframeDashboard dashboardKey="application"`。 |
| Dashboard Page | `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx` | 使用 `IframeDashboard dashboardKey="middleware"`，并额外展示文档提示。 |
| Dashboard Page | `frontend/src/pro/views/MysqlDashboard/index.jsx` | 试图直接传入 MySQL dashboard URL。 |
| Health Page | `frontend/src/pro/views/serviceHealthy/index.js` | 调 `/api/service/ryglight`，展示服务健康状态表。 |
| Shared Component | `frontend/src/core/components/Dashboard/IframeDashboard.jsx` | 通过 `dashboardKey` 从 `menuItems` 读取 `router.page.url`，拼接时间范围和主题后渲染 iframe。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/pro/api/serviceHealthy.js` | `getServiceHealthyApi` | `/api/service/ryglight` | 当前健康页唯一直接消费的模块 API。 |
| `frontend/src/core/api/permission.js` | `getUserPermissionApi` / `getRouterPermissionApi` | `/api/permission/config`, `/api/permission/router` | dashboard 页面不直接调 dashboard API，而是依赖权限配置中的 `router.page.url`。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router | `backend/pkg/router/router_api.go` | `/api/service/ryglight`、`/api/service/monitor/status` 都在 `/api/service` route group 下，受 `AuthMiddleware` 与 `LicMiddleware` 保护。 |
| Handler Root | `backend/pkg/api/serviceoverview/handler.go` | 组装 `serviceoverview.Service` 与 `dataService`。 |
| Handler | `backend/pkg/api/serviceoverview/func_getryglight.go` | 绑定 `GetRygLightRequest`，执行 `CheckGroupPermission`。 |
| Handler | `backend/pkg/api/serviceoverview/func_getmonitorstatus.go` | 绑定 `GetMonitorStatusRequest`，转时间范围后查询监控状态。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Health Score | `backend/pkg/services/serviceoverview/service_ryg_light.go` | 聚合 Prometheus 与 ClickHouse 数据，计算 RYG 健康评分。 |
| Monitor Status | `backend/pkg/services/serviceoverview/service_get_monitor_status.go` | 从 Prometheus 读取 `uptime-kuma` 风格监控项状态。 |
| Shared Interface | `backend/pkg/services/serviceoverview/service.go` | 暴露 `GetServicesRYGLightStatus` 与 `GetMonitorStatus`。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/service_req.go` | `GetRygLightRequest`, `GetMonitorStatusRequest`。 |
| Response | `backend/pkg/model/response/service_resp.go` | `ServiceRYGLightRes`, `ServiceRYGResult`, `RYGResult`, `RYGScoreDetail`, `GetMonitorStatusResponse`, `MonitorStatus`。 |
| Domain Model | `backend/pkg/services/serviceoverview/service_ryg_light.go` | `RYGLightStatus` 聚合 Prometheus 指标、告警状态和实例数。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Menu Config | `frontend/src/core/contexts/UserContext.js` | `IframeDashboard` 通过 `menuItems` 找到 dashboard URL。 |
| Time Range | `frontend/src/core/store/reducers/timeRangeReducer.js` | dashboard URL 会附带时间范围；健康页也使用共享时间范围。 |
| Permission/Menu Storage | `backend/pkg/repository/database/init_menu_item.go`, `init_feature_mapping.go`, `init_router.go` | dashboard / healthy 页面并非纯前端硬编码，而有数据库初始化映射。 |
| Router Page Storage | `backend/pkg/repository/database/init_router_page.go` | 初始化 Grafana / Jaeger 页面 URL。 |
| Prometheus Repo | `backend/pkg/repository/prometheus` | 健康评分与监控状态的主要指标来源。 |
| ClickHouse Repo | `backend/pkg/repository/clickhouse` | 健康评分中告警事件分组计数依赖它。 |
| Data Group Permission | `backend/pkg/services/data` | `/api/service/ryglight` 的数据组权限检查依赖它。 |
| Shared Dashboard Embedding | `frontend/src/oss/views/serviceInfo/component/infoUni/MetricsDashboard.jsx`, `frontend/src/pro/components/DataViewCom/MetricsDashboard.jsx` | 其他模块也复用 `IframeDashboard` 能力。 |

## Evidence Notes

1. dashboard 页面当前没有各自的业务 API；它们依赖 `menuItems.find(item.key === dashboardKey).router.page.url` 动态获得 iframe 目标地址。
2. `/api/service/monitor/status` 在当前主前端代码中没有观察到消费点。
3. `frontend/src/pro/views/MysqlDashboard/index.jsx` 与 `frontend/src/pro/components/DataViewCom/MetricsDashboard.jsx` 传给 `IframeDashboard` 的是 `src`，而组件参数名是 `srcProp`，存在明显接线偏差。
4. `frontend/src/pro/views/serviceHealthy/index.js` 调 `/api/service/ryglight` 时传的是 `dataGroupId`，而后端 `GetRygLightRequest` 绑定的是 `groupId`，需要单独核对。
