# Automated Diagnosis And Workflows Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `alert check` workflow 输出 `"false"` 时触发 `auto analyze` 的业务语义只能从代码推断为“未通过检查则继续分析”，更细的 workflow 业务定义无法仅凭仓库确认。
2. `GetDescendantContribution` 同时支持 `latency,errorRate` 排序，但具体“贡献度”算法细节委托给 Polaris，仓库中无法完整复原。
3. `WorkflowParams` 中 `edition: "ee"` 是硬编码，是否意味着该自动 analyze 仅针对 EE 版本需要确认。

## Legacy Or Dead-Code Suspicion

1. `workflowLoginApi` 和 `workflowLogoutApi` wrapper 存在，但当前 `WorkflowsIframe` 只观察到 `workflowAnonymousLoginApi`。
2. `getAlertDescendantApi` wrapper 指向 `/api/alerts/descendant/anormal`，当前主后端未见该路径。
3. `service_integration_doc.go` 和数据接入模块的安装说明已标记 `Deprecated`，但这里的 workflow console入口仍是 iframe 方式，可能与未来控制台集成方式有关。

## Missing Frontend-Backend Alignment

1. 前端 `getAlertDescendantApi` 与后端实际 `/api/alerts/descendant/anormal/list` 路径不一致。
2. `ImpactList` 通过 `window.open('#/alert-analyze/info?...')` 传参，`ServiceAnalyzeContext` 再从 query 读取，缺少显式的共享路由状态模型。
3. 前端 workflow 入口依赖 `/dify/console/api/*`，当前主后端路由中未观察到对应实现，应继续确认是否由网关或代理提供。
4. `repository/dify` 是后端核心自动化执行边界，但前端 `/workflows` 页面直接面向 Dify console，两端所连接的 Dify 能力并不完全相同。

## Permission Uncertainty

1. `/workflows` 虽被 `AuthRouter` 包装，但前端 route 权限检查逻辑当前未实际执行。
2. `alertanalyze` routes 位于 `/api/alerts` 的 `AuthMiddleware()` 之后，但具体 feature 到 API 的授权映射还需结合权限模块继续确认。
3. Dify console `/dify/console/api/*` 的权限模型在当前仓库中未完整呈现。

## Data Model Uncertainty

1. `response/alert_input.go` 中同时承载了告警来源模块和自动化诊断模块的返回类型，模型边界较混杂。
2. `WorkflowRecord` 的实际写入路径在当前文档范围内只看到模型和 Dify worker，最终持久化 DAO 需要结合 ClickHouse 仓储继续核对。
3. `DetectMutation` 既表示执行记录又被规则列表复用，前后端对“规则”与“执行”的区分主要依赖查询方式。

## To Be Confirmed

- To Be Confirmed: `getAlertDescendantApi` 的路径漂移是否仍在真实页面中使用。
- To Be Confirmed: 非匿名用户的 Dify token 生命周期在当前前端如何维护。
- To Be Confirmed: `WorkflowRecord` 最终写入和展示闭环由哪些模块共同完成。
- To Be Confirmed: `alert check -> auto analyze` 的业务开关和 workflow 参数是否还有外部配置来源。
