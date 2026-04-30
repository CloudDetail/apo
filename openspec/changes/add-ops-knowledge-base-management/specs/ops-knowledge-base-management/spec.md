## ADDED Requirements

### Requirement: APO 提供知识库管理页面入口
系统 SHALL 在 APO 商业前端提供“知识库管理”页面，并在 pro 构建中通过 `/knowledge-bases` route 渲染该页面。该页面 MUST 至少支持知识库列表、文档列表、上传文档、上传前预览文档切片、预览已上传文档切片、下载原文件、删除文档、导入状态查看、失败原因展示、刷新和空状态。

#### Scenario: 用户直达知识库管理页面
- **WHEN** 用户在 pro 构建中访问 `/knowledge-bases`
- **THEN** 系统 MUST 渲染知识库管理页面，而不是跳转到 `/agent` 或其他诊断页面

#### Scenario: 管理员通过侧边栏进入知识库管理页面
- **WHEN** 管理员拥有 `知识库管理` feature 权限并请求 `/api/permission/config`
- **THEN** 系统 MUST 返回顶层菜单项 `knowledgeBases`
- **AND** 该菜单项 MUST 指向 `/knowledge-bases`

#### Scenario: 页面空状态可用
- **WHEN** 当前选中的知识库没有任何文档
- **THEN** 页面 MUST 展示空状态，并且对可写知识库展示上传入口

### Requirement: APO 通过 `/rag/api/kb/**` 消费 RAG 管理接口
系统 SHALL 通过同源路径 `/rag/api/kb/**` 调用 RAG 管理接口，并且前端业务代码 MUST NOT 直接硬编码外部 RAG 服务真实地址。开发态 Vite 代理 MUST 将 `/rag` 转发到 `http://192.168.1.69:5003` 并去掉 `/rag` 前缀。

#### Scenario: 前端请求知识库列表使用同源路径
- **WHEN** 知识库管理页面加载知识库列表
- **THEN** 前端 MUST 请求 `/rag/api/kb/knowledge_bases`

#### Scenario: 开发代理重写 RAG 前缀
- **WHEN** 本地开发环境收到 `/rag/api/kb/knowledge_bases` 请求
- **THEN** Vite proxy MUST 将其转发为 `http://192.168.1.69:5003/api/kb/knowledge_bases`

#### Scenario: 前端代码不写真实 RAG 地址
- **WHEN** 实现 API wrapper 或页面组件
- **THEN** 前端业务代码 MUST NOT 直接引用 `http://192.168.1.69:5003` 作为请求地址

### Requirement: APO 展示 RAG 返回的可写知识库
系统 SHALL 调用 RAG 管理接口获取知识库列表，并且页面 MUST 只向用户展示 `writable === true` 的知识库。`syncause_default` MUST 作为一期默认运维知识库，但只有当 RAG 明确返回其 `writable` 为 `true` 时才可在页面中展示；`support`、`intention_examples` 以及其他只读或未声明可写的知识库 MUST NOT 出现在用户可选知识库列表中。用户界面 MUST 使用知识库 `name` 作为展示文案，MUST NOT 拼接展示知识库 ID。

#### Scenario: syncause_default 为可写知识库
- **WHEN** RAG 返回 `syncause_default` 且 `writable` 为 `true`
- **THEN** 页面 MUST 展示该知识库，并允许用户在该知识库中上传和删除文档

#### Scenario: 只读知识库不向用户展示
- **WHEN** RAG 返回 `support` 或 `intention_examples` 且 `writable` 为 `false`
- **THEN** 页面 MUST 不在知识库选择器中展示这些知识库

#### Scenario: writable 缺失时不展示知识库
- **WHEN** RAG 知识库响应缺少 `writable` 字段
- **THEN** 前端 MUST NOT 将该知识库默认视为可写，也 MUST NOT 将其展示给用户管理

#### Scenario: 知识库选择器只展示名称
- **WHEN** RAG 返回某个可写知识库的 `id` 和 `name`
- **THEN** 知识库选择器 MUST 只展示 `name`
- **AND** 前端 MUST NOT 展示 `name (id)` 或单独展示 `id`

### Requirement: APO 支持服务端分页、搜索和文档状态筛选
系统 SHALL 支持按知识库查询文档列表，并使用 RAG 文档列表接口的 `limit`、`offset`、`filename` 或 `keyword`、`status` 查询参数完成分页、服务端文件名搜索和文档状态筛选。APO MUST NOT 将当前页本地过滤作为正式搜索能力。

#### Scenario: 文档列表分页加载
- **WHEN** 用户打开某个知识库的文档列表
- **THEN** 前端 MUST 请求 `/rag/api/kb/knowledge_bases/{kb_id}/documents` 并传递 `limit` 和 `offset`

#### Scenario: 文件名搜索走服务端过滤
- **WHEN** 用户输入文档搜索关键词
- **THEN** 前端 MUST 使用 `filename` 或 `keyword` 查询参数请求服务端过滤结果

#### Scenario: 文档状态筛选走 status 参数
- **WHEN** 用户按文档状态筛选列表
- **THEN** 前端 MUST 使用 `status` 查询参数请求服务端筛选结果

### Requirement: APO 展示文档列表字段
系统 SHALL 在文档列表中展示 RAG 返回的文档核心字段，包括文档 ID、文档名称、知识库名称、文件类型、文件大小、文档状态、导入状态、上传时间、失败原因和操作入口。文档响应中的 `kb_id` SHALL 仅用于前端映射当前已加载知识库的 `name`，用户界面 MUST NOT 直接展示 `kb_id`。

#### Scenario: 文档列表展示核心字段
- **WHEN** RAG 文档列表返回 `items[].id`、`filename`、`kb_id`、`extension`、`mime_type`、`size_bytes`、`status`、`index_status`、`created_at` 和 `error_message`
- **THEN** 页面 MUST 将这些字段映射为用户可读的列表列或详情信息，并将 `kb_id` 映射为知识库名称展示

#### Scenario: 处理失败展示失败原因
- **WHEN** 某个文档的 `index_status` 为 `failed`
- **THEN** 页面 MUST 展示 `error_message`，如果上游没有返回具体原因则展示明确的本地兜底失败文案

### Requirement: APO 使用 index_status 展示和轮询导入状态
系统 SHALL 使用 RAG 返回的 `index_status` 表示文档导入状态，并将 `pending`、`indexing`、`indexed`、`failed` 分别映射为 `待处理`、`处理中`、`可用`、`处理失败`。系统 MUST NOT 依赖 `job_id`、`/jobs/{job_id}` 或进度百分比作为一期状态模型。

#### Scenario: pending 状态映射为待处理
- **WHEN** RAG 返回 `index_status` 为 `pending`
- **THEN** 页面 MUST 展示待处理状态

#### Scenario: indexing 状态触发有界轮询
- **WHEN** 上传接口或状态接口返回 `index_status` 为 `indexing`
- **THEN** 前端 MAY 轮询 `/rag/api/kb/documents/{doc_id}/index_status`，且 MUST 设置停止条件和轮询上限

#### Scenario: indexed 或 failed 停止轮询
- **WHEN** 导入状态变为 `indexed` 或 `failed`
- **THEN** 前端 MUST 停止该文档的自动轮询

#### Scenario: 用户手动刷新单个文档导入状态
- **WHEN** 用户点击导入状态列右侧的刷新图标
- **THEN** 前端 MUST 查询该文档的导入状态并更新当前行
- **AND** 操作列 MUST NOT 再展示独立的“状态”按钮

### Requirement: APO 支持上传文档到 syncause_default
系统 SHALL 支持向 `syncause_default` 上传文档，并按 RAG 契约以 `multipart/form-data` 提交文件字段 `file`。一期允许的文件扩展名 MUST 仅包含 `.txt`、`.md`、`.docx`、`.csv`、`.xls`、`.xlsx`，且 MUST NOT 支持 PDF。

#### Scenario: 上传合法文件
- **WHEN** 用户在 `syncause_default` 中选择合法格式且大小不超过限制的文件并确认上传
- **THEN** 前端 MUST 调用 `POST /rag/api/kb/knowledge_bases/syncause_default/documents` 并使用表单字段 `file`
- **AND** 请求体 MUST 使用 `chunk_method=separator` 和 `chunk_params={"separator":["\n\n","\n",".","。","?","？","!","！"," "],"chunk_size":1024,"chunk_overlap":100}`

#### Scenario: 阻止 PDF 上传
- **WHEN** 用户选择 `.pdf` 文件
- **THEN** 前端 MUST 在提交前阻止上传并提示该格式不属于一期支持范围

#### Scenario: 上传成功后更新列表
- **WHEN** 上传接口返回 `doc_id` 和 `index_status`
- **THEN** 页面 MUST 展示本次上传结果，并刷新文档列表或将新文档插入列表后再刷新确认

#### Scenario: 上传失败展示上游错误
- **WHEN** RAG 管理接口返回上传错误
- **THEN** 页面 MUST 优先展示 `error_message`、`error` 或 `message` 中的上游错误信息

### Requirement: APO 支持上传前预览切片
系统 SHALL 在可写知识库中支持调用上传前预览切片接口。预览接口 MUST 使用 `multipart/form-data`，文件字段 MUST 为 `file`，查询参数 MUST 使用 `limit` 和 `offset` 控制分页，并且 MUST NOT 创建正式文档记录、chunks 或要求 APO 前端写入向量索引。多文件待上传时，预览入口 MUST 归属于具体文件项，MUST NOT 使用只预览第一个文件的全局按钮。

#### Scenario: 可写知识库展示预览入口
- **WHEN** 当前知识库 `writable` 为 `true`
- **THEN** 页面 MAY 展示上传前预览切片入口

#### Scenario: 预览请求使用 RAG preview 接口
- **WHEN** 用户请求预览切片
- **THEN** 前端 MUST 调用 `POST /rag/api/kb/knowledge_bases/{kb_id}/documents/preview?limit=50&offset=0`
- **AND** 请求体 MUST 包含当前文件项对应的 `file`
- **AND** 请求体 MUST 使用 `chunk_method=separator` 和 `chunk_params={"separator":["\n\n","\n",".","。","?","？","!","！"," "],"chunk_size":1024,"chunk_overlap":100}`，除非后续 UI 明确允许用户选择切片方法
- **AND** 前端 MUST 归一化并展示返回的 `nodes[].content`

#### Scenario: 上传前预览滚动加载后续切片
- **WHEN** 上传前预览返回的 `total` 大于当前已加载切片数
- **THEN** 页面 MUST 在预览弹窗滚动到底部时继续请求下一页
- **AND** 如果当前已加载内容不足以撑出滚动条但仍有后续切片，页面 MUST 自动继续请求下一页
- **AND** 后续请求 MUST 重新带同一个 `file`、`chunk_method` 和 `chunk_params`
- **AND** 后续请求 MUST 使用 RAG 返回的 `offset + limit` 推进新的 `offset`，并保留 `limit=50`
- **AND** 后续切片 MUST 追加到当前弹窗内容中

#### Scenario: 上传前预览忽略过期请求
- **WHEN** 用户在预览请求未完成前切换到另一个文件预览或关闭预览弹窗
- **THEN** 旧请求返回的切片 MUST NOT 追加到新的预览弹窗
- **AND** 旧请求失败时 MUST NOT 覆盖当前预览状态

#### Scenario: 多文件上传时逐文件预览
- **WHEN** 用户在上传弹窗中选择多个文件
- **THEN** 每个文件项旁边 MUST 展示该文件自己的预览入口
- **AND** 页面 MUST NOT 通过 `uploadFiles[0]` 或等价逻辑只预览第一个文件

### Requirement: APO 支持预览已上传文档切片
系统 SHALL 在文档列表操作列提供已上传文档切片预览入口。预览 MUST 通过 RAG 管理接口 `GET /rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/preview` 完成，APO 前端 MUST 使用接口的 `limit` 和 `offset` 查询参数做滚动分页加载，MUST 只展示 RAG 返回的 `nodes[].content`，MUST NOT 重新下载原文件、重新解析文件或重新索引文档。

#### Scenario: 用户预览已上传文档切片
- **WHEN** 用户点击某个文档的预览操作
- **THEN** 前端 MUST 请求 `/rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/preview?limit=50&offset=0`
- **AND** 预览请求 MUST 继续走同源 `/rag` 代理
- **AND** 页面 MUST 在弹窗中展示返回的切片内容

#### Scenario: 用户滚动加载后续切片
- **WHEN** 已上传文档切片预览接口返回 `total` 大于当前已加载页
- **THEN** 页面 MUST 在预览弹窗滚动到底部时使用下一页 `offset` 继续请求该 preview 接口
- **AND** 如果当前已加载内容不足以撑出滚动条但仍有后续切片，页面 MUST 自动继续请求下一页
- **AND** 下一页 `offset` MUST 按 RAG 返回的 `offset + limit` 推进，而不是按前端过滤后的切片数量推进
- **AND** 后续切片 MUST 追加到当前弹窗内容中
- **AND** 页面 MUST NOT 使用固定本地截断只展示前 20 条或前 50 条切片

#### Scenario: 已上传文档预览忽略过期请求
- **WHEN** 用户在预览请求未完成前切换到另一个文档预览或关闭预览弹窗
- **THEN** 旧请求返回的切片 MUST NOT 追加到新的预览弹窗
- **AND** 旧请求失败时 MUST NOT 覆盖当前预览状态

#### Scenario: 切片预览弹窗继承当前主题
- **WHEN** 页面展示已上传文档切片预览
- **THEN** 前端 MUST 使用组件内受控弹窗或等价方式渲染弹窗
- **AND** 前端 MUST NOT 使用脱离当前 AntD `ConfigProvider` 主题上下文的静态 `Modal.info()`

### Requirement: APO 支持删除可写知识库中的文档
系统 SHALL 支持删除 `syncause_default` 中的文档。删除操作 MUST 只在可写知识库中展示，MUST 在提交前二次确认，并在删除失败时保留原列表状态。

#### Scenario: 删除前二次确认
- **WHEN** 用户点击删除某个文档
- **THEN** 页面 MUST 先展示确认交互，而不是立即发送删除请求

#### Scenario: 删除成功后刷新列表
- **WHEN** RAG 删除接口返回成功
- **THEN** 页面 MUST 刷新文档列表或移除该文档并随后以服务端列表确认

#### Scenario: 删除部分失败不移除文档
- **WHEN** RAG 删除接口返回错误且包含 `failed_step`
- **THEN** 页面 MUST 提示删除未完成和失败步骤，并且 MUST NOT 直接从列表中移除该文档

### Requirement: APO 支持下载已上传文档原文件
系统 SHALL 在文档列表操作列提供原文件下载入口。下载 MUST 通过 RAG 管理接口 `GET /rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/download` 完成，APO 前端 MUST NOT 直接猜测或拼接 RAG 文件存储路径。前端 SHOULD 使用显式请求获取文件 blob，并在请求失败时展示错误提示，而不是仅依赖静态浏览器下载链接。

#### Scenario: 用户下载文档原文件
- **WHEN** 用户点击某个文档的下载操作
- **THEN** 前端 MUST 请求 `/rag/api/kb/knowledge_bases/{kb_id}/documents/{doc_id}/download`
- **AND** 下载请求 MUST 继续走同源 `/rag` 代理
- **AND** 请求失败时 MUST 展示下载失败提示

### Requirement: APO 前端不实现 RAG 检索和 AgentSRE 意图路由
系统 SHALL 保持 `/agent` 现有 AG-UI / AgentSRE 调用链。APO 前端 MUST NOT 根据关键词硬编码 `syncause_default` 路由，MUST NOT 直接调用 RAG 检索接口回答用户问题，也 MUST NOT 将知识库管理页面新增为 AgentSRE 管理代理。

#### Scenario: 运维知识问题仍发送给 AgentSRE
- **WHEN** 用户在 `/agent` 中提问运维方法类问题
- **THEN** APO 前端 MUST 按现有 AgentSRE 消息链路发送请求，而不是直接调用 `/retrieval` 或 `/rag/api/kb/**/retrieve`

#### Scenario: 知识库管理页面不影响实时诊断链路
- **WHEN** 用户在 `/agent` 中提问 RCA、resource、findings 或 general 类实时诊断问题
- **THEN** APO 前端 MUST 保持现有调用链路，不因知识库管理页面新增而改变实时诊断取数逻辑

### Requirement: APO 不新增 Go 后端知识库管理 API
系统 SHALL 将知识库管理能力归属到外部 RAG 服务。APO Go 后端 MUST NOT 在本 change 中新增用于上传、删除、列出文档或查询导入状态的知识库管理 API。

#### Scenario: 文档上传不经过 APO Go 后端管理接口
- **WHEN** 用户上传知识库文档
- **THEN** 前端 MUST 通过 `/rag/api/kb/**` 转发到 RAG 管理接口，而不是请求 APO Go 后端新增的 `/api/knowledge-bases/**` 接口

#### Scenario: APO 后端不承担文档处理能力
- **WHEN** 文档上传成功
- **THEN** 文档解析、切片、embedding 和索引 MUST 由 RAG 服务完成，而不是由 APO Go 后端完成

### Requirement: 页面错误状态可区分且可恢复
系统 SHALL 区分知识库列表加载失败、文档列表加载失败、上传失败、删除失败和导入状态查询失败，并为每类失败展示可恢复的用户反馈。

#### Scenario: 知识库列表加载失败
- **WHEN** 知识库列表接口请求失败
- **THEN** 页面 MUST 展示加载失败状态，并提供重新加载入口

#### Scenario: 导入状态查询失败
- **WHEN** 文档导入状态查询接口失败
- **THEN** 页面 MUST 展示状态查询失败提示，且 MUST NOT 让该文档永久停留在无限 loading 状态
