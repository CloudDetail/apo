## Final Scope

本轮落地范围已明确收窄为 **Elastic Log-only 接入**，聚焦 AgentSRE 会消费的日志链路：APO 负责把 `elastic` 作为正式 dataplane type 种子暴露出来，`apo-dataplane` 负责注册 `elastic` provider，并让启用 `Log` capability 的 Elastic 实例可以通过 `/datasource/queryLogs` 被查询。

未在本轮完成 APO 共享后端 capability 治理改造、完整 UI capability 选择、结构化测试结果 UI 展示，以及 full APM 联调。

## Actual Code Changes

- APO 商业仓：
  - `backend/sqlscripts/data_plane_type.sql` 新增 `elastic` 类型，类型上限 capability 为 `RedMetric`、`ServiceInstance`、`Topology`、`Log`、`Trace`、`Metrics`。
  - `backend/sqlscripts/upgrade/data_plane_type/1.14.0-upgrade-mysql.sql`、`1.14.0-upgrade-postgres.sql`、`1.14.0-upgrade-sqllite.sql` 新增幂等升级脚本，其中 `sqllite` 命名与现有数据库配置常量保持一致。
  - `frontend/src/pro/components/Dataplane/DataplaneGuide/guideMd.ts` 补充 Elastic Log-only 配置指南，说明主流程只需要 URL / API Key，并解释 AgentSRE 服务日志分析依赖常见 Kubernetes 日志字段。
  - `frontend/src/pro/constants.ts` 为 `elastic` dataplane type 复用现有 Elastic 图标。
  - `frontend/src/pro/components/Dataplane/DataplaneConfigForm/index.tsx` 将 `success:true + message` 的 check 结果展示为 warning，避免 Elastic 有日志但缺少关键过滤字段时被 UI 误判成连接失败。
  - Elastic 表单只对 `elastic` 类型强制 `Log` capability；非 Elastic 数据源继续沿用原有 `capabilitySpec` 行为，避免改动 Datadog、SigNoz、Loki 等既有保存语义。
  - `frontend/public/locales/en/pro/agentDatasource.json`、`frontend/public/locales/zh/pro/agentDatasource.json` 补充测试 warning 文案。
- `apo-dataplane` 仓：
  - `pkg/datasource/provider/model.go` 新增 `elastic` datasource 常量和结构化 check response 模型。
  - `pkg/datasource/datasource.go` 注册 Elastic provider factory、checker 和 capability 上限。
  - `pkg/datasource/manager.go` 和 `pkg/httpserver/check_provider.go` 支持 capability-aware 结构化 check response，同时保留顶层 `success/message` 兼容语义。
  - `pkg/datasource/elastic/provider.go` 新增 Elastic 连接、认证、参数解析、Log 查询和 capability check；日志过滤支持 AgentSRE 传入的 `podName` / `k8sNamespace`，并覆盖 ECS 与 OTel 常见字段候选。
  - Elastic Log check 改为先读取日志样本，再检测是否存在 AgentSRE 可用于服务实例过滤的 pod 与 namespace 字段；能读到日志但缺少关键过滤字段时返回 `warning`，不阻塞保存。
  - Elastic Log check 的样本查询限定在最近 30 分钟验证窗口内，避免历史日志让测试连接误判为当前可用。
  - `pkg/datasource/elastic/provider_test.go` 覆盖日志-only warning、认证推断、Log 查询转换、AgentSRE pod/namespace 过滤、自定义字段映射，以及“有日志但缺少 pod 或 namespace 字段”的诊断 warning。
  - `frontend/src/pro/components/Dataplane/DataplaneConfigForm/elasticCopy.test.ts` 覆盖 Elastic 主流程文案、默认参数说明、图标映射，以及非 Elastic capability 行为不被本次变更改动。

## Document Updates

- 已更新需求文档：`docs/requirements/2026-04-23-elastic-dataplane-requirements.md`，将目标收窄为 Elastic Log-only。
- 已更新 OpenSpec change docs：`proposal.md`、`design.md`、`specs/elastic-dataplane/spec.md`、`specs/dataplane-capability-governance/spec.md`、`tasks.md`。
- 已更新本文件：`openspec/changes/add-elastic-dataplane/implementation-summary.md`。
- 已更新 Elastic 配置指南：`frontend/src/pro/components/Dataplane/DataplaneGuide/guideMd.ts`。
- 未更新 `docs/openspec-system/**`：本轮没有改变既有仓间边界或系统级源码归属，仍是 APO 管理 dataplane 配置，`apo-dataplane` 执行 provider 查询。

## Design Deviations

- 本轮正式把验收目标从完整 Elastic dataplane 收窄为 Elastic Log-only。
- `Trace`、`Topology`、`ServiceInstance`、`RedMetric`、`Metrics` 的代码骨架已在 `apo-dataplane` 中存在，但 full APM 端到端联调和 APO UI/API 治理项未在本轮收口，因此已移入 follow-up。
- APO 共享后端候选路径未改动；后续如果要做实例级 capability 持久化和结构化 check response 的 API contract，需要先确认是否回 OSS 以及目标 OSS 分支。

## Compatibility And Rollout

- 旧的 `checkProvider` 调用方仍可继续读取顶层 `success/message`。
- 新的结构化结果通过 `connection` 和 `results` 暴露 capability 级状态。
- Elastic Log-only 的主流程保持简单：标准 Elastic / Kubernetes / OTel 日志默认只需要 URL 和 API Key；非业务日志或缺少 pod / namespace 字段的日志会给出 warning，提示 AgentSRE 可能无法按服务过滤。
- AgentSRE 不需要改代码；只要 `config/dev/agentsre.yaml` 的 `apo_dataplane_url` 指向包含本实现的 dataplane，且 APO 已保存启用 `Log` capability 的 Elastic 实例，AgentSRE 的 LogAnalysisTool 会继续调用 `/datasource/queryLogs` 并命中 Elastic provider。
- 本地或部署环境需要让 `apo-dataplane` 重新加载 APO 数据库中的 dataplane 实例，provider list 才会出现 `elastic`。

## Follow-ups

- 完成 APO `/dataplane` 的实例级 capability 选择、保存和 coverage 聚合治理。
- 完成结构化 check result 在 APO UI 中的 warning / failed 展示和保存阻塞规则。
- 开放 Elastic Basic Auth、自定义 log index、environment 和字段映射 UI。
- 对 Elastic full APM capability 做真实环境联调，再决定是否勾选 4.4、4.5、5.2、5.3。
