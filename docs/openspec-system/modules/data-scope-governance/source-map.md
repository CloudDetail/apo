# Data Scope Governance Source Map

## Scope
本文档固定 `data-scope-governance` 模块在当前仓库中的真实代码证据。该模块覆盖数据组树、数据源范围选择、按用户/团队的数据组授权、数据组与 scope 的 V2 管理能力，以及旧版 `/api/data/group*` 与新版 `/api/v2/data/group*` 共存时的边界差异。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/core/views/DataGroup` | `/system/data-group` 主页面，包含数据组树、详情面板、新增/编辑弹窗和授权弹窗。 |
| Frontend | `frontend/src/core/components/PermissionAuthorize` | 面向 user/team 的数据组授权弹窗。 |
| Frontend | `frontend/src/core/contexts/DataGroupContext.tsx` | 数据组树全局上下文。 |
| Frontend | `frontend/src/core/api/dataGroup.ts` | 旧版与 V2 数据组 API wrapper 共存。 |
| Backend | `backend/pkg/api/data` | `/api/data/*` 与 `/api/v2/data/*` handler 集合。 |
| Backend | `backend/pkg/services/data` | 数据组、scope、主体授权、过滤和清理逻辑。 |
| Backend | `backend/pkg/services/common/data_group_store_v2.go` | `DataGroupStorage` 初始化、scope 扫描与刷新。 |
| Backend | `backend/pkg/model/datagroup` | 数据组、scope、group2scope 等模型。 |
| Backend | `backend/pkg/repository/database/dao_data_group.go` | 旧版数据组 DAO。 |
| Backend | `backend/pkg/repository/database/dao_data_group_new.go` | V2 数据组树、root group 和权限树加载。 |
| Backend | `backend/pkg/repository/database/dao_data_scope.go` | scope 查询与删除。 |
| Backend | `backend/pkg/repository/database/dao_auth_data_group.go` | user/team 与 data group 关系。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/routes.js` | `/system/data-group`。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/data/datasource`, `/api/data/group`, `/api/data/group/data`, `/api/data/group/operation`, `/api/data/sub/group`, `/api/data/user/group`, `/api/data/subs`, `/api/data/subs/operation`。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/v2/data/group`, `/api/v2/data/group/datasource/list`, `/api/v2/data/group/detail`, `/api/v2/data/group/add`, `/api/v2/data/group/update`, `/api/v2/data/group/delete`, `/api/v2/data/group/filter`, `/api/v2/data/group/datasource/refresh`。 |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/core/views/DataGroup/index.tsx` | 主页面入口，左树右详情。 |
| Tree | `frontend/src/core/views/DataGroup/DataGroupTree.tsx` | 按 `permissionType` 控制可见操作。 |
| Info Panel | `frontend/src/core/views/DataGroup/DataGroupInfoPanel/index.tsx` | 展示当前组子组和 scope 摘要。 |
| Edit Modal | `frontend/src/core/views/DataGroup/InfoModal.tsx` | 新增/编辑数据组，并选择 scope。 |
| Group Authorization Modal | `frontend/src/core/views/DataGroup/PermissionModal/index.tsx` | 从“数据组 -> 用户/团队”方向授权。 |
| Subject Authorization Modal | `frontend/src/core/components/PermissionAuthorize/DataGroupAuthorizeModal.tsx` | 从“用户/团队 -> 数据组”方向授权。 |
| Datasource Selector | `frontend/src/core/views/DataGroup/component/DatasourceSelector.tsx` | 读取可选 scope tree，支持 refresh/clean。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/dataGroup.ts` | `getDatasourceByGroupApiV2` | `/api/v2/data/group` | 当前主页面数据组树入口。 |
| `frontend/src/core/api/dataGroup.ts` | `getCheckableDatasourceApi` | `/api/v2/data/group/datasource/list` | 新增/编辑数据组时读取可选 scope。 |
| `frontend/src/core/api/dataGroup.ts` | `addDataGroupApi` | `/api/v2/data/group/add` | 新建数据组。 |
| `frontend/src/core/api/dataGroup.ts` | `updateDataGroupApiV2` | `/api/v2/data/group/update` | 更新数据组。 |
| `frontend/src/core/api/dataGroup.ts` | `getSubGroupsApiV2` | `/api/v2/data/group/detail` | 详情面板读取子组和 datasource。 |
| `frontend/src/core/api/dataGroup.ts` | `deleteDataGroupApiV2` | `/api/v2/data/group/delete` | 删除数据组。 |
| `frontend/src/core/api/dataGroup.ts` | `getDatasourceFilterApiV2` | `/api/v2/data/group/filter` | 读取 endpoint/instance 扩展过滤树。 |
| `frontend/src/core/api/dataGroup.ts` | `refreshGroupDatasourceApiV2` | `/api/v2/data/group/datasource/refresh` | 清理过期 scope。 |
| `frontend/src/core/api/dataGroup.ts` | `getSubsDataGroupApi` | `/api/data/sub/group` | 按主体读取已授权数据组。 |
| `frontend/src/core/api/dataGroup.ts` | `updateSubsDataGroupApi` | `/api/data/group/operation` | 从 user/team 视角更新授权。 |
| `frontend/src/core/api/dataGroup.ts` | `getDataGroupPermissionSubsApi` | `/api/data/subs` | 从 data group 视角读取已授权主体。 |
| `frontend/src/core/api/dataGroup.ts` | `updateDataGroupSubsApi` | `/api/data/subs/operation` | 从 data group 视角更新主体列表。 |
| `frontend/src/core/api/dataGroup.ts` | `creatDataGroupApi`, `updateDataGroupApi`, `deleteDataGroupApi` | `/api/data/group/create|update|delete` | 旧版 wrapper 仍保留，但当前主页面未使用。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | 注册 `/api/data` 与 `/api/v2/data`。 |
| Handler Definition | `backend/pkg/api/data/handler.go` | 汇总旧版与 V2 handler。 |
| V2 Group CRUD | `func_getdatagroup_v2.go`, `func_createdatagroup_v2.go`, `func_updatedatagroup_v2.go`, `func_deletedatagroup.go` | 数据组树与增删改。 |
| V2 Scope APIs | `func_getdgdetail_v2.go`, `func_getdgscopelist_v2.go`, `func_getfilterbygroupid_v2.go`, `func_cleanexpireddatascope.go` | 详情、scope 列表、过滤树和清理。 |
| Old Subject/Auth APIs | `func_getsubjectdatagroup.go`, `func_getgroupsubs.go`, `func_groupsubsoperation.go`, `func_datagroupoperation.go` | user/team 与 data group 授权关系维护。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| DataGroup V2 | `backend/pkg/services/data/service_data_group_v2.go` | 数据组树读取、详情、新增、更新、删除。 |
| Data Scope | `backend/pkg/services/data/service_data_scope.go` | scope tree、endpoint/instance 过滤和过期 scope 清理。 |
| Subject Authorization | `backend/pkg/services/data/service_get_subject_data_group.go`, `service_data_group_operation.go` | user/team -> data group 方向。 |
| Group Authorization | `backend/pkg/services/data/service_get_group_subs.go`, `service_group_subs_operation.go` | data group -> user/team 方向。 |
| Legacy Group APIs | `backend/pkg/services/data/service_data_group.go`, `service_data_group_operation.go` | 旧版数据组接口。 |
| Shared Storage | `backend/pkg/services/common/data_group_store_v2.go` | root group 初始化、scope 刷新和树缓存。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| `DataGroup` / `DataGroupTreeNode` | `backend/pkg/model/datagroup/data_group.go` | 数据组主实体和树节点，带 `permissionType`。 |
| `DataScope` | `backend/pkg/model/datagroup/data_scope.go` | 可治理的数据范围实体。 |
| `DataGroup2Scope` | `backend/pkg/model/datagroup/group2scope.go` | 组与 scope 关系。 |
| Auth Relation | `backend/pkg/repository/database/dao_auth_data_group.go` | user/team 与 group 关系。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| `UserContext` | `frontend/src/core/contexts/UserContext.js` | 依赖当前登录用户上下文。 |
| Team/User APIs | `frontend/src/core/api/team.ts`, `frontend/src/core/api/user.js` | 授权弹窗需要可选主体列表。 |
| Prometheus / ClickHouse | `service_data_scope.go`, `data_group_store_v2.go` | scope 扫描和过滤依赖观测数据。 |
| `DataGroupStorage` | `backend/pkg/services/common/data_group_store_v2.go` | V2 树和 scope 缓存的核心。 |

## Evidence Notes

1. 当前主页面明确走 V2：`/api/v2/data/group*`。
2. `DataGroupContext` 会把 `permissionType === 'known'` 的节点标记为 disabled。
3. `DeleteDataGroupV2` 后端和前端都明确禁止删除仍有 `subGroups` 的数据组。
4. `InitRootGroup` 会创建 `group_id = 0` 的 root group，并绑定系统 scope `APO_ALL_DATA`。
5. `DataGroupAuthorizeModal` 在 user 维度会区分 `source === user` 与 team 继承的数据组。
