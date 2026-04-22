# Identity Access And Organization Backend

## API Summary

| Category | Method | API Path | Handler | Notes |
|---|---|---|---|---|
| Auth | POST | `/api/user/login` | `Login` | 登录并返回 access/refresh token。 |
| Auth | POST | `/api/user/logout` | `Logout` | 登出。 |
| Auth | GET | `/api/user/refresh` | `RefreshToken` | 刷新 access token。 |
| User | GET | `/api/user/info` | `GetUserInfo` | 获取当前用户信息。 |
| User | GET | `/api/user/list` | `GetUserList` | 用户列表。 |
| User | GET | `/api/user/team` | `GetUserTeam` | 获取用户所属团队。 |
| User | POST | `/api/user/create` | `CreateUser` | 创建用户。 |
| User | POST | `/api/user/update/password` | `UpdateUserPassword` | 用户自己改密。 |
| User | POST | `/api/user/update/phone` | `UpdateUserPhone` | 更新手机号。 |
| User | POST | `/api/user/update/email` | `UpdateUserEmail` | 更新邮箱。 |
| User | POST | `/api/user/update/info` | `UpdateUserInfo` | 更新资料。 |
| User | POST | `/api/user/update/self` | `UpdateSelfInfo` | 自己更新资料。 |
| User | POST | `/api/user/remove` | `RemoveUser` | 删除用户。 |
| User | POST | `/api/user/reset` | `ResetPassword` | 管理员重置密码。 |
| Permission | GET | `/api/permission/config` | `GetUserConfig` | 返回菜单与 routes。 |
| Permission | GET | `/api/permission/feature` | `GetFeature` | 返回完整 feature 树。 |
| Permission | GET | `/api/permission/sub/feature` | `GetSubjectFeature` | 返回某主体当前 feature。 |
| Permission | GET | `/api/permission/router` | `CheckRouterPermission` | 单路由权限检查。 |
| Permission | POST | `/api/permission/operation` | `PermissionOperation` | 给 user 直接授予或撤销 feature。 |
| Permission | POST | `/api/permission/menu/configure` | `ConfigureMenu` | 全局菜单配置接口，当前前端主流程未见使用。 |
| Role | GET | `/api/role/roles` | `GetRole` | 角色列表。 |
| Role | GET | `/api/role/user` | `GetUserRole` | 获取用户角色。 |
| Role | POST | `/api/role/operation` | `RoleOperation` | 给用户授予/撤销角色。 |
| Role | POST | `/api/role/create` | `CreateRole` | 创建角色。 |
| Role | POST | `/api/role/update` | `UpdateRole` | 更新角色及其权限。 |
| Role | POST | `/api/role/delete` | `DeleteRole` | 删除角色。 |
| Team | GET | `/api/team` | `GetTeam` | 团队列表。 |
| Team | GET | `/api/team/user` | `GetTeamUser` | 获取团队成员。 |
| Team | POST | `/api/team/create` | `CreateTeam` | 创建团队。 |
| Team | POST | `/api/team/update` | `UpdateTeam` | 更新团队。 |
| Team | POST | `/api/team/delete` | `DeleteTeam` | 删除团队。 |
| Team | POST | `/api/team/operation` | `TeamOperation` | 从用户角度批量调整团队。 |
| Team | POST | `/api/team/user/operation` | `TeamUserOperation` | 从团队角度批量调整成员。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 注册 `/api/user`, `/api/permission`, `/api/role`, `/api/team`。 |
| `backend/pkg/api/user/handler.go` | 用户与认证相关接口。 |
| `backend/pkg/api/permission/handler.go` | feature、菜单、router 相关接口。 |
| `backend/pkg/api/role/handler.go` | 角色相关接口。 |
| `backend/pkg/api/team/handler.go` | 团队相关接口。 |

## Services

| Service | File | Behavior |
|---|---|---|
| `Login` | `backend/pkg/services/user/service_login.go` | 校验用户名密码后调用 `jwt.GenerateTokens`。 |
| `CreateUser` | `service_create_user.go` | 校验用户名格式、密码复杂度、角色、团队，然后创建 DB 用户并同步创建 Dify 用户。 |
| `RemoveUser` | `service_remove_user.go` | 撤销角色、权限、数据组、团队关系后删除用户，并同步删除 Dify 用户。 |
| `UpdateUserPassword` / `RestPassword` | `service_update_info.go` | 更新本地密码后同步更新/重置 Dify 密码。 |
| `UpdateUserInfo` / `UpdateSelfInfo` | `service_update_info.go` | 更新手机、邮箱、公司信息。 |
| `GetUserConfig` | `backend/pkg/services/permission/service_user_configs.go` | 聚合 feature、菜单、router 和国际化文本，返回前端菜单树与 routes。 |
| `getUserFeatureIDs` | `service_get_user_permission.go` | 把 role、team、user 直授 feature 合并为统一 feature 列表。 |
| `CheckRouterPermission` | `service_check_router.go` | 先检查白名单，再按菜单和 route 映射判断。 |
| `CreateRole` | `backend/pkg/services/role/service_create_role.go` | 校验重名、feature 合法性和 user 存在性后创建角色并授权。 |
| `UpdateRole` | `service_update_role.go` | 计算 feature 的增删集并事务更新。 |
| `DeleteRole` | `service_delete_role.go` | 若角色仍被授予则拒绝删除。 |
| `CreateTeam` | `backend/pkg/services/team/service_create_team.go` | 校验团队名、feature、data group、user 后创建团队并建立成员/权限关系。 |
| `UpdateTeam` | `service_update_team.go` | 当前只稳定更新团队名、描述和用户成员；feature/data group 更新逻辑已注释。 |
| `DeleteTeam` | `service_delete_team.go` | 删除团队并撤销 feature、成员关系和数据组授权。 |

## Models And Schemas

| Model | File | Mapping |
|---|---|---|
| `User` | `backend/pkg/model/profile/user.go` | `TableName() == "user"`。 |
| `Role` | `backend/pkg/model/profile/role.go` | `TableName() == "role"`。 |
| `UserRole` | `backend/pkg/model/profile/role.go` | `TableName() == "user_role"`。 |
| `Team` | `backend/pkg/model/profile/team.go` | `TableName() == "team"`。 |
| `UserTeam` | `backend/pkg/model/profile/team.go` | `TableName() == "user_team"`。 |
| `Feature` | `backend/pkg/model/profile/feature.go` | `TableName() == "feature"`。 |

## Database Mapping

| Storage | Evidence | Notes |
|---|---|---|
| `user` | `profile/user.go` | 用户主表。 |
| `role` | `profile/role.go` | 角色主表。 |
| `user_role` | `profile/role.go` | 用户与角色关系表。 |
| `team` | `profile/team.go` | 团队主表。 |
| `user_team` | `profile/team.go` | 用户与团队关系表。 |
| `feature` | `profile/feature.go` | 功能权限表。 |
| 菜单/路由/映射 | `init_menu_item.go`, `init_router.go`, `init_feature_mapping.go` | 权限导航元数据由 repository/database 初始化维护。 |

## Validation And Rules

| Area | Implemented Rule |
|---|---|
| Username | `CreateUser` 会把 `username + "@apo.com"` 交给 `mail.ParseAddress`，用户名格式非法会失败。 |
| Password Complexity | 后端要求至少 8 位，且必须包含大写、小写、数字和特殊字符。 |
| Default Role | 创建用户时如果未传 `RoleList`，会尝试默认分配 `admin` 角色。 |
| Team Exists | 创建用户、创建团队、更新团队等流程都会检查团队存在性。 |
| Role Exists | 创建用户和更新角色时都会校验目标角色存在。 |
| Role Delete | 如果角色仍被授予给用户，`DeleteRole` 会直接拒绝删除。 |
| Router Permission | `CheckRouterPermission` 先放行白名单，再按 feature 到菜单/路由映射匹配。 |
| Menu Construction | `GetUserConfig` 只把用户具备 feature 映射的 menuItem 加入结果树。 |

## Error Handling

| Area | Behavior |
|---|---|
| Login Failure | `dbRepo.Login` 失败时直接返回错误。 |
| Role Duplicate | 创建/更新角色时重名会返回 `RoleExistsError`。 |
| Team Duplicate | 创建/更新团队时重名会返回 `TeamAlreadyExistError`。 |
| Missing User | 创建角色、创建团队、删除用户等流程都会对用户不存在返回专门错误。 |
| Missing Feature | 角色/团队使用不存在的 feature 时返回 `PermissionNotExistError`。 |
| Dify Sync Failure | 创建用户、删用户、改密、重置密码任一步 Dify 失败都会让事务失败。 |

## Notes

- `GetUserConfig` 不是单纯查用户自己的 feature，而是把 role、team、user 直授权限三层合并后再映射菜单和 routes。
- `CheckRouterPermission` 的白名单逻辑是字符串匹配，不依赖数据库配置。
- `handler.go` 中 `GetUserTeam` 的 swagger 注释写 `[post]`，而 `router_api.go` 实际注册为 `GET`，文档时要以真实注册为准。
- 团队更新逻辑里的 feature/data group 更新块被注释，说明团队管理与数据授权已经部分拆分到其他页面。
