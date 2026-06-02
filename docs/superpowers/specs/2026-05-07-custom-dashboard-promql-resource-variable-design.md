# 自定义大盘 PromQL 资源变量识别与提示设计

## 背景

自定义大盘已经把 APO 顶层数据组作为页面上下文，并将 `cluster`、`namespace`、`service`、`endpoint`、`pod`、`instance` 等标准资源变量的候选值切到 `/api/v2/data/group/filter`。

接下来需要解决一个语义识别问题：用户在 PromQL 中输入的 label 名称可能不是 APO 内部推荐写法，但含义仍然对应标准资源。例如 `clusterId` 实际上应被识别为系统的 `Cluster` 筛选项，而不是普通 label 变量。

同时，用户编辑 PromQL 时也需要知道哪些 label/变量写法能被系统识别，否则容易写出候选值来源不符合预期的变量。

## 目标

- 将常见 PromQL label alias 归一到标准资源变量。
- 保留用户 PromQL 中的原始 label 名称，不强行改写查询语句。
- 标准资源变量继续使用 APO 既有资源筛选接口获取候选值。
- 未识别的 label 保守按普通 label 变量处理。
- 在 PromQL 编辑界面提供轻量提示，帮助用户发现和使用支持的参数名。
- 不为了提示能力额外引入新的重型编辑器依赖。

## 非目标

- 不实现完整 Grafana 变量编辑器。
- 不自动重写用户 PromQL 的 label 名称。
- 不把未知 label 强行猜测为标准资源。
- 不在本阶段把自定义大盘 PromQL 输入框强制替换成 Monaco。
- 不改变后端数据组权限作为最终安全边界的原则。

## 当前代码事实

- 项目已经引入 `monaco-editor`、`@monaco-editor/react` 和 `monaco-promql`。
- 告警规则编辑已经使用 `MonacoEditorWrapper`。
- 自定义大盘编辑器当前 PromQL 输入仍使用 `Input.TextArea`。
- 自定义大盘现有标准资源变量判断依赖 `type` 是否为 `cluster/namespace/service/endpoint/pod/instance`。
- 标准资源变量候选值来自 `/api/v2/data/group/filter`。
- 普通 label / Grafana label 变量候选值来自 `/api/custom-dashboard/variables`。

## 推荐方案

采用“两层模型”：

```text
PromQL 原始 label
  clusterId / cluster_id / cluster
        |
        v
标准资源语义
  type: cluster
        |
        v
候选值来源
  /api/v2/data/group/filter
        |
        v
查询替换
  仍然替换原始 label: clusterId
```

也就是说，系统可以把 `clusterId` 识别成 `Cluster`，但真正生成或修改 PromQL matcher 时仍使用 `clusterId` 这个原始 label。

## Alias 表

建议集中维护标准资源 alias 表，作为变量发现、变量启用、候选值路由和 UI 提示的共同来源。

| 标准资源变量 | 推荐显示名 | Label aliases |
|---|---|---|
| `cluster` | Cluster | `cluster`, `cluster_id`, `clusterId`, `clusterID`, `cluster_name` |
| `namespace` | Namespace | `namespace`, `namespace_name`, `namespaceName` |
| `service` | Service | `svc`, `svc_name`, `service`, `service_name` |
| `endpoint` | Endpoint | `endpoint`, `endpoint_name`, `content_key`, `contentKey` |
| `pod` | Pod | `pod`, `pod_name` |
| `instance` | Instance | `instance`, `instance_name`, `node_name`, `nodeName` |

后续如果发现用户指标中存在稳定的新写法，应先加入 alias 表，再让它进入标准资源候选值链路。

## 变量定义模型

命中 alias 时，变量定义应同时保存标准资源类型和原始 label：

```json
{
  "name": "cluster",
  "label": "Cluster",
  "type": "cluster",
  "labels": ["clusterId"]
}
```

含义：

- `type: "cluster"` 决定候选值来源走 `/api/v2/data/group/filter`。
- `name: "cluster"` 决定 UI 和变量 token 的标准名称。
- `labels: ["clusterId"]` 决定这条 PromQL 中实际注入 matcher 的 label 名称。

如果同一个大盘里同时出现多个 cluster alias，例如：

```promql
up{cluster_id=~"$cluster"}
node_cpu_seconds_total{clusterId=~"$cluster"}
```

UI 仍只展示一个 `Cluster` 筛选框，但替换查询时应分别命中 `cluster_id` 和 `clusterId`。

## 候选值来源规则

### 标准资源变量

只要变量 `type` 是以下之一，就走 APO 标准资源接口：

- `cluster`
- `namespace`
- `service`
- `endpoint`
- `pod`
- `instance`

接口：

```text
POST /api/v2/data/group/filter
```

请求包含：

- `groupId`
- `category: "apm"`
- 当前全局时间范围
- `extra: "endpoint"` 用于 endpoint
- `extra: "instance"` 用于 pod / instance

前端将接口返回的资源树拍平成变量候选值，并按已选上游变量级联过滤。

### 普通 label 变量

未命中标准资源 alias 的 label，仍按普通 label 变量处理：

```text
GET /api/custom-dashboard/variables
```

请求包含：

- `type`
- `label`
- `variableName`
- `query`
- `variablesJson`
- `groupId`

### 静态变量

有本地 `options` 的变量不请求后端，直接使用 dashboard 定义里的静态候选值。

## PromQL 编辑提示

第一阶段建议保留 `Input.TextArea`，提供轻量提示，而不是立即把输入框替换成 Monaco。

### 支持参数提示

在 PromQL 输入区域旁边或下方显示“可用筛选参数”：

```text
可用筛选参数
Cluster       cluster_id / clusterId / clusterID / cluster_name
Namespace     namespace / namespace_name
Service       svc_name / service_name
Endpoint      endpoint / content_key
Pod           pod / pod_name
Instance      instance / instance_name / node_name
```

每一项可提供点击插入片段：

```promql
cluster_id=~"$cluster"
namespace=~"$namespace"
svc_name=~"$service"
endpoint=~"$endpoint"
pod=~"$pod"
instance_name=~"$instance"
```

### 已识别提示

解析当前 PromQL 后，展示系统已识别的资源变量：

```text
已识别
clusterId -> Cluster
namespace -> Namespace
svc_name -> Service
```

这能让用户确认候选值会来自 APO 标准资源接口。

### 未识别提示

对未命中 alias 的 label，提示其将作为普通 label 变量处理：

```text
未识别
clusterID2 将作为普通 label 变量处理
```

不要把相似 label 自动改成标准资源。可以后续提供“映射为 Cluster”的显式操作，但不能静默猜测。

## Monaco 取舍

当前项目已经有 Monaco 依赖，因此未来可以复用，但本需求不应依赖 Monaco 才能落地。

### 不建议当前阶段切换 Monaco 的原因

- 自定义大盘当前 PromQL 输入是轻量 TextArea。
- 本阶段核心问题是语义识别和提示，不是完整编辑器体验。
- 自动补全、inline warning、语法诊断会扩大实现和验证范围。
- 如果直接切换，容易把“变量识别”需求扩散成“PromQL 编辑器重做”。

### 适合后续引入 Monaco 的条件

当产品需要以下能力时，再把自定义大盘 PromQL 输入迁移到懒加载 Monaco：

- PromQL 语法高亮。
- label key 自动补全。
- `$cluster` / `$service` 变量 token 自动补全。
- alias 未命中的 inline warning。
- 查询语法错误位置提示。
- 更完整的 snippet 和 hover 说明。

如果后续引入，应复用现有 `MonacoEditorWrapper`，并确保只在进入 PromQL 编辑态时加载，不影响大盘列表和查看页首屏。

## 数据流

```text
用户输入 PromQL
      |
      v
解析 label keys
      |
      v
alias 表匹配
      |
      +-- 命中标准资源 alias
      |       |
      |       v
      |   生成标准资源变量 type/name
      |   保留原始 labels
      |       |
      |       v
      |   候选值走 /api/v2/data/group/filter
      |
      +-- 未命中 alias
              |
              v
          生成普通 label 变量
              |
              v
          候选值走 /api/custom-dashboard/variables
```

## 验收标准

- `clusterId`、`cluster_id`、`cluster` 都能识别为同一个 `Cluster` 筛选框。
- 标准资源筛选框候选值来自 `/api/v2/data/group/filter`。
- 查询替换使用 PromQL 中的原始 label，不强制改写为 canonical label。
- 未命中 alias 的 label 不走标准资源接口。
- 编辑器能展示支持的参数名和当前 PromQL 的识别结果。
- 用户可通过点击插入常用 matcher 片段。
- legacy `dataGroup` 变量仍不作为 dashboard 变量展示或保存。

## 测试建议

- alias 命中测试：`clusterId`、`clusterID`、`cluster_name` 均归一为 `cluster`。
- 多 alias 合并测试：同一大盘多个 cluster alias 只展示一个 `Cluster` 控件。
- 原始 label 保留测试：`clusterId` 查询注入 matcher 时仍使用 `clusterId`。
- 候选值路由测试：标准资源变量走 `/api/v2/data/group/filter`，普通 label 走 `/api/custom-dashboard/variables`。
- UI 提示测试：支持参数列表、已识别列表、未识别列表都能按当前 PromQL 更新。
- 回归测试：静态 options 变量不请求后端，legacy `dataGroup` 不展示、不保存。

## 推荐实施顺序

1. 抽出统一 alias 表，作为变量发现、候选值路由和提示文案的数据源。
2. 扩展变量发现逻辑，命中 alias 时生成标准资源变量并保留原始 labels。
3. 调整查询替换逻辑，确保原始 label 被用于 matcher 注入。
4. 增加轻量 PromQL 参数提示和识别结果展示。
5. 增加点击插入 snippet。
6. 补充行为测试和必要的 UI 测试。
7. 后续如需要，再单独评估 Monaco 化。

## 关键原则

- 语义识别要主动，PromQL 改写要克制。
- 候选值来源由标准资源类型决定，不由用户变量名猜测。
- 未知 label 宁可走普通 label 变量，也不要误归类。
- UI 提示应帮助用户写对，而不是替用户悄悄改。
