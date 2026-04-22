# Dataplane And Topology Governance Source Map

## Scope
本文档固定 `dataplane-and-topology-governance` 模块在当前仓库中的真实代码证据。该模块覆盖数据平面接入配置、服务匹配规则治理、数据覆盖状态展示，以及后端的服务拓扑与自定义拓扑合并逻辑。本文档不把未落到主页面的 mock 页面或外部系统说明误写成已实现主流程。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/pro/views/agentDataplane` | `/dataplane` 主页面，包含 `dataplane` 和 `serviceMatcher` 两个 tab。 |
| Frontend | `frontend/src/pro/components/Dataplane` | 数据平面类型列表、配置表单、数据覆盖概览和配置引导。 |
| Frontend | `frontend/src/pro/api/dataplane.ts` | 数据平面、服务匹配规则、拓扑、状态检查相关 API wrapper。 |
| Frontend | `frontend/src/pro/components/AGUIChat/components/DataStatus.tsx` | Agent 入口中的数据状态提示，会跳转到 `/dataplane`。 |
| Frontend | `frontend/src/pro/views/probeInstaller/DataplaneStep` | OneAgent/Probe 安装流程中的数据平面类型选择和配置复用入口。 |
| Backend | `backend/pkg/api/dataplane` | `/api/dataplane/*` route handler 集合。 |
| Backend | `backend/pkg/services/dataplane` | 数据平面 CRUD、服务匹配规则、实例归因、拓扑和数据平面状态判断。 |
| Backend | `backend/pkg/repository/dataplane` | 面向 ClickHouse/Prometheus 的服务、实例、拓扑和数据源查询。 |
| Backend | `backend/pkg/repository/database/dao_dataplane.go` | `data_plane`、`data_plane_type`、`cluster2data_plane` 相关持久化。 |
| Backend | `backend/pkg/repository/database/dao_service_name_rule.go` | `service_name_rule` 主表。 |
| Backend | `backend/pkg/repository/database/dao_service_name_rule_condition.go` | `service_name_rule_condition` 条件表及匹配逻辑。 |
| Backend | `backend/pkg/repository/database/dao_custom_service_topology.go` | `custom_service_topology` 静态拓扑边存储。 |
| Backend | `backend/pkg/model/request/data_plane_req.go` | 模块请求结构定义。 |
| Backend | `backend/pkg/model/response/dataplane_resp.go` | 模块响应结构定义。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/pro/routes.js` | `/dataplane`。 |
| Frontend Shared | `frontend/src/routes.js` | `/demo/service-config` 仍指向 `Config/index.tsx`，但属于带 `comingSoon` 覆盖的 mock 页，不应视为主流程入口。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/dataplane/services`, `/redcharts`, `/redcharts/list`, `/endpoints`, `/instances`, `/servicename`, `/topology`, `/checkDataSource`, `/customtopology/*`, `/servicename/*`, `/type/list`, `/list`, `/create`, `/update`, `/delete`, `/data/check`, `/appinfo/tags`, `/appinfo/tags/values`, `/datasources`, `/servicesWithInstances`。 |
| Permission Metadata | `backend/pkg/repository/database/init_router.go` | 当前已预置 `/dataplane`；历史 allowlist 仍在 permission service 中保留 `/agent-dataplane`。 |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/pro/views/agentDataplane/index.tsx` | `/dataplane` 页面根入口，使用 URL query `tab` 在 `dataplane` 和 `serviceMatcher` 间切换。 |
| Dataplane Tab | `frontend/src/pro/views/agentDataplane/DataplanePage.tsx` | 左侧可配置数据平面类型列表，右侧已配置数据源列表，抽屉复用同一表单。 |
| Config Drawer | `frontend/src/pro/views/agentDataplane/ConfigDrawer/index.tsx` | 复用 `DataplaneConfigForm`，支持新增和编辑。 |
| Configured Sources | `frontend/src/pro/views/agentDataplane/components/ConfiguredSources.tsx` | 调 `/api/dataplane/list` 与 `/api/dataplane/delete` 展示已配置数据源。 |
| Service Matcher Tab | `frontend/src/pro/views/agentDataplane/ServiceMatcherPage.tsx` | 调 `/api/dataplane/servicesWithInstances` 展示服务与实例关系，可展开规则列表。 |
| Rule List | `frontend/src/pro/views/agentDataplane/ServiceMatcher/index.tsx` | 读取 `/api/dataplane/servicename/listRule`，支持编辑和删除规则。 |
| Rule Modal | `frontend/src/pro/views/agentDataplane/ServiceMatcher/ConfigModal/index.tsx` | 使用 `/api/dataplane/servicename/checkRule` 做试跑，再用 `/api/dataplane/servicename/upsertRule` 保存。 |
| Dataplane Type List | `frontend/src/pro/components/Dataplane/DataplaneList.tsx` | 读取 `/api/dataplane/type/list`，并额外调用 `/api/dataplane/capabilities` 计算能力状态。 |
| Data Coverage | `frontend/src/pro/components/Dataplane/DataCoverageOverview.tsx` | 通过 `/api/dataplane/capabilities` 展示能力覆盖详情。 |
| Agent Badge | `frontend/src/pro/components/AGUIChat/components/DataStatus.tsx` | 通过 `/api/dataplane/datasources` 和 `/api/dataplane/capabilities` 呈现 Agent 前置数据状态。 |
| Probe Installer Step | `frontend/src/pro/views/probeInstaller/DataplaneStep/index.tsx` | 复用 `DataplaneConfigForm`，说明本模块同时被接入安装流程消费。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/pro/api/dataplane.ts` | `getDataplaneTypeListApi` | `/api/dataplane/type/list` | 数据平面类型列表。 |
| `frontend/src/pro/api/dataplane.ts` | `getDataplaneListApi` | `/api/dataplane/list` | 已配置数据平面列表。 |
| `frontend/src/pro/api/dataplane.ts` | `createDataplaneApi` | `/api/dataplane/create` | 新增数据平面。 |
| `frontend/src/pro/api/dataplane.ts` | `updateDataplaneApi` | `/api/dataplane/update` | 更新数据平面。 |
| `frontend/src/pro/api/dataplane.ts` | `deleteDataplaneApi` | `/api/dataplane/delete` | 删除数据平面。 |
| `frontend/src/pro/api/dataplane.ts` | `checkDataplaneDatasourceApi` | `/api/dataplane/checkDataSource` | 配置表单“测试连接”调用。 |
| `frontend/src/pro/api/dataplane.ts` | `checkDataplaneDataApi` | `/api/dataplane/data/check` | Agent 会话入口前置检查。 |
| `frontend/src/pro/api/dataplane.ts` | `getDataplaneDatasourcesApi` | `/api/dataplane/datasources` | Agent 数据状态 badge 使用。 |
| `frontend/src/pro/api/dataplane.ts` | `getCapabilitiesApi` | `/api/dataplane/capabilities` | 前端多处调用；后端当前已注册并返回 capability 摘要。 |
| `frontend/src/pro/api/dataplane.ts` | `getServiceWithIstanceApi` | `/api/dataplane/servicesWithInstances` | `serviceMatcher` 主表使用。 |
| `frontend/src/pro/api/dataplane.ts` | `getServicesApi` | `/api/dataplane/services` | 规则配置时读取服务清单。 |
| `frontend/src/pro/api/dataplane.ts` | `listServiceNameRuleApi` | `/api/dataplane/servicename/listRule` | 某服务下规则列表。 |
| `frontend/src/pro/api/dataplane.ts` | `checkServiceNameRuleApi` | `/api/dataplane/servicename/checkRule` | 规则试跑。 |
| `frontend/src/pro/api/dataplane.ts` | `upsertServiceNameRuleApi` / `saveServiceRuleApi` | `/api/dataplane/servicename/upsertRule` | 保存规则，存在重复 wrapper。 |
| `frontend/src/pro/api/dataplane.ts` | `deleteServiceNameRuleApi` / `deleteServiceRuleApi` | `/api/dataplane/servicename/deleteRule` | 删除规则，存在重复 wrapper。 |
| `frontend/src/pro/api/dataplane.ts` | `getAppinfoTagsApi` / `getServiceNameRuleApi` | `/api/dataplane/appinfo/tags` | 读取标签 key，存在重复 wrapper。 |
| `frontend/src/pro/api/dataplane.ts` | `getAppinfoTagValuesApi` / `getServiceValuesRuleApi` | `/api/dataplane/appinfo/tags/values` | 读取标签值，存在重复 wrapper。 |
| `frontend/src/pro/api/dataplane.ts` | `getTopologyApi` | `/api/dataplane/topology` | wrapper 存在，当前主页面未见实际消费。 |
| `frontend/src/pro/api/dataplane.ts` | `createCustomTopologyApi` | `/api/dataplane/customtopology/create` | wrapper 存在，当前主页面未见实际消费。 |
| `frontend/src/pro/api/dataplane.ts` | `getCustomTopologyListApi` | `/api/dataplane/customtopology/list` | wrapper 存在，当前主页面未见实际消费。 |
| `frontend/src/pro/api/dataplane.ts` | `deleteCustomTopologyApi` | `/api/dataplane/customtopology/delete` | wrapper 存在，当前主页面未见实际消费。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | 注册整个 `/api/dataplane` 路由组，并挂 `LicMiddleware()`。 |
| Handler Definition | `backend/pkg/api/dataplane/handler.go` | 汇总模块 handler interface。 |
| DataPlane CRUD | `func_listdataplanetype.go`, `func_listdataplane.go`, `func_createdataplane.go`, `func_updatedataplane.go`, `func_deletedataplane.go`, `func_checkdataplane.go` | 数据平面类型、实例和前置状态管理。 |
| Service Query | `func_queryservices.go`, `func_getservicewithinstances.go`, `func_queryserviceinstances.go`, `func_queryserviceendpoints.go` | 服务、实例、endpoint 读取。 |
| Rule Governance | `func_listservicenamerule.go`, `func_setservicenamerule.go`, `func_deleteservicenamerule.go`, `func_checkservicenamerule.go`, `func_cleanruleservices.go`, `func_cleandatasourceservices.go` | 服务匹配规则管理。 |
| Tag Query | `func_queryappinfotags.go`, `func_queryappinfotagvalues.go`, `func_queryservicename.go` | APPInfo 标签与服务名查询。 |
| Topology | `func_querytopology.go`, `func_createcustomtopology.go`, `func_listcustomtopology.go`, `func_deletecustomtopology.go` | 拓扑查询与静态拓扑边管理。 |
| Data Status | `func_checkdatasource.go`, `func_getdatasources.go`, `func_getdataplanecapabilities.go` | 测试数据源可连通性、汇总数据源/服务数量，以及 capability 摘要。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Service Interface | `backend/pkg/services/dataplane/service.go` | 模块 service 聚合入口，依赖 ClickHouse、Prometheus、DB、`dataplaneRepo`。 |
| DataPlane CRUD | `service_data_plane.go` | 类型排序、`apo` 数据平面隐藏、参数校验、创建/更新/删除和最小数据平面检测。 |
| Datasource Health | `service_check_datasource.go`, `service_get_datasources.go` | 测试连接与汇总数据源状态。 |
| Capability Summary | `service_get_integrated_capabilities.go` | 基于 DB 中已配置 `data_plane` 聚合 telemetry capability 摘要。 |
| Service Discovery | `service_get_service.go`, `service_get_service_instances.go`, `service_get_service_endpoints.go` | 服务和实例聚合，实例来源归因为 `rule` 或 `source`。 |
| Service Matcher | `service_set_servicename_rule.go`, `service_list_servicename_rule.go`, `service_check_servicename_rule.go`, `service_delete_servicename_rule.go`, `service_clean_relate_service.go` | 规则保存、试跑、查询、删除及清理关联。 |
| APPInfo Labels | `service_list_app_info_labels_keys.go` | 从 APPInfo 提取标签 key。 |
| Topology | `service_get_service_topology.go`, `service_create_custom_topology.go`, `service_list_custom_topology.go`, `service_delete_custom_topology.go` | 动态拓扑与静态拓扑合并管理。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Request | `backend/pkg/model/request/data_plane_req.go` | 定义 `CreateDataPlaneRequest`、`SetServiceNameRuleRequest`、`CreateCustomTopologyRequest`、`GetServiceWithInstancesRequest` 等。 |
| Response | `backend/pkg/model/response/dataplane_resp.go` | 定义 `ListDataPlaneTypeResponse`、`ListServiceNameRuleResponse`、`CheckMinimalDataPlaneResponse` 等。 |
| Data Plane Storage | `backend/pkg/model/integration/data_plane*.go` | `service_data_plane.go` 中直接使用 `integration.DataPlane`、`DataPlaneType`、`DataPlaneWithClusterIDs`。 |
| `data_plane` DAO | `backend/pkg/repository/database/dao_dataplane.go` | 管理数据平面、类型和 cluster 关系。 |
| `service_name_rule` DAO | `backend/pkg/repository/database/dao_service_name_rule.go` | 规则主表。 |
| `service_name_rule_condition` DAO | `backend/pkg/repository/database/dao_service_name_rule_condition.go` | 条件表，内置 `equals`、`startsWith`、`endsWith`、`contains`、`has` 五类匹配。 |
| `custom_service_topology` DAO | `backend/pkg/repository/database/dao_custom_service_topology.go` | 自定义静态拓扑边及过期时间。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Time Range | `frontend/src/core/store/reducers/timeRangeReducer` | `/dataplane` 的 `serviceMatcher` tab 和 Agent badge 都依赖时间范围。 |
| Route/Menu Permission | `frontend/src/core/components/AuthRouter.tsx`, `backend/pkg/services/permission/service_check_router.go`, `backend/pkg/repository/database/init_router.go` | 数据库初始化已切到 `/dataplane`，但权限白名单仍保留 `/agent-dataplane` 历史字符串。 |
| Agent Entry | `frontend/src/pro/components/AGUIChat/components/DataStatus.tsx` | Agent 页面通过本模块状态引导用户回到 `/dataplane` 补齐接入。 |
| Probe Installer | `frontend/src/pro/views/probeInstaller/DataplaneStep` | 接入安装流程复用数据平面类型和配置表单。 |
| ClickHouse / Prometheus | `backend/pkg/repository/dataplane/*` | 服务、实例、拓扑、节点状态查询的主要数据源。 |

## Evidence Notes

1. `/dataplane` 页面当前只有两个稳定 tab：`dataplane` 和 `serviceMatcher`，没有前端证据表明自定义拓扑已接入主界面。
2. `DataplaneList`、`DataCoverageOverview`、`DataStatus` 都会请求 `/api/dataplane/capabilities`；当前后端已实现该接口，但它只汇总“已接入 capability”，不直接代表运行时健康。
3. `service_get_service_topology.go` 明确会把实时拓扑和 `custom_service_topology` 静态边合并后统一返回。
4. `CheckMinimalDataPlaneExist` 既检查 Prometheus 侧最近 7 天是否存在 agent cluster，也检查 DB 中是否已有 `data_plane` 记录，结果同时返回 `agentReady` 和 `dataPlaneExist`。
5. `frontend/src/core/views/Config/index.tsx` 虽然和 service config 命名相关，但当前整个页面被 `GlassCard` 的 `comingSoon` 覆盖，不应写成已交付的数据治理能力。
