# Identity Access And Organization Module Overview

## Module Goal
本模块负责系统中的身份、访问控制和组织边界。它不仅包含登录与个人中心，还覆盖用户、角色、菜单权限、团队和功能权限模型的管理；前端通过系统管理页面和 `UserContext` 消费这些能力，后端则通过 `user`、`permission`、`role`、`team` 四组 API 提供实现。

## Scope
当前代码证据支持以下范围：

- `/login` 登录、token 刷新与登出。
- `/user` 个人资料修改与密码修改。
- `/system/user-manage` 用户列表、新增、编辑、删除、重置密码。
- `/system/role-manage` 角色创建、编辑、删除和权限配置。
- `/system/menu-manage` 以角色为维度配置功能树。
- `/system/team` 团队创建、编辑、删除和成员维护。
- 基于 `/api/permission/config` 的菜单与 route 装载。
- router 级权限检查逻辑和后端白名单。
- 用户与 Dify 用户之间的创建、删除、密码同步。

当前代码不应直接纳入本模块主体范围的内容：

- `data-scope-governance` 的数据组树和数据源范围治理；这里只记录其在用户/团队授权中的复用。
- `platform-administration` 中的 deploy mode / license；这里只在登录和布局处作为外部条件出现。
- `/dify/console/api/*` 的真实后端实现；当前仓库只体现为外部边界。

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Login Route | `frontend/src/App.js` | `/login`。 |
| System Routes | `frontend/src/routes.js` | `/user`, `/system/user-manage`, `/system/menu-manage`, `/system/role-manage`, `/system/team`。 |
| Context | `frontend/src/core/contexts/UserContext.js` | 菜单和 routes 的全局状态来源。 |
| Route Guard | `frontend/src/core/components/AuthRouter.tsx` | 所有系统管理页面统一包装。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Groups | `backend/pkg/router/router_api.go` | `/api/user`, `/api/permission`, `/api/role`, `/api/team`。 |
| User Services | `backend/pkg/services/user` | 登录、用户 CRUD、改密、资料更新。 |
| Permission Services | `backend/pkg/services/permission` | feature 聚合、菜单和 router 配置。 |
| Role Services | `backend/pkg/services/role` | 角色与 feature 绑定。 |
| Team Services | `backend/pkg/services/team` | 团队与成员关系维护。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `feature` | 系统的最小权限能力单元，菜单、路由和主体权限都围绕它映射。 |
| `menuItem` | 菜单树节点，最终由 `GetUserConfig` 返回给前端。 |
| `router` | 可访问路由的元数据，与菜单一起构成前端导航模型。 |
| `role` | feature 权限集合，可授予用户。 |
| `team` | 用户组织单元，既可承载成员，也可承载 feature/data group 权限。 |
| `UserContext` | 前端全局用户、菜单和 routes 状态上下文。 |
| Dify user sync | 创建用户、删用户、改密码时与外部 Dify 用户侧保持同步。 |

## Related Files

- `frontend/src/App.js`
- `frontend/src/routes.js`
- `frontend/src/core/views/Login/Login.jsx`
- `frontend/src/core/views/UserPage/index.jsx`
- `frontend/src/core/views/UserManage/index.tsx`
- `frontend/src/core/views/MenuManage/index.tsx`
- `frontend/src/core/views/RoleManage/index.tsx`
- `frontend/src/core/views/Team/index.tsx`
- `frontend/src/core/components/AuthRouter.tsx`
- `frontend/src/core/components/PermissionTree/index.tsx`
- `frontend/src/core/contexts/UserContext.js`
- `frontend/src/core/api/user.js`
- `frontend/src/core/api/permission.js`
- `frontend/src/core/api/role.ts`
- `frontend/src/core/api/team.ts`
- `backend/pkg/api/user/handler.go`
- `backend/pkg/api/permission/handler.go`
- `backend/pkg/api/role/handler.go`
- `backend/pkg/api/team/handler.go`
- `backend/pkg/services/user/service_create_user.go`
- `backend/pkg/services/user/service_remove_user.go`
- `backend/pkg/services/user/service_update_info.go`
- `backend/pkg/services/permission/service_user_configs.go`
- `backend/pkg/services/permission/service_check_router.go`
- `backend/pkg/services/role/service_create_role.go`
- `backend/pkg/services/role/service_update_role.go`
- `backend/pkg/services/team/service_create_team.go`
- `backend/pkg/services/team/service_update_team.go`
- `backend/pkg/model/profile/user.go`
- `backend/pkg/model/profile/role.go`
- `backend/pkg/model/profile/team.go`
- `backend/pkg/model/profile/feature.go`
- `backend/pkg/repository/database/init_feature.go`
- `backend/pkg/repository/database/init_feature_mapping.go`
- `backend/pkg/repository/database/init_menu_item.go`
- `backend/pkg/repository/database/init_router.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `request.js` | 负责 token 注入、过期跳转登录。 |
| Cross Module | `data-scope-governance` | 用户和团队页复用 `DataGroupAuthorizeModal`。 |
| Cross Module | `platform-administration` | 登录页 `useDeployMode` 会影响品牌名和是否登录 Dify。 |
| Backend Shared | `jwt.GenerateTokens` | 登录与刷新 token 的基础。 |
| Backend Shared | repository/database init metadata | 功能、菜单、路由初始化决定整个权限模型边界。 |
| External Service | Dify | workflow 登录和用户同步。 |

## Risks And Unknowns

1. `AuthRouter` 的前端 route permission 检查已注释，当前用户可见页面与后端权限模型可能不一致。
2. `MenuManage` 实际调用的是 `updateRoleApi`，而不是 `/api/permission/menu/configure`，页面名称与后端接口设计存在偏差。
3. 前端个人中心 `UpdatePassword` 的密码正则与后端复杂度校验并不完全一致。
4. 用户资料页使用 `/api/user/update/info`，而不是看起来更语义化的 `/api/user/update/self`。
5. 团队更新 service 中与 feature/data group 的更新逻辑被注释，说明团队功能存在历史收缩或拆分。
