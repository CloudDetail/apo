## 1. Scope And Governance

- [x] 1.1 收窄本 change 目标为 Elastic Log-only 接入，full APM capability governance 移入后续 change。
- [x] 1.2 明确本轮不改 APO 共享后端候选路径；后续若恢复完整 capability governance，再确认是否回 OSS 以及目标分支。
- [x] 1.3 更新 proposal、design、spec 和 implementation summary，避免继续按完整 Elastic dataplane 验收。

## 2. APO Dataplane UI

- [x] 2.1 在 `/dataplane` 类型列表中展示 `Elastic`。
- [x] 2.2 为 Elastic 配置抽屉提供 Log-only 主流程，只要求 `Name`、`URL`、`API Key`。
- [x] 2.3 Elastic 保存时固定启用 `Log` capability，并默认 `authType=apiKey`、`logIndexPattern=logs-*`。
- [x] 2.4 测试连接中将 `success=true + message` 展示为 warning，避免 Log-only warning 被误判为失败。
- [x] 2.5 补充 Elastic Log-only 接入说明和 i18n 文案。

## 3. APO Dataplane Backend

- [x] 3.1 在 `backend/sqlscripts/data_plane_type.sql` 和对应升级脚本中新增 `elastic` 类型种子，并定义未来能力上限 capability 与参数 schema。
- [x] 3.2 保持现有 `success/message` checkDataSource 兼容契约；结构化 capability 结果不作为本 change 的 APO 后端验收项。
- [x] 3.3 保持现有 `/api/dataplane/capabilities` 行为；完整 coverage 聚合治理不作为本 change 的验收项。

## 4. `apo-dataplane` Elastic Provider

- [x] 4.1 在 `apo-dataplane: pkg/datasource/**` 中注册 `elastic` datasource，并补齐 provider factory、checker 和参数解析入口。
- [x] 4.2 实现 Elastic 共享客户端、认证处理和基础参数模型，支持 Elasticsearch endpoint 与 API Key；Basic Auth 作为 provider 预留能力，不在 APO UI 主流程开放。
- [x] 4.3 实现 Elastic `LogProvider` 与默认字段映射，并支持 AgentSRE 传入的 `podName` / `k8sNamespace` 过滤。
- [x] 4.4 扩展 `checkProvider` 逻辑，输出 Log-only warning，并保留顶层 `success/message` 兼容语义。
- [x] 4.5 为 `apo-dataplane` Elastic provider 补充针对参数解析、Log check、Log 查询和 AgentSRE 过滤语义的测试。

## 5. Verification And Documentation

- [x] 5.1 运行前端 Elastic 文案/行为测试，验证 Log-only 主流程约束。
- [x] 5.2 运行 `apo-dataplane` Elastic provider 测试，确认 Elastic 包测试通过。
- [ ] 5.3 联调验证真实 Elastic Log-only 实例：只启用 `Log` 时可测试、可保存、可查询日志，且 AgentSRE 可按 pod / namespace 查询。
- [x] 5.4 更新 `openspec/changes/add-elastic-dataplane/implementation-summary.md`，记录真实落地范围、偏差、兼容策略和 follow-up。
- [x] 5.5 评估 `docs/openspec-system/**`：本轮目标收窄且未改变系统级源码归属，无需更新。

## Follow-up Change Candidates

- [ ] F1 完整实例级 capability 选择、保存、coverage 聚合和结构化 check result UI。
- [ ] F2 Elastic Basic Auth、自定义 log index、environment 和字段映射 UI。
- [ ] F3 Elastic Trace / Metrics / Topology / ServiceInstance / RedMetric provider 真实环境联调与验收。
