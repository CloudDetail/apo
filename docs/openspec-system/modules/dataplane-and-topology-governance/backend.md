# Dataplane And Topology Governance Backend

## API Summary

| Category | Method | API Path | Handler | Notes |
|---|---|---|---|---|
| DataPlane | GET | `/api/dataplane/type/list` | `ListDataPlaneType` | 返回支持的数据平面类型模板。 |
| DataPlane | GET | `/api/dataplane/list` | `ListDataPlane` | 返回已配置数据平面。 |
| DataPlane | POST | `/api/dataplane/create` | `CreateDataPlane` | 创建数据平面。 |
| DataPlane | POST | `/api/dataplane/update` | `UpdateDataPlane` | 更新数据平面。 |
| DataPlane | POST | `/api/dataplane/delete` | `DeleteDataPlane` | 逻辑删除数据平面。 |
| DataPlane | POST | `/api/dataplane/checkDataSource` | `CheckDataSource` | 测试数据源配置可达性。 |
| DataPlane | GET | `/api/dataplane/data/check` | `CheckDataPlane` | 返回 `agentReady`、`dataPlaneExist`、`clusters`、`dataSource`。 |
| Data Status | GET | `/api/dataplane/datasources` | `GetDatasources` | 汇总数据源数量和可监控服务数。 |
| Data Status | GET | `/api/dataplane/capabilities` | `GetIntegratedCapabilities` | 汇总已配置数据平面的 telemetry capability 与来源。 |
| Services | GET | `/api/dataplane/services` | `QueryServices` | 返回服务清单。 |
| Services | GET | `/api/dataplane/servicesWithInstances` | `GetServiceWithInstances` | 返回服务及实例来源。 |
| Services | GET | `/api/dataplane/endpoints` | `QueryServiceEndpoints` | 查询 service endpoint。 |
| Services | GET | `/api/dataplane/instances` | `QueryServiceInstances` | 查询某 service 实例。 |
| Services | POST | `/api/dataplane/servicename` | `QueryServiceName` | 根据标签反推服务名。 |
| Service Rule | GET | `/api/dataplane/servicename/listRule` | `ListServiceNameRule` | 列出服务匹配规则。 |
| Service Rule | POST | `/api/dataplane/servicename/upsertRule` | `SetServiceNameRule` | 新增或更新规则。 |
| Service Rule | POST | `/api/dataplane/servicename/deleteRule` | `DeleteServiceNameRule` | 删除规则。 |
| Service Rule | POST | `/api/dataplane/servicename/checkRule` | `CheckServiceNameRule` | 试跑规则，返回匹配实例。 |
| Service Rule | POST | `/api/dataplane/servicename/cleanRuleServices` | `CleanRuleServices` | 清理规则关联服务。 |
| Service Rule | POST | `/api/dataplane/servicename/cleanDataSourceServices` | `CleanDataSourceServices` | 清理数据源关联服务。 |
| APPInfo | GET | `/api/dataplane/appinfo/tags` | `QueryAPPInfoTags` | 返回 APPInfo 标签 key。 |
| APPInfo | GET | `/api/dataplane/appinfo/tags/values` | `QueryAPPInfoTagValues` | 返回指定标签的值。 |
| Topology | GET | `/api/dataplane/topology` | `QueryTopology` | 合并实时拓扑和静态拓扑边。 |
| Topology | POST | `/api/dataplane/customtopology/create` | `CreateCustomTopology` | 创建静态拓扑边。 |
| Topology | GET | `/api/dataplane/customtopology/list` | `ListCustomTopology` | 读取静态拓扑边列表。 |
| Topology | POST | `/api/dataplane/customtopology/delete` | `DeleteCustomTopology` | 删除静态拓扑边。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 注册 `/api/dataplane` 全部接口。 |
| `backend/pkg/api/dataplane/handler.go` | 模块 handler interface 与构造入口。 |
| `func_listdataplanetype.go`, `func_listdataplane.go` | 数据平面类型/实例读取。 |
| `func_createdataplane.go`, `func_updatedataplane.go`, `func_deletedataplane.go` | 数据平面 CRUD。 |
| `func_checkdatasource.go`, `func_checkdataplane.go`, `func_getdatasources.go`, `func_getdataplanecapabilities.go` | 配置测试、前置状态和 capability 摘要接口。 |
| `func_queryservices.go`, `func_getservicewithinstances.go`, `func_queryserviceinstances.go`, `func_queryserviceendpoints.go` | 服务、实例、endpoint 查询。 |
| `func_setservicenamerule.go`, `func_listservicenamerule.go`, `func_deleteservicenamerule.go`, `func_checkservicenamerule.go` | 服务匹配规则 handler。 |
| `func_queryappinfotags.go`, `func_queryappinfotagvalues.go`, `func_queryservicename.go` | APPInfo 标签和服务名辅助接口。 |
| `func_querytopology.go`, `func_createcustomtopology.go`, `func_listcustomtopology.go`, `func_deletecustomtopology.go` | 拓扑相关 handler。 |

## Services

| Service | File | Behavior |
|---|---|---|
| `ListDataPlaneType` | `service_data_plane.go` | 读取 `data_plane_type`，按 `specialOrder` 排序，并在缺省环境下隐藏 `apo`。 |
| `CreateDataPlane` | `service_data_plane.go` | 校验 type 存在、参数合法；若 `capability` 为空则使用完整能力；若带 `ClusterID` 则同时写 cluster 关联。 |
| `UpdateDataPlane` | `service_data_plane.go` | 校验目标存在、type 合法后更新。 |
| `DeleteDataPlane` | `service_data_plane.go` | 逻辑删除 `data_plane` 并清理 `cluster2data_plane`。 |
| `CheckMinimalDataPlaneExist` | `service_data_plane.go` | 同时读取 Prometheus 最近 7 天 agent cluster 与 DB 中数据平面记录。 |
| `GetDatasources` | `service_get_datasources.go` | 统计 `data_plane` 类型数量，并进一步查询这些 source 对应的服务总数。 |
| `GetIntegratedCapabilities` | `service_get_integrated_capabilities.go` | 枚举所有 telemetry type，并从已配置 `data_plane` 聚合 enabled 状态和 source 列表。 |
| `GetServices` | `service_get_service.go` | 从 ClickHouse 查询服务后按名称排序。 |
| `GetServiceInstances` | `service_get_service_instances.go`, `service_instances_lookup.go` | 对 `/api/dataplane/instances` 采用 ClickHouse 优先、显式查询预算、APO ownership gate 控制的 Prometheus fallback，以及短 TTL 成功结果缓存。 |
| `GetServiceWithInstances` | `service_get_service.go` | 过滤出来源于已配置数据源或规则归因的服务，并对实例补充 `InstanceSource`。 |
| `SetServiceNameRule` | `service_set_servicename_rule.go` | 新建或更新 `service_name_rule` 与其条件。 |
| `ListServiceNameRule` | `service_list_servicename_rule.go` | 读取规则与条件，并通过 `GetResolvedApps` 计算每条规则当前匹配实例。 |
| `CheckServiceNameRule` | `service_check_servicename_rule.go` | 使用 `GetToResolveApps` 对未保存规则做即时试跑。 |
| `GetServiceTopology` | `service_get_service_topology.go` | 把实时拓扑与未过期 `custom_service_topology` 合并成统一图。 |
| `CreateCustomTopology` | `service_create_custom_topology.go` | 创建静态拓扑边。 |
| `ListCustomTopology` | `service_list_custom_topology.go` | 列出静态拓扑边。 |
| `DeleteCustomTopology` | `service_delete_custom_topology.go` | 删除静态拓扑边。 |

## Models And Schemas

| Model | File | Mapping |
|---|---|---|
| `integration.DataPlane` | `backend/pkg/model/integration` | 数据平面主实体。 |
| `integration.DataPlaneType` | `backend/pkg/model/integration` | 支持的数据平面类型模板。 |
| `QueryServicesRequest` 等 | `backend/pkg/model/request/data_plane_req.go` | 定义服务、拓扑、规则和数据平面请求结构。 |
| `ListDataPlaneResponse`、`GetIntegratedCapabilitiesResponse` 等 | `backend/pkg/model/response/dataplane_resp.go` | 定义模块响应体。 |
| `ServiceNameRule` | `backend/pkg/repository/database/dao_service_name_rule.go` | `TableName() == "service_name_rule"`。 |
| `ServiceNameRuleCondition` | `backend/pkg/repository/database/dao_service_name_rule_condition.go` | `TableName() == "service_name_rule_condition"`。 |
| `CustomServiceTopology` | `backend/pkg/repository/database/dao_custom_service_topology.go` | `TableName() == "custom_service_topology"`。 |

## Database Mapping

| Storage | Evidence | Notes |
|---|---|---|
| `data_plane` | `dao_dataplane.go` | 保存已接入的数据平面配置，删除时通过 `is_deleted` 标记。 |
| `data_plane_type` | `dao_dataplane.go` | 保存可支持的数据平面类型及配置模板。 |
| `cluster2data_plane` | `dao_dataplane.go` | cluster 与数据平面的关联。 |
| `service_name_rule` | `dao_service_name_rule.go` | 保存规则头信息：service 与 cluster。 |
| `service_name_rule_condition` | `dao_service_name_rule_condition.go` | 保存每条规则的条件列表。 |
| `custom_service_topology` | `dao_custom_service_topology.go` | 保存静态拓扑边及过期时间。 |

## Validation And Rules

| Area | Implemented Rule |
|---|---|
| DataPlane Type | 创建和更新时必须能在 `data_plane_type` 中找到 `req.Typ`。 |
| DataPlane Params | 通过 `integration.CheckInvalid` 校验参数结构。 |
| Full Capability | 如果 `req.Capability.Obj` 为空，则自动采用 `dpType.CapabilitySpec.Obj`。 |
| Prometheus Hack | 当 `dpType.TypeName == "prometheus"` 时，后端会强制把 `promStorage` 写成 `"prom"`。 |
| APO Visibility | `SHOW_APO_DATA_PLANE_TYPE != "true"` 时，类型列表和已配置列表都会过滤掉 `apo`。 |
| Capability Summary | `GetIntegratedCapabilities` 会始终枚举全部 telemetry type；只要任一已配置 dataplane 声明支持该 capability，就标记 `enabled=true` 并记录来源。 |
| Rule Save | `SetServiceNameRule` 要求 `ServiceName` 非空。 |
| Rule Condition Match | `ServiceNameRuleCondition.Match` 支持 `equals`、`startsWith`、`endsWith`、`contains`、`has`。 |
| Custom Topology Create | `CreateCustomServiceTopology` 会检查 `cluster_id + left_node + right_node` 唯一。 |
| Minimal Agent Ready | `CheckMinimalDataPlaneExist` 最近 7 天 Prometheus cluster 非空时才视为 `agentReady`。 |

## Error Handling

| Area | Behavior |
|---|---|
| Binding | 各 handler 对参数绑定失败返回 `ParamBindError`。 |
| Missing Type | 创建或更新数据平面时找不到 type，返回 `DataPlaneTypeNotExistError`。 |
| Invalid Params | 数据平面参数非法时返回 `DataPlaneParamInvalidError`。 |
| Missing DataPlane | 更新或删除不存在的数据平面时返回 `DataPlaneNotExistError`。 |
| Capability Summary | 读取 `data_plane` 或 `data_plane_type` 失败时，handler 返回 `GetIntegratedCapabilitiesError`。 |
| Missing Rule | 删除规则前若规则不存在，DAO/service 会报错。 |
| Topology Query | 查询实时拓扑或静态拓扑失败时，`QueryTopologyResponse.Msg` 会写入错误信息。 |
| Datasource Stats | `GetDatasources` 若 DB 读取失败则直接返回 error。 |

## Notes

- `/api/dataplane/capabilities` 已在主路由注册，但它聚合的是“已配置 capability”，不是实时连通性或数据完整性校验。
- `service_get_service_topology.go` 会忽略 `ExpireTime < req.EndTime` 的静态拓扑边，说明静态边是可过期的。
- `GetServiceWithInstances` 先按已配置数据源过滤服务，再对实例打 `rule` 或 `source` 来源标记，这意味着未归因服务不会出现在前端 `serviceMatcher` 主表。
- `CheckMinimalDataPlaneExist` 中如果读取 Prometheus cluster 失败，会把 `clusterIDs` 置空继续返回，因此 `agentReady=false` 既可能表示没有 agent，也可能表示读取失败。
