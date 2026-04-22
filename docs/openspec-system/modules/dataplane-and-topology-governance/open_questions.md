# Dataplane And Topology Governance Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `/api/dataplane/capabilities` 已有后端实现，但它根据 DB 中 `data_plane.capability` 聚合，不直接验证实时数据可用性；前端应把它解释为“配置覆盖”还是“运行态健康”仍需确认。
2. `custom_service_topology` 的创建和过期语义已经在后端实现，但当前前端主界面没有稳定入口，无法确认实际业务使用方式。
3. `CleanRuleServices` 与 `CleanDataSourceServices` 的真实触发场景在前端未见主流程使用，业务语义需要确认。

## Legacy Or Dead-Code Suspicion

1. `frontend/src/pro/api/dataplane.ts` 中存在多组重复 wrapper，说明接口层有明显历史兼容痕迹。
2. `service_check_router.go` 仍使用 `/agent-dataplane`，而 `init_router.go` 和当前主路由都已切到 `/dataplane`。
3. `/demo/service-config` 页面与服务治理命名相近，但当前是 mock 页面，容易被误判为旧版入口。

## Missing Frontend-Backend Alignment

1. 前端存在 `customtopology` 的 wrapper，但当前页面没有明确调用链。
2. Agent UI 跳转使用 `/dataplane`，权限白名单仍围绕 `/agent-dataplane`。
3. `DataCoverageOverview` 和 `DataStatus` 把 capabilities 当作覆盖状态展示，但后端接口只聚合已配置数据源声明，不包含实时健康校验。

## Permission Uncertainty

1. `AuthRouter` 当前不会真正执行 route permission 检查，因此前端页面可见性和后端权限模型并未完全一致。
2. `CheckRouterPermission` 的白名单是历史字符串匹配逻辑，是否足以覆盖当前前端路由需要确认。

## Data Model Uncertainty

1. `integration.DataPlane`、`DataPlaneType` 的完整表结构分布在集成模型中，当前模块文档只能确认其被当前服务使用的字段。
2. `service_name_rule_condition.value` 的真实数据来源受 APPInfo labels 限制，但标签全集依赖观测数据，无法仅从治理侧代码完整推断。
