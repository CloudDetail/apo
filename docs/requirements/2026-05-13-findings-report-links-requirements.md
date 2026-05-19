# 推理树节点详情关联故障报告需求文档

## 一、文档目的

本文档用于定义 Agent 推理树节点详情中展示“相关故障报告”入口的需求范围、跨系统职责、数据契约、展示规则和验收标准。

本需求的核心原则是：

```text
能点就是能看；没有可用报告就不展示链接。
```

因此，本需求不是让前端看到异常节点后自行猜测报告链接，也不是要求所有推理树节点都支持查看报告，而是在节点确实有可验证、可打开、与当前证据同源的慢调用或错误报告时，才在节点详情中提供下钻入口。

## 二、背景现状

当前 `/agent` 页面中的推理树用于展示 AgentSRE 对一次故障分析的判断路径。用户点击树节点后，右侧节点详情面板可以展示：

- 节点结论
- 节点状态
- 指标证据
- 表格证据
- 日志证据
- trace 聚合证据

APO 当前已经存在两类故障报告页面：

| 报告类型 | 路由 | 必需参数 | 说明 |
|---|---|---|---|
| 慢调用报告 | `/#/cause/report/:traceId/:spanId?mutatedType=...` | `traceId`、`spanId` | 展示慢调用根因链、拓扑、指标、profiling、flow 等信息 |
| 错误报告 | `/#/cause/error-report/:traceId` | `traceId` | 展示错误 trace 拓扑、异常节点、错误 flow、日志和 exception 信息 |

这些报告页面不是读取预先保存的报告快照，而是在页面打开后根据 `traceId` / `spanId` 调用 OriginX 相关接口实时查询和组装报告内容。

这意味着：

- 有合法参数时，前端可以直接跳转到对应报告页面。
- 但能打开 URL 不等于一定有有效报告内容。
- 如果 OriginX 查不到对应 trace/span，或者 trace/span 不满足报告生成条件，页面仍可能显示空态、数据缺失或超时。

因此，推理树节点详情中的报告入口必须由 AgentSRE 在 producer 侧确认可用后再返回给 APO 前端。

## 三、总体目标

本需求目标如下：

1. 在推理树节点详情中增加“相关故障报告”区域。
2. 当节点存在可用慢调用报告时，展示“查看慢调用报告”入口。
3. 当节点存在可用错误报告时，展示“查看错误报告”入口。
4. 报告入口必须来自 AgentSRE 返回的结构化字段。
5. 前端只做字段完整性和 URL 安全过滤，不做业务推断。
6. 没有关联报告、报告不可用、字段不完整或来源不明确时，不展示报告入口。
7. 现有推理树、节点详情、findings stream、node retry 和 evidence 展示逻辑不受影响。

## 四、非目标

以下内容不属于本需求范围：

1. 不要求所有推理树节点都支持查看报告。
2. 不在前端根据 service、endpoint、时间范围、节点状态自行反查并拼接报告链接。
3. 不把 markdown 文本中的 report link 当作稳定契约。
4. 不改造 `/cause/report` 和 `/cause/error-report` 页面本身的报告查询逻辑。
5. 不要求 AgentSRE 为每个异常节点都补齐报告链接。
6. 不为只有聚合指标、聚合日志或聚合 trace 摘要的节点强行生成报告入口。
7. 不在节点详情中展示“暂无报告”空态，避免增加无意义噪音。

## 五、核心原则

### 5.1 能点就是能看

报告入口一旦展示，用户在数据保留期内点击后应该能够看到有效报告内容。

这里的“能看”要求是：

- 链接指向现有报告页面。
- 链接参数完整。
- 链接对应的 trace/span 是当前节点证据的来源或强相关样本。
- AgentSRE 在返回链接前已通过可信数据源确认该报告在当前分析窗口内理论上可用。

这里的“能看”不是永久可看承诺，而是要求 AgentSRE 返回链接时已经确认当前分析窗口内存在可用报告数据。

如果报告数据后续因为数据生命周期过期而不可查，这属于数据保留周期问题，不要求前端强行补救。但 AgentSRE 必须只返回当前分析窗口内已确认可用的链接。

### 5.2 没报告就别强硬塞链接

如果节点只是异常，但没有可用报告，不展示入口。

不允许以下行为：

- 看到 `status=abnormal` 就展示报告按钮。
- 只有 serviceName 就去猜测报告。
- 只有时间范围就拉 top N 报告塞给节点。
- 只有错误率升高但没有具体错误 trace，就展示错误报告。
- 只有慢调用方向但没有 `spanId`，就展示慢调用报告。

### 5.3 不是所有节点都需要支持查看报告

报告入口只适用于拥有具体 trace/span 证据的节点。

一句话规则：

```text
不是节点异常就展示报告入口，而是节点拥有可验证、可打开、与当前证据同源的故障报告时才展示入口。
```

## 六、节点适用范围

### 6.1 应支持展示报告入口的节点

以下节点可以展示报告入口，但前提是 AgentSRE 返回了可用的结构化 `reportLinks`：

1. 慢调用 / eBPF / Polaris 类异常节点
   - 例如 CPU、network、scheduling、disk I/O、lock/GC 等方向。
   - 必须能关联到具体异常 span。
   - 慢调用报告必须具备 `traceId + spanId`。

2. trace 类异常节点
   - 例如错误 trace、慢 trace、数据库调用异常 trace。
   - 错误报告必须具备 `traceId`。
   - 慢调用报告必须具备 `traceId + spanId`。

3. 数据库 trace 节点
   - 只有在能落到具体异常 trace/span 时展示。
   - 如果只有“数据库调用变慢”的聚合统计，不展示。

4. 日志节点
   - 只有在日志证据能明确关联到某条错误 trace 时展示错误报告。
   - 如果只是日志数量、错误关键字或异常栈聚合，不展示。

### 6.2 不应展示报告入口的节点

以下节点默认不展示报告入口：

1. `summary` 汇总节点
   - 例如“故障排查开始”“依赖服务检查”“代码运行状态检查”等聚合判断节点。

2. `normal` 正常节点
   - 没有异常报告价值。

3. `unsupported` 节点
   - 没有真实证据，不应给报告入口。
   - 如果后续能够提供真实 trace/span 证据，应将节点归类为 `evidence`，而不是继续以 `unsupported` 节点展示报告入口。

4. 只有指标聚合的节点
   - 例如只有 CPU 使用率、内存使用率、错误率曲线。
   - 如果没有具体 `traceId/spanId`，不展示。

5. 只有日志聚合的节点
   - 除非日志能明确关联到某条错误 trace。

6. 只有服务级 RED 结论的节点
   - 除非 AgentSRE 能给出对应异常 trace/span。

## 七、跨系统职责边界

### 7.1 AgentSRE 职责

AgentSRE 是报告链接的 producer，负责：

1. 判断当前 findings node 是否应该有关联报告。
2. 查询或复用节点证据中的 `traceId` / `spanId`。
3. 确认 report link 与当前节点证据同源或强相关。
4. 确认报告理论上可打开、可消费。
5. 生成结构化 `reportLinks` 字段。
6. 不返回不完整、不确定或不可用的链接。

AgentSRE 不应该：

1. 把不确定的候选 trace 强行塞给节点。
2. 把服务级 top N 异常报告当成当前节点的精确证据。
3. 只在 markdown 文本里附链接而不提供结构化字段。
4. 为没有 trace/span 证据的节点伪造 report link。

### 7.2 APO 前端职责

APO 前端是 report link consumer，负责：

1. 扩展 findings node 类型，保留 `reportLinks` 字段。
2. 在节点详情面板中渲染“相关故障报告”区域。
3. 对 `reportLinks` 做基本完整性过滤。
4. 对 URL 做范围校验，只允许现有故障报告页面。
5. 点击链接时新窗口打开，避免丢失当前 Agent 推理树上下文。

APO 前端不应该：

1. 自行查询 `/api/alerts/anomaly-span/list` 猜测慢调用报告。
2. 自行查询错误实例或 trace 列表猜测错误报告。
3. 根据节点状态自动拼 report link。
4. 在没有合法 link 时展示空的报告区域。

### 7.3 OriginX / 报告页面职责

OriginX 和现有报告页面继续负责：

1. 根据 `traceId` / `spanId` 查询拓扑、flow、日志和报告内容。
2. 在报告不可查时展示已有空态、数据缺失或超时状态。

本需求不改变现有报告页面的内部实现。

## 八、AgentSRE 输出契约

AgentSRE 在 findings node 上新增可选字段 `reportLinks`。

建议结构如下：

```ts
reportLinks?: Array<{
  type: 'slow' | 'error'
  title?: string
  serviceName?: string
  errorGroupKey?: string
  representative?: boolean
  groupCount?: number
  traceId: string
  spanId?: string
  mutatedType?: string
  url: string
  source?: string
}>
```

字段说明：

| 字段 | 必需 | 说明 |
|---|---|---|
| `type` | 是 | 报告类型，取值为 `slow` 或 `error` |
| `title` | 否 | 展示标题，例如 `CPU 慢调用报告`、`错误报告` |
| `serviceName` | 错误报告建议提供 | 错误报告对应的异常服务或依赖服务名，用于前端分组展示 |
| `errorGroupKey` | 错误报告建议提供 | 错误类型、异常签名、endpoint 或其他稳定分组键，用于区分同一服务下的不同错误模式 |
| `representative` | 否 | 是否为同一错误分组内的代表报告；仅影响排序，不追加额外代表文案，也不代表其他报告不可见 |
| `groupCount` | 否 | producer 已知的同组报告数量，用于展示计数；缺失时前端可按当前返回列表计算 |
| `traceId` | 是 | trace ID |
| `spanId` | 慢调用必需 | span ID，慢调用报告必须提供 |
| `mutatedType` | 慢调用建议提供 | 慢调用方向，例如 `cpu_time`、`network_time`、`scheduling_time` |
| `url` | 是 | APO 相对 hash 路由链接，必须精确匹配允许的报告路由 |
| `source` | 否 | 链接来源，例如 `anomaly-span-list`、`error-trace-query`、`ebpf-report-list` |

### 8.1 慢调用报告链接规则

慢调用报告链接必须满足：

1. `type = "slow"`。
2. `traceId` 非空。
3. `spanId` 非空。
4. `url` 必须是 APO 相对 hash 路由，并精确匹配 `/#/cause/report/{traceId}/{spanId}`。
5. 如果有 `mutatedType`，应附加到 query：`?mutatedType={mutatedType}`。
6. `mutatedType` 应与当前节点异常方向一致。

典型示例：

```json
{
  "type": "slow",
  "title": "CPU 慢调用报告",
  "traceId": "abc123",
  "spanId": "def456",
  "mutatedType": "cpu_time",
  "url": "/#/cause/report/abc123/def456?mutatedType=cpu_time",
  "source": "anomaly-span-list"
}
```

### 8.2 错误报告链接规则

错误报告链接必须满足：

1. `type = "error"`。
2. `traceId` 非空。
3. `url` 必须是 APO 相对 hash 路由，并精确匹配 `/#/cause/error-report/{traceId}`。
4. 该 `traceId` 必须来自当前节点使用的错误 trace 证据。
5. AgentSRE 可以使用 OriginX `/originx/api/error/traceIds` 作为错误 trace 候选来源，但查询必须受当前节点语义、服务名和 runtime window 约束。
6. AgentSRE 必须在返回链接前确认 `/originx/api/query/errortraces/{traceId}` 可成功解析；仅 dataplane 中存在 error trace 不代表错误报告可打开。
7. 多个错误报告属于同一依赖服务或同一错误类型时，AgentSRE 不应在数据层丢弃报告，而应尽量提供 `serviceName`、`errorGroupKey`、`representative`、`groupCount`，让前端按错误组展示。
8. 同一错误组内的代表报告由 AgentSRE 选择；建议优先选择当前节点证据直接命中的 trace，其次选择影响面、错误次数或证据完整度更高的 trace，最后用时间和 traceId 做稳定排序兜底。

典型示例：

```json
{
  "type": "error",
  "title": "ts-travel-plan-service error report",
  "serviceName": "ts-travel-plan-service",
  "errorGroupKey": "POST /api/v1/travel-plans:TimeoutException",
  "representative": true,
  "groupCount": 2,
  "traceId": "abc123",
  "url": "/#/cause/error-report/abc123",
  "source": "error-trace-query"
}
```

### 8.3 不返回链接的情况

AgentSRE 在以下情况下不应返回 `reportLinks`：

1. 没有查到可用异常 span。
2. 慢调用报告缺少 `spanId`。
3. 错误报告缺少 `traceId`。
4. 只有聚合指标，没有具体 trace/span。
5. 只有服务名和时间范围，无法证明与当前节点证据同源。
6. 错误报告接口 `/originx/api/query/errortraces/{traceId}` 无法解析该 trace。
7. 查询异常、权限不足或数据源不可用。
8. 报告链接只是历史 markdown 文本中的弱引用，不能作为结构化证据。
9. 节点仍被标记为 `unsupported`，说明该节点没有可消费的真实证据。

## 九、前端展示设计

### 9.1 展示位置

在推理树节点详情面板中新增“相关故障报告”区域。

建议位置：

1. 节点结论之后。
2. 指标、表格、日志、trace 证据之前。

这样用户先看到结论，再看到可下钻的报告入口，然后继续查看当前节点内嵌证据。

### 9.2 展示条件

仅当满足以下条件时展示：

1. `selectedNode.reportLinks` 存在。
2. 过滤后至少有一条合法链接。
3. 链接类型是 `slow` 或 `error`。
4. 链接 URL 属于允许的报告路由。

如果过滤后为空，不展示标题、不展示空态、不展示占位。

### 9.3 链接过滤规则

前端需要执行基础过滤：

| 类型 | 必需字段 | URL 要求 |
|---|---|---|
| `slow` | `traceId`、`spanId`、`url` | 必须是 APO 相对 hash 路由，并精确匹配 `/#/cause/report/{traceId}/{spanId}`，允许携带 `mutatedType` query |
| `error` | `traceId`、`url` | 必须是 APO 相对 hash 路由，并精确匹配 `/#/cause/error-report/{traceId}` |

不合法链接直接丢弃。

前端不能只通过字符串包含关系判断 URL 合法性，例如仅检查是否包含 `/cause/report/` 不足以通过校验。

前端也不能在 `url` 缺失时根据 `traceId` / `spanId` 自行补齐报告路由；`url` 必须由 AgentSRE producer 显式返回。

### 9.4 展示形式

展示为一个轻量操作区：

- 标题：`相关故障报告`
- 慢调用报告按钮：`查看慢调用报告`
- 错误报告按钮：`查看错误报告`
- 多条同类报告可以使用标题区分，例如：
  - `CPU 慢调用报告`
  - `网络慢调用报告`
  - `错误报告 1`
- 错误报告应按“依赖服务 + 错误分组”聚合展示，但组内所有链接都应直接可见，不默认收起、不要求用户展开后才能看到其他 trace。
- 同一错误组内如果存在代表报告，应把代表报告排在前面；其他报告仍作为同组入口展示，不追加额外“代表”文案。
- 当 AgentSRE 提供 `serviceName` / `errorGroupKey` 时，前端优先使用这些字段生成分组；缺失时可退化为从 `title` 或默认文案生成单组，但不得为了分组而丢弃合法链接。

点击行为：

- 使用新窗口打开。
- 不替换当前 `/agent` 页面。
- 不清理当前推理树状态。

### 9.5 国际化要求

需要补充中英文文案：

| Key | 中文 | 英文 |
|---|---|---|
| `evidencePanel.relatedReports` | 相关故障报告 | Related reports |
| `evidencePanel.viewSlowReport` | 查看慢调用报告 | View slow report |
| `evidencePanel.viewErrorReport` | 查看错误报告 | View error report |

如果使用 `title` 字段展示，优先使用 AgentSRE 返回的 `title`；缺失时回退到上述默认文案。

## 十、用户故事

### 10.1 慢调用异常下钻

作为一个排障用户，当我点击推理树中的 CPU/eBPF 异常节点时，如果该节点关联了具体慢调用 span，我希望能直接打开慢调用报告，查看完整调用链、耗时传播、profiling 和相关指标证据。

### 10.2 错误 trace 下钻

作为一个排障用户，当我点击推理树中的错误 trace 节点时，如果该节点关联了具体错误 trace，我希望能直接打开错误报告，查看错误传播链和异常节点详情。

### 10.3 无报告时保持干净

作为一个排障用户，当节点没有可用报告时，我不希望看到无效按钮或空的报告区域，以免误以为系统漏查或报告生成失败。

## 十一、验收标准

### 11.1 慢调用报告展示

给定 AgentSRE 返回以下节点：

```json
{
  "nodeId": "code_cpu_check",
  "status": "abnormal",
  "reportLinks": [
    {
      "type": "slow",
      "traceId": "trace-1",
      "spanId": "span-1",
      "mutatedType": "cpu_time",
      "url": "/#/cause/report/trace-1/span-1?mutatedType=cpu_time"
    }
  ]
}
```

当用户点击该节点时：

- 节点详情展示“相关故障报告”区域。
- 区域内展示慢调用报告入口。
- 点击后新窗口打开 `/#/cause/report/trace-1/span-1?mutatedType=cpu_time`。

### 11.2 错误报告展示

给定 AgentSRE 返回以下节点：

```json
{
  "nodeId": "error_trace_check",
  "status": "abnormal",
  "reportLinks": [
    {
      "type": "error",
      "traceId": "trace-2",
      "url": "/#/cause/error-report/trace-2"
    }
  ]
}
```

当用户点击该节点时：

- 节点详情展示错误报告入口。
- 点击后新窗口打开 `/#/cause/error-report/trace-2`。

### 11.3 无报告不展示

给定异常节点没有 `reportLinks`：

- 节点详情不展示“相关故障报告”标题。
- 不展示“暂无报告”。
- 不影响指标、表格、日志、trace 证据展示。

### 11.4 不完整链接被过滤

给定以下链接：

```json
{
  "type": "slow",
  "traceId": "trace-1",
  "url": "/#/cause/report/trace-1/"
}
```

因为缺少 `spanId`：

- 前端丢弃该链接。
- 如果没有其他合法链接，不展示报告区域。

### 11.5 非法或非报告路由被过滤

给定以下链接：

```json
{
  "type": "error",
  "traceId": "trace-1",
  "url": "https://example.com/report"
}
```

因为 URL 不是 APO 相对 hash 路由，且不属于允许的 APO 故障报告路由：

- 前端丢弃该链接。
- 不展示该按钮。

给定以下链接：

```json
{
  "type": "slow",
  "traceId": "trace-1",
  "spanId": "span-1",
  "url": "/#/foo/cause/report/trace-1/span-1"
}
```

因为 URL 虽然包含 `/cause/report/`，但没有精确匹配允许路由：

- 前端丢弃该链接。
- 不展示该按钮。

### 11.6 多报告共存

给定同一节点同时返回慢调用和错误报告：

- 两个入口都展示。
- 每个入口点击后打开对应报告页面。
- 展示顺序建议为慢调用报告在前，错误报告在后。
- 错误报告如果属于多个依赖服务或错误组，应按组展示。
- 同一错误组内的多个错误报告全部展示，不默认折叠；代表报告只影响排序，不隐藏其他报告，也不追加额外“代表”文案。

### 11.7 历史兼容

历史会话中的 findings node 没有 `reportLinks` 字段时：

- 页面正常展示。
- 不报错。
- 不展示报告区域。

### 11.8 不影响现有交互

新增报告入口后：

- 推理树节点点击逻辑不变。
- findings stream 渐进加载不变。
- node retry 不变。
- `metrics/table/logs/traces` 展示不变。
- `/cause/report` 和 `/cause/error-report` 页面逻辑不变。

## 十二、实现建议

### 12.1 AgentSRE 优先落地顺序

建议分两期：

1. 第一期支持慢调用报告。
   - AgentSRE 已有 `anomaly-span/list -> traceId/spanId -> cause/report` 链路基础。
   - 优先在 eBPF / Polaris 异常节点中输出结构化 `reportLinks`。
   - 第一期验收只要求慢调用报告链路可用，不强制错误报告同时上线。

2. 第二期支持错误报告。
   - AgentSRE 需要从错误 trace 数据中提取 `traceId`。
   - 需要确认该 `traceId` 与当前错误节点证据同源。
   - 对依赖 RED 错误率异常节点，AgentSRE 可以按当前 runtime window 和异常下游服务查询 OriginX `/originx/api/error/traceIds` 获取候选错误 trace。
   - 需要确认 OriginX 错误报告接口能解析该 `traceId`。
   - 再输出 `type=error` 的 `reportLinks`。
   - 当同一依赖服务存在多条错误 trace 时，AgentSRE 应返回完整合法链接，并提供 `serviceName`、`errorGroupKey`、`representative`、`groupCount` 等分组辅助字段；APO 前端负责分组展示但不折叠、不隐藏组内链接。
   - 如果无法证明错误 `traceId` 与当前节点证据同源，或 OriginX 无对应错误报告，则不输出错误报告链接。

### 12.2 APO 前端落地范围

APO 前端建议改动范围：

1. 扩展 `FindingsNodePayload` 类型，新增 `reportLinks`。
2. 在 findings adapter 中保留并规范化 `reportLinks`。
3. 扩展 `FindingsNodeData`。
4. 在 `NodeDetail` 中新增 `RelatedReports` 子组件。
5. 补充中英文文案。
6. 增加单元测试覆盖合法链接、非法链接和无链接场景。

## 十三、风险与约束

### 13.1 主要风险

最大风险是链接看似可用，但与当前推理节点不是同一条证据。

这种情况会误导用户，让用户以为报告内容就是该节点结论的证据来源。因此必须由 AgentSRE producer 侧保证来源关系，前端不能通过 service/time range 自行反查并拼链接。

### 13.2 数据生命周期风险

报告页面依赖 OriginX 实时查询 trace/span。即使 AgentSRE 返回链接时报告可用，用户稍后点击也可能因为数据过期而不可查。

这类问题应通过数据保留周期、报告可用性校验或提示策略解决，不应通过前端伪造链接解决。

### 13.3 契约漂移风险

当前 findings contract 已强调 consumer 只消费 canonical fields。`reportLinks` 应作为新的 canonical extension 进入 AgentSRE contract，再由 APO 前端消费。

不要长期依赖 markdown 中的 `ReportLink` 文本解析。

## 十四、结论

本需求有明确价值：推理树提供诊断结论和判断路径，故障报告提供更底层的 trace/span 级可审计证据。两者打通后，用户可以从“为什么判断异常”直接下钻到“哪一次调用证明异常”。

但该能力必须保持克制：

```text
只有具体 trace/span 证据支撑的节点才展示报告入口；
只有 AgentSRE 确认可用的报告才展示入口；
没有报告就不展示，不占位，不猜测，不强塞。
```
