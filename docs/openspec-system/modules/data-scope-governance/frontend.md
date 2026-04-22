# Data Scope Governance Frontend

## Pages

| Page | Route | Description |
|---|---|---|
| `DataGroupPage` | `/system/data-group` | 数据组主页面，左树右详情。 |
| `InfoModal` | 页面内弹窗 | 新增/编辑数据组和其 datasource scope。 |
| `PermissionModal` | 页面内弹窗 | 从“数据组 -> 用户/团队”方向授权。 |
| `DataGroupAuthorizeModal` | 复用弹窗 | 从“用户/团队 -> 数据组”方向授权。 |

## Routes

| File | Route | Component | Notes |
|---|---|---|---|
| `frontend/src/routes.js` | `/system/data-group` | `DataGroupPage` | 通过 `AuthRouter` 包装。 |

## Menus

| Source | Key | Route | Notes |
|---|---|---|---|
| `backend/pkg/repository/database/init_menu_item.go` | `dataGroup` | `/system/data-group` | 父级为 `manage`。 |
| `backend/pkg/repository/database/init_router.go` | `/system/data-group` | router metadata | `HideTimeSelector: true`。 |

## Key Components

| Component | File | Implemented Behavior |
|---|---|---|
| `DataGroupPage` | `frontend/src/core/views/DataGroup/index.tsx` | 管理当前选中组、弹窗和刷新逻辑。 |
| `DataGroupTree` | `frontend/src/core/views/DataGroup/DataGroupTree.tsx` | 以 `permissionType` 决定是否显示新增、编辑、授权、删除操作。 |
| `DataGroupInfoPanel` | `frontend/src/core/views/DataGroup/DataGroupInfoPanel/index.tsx` | 展示当前组的子组、datasource 和操作入口。 |
| `InfoModal` | `frontend/src/core/views/DataGroup/InfoModal.tsx` | 新增时调用 `/api/v2/data/group/add`，编辑时调用 `/api/v2/data/group/update`。 |
| `DatasourceSelector` | `frontend/src/core/views/DataGroup/component/DatasourceSelector.tsx` | 读取 `/api/v2/data/group/datasource/list`，支持 refresh 并显示 `protected` 项。 |
| `PermissionModal` | `frontend/src/core/views/DataGroup/PermissionModal/index.tsx` | 读取 `/api/data/subs`，保存到 `/api/data/subs/operation`。 |
| `DataGroupAuthorizeModal` | `frontend/src/core/components/PermissionAuthorize/DataGroupAuthorizeModal.tsx` | 读取 `/api/data/sub/group`，保存到 `/api/data/group/operation`。 |
| `DataGroupPermission` | `frontend/src/core/components/PermissionAuthorize/DataGroupPermission.tsx` | 以树选择方式展示可授权数据组。 |

## State Management

| State | File | Notes |
|---|---|---|
| Data Group Tree | `frontend/src/core/contexts/DataGroupContext.tsx` | 调 `/api/v2/data/group` 拉树并缓存。 |
| Disabled Node Logic | `DataGroupContext.tsx` | `permissionType === 'known'` 的节点会被禁用。 |
| Current Group | `DataGroupPage` | 当前选中组和父组由页面本地状态维护。 |
| Datasource Selection | `DatasourceSelector.tsx` | 本地维护 `checkedTreeKeys`、`datasourceList`、`protectedList`。 |
| Authorization State | `PermissionModal` / `DataGroupAuthorizeModal` | 分别维护 user/team 选择和主体已授权的数据组。 |

## API Usage

| UI Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 拉数据组树 | `getDatasourceByGroupApiV2` | `/api/v2/data/group` | 主页面入口。 |
| 拉可选 datasource | `getCheckableDatasourceApi` | `/api/v2/data/group/datasource/list` | 新增/编辑弹窗使用。 |
| 新增数据组 | `addDataGroupApi` | `/api/v2/data/group/add` | `InfoModal`。 |
| 更新数据组 | `updateDataGroupApiV2` | `/api/v2/data/group/update` | `InfoModal`。 |
| 读取组详情 | `getSubGroupsApiV2` | `/api/v2/data/group/detail` | 详情面板使用。 |
| 删除数据组 | `deleteDataGroupApiV2` | `/api/v2/data/group/delete` | 页面删除操作。 |
| 读取过滤树 | `getDatasourceFilterApiV2` | `/api/v2/data/group/filter` | endpoint / instance 过滤场景。 |
| 清理过期 scope | `refreshGroupDatasourceApiV2` | `/api/v2/data/group/datasource/refresh` | `DatasourceSelector` refresh 按钮。 |
| 读取组已授权主体 | `getDataGroupPermissionSubsApi` | `/api/data/subs` | `PermissionModal` 使用。 |
| 保存组授权主体 | `updateDataGroupSubsApi` | `/api/data/subs/operation` | `PermissionModal` 使用。 |
| 读取主体已授权组 | `getSubsDataGroupApi` | `/api/data/sub/group` | `DataGroupAuthorizeModal` 使用。 |
| 保存主体授权组 | `updateSubsDataGroupApi` | `/api/data/group/operation` | `DataGroupAuthorizeModal` 使用。 |

## User Interaction Flow

1. 用户进入 `/system/data-group`。
2. `DataGroupContext` 调 `/api/v2/data/group`，构造带 `permissionType` 的树。
3. 页面左侧 `DataGroupTree` 展示可见节点，右侧 `DataGroupInfoPanel` 展示当前组详情。
4. 用户新增或编辑数据组时，`InfoModal` 打开并调 `/api/v2/data/group/datasource/list` 加载 scope tree。
5. 用户保存后调用 `/api/v2/data/group/add` 或 `/api/v2/data/group/update`。
6. 如果用户删除数据组，前端会先检查当前节点是否还有 `subGroups`，有则直接报错，不发请求。
7. 用户在数据组详情中打开授权弹窗，可把该组授权给 user/team，实际走 `/api/data/subs` 与 `/api/data/subs/operation`。
8. 在用户或团队页面打开 `DataGroupAuthorizeModal` 时，前端从相反方向读取 `/api/data/sub/group` 并保存到 `/api/data/group/operation`。

## Permissions

| Area | Evidence | Notes |
|---|---|---|
| Route Guard | `AuthRouter` | 当前 route permission 检查被注释。 |
| Tree Editing | `DataGroupContext` + `DataGroupTree.tsx` | `known` 节点不可勾选、不可编辑。 |
| Backend Group Permission | `func_getdgscopelist_v2.go` | 读取 group scope 前会 `CheckGroupPermission`。 |
| Refresh Clean Protection | `DatasourceSelector.tsx` + `CleanExpiredDataScope` | 没权限删除的 scope 会返回到 `protected` 列表。 |

## Notes

- 前端主页面已经全面转向 V2，不应把旧版 `/api/data/group/create|update|delete` 写成当前页面主流程。
- `DataGroupAuthorizeModal` 会把 user 从 team 继承来的组单独展示在只读区域，这和 `PermissionModal` 是两个不同方向的授权视角。
- `DataGroupContext` 中写的是 `permissionType !== 'know'`，而模型常量是 `known`，这属于值得留意的拼写差异。
