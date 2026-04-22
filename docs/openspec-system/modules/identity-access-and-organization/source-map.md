# Identity Access And Organization Source Map

## Scope
本文档固定 `identity-access-and-organization` 模块在当前仓库中的真实代码证据。该模块覆盖登录、个人中心、用户管理、角色管理、菜单/功能权限配置、团队管理，以及这些能力与后端权限模型、菜单路由元数据和 Dify 用户同步之间的连接关系。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/core/views/Login` | `/login` 登录页。 |
| Frontend | `frontend/src/core/views/UserPage` | `/user` 个人中心，包含资料修改和密码修改。 |
| Frontend | `frontend/src/core/views/UserManage` | `/system/user-manage` 用户管理页。 |
| Frontend | `frontend/src/core/views/MenuManage` | `/system/menu-manage` 菜单/权限配置页。 |
| Frontend | `frontend/src/core/views/RoleManage` | `/system/role-manage` 角色管理页。 |
| Frontend | `frontend/src/core/views/Team` | `/system/team` 团队管理页。 |
| Frontend | `frontend/src/core/components/AuthRouter.tsx` | 路由级权限包装。 |
| Frontend | `frontend/src/core/components/PermissionTree/index.tsx` | 功能权限树组件。 |
| Frontend | `frontend/src/core/contexts/UserContext.js` | 基于 `/api/permission/config` 拉菜单和 routes。 |
| Frontend | `frontend/src/core/api/user.js` | `/api/user/*` wrapper。 |
| Frontend | `frontend/src/core/api/permission.js` | `/api/permission/*` wrapper。 |
| Frontend | `frontend/src/core/api/role.ts` | `/api/role/*` wrapper。 |
| Frontend | `frontend/src/core/api/team.ts` | `/api/team/*` wrapper。 |
| Backend | `backend/pkg/api/user` | 登录、用户 CRUD、密码和个人资料相关 handler。 |
| Backend | `backend/pkg/api/permission` | 菜单路由配置、feature 列表、router permission 检查。 |
| Backend | `backend/pkg/api/role` | 角色 CRUD 与用户角色操作。 |
| Backend | `backend/pkg/api/team` | 团队 CRUD、成员关系维护。 |
| Backend | `backend/pkg/services/user` | 登录、建删用户、密码和 Dify 同步逻辑。 |
| Backend | `backend/pkg/services/permission` | feature 聚合、用户菜单配置、router 白名单。 |
| Backend | `backend/pkg/services/role` | 角色与功能权限关系维护。 |
| Backend | `backend/pkg/services/team` | 团队、成员、权限和数据组关联维护。 |
| Backend | `backend/pkg/model/profile` | `user`、`role`、`team`、`feature` 等主实体。 |
| Backend | `backend/pkg/repository/database/init_feature.go` | 预置功能项。 |
| Backend | `backend/pkg/repository/database/init_feature_mapping.go` | 功能到菜单/路由的映射。 |
| Backend | `backend/pkg/repository/database/init_menu_item.go` | 菜单树元数据。 |
| Backend | `backend/pkg/repository/database/init_router.go` | 路由元数据。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/App.js` | `/login`。 |
| Frontend | `frontend/src/routes.js` | `/user`, `/system/user-manage`, `/system/menu-manage`, `/system/role-manage`, `/system/team`。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/user/*`, `/api/permission/*`, `/api/role/*`, `/api/team/*`。 |
| Permission Metadata | `backend/pkg/repository/database/init_router.go` | `/user`, `/system/user-manage`, `/system/menu-manage`, `/system/team`, `/system/role-manage` 等路由的系统元数据。 |

## Page Files

| Type | File | Notes |
|---|---|---|
| Login Page | `frontend/src/core/views/Login/Login.jsx` | 调 `/api/user/login`，成功后在非 `syncause` 模式下额外调用 `workflowLoginApi`。 |
| User Center | `frontend/src/core/views/UserPage/index.jsx` | 组合 `UserInfo` 与 `UpdatePassword`。 |
| User Info Form | `frontend/src/core/views/UserPage/component/UserInfo.jsx` | 读取 `/api/user/info`，保存时调用 `/api/user/update/info`。 |
| Password Form | `frontend/src/core/views/UserPage/component/UpdatePassword.jsx` | 调 `/api/user/update/password`，成功后立即 `/api/user/logout` 并跳转登录页。 |
| User Management | `frontend/src/core/views/UserManage/index.tsx` | 用户列表、新增、编辑、重置密码和数据组授权入口。 |
| Menu Management | `frontend/src/core/views/MenuManage/index.tsx` | 以角色为维度配置功能权限树。 |
| Role Management | `frontend/src/core/views/RoleManage/index.tsx` | 角色新增、编辑、删除和权限配置。 |
| Team Management | `frontend/src/core/views/Team/index.tsx` | 团队列表、新增、编辑、删除和数据组授权入口。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/user.js` | `loginApi` | `/api/user/login` | 登录。 |
| `frontend/src/core/api/user.js` | `logoutApi` | `/api/user/logout` | 登出。 |
| `frontend/src/core/api/user.js` | `getUserInfoApi` | `api/user/info` | 个人中心读取资料。 |
| `frontend/src/core/api/user.js` | `getUserListApi` | `/api/user/list` | 用户管理列表。 |
| `frontend/src/core/api/user.js` | `createUserApi` | `/api/user/create` | 新增用户。 |
| `frontend/src/core/api/user.js` | `updatePasswordApi` | `/api/user/update/password` | 用户自己修改密码。 |
| `frontend/src/core/api/user.js` | `updateCorporationApi` | `/api/user/update/info` | 个人中心和管理页都复用。 |
| `frontend/src/core/api/user.js` | `removeUserApi` | `/api/user/remove` | 删除用户。 |
| `frontend/src/core/api/user.js` | `updatePasswordWithNoOldPwdApi` | `/api/user/reset` | 管理员重置密码。 |
| `frontend/src/core/api/permission.js` | `getUserPermissionApi` | `/api/permission/config` | 拉菜单和 routes。 |
| `frontend/src/core/api/permission.js` | `getAllPermissionApi` | `/api/permission/feature` | 拉完整 feature 树。 |
| `frontend/src/core/api/permission.js` | `getSubjectPermissionApi` | `/api/permission/sub/feature` | 拉 role/user 的 feature。 |
| `frontend/src/core/api/permission.js` | `getRouterPermissionApi` | `/api/permission/router` | 单路由校验。 |
| `frontend/src/core/api/permission.js` | `configMenuApi` | `/api/permission/menu/configure` | wrapper 存在，当前主页面未见使用。 |
| `frontend/src/core/api/role.ts` | `getAllRolesApi` | `/api/role/roles` | 角色列表。 |
| `frontend/src/core/api/role.ts` | `createRoleApi` | `/api/role/create` | 新增角色。 |
| `frontend/src/core/api/role.ts` | `updateRoleApi` | `/api/role/update` | 更新角色及其权限。 |
| `frontend/src/core/api/role.ts` | `deleteRoleApi` | `/api/role/delete` | 删除角色。 |
| `frontend/src/core/api/role.ts` | `revokeUserRoleApi` | `/api/role/operation` | 用户角色调整。 |
| `frontend/src/core/api/team.ts` | `getTeamsApi` | `/api/team` | 团队列表。 |
| `frontend/src/core/api/team.ts` | `addTeamApi` | `/api/team/create` | 新增团队。 |
| `frontend/src/core/api/team.ts` | `updateTeamApi` | `/api/team/update` | 更新团队。 |
| `frontend/src/core/api/team.ts` | `deleteTeamApi` | `/api/team/delete` | 删除团队。 |
| `frontend/src/core/api/workflows.ts` | `workflowLoginApi` | `/dify/console/api/login/apo` | 非 `syncause` 模式登录后额外调用。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| User Router | `backend/pkg/api/user/handler.go` | 登录、刷新 token、用户 CRUD、更新资料、重置密码。 |
| Permission Router | `backend/pkg/api/permission/handler.go` | 菜单配置、feature 树、subject feature、router permission。 |
| Role Router | `backend/pkg/api/role/handler.go` | 角色 CRUD 和 user-role 操作。 |
| Team Router | `backend/pkg/api/team/handler.go` | 团队 CRUD、成员操作和团队成员查询。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Login/Auth | `backend/pkg/services/user/service_login.go`, `service_refresh_token.go`, `service_logout.go` | 登录与 token 刷新。 |
| User CRUD | `service_create_user.go`, `service_remove_user.go`, `service_get_user_info.go` | 用户创建、删除、列表和详情。 |
| User Profile | `service_update_info.go` | 更新资料、修改密码、重置密码，并同步 Dify。 |
| Permission Aggregation | `backend/pkg/services/permission/service_user_configs.go`, `service_get_user_permission.go`, `service_permission.go` | 汇总 role/team/user 权限，构造菜单与路由。 |
| Router Check | `service_check_router.go` | 路由匹配与白名单逻辑。 |
| Role Governance | `backend/pkg/services/role/service_create_role.go`, `service_update_role.go`, `service_delete_role.go`, `service_user_role.go` | 角色及其 feature 关系维护。 |
| Team Governance | `backend/pkg/services/team/service_create_team.go`, `service_update_team.go`, `service_delete_team.go`, `service_team_opreation.go`, `service_team_user_operation.go` | 团队及成员关系维护。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| `User` | `backend/pkg/model/profile/user.go` | `TableName() == "user"`。 |
| `Role` / `UserRole` | `backend/pkg/model/profile/role.go` | `role` 与 `user_role`。 |
| `Team` / `UserTeam` | `backend/pkg/model/profile/team.go` | `team` 与 `user_team`。 |
| `Feature` | `backend/pkg/model/profile/feature.go` | `TableName() == "feature"`。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| `UserContext` | `frontend/src/core/contexts/UserContext.js` | 全站菜单和 routes 来源。 |
| `PermissionTree` | `frontend/src/core/components/PermissionTree/index.tsx` | 角色/菜单权限配置复用组件。 |
| `DataGroupAuthorizeModal` | `frontend/src/core/components/PermissionAuthorize/DataGroupAuthorizeModal.tsx` | 用户和团队页面都复用数据组授权弹窗。 |
| Dify Repo | `backend/pkg/repository/dify` | 用户创建、删除、改密、重置密码和 workflow 登录边界。 |
| Init Metadata | `init_feature.go`, `init_feature_mapping.go`, `init_menu_item.go`, `init_router.go` | 权限与菜单体系的底层元数据。 |

## Evidence Notes

1. `AuthRouter` 中真正调用 `/api/permission/router` 的 `useEffect` 已被整段注释，前端当前不会实际拦截无权限页面。
2. `UserContext` 通过 `/api/permission/config` 读取 `menuItem` 和 `routes`，是全站导航的主数据源。
3. `CheckRouterPermission` 后端存在硬编码放行：`/agent-dataplane`、`agent`、`probe-installer`。
4. 登录成功后，前端会在非 `syncause` 模式下额外调用 `/dify/console/api/login/apo` 获取 Dify token。
5. 后端 `CreateUser`、`RemoveUser`、`UpdateUserPassword`、`RestPassword` 都会联动 Dify 用户。
