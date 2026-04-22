## Context

当前 `/agent` 页面中的 findings evidence tree 由商业仓前端渲染，但节点 payload 主要来自外部 AgentSRE 服务的 `/api/v1/conversations/findings-data`。现状存在三类问题：

- 节点类型和职责没有被显式建模，汇总节点、证据节点和占位节点混杂在一棵树里。
- investigations-first 与 dataplane fallback 的合并规则不完整，部分 dataplane builder 已经计算出状态和结论，但最终 payload 只回填了 `metrics`、`table` 或 `content`，没有同步状态。
- 前端 tree 到 detail panel 的字段映射不稳定，导致后端返回字段存在但右侧面板未消费或错误消费。

这个变更横跨三个边界：

- Commercial frontend 的 AGUI EvidencePanel / FindingsTree / NodeDetail 渲染层
- AgentSRE findings payload 组装层
- 两仓之间的协议文档与演进约束

为避免“实现提案只留在消费者仓”的漂移，本变更还对应 AgentSRE 仓中的 companion OpenSpec change：`openspec/changes/fix-findings-evidence-tree-contract`。商业仓 proposal 保留跨仓与 consumer 视角，AgentSRE proposal 负责 producer-side 实现计划。

目标不是重新设计诊断体验，而是把当前“会讲故事的诊断树”收敛成“可点击、可自证、可维护的证据树”。

## Goals / Non-Goals

**Goals:**

- 定义一个稳定的 findings tree 节点契约，覆盖顶层 payload、节点字段、状态语义和证据字段。
- 明确节点分类，区分汇总节点与证据节点，并定义哪些节点允许仅聚合，哪些节点必须返回真实证据。
- 统一 investigations-first 与 dataplane fallback 的最终投影规则，避免状态、结论和证据来源不一致。
- 规范前端点击节点后的右侧面板展示行为，确保后端返回的证据字段被正确消费。
- 建立节点覆盖率治理方式，允许显式“未支持/无数据”，但不允许静默返回默认正常占位节点冒充真实判断。

**Non-Goals:**

- 不重新设计 AGUI 页面整体布局、视觉风格或交互结构。
- 不为 findings API 引入新的公开版本号或全新路由前缀。
- 不在本次设计中扩展新的诊断领域，只聚焦现有节点和现有 evidence tree。
- 不要求一次性为所有节点引入全量深度分析能力；本次重点是契约正确性和证据可信度。

## Decisions

### Decision 1: 为 findings tree 定义规范化节点分类

节点将被分成三类：

- `summary`：汇总节点，可基于子节点状态或聚合结果生成总结，不要求自带原始证据
- `evidence`：证据节点，点击后必须能展示至少一种有效证据
- `unsupported`：当前树结构中保留位置，但尚未实现真实判断或真实证据；这类节点不得伪装成正常证据节点

选择这个方案，是因为当前最大问题不是单个字段缺失，而是节点语义不清。先把节点分类固定下来，才能判断哪些节点允许只有 `summaryConclusion`，哪些节点必须带 `metrics/table/logs/traces`。

替代方案：

- 继续沿用“所有节点长得差不多”的弱契约：实现成本低，但会继续放大默认模板和真实证据混淆的问题。
- 把整棵树都变成 evidence-only：表达简单，但会丢失父节点对诊断叙事和聚合判断的价值。

### Decision 2: 最终节点状态必须与最终证据来自同一来源

最终 payload 中每个节点的 `status`、`statusColor`、`summaryConclusion` 和右侧面板证据必须来自同一个最终投影结果。若 dataplane fallback 为节点补充了 `metrics/table/logs/traces`，且该 builder 产出了更具体的状态或总结，则最终节点必须同步使用这些值，而不是保留模板默认状态。

选择这个方案，是为了修复“树节点正常但右侧指标异常”的表达断裂。只要证据和状态不是同源，用户就无法判断系统到底想表达什么。

替代方案：

- 允许 investigations 覆盖状态、dataplane 仅补证据：实现简单，但会持续制造状态与证据矛盾。
- 完全由 investigations 决定状态：对 investigations 质量依赖过高，且无法利用 dataplane fallback 的真实检测结果。

### Decision 3: 证据节点必须满足最小证据完整性要求

所有 `evidence` 节点点击后必须满足以下之一：

- 有 `metrics`
- 有 `table`
- 有 `logs`
- 有 `traces`
- 或返回显式 `unsupported/unavailable` 说明

不再允许“默认 normal + 默认总结 + 无任何 detail payload”这种静默占位行为。

选择这个方案，是因为当前 `code_check`、`db_log_check`、`code_lock_check` 一类节点虽然在树上可见，但没有真实证据，对用户来说等同于虚假确定性。

替代方案：

- 继续保留空节点，只靠文案提示：对实现最省事，但会继续削弱诊断可信度。
- 直接删掉所有不完整节点：能立刻提升可信度，但会破坏当前树叙事，也不利于后续增量补齐。

### Decision 4: 前端右侧详情面板以节点契约为唯一展示入口

前端 FindingsTree 到 NodeDetail 的映射必须只消费规范化节点契约中的字段，不允许再依赖 `details` 这类后端未定义字段。详情面板的展示优先级为：

1. 标题与状态
2. 单一 canonical 文本槽位：优先 `summaryConclusion`，缺失时回退到 `content`
3. `abnormalMetrics`
4. `normalMetrics`
5. `table`
6. `logs`
7. `traces`

对 `unsupported` 节点，consumer 侧保留 `nodeKind` / `unsupportedReason` 元数据用于点击态与后续聚合判断，但为了让用户注意力聚焦在真正异常的数据上，树节点和详情头部可以沿用与 `normal` 相同的视觉样式，不额外抬升为 `warning`，也不强制插入 unavailable banner。

选择这个方案，是因为当前问题并不只在后端；前端字段映射错误会让已经返回的 `content` 丢失，进一步放大“不可信”的感受。同时 `summaryConclusion` 与 `content` 在很多节点上语义高度接近，详情面板若把两者同时铺开，只会制造重复感而不是增加解释力。

替代方案：

- 保持前端容错映射，兼容多个历史字段名：短期兼容性更强，但会让契约长期漂移。

### Decision 5: 采用“先契约、后覆盖率”的分阶段治理

实施顺序分两层：

- 第一层：统一字段契约、状态枚举、合并规则和前端映射
- 第二层：按节点覆盖率补齐空实现和弱实现节点

优先级建议：

- P0: `code_check`, `db_log_check`, `code_lock_check`, `Tree -> NodeDetail` 字段映射
- P1: `db_metrics_check`, `cpu_check`, `memory_check`, `file_check`
- P2: 父节点状态聚合与 unsupported 节点降级策略

选择这个方案，是为了避免一次性“重写整棵树”。先把契约拉齐，已有节点即使能力不完整，也能以明确方式暴露不足。

替代方案：

- 一次性补齐所有节点：理论上最好，但跨仓协调、验证和回归成本过高。

## Risks / Trade-offs

- [跨仓实现边界复杂] → Commercial frontend 与 AgentSRE backend 不在同一仓。需要把协议定义写清楚，并在任务中显式区分本仓改动、外部服务改动和文档同步。
- [节点视觉与语义分离带来的认知成本] → `unsupported` 节点在 consumer UI 中会按 normal-equivalent 样式展示，但它仍不是“检查通过”。需要把这个边界写进文档，避免后续实现把视觉策略误当成契约语义。
- [历史 investigations 数据质量不稳定] → investigations 事件可能只给摘要，不给完整证据；因此必须允许 fallback 产出最终节点，但要同步状态和结论。
- [前端兼容期存在双协议风险] → 在切换期间，前端需要对旧 payload 保持最低容错，但主展示逻辑必须收敛到新契约，避免长期双轨。
- [父节点聚合规则容易失真] → 若父节点继续靠模板文案或散落逻辑拼接，树仍然会讲不清故事。需要在后续实现中把聚合规则显式化。

## Migration Plan

1. 先冻结并文档化 canonical findings node contract。
2. 在 AgentSRE 侧按 canonical contract 修复最终 payload 投影与空节点策略。
3. 在前端侧切换 FindingsTree 与 NodeDetail 到 canonical contract 字段。
4. 联合验证代表性节点：
   - 纯 investigations 节点
   - dataplane fallback 节点
   - metrics 节点
   - table 节点
   - logs 节点
   - unsupported 节点
5. 文档更新后再扩大节点覆盖率改造。

回滚策略：

- 如果 AgentSRE contract 改动导致前端不可用，可暂时恢复前端的兼容映射，但不回退 proposal 中定义的 canonical contract。
- 如果某个节点无法在当前阶段补齐真实证据，应回退为显式 unsupported，而不是回退成默认 normal 模板。

## Open Questions

- 是否需要在节点契约中加入显式 `nodeKind` 或 `evidenceType` 字段，而不是仅依赖文档分类？
- 父节点最终状态应完全根据子节点聚合，还是允许 investigations 给出父节点级直接判断并覆盖聚合结果？
- `logs` 与 `traces` 的前端展示契约是否需要进一步标准化，避免不同节点返回不同数据形状？
