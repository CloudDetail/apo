# Data Scope Governance Module Overview

## Module Goal
本模块负责把平台可见的 cluster/namespace/service 等观测数据范围组织成一棵可授权、可继承、可裁剪的数据组树。前端通过 `/system/data-group` 管理数据组、scope 和授权关系；后端通过 `/api/v2/data/group*` 与 `/api/data/*` 提供树加载、scope 选择、授权和清理能力。

## Scope
当前代码证据支持以下范围：

- `/system/data-group` 数据组树展示、详情、新增、编辑、删除。
- 基于 scope tree 的数据范围选择。
- endpoint / instance 级的附加过滤树查询。
- 过期 scope 扫描和清理。
- user/team 维度的数据组授权。
- data group 维度的用户/团队授权。
- root group 与系统 scope 初始化。

当前代码不应直接写入本模块主范围的内容：

- 用户、团队本身的管理流程；这里只记录它们如何被授权数据组。
- 菜单权限与功能权限配置；它们属于 `identity-access-and-organization`。
- 观测查询页面的实际过滤消费；这里只记录过滤对象与数据范围模型。

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/routes.js` | `/system/data-group`。 |
| Page Root | `frontend/src/core/views/DataGroup/index.tsx` | 主页面入口。 |
| Context | `frontend/src/core/contexts/DataGroupContext.tsx` | 数据组树共享状态。 |
| Authorization Modal | `frontend/src/core/components/PermissionAuthorize/DataGroupAuthorizeModal.tsx` | 用户和团队页复用。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Groups | `backend/pkg/router/router_api.go` | `/api/data` 与 `/api/v2/data`。 |
| Data Service | `backend/pkg/services/data` | 数据组、scope 和授权逻辑。 |
| Shared Storage | `backend/pkg/services/common/data_group_store_v2.go` | 树缓存与 scope 刷新。 |
| DAO Layer | `dao_data_group*.go`, `dao_data_scope.go`, `dao_auth_data_group.go` | 数据组与 scope 持久化。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `DataGroup` | 可授权的数据分组节点。 |
| `DataScope` | 具体的数据范围单元，通常由 cluster / namespace / service 层级构成。 |
| `DataGroup2Scope` | 数据组与 scope 的绑定关系。 |
| `permissionType` | 当前用户在某个数据组节点上的可见/可编辑状态，主要有 `known`、`view`、`edit`。 |
| Root Group `0` | 系统级根组，名称为 `ALL`，持有 `APO_ALL_DATA`。 |
| `DataGroupStorage` | 缓存后的数据组树和 scope 树，同时负责周期刷新。 |
| Subject Authorization | user/team 拥有哪些数据组。 |
| Group Authorization | 某个数据组授权给了哪些 user/team。 |

## Related Files

- `frontend/src/routes.js`
- `frontend/src/core/views/DataGroup/index.tsx`
- `frontend/src/core/views/DataGroup/DataGroupTree.tsx`
- `frontend/src/core/views/DataGroup/DataGroupInfoPanel/index.tsx`
- `frontend/src/core/views/DataGroup/InfoModal.tsx`
- `frontend/src/core/views/DataGroup/PermissionModal/index.tsx`
- `frontend/src/core/views/DataGroup/component/DatasourceSelector.tsx`
- `frontend/src/core/components/PermissionAuthorize/DataGroupAuthorizeModal.tsx`
- `frontend/src/core/contexts/DataGroupContext.tsx`
- `frontend/src/core/api/dataGroup.ts`
- `backend/pkg/api/data/handler.go`
- `backend/pkg/services/data/service_data_group_v2.go`
- `backend/pkg/services/data/service_data_scope.go`
- `backend/pkg/services/data/service_get_subject_data_group.go`
- `backend/pkg/services/data/service_data_group_operation.go`
- `backend/pkg/services/data/service_group_subs_operation.go`
- `backend/pkg/services/common/data_group_store_v2.go`
- `backend/pkg/model/datagroup/data_group.go`
- `backend/pkg/model/datagroup/data_scope.go`
- `backend/pkg/model/datagroup/group2scope.go`
- `backend/pkg/repository/database/dao_data_group_new.go`
- `backend/pkg/repository/database/dao_data_scope.go`
- `backend/pkg/repository/database/dao_auth_data_group.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `UserContext` | 当前用户身份决定可见数据组树。 |
| Frontend Shared | `team` / `user` APIs | 授权弹窗需要可选主体列表。 |
| Cross Module | `identity-access-and-organization` | user/team 页面复用本模块授权弹窗。 |
| Backend Shared | Prometheus / ClickHouse | scope 扫描和过滤依赖实时观测数据。 |
| Backend Shared | `DataGroupStorage` | V2 页面和授权链路都依赖这份共享缓存。 |

## Risks And Unknowns

1. 前端旧版 `/api/data/group/create|update|delete` wrapper 仍保留，但当前后端主注册和页面使用都已转到 V2。
2. 授权存在两条方向相反的接口链路，容易在文档里混淆。
3. `DataGroupStorage` 同时承担缓存和刷新职责，状态来源并非只有数据库静态表。
4. 过期 scope 清理会涉及权限保护逻辑，用户看到的 `protected` 项并不一定意味着它们在观测侧仍存在。
