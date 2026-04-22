# Dataplane And Topology Governance Spec

本文件保留为模块索引页。

当前已基于现有代码补齐以下详细文档：

- `source-map.md`
- `module_overview.md`
- `frontend.md`
- `backend.md`
- `flows.md`
- `dependencies.md`
- `open_questions.md`

后续如需继续扩写，应优先以这些文件中的现有代码证据为准，尤其要把以下边界分开描述：

- 已实现的 `/api/dataplane/*` 主后端接口
- 已实现的 `/api/dataplane/capabilities`，以及它只聚合已配置 dataplane capability 的边界
- 已存在 wrapper 但尚未进入主页面流程的 `customtopology` 相关能力
- 历史权限元数据中的 `/agent-dataplane` 与当前主路由 `/dataplane` 的差异
