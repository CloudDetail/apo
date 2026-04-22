# Identity Access And Organization Flows

## Main Flow 1

### Login And Bootstrap Flow

1. 用户访问系统，若本地没有有效 token，主应用跳到 `/login`。
2. 用户在 `Login` 页面提交用户名和密码。
3. 前端调用 `/api/user/login`。
4. 后端校验账号密码并返回 `accessToken`、`refreshToken` 与 `user`。
5. 前端保存 `token` 和 `refreshToken` 到 `localStorage`。
6. 如果 `deployMode !== 'syncause'`，前端继续调用 `/dify/console/api/login/apo` 获取 `difyToken`。
7. 登录完成后进入首页，`UserContext` 调 `/api/permission/config` 获取菜单和 routes。
8. 前端根据返回结果渲染侧边栏和页面导航。

## Main Flow 2

### User Creation And External Sync Flow

1. 管理员进入 `/system/user-manage`。
2. 前端通过 `/api/user/list` 展示用户列表，通过 `/api/role/roles` 加载角色选项。
3. 管理员打开 `AddUserModal`，填写用户名、密码、角色等信息。
4. 前端调用 `/api/user/create`。
5. 后端校验用户名格式、密码复杂度、角色存在性、团队存在性。
6. 后端创建本地 `user` 记录。
7. 后端调用 `difyRepo.AddUser` 创建对应 Dify 用户。
8. 后端为用户授予角色、团队与默认数据组授权。
9. 前端刷新用户列表。

## Main Flow 3

### Role Permission Configuration Flow

1. 管理员进入 `/system/role-manage` 或 `/system/menu-manage`。
2. 页面调用 `/api/role/roles` 获取角色列表。
3. 选择某个角色后，`PermissionTree` 调 `/api/permission/feature` 获取完整 feature 树。
4. `PermissionTree` 再调 `/api/permission/sub/feature` 获取当前角色已有 feature。
5. 用户勾选权限后保存。
6. 前端调用 `/api/role/update`，传 `roleId`、`roleName`、`permissionList`。
7. 后端计算需要新增和删除的 feature 授权，并事务更新。
8. 如果是 `MenuManage` 页面，还会触发 `getUserPermission()` 刷新当前用户菜单。

## Main Flow 4

### Password Update Flow

1. 用户进入 `/user` 页面。
2. `UpdatePassword` 表单校验旧密码、新密码和确认密码。
3. 前端调用 `/api/user/update/password`。
4. 后端校验密码复杂度，更新本地用户密码。
5. 后端调用 `difyRepo.UpdatePassword` 同步 Dify 密码。
6. 前端随后调用 `/api/user/logout`，清理本地 token 并跳转 `/login`。

## Edge Cases

| Case | Implemented Behavior |
|---|---|
| 新建用户未传角色 | 后端尝试自动分配 `admin`。 |
| 删除仍被授予的角色 | 后端返回 `RoleGrantedError`，拒绝删除。 |
| 团队更新时只改成员 | 当前 service 稳定支持；feature/data group 更新逻辑已注释。 |
| 登录为 `syncause` 部署模式 | 前端不会调用 Dify 登录。 |
| 用户密码修改成功 | 前端会强制重新登录。 |
| `UserContext` 语言切换 | 会重新调用 `/api/permission/config` 以刷新菜单文案。 |

## Failure Paths

| Path | Failure Behavior |
|---|---|
| 用户名格式不合法 | `/api/user/create` 返回 `UserNameError`。 |
| 密码复杂度不足 | 创建、修改或重置密码时返回 `UserPasswdSimpleError`。 |
| Dify 同步失败 | 创建/删用户/改密流程整体失败。 |
| feature 不存在 | 角色或团队相关接口返回 `PermissionNotExistError`。 |
| team 不存在 | 用户或团队操作返回 `TeamNotExistError`。 |
| `AuthRouter` 权限检查 | 当前前端不会真正执行，因此失败路径只存在于预留代码。 |

## Permission Constraints

| Layer | Constraint |
|---|---|
| Frontend Route | 虽然经 `AuthRouter` 包装，但当前不会实际阻止访问。 |
| Menu Visibility | 真正的导航可见性由 `/api/permission/config` 返回结果决定。 |
| Backend Middleware | 用户信息接口需登录；修改型接口通常还叠加 `LicMiddleware()`。 |
| Router Allowlist | `/agent-dataplane`、`agent`、`probe-installer` 永久放行。 |

## To Be Confirmed

- `/api/permission/menu/configure` 仍然存在，但当前主页面看起来已经转向“通过更新角色权限来影响菜单”，两种模式的真实分工需要确认。
- `AuthRouter` 被注释后，前端与后端权限边界谁负责最终兜底，需要确认。
- 个人中心为何实际调用 `/api/user/update/info` 而不是 `/api/user/update/self`，需要确认是复用还是旧接口残留。
