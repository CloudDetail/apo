# Identity Access And Organization Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `/api/permission/menu/configure` 与当前“通过更新角色 feature 来影响菜单”的模式是否同时有效，代码里没有完整前端消费证据。
2. 团队在权限模型中的定位是“只承载成员关系”，还是也应承载 feature 和数据组授权，当前实现呈现出过渡态。

## Legacy Or Dead-Code Suspicion

1. `AuthRouter` 中真正的路由权限检查逻辑被注释，说明前端权限拦截可能被临时关闭。
2. `permission.js` 中 `getAllRoleList` 指向 `/api/permission/roles`，但当前后端未见对应 route，疑似旧 wrapper。
3. `MenuManage` 页名与其真实行为不完全一致，它更像“角色权限配置”的另一种视图。

## Missing Frontend-Backend Alignment

1. 个人中心实际使用 `/api/user/update/info`，而不是 `/api/user/update/self`。
2. `handler.go` 中 `GetUserTeam` swagger 标注为 `[post]`，实际 router 注册为 `GET`。
3. 前端密码校验规则和后端复杂度校验规则不一致。

## Permission Uncertainty

1. 前端当前并未真正执行 route permission 检查，导致页面可达性与后端权限配置可能脱节。
2. `CheckRouterPermission` 的硬编码白名单是否只是临时过渡，需要确认。

## Data Model Uncertainty

1. 角色、用户、团队与 feature 之间的所有关系表由 repository/database 维护，当前模块文档只确认了主实体表和初始化元数据，未展开全部关系表实现。
2. 团队更新流程中 feature/data group 更新逻辑被注释，说明数据模型的写入入口可能已迁移，需要后续确认。
