# Dashboards And Health Monitoring Dependencies

## Upstream Dependencies

| Dependency | Evidence | Why It Matters |
|---|---|---|
| `identity-access-and-organization` | `UserContext`, `menuItems`, `AuthRouter`, `/api/permission/config` | dashboard 页面要靠权限配置下发 `router.page.url`。 |
| `data-scope-governance` | `dataGroupReducer`, `CheckGroupPermission`, `GetPQLFilterByGroupID` | `/healthy-service` 的健康度依赖数据组权限。 |
| Shared Time Range Infrastructure | `timeRangeReducer` | dashboard URL 拼接与健康页查询都依赖统一时间范围。 |
| Grafana / Jaeger Embedding | `init_router_page.go`, `IframeDashboard.jsx` | dashboard 页面本质上是外部页面嵌入。 |
| Prometheus | `service_ryg_light.go`, `service_get_monitor_status.go` | 健康度与监控状态的主要数据来源。 |
| ClickHouse | `service_ryg_light.go` | 健康度中的告警事件统计依赖它。 |

## Downstream Dependencies

| Consumer | Evidence | Dependency Form |
|---|---|---|
| `service-observability` | `frontend/src/oss/views/serviceInfo/component/infoUni/MetricsDashboard.jsx` | 复用 `IframeDashboard` 机制嵌入 dashboard。 |
| `conversational-diagnosis-and-reporting` / `pro DataView` 相关页面 | `frontend/src/pro/components/DataViewCom/MetricsDashboard.jsx` | 同样复用 dashboard iframe 能力。 |

## Shared Components

| Shared Component | File | Usage |
|---|---|---|
| `IframeDashboard` | `frontend/src/core/components/Dashboard/IframeDashboard.jsx` | 所有 dashboard 页面共用。 |
| `BasicTable` | `src/core/components/Table/basicTable` | 健康状态页表格。 |
| `LoadingSpinner` | `src/core/components/Spinner` | 健康页加载态。 |
| `BorderTag` | `src/core/components/Tag/BorderTag` | 展示评分详情级别标签。 |

## Shared Utilities

| Utility | File | Usage |
|---|---|---|
| `request` | `src/core/utils/request` | 健康页 API 请求。 |
| `selectProcessedTimeRange` | `timeRangeReducer.js` | 健康页与 dashboard 时间参数来源。 |
| `timeRangeList` | `timeRangeReducer.js` | `IframeDashboard` 在 `rangeType` 模式下用它构造 `from` / `to`。 |
| `userReducer.menuItems` | Redux state | dashboard URL 配置读取入口。 |

## External Services

| External Service | Evidence | Purpose |
|---|---|---|
| Grafana page URLs | `init_router_page.go` | `/system-dashboard`、`/basic-dashboard`、`/application-dashboard`、`/middleware-dashboard`、`/mysql-dashboard` 的实际内容来源。 |
| Jaeger page URL | `init_router_page.go` | `trace/full` 也采用相同的 router page 配置模式，表明 dashboard 嵌入能力是通用机制。 |
| Prometheus | `QueryMetricsWithPQLFilter`, `QueryAggMetricsWithFilter`, `GetInstanceList` | 提供健康度与监控状态所需指标。 |
| ClickHouse | `GetAlertEventCountGroupByInstance` | 提供健康度中的告警事件统计。 |

## Cross-Module Linkage Notes

1. dashboard 页面文档必须强调其“配置驱动”特征，否则容易误以为它们各自拥有独立业务 API。
2. `IframeDashboard` 既服务本模块，也被其他模块复用，因此在系统架构中更接近共享基础设施，但本模块是它的主要业务承载面。
3. `/healthy-service` 虽然挂在 pro 路由下，但其后端接口位于 `/api/service` 的 `serviceoverview` 包中，文档应显式连接这两端。
