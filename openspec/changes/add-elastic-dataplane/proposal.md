## Why

当前商业版 APO 的 `/dataplane` 主流程还没有正式的 `elastic` 数据平面类型，导致已经把日志沉淀在 Elastic / ELK 体系中的用户，无法通过当前 dataplane 入口完成统一接入、测试、保存和 AgentSRE 日志消费。

本 change 目标已收窄为 **Elastic Log-only 接入**。完整 Elastic Observability / APM 能力，包括 `Trace`、`Metrics`、`Topology`、`ServiceInstance`、`RedMetric`，不作为本 change 的验收目标，后续单独立项。

## What Changes

- 在 `/dataplane` 主流程中新增正式的 `elastic` 数据平面类型，支持 Elastic Cloud 和自建 Elastic Stack 的 Elasticsearch endpoint。
- Elastic 当前只开放 `Log` capability，前端主流程只要求 `URL` 和 `API Key`，默认使用 `authType=apiKey`、`logIndexPattern=logs-*`。
- 在 `apo-dataplane` 仓新增 `elastic` provider，实现 provider 注册、连接检查、日志查询，以及 AgentSRE 所需的 `podName` / `k8sNamespace` 日志过滤。
- 测试连接支持 Log-only 场景：能访问日志索引但当前时间窗无日志，或日志缺少 AgentSRE 服务过滤字段时，返回 warning，不阻塞保存。
- 保留 `Trace`、`Metrics`、`Topology`、`ServiceInstance`、`RedMetric` 作为类型上限和后续演进方向，但本 change 不要求 UI 配置、保存、coverage 或联调这些能力。

## Capabilities

### New Capabilities
- `elastic-dataplane`: 定义 `/dataplane` 中 `elastic` 类型的 Log-only 可见行为、配置参数、部署形态、字段默认值和 AgentSRE 日志消费边界。
- `dataplane-capability-governance`: 本 change 仅保留 Log-only 测试连接兼容语义；完整实例级 capability 选择、coverage 聚合治理和保存阻塞规则移入后续工作。

### Modified Capabilities

None.

## Impact

- Code:
  - APO 商业仓 `/dataplane` 页面与 `frontend/src/pro/**` 相关配置 UI
  - APO 商业仓 `backend/sqlscripts/data_plane_type.sql` 及升级脚本
  - 外部仓 `apo-dataplane: pkg/datasource/**`
- APIs:
  - `/api/dataplane/type/list`
  - `/api/dataplane/checkDatasource`
  - `apo-dataplane` `/datasource/checkProvider`
  - `apo-dataplane` `/datasource/queryLogs`
- Systems:
  - APO 商业版 `/dataplane` 治理链路
  - `apo-dataplane` provider 注册、Log 能力检查与日志查询分发链路
- Dependencies:
  - Elastic Cloud / Elasticsearch API
  - API Key 认证
  - ECS / OTel / Kubernetes 常见日志字段约定
- Governance:
  - APO 前端主要位于 `frontend/src/pro/**`，属于商业仓独有范围
  - 本轮不改 APO 共享后端候选路径；后续若恢复完整 capability governance，需要重新确认是否回 OSS 以及目标分支
