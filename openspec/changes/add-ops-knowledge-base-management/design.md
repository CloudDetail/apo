## Context

APO 当前商业前端已经通过 `/agent` 页面消费外部 AgentSRE，但没有独立的知识库管理页面。用户无法在 APO 内维护运维知识文档，`syncause_default` 中的 Runbook、SOP、排障方法等内容只能依赖外部系统管理。

本次需求的职责边界已经收敛为：

```text
APO：提供知识库管理页面和前端交互。
rag-management：提供真正的知识库管理、文档解析、切片、embedding、索引和检索能力。
AgentSRE：只读消费 RAG 检索结果，不参与知识库管理。
```

当前 `rag-management` 已提供面向 APO 的管理接口文档：

```text
rag-management: docs/api/apo-knowledge-base-api.md
```

APO 前端应按该文档的 `/api/kb/**` 接口调用，但在同源路径前加 `/rag` 前缀。开发态代理目标为 `http://192.168.1.69:5003`，即：

```text
/rag/api/kb/** -> http://192.168.1.69:5003/api/kb/**
```

当前仓库的路由和菜单存在两个重要约束：

- 本地 route 渲染由 `frontend/src/routes.js` 和 `frontend/src/pro/routes.js` 决定。
- 可见菜单与默认入口依赖后端 `/api/permission/config` 返回的菜单/路由元数据，而不是只看前端本地 route。

因此，本 change 可以先实现 `/knowledge-bases` 直达页面与本地开发代理；是否进入主导航、是否需要 license gating、生产环境 `/rag` 代理由哪个部署仓承载，需要作为 rollout 决策单独确认。

## Goals / Non-Goals

**Goals:**

- 新增 APO 商业前端“知识库管理”页面，覆盖知识库列表、文档列表、上传、下载原文件、删除、导入状态和失败原因展示。
- 新增前端 RAG 管理 API wrapper，统一消费 `/rag/api/kb/**`。
- 本地 Vite 开发态新增 `/rag` 代理，并 rewrite 掉 `/rag` 前缀。
- 一期只允许写入 `syncause_default`，并用 RAG 返回的 `writable` 控制上传、预览和删除入口。
- 一期使用 `index_status` 离散状态，不使用 `job_id`、jobs API 或进度百分比。
- 继续保持 `/agent` 前端只消费 AgentSRE，不在 APO 前端做关键词路由或 RAG 检索。

**Non-Goals:**

- 不在 APO Go 后端新增 RAG 管理代理接口。
- 不在 APO 内实现文档解析、切片、embedding、向量索引或向量数据库连接。
- 不在 APO 前端实现 AgentSRE 意图识别或问答路由。
- 不支持 PDF 上传。
- 不展示上传人、来源类型或索引进度百分比。
- 不实现多租户知识库隔离、引用溯源、metadata filter、重建索引、批量删除或审计日志。

## Decisions

### Decision 1: APO 前端只通过 `/rag/api/kb/**` 同源路径消费 RAG 管理接口

前端 API wrapper 不直接写外部 RAG 真实地址，只请求 `/rag/api/kb/**`。开发态由 Vite proxy 转发到 `http://192.168.1.69:5003` 并去掉 `/rag` 前缀。

这样可以保持浏览器侧同源调用，也避免把环境地址写入业务代码。生产环境仍需要由部署或网关层补齐等价转发。

Alternatives considered:

- 前端直接调用 `http://192.168.1.69:5003`：开发最快，但会把环境地址泄露进前端，并引入跨域和部署差异。
- APO Go 后端做代理：鉴权和审计更集中，但会把知识库管理职责拉回 APO 后端，违背当前边界，并触发后端共享包治理问题。

### Decision 2: RAG 管理接口契约以 rag-management 文档为 owner

APO 只作为 consumer 适配以下接口：

```http
GET    /api/kb/knowledge_bases
GET    /api/kb/knowledge_bases/{kb_id}
GET    /api/kb/knowledge_bases/{kb_id}/documents
POST   /api/kb/knowledge_bases/{kb_id}/documents/preview
POST   /api/kb/knowledge_bases/{kb_id}/documents
GET    /api/kb/documents/{doc_id}/index_status
GET    /api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/download
DELETE /api/kb/knowledge_bases/{kb_id}/documents/{doc_id}
```

APO wrapper 负责 TypeScript 类型、响应归一化、错误消息提取和页面状态映射，但不重新定义上游 contract。

Alternatives considered:

- 在 APO change 内重新设计一套 `/knowledge-bases/**` API：会制造双 owner 契约，也容易和 rag-management 实际接口漂移。
- 直接在组件内散落 fetch 调用：实现少一层，但会让错误处理、字段映射和后续接口变化难以集中维护。

### Decision 3: 一期只展示 `writable === true` 的知识库

知识库列表和详情接口返回 `writable`。页面知识库选择器只展示 `writable === true` 的知识库，上传、预览和删除入口也只在当前知识库明确可写时展示；如果上游缺少 `writable` 字段，前端不将该知识库默认视为可管理。

这样可以把“哪些知识库可管理”的最终权威留给 RAG 服务，避免把 `support`、`intention_examples` 或字段缺失的知识库暴露给用户。

Alternatives considered:

- 在 APO 前端硬编码只有 `syncause_default` 可写：实现简单，但后续 RAG 调整可写性时需要改前端。
- 所有知识库都显示上传入口，依赖后端 403：安全上可行，但用户体验差。
- 对 `syncause_default` 做字段缺失兜底：可兼容旧接口，但会让可见性规则不够明确。

### Decision 4: 文档导入状态采用 `index_status` 离散状态模型

一期不引入 `job_id`、`/jobs/{job_id}` 或进度百分比。页面内部仍消费 RAG 的 `index_status` 字段，但 UI 展示为用户更容易理解的“导入状态”：

| index_status | 页面语义 |
|---|---|
| `pending` | 待处理 |
| `indexing` | 处理中 |
| `indexed` | 可用 |
| `failed` | 处理失败 |

上传返回 `indexing` 时，前端可轮询 `GET /rag/api/kb/documents/{doc_id}/index_status`，直到 `indexed` 或 `failed`，并设置轮询上限。

Alternatives considered:

- 引入 job 模型：更适合复杂异步任务，但当前 RAG API 已收敛为 doc-level index status。
- 展示百分比：体验更细，但上游 contract 没有稳定 progress 字段。

### Decision 5: 页面主体放在 `frontend/src/pro/**`，避免扩散到共享路径

页面、组件、类型和 API wrapper 应优先落在：

```text
frontend/src/pro/views/knowledgeBases/**
frontend/src/pro/api/knowledgeBase.ts
```

本地 route 可在 `frontend/src/pro/routes.js` 注册 `/knowledge-bases`。如果需要新增独立 i18n namespace，可能触及 `frontend/src/i18n.js` 的 `proNsList`；实现时可以选择复用现有 pro namespace 降低共享路径改动，也可以按共享前端治理先处理 OSS 同步。

Alternatives considered:

- 把页面放入 `frontend/src/core/**`：不符合商业特性归属，也会触发共享前端同步。
- 直接改主 `_nav.js` 和后端菜单种子作为第一步：能让入口立即可见，但会放大共享路径和权限治理范围。

### Decision 6: 主导航与权限菜单已按当前仓顶层菜单落地

一期实现先保证 `/knowledge-bases` 本地 route 在 pro 构建中可渲染。随后已确认知识库管理进入侧边栏顶层主导航，并按当前商业仓本地变更落地，不先回 OSS。

主导航落地约定：

- 菜单 key 使用 `knowledgeBases`，路由使用 `/knowledge-bases`。
- 菜单为顶层菜单，不挂到其他父级下。
- 后端初始化补齐 `feature`、`router`、`menu`、`feature_mapping`、默认 admin feature 权限和 i18n。
- 普通 React 页面不新增 `router_insert_page`。
- 本期不增加 license gating；后续如产品要求，再由前端隐藏、后端菜单权限或 license 中间件单独承载。

Alternatives considered:

- 开发前必须先定完主导航和后端权限：治理更完整，但会阻塞页面与 RAG API 的独立开发验证。
- 完全不考虑菜单权限：直达可用，但上线后用户可能无法发现入口。
- 只挂到系统管理等父菜单下：权限层次更收敛，但当前用户明确要求顶层菜单。

### Decision 7: `/agent` 问答路由仍由 AgentSRE 负责

APO `/agent` 页面不识别关键词，不直接请求 RAG 检索接口，不把 `syncause_default` 写成前端业务意图。AgentSRE 后续应在自身仓库中完成 `ops_knowledge -> syncause_default`、`support -> support` 和 `intention_examples` 辅助识别的配置与实现。

APO 只要求现有 AG-UI/AgentSRE 消息链路在知识库页面新增后不回归，并能正常展示 AgentSRE 返回的 assistant 消息或错误降级。

Alternatives considered:

- 在 APO 前端根据关键词直接调用 RAG：会绕过 AgentSRE 的 SRE 分析和意图识别边界。
- 把 RAG 管理和检索都代理到 APO 后端：边界过重，不符合当前系统划分。

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| 本地 `/rag` 可用但生产环境未配置等价转发 | 将生产 `/rag/api/kb/**` 转发归属保留为 rollout 前置确认；页面实现阶段只承诺本地 Vite proxy |
| 页面 route 存在但侧边栏不可见 | 将主导航和后端权限菜单作为单独任务和开放问题；如果产品要求可见入口，则在实现前确认共享前端/共享后端治理 |
| RAG 接口字段继续变化 | API wrapper 集中做响应归一化，并以 rag-management 文档为 contract owner |
| 上传大文件或多文件导致前端状态混乱 | 一期限制格式、20 MB 单文件上限和最多 5 个文件；上传后刷新列表或按 doc_id 局部更新 |
| 删除接口出现 `failed_step` 部分失败 | 前端提示“删除未完成”，不直接移除列表项，刷新后再以服务端状态为准 |
| `writable` 缺失或错误导致入口展示不一致 | 前端只展示 `writable === true` 的知识库；RAG 服务仍负责最终权限校验 |
| AgentSRE ops_knowledge 路由未同步完成 | APO 知识库管理页面仍可独立上线；`/agent` 运维知识问答命中率作为跨仓联调项，不由 APO 前端硬编码补偿 |

## Migration Plan

1. 先实现前端 API wrapper、类型和状态映射，并用 mock 或真实 RAG dev 服务验证契约。
2. 新增 `/knowledge-bases` 页面和 pro route，保证直达路由可用。
3. 增加 Vite `/rag` 开发代理，验证本地管理接口闭环。
4. 接入文档列表、上传、删除、导入状态查询和错误处理。
5. 决定是否进入主导航；如果进入，再处理菜单、权限和可能的 license gating。
6. 生产上线前确认部署层 `/rag/api/kb/**` 转发链路。

Rollback strategy:

- 前端页面可以通过撤销 route/menu 入口隐藏；若仅直达 route 已合入，移除菜单即可降低用户可见影响。
- `/rag` proxy 是新增路径，不影响现有 `/api` 和 `/ai/api`。
- 不改 APO Go 后端和 AgentSRE 管理能力，因此回滚不涉及数据库迁移。

## Open Questions

1. 生产环境 `/rag/api/kb/**` 转发链路由哪个仓或部署配置负责。
2. 知识库管理页面是否进入主导航；如果进入，菜单 key、图标、排序和权限归属如何定义。
3. 知识库管理是否需要 license gating。
4. 文档删除权限最终由 RAG 按管理员、团队还是全局知识库管理员控制。
5. `support` 一期是只读展示还是完全不向普通用户展示。
6. `intention_examples` 一期是只读展示还是对普通用户隐藏。
