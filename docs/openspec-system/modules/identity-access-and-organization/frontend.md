# Identity Access And Organization Frontend

## Pages

| Page | Route | Description |
|---|---|---|
| `Login` | `/login` | 登录页，支持部署模式品牌切换，并在非 `syncause` 模式下联动 Dify 登录。 |
| `UserPage` | `/user` | 个人中心，包含资料修改和密码修改。 |
| `UserManage` | `/system/user-manage` | 用户列表、新增、编辑、删除、重置密码和数据组授权。 |
| `MenuManagePage` | `/system/menu-manage` | 按角色切换并配置权限树。 |
| `RoleManage` | `/system/role-manage` | 角色新增、编辑、删除和权限配置。 |
| `TeamPage` | `/system/team` | 团队列表、新增、编辑、删除和数据组授权。 |

## Routes

| File | Route | Component | Notes |
|---|---|---|---|
| `frontend/src/App.js` | `/login` | `Login` | 未登录时主入口会跳转这里。 |
| `frontend/src/routes.js` | `/user` | `UserPage` | 个人中心。 |
| `frontend/src/routes.js` | `/system/user-manage` | `UserManage` | 系统管理。 |
| `frontend/src/routes.js` | `/system/menu-manage` | `MenuManage` | 系统管理。 |
| `frontend/src/routes.js` | `/system/role-manage` | `RoleManage` | 系统管理。 |
| `frontend/src/routes.js` | `/system/team` | `TeamPage` | 系统管理。 |

## Menus

| Source | Key | Route | Notes |
|---|---|---|---|
| `backend/pkg/repository/database/init_menu_item.go` | `userManage` | `/system/user-manage` | 父级为 `manage`。 |
| `backend/pkg/repository/database/init_menu_item.go` | `menuManage` | `/system/menu-manage` | 父级为 `manage`。 |
| `backend/pkg/repository/database/init_menu_item.go` | `team` | `/system/team` | 父级为 `manage`。 |
| `backend/pkg/repository/database/init_menu_item.go` | `role` | `/system/role-manage` | 父级为 `manage`。 |
| `backend/pkg/repository/database/init_router.go` | `/user` | 用户中心 | `HideTimeSelector: true`。 |

## Key Components

| Component | File | Implemented Behavior |
|---|---|---|
| `Login` | `frontend/src/core/views/Login/Login.jsx` | 调 `/api/user/login` 获取 `token`/`refreshToken`；非 `syncause` 模式下再调 `workflowLoginApi` 获取 Dify token。 |
| `UserInfo` | `frontend/src/core/views/UserPage/component/UserInfo.jsx` | 用 `/api/user/info` 读取资料，保存时调用 `/api/user/update/info`。 |
| `UpdatePassword` | `frontend/src/core/views/UserPage/component/UpdatePassword.jsx` | 调 `/api/user/update/password` 后立即 `/api/user/logout` 并清理本地 token。 |
| `UserManage` | `frontend/src/core/views/UserManage/index.tsx` | 组合搜索、用户表格、新增/编辑弹窗和数据组授权弹窗。 |
| `PermissionTree` | `frontend/src/core/components/PermissionTree/index.tsx` | 读取 `/api/permission/feature` 和 `/api/permission/sub/feature`，支持全选和保存。 |
| `MenuManagePage` | `frontend/src/core/views/MenuManage/index.tsx` | 左侧角色列表，右侧 `PermissionTree`，保存时调用 `updateRoleApi`。 |
| `RoleManage` | `frontend/src/core/views/RoleManage/index.tsx` | 角色表格 + 新增/编辑/权限配置弹窗。 |
| `TeamPage` | `frontend/src/core/views/Team/index.tsx` | 团队列表，支持新增、编辑、删除和 `DataGroupAuthorizeModal`。 |
| `AuthRouter` | `frontend/src/core/components/AuthRouter.tsx` | 预留路由权限拦截，但当前 effect 被注释。 |

## State Management

| State | File | Notes |
|---|---|---|
| User/Menu Context | `frontend/src/core/contexts/UserContext.js` | `getUserPermissionApi` 返回的 `menuItem` 和 `routes` 会写入 redux。 |
| Login State | `Login.jsx` | 本地持有 `loading`、`remeberMe`，并写 `token`、`refreshToken`、`difyToken`。 |
| User List State | `UserManage/index.tsx` | 本地维护分页、搜索条件、选中用户和弹窗状态。 |
| Role State | `useRoleManage.ts` | 管理角色列表和增删改请求状态。 |
| Menu Role State | `useMenuPermission.ts` | 管理角色选择和权限保存状态。 |
| Team State | `Team/index.tsx` | 管理团队列表、分页、当前编辑团队和授权弹窗。 |

## API Usage

| UI Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 登录 | `loginApi` | `/api/user/login` | 主登录请求。 |
| 登录后获取 Dify token | `workflowLoginApi` | `/dify/console/api/login/apo` | 仅 `deployMode !== 'syncause'` 时调用。 |
| 拉用户菜单和 routes | `getUserPermissionApi` | `/api/permission/config` | `UserContext` 使用。 |
| 单路由权限检查 | `getRouterPermissionApi` | `/api/permission/router` | `AuthRouter` 预留使用，当前未启用。 |
| 拉功能树 | `getAllPermissionApi` | `/api/permission/feature` | `PermissionTree` 使用。 |
| 拉主体权限 | `getSubjectPermissionApi` | `/api/permission/sub/feature` | `PermissionTree` 使用。 |
| 拉角色列表 | `getAllRolesApi` | `/api/role/roles` | 角色管理和菜单配置都依赖。 |
| 新增角色 | `createRoleApi` | `/api/role/create` | `RoleManage` 使用。 |
| 更新角色/权限 | `updateRoleApi` | `/api/role/update` | `RoleManage` 和 `MenuManage` 都使用。 |
| 删除角色 | `deleteRoleApi` | `/api/role/delete` | `RoleManage` 使用。 |
| 拉用户列表 | `getUserListApi` | `/api/user/list` | `UserManage` 主表。 |
| 新增用户 | `createUserApi` | `/api/user/create` | `AddUserModal` 使用。 |
| 删除用户 | `removeUserApi` | `/api/user/remove` | `UserManage` 使用。 |
| 管理员重置密码 | `updatePasswordWithNoOldPwdApi` | `/api/user/reset` | `EditUserModal` 使用。 |
| 个人资料读取 | `getUserInfoApi` | `api/user/info` | 个人中心使用。 |
| 个人资料保存 | `updateCorporationApi` | `/api/user/update/info` | 个人中心实际使用。 |
| 自己修改密码 | `updatePasswordApi` | `/api/user/update/password` | 个人中心使用。 |
| 团队列表 | `getTeamsApi` | `/api/team` | 团队页与授权弹窗使用。 |
| 新增团队 | `addTeamApi` | `/api/team/create` | `InfoModal` 使用。 |
| 更新团队 | `updateTeamApi` | `/api/team/update` | `InfoModal` 使用。 |
| 删除团队 | `deleteTeamApi` | `/api/team/delete` | `TeamPage` 使用。 |

## User Interaction Flow

1. 用户访问受保护页面时，主应用检测本地 token；未登录则跳到 `/login`。
2. 登录成功后，前端保存 `token` 与 `refreshToken`。
3. 如果 `deployMode !== 'syncause'`，前端继续请求 `/dify/console/api/login/apo`，保存 `difyToken`。
4. 用户进入系统后，`UserContext` 调 `/api/permission/config` 获取菜单和 routes。
5. `UserManage` 页通过 `/api/user/list` 管理用户，并可打开角色选择和数据组授权。
6. `RoleManage` 页通过 `/api/role/roles` 管理角色，并通过 `PermissionModal` 保存 feature 权限。
7. `MenuManage` 页左侧选中角色，右侧 `PermissionTree` 实际也是修改该角色的 `permissionList`。
8. `TeamPage` 通过 `/api/team` 管理团队，并可单独弹出 `DataGroupAuthorizeModal` 做数据组授权。
9. 用户在 `/user` 修改密码后，前端立即登出并跳回 `/login`。

## Permissions

| Area | Evidence | Notes |
|---|---|---|
| Route Guard | `AuthRouter` | 当前 `isAllowed` 初始就是 `true`，且真正检查逻辑已注释。 |
| Menu Visibility | `UserContext` + `/api/permission/config` | 真实菜单由后端 feature/menu/router 映射控制。 |
| Backend Middleware | `router_api.go` | `/api/permission`, `/api/role`, `/api/team` 需要 `AuthMiddleware()`；其中修改操作再叠 `LicMiddleware()`。 |
| Router Allowlist | `service_check_router.go` | `/agent-dataplane`, `agent`, `probe-installer` 被硬编码放行。 |

## Notes

- `MenuManage` 页面名字是“菜单管理”，但实际操作的是某个 role 的 `permissionList`，不是独立菜单配置接口。
- `configMenuApi` wrapper 当前未见主页面使用，不能把 `/api/permission/menu/configure` 写成现行前端主流程。
- 个人中心资料修改复用的是 `/api/user/update/info`，而不是 `/api/user/update/self`。
- `UpdatePassword` 前端正则要求至少 9 位且未显式要求数字；后端要求至少 8 位且必须包含数字，两端约束并不一致。
