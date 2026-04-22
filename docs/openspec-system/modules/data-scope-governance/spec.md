# Data Scope Governance Spec

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

- 当前主页面实际使用的 `/api/v2/data/group*`
- 仍然保留但未进入主流程的旧版 `/api/data/group/create|update|delete`
- “数据组 -> 主体授权”和“主体 -> 数据组授权”两条方向不同的接口链
- `DataGroupStorage` 作为共享缓存而非纯页面私有状态的实现方式
