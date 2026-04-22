# Dashboards And Health Monitoring Module Overview

## Module Goal
本模块为系统提供两类聚合视图能力：

- 通过 iframe 直接承载 Grafana / Jaeger 风格的大盘页面，覆盖全局资源、基础设施、应用指标、中间件以及 MySQL 监控场景。
- 通过 `/healthy-service` 页面提供服务健康状态总览，按 service 输出 RYG 风险灯、百分制健康分和评分详情。

它不是单一业务流程，而是“宏观大盘入口 + 健康评分页”的组合模块。

## Scope
当前代码证据支持以下范围：

- `/system-dashboard`
- `/basic-dashboard`
- `/application-dashboard`
- `/middleware-dashboard`
- `/mysql-dashboard`
- `/healthy-service`
- `/api/service/ryglight`
- `/api/service/monitor/status`
- dashboard URL 的数据库初始化映射

当前代码不应纳入本模块主体范围的能力：

- `service-observability` 中嵌入的 service 详情 dashboard 面板业务逻辑
- `trace/full` 页面主体逻辑
- 告警事件、告警规则和 AI 诊断页面

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/oss/routes.js` | 注册四个 OSS dashboard 页面。 |
| Route | `frontend/src/pro/routes.js` | 注册 `/mysql-dashboard` 与 `/healthy-service`。 |
| Page | `frontend/src/oss/views/dashboard/*.jsx` | 四个 OSS dashboard 页面。 |
| Page | `frontend/src/pro/views/MysqlDashboard/index.jsx` | MySQL dashboard 页面。 |
| Page | `frontend/src/pro/views/serviceHealthy/index.js` | 服务健康状态页。 |
| Shared Component | `frontend/src/core/components/Dashboard/IframeDashboard.jsx` | 所有 dashboard 页面共用的 iframe 组件。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | 健康相关接口挂在 `/api/service` 下。 |
| Handler Package | `backend/pkg/api/serviceoverview` | `GetRYGLight`、`GetMonitorStatus`。 |
| Service Package | `backend/pkg/services/serviceoverview` | 健康评分与监控状态的核心实现。 |
| DB Init Mapping | `backend/pkg/repository/database/init_router_page.go` 等 | dashboard 页面 URL 与菜单并非只在前端硬编码。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `dashboardKey` | `IframeDashboard` 用它从 `menuItems` 中找到目标 dashboard URL。 |
| `router.page.url` | 后端初始化并最终通过权限配置下发给前端的 dashboard iframe 地址。 |
| `RYGResult` | 服务健康评分结果，包含 `percentScore`、`score`、`status`、`scoreDetail`。 |
| `RYGStatus` | 健康状态枚举，取值 `red` / `yellow` / `green`。 |
| `RYGScoreDetail` | 每个检查项的分值和解释文案。 |
| `MonitorStatus` | 监控项名称与存活状态结果。 |

## Related Files

- `frontend/src/oss/views/dashboard/SystemDashboard.jsx`
- `frontend/src/oss/views/dashboard/BasicDashboard.jsx`
- `frontend/src/oss/views/dashboard/ApplicationDashboard.jsx`
- `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx`
- `frontend/src/pro/views/MysqlDashboard/index.jsx`
- `frontend/src/pro/views/serviceHealthy/index.js`
- `frontend/src/core/components/Dashboard/IframeDashboard.jsx`
- `frontend/src/pro/api/serviceHealthy.js`
- `backend/pkg/api/serviceoverview/func_getryglight.go`
- `backend/pkg/api/serviceoverview/func_getmonitorstatus.go`
- `backend/pkg/services/serviceoverview/service_ryg_light.go`
- `backend/pkg/services/serviceoverview/service_get_monitor_status.go`
- `backend/pkg/model/request/service_req.go`
- `backend/pkg/model/response/service_resp.go`
- `backend/pkg/repository/database/init_router_page.go`
- `backend/pkg/repository/database/init_menu_item.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `UserContext` + `userReducer.menuItems` | dashboard iframe URL 解析依赖权限配置下发的菜单结构。 |
| Frontend Shared | `timeRangeReducer` | dashboard URL 和健康页查询都依赖统一时间范围。 |
| Frontend Shared | `AuthRouter` | 页面入口被统一 guard 包装。 |
| Backend Shared | `data.Service.CheckGroupPermission` | `/api/service/ryglight` 会检查数据组权限。 |
| Repository | `prometheus.Repo` | dashboard 健康评分和 monitor status 的主要数据来源。 |
| Repository | `clickhouse.Repo` | 健康评分中告警事件计数依赖它。 |
| Repository | `database.Repo` | dashboard / menu / feature / router page 初始化依赖它。 |
| External Service | Grafana / Jaeger page URLs | dashboard 页面本质上是外部可视化页面嵌入。 |
| Cross Module | `service-observability` | service 详情里也复用 `IframeDashboard` 机制嵌入 dashboard。 |

## Risks And Unknowns

1. 多个 dashboard 页面本身没有直接业务 API，运行效果高度依赖 `menuItems` 和 `router.page.url` 是否正确下发。
2. `MysqlDashboard` 当前传的是 `src` 而不是 `srcProp`，与 `IframeDashboard` 参数名不一致，存在真实接线风险。
3. `/healthy-service` 前端请求参数名使用 `dataGroupId`，后端绑定的是 `groupId`，可能导致权限过滤失效。
4. `/api/service/monitor/status` 当前未观察到主前端消费点，其业务归属和启用状态需要确认。
