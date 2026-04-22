## Why

当前 findings evidence tree 试图向用户表达一条“可解释的故障排查链路”，但实际前后端契约并不闭环：部分节点只有默认模板状态，没有真实取数或真实判断；部分节点有结论文案却没有可展示证据；还有一部分节点拿到了异常指标却没有把状态同步回树节点。结果是树结构看起来完整，但用户点击节点后，右侧证据面板并不总能回答“为什么这样判断”。

这个问题已经影响诊断可信度和可维护性。现在需要把 findings tree 从“半结构化摘要视图”收敛成“每个非汇总节点都能自证的证据树”，并明确 AgentSRE payload 与前端渲染之间的稳定契约。

## What Changes

- 定义 findings evidence tree 的稳定节点契约，明确顶层 payload、节点必填字段、证据字段和状态枚举。
- 区分汇总节点与证据节点，要求非汇总节点在可点击时必须返回至少一种有效证据载荷，而不是仅返回默认模板文案。
- 规范 AgentSRE investigations-first 与 dataplane fallback 的合并规则，确保 dataplane builder 产出的状态、结论和证据能够一致投影到最终节点。
- 识别并治理当前占位或弱实现节点，明确哪些节点允许仅聚合，哪些节点必须真实取数和真实判断。
- 规范前端 findings tree 到右侧详情面板的字段映射，避免后端返回字段存在但前端未消费或错误消费。
- 为跨仓边界补充系统文档，明确 commercial frontend 与 AgentSRE backend 的协议责任和演进规则。
- 在 AgentSRE 仓同步维护一个同名 companion OpenSpec change，承载 producer-side contract、merge 规则和后端回归覆盖。

## Capabilities

### New Capabilities
- `findings-evidence-tree-contract`: 定义 findings tree 的节点层级、状态语义、证据契约、前后端字段映射，以及“可点击节点必须可自证”的行为要求。

### Modified Capabilities
- None.

## Impact

- Frontend:
  - `frontend/src/pro/components/AGUIChat/components/EvidencePanel/*`
  - `frontend/src/pro/components/AGUIChat/components/FindingsTree/*`
  - `frontend/src/pro/api/conversation.ts`
- External backend boundary:
  - AgentSRE `GET /api/v1/conversations/findings-data`
  - AgentSRE `GET /api/v1/conversations/investigations`
  - AgentSRE node builder and payload assembly logic in `api/controllers/findings_data.py`
- Documentation:
  - `docs/openspec-system/modules/conversational-diagnosis-and-reporting/*`
  - AgentSRE `docs/openspec-backend/modules/investigation_findings/*`
  - AgentSRE `openspec/changes/fix-findings-evidence-tree-contract/*`
- Cross-system behavior:
  - findings node status normalization
  - node evidence completeness rules
  - frontend right-panel rendering consistency
