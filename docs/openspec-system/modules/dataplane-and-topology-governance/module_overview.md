# Dataplane And Topology Governance Module Overview

## Module Goal
本模块负责把“观测数据接入情况”转换成平台内可治理的配置对象。前端提供 `/dataplane` 页面管理数据平面与服务匹配规则，并在 Agent/Probe 入口复用这些状态；后端通过 `/api/dataplane/*` 提供数据平面 CRUD、服务归因、服务匹配规则和拓扑合并能力。

## Scope
当前代码证据支持以下范围：

- 数据平面类型展示、配置测试、创建、编辑、删除。
- 已配置数据源清单展示。
- 服务列表与实例列表展示，以及实例来源归因为 `rule` 或 `source`。
- 服务匹配规则的新增、试跑、更新、删除。
- APPInfo 标签 key/value 的辅助查询。
- 最小数据平面存在性检查，供 Agent 入口判断是否就绪。
- 已接入 dataplane capability 摘要查询，供 `/dataplane` 页面和 Agent 数据状态组件消费。
- 实时服务拓扑与 `custom_service_topology` 静态边的合并查询。

当前代码不应直接写入本模块主范围的内容：

- `service-observability` / `logs-exploration-and-parsing` / `trace-exploration-and-profiling` 中的观测查询页面。
- `conversational-diagnosis-and-reporting` 中的 Agent 会话本身；本模块只为其提供数据前置状态。
- `frontend/src/core/views/Config/index.tsx` 的 `/demo/service-config` mock 页面。
## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/pro/routes.js` | 主入口是 `/dataplane`。 |
| Page Root | `frontend/src/pro/views/agentDataplane/index.tsx` | 根页面用 query `tab` 切换 `dataplane` 和 `serviceMatcher`。 |
| Dataplane Tab | `frontend/src/pro/views/agentDataplane/DataplanePage.tsx` | 左侧数据平面类型，右侧已配置源。 |
| Service Matcher Tab | `frontend/src/pro/views/agentDataplane/ServiceMatcherPage.tsx` | 服务与实例映射表，支持展开规则。 |
| Shared Config Form | `frontend/src/pro/components/Dataplane/DataplaneConfigForm/index.tsx` | 被 `/dataplane` 和 probe 安装流程复用。 |
| Agent Status Entry | `frontend/src/pro/components/AGUIChat/components/DataStatus.tsx` | 会打开 `/#/dataplane?tab=...`。 |
| Installer Entry | `frontend/src/pro/views/probeInstaller/DataplaneStep/index.tsx` | 安装时选择并配置数据平面。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | `/api/dataplane` 整组挂 `LicMiddleware()`。 |
| Handler Package | `backend/pkg/api/dataplane` | 聚合数据平面、规则、拓扑和状态接口。 |
| Service Package | `backend/pkg/services/dataplane` | 实现数据平面 CRUD、规则匹配、实例归因和拓扑合并。 |
| Repository Package | `backend/pkg/repository/dataplane` | 面向观测存储的服务、实例、拓扑查询。 |
| Database DAO | `backend/pkg/repository/database/dao_dataplane.go` 等 | 负责治理侧关系表持久化。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `data_plane` | 平台记录的外部数据平面配置实体，包含 `typ`、`params`、`capability`、`name` 等字段。 |
| `data_plane_type` | 平台支持的数据平面类型模板，提供 `paramSpec`、`capabilitySpec`、描述与展示名。 |
| `service_name_rule` | 针对某个 service 的归因规则主表。 |
| `service_name_rule_condition` | 规则条件表，定义 label key、匹配类型和值。 |
| `custom_service_topology` | 静态拓扑边，用于补齐实时拓扑缺失的服务关系。 |
| `agentReady` | 最近窗口内是否存在最小 agent cluster 的观测信号。 |
| `dataPlaneExist` | 数据库中是否至少存在一条 `data_plane` 配置。 |
| `InstanceSource` | 实例归因来源，当前由 `rule_id` 或 `source` label 反推为 `rule` 或 `source`。 |

## Related Files

- `frontend/src/pro/routes.js`
- `frontend/src/pro/views/agentDataplane/index.tsx`
- `frontend/src/pro/views/agentDataplane/DataplanePage.tsx`
- `frontend/src/pro/views/agentDataplane/ServiceMatcherPage.tsx`
- `frontend/src/pro/views/agentDataplane/ServiceMatcher/index.tsx`
- `frontend/src/pro/views/agentDataplane/ServiceMatcher/ConfigModal/index.tsx`
- `frontend/src/pro/components/Dataplane/DataplaneList.tsx`
- `frontend/src/pro/components/Dataplane/DataplaneConfigForm/index.tsx`
- `frontend/src/pro/components/Dataplane/DataCoverageOverview.tsx`
- `frontend/src/pro/components/AGUIChat/components/DataStatus.tsx`
- `frontend/src/pro/views/probeInstaller/DataplaneStep/index.tsx`
- `frontend/src/pro/api/dataplane.ts`
- `backend/pkg/api/dataplane/handler.go`
- `backend/pkg/services/dataplane/service.go`
- `backend/pkg/services/dataplane/service_data_plane.go`
- `backend/pkg/services/dataplane/service_get_service.go`
- `backend/pkg/services/dataplane/service_set_servicename_rule.go`
- `backend/pkg/services/dataplane/service_list_servicename_rule.go`
- `backend/pkg/services/dataplane/service_check_servicename_rule.go`
- `backend/pkg/services/dataplane/service_get_service_topology.go`
- `backend/pkg/repository/database/dao_dataplane.go`
- `backend/pkg/repository/database/dao_service_name_rule.go`
- `backend/pkg/repository/database/dao_service_name_rule_condition.go`
- `backend/pkg/repository/database/dao_custom_service_topology.go`
- `backend/pkg/model/request/data_plane_req.go`
- `backend/pkg/model/response/dataplane_resp.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `timeRangeReducer` | `serviceMatcher` 和 Agent 数据状态都依赖时间范围。 |
| Frontend Shared | `AuthRouter` | `/dataplane` 路由经统一包装，但实际 route permission 检查当前被注释。 |
| Cross Module | `conversational-diagnosis-and-reporting` | Agent 页面依赖本模块的 `data/check` 和数据状态。 |
| Cross Module | `data-ingestion-and-oneagent` | 安装流程复用数据平面类型和配置表单。 |
| Backend Shared | ClickHouse / Prometheus repository | 服务、实例、拓扑和最小 agent cluster 判断来自观测后端。 |
| Permission Metadata | `service_check_router.go` + `init_router.go` | router 初始化已切到 `/dataplane`，但权限白名单仍保留 `/agent-dataplane`。 |

## Risks And Unknowns

1. `/api/dataplane/capabilities` 已实现，但它基于已配置 `data_plane.capability` 聚合摘要，不等于运行时可用性探测。
2. `customtopology` 前端 wrapper 已存在，但主页面当前没有稳定入口消费这些接口，存在半成品或预留分支。
3. `service_check_router.go` 的历史白名单仍写 `/agent-dataplane`，而实际前端主路由是 `/dataplane`。
4. `/demo/service-config` 页面与“服务治理”命名接近，但当前是被 `comingSoon` 覆盖的 mock 页面，容易误归类。
5. `DataplaneConfigForm` 会展示能力状态和接入指引，但 `/api/dataplane/capabilities` 当前只反映“已接入配置声明”，不是实时拓扑或服务采集健康。
