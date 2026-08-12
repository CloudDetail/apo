# Hermes 查询结果 Evidence 展示开发方案

## 1. 建设目标

Hermes 在调用 APO API 查询指标、拓扑和日志后，将查询结果转换为前端可以直接展示的结构化 Evidence。APO Front 不负责理解不同 API 的原始返回结构，只负责识别 Evidence、按 Hermes 事件时间线保存，并调用现有展示组件。

整体链路如下：

```text
Hermes 调用 APO API
→ Hermes 根据数据源和 operationId 解析返回结果
→ 转换为 metrics、topology 或 logs Evidence
→ Evidence 随 tool.complete 事件返回并持久化
→ APO Front 将 Evidence 放入当前 Run 时间线
→ EvidenceRenderer 调用现有指标、拓扑或日志组件
```

职责边界：

| 模块 | 职责 |
| --- | --- |
| APO Front | 识别 Evidence、保持事件顺序和 Run 归属、选择展示组件、提供 JSON 兜底 |
| Hermes | 识别 APO API、解析原始返回、处理字段和时间单位、生成 Evidence、限制数据量、持久化并返回 |
| APO API | 保持现有接口和返回结构，不要求为本需求修改接口 |

## 2. 设计原则

1. Evidence 数据结构按照 APO Front 现有展示组件的输入结构定义，减少前端改造。
2. 前端不能根据任意 API JSON 猜测结果类型和字段路径。
3. Hermes 通过 `sourceId + operationId` 选择确定性转换器。
4. Evidence 与产生它的工具调用使用同一个 `runId`、`toolId` 和时间线位置。
5. 实时事件和历史会话必须使用相同 Evidence 数据，页面刷新后仍可恢复展示。
6. 单个工具调用可以产生零条、一条或多条 Evidence。
7. 转换失败不能影响 Hermes 最终回答，失败结果使用 JSON Evidence 兜底，并记录转换错误。

## 3. Evidence 通用协议

Hermes 在 `tool.complete` 或 `tool.completed` 事件的 `payload` 中增加 `evidence`：

```json
{
  "type": "tool.complete",
  "payload": {
    "tool_id": "tool-001",
    "name": "查询服务指标",
    "result": {},
    "evidence": [
      {
        "id": "evidence-001",
        "kind": "apo_metrics",
        "title": "订单服务 CPU 使用率",
        "summary": "返回 2 条指标时序",
        "source": {
          "sourceId": "system-apo",
          "operationId": "queryMetrics"
        },
        "data": []
      }
    ]
  }
}
```

通用字段：

```ts
interface HermesEvidence<T = unknown> {
  id: string
  kind: 'apo_metrics' | 'apo_topology' | 'apo_logs' | 'apo_json'
  title: string
  summary?: string
  source: {
    sourceId: string
    operationId: string
  }
  query?: {
    startTime?: number
    endTime?: number
    service?: string
    instance?: string
    host?: string
  }
  data: T
}
```

`result` 暂时保留，用于执行详情和问题排查；用户主要查看 `evidence`。后续确认没有兼容需求后，可以只保存经过裁剪的原始结果摘要。

## 4. Evidence 数据结构

### 4.1 指标 Evidence

直接使用现有 Metrics 组件接收的结构：

```ts
interface ApoMetricItem {
  title: string
  unit: string
  timeseries: Array<{
    legend: string
    labels?: Record<string, unknown>
    chart: {
      chartData: Record<string, number>
    }
  }>
  detection?: string
}

interface ApoMetricsEvidence extends HermesEvidence<ApoMetricItem[]> {
  kind: 'apo_metrics'
}
```

完整示例：

```json
{
  "id": "evidence-metric-001",
  "kind": "apo_metrics",
  "title": "订单服务 CPU 使用率",
  "summary": "返回 2 个实例的 CPU 时序",
  "source": {
    "sourceId": "system-apo",
    "operationId": "queryMetrics"
  },
  "query": {
    "startTime": 1786464000000,
    "endTime": 1786467600000,
    "service": "order-service"
  },
  "data": [
    {
      "title": "CPU 使用率",
      "unit": "%",
      "timeseries": [
        {
          "legend": "order-service-1",
          "labels": {
            "service": "order-service",
            "instance": "order-service-1"
          },
          "chart": {
            "chartData": {
              "1786464000000000": 32.5,
              "1786464060000000": 38.1
            }
          }
        }
      ]
    }
  ]
}
```

时间戳按照现有组件约定使用微秒字符串，前端现有转换逻辑会除以 `1000` 后交给 ECharts。该约定应通过契约测试固定，避免不同接口混用秒、毫秒和微秒。

### 4.2 日志 Evidence

现有日志组件的核心输入是 `timestamp` 和 `message`，可保留扩展字段：

```ts
interface ApoLogRecord {
  timestamp: number | string
  message: string
  level?: string
  service?: string
  instance?: string
  traceId?: string
}

interface ApoLogsEvidence extends HermesEvidence<ApoLogRecord[]> {
  kind: 'apo_logs'
  total?: number
}
```

示例：

```json
{
  "id": "evidence-log-001",
  "kind": "apo_logs",
  "title": "订单服务错误日志",
  "summary": "命中 126 条，展示前 100 条",
  "total": 126,
  "source": {
    "sourceId": "system-apo",
    "operationId": "queryLogs"
  },
  "data": [
    {
      "timestamp": 1786464000000000,
      "level": "ERROR",
      "service": "order-service",
      "instance": "order-service-1",
      "traceId": "4f05a6c1",
      "message": "connection timeout"
    }
  ]
}
```

时间戳同样按照现有 `NodeLogs` 组件约定使用微秒。Hermes 必须在转换阶段统一时间单位。

### 4.3 拓扑 Evidence

拓扑数据直接使用现有 ReactFlow `Topology` 组件所需的 `nodes/edges` 结构：

```ts
interface ApoTopologyNode {
  id: string
  type?: string
  position: { x: number; y: number }
  data: {
    label: string
    service?: string
    endpoint?: string
    instance?: string
    isTraced?: boolean
    outOfGroup?: boolean
    status?: 'normal' | 'warning' | 'critical'
    [key: string]: unknown
  }
}

interface ApoTopologyEdge {
  id: string
  source: string
  target: string
  type?: string
  data?: {
    latency?: number
    errorRate?: number
    requestCount?: number
    [key: string]: unknown
  }
}

interface ApoTopologyEvidence
  extends HermesEvidence<{
    nodes: ApoTopologyNode[]
    edges: ApoTopologyEdge[]
  }> {
  kind: 'apo_topology'
}
```

Hermes 只需要提供稳定的节点、边和业务数据。`position` 可以统一设置为 `{ "x": 0, "y": 0 }`，前端现有拓扑组件会重新计算布局。

示例：

```json
{
  "id": "evidence-topology-001",
  "kind": "apo_topology",
  "title": "订单服务调用拓扑",
  "summary": "包含 3 个服务和 2 条调用关系",
  "source": {
    "sourceId": "system-apo",
    "operationId": "queryServiceTopology"
  },
  "data": {
    "nodes": [
      {
        "id": "gateway-api",
        "position": { "x": 0, "y": 0 },
        "data": {
          "label": "gateway",
          "service": "gateway",
          "endpoint": "/api/orders",
          "status": "normal"
        }
      },
      {
        "id": "order-service-create",
        "position": { "x": 0, "y": 0 },
        "data": {
          "label": "order-service",
          "service": "order-service",
          "endpoint": "/orders/create",
          "status": "warning"
        }
      }
    ],
    "edges": [
      {
        "id": "gateway-api-order-service-create",
        "source": "gateway-api",
        "target": "order-service-create",
        "data": {
          "latency": 326,
          "errorRate": 0.08,
          "requestCount": 1520
        }
      }
    ]
  }
}
```

### 4.4 JSON 兜底 Evidence

未配置转换器、返回结构不符合契约或转换失败时，Hermes 返回：

```ts
interface ApoJsonEvidence extends HermesEvidence<unknown> {
  kind: 'apo_json'
  conversionError?: string
}
```

该类型由前端折叠展示，避免未知结果导致页面报错或数据完全不可见。

## 5. APO Front 开发方案

### 5.1 改造范围

前端不解析 APO API 原始数据，只完成以下工作：

1. 在 Hermes 类型中增加 `HermesEvidence` 和各 Evidence 类型。
2. 在 `HermesToolActivity` 中增加 `evidence?: HermesEvidence[]`。
3. 实时事件处理从 `payload.evidence` 读取 Evidence。
4. 历史事件恢复同样从持久化事件的 `payload.evidence` 读取 Evidence。
5. Evidence 保留当前事件的 `sequenceNo`、`occurredAt`、`runId` 和 `toolId` 归属。
6. 新增 `EvidenceRenderer`，按照 `kind` 选择现有组件。
7. 未识别的类型进入 JSON 兜底，不影响消息时间线。

### 5.2 时间线处理

推荐将 Evidence 作为产生它的工具活动的一部分，不单独重新排序：

```ts
interface HermesToolActivity {
  id: string
  kind: 'tool'
  name: string
  status: 'running' | 'completed' | 'failed'
  args?: Record<string, unknown>
  result?: unknown
  evidence?: HermesEvidence[]
  durationMs?: number
}
```

`tool.start` 创建执行中活动；相同 `runId + toolId` 的 `tool.complete` 更新该活动并附加 Evidence。因此 Evidence 会自然保留在工具调用发生的时间线位置，不需要前端另外计算排序。

### 5.3 展示方式

`ToolActivity` 保留现有“查看执行过程”，并在工具完成后增加 Evidence 摘要：

```text
✓ 查询服务指标
  查询完成，返回 2 条指标时序
  [查看指标趋势]

  调用参数
  执行结果 JSON
```

第一期建议在对应工具活动内展示，不额外建立全局 Evidence 面板。后续如果需要在最终回答下面汇总“分析依据”，可以按 `runId` 从时间线聚合 Evidence，不需要调整 Hermes 协议。

### 5.4 组件复用

| Evidence Kind | 复用组件 | 前端处理 |
| --- | --- | --- |
| `apo_metrics` | `AGUIChat/components/ToolRender/Metrics.tsx` | 传入 `evidence.data` |
| `apo_logs` | `AGUIChat/components/EvidencePanel/NodeLogs.tsx` | 解除 `FindingsLogItem` 类型依赖后传入 `evidence.data` |
| `apo_topology` | `core/components/ReactFlow/Topology.jsx` | 传入 `evidence.data`，可增加 Hermes 专用简化节点 |
| `apo_json` | 新增简易 JSON 折叠组件 | 展示 `evidence.data` 和转换失败原因 |

前端主要工作量集中在类型适配和样式嵌入，不承担 API 字段转换。

### 5.5 前端测试

1. `tool.complete` 携带指标 Evidence 时展示指标图。
2. 日志 Evidence 可以展示、搜索和复制。
3. 拓扑 Evidence 可以完成布局并展示节点和边。
4. 同一次 Run 多个工具产生的 Evidence 顺序正确。
5. 多个 Run 的 Evidence 不串到其他回答。
6. 页面刷新和打开历史会话后 Evidence 仍能展示。
7. 未知 Evidence Kind 使用 JSON 兜底。
8. Evidence 数据异常时仅当前卡片提示错误，不导致页面崩溃。

## 6. Hermes 开发方案

### 6.1 改造范围

Hermes 是本需求的主要改造方，需要完成：

1. 获取工具调用对应的 `sourceId`、`operationId` 和原始 API 结果。
2. 建立 APO Evidence 转换器注册表。
3. 将 API 原始结果转换成前端现有组件需要的数据结构。
4. 统一时间戳、数值、单位、标签、节点 ID 和空数据行为。
5. 对指标点数、日志条数和拓扑规模进行限制。
6. 将 Evidence 写入 `tool.complete` 事件。
7. 将包含 Evidence 的完整事件持久化到会话历史。
8. 转换失败时记录内部错误并返回 `apo_json` Evidence。

### 6.2 转换器注册

建议按照数据源和接口注册，不按返回 JSON 的偶然字段猜测：

```python
EVIDENCE_NORMALIZERS = {
    ("system-apo", "queryMetrics"): normalize_metrics,
    ("system-apo", "queryLogs"): normalize_logs,
    ("system-apo", "queryServiceTopology"): normalize_topology,
}
```

处理流程：

```text
工具调用完成
→ 获取 sourceId 和 operationId
→ 查找 Evidence Normalizer
→ 校验 API 返回结果
→ 转换字段及时间单位
→ 裁剪数据量
→ 生成 Evidence
→ 写入 tool.complete payload
→ 持久化事件
→ 推送前端
```

一个接口如同时返回指标和拓扑，可以返回多条 Evidence。

### 6.3 指标转换要求

Hermes 负责：

- 从实际返回路径读取指标列表和时序。
- 将指标名称转换为 `title`。
- 明确 `unit`，没有单位时使用空字符串，不能由前端猜测。
- 将标签组合成稳定的 `legend`。
- 将时间戳统一转换为微秒字符串。
- 将数值转换为有限数字，过滤 `NaN`、无穷值和非法字符串。
- 每条时序最多保留 300～500 个点，超出时降采样。
- 空时序返回空数组和明确 `summary`，不能伪造成数值零。

### 6.4 日志转换要求

Hermes 负责：

- 提取日志时间和正文。
- 将时间统一为微秒。
- 尽可能补充 `level`、`service`、`instance` 和 `traceId`。
- 单条消息限制最大长度，防止超大日志拖垮页面和历史存储。
- 默认最多保存前 100～200 条，同时在 `total` 中记录总命中数。
- 空日志结果返回 `data: []`，并在摘要中说明未查询到日志。

### 6.5 拓扑转换要求

Hermes 负责：

- 为每个业务节点生成稳定且唯一的字符串 ID。
- 将服务、接口、实例等信息写入节点 `data`。
- 边的 `source` 和 `target` 必须引用已有节点。
- 去除重复节点、重复边以及无法关联节点的边。
- 自调用边允许 `source === target`。
- `position` 统一写入 `{x: 0, y: 0}`，实际布局由前端完成。
- 首期限制节点和边数量；超过限制时裁剪并在摘要中说明。

### 6.6 事件与持久化

Evidence 必须写入浏览器收到的同一条 `tool.complete` 事件，同时原样进入持久化会话事件：

```json
{
  "event_type": "tool.complete",
  "run_id": "run-001",
  "sequence_no": 18,
  "payload": {
    "tool_id": "tool-001",
    "result": {},
    "evidence": []
  }
}
```

不能只通过 WebSocket 临时推送 Evidence 而不持久化，否则页面刷新和历史会话恢复后将丢失图表、日志和拓扑。

### 6.7 Hermes 测试

每个纳入支持范围的 APO API 都需要固定契约样例和转换测试：

1. 正常返回可以生成正确 Evidence Kind。
2. 时间戳单位转换正确。
3. 指标单位和图例正确。
4. 空结果生成空 Evidence，不误判为正常或零值。
5. 缺少可选字段时仍能生成 Evidence。
6. 返回结构变化时进入 JSON 兜底并记录转换错误。
7. 超量指标、日志和拓扑能够正确裁剪。
8. Evidence 写入实时事件和历史事件的内容一致。
9. 一个 API 返回多种数据时可以生成多条 Evidence。

## 7. 联调顺序

### 第一阶段：Metrics

1. 冻结一个真实 APO metrics API 及返回样例。
2. Hermes 开发指标转换器。
3. Hermes 在 `tool.complete.payload.evidence` 返回 `apo_metrics`。
4. 前端识别并调用现有 Metrics 组件。
5. 验证实时展示和历史恢复。

### 第二阶段：Logs

1. 冻结日志 API 及返回样例。
2. Hermes 统一时间、正文和条数限制。
3. 前端解除 `NodeLogs` 对 Findings 类型的依赖。
4. 验证搜索、复制和历史恢复。

### 第三阶段：Topology

1. 冻结拓扑 API 及返回样例。
2. Hermes 转换为 ReactFlow `nodes/edges`。
3. 前端增加 Hermes 简化拓扑节点并复用布局组件。
4. 验证重复节点、自调用、空拓扑和大拓扑裁剪。

## 8. 验收标准

1. Hermes 查询 metrics、logs 或 topology API 后能够生成对应 Evidence。
2. 前端不解析 API 原始返回结构，只根据 Evidence Kind 展示。
3. 指标能够使用现有 ECharts 组件展示标题、单位、图例和时序。
4. 日志能够使用现有日志组件展示，并支持搜索和复制。
5. 拓扑能够使用现有 ReactFlow 组件展示节点和调用关系。
6. Evidence 在工具调用发生的位置按时间线展示，且不会跨 Run 串联。
7. 页面刷新、会话切换和历史恢复后 Evidence 仍然存在。
8. 未知接口、结构异常和转换失败时使用 JSON 兜底，不影响最终回答。
9. 大量指标点、日志和拓扑数据经过 Hermes 裁剪，不导致页面卡顿或会话数据异常膨胀。

## 9. 结论

本需求采用“前端轻处理、Hermes 重转换”的实现方式。前端只扩展 Evidence 类型、时间线读取和 Renderer，展示层最大限度复用现有组件；Hermes 按 APO API 的 `sourceId + operationId` 建立确定性转换器，将原始查询结果转换为现有前端组件可以直接消费的数据结构，并负责时间单位、字段校验、数据裁剪、异常兜底和历史持久化。

该方案不会重新引入 AgentSRE Findings 业务结构，也不会让 APO Front 依赖具体 API 返回字段。后续增加 Trace、表格或其他数据类型时，只需在 Hermes 增加转换器，并在前端 Renderer Registry 中增加对应组件。
