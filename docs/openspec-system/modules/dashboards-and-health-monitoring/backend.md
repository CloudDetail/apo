# Dashboards And Health Monitoring Backend

## API Summary

| API | Handler | Service | Notes |
|---|---|---|---|
| `GET /api/service/ryglight` | `GetRYGLight` | `GetServicesRYGLightStatus` | 服务健康状态评分与 RYG 风险灯。 |
| `GET /api/service/monitor/status` | `GetMonitorStatus` | `GetMonitorStatus` | 监控项存活状态查询。 |

补充说明：

- dashboard 页面本身没有对应的运行时业务 API。
- dashboard URL 主要通过数据库初始化的 `router.page.url` 配置，再经权限配置接口下发到前端。

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 在 `/api/service` route group 下注册 `/ryglight` 与 `/monitor/status`；该 group 已套用 `AuthMiddleware` 与 `LicMiddleware`。 |
| `backend/pkg/api/serviceoverview/handler.go` | 组装 `serviceoverview.Service` 和 `dataService`。 |
| `backend/pkg/api/serviceoverview/func_getryglight.go` | 绑定 `GetRygLightRequest`，执行 `CheckGroupPermission` 后调用 service。 |
| `backend/pkg/api/serviceoverview/func_getmonitorstatus.go` | 绑定 `GetMonitorStatusRequest`，把微秒时间转为 `time.Time` 后查询。 |

## Services

### Health Score

| File | Behavior |
|---|---|
| `backend/pkg/services/serviceoverview/service_ryg_light.go` | 按 service 聚合 Prometheus 指标、告警计数与实例数量，产出 `RYGResult`。 |

关键实现点：

1. 先调用 `common.GetPQLFilterByGroupID(ctx, dbRepo, "apm", req.GroupID)` 得到数据组过滤条件。
2. 再叠加 serviceName / endpointName / namespace 的页面过滤条件。
3. 依次查询：
   - `avgLatency`
   - `avgLatencyDoD`
   - `avgErrorRateDoD`
   - `avgLogErrorCountDoD`
4. 调 ClickHouse `GetAlertEventCountGroupByInstance` 获取 firing 告警计数。
5. 调 Prometheus `GetInstanceList` 获取实例列表。
6. 通过 `ExposeRYGLightStatus` 生成评分详情、总分、百分比分和 `red` / `yellow` / `green` 状态。

### Monitor Status

| File | Behavior |
|---|---|
| `backend/pkg/services/serviceoverview/service_get_monitor_status.go` | 调 `promRepo.QueryAggMetricsWithFilter(..., prometheus.PQLMonitorStatus, ...)` 读取监控项状态，并把 `1.0` 映射为 `IsAlive=true`。 |

### Dashboard Metadata Initialization

| File | Behavior |
|---|---|
| `backend/pkg/repository/database/init_router.go` | 初始化 dashboard / healthy 路由定义。 |
| `backend/pkg/repository/database/init_menu_item.go` | 初始化 dashboard / healthy 菜单项。 |
| `backend/pkg/repository/database/init_feature_mapping.go` | 初始化 feature 到菜单映射。 |
| `backend/pkg/repository/database/init_router_page.go` | 初始化每个 dashboard route 对应的 Grafana / Jaeger URL。 |

## Models And Schemas

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/service_req.go` | `GetRygLightRequest`, `GetMonitorStatusRequest`。 |
| Response | `backend/pkg/model/response/service_resp.go` | `ServiceRYGLightRes`, `ServiceRYGResult`, `RYGResult`, `RYGScoreDetail`, `GetMonitorStatusResponse`, `MonitorStatus`。 |
| Domain Aggregate | `backend/pkg/services/serviceoverview/service_ryg_light.go` | `RYGLightStatus` 聚合 Prometheus 指标、实例和告警统计。 |

## Database Mapping

| Storage | Entity | Mapping |
|---|---|---|
| Relational DB | Router definitions | `init_router.go` 中初始化 `/system-dashboard`、`/basic-dashboard`、`/application-dashboard`、`/middleware-dashboard`、`/mysql-dashboard`、`/healthy-service`。 |
| Relational DB | Menu items | `init_menu_item.go` 中初始化 `system`、`basic`、`application`、`middleware`、`mysql`、`healthy`。 |
| Relational DB | Router pages | `init_router_page.go` 中为 dashboard route 绑定 Grafana / Jaeger URL。 |
| Prometheus | Health metrics | `service_ryg_light.go` 和 `service_get_monitor_status.go` 的主要指标来源。 |
| ClickHouse | Alert event counts | 健康度里按实例聚合告警事件数。 |

## Validation And Rules

1. `/api/service/ryglight` 与 `/api/service/monitor/status` 都处在 `/api/service` route group 下，因此统一受 `AuthMiddleware` 与 `LicMiddleware` 保护。
2. `GetRYGLight` 在 handler 层显式调用 `CheckGroupPermission`。
3. `GetMonitorStatus` 当前只校验时间范围参数，没有看到 `groupId` 相关数据范围过滤。
4. `GetServicesRYGLightStatus` 会把 serviceName / endpointName / namespace 的页面筛选条件和数据组 PQL 过滤条件合并。
5. `ScoreFromDoD` 会把同比增长映射成 0 到 3 的离散分值。
6. `AlertScore` 会根据高优先级告警、warning 告警数量计算 0 到 3 的分值。
7. 副本数小于 2 时，`replica` 检查项给 0 分；否则给 3 分。
8. 总分会换算为百分比 `percentScore = score * 100 / MAX_RYG_SCORE`，再映射成 `GREEN`、`YELLOW`、`RED`。

## Error Handling

1. 参数绑定失败统一通过 `AbortWithError(..., code.ParamBindError, err)` 返回。
2. `GetRYGLight` 权限失败时返回空 `serviceList` 的响应体。
3. `GetMonitorStatus` service 在 Prometheus 查询失败时返回空结果和 `nil` error，这意味着 handler 侧可能返回一个空 `monitorList` 而不是显式失败。
4. 健康度逻辑中，多项 Prometheus 查询如果单项失败，会被跳过，不会直接终止整个评分流程。

## Notes

1. dashboard 页面运行所需 URL 不是仅靠前端常量，而是由数据库初始化与权限配置链路共同提供。
2. `GetMonitorStatus` 虽然实现存在，但当前主前端没有观察到消费点。
3. `service_ryg_light.go` 同时使用 Prometheus、ClickHouse 和数据组 PQL 过滤，是该模块最核心的后端聚合点。
