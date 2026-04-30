## Final Scope

本次实现完成 APO 商业前端侧的知识库管理最小闭环：

- 新增 `/knowledge-bases` 直达页面，用于管理 RAG 知识库文档。
- 新增前端 RAG 管理 API wrapper，统一请求 `/rag/api/kb/**`。
- 新增 Vite 开发态 `/rag` proxy，转发到 `http://192.168.1.69:5003` 并去掉 `/rag` 前缀。
- 支持可写知识库列表、文档列表、服务端分页、文件名搜索、文档状态筛选、上传前预览滚动分页、已上传文档切片滚动分页预览、上传并索引、导入状态轮询和单行手动刷新、原文件下载、删除确认和失败原因展示。
- 新增 `knowledgeBases` 顶层侧边栏菜单和 `/knowledge-bases` 后端 router/feature/menu 权限初始化；新库默认授予 admin 角色。
- 一期上传格式限制为 `.txt`、`.md`、`.docx`、`.csv`、`.xls`、`.xlsx`，不支持 PDF。
- 保持 APO Go 后端和 `/agent` AgentSRE 消费链路不变；APO 前端不做 RAG 检索路由。

未在本次实现中落地：

- license gating。
- 生产环境 `/rag/api/kb/**` 转发配置。
- 上传人、来源类型、进度百分比、jobs API、批量删除、审计日志、租户隔离。

## Actual Code Changes

- `frontend/src/pro/api/knowledgeBase.ts`
  - 新增知识库和文档类型、RAG 响应归一化、上传校验、导入状态映射、错误消息提取和 multipart `FormData` 构造。
  - 新增可写知识库过滤逻辑；页面只展示 RAG 明确返回 `writable === true` 的知识库。
  - 新增原文件下载 URL 构造和下载文件名解析，统一走 `/rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/download`。
  - 新增已上传文档切片预览 URL、滚动分页查询参数构造和响应归一化，统一走 `/rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/preview`；预览响应只过滤纯空白 `nodes[]`，保留非空 chunk 的原始前后空白和缩进。
  - 新增上传前和已上传文档预览共用的 `limit`、`offset` 查询参数构造；上传和上传前预览 `FormData` 默认使用 `chunk_method=separator` 和 `chunk_params={"separator":["\n\n","\n",".","。","?","？","!","！"," "],"chunk_size":1024,"chunk_overlap":100}`。
- `frontend/src/pro/api/knowledgeBaseClient.ts`
  - 新增 RAG 管理接口 wrapper：知识库列表、知识库详情、文档列表、上传前预览、已上传文档切片预览、上传、导入状态、下载、删除。
  - 已上传文档切片预览 wrapper 支持 `limit` 和 `offset`，默认按 `limit=50&offset=0` 请求，并将 `limit` 限制在 RAG 契约允许的 `0..200` 范围内。
  - 上传前预览 wrapper 支持 `limit` 和 `offset`，通过 `POST /rag/api/kb/knowledge_bases/{kb_id}/documents/preview?limit=50&offset=0` 提交当前文件，并复用 preview 响应归一化。
- `frontend/src/pro/api/knowledgeBase.test.ts`
  - 覆盖列表归一化、服务端查询参数、导入状态映射、上传格式/大小/数量限制、删除失败错误消息提取。
  - 覆盖可写知识库过滤、知识库名称展示、原文件下载 URL、文件名解析、页面下载操作入口和导入状态列刷新入口。
  - 覆盖已上传文档切片预览 URL、滚动分页查询参数、响应归一化、操作列预览入口、上传弹窗逐文件预览、去除固定本地截断，以及切片预览使用受控 Modal 继承当前主题而不是静态 `Modal.info()`。
  - 覆盖上传和上传前预览 multipart 默认切片字段、分页查询参数、响应归一化、过期预览请求失效、预览 chunk 保留原始空白，以及滚动分页时重新提交同一个待上传文件。
- `frontend/src/pro/views/knowledgeBases/index.tsx`
  - 新增知识库管理页面，包含可写知识库选择、文档表格、搜索筛选、刷新、上传弹窗、预览、轮询和删除确认。
  - 知识库选择器和文档列表知识库列只展示知识库 `name`；`id`/`kb_id` 仅用于请求和前端映射，不直接展示给用户。
  - 移除操作栏中的“状态”按钮，将单行导入状态刷新入口收敛到“导入状态”列右侧刷新图标。
  - 将上传前预览从底部全局按钮改为每个待上传文件项自己的“预览”按钮，避免多文件时默认只预览第一个文件。
  - 上传前预览首次请求 `limit=50&offset=0`；预览弹窗滚动到底部或内容不足以撑出滚动条但仍有更多切片时，使用同一个 `File` 重新 POST 下一页并追加切片，适配 RAG 不保存上传前预览状态的契约。
  - 在文档操作列新增“预览”按钮，请求已上传文档切片预览接口，并在受控 AntD Modal 中展示返回切片；弹窗滚动到底部或内容不足以撑出滚动条但仍有更多切片时，按 RAG 返回的 `offset + limit` 推进 `offset` 并追加后续切片，不再固定截断前 20 条。
  - 为预览请求增加序号失效机制，用户切换预览对象或关闭弹窗后，未完成的旧请求结果和错误不会追加或覆盖当前预览状态。
  - 在文档操作列新增“下载”按钮，用显式请求获取 blob 后触发浏览器保存，并延迟释放 object URL；失败时展示下载失败提示，避免静态下载链接直接保存错误响应。
  - 收敛页面布局：标题下不再展示知识库描述；知识库选择、文档搜索和状态筛选放在同一工具栏左侧，刷新和上传文档放在同一行右侧。
- `frontend/src/pro/views/knowledgeBases/index.module.scss`
  - 新增页面布局样式，并调整工具栏在桌面和移动端的排列。
- `frontend/src/pro/routes.js`
  - 注册 `/knowledge-bases` pro route。
- `frontend/vite.config.mjs`
  - 新增开发态 `/rag` proxy；同时保留当前工作树中的 `/api`、`/ai/api` 开发代理配置。
- `backend/pkg/repository/database/init_feature.go`
  - 新增 `知识库管理` feature，并同时加入默认产品模式和 `PRODUCT_MODE=syncause` 的 feature 白名单。
- `backend/pkg/repository/database/init_router.go`
  - 新增 `/knowledge-bases` router，`HideTimeSelector` 为 `true`，并同时加入默认产品模式和 `syncause` router 白名单。
- `backend/pkg/repository/database/init_menu_item.go`
  - 新增 `knowledgeBases` 顶层菜单，排序为 `2`，指向 `/knowledge-bases`。
- `backend/pkg/repository/database/init_feature_mapping.go`
  - 新增 `知识库管理 -> knowledgeBases` feature-to-menu 映射。
- `backend/pkg/repository/database/init_permission.go`
  - 将 `知识库管理` 加入 admin 默认 feature 权限；不默认授予 viewer 或 anonymous。
- `backend/sqlscripts/i18n.yml`
  - 新增 `knowledgeBases` 菜单 label/abbreviation 翻译和 `知识库管理` feature 翻译。
- `frontend/src/_nav.js`
  - 新增 `knowledgeBases` 顶层菜单图标。
- `backend/pkg/repository/database/validation_test.go`
  - 补充 knowledge-bases feature/router/menu seed validator 测试，先红后绿。

## Document Updates

Change docs:

- `openspec/changes/add-ops-knowledge-base-management/proposal.md`
- `openspec/changes/add-ops-knowledge-base-management/design.md`
- `openspec/changes/add-ops-knowledge-base-management/specs/ops-knowledge-base-management/spec.md`
- `openspec/changes/add-ops-knowledge-base-management/tasks.md`
- `openspec/changes/add-ops-knowledge-base-management/implementation-summary.md`

System docs:

- `docs/openspec-system/page_inventory.md`
- `docs/openspec-system/route_to_module_map.md`
- `docs/openspec-system/api_to_module_map.md`
- `docs/openspec-system/frontend_backend_mapping.md`
- `docs/openspec-system/architecture_notes.md`
- `docs/openspec-system/cross-cutting/permission-menu-router-model.md`

本次补充已上传文档切片预览时同步更新：

- `docs/openspec-system/api_to_module_map.md`
- `docs/openspec-system/frontend_backend_mapping.md`
- `docs/openspec-system/architecture_notes.md`

本次修复已上传文档切片滚动分页时同步更新：

- `openspec/changes/add-ops-knowledge-base-management/design.md`
- `openspec/changes/add-ops-knowledge-base-management/specs/ops-knowledge-base-management/spec.md`
- `openspec/changes/add-ops-knowledge-base-management/implementation-summary.md`

本次滚动分页修复没有更新 `docs/openspec-system/**`，原因是模块归属、跨仓边界和 `/rag/api/kb/**` 系统级 ownership 未变化；只是在既有外部 RAG preview 契约内补齐前端分页消费行为。

本次补充上传前预览分页契约时同步更新：

- `openspec/changes/add-ops-knowledge-base-management/proposal.md`
- `openspec/changes/add-ops-knowledge-base-management/design.md`
- `openspec/changes/add-ops-knowledge-base-management/specs/ops-knowledge-base-management/spec.md`
- `openspec/changes/add-ops-knowledge-base-management/tasks.md`
- `openspec/changes/add-ops-knowledge-base-management/implementation-summary.md`
- `docs/openspec-system/api_to_module_map.md`
- `docs/openspec-system/frontend_backend_mapping.md`
- `docs/openspec-system/architecture_notes.md`

## Design Deviations

- 主导航入口在初始页面实现后追加落地。用户已确认本次只落当前商业仓并作为顶层菜单，因此本次没有先回 OSS；该点是共享后端候选和共享前端路径的本地商业变更，需要后续按仓库治理决定是否补同步。
- 没有新增 APO Go 后端代理。原因是知识库管理能力归属 `rag-management`，APO 本次只作为前端 consumer。
- 没有新增 license gating。当前只通过菜单 feature 权限控制可见性，RAG 服务仍负责最终写权限。
- 没有展示上传人、来源类型和进度百分比。原因是一期 RAG 契约以文档级 `index_status` 为准，不引入 jobs/progress 模型。
- 已上传文档切片预览依赖 RAG 新增 `GET /api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/preview` 只读接口。APO 不通过下载原文件后重新上传预览来模拟切片结果，避免展示内容和实际索引切片漂移；该接口按 `limit` 和 `offset` 滚动分页读取，APO 不再使用本地固定条数截断替代分页，后续页 offset 按 RAG 返回的 `offset + limit` 推进。
- 上传和上传前预览显式使用相同默认切片参数：`chunk_method=separator`、`chunk_params={"separator":["\n\n","\n",".","。","?","？","!","！"," "],"chunk_size":1024,"chunk_overlap":100}`。
- 上传前预览依赖 RAG `POST /api/kb/knowledge_bases/{kb_id}/documents/preview` 接口支持 `limit` 和 `offset` 查询参数。APO 不保存本地预览状态，也不假设服务端保存预览状态；后续页使用同一个 `File` 和同一组切片参数重新发起 multipart 请求，offset 同样按 RAG 返回的 `offset + limit` 推进。
- `writable` 从写入口控制收敛为知识库可见性控制；RAG 未明确返回 `writable === true` 的知识库不会展示给用户管理，不再对 `syncause_default` 做缺省可写兜底。

## Compatibility And Rollout

- 前端业务代码只请求 `/rag/api/kb/**`，不直接硬编码 RAG 服务地址。
- 本地开发依赖 Vite `/rag` proxy：`/rag/api/kb/** -> http://192.168.1.69:5003/api/kb/**`。
- 已上传文档切片预览依赖 RAG 暴露 `GET /api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/preview`，前端通过同源 `/rag` 代理调用；APO 首次请求 `limit=50&offset=0`，弹窗滚动加载后续页。
- 上传和上传前预览依赖 RAG 接受 `chunk_method=separator` 和 `chunk_params={"separator":["\n\n","\n",".","。","?","？","!","！"," "],"chunk_size":1024,"chunk_overlap":100}`；上传前预览通过同源 `/rag` 代理提交 multipart 文件，首次请求 `limit=50&offset=0`，弹窗滚动加载或内容未撑满弹窗时重新带同一个 `file` 和同一组切片参数请求后续页。
- 生产上线前必须由部署、网关、Ingress 或其他承载方补齐等价 `/rag/api/kb/**` 转发。
- 新库默认 admin 角色会获得 `知识库管理` feature；已有环境中的角色如果已经存在任意 feature 权限，不会因 `init_permission.go` 自动追加该权限，需要通过迁移、菜单管理或手动授权补齐。
- 新增 `/knowledge-bases` route 不改变 `/agent`、RCA、resource、findings、general 等现有前端调用链。
- 回滚时可移除 `/knowledge-bases` route、`knowledgeBases` 菜单/权限种子和页面入口；`/rag` 是新增开发代理路径，不影响现有 `/api` 或 `/ai/api`。

## Follow-ups

- 确认生产环境 `/rag/api/kb/**` proxy 的 owner、配置仓库、鉴权方式和发布顺序。
- 确认已有部署环境如何给现有 admin 角色补授 `知识库管理` feature 权限。
- 如果后续要求跨仓一致，按共享后端候选和共享前端路径治理补 OSS 同步。
- 如果后续需要 license gating，再决定由前端隐藏、后端菜单权限还是 license 中间件承载。
- 与 `rag-management` 联调真实删除部分失败、索引失败、上传前预览分页和已上传文档切片预览响应。
- 与 AgentSRE 仓联调 `ops_knowledge -> syncause_default` 消费路由；APO 前端不在本仓补偿该逻辑。
