## 1. Scope And Governance

- [x] 1.1 确认本 change 只实现 APO 商业前端知识库管理页面、`/rag` 本地代理和前端 API wrapper，不新增 APO Go 后端知识库管理 API。
- [x] 1.2 确认本期是否进入主导航；如果进入，先按共享前端和共享后端治理确认 `_nav.js`、菜单权限种子和 OSS 同步策略。
- [x] 1.3 确认本期是否需要 license gating；如果需要，明确前端隐藏、后端菜单权限和 license 中间件的责任边界。
- [ ] 1.4 确认生产环境 `/rag/api/kb/**` 转发链路由部署、网关、Ingress 还是其他仓库负责，并记录 rollout 前置条件。

## 2. RAG API Wrapper And Proxy

- [x] 2.1 在 `frontend/src/pro/api/knowledgeBase.ts` 新增知识库列表、详情、文档列表、上传前预览、上传并索引、导入状态查询和删除文档 wrapper。
- [x] 2.2 为 RAG 管理响应补充 TypeScript 类型和轻量归一化逻辑，覆盖 `writable`、`index_status`、`error_message`、`failed_step` 等字段。
- [x] 2.3 在前端错误处理工具或 wrapper 内统一错误提示优先级：`error_message > error > message > 默认文案`。
- [x] 2.4 在 `frontend/vite.config.mjs` 增加 `/rag` 开发代理，target 为 `http://192.168.1.69:5003`，并 rewrite 掉 `/rag` 前缀。
- [x] 2.5 验证前端业务代码只请求 `/rag/api/kb/**`，不直接硬编码 RAG 真实地址。

## 3. Knowledge Base Management Page

- [x] 3.1 在 `frontend/src/pro/views/knowledgeBases/**` 新增知识库管理页面骨架，并在 `frontend/src/pro/routes.js` 注册 `/knowledge-bases` route。
- [x] 3.1.1 在当前商业仓补齐 `/knowledge-bases` 顶层菜单、后端 router/feature/menu 权限初始化和菜单图标。
- [x] 3.2 实现知识库列表加载，只向用户展示 `writable === true` 的知识库，并在 `syncause_default` 可写时优先默认选中；知识库选择器只显示 `name`，不拼接 ID。
- [x] 3.3 实现文档列表表格，展示文档名称、知识库名称、文件类型、文件大小、上传时间、文档状态、导入状态、失败原因和操作列。
- [x] 3.4 实现分页、服务端文件名搜索、文档 `status` 筛选和手动刷新，避免使用当前页本地过滤作为正式搜索。
- [x] 3.5 实现空状态、列表加载失败、知识库加载失败和刷新状态。

## 4. Upload, Preview, Index Status, Delete

- [x] 4.1 实现上传入口，只允许可写知识库上传，并限制一期格式为 `.txt`、`.md`、`.docx`、`.csv`、`.xls`、`.xlsx`，明确阻止 PDF。
- [x] 4.2 实现上传大小和数量校验：单文件不超过 20 MB，单次最多 5 个文件；多文件上传按 RAG 单文件接口逐个提交或按页面约束收敛为单文件提交。
- [x] 4.3 实现上传前预览切片入口，使用 `POST /rag/api/kb/knowledge_bases/{kb_id}/documents/preview`，并只在可写知识库展示。
- [x] 4.4 实现正式上传并索引，使用 `multipart/form-data` 和固定字段 `file`，上传成功后展示 `doc_id`、`index_status` 并刷新列表。
- [x] 4.5 实现 `index_status` 映射和有界轮询：`pending`、`indexing`、`indexed`、`failed`，UI 分别展示为待处理、处理中、可用、处理失败，终态停止轮询；单行手动刷新入口放在导入状态列右侧。
- [x] 4.6 实现处理失败原因展示；上游未返回原因时展示本地兜底文案。
- [x] 4.7 实现删除文档二次确认、成功刷新、失败回滚；当返回 `failed_step` 时提示删除未完成且不直接移除列表项。
- [x] 4.8 实现原文件下载入口，使用显式请求调用 `GET /rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/download` 并下载 blob，不在 APO 前端拼接 RAG 文件存储路径。

## 5. AgentSRE Boundary And Existing Flow Safety

- [x] 5.1 确认 `/agent` 页面不新增前端关键词识别，不直接调用 RAG `/retrieval` 或 `/rag/api/kb/**/retrieve`。
- [x] 5.2 验证知识库管理页面新增后，`/agent` 仍按现有 AG-UI / AgentSRE 链路发送和展示消息。
- [x] 5.3 验证 RCA、resource、findings、general 等实时诊断问题不因本页面新增而改变前端调用链。

## 6. Tests And Verification

- [x] 6.1 补充 API wrapper 单元测试或等价验证，覆盖知识库列表、文档列表、上传成功、上传失败、删除失败和导入状态响应归一化。
- [x] 6.2 补充 `index_status` 导入状态映射和轮询终止条件测试。
- [x] 6.3 补充上传格式、PDF 拦截、大小限制和数量限制测试。
- [x] 6.4 补充删除确认、删除失败保留列表状态和 `failed_step` 展示测试。
- [x] 6.5 补充页面空状态、加载失败和仅展示可写知识库的交互验证。
- [x] 6.6 补充原文件下载 URL、操作入口和导入状态列刷新入口验证。
- [x] 6.7 运行前端构建或项目约定的最小验证命令，确认新增页面和 proxy 配置不会破坏现有编译。

## 7. Documentation And Rollout

- [x] 7.1 实现完成后新增或更新 `openspec/changes/add-ops-knowledge-base-management/implementation-summary.md`，记录实际落地范围、偏差、兼容策略和 follow-up。
- [x] 7.2 如果实现触及主导航、权限菜单、license gating 或生产代理配置，同步更新对应 `docs/openspec-system/**` 文档。
- [x] 7.3 如果实现不触及系统级文档，必须在 implementation summary 中说明豁免原因。
- [ ] 7.4 在上线前确认生产 `/rag/api/kb/**` 代理可用，并记录未解决的部署或权限 follow-up。
