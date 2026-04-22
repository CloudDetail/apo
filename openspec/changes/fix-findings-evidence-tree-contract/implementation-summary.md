# Implementation Summary

## Final Scope

截至当前时间点，这个 change 已实际推进到以下范围：

- 已完成 canonical findings contract 的文档冻结
- 已完成 AgentSRE producer-side runtime contract enforcement 的第一批实现
- 已完成 AgentSRE `db_traces_check` 弱 builder 收口，不再返回伪造 sample evidence
- 已完成 AgentSRE rollout 兼容字段清理，最终 payload 不再暴露 node-level alias / internal helper fields
- 已完成商业仓 frontend consumer-side canonical consumption 适配
- 已补充后端 runtime contract 测试与商业仓前端 tree-to-detail 纯函数回归测试

当前仍未完成的范围：

- 无阻塞本 change 关闭的实现项；后续只剩归档动作与可选增强验证

因此，本 change 的实现、消费适配和代表性联调验证已完成。

## Actual Code Changes

### AgentSRE owner-side changes

外部 owner 仓 `/Users/keting/workspace/gitlabProject/AgentSRE` 已落地的关键实现包括：

- `api/controllers/findings_data.py`
  - findings payload 最终收口到 canonical node projection
  - 统一写回 `status`、`statusColor`、`summaryConclusion`、`content`
  - evidence section 收口为 `metrics`、`table`、`logs`、`traces`
  - dataplane 计算出的最终状态会写回最终节点
  - 不满足 canonical evidence contract 的 leaf 节点会降级为 explicit unsupported
  - `db_traces_check` 不再伪造 sample table；只有真实 trace table 才保留 evidence，否则降级为 unsupported
  - investigations 提供真实 `db_traces_check.table` 时，会同步写回 canonical `summaryConclusion` / `content`
  - 最终 payload emit 前会剥离 rollout 兼容字段和 internal helper fields，例如 `id`、`titleEn`、`source`、`type`、`has_investigations_data`、`_infra_*`、`_db_*`
- `tests/core/test_findings_runtime_contract.py`
  - 覆盖 investigations-only、dataplane-enriched、unsupported、empty-evidence regression

### Commercial consumer-side changes

当前仓 `/Users/keting/workspace/product/commercial/apo` 已落地的关键实现包括：

- `frontend/src/pro/api/conversation.ts`
  - 新增 findings canonical 类型定义
  - 新增 response / node runtime normalization
  - 收口前端消费状态到 `normal / warning / abnormal`
- `frontend/src/pro/api/findingsPresentation.ts`
  - 抽出 consumer 侧纯状态归一化与 unsupported reason 映射
  - 保留 producer 已明确写回的 `unsupported + normal`，不再由前端强制抬升为 `warning`
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.ts`
  - 新增 tree-to-detail 共享 view-model / helper
  - 收口树节点点击态、detail section presence 判断、unsupported 元数据传递
  - 新增 consumer-side `visualStatus` 派生：`unsupported` 节点保留原始 `status` 语义，但在 UI 中按 `normal` 等价视觉处理
  - 右侧文本区收口为单一 canonical 文本槽位：优先 `summaryConclusion`，缺失时回退到 `content`
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/detailSectionPolicy.ts`
  - 抽出 detail text visibility 纯函数
  - 保持 canonical payload 不变，只在 consumer 侧收敛重复提示
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/Tree.tsx`
  - 从旧的 `details` 主映射切换到 canonical `content`
  - FindingsTree 直接消费 canonical node payload
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/FindingsNode.tsx`
  - 状态视觉语义收口到 `normal / warning / abnormal`
  - `unsupported` 节点在树上沿用 `normal` 视觉样式
  - 点击 payload 显式传递 `nodeKind`、`unsupportedReason` 与 `visualStatus`
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/FindingsTopology.tsx`
  - 自动聚焦、离屏方向提示、边样式和 MiniMap 改为消费 `visualStatus`
  - `unsupported` 节点不再参与 warning attention flow
  - findings tree 布局改为消费抽出的纯函数 layout helper，不再把所有“缺失子节点”情况直接交给 `d3.tree()` 压缩
  - 对 `dependency_check`、`db_query_check` 这类固定诊断步骤补充 invisible reserved child slot，节点本身无子节点时仍保留对应子树带宽，避免同级其他节点的子树向上补位造成层级误读
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsLayout.ts`
  - 新增 findings tree 纯布局 helper
  - 在布局阶段注入仅用于占位的 placeholder children，最终渲染时过滤掉 placeholder node / edge
  - 保持现有 ReactFlow 交互不变，只修正固定诊断节点缺失子树时的 subtree positioning
- `frontend/src/pro/components/AGUIChat/components/EvidencePanel/NodeDetail.tsx`
  - 右侧详情面板改为按 canonical fields 渲染
  - 空 section clean hide
  - `unsupported` 节点详情头部沿用 `normal` 视觉样式，不额外插入 unavailable banner
  - `unsupportedReason` 保留在 consumer state 中用于内部判断，但不在右侧详情面板直接展示
  - 修复右侧详情面板异常指标区图标 import 回归，避免点击带 abnormal metrics 的节点时触发 `ReferenceError`
- `frontend/src/pro/components/AGUIChat/components/EvidencePanel/*`
  - logs/table/detail 类型改为依赖 canonical contract
- `frontend/src/pro/components/AGUIChat/components/StatusIcon.tsx`
  - 去除 `info / muted` 作为主消费状态
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts`
  - 新增前端 mapping regression test
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsLayout.test.ts`
  - 新增固定诊断节点空槽位布局回归测试，覆盖 `dependency_check` / `db_query_check` 无子节点时 sibling subtree 不补位
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/detailSectionPolicy.test.ts`
  - 新增 canonical 文本槽位优先级回归测试
- `frontend/src/pro/api/findingsPresentation.test.ts`
  - 新增 unsupported normal 状态保留与 canonical reason 映射回归测试
- `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsVisualStatus.test.ts`
  - 新增 unsupported 视觉状态回归测试

## Document Updates

### Updated change docs

- `openspec/changes/fix-findings-evidence-tree-contract/design.md`
- `openspec/changes/fix-findings-evidence-tree-contract/specs/findings-evidence-tree-contract/spec.md`
- `openspec/changes/fix-findings-evidence-tree-contract/implementation-summary.md`

### Updated system docs in commercial repo

- `docs/openspec-system/modules/conversational-diagnosis-and-reporting/contracts.md`
- `docs/openspec-system/modules/conversational-diagnosis-and-reporting/frontend.md`

### Related owner-side docs in AgentSRE

以下文档位于 owner 仓，记录 producer-side canonical contract：

- `/Users/keting/workspace/gitlabProject/AgentSRE/docs/openspec-backend/modules/investigation_findings/contracts.md`
- `/Users/keting/workspace/gitlabProject/AgentSRE/docs/openspec-backend/modules/investigation_findings/backend.md`
- `/Users/keting/workspace/gitlabProject/AgentSRE/openspec/changes/fix-findings-evidence-tree-contract/tasks.md`
- `/Users/keting/workspace/gitlabProject/AgentSRE/openspec/changes/fix-findings-evidence-tree-contract/implementation-summary.md`

## Design Deviations

相对 proposal / design / tasks，当前实现存在以下偏差或折中：

- 商业仓前端回归测试采用了“纯函数 mapping test”，而不是直接补完整组件测试基线
  - 原因是当前仓没有现成的 findings 组件测试基础设施
  - 这次优先覆盖 contract-sensitive 的 tree-to-detail mapping、status normalization、unsupported rendering decision
  - 这不等同于组件层级的更高保真渲染覆盖；本次已通过真实 thread live validation 完成 `4.3`
- unsupported 节点的 consumer 视觉策略在实现收尾中收敛为：
  - 保留 `nodeKind` / `unsupportedReason` 元数据与 clickable 交互
  - 树节点、拓扑聚焦和详情头部按 `normal` 等价视觉处理
  - 不额外插入 unavailable banner，避免分散用户对 abnormal / warning 证据的注意力
  - 这个策略已同步回 consumer-side contract 文档
- 商业仓 consumer UI 在后续收尾中进一步收紧为：
  - 无论节点类型如何，右侧文本区只渲染一个 canonical 文本块
  - 当 `summaryConclusion` 与 `content` 同时存在时，consumer 优先展示 `summaryConclusion`
  - 这样做是为了降低详情面板的重复感，同时不改 producer contract
- 商业仓前端单测执行仍依赖仓内现有的 JSX loader 配置
  - 本轮尝试用 `tsx --test` 直接重跑时，命中了仓内既有的 `.js` JSX 解析问题
  - 这不影响本次 producer payload contract 清理本身，但说明 consumer test runner 后续仍值得单独整理

## Compatibility And Rollout

- rollout 采用 contract-first 顺序：
  - 先冻结 contract
  - 再落 AgentSRE runtime enforcement
  - 再切商业仓前端 canonical consumption
  - 最后做代表性线程联调与 backlog 收口
- 过渡期兼容策略：
  - AgentSRE 可以短期继续返回历史 alias
  - 商业仓前端不再把这些 alias 视为主 contract
  - 若 canonical 字段与 alias 同时存在，以 canonical 字段为准
- unsupported 策略：
  - 当前批次保持 clickable
  - tree / topology / detail header 使用 `normal` 等价视觉样式
  - 右侧面板不直接展示 `unsupportedReason` code
  - 若 producer 已提供 `summaryConclusion` 或 `content`，仍按 canonical 文本槽位展示
  - 不再允许以默认正常占位节点冒充真实 evidence 节点
- 跨仓同步要求：
  - producer-side contract 以 AgentSRE owner-side 文档为准
  - 商业仓只维护 consumer-side contract 与渲染行为

## Validation Snapshot

已完成的验证：

- AgentSRE runtime contract tests 通过
- 商业仓 findings contract 纯函数测试通过
- 商业仓 unsupported visual status 纯函数测试通过
- 商业仓 detail text policy 纯函数测试通过
- 商业仓 consumer status / unsupported reason 纯函数测试通过
- 2026-04-17 基于真实 thread `thread_1776303854393_k00vr4nsh` / `ts-station-service` 完成 live payload 验证
  - `/api/v1/conversations/investigations` 中原始 findings 树保留 `infra_check=warning`
  - `/api/v1/conversations/findings-data` 最终 payload 输出 19 个 canonical nodes，状态全部收口到 `normal / warning / abnormal`
  - `infra_check` 在最终 payload 中保持 `warning`
  - `network_check` 在最终 payload 中保留 `warning` 且带真实 `metrics`
  - `self_log_check` 从原始 findings 事件中的 `normal` 弱节点，降级为 `warning + unsupported + no-valid-evidence`
  - 通过前端 canonical adapter / topology helper 复核，未出现 invalid status 或 `statusColor != status` 的漂移
  - 在 `db_traces_check` builder 修复后复核同一真实 thread，确认该节点不再返回伪造 sample table，而是收口为 `warning + unsupported + no-valid-evidence`
  - 在兼容字段清理后再次复核同一真实 thread，确认节点层仅保留 canonical 字段，`id/titleEn/source/type/has_investigations_data/_infra_*/_db_*` 已从最终 payload 移除

尚未完成的验证：

- 若要获得一条新的 consumer-side 自动化证据，需要先修复商业仓当前 `tsx --test` 直接导入 `request.js` 时的 JSX loader 问题

## Follow-Ups

- 归档 `fix-findings-evidence-tree-contract`
- 在仓内既有 TypeScript / JSX test runner 问题清理后，补更高层级的 frontend render/integration coverage

## Current Status

当前 change 的核心 contract、runtime、frontend 和代表性 live validation 已完成，任务层面可视为完成并进入归档准备状态。
