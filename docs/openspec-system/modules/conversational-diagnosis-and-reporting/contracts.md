# Conversational Diagnosis And Reporting Contracts

## Ownership Rule

本模块消费的 findings tree / evidence payload 的 producer-side canonical contract 不在当前仓，而在 AgentSRE：

- [AgentSRE investigation_findings/contracts.md](/Users/keting/workspace/gitlabProject/AgentSRE/docs/openspec-backend/modules/investigation_findings/contracts.md)

当前商业仓只记录 consumer-side contract，也就是：

- 前端实际消费哪些字段
- 树节点点击后如何把字段传给右侧详情面板
- 当外部 payload 缺字段或节点未支持时，前端如何降级展示

因此，两边的职责必须分开：

- AgentSRE 负责节点结构、状态枚举、节点分类、merge 规则和 evidence 完整性
- Commercial frontend 负责按 canonical fields 渲染，不扩散 undocumented alias

## External Producer Boundary

以下接口由外部 AgentSRE 提供，当前仓是消费者，不是契约主导方：

| API | Role |
|---|---|
| `/ai/api/v1/conversations/findings-data` | 会话级 findings payload |
| `/ai/api/v1/conversations/investigations` | findings 调查树事件重建 |
| `/ai/api/v1/conversations/messages` | Agent 历史消息与 AG-UI 事件来源 |

如果这些接口的节点结构、状态值或 evidence schema 发生变化，应优先更新 AgentSRE 文档，再回到当前仓调整消费逻辑。

## Consumer-Side Expectations

当前仓前端对 findings node 的稳定消费顺序应理解为：

1. `title`
2. `status`
3. 单一 canonical 文本槽位：优先 `summaryConclusion`，缺失时回退到 `content`
4. `metrics`
5. `table`
6. `logs`
7. `traces`

说明：

- `details` 一类历史 alias 不应再被视为主展示字段
- 缺失 section 时应直接省略，而不是构造空壳内容
- `unsupported` 节点必须保留分类语义，但 consumer UI 可以采用与 `normal` 相同的视觉样式，避免分散用户对异常数据的注意力
- `unsupportedReason` 仍保留为 consumer-side metadata，用于点击态和后续扩展，但不要求占用右侧详情的主注意力层
- 前端不得把 producer 已明确给出的 `unsupported + normal` 自动抬升成 `warning`；consumer 只负责归一化状态词汇，不负责重写最终状态语义

## Compatibility And Unsupported Policy

Batch 1 冻结的 consumer 侧策略如下：

- AgentSRE 在过渡期可以继续返回历史 alias，但当前仓不把这些字段视为稳定 contract
- 如果 canonical 字段和 alias 同时存在，以 canonical 字段为准
- 当前 rollout 中，`unsupported` 节点保持 clickable，而不是静默退化成 summary-only 节点
- 当前 rollout 中，consumer 可以为 `unsupported` 节点派生 `normal` 等价的视觉状态，但不得丢失 `nodeKind: unsupported`
- 点击任意节点时，右侧面板只渲染一个 canonical 文本块：优先 `summaryConclusion`，缺失时回退到 `content`
- 点击 `unsupported` 节点时，右侧面板继续按 canonical 文本槽位渲染；如果 producer 未提供 `summaryConclusion` / `content`，consumer 不强制插入 warning 或 unavailable banner
- 已冻结的 canonical `unsupportedReason` code 应映射为本地化说明，当前至少包括：`no-downstream-services`、`red-metrics-unavailable`、`red-query-failed`、`no-valid-evidence`
- 只有在 producer contract 明确变更后，当前仓才允许改变 unsupported 节点的交互方式

## Tree To Detail Mapping Rule

`/agent` 页面里的 findings tree 与右侧详情面板必须遵守以下约束：

1. 树上的节点状态和右侧详情必须来自同一份 canonical node data
2. 点击 evidence 节点时，右侧面板应展示真实 evidence payload，而不是仅展示模板默认文案
3. summary 节点可以只展示聚合结论
4. `unsupported` 节点在当前 rollout 中保持可点击；consumer UI 可以将其视觉优先级下调到与 `normal` 等价，但不得把它当作已完成且结果正常的 evidence check

如果出现“树节点正常但右侧 evidence 显示异常”的情况，应优先判断为 producer contract 或 merge projection 失真，而不是前端自行修饰结果。

## Consumer Responsibilities

当前仓对 findings contract 的责任边界如下：

| Area | Responsibility |
|---|---|
| `frontend/src/pro/api/conversation.ts` | 保持 API type / adapter 与 canonical contract 对齐 |
| `frontend/src/pro/components/AGUIChat/components/FindingsTree/*` | 只基于 canonical node fields 构造树和点击态，并可派生不改变语义的 visual status |
| `frontend/src/pro/components/AGUIChat/components/EvidencePanel/*` | 只渲染 canonical detail fields，并对缺失 section 做 clean hide |

这些职责不包括重新解释 node status，也不包括为缺失 evidence 的节点补造业务结论。

## Drift Handling

当商业仓与 AgentSRE 行为不一致时，按以下顺序处理：

1. 先检查 AgentSRE `contracts.md` 是否已经定义该字段或语义
2. 如果 AgentSRE 文档未定义，先补 producer-side contract
3. 如果 AgentSRE 文档已定义但 payload 未满足，修 AgentSRE 实现
4. 如果 AgentSRE payload 已满足但前端未正确消费，修当前仓消费逻辑

不要在当前仓文档中重新发明一套独立 findings contract。

## Related Documents

- [module_overview.md](/Users/keting/workspace/product/commercial/apo/docs/openspec-system/modules/conversational-diagnosis-and-reporting/module_overview.md)
- [frontend.md](/Users/keting/workspace/product/commercial/apo/docs/openspec-system/modules/conversational-diagnosis-and-reporting/frontend.md)
- [dependencies.md](/Users/keting/workspace/product/commercial/apo/docs/openspec-system/modules/conversational-diagnosis-and-reporting/dependencies.md)
- [AgentSRE investigation_findings/module_overview.md](/Users/keting/workspace/gitlabProject/AgentSRE/docs/openspec-backend/modules/investigation_findings/module_overview.md)
- [AgentSRE investigation_findings/contracts.md](/Users/keting/workspace/gitlabProject/AgentSRE/docs/openspec-backend/modules/investigation_findings/contracts.md)
