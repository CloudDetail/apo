# 运维知识库管理与 AgentSRE RAG 消费路由需求文档

## 一、文档目的

本文档用于把“运维知识库管理与 RAG 问答路由”需求重新收敛为清晰的跨系统职责、APO 侧需求范围、RAG 服务接口要求、AgentSRE 运行时消费边界和验收标准。

本文档的核心边界是：

```text
APO 负责知识库管理页面；
RAG 负责真正的知识库管理和检索能力；
AgentSRE 只负责消费 RAG 检索结果，不参与知识库管理。
```

因此，本需求不是让 AgentSRE 成为知识库管理系统，也不是让 APO Go 后端实现 RAG 管理代理，而是完成两件事：

```text
1. APO 前端通过同源转发路径调用 RAG 管理接口，管理 syncause_default 中的运维知识文档。
2. AgentSRE 在用户提问时，根据意图识别结果消费 RAG 检索接口，将运维知识类问题路由到 syncause_default 进行问答。
```

## 二、模块定位与总体意图

本需求涉及三个系统：APO、外部 RAG 服务、AgentSRE。

三者职责必须明确拆分：

| 系统 | 定位 | 核心职责 |
|---|---|---|
| APO | 用户操作入口 | 提供知识库管理页面，调用 RAG 管理接口，展示上传、列表、删除、索引状态和失败原因 |
| 外部 RAG 服务 | 知识库能力 owner | 提供知识库管理、文档处理、embedding、索引和检索能力 |
| AgentSRE | RAG 只读消费方 | 做意图识别、问答路由和 SRE 分析，按分类消费 RAG 检索接口 |

目标不是：

```text
让 AgentSRE 或 APO 后端具备把文档上传到外部 RAG 的能力。
```

目标是：

```text
让外部 RAG 提供知识库管理接口给 APO 调用；
让 AgentSRE 具备基于配置消费 syncause_default 的检索问答能力。
```

## 三、三个系统的职责边界

### 3.1 APO 的职责

APO 是用户操作入口，负责提供“知识库管理”页面。

APO 应该负责：

1. 展示知识库管理页面。
2. 展示知识库列表。
3. 展示文档列表。
4. 上传文档到 `syncause_default`。
5. 删除文档。
6. 查询文档索引状态。
7. 展示文档解析和索引的离散状态。
8. 展示失败原因。
9. 做基础前端交互校验，例如文件格式、文件大小和只读知识库隐藏操作。
10. 通过同源转发路径访问 RAG 管理接口，避免在前端硬编码 RAG 服务真实地址。

APO 不应该负责：

1. 文档解析。
2. 文档切片。
3. embedding。
4. 向量索引。
5. RAG 检索实现。
6. 意图识别。
7. 问答生成。
8. 知识库管理后端 API。
9. RAG 管理代理。

### 3.2 RAG 服务的职责

RAG 服务是知识库能力 owner，负责真正的知识库管理和检索能力。

RAG 应该负责：

1. 知识库创建或维护。
2. 知识库列表。
3. 文档上传。
4. 文档列表。
5. 文档删除。
6. 文档索引状态。
7. 文档解析。
8. 文档切片。
9. embedding。
10. 向量索引。
11. 检索接口。
12. 检索结果返回。
13. 管理接口的最终权限校验和错误语义。

对 APO 来说，RAG 提供：

```text
知识库管理接口。
```

对 AgentSRE 来说，RAG 提供：

```text
检索接口。
```

### 3.3 AgentSRE 的职责

AgentSRE 是 RAG 的只读消费方，负责在运行时决定“该不该查知识库、查哪个知识库、如何基于检索结果回答”。

AgentSRE 应该负责：

1. 接收用户问题。
2. 执行意图识别。
3. 区分实时诊断请求和知识问答请求。
4. 对 RCA、resource、general 类问题调用实时观测数据。
5. 对 support 类问题检索 `support` 知识库。
6. 对 ops_knowledge 类问题检索 `syncause_default` 知识库。
7. 基于检索结果生成回答。
8. 当 RAG 检索为空时，明确说明当前知识库无法回答。
9. 当 RAG 服务异常时降级处理，不影响实时诊断主流程。

AgentSRE 不应该负责：

1. 知识库创建。
2. 文档上传。
3. 文档删除。
4. 文档列表。
5. 文档解析。
6. 文档切片。
7. embedding。
8. 向量索引。
9. 文档索引状态管理。
10. 知识库管理 API 代理。

也就是说，AgentSRE 对 RAG 的定位应该是：

```text
只读消费方
```

不是：

```text
知识库管理代理层
```

## 四、整体架构与两条链路

### 4.1 总体架构

```text
┌──────────────────────┐
│          APO          │
│  知识库管理页面       │
└──────────┬───────────┘
           │
           │ 同源转发路径
           v
┌──────────────────────┐
│      外部 RAG 服务    │
│ 知识库 / 文档 / 向量索引 │
└──────────┬───────────┘
           ^
           │ 检索接口
           │
┌──────────┴───────────┐
│       AgentSRE        │
│ 意图识别 / 问答路由 / SRE 分析 │
└──────────────────────┘
```

### 4.2 知识库管理链路

```text
APO 前端知识库管理页面
  ↓
通过同源转发路径调用 RAG 管理接口
  ↓
外部 RAG 完成文档上传、解析、切片、embedding、索引
  ↓
文档进入 syncause_default
```

这条链路中：

```text
AgentSRE 不参与。
```

除非将来有统一网关、安全审计或鉴权要求，才考虑让某个后端做代理。但从职责上看，知识库管理仍然属于 RAG，不属于 AgentSRE。

### 4.3 AgentSRE 问答消费链路

```text
用户在 /agent 提问
  ↓
AgentSRE 做意图识别
  ↓
根据分类决定是否调用 RAG 检索接口
  ↓
RAG 返回检索结果
  ↓
AgentSRE 基于检索结果生成回答
  ↓
APO /agent 页面展示回答
```

当用户问：

```text
CPU 飙高一般怎么排查？
```

目标链路是：

```text
classify = ops_knowledge
knowledge_base = syncause_default
```

当用户问：

```text
帮我分析 ts-order-service 最近 15 分钟错误率为什么升高。
```

目标链路是：

```text
classify = system_error_rca
走实时观测数据分析
不使用知识库替代实时分析
```

## 五、知识库职责拆分

知识库都由外部 RAG 服务维护，但使用方不同。

| 知识库 | owner | 主要使用方 | 职责 |
|---|---|---|---|
| `intention_examples` | 外部 RAG | AgentSRE | 意图识别辅助库，只放分类样例、路由规则、边界说明 |
| `syncause_default` | 外部 RAG | AgentSRE / APO 管理页面 | 运维知识库，放 Runbook、SOP、排障方法、最佳实践 |
| `support` | 外部 RAG | AgentSRE / APO 管理页面 | 产品支持知识库，放安装部署、配置说明、产品 FAQ |

### 5.1 `intention_examples`

用途：

```text
辅助 AgentSRE 做意图识别。
```

内容包括：

```text
分类样例
路由规则
分类边界说明
少量高质量示例
```

不应该放：

```text
大段产品文档
Runbook 正文
SOP 正文
故障复盘全文
```

原因是意图识别需要稳定、轻量、结构化的分类参考，不需要大量正文知识。

### 5.2 `syncause_default`

用途：

```text
作为运维知识库，回答运维方法论类问题。
```

内容包括：

```text
CPU 飙高排查方法
内存泄漏排查方法
数据库连接池耗尽处理方法
慢请求排查 Runbook
服务重启 SOP
故障处理最佳实践
常见告警排查步骤
```

典型问题：

```text
CPU 飙高一般怎么排查？
慢请求问题有哪些常见排查步骤？
数据库连接池耗尽应该怎么处理？
服务频繁重启一般看哪些指标？
```

这些问题应该由 AgentSRE 分类为：

```text
ops_knowledge
```

并检索：

```text
syncause_default
```

### 5.3 `support`

用途：

```text
作为产品支持知识库，回答产品安装、配置、使用说明和 FAQ。
```

内容包括：

```text
Syncause 安装部署说明
Agent 探针部署说明
Server 配置说明
数据源接入说明
产品使用 FAQ
```

典型问题：

```text
Syncause 怎么安装？
Agent 探针怎么部署？
Elastic 数据源怎么接？
这个配置项是什么意思？
```

这些问题应该由 AgentSRE 检索：

```text
support
```

并基于文档回答。

## 六、AgentSRE 问答路由规则

AgentSRE 只负责“根据问题选择该查什么”。

推荐路由如下：

| 用户问题类型 | AgentSRE 分类 | AgentSRE 行为 | 使用知识库 |
|---|---|---|---|
| 慢请求、超时、延迟升高，并要求分析当前根因 | `system_slow_rca` | 查实时指标、日志、Trace | 不用知识库替代实时分析 |
| 错误率升高、请求失败、大量异常，并要求分析当前根因 | `system_error_rca` | 查实时指标、日志、Trace | 不用知识库替代实时分析 |
| CPU、内存、磁盘、网络等当前资源异常 | `resource` | 查实时资源指标 | 不用知识库替代实时分析 |
| 当前服务健康状态、当前系统状态、泛化 SRE 分析 | `general` | 走实时分析或通用 SRE 分析 | 不用知识库替代实时分析 |
| 产品安装、配置、使用说明、FAQ | `support` | 检索产品支持文档 | `support` |
| 运维方法、排障方法、Runbook、故障 SOP、最佳实践 | `ops_knowledge` | 检索运维知识库 | `syncause_default` |
| 非 SRE / 非 Syncause 问题 | `not_sre` | 拒答或提示范围外 | 不检索 |

分类名不应该写成：

```text
syncause_default
```

因为 `syncause_default` 是知识库名，不是业务意图名。

更合理的是：

```text
classify = ops_knowledge
knowledge_base = syncause_default
```

## 七、当前 APO 仓现状

当前商业 APO 仓包含：

- 商业前端页面与组件，主要位于 `frontend/src/pro/**`
- `/agent` 页面与 AG-UI 对话能力，主要位于 `frontend/src/pro/views/agent/**`、`frontend/src/pro/components/AGUIChat/**`
- 对外部 AgentSRE 服务的 `/ai/api/**` 调用封装，例如 `frontend/src/pro/api/conversation.ts`、`frontend/src/pro/api/modelSetting.ts`
- 开发态 Vite 代理已把 `/ai/api` 转发到本地 AgentSRE，并去掉 `/ai` 前缀
- 主 Go 后端当前不实现 `/ai/api/**`，系统文档已将其标记为外部 AI / AgentSRE 边界

当前 APO 仓没有独立的“知识库管理”页面，也没有文档上传、文档列表、删除文档、索引状态查看等前端能力。

当前仓库中开发态 Vite 只看到 `/api` 和 `/ai/api` 代理配置，尚未看到面向 RAG 管理接口的独立开发代理；`frontend/nginx.conf` 中也未看到该生产代理配置。生产路径可能由外部 Ingress、网关或部署 Helm 负责。实现前需要确认真实部署链路，避免本地可用但线上 404。

## 八、APO 功能需求

### 8.1 新增知识库管理页面

APO 需要新增一个商业前端页面，页面名称为“知识库管理”。

建议入口路径为：

```text
/knowledge-bases
```

页面至少包含：

- 知识库选择器。
- 文档列表。
- 上传文档按钮。
- 文档状态展示。
- 删除文档操作。
- 索引状态查看。
- 失败原因展示。
- 刷新和空状态。

一期页面默认聚焦 `syncause_default`，但接口和展示层应能读取 RAG 管理接口返回的知识库列表，以便同时展示只读的 `support` 和 `intention_examples`。

### 8.2 新增 RAG 管理接口消费层

APO 前端需要新增 API wrapper，通过同源转发路径消费 RAG 服务提供的知识库管理接口。接口契约以 `rag-management` 仓库中的 `docs/api/apo-knowledge-base-api.md` 为准，APO 前端访问路径在 RAG 原始接口前加 `/rag` 前缀，即：

```text
/rag/api/kb/** -> http://192.168.1.69:5003/api/kb/**
```

本地开发代理需要将 `/rag` 转发到：

```text
http://192.168.1.69:5003
```

并 rewrite 掉 `/rag` 前缀。

APO API wrapper 需要负责：

- 复用 APO 前端现有请求能力、登录态和语言头处理。
- 规范化知识库列表、文档列表、上传、删除、索引状态响应。
- 处理 RAG 管理接口经转发返回的错误信息。
- 为页面提供稳定的 TypeScript 类型。
- 对 `multipart/form-data` 上传保持文件字段与 RAG 管理接口契约一致。

APO 仍应只依赖前端同源转发后的 `/rag/api/kb/**` 路径，而不是把外部 RAG 的真实 URL 写入前端，也不是新增 APO Go 后端知识库管理 API。

### 8.3 知识库列表展示

APO 页面需要展示 RAG 管理接口返回的知识库列表。

推荐展示字段：

| 字段 | 说明 |
|---|---|
| 知识库 ID | `syncause_default`、`support`、`intention_examples` |
| 名称 | 用户可读名称 |
| 描述 | 知识库用途说明 |
| 是否可写 | 一期只有 `syncause_default` 可写 |

一期规则：

- `syncause_default`：运维知识库，可上传、删除、查看索引状态。
- `support`：产品支持知识库，只读。
- `intention_examples`：意图识别辅助库，只读，不向普通用户开放上传。

### 8.4 文档列表展示

APO 页面需要支持查询和展示指定知识库下的文档列表。

列表字段至少包括：

| 字段 | 说明 |
|---|---|
| 文档 ID | RAG 管理接口返回的文档标识 |
| 文档名称 | 原始文件名 |
| 知识库 | 当前所属知识库 |
| 文件类型 | 文件扩展名或 MIME 类型 |
| 文档状态 | 文档记录状态 |
| 索引状态 | RAG 返回的 `index_status` |
| 上传时间 | 文档上传时间 |
| 文件大小 | 原始文件大小 |
| 失败原因 | 失败时展示 |
| 操作 | 查看状态、删除 |

查询能力至少包括：

- 分页。
- 服务端文件名搜索，使用 `filename` 或 `keyword` 查询参数。
- 文档状态筛选，使用 `status` 查询参数。
- 手动刷新。

### 8.5 文档状态映射

APO 需要把 RAG 管理接口返回的 `index_status` 映射为用户可理解的展示状态。

一期建议状态为：

| 状态 | 页面语义 |
|---|---|
| `pending` | 待索引 |
| `indexing` | 索引中 |
| `indexed` | 完成 |
| `failed` | 失败 |

当状态为 `failed` 时，页面必须展示 `error_message` 或等价失败原因。如果上游没有返回失败原因，页面应展示明确兜底文案，例如“索引失败，未返回具体原因”。

### 8.6 上传文档

APO 页面需要支持向 `syncause_default` 上传文档。

上传要求：

- 请求类型为 `multipart/form-data`。
- 文件字段名按 RAG 管理接口契约固定为 `file`。
- 一期仅允许目标知识库为 `syncause_default`。
- 一期支持格式以 RAG 管理接口为准：`.txt`、`.md`、`.docx`、`.csv`、`.xls`、`.xlsx`。
- 一期建议单文件大小限制为 20 MB。
- 一期建议单次最多上传 5 个文件。
- 上传成功后页面应展示返回的 `doc_id`、`index_status`。
- 上传成功后应刷新文档列表，或把新文档乐观插入列表并随后刷新。

上传失败时：

- 文件格式或大小不符合限制，应在前端直接阻止并提示。
- RAG 管理接口经转发返回业务错误时，应展示后端错误信息。
- 网络错误或超时，应提示用户稍后重试，不应误认为上传成功。

### 8.7 删除文档

APO 页面需要支持删除 `syncause_default` 下的文档。

删除要求：

- 只对可写知识库显示删除操作。
- 删除前必须二次确认。
- 删除成功后应从列表移除或刷新列表。
- 删除失败时保留原列表状态并展示错误信息。

一期不要求实现批量删除。

### 8.8 文档索引状态

上传接口返回 `doc_id` 后，APO 需要支持查看文档索引状态。

页面行为：

- 上传成功后如返回 `indexing`，可以自动轮询索引状态，直到进入终态。
- 终态包括 `indexed`、`failed`。
- 轮询应有合理上限，避免页面无限请求。
- 用户应能手动查看或刷新索引状态。
- 索引失败时展示 `message` 和 `error_message`。

### 8.9 `/agent` 运维知识问答消费

APO `/agent` 页面本身不需要实现意图识别或 RAG 检索逻辑。

APO 需要保证：

- `/agent` 继续按现有 AG-UI / AgentSRE 调用链发送用户问题。
- 不在前端根据关键词硬编码路由到 `syncause_default`。
- 不在 APO 前端直接调用外部 RAG 检索接口。
- 如果 AgentSRE 返回运维知识问答结果，前端按普通 assistant 消息展示。
- 如果 AgentSRE 后续增加“基于知识库回答”的引用或提示字段，再另立需求扩展前端展示。

## 九、RAG 接口要求

RAG 需要提供两类接口：给 APO 用的管理接口，给 AgentSRE 用的检索接口。

### 9.1 管理接口，给 APO 用

以下为 APO 前端通过同源转发路径访问后的推荐路径。RAG 服务真实接口前缀为 `/api/kb`，APO 前端统一在前面加 `/rag`：

```http
GET    /rag/api/kb/knowledge_bases
GET    /rag/api/kb/knowledge_bases/{kb_id}
GET    /rag/api/kb/knowledge_bases/{kb_id}/documents
POST   /rag/api/kb/knowledge_bases/{kb_id}/documents/preview
POST   /rag/api/kb/knowledge_bases/{kb_id}/documents
GET    /rag/api/kb/documents/{doc_id}/index_status
DELETE /rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}
```

#### 9.1.1 知识库列表

```http
GET /rag/api/kb/knowledge_bases
```

响应至少包含：

```json
{
  "total": 3,
  "current_page": 1,
  "page_size": 10,
  "data": [
    {
      "id": "syncause_default",
      "name": "运维知识库",
      "description": "用于运维方法、Runbook、排障 SOP、最佳实践和常见故障说明",
      "status": "active",
      "writable": true
    }
  ]
}
```

#### 9.1.2 文档列表

```http
GET /rag/api/kb/knowledge_bases/{kb_id}/documents?filename=cpu&limit=20&offset=0
```

响应至少包含：

```json
{
  "items": [
    {
      "id": "doc_001",
      "kb_id": "syncause_default",
      "filename": "CPU高排查Runbook.md",
      "mime_type": "text/markdown",
      "extension": ".md",
      "size_bytes": 10240,
      "status": "active",
      "index_status": "indexed",
      "created_at": "2026-04-28T10:00:00",
      "meta": {
        "description": "CPU 排查方法"
      },
      "error_message": null
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

文档搜索必须走服务端过滤参数：

```http
GET /rag/api/kb/knowledge_bases/{kb_id}/documents?filename=cpu
GET /rag/api/kb/knowledge_bases/{kb_id}/documents?keyword=cpu
```

APO 不应把“当前页本地过滤”作为正式搜索能力。

文档状态筛选使用 `status` 查询参数；`index_status` 一期只作为展示字段和轮询结果，不作为后端筛选条件要求。

#### 9.1.3 上传前预览切片

```http
POST /rag/api/kb/knowledge_bases/{kb_id}/documents/preview
Content-Type: multipart/form-data
```

请求字段：

```text
file: File
chunk_method: separator
chunk_params: {"chunk_size":512,"chunk_overlap":20}
```

响应至少包含：

```json
{
  "success": true,
  "nodes": [
    {
      "id": "node_001",
      "content": "CPU 飙高一般先确认进程、线程、负载和容器限制...",
      "metadata": {
        "kb_id": "syncause_default",
        "file_name": "CPU高排查Runbook.md",
        "doc_id": "doc_001"
      }
    }
  ]
}
```

#### 9.1.4 上传文档并索引

```http
POST /rag/api/kb/knowledge_bases/{kb_id}/documents
Content-Type: multipart/form-data
```

请求字段：

```text
file: File
chunk_method: separator
chunk_params: {"chunk_size":512,"chunk_overlap":20}
```

响应至少包含：

```json
{
  "success": true,
  "message": "文件 CPU高排查Runbook.md 已成功上传并添加到索引",
  "filename": "CPU高排查Runbook.md",
  "kb_id": "syncause_default",
  "doc_id": "doc_001",
  "document_ids": ["doc_001"],
  "index_status": "indexed"
}
```

#### 9.1.5 查询文档索引状态

```http
GET /rag/api/kb/documents/{doc_id}/index_status
```

响应至少包含：

```json
{
  "success": true,
  "index_status": "indexed",
  "error_message": null
}
```

失败状态响应：

```json
{
  "success": true,
  "index_status": "failed",
  "error_message": "vector backend unavailable"
}
```

#### 9.1.6 删除文档

```http
DELETE /rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}
```

响应至少包含：

```json
{
  "message": "文档删除成功"
}
```

如果删除失败并返回 `failed_step`，APO 应提示“删除未完成”，不要直接从列表中移除该文档，除非刷新列表后确认已经不存在。

### 9.2 检索接口，给 AgentSRE 用

AgentSRE 需要消费 RAG 检索接口，而不是管理接口。

推荐形态：

```http
POST /retrieval
```

AgentSRE 侧应通过配置传入：

```text
query
knowledge_id
top_k
score_threshold
```

AgentSRE 不应实现以下管理接口：

```http
POST /api/kb/knowledge_bases/{kb_id}/documents
DELETE /api/kb/knowledge_bases/{kb_id}/documents/{doc_id}
GET /api/kb/knowledge_bases/{kb_id}/documents
GET /api/kb/documents/{doc_id}/index_status
```

除非后续架构明确要求 AgentSRE 做统一网关或权限代理。但从职责设计上，不推荐这么做。

## 十、非目标

以下内容不属于 APO 本仓一期实现范围：

- 不在 APO 前端硬编码外部 RAG 服务真实地址；只通过同源转发路径访问 RAG 管理接口。
- 不在 APO Go 后端实现 RAG 文档解析、切片、embedding、索引。
- 不在 APO Go 后端直接连接向量数据库。
- 不在 APO Go 后端新增知识库管理 API 或 RAG 管理代理。
- 不在 APO 前端实现意图识别规则。
- 不在 APO 前端把关键词硬编码路由到 `syncause_default`。
- 不改变 RCA、resource、findings、general 等实时诊断链路的核心取数逻辑。
- 不把 `syncause_default` 同时当作意图识别样例库和运维知识库。
- 不在一期实现多租户知识库隔离、文档重建索引、引用溯源展示或 metadata filter。

如果后续决定由 APO Go 后端承担代理层，需要重新评估后端共享包治理、接口契约和部署边界。

## 十一、页面与交互需求

### 11.1 页面布局

页面建议采用工作台式布局：

- 顶部：页面标题、知识库选择器、刷新按钮。
- 主区域：文档列表。
- 右上角：上传文档按钮。
- 列表操作区：查看状态、删除。

页面应避免营销式或说明页式布局，应以重复操作效率为主。

### 11.2 空状态

当选中的知识库没有文档时：

- 页面显示空状态。
- 对可写知识库显示上传入口。
- 对只读知识库说明当前知识库不可上传。

### 11.3 错误状态

页面需要区分：

- 知识库列表加载失败。
- 文档列表加载失败。
- 上传失败。
- 删除失败。
- 索引状态查询失败。

错误信息应尽量使用 RAG 管理接口经转发返回的 `message` 或 `error_message`，没有明确返回时使用本地兜底文案。

### 11.4 可写性控制

页面必须以 RAG 管理接口返回的 `writable` 为准控制上传和删除入口。

即使用户手动构造请求，RAG 管理接口或转发层仍应负责最终权限校验。APO 前端只做用户体验层面的入口隐藏和提示。

## 十二、源码归属与实现注意事项

### 12.1 商业前端独有范围

以下预期改动属于商业前端独有范围，可以在当前商业开发分支直接实现：

- `frontend/src/pro/views/knowledgeBases/**`
- `frontend/src/pro/api/knowledgeBase.ts`
- `frontend/src/pro/routes.js` 中新增商业路由项
- `frontend/public/locales/*/pro/knowledgeBase.json`
- 与该页面强相关的 `frontend/src/pro/**` 组件或类型

### 12.2 共享前端路径风险

如果实现需要修改以下共享前端路径，应先按仓库治理规则处理：

- `frontend/src/_nav.js`
- `frontend/src/routes.js`
- `frontend/src/i18n.js`
- `frontend/public/locales/*/common.json`

该类改动属于共享前端，应先在 OSS 仓 `feat/agent-ce` 落地，再同步回商业分支，除非用户明确要求直接公开到 OSS `main`。

实现时应优先把页面本体和路由定义放在 `frontend/src/pro/**`，减少共享壳层变更。若必须新增主导航项，需把 `_nav.js` 和菜单种子改动单独标记为共享/商业变体治理项。

如果新增 `frontend/public/locales/*/pro/knowledgeBase.json` 作为独立 namespace，由于当前 `frontend/src/i18n.js` 维护了 `proNsList`，实现时也会触及共享前端壳层。可以选择复用已有 `pro/agent` namespace 降低共享路径改动，也可以按共享前端规则先落 OSS `feat/agent-ce` 后同步。

### 12.3 后端共享包风险

如果实现需要修改 APO Go 后端的菜单、路由、权限种子，例如：

- `backend/pkg/repository/database/init_router.go`
- `backend/pkg/repository/database/init_menu_item.go`
- `backend/pkg/services/permission/**`

这些路径属于共享后端候选。实现前必须先确认是否要回 OSS，以及应使用哪个 OSS 分支。不要擅自默认走 `feat/agent-ce`。

如果一期不需要通过 APO Go 后端新增代理 API，则不应改动主 Go 后端。

## 十三、验收标准

### 13.1 知识库管理页面

1. 用户能进入“知识库管理”页面。
2. 页面能加载知识库列表。
3. 页面能识别 `syncause_default` 为可写知识库。
4. 页面能识别 `support` 和 `intention_examples` 为只读知识库。
5. 页面能分页展示文档列表。
6. 页面能展示文档状态、索引状态、上传时间、文件大小、失败原因。
7. 页面能按关键词和文档状态筛选文档。

### 13.2 上传与索引状态

1. 用户可以上传符合格式和大小限制的文档到 `syncause_default`。
2. 上传成功后页面能展示返回的 `index_status`。
3. 页面能查询或轮询文档索引状态。
4. 索引完成后文档状态更新为完成。
5. 索引失败时页面展示失败原因。
6. 用户不能向只读知识库上传文档。

### 13.3 删除文档

1. 用户可以删除 `syncause_default` 下允许删除的文档。
2. 删除前有二次确认。
3. 删除成功后列表状态正确刷新。
4. 删除失败时页面展示错误并保留原列表状态。
5. 只读知识库不展示删除入口。

### 13.4 AgentSRE 问答链路

1. 用户在 `/agent` 提问运维方法类问题时，APO 前端仍能正常发送消息并展示 AgentSRE 回答。
2. APO 前端不通过关键词硬编码知识库路由。
3. AgentSRE 返回 RAG 空结果或失败时，APO 前端按普通错误或 assistant 消息展示，不应卡死或无限 loading。
4. RCA、resource、findings、general 等实时诊断问题不因知识库页面新增而改变前端取数链路。
5. AgentSRE 不调用 RAG 管理接口管理文档。

### 13.5 边界验收

1. 文档管理链路为 `APO -> RAG 管理接口 -> syncause_default`。
2. 问答消费链路为 `用户 -> AgentSRE -> RAG 检索接口 -> syncause_default / support / intention_examples`。
3. APO 前端不硬编码 RAG 服务真实地址。
4. APO Go 后端不新增知识库管理 API。
5. AgentSRE 不成为知识库管理代理层。

## 十四、测试要求

APO 实现阶段至少需要补充以下验证：

- API wrapper response normalization 测试。
- 文档状态映射测试。
- 上传前文件格式和大小校验测试。
- 删除确认与失败回滚测试。
- 索引状态轮询终止条件测试。
- i18n key 完整性检查。
- `/agent` 现有会话测试不回归。
- `frontend` 构建验证。

如果实现涉及共享前端路径或 Go 后端菜单、权限种子，还需要补充对应路径的回归测试或初始化数据验证。

## 十五、MVP 范围

一期 MVP 建议只做最小闭环：

1. RAG 服务确认并提供知识库管理 API。
2. 部署或网关确认 APO 前端可用的同源转发路径。
3. APO 新增知识库管理页面。
4. APO 页面只允许向 `syncause_default` 上传文档。
5. APO 支持文档列表、上传、删除、索引状态和失败原因展示。
6. AgentSRE 将 `ops_knowledge` 路由到 `syncause_default`。
7. AgentSRE 将意图识别辅助库从 `syncause_default` 切到 `intention_examples`。
8. APO `/agent` 继续消费 AgentSRE 回答，不做前端路由硬编码。

二期再考虑：

- 多知识库写入。
- 文档引用展示。
- metadata filter。
- 租户级权限隔离。
- 文档重建索引。
- 批量删除或重试索引。
- 知识库审计日志。

## 十六、待确认问题

1. 生产环境 `/rag/api/kb/**` 转发链路由哪个仓或部署配置负责。
2. 知识库管理页面是否需要进入主导航；如果需要，菜单 key、图标、排序和权限归属如何定义。
3. 知识库管理是否需要 license gating。
4. 文档删除权限是按管理员、团队还是全局知识库管理员控制。
5. `support` 是否一期只读展示，还是也允许 APO 页面管理。
6. `intention_examples` 是否完全隐藏于普通用户页面，还是只读展示。
