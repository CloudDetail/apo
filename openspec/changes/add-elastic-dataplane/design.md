## Context

当前商业版 APO 已经把 `/dataplane` 作为正式的数据平面治理入口，但系统中没有正式的 `elastic` dataplane type，Elastic / ELK 仍停留在旧接入常量层，未进入当前主链路。

本 change 目标已收窄为 **Elastic Log-only 接入**：APO 负责暴露 `elastic` 类型和简化配置入口，`apo-dataplane` 负责注册 Elastic provider、测试连接和日志查询。完整 Elastic Observability / APM 能力不在本 change 内验收。

## Goals / Non-Goals

**Goals:**

- 在 `/dataplane` 中新增正式的 `elastic` 类型，并让用户能够创建、编辑、删除 Log-only Elastic 数据平面实例。
- 前端主流程只开放 `URL` 和 `API Key`，默认保存 `authType=apiKey`、`logIndexPattern=logs-*`、`capability=["Log"]`。
- 在 `apo-dataplane` 中新增 `elastic` provider，接入 `Log` 查询主链路。
- 测试连接保留 `success/message` 兼容语义，并允许 Log-only warning 不阻塞保存。
- 让 AgentSRE 继续通过现有 `/datasource/queryLogs` 查询 Elastic 日志，支持 `podName` / `k8sNamespace` 过滤。

**Non-Goals:**

- 不在一期实现完全任意字段映射系统。
- 不在本 change 实现前端可配置字段映射；字段映射只在 provider 参数模型中预留，后续再开放 UI。
- 不在本 change 支持 Basic Auth 的前端主流程；Basic Auth 作为 provider 能力预留，后续再开放 UI。
- 不在本 change 完成 `Trace`、`Metrics`、`Topology`、`ServiceInstance`、`RedMetric` 的 UI、保存、coverage 或端到端联调。
- 不在一期实现 Kibana 深链、Alert / Incident 联动、Elastic ML 联动或旧 `/logs/full` 体系迁移。
- 不在本次设计里替共享后端治理做分支决策；该问题作为实现前 gate 单独确认。
- 不在一期尝试覆盖所有历史 Elastic 版本或非标准部署差异。

## Decisions

### 1. 使用单一 `elastic` 类型，本 change 只启用 Log

Elastic 接入采用单一 `elastic` dataplane type，而不是拆成 `elastic-log`、`elastic-apm` 等多个类型。类型模板可以保留未来能力上限，但本 change 的 UI 和保存行为只启用 `Log`。

这样可以先把 AgentSRE 真实需要的日志链路接通，同时避免现在就改 APO 共享后端 capability governance。完整能力选择后续再做。

### 2. APO 商业仓负责治理与显示，`apo-dataplane` 负责 provider 实现

APO 商业仓负责 `/dataplane` 类型展示、Log-only 配置表单和类型种子；`apo-dataplane` 负责 `elastic` provider 注册、连接校验和 Log 查询能力实现。

这样符合当前仓间职责边界，也避免把 Elastic 查询逻辑直接塞进 APO。替代方案是在 APO 仓里新增 Elasticsearch 直连逻辑，但这会绕过现有 dataplane provider 体系，并与已有 Datadog / SigNoz 模式不一致。

### 3. `checkDataSource` 契约本轮只消费兼容字段

当前 APO 与 `apo-dataplane` 的测试连接返回只有 `success` 和 `message`。本 change 中 `apo-dataplane` 可以返回结构化结果，但 APO UI 只消费顶层 `success/message`，并把 `success=true + message` 展示为 warning。

完整 capability 级结果展示、保存阻塞规则和 Swagger 契约同步移入后续工作。

### 4. 参数模型先采用 Log-only 简化主流程

前端主流程只展示 `URL` 和 `API Key`，并自动补齐 `authType=apiKey`、`logIndexPattern=logs-*`。字段映射、Basic Auth、自定义索引和多 capability 参数仍由 provider 模型预留，但不在本 change 的 UI 中开放。

这样能让日志-only 用户最快完成接入，也符合当前真实落地范围。

### 5. AgentSRE 日志过滤是本轮关键约束

AgentSRE 通过 `/datasource/queryLogs` 按 `podName` / `k8sNamespace` 查日志。因此 Elastic Log check 不只验证 endpoint 可达，还会检查日志样本中是否存在常见 Kubernetes pod / namespace 字段。

如果能读到日志但缺少这些字段，测试返回 warning，不阻塞保存，但提示 AgentSRE 可能无法稳定查到服务级日志。

### 6. full APM 能力移入后续

`Trace`、`Metrics`、`Topology`、`ServiceInstance`、`RedMetric` 的 provider 代码可以继续作为探索骨架存在，但不作为本 change 的验收范围。后续恢复完整能力时，需要重新完成 UI、APO 后端契约、coverage、保存规则和真实环境联调。

这样避免把未联调能力误报为已完成。

## Risks / Trade-offs

- [功能边界比原始需求更窄] → 明确本 change 只交付 Elastic Log-only，full APM 另起 change。
- [日志中缺少 pod / namespace 时 AgentSRE 查不到服务日志] → Test 返回 warning，配置指南提示需要标准 Kubernetes 日志字段。
- [默认 `logs-*` 不覆盖所有现场索引] → 本 change 主流程先固定默认值；后续再开放自定义 index 和字段映射 UI。
- [Basic Auth 现场暂时不能通过 UI 配置] → provider 预留 Basic Auth，前端主流程后续再开放。

## Migration Plan

1. 先补 OpenSpec proposal、design、spec、tasks，作为实现合同。
2. 在 APO 商业仓实现 `elastic` 类型种子和 Log-only 配置入口。
3. 在 `apo-dataplane` 仓实现 `elastic` provider、注册、Log check 和 Log 查询。
4. 完成 APO 与 `apo-dataplane` 的 Log-only 联调。
5. 更新实现总结与系统文档，记录真实落地范围、偏差和跨仓治理结论。

回滚策略：

- 如果 APO UI 已上线但 provider 未稳定，可先隐藏 `elastic` 类型或撤回种子数据。
- 如果结构化 Test 结果存在兼容问题，可继续回退到只消费 `success/message`，不影响旧路径。
- 如果 Elastic Log 查询在联调中不稳定，可先隐藏 `elastic` 类型或撤回种子数据。

## Open Questions

- 后续完整 capability governance 是否需要同步回 OSS；如果需要，应使用哪个 OSS 分支？
- 后续是否开放 Basic Auth、自定义 index 和字段映射 UI，及其优先级。
- 后续 full APM change 中，`Topology` 应优先走 Elastic service map、trace 推导边关系，还是两者并存并按可用性选择。
