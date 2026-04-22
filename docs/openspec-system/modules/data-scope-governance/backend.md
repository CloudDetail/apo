# Data Scope Governance Backend

## API Summary

| Category | Method | API Path | Handler | Notes |
|---|---|---|---|---|
| Legacy | GET | `/api/data/datasource` | `GetDatasource` | 获取全部 datasource。 |
| Legacy | POST | `/api/data/group` | `GetDataGroup` | 旧版数据组查询。 |
| Legacy | ANY | `/api/data/group/data` | `GetGroupDatasource` | 旧版 group datasource 查询。 |
| Legacy | POST | `/api/data/group/operation` | `DataGroupOperation` | 从主体视角更新数据组授权。 |
| Legacy | GET | `/api/data/sub/group` | `GetSubjectDataGroup` | 读取主体已授权数据组。 |
| Legacy | GET | `/api/data/user/group` | `GetUserDataGroup` | 读取用户组数据。 |
| Legacy | GET | `/api/data/subs` | `GetGroupSubs` | 读取数据组已授权主体。 |
| Legacy | POST | `/api/data/subs/operation` | `GroupSubsOperation` | 从数据组视角更新主体列表。 |
| V2 | GET | `/api/v2/data/group` | `GetDataGroupV2` | 返回当前用户可见的数据组树。 |
| V2 | GET | `/api/v2/data/group/datasource/list` | `GetDGScopeList` | 返回某组可选/已选 scope tree。 |
| V2 | GET | `/api/v2/data/group/detail` | `GetDGDetailV2` | 返回当前组子组与 datasource。 |
| V2 | POST | `/api/v2/data/group/add` | `CreateDataGroupV2` | 新建数据组。 |
| V2 | POST | `/api/v2/data/group/update` | `UpdateDataGroupV2` | 更新数据组。 |
| V2 | DELETE | `/api/v2/data/group/delete` | `DeleteDataGroupV2` | 删除数据组。 |
| V2 | POST | `/api/v2/data/group/filter` | `GetFilterByGroupIDV2` | 返回 endpoint 或 instance 扩展树。 |
| V2 | ANY | `/api/v2/data/group/datasource/refresh` | `CleanExpiredDataScope` | 清理过期 scope，并返回 `protected` 与 `toBeDeleted`。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 注册旧版 `/api/data` 与新版 `/api/v2/data`。 |
| `backend/pkg/api/data/handler.go` | 模块 handler interface。 |
| `func_getdatagroup_v2.go`, `func_createdatagroup_v2.go`, `func_updatedatagroup_v2.go`, `func_deletedatagroup.go` | 数据组 V2 CRUD。 |
| `func_getdgdetail_v2.go`, `func_getdgscopelist_v2.go`, `func_getfilterbygroupid_v2.go`, `func_cleanexpireddatascope.go` | V2 详情、scope 列表、过滤与清理。 |
| `func_getsubjectdatagroup.go`, `func_datagroupoperation.go` | 主体视角授权。 |
| `func_getgroupsubs.go`, `func_groupsubsoperation.go` | 数据组视角授权。 |

## Services

| Service | File | Behavior |
|---|---|---|
| `ListDataGroupV2` | `service_data_group_v2.go` | 基于当前用户的可见 groupID，用 `DataGroupStorage.CloneWithPermission` 构造树。 |
| `GetGroupDetailWithSubGroup` | `service_data_group_v2.go` | 返回当前组子组及其 datasource 详情。 |
| `CreateDataGroupV2` | `service_data_group_v2.go` | 校验父组存在与父组可授权范围后创建新组，并写 group2scope。 |
| `UpdateDataGroupV2` | `service_data_group_v2.go` | 校验新 scope 是否仍在父组授权范围内，并在必要时清理子组已不再可见的 scope。 |
| `DeleteDataGroupV2` | `service_data_group_v2.go` | 不允许删除仍有 `subGroups` 的组。 |
| `ListDataScopeByGroupID` | `service_data_scope.go` | 返回可选/已选 scope tree，并按 cluster 名称补全文案。 |
| `GetFilterByGroupID` | `service_data_scope.go` | 在 group 范围内补充 endpoint 或 instance 子节点。 |
| `CleanExpiredDataScope` | `service_data_scope.go` | 对比实时 scope 与当前 group 范围，返回可删和受保护列表，可选执行删除。 |
| `GetSubjectDataGroup` | `service_get_subject_data_group.go` | 对 user 会区分来自 user 自身还是 team 继承。 |
| `DataGroupOperation` | `service_data_group_operation.go` | 从 user/team 视角更新数据组授权。 |
| `GroupSubsOperation` | `service_group_subs_operation.go` | 从 data group 视角更新主体授权。 |

## Models And Schemas

| Model | File | Mapping |
|---|---|---|
| `DataGroup` | `backend/pkg/model/datagroup/data_group.go` | `TableName() == "data_group"`。 |
| `DataGroupTreeNode` | `data_group.go` | 树节点，带 `permissionType`。 |
| `DataScope` | `backend/pkg/model/datagroup/data_scope.go` | datasource scope 实体。 |
| `DataGroup2Scope` | `backend/pkg/model/datagroup/group2scope.go` | `group_id` 到 `scope_id` 关系。 |
| Auth Relation | `backend/pkg/repository/database/dao_auth_data_group.go` | user/team 到 group 授权关系。 |

## Database Mapping

| Storage | Evidence | Notes |
|---|---|---|
| `data_group` | `DataGroup.TableName()` | 数据组主表。 |
| `data_scope` | `InitRootGroup`, `dao_data_scope.go` | 实际 scope 存储。 |
| `data_group2scope` | `group2scope.go` | 数据组与 scope 的关联。 |
| `auth_data_group` | `dao_auth_data_group.go` | user/team 与 group 的授权关系。 |
| `datasource_group` | `InitRootGroup` 迁移逻辑 | 旧版 datasource group 迁移来源。 |

## Validation And Rules

| Area | Implemented Rule |
|---|---|
| Visible Tree | 如果当前用户没有任何 `permGroupIDs`，`CloneWithPermission` 返回只含 root 且 `permissionType=known` 的树。 |
| Create Under Parent | 新建数据组时，选中的 scope 必须在父组可授权范围内。 |
| Update Group Scope | 更新时，超出父组权限的新 scope 会被拒绝。 |
| Delete Group | 有子组时不能删除。 |
| Subject Operation | `DataGroupOperation` 会先校验 subjectType 和主体存在性。 |
| Group Subs Operation | `GroupSubsOperation` 会分别校验 user/team 是否存在。 |
| Root Group | `InitRootGroup` 会确保 `group_id = 0` 和 `APO_ALL_DATA` 存在。 |
| Protected Cleanup | 清理过期 scope 时，没有权限删除的 scope 会进 `protected` 列表。 |

## Error Handling

| Area | Behavior |
|---|---|
| Query Scope Without Permission | `GetDGScopeList` 会返回 permission error 和空结果结构。 |
| Create/Update Unauthorized Scope | V2 service 返回 `CreateDataGroupError` / `UpdateDataGroupError`。 |
| Delete Missing Group | 删除不存在的数据组会返回 `DataGroupNotExistError`。 |
| Delete Group With Children | 返回 `DeleteDataGroupError: data group has sub groups`。 |
| Refresh Scope | `CleanExpiredDataScope` 出错时返回 `CleanExpiredDataScopeError`。 |

## Notes

- V2 页面不是直接每次扫库构树，而是高度依赖 `DataGroupStorage` 这份共享缓存。
- `CleanExpiredDataScope` 在 `clean=true` 时不仅删 `group2scope` 关系，还会删除对应 scope 记录，并刷新缓存。
- `GetSubjectDataGroup` 对 team 直接返回 DB 结果，对 user 则会把 user 自身和 team 继承来源合并并标注 `source`。
