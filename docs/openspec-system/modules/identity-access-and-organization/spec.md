# Identity Access And Organization Spec

本文件保留为模块索引页。

当前已基于现有代码补齐以下详细文档：

- `source-map.md`
- `module_overview.md`
- `frontend.md`
- `backend.md`
- `flows.md`
- `dependencies.md`
- `open_questions.md`

后续如需继续扩写，应优先以这些文件中的现有代码证据为准，尤其需要持续区分：

- 本仓库内的 `user` / `permission` / `role` / `team` 主后端实现
- 通过 `UserContext` 驱动的菜单与路由可见性
- 前端已存在但主页面未使用的 `configMenuApi`
- 登录和用户管理链路里对 Dify 的外部同步边界
