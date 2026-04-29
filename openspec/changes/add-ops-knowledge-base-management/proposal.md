## Why

当前 APO 没有面向用户的运维知识库管理入口，用户无法在 APO 内上传、查看、删除 Runbook、SOP、排障方法等运维知识文档。与此同时，RAG 管理能力已经由外部 `rag-management` 服务承载，APO 需要作为前端消费方接入这些管理接口，并保持 AgentSRE 只读消费 RAG 检索的边界。

## What Changes

- 在 APO 商业前端新增“知识库管理”页面，用于展示知识库列表、文档列表、上传文档、下载原文件、删除文档、查看导入状态和失败原因。
- 新增 APO 前端 RAG 管理 API wrapper，统一调用同源路径 `/rag/api/kb/**`，并由开发态代理转发到 `http://192.168.1.69:5003/api/kb/**`。
- 按 `rag-management` 的 `docs/api/apo-knowledge-base-api.md` 作为管理接口契约，接入知识库列表、知识库详情、文档列表、上传前预览、上传并索引、文档导入状态、原文件下载和删除文档接口。
- 页面只向用户展示 RAG 明确返回 `writable === true` 的知识库；`support`、`intention_examples` 以及未声明可写的知识库不展示给用户管理。
- 使用 RAG 返回的 `writable` 控制知识库可见性和上传、预览、删除入口；不再对 `syncause_default` 做字段缺失兜底。
- 使用 `index_status` 展示文档导入状态，一期不引入 `job_id`、任务进度百分比或独立 jobs API。
- 一期上传格式限制为 `.txt`、`.md`、`.docx`、`.csv`、`.xls`、`.xlsx`，不支持 PDF。
- APO `/agent` 页面继续只消费 AgentSRE 返回结果，不在前端做关键词路由，不直接调用 RAG 检索接口。
- 不在 APO Go 后端新增知识库管理 API，不让 AgentSRE 成为知识库管理代理层。

## Capabilities

### New Capabilities

- `ops-knowledge-base-management`: 定义 APO 商业前端知识库管理页面、RAG 管理接口消费、文档列表/上传/删除/导入状态、可写性控制、同源代理和 AgentSRE/RAG 边界。

### Modified Capabilities

None.

## Impact

- Code:
  - `frontend/src/pro/views/knowledgeBases/**`
  - `frontend/src/pro/api/knowledgeBase.ts`
  - `frontend/src/pro/routes.js`
  - `frontend/vite.config.mjs`
  - 可能的 `frontend/public/locales/*/pro/**`
  - `frontend/src/_nav.js` 和后端菜单/权限初始化数据，用于把 `/knowledge-bases` 接入顶层主导航；本次按用户确认只落当前商业仓。
- APIs:
  - `GET /rag/api/kb/knowledge_bases`
  - `GET /rag/api/kb/knowledge_bases/{kb_id}`
  - `GET /rag/api/kb/knowledge_bases/{kb_id}/documents`
  - `POST /rag/api/kb/knowledge_bases/{kb_id}/documents/preview`
  - `POST /rag/api/kb/knowledge_bases/{kb_id}/documents`
  - `GET /rag/api/kb/documents/{doc_id}/index_status`
  - `GET /rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/download`
  - `DELETE /rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}`
- Systems:
  - APO 商业版知识库管理页面
  - 外部 `rag-management` 知识库管理服务
  - AgentSRE `/agent` 问答消费链路保持现状，由 AgentSRE 自身负责 `ops_knowledge -> syncause_default` 检索路由
- Dependencies:
  - `rag-management` 管理接口契约：`/Users/keting/workspace/gitlabProject/rag-management/docs/api/apo-knowledge-base-api.md`
  - 本地开发代理目标：`http://192.168.1.69:5003`
- Governance:
  - 页面主体与 API wrapper 属于 `frontend/src/pro/**` 商业前端范围。
  - `frontend/vite.config.mjs`、`frontend/src/_nav.js`、`frontend/src/i18n.js` 属于共享前端变体路径；实现前需要按仓库治理明确是否需要同步 OSS `feat/agent-ce`。
  - 如果实现触及 APO Go 后端菜单、路由、权限种子，则属于共享后端候选，需要先确认是否回 OSS 以及目标分支。
- Explicitly out of scope:
  - APO Go 后端 RAG 管理代理。
  - 文档解析、切片、embedding、向量索引。
  - APO 前端意图识别或 RAG 检索路由。
  - PDF 上传。
  - 上传人、来源类型、进度百分比展示。
  - 多租户知识库隔离、文档重建索引、引用溯源、metadata filter、批量删除。
