# Service Observability Open Questions

## Business semantics that cannot be fully inferred from code

| ID | Issue | Evidence | Why It Needs Attention |
|---|---|---|---|
| SQ-01 | `DelaySource` 的最终业务定义只能从文案推断，无法仅凭代码严格确认 | 前端只把值渲染为 `self`、`dependency`、`unknown`；后端通过阈值比较与 `DelaySource` 数值填充 | 文档可以说明表现形式，但不能杜撰更强的业务定义。 |
| SQ-02 | `EntryImpact` 中“入口”的业务口径需要 `To Be Confirmed` | 后端 `GetServiceEntryEndpointsInGroup` 是通过最短祖先路径和组内可见 service 推断入口 | 代码能看出算法，但不能完全确认产品语义命名是否等同“首层入口服务”。 |

## Legacy or dead-code suspicion

| ID | Issue | Evidence | Why It Needs Attention |
|---|---|---|---|
| SQ-03 | `getServiceTableApi` 似乎是遗留 wrapper | `frontend/src/core/api/service.js` 仍指向 `api/service/top3`，但当前后端 `router_api.go` 未注册 `top3` | 需要避免在文档中把它误写成现行接口。 |
| SQ-04 | `ThresholdConfigModal` 当前更像未完成 UI | 组件存在，但未见被 `ServiceTable` 或其他页面引用，也未调用 `getThreshold` / `setThreshold` | 需要避免写成“用户已可配置 threshold”。 |
| SQ-05 | `/service/overview` 可能是遗留 route 标识 | `backend/pkg/repository/database/init_feature_mapping.go` 把 `服务概览` 绑定到 `/service/info` 和 `/service/overview`，但当前前端未见 `/service/overview` | 需要确认是历史兼容、插页页签，还是未清理的旧路径。 |

## Missing frontend-backend alignment

| ID | Issue | Evidence | Why It Needs Attention |
|---|---|---|---|
| SQ-06 | `LogsInfo` 没有像 `TraceInfo` 那样显式传 `groupId` 和 `clusterIds` | `frontend/src/oss/views/serviceInfo/component/infoUni/LogsInfo.jsx` 只传 `startTime/endTime/service/endpoint/step`；而后端 `func_getlogmetrics.go` 依赖 `req.GroupID` 做权限检查 | 对非 root 用户可能造成权限失败或范围不一致。 |
| SQ-07 | `GetServiceEntryEndpoints` 的二次聚合忽略了 `GroupID` 与 `ClusterIDs` | `backend/pkg/api/service/func_getserviceentryendpoints.go` 中补齐告警摘要时调用 `GetServicesAlert(c, 0, nil, ...)` | 入口影响面板的第二阶段聚合可能与第一阶段范围不一致。 |
| SQ-08 | 前端 route guard 与后端权限模型并未真正对齐 | `AuthRouter.tsx` 中 `checkRoutePermission` 存在，但实际调用被注释 | 当前更像依赖后端数据接口做兜底，而不是前端先行拦截。 |

## Permission uncertainty

| ID | Issue | Evidence | Why It Needs Attention |
|---|---|---|---|
| SQ-09 | `/service/info` 的权限入口是 feature route 还是仅依赖数据接口结果，仍需 `To Be Confirmed` | 菜单只显式挂 `/service`；`/service/info` 通过 `init_router.go` 和 `init_feature_mapping.go` 注册为 feature route | 这会影响后续如何描述页面可见性与收藏/直链行为。 |
| SQ-10 | `CheckGroupPermission(0, ...)` 的业务含义需要谨慎对待 | 多个接口若前端未传 `groupId`，后端会落到 `groupID = 0` 分支 | 文档中不能简单把 `groupId` 缺省解释成“全部数据可见”。 |

## Data model uncertainty

| ID | Issue | Evidence | Why It Needs Attention |
|---|---|---|---|
| SQ-11 | threshold 写入逻辑看起来不完整 | `backend/pkg/services/serviceoverview/service_set_threshold.go` 只填 `Latency/Tps/ErrorRate/Log`，没有写 `ServiceName`、`EndPoint`，且仅在 `level == global` 时写 `Level` | 当前无法确认 service 级别或 endpoint 级别 threshold 是否真的能落库。 |
| SQ-12 | 模块没有独立“服务业务表”，而是多数据源读模型 | 代码主要依赖 `Prometheus`、`ClickHouse`、`database` 权限表和 `threshold` | 后续文档应避免使用传统 CRUD 模块的表设计叙述方式。 |
