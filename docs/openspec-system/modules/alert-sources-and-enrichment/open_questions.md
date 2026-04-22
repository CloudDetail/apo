# Alert Sources And Enrichment Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `ProcessAlertEvents` 在请求带 `sourceId` 但 source 未创建时直接返回 `nil`，业务上这是有意丢弃未配置来源的事件，还是应该返回错误，无法仅从代码确认。
2. source 新建成功后自动 `initDefaultAlertSource` 的默认规则来源和业务含义需要确认，尤其是 `json` 默认规则作为回退规则的语义。
3. `target_tags` 中 id 为 `0` 被前端用于自定义 tag，但数据库种子数据和后端是否保留 id 为 `0` 的特殊语义需要进一步确认。
4. schema mapping 的业务目标是“静态数据补充”还是“外部字典映射”，当前只能从 schema table 和 UI 表单推断，不能扩写成更具体规则。

## Legacy Or Dead-Code Suspicion

1. `ClusterSelector.tsx` 已实现但在 `BaseInfoContent/index.tsx` 中被注释，可能是未完成、临时关闭或遗留能力。
2. `getSchemaListApi` 和 `getSchemaColumnsApi` wrapper 存在，但当前主页面未观察到直接调用。
3. cluster CRUD、schema CRUD、schema data、default enrich rule、provider params spec、provider webhook setup 的后端接口已暴露，但当前主前端未观察到入口。
4. `func_setupproviderwebhook.go`、`func_getdefaultalertenrichrule.go`、`func_listtargettags.go` 的 swagger 注释路径与 router 实际路径存在差异，可能是注释遗留。
5. `JsonHandler` 默认 source type 为 `"unkonwn"`，拼写可能是历史问题，但属于当前运行契约。

## Missing Frontend-Backend Alignment

1. 前端 `AlertKey` 只允许 `json`、`zabbix`、`prometheus`，后端 `ProviderRegistry` 还包含 `datadog` 和 `pagerduty`。
2. 前端 `TagRuleFormListCom` 的非 tag mapping 选项值是 `staticEnrich`，后端模型注释和字段使用 `schemaMapping`，需要确认实际 enricher 是否以非 `tagMapping` 兜底处理。
3. 前端 `BaseInfoContent` 会提交 `clusters`，但 cluster 选择 UI 被注释，后端 source-cluster 关系的实际使用路径不清晰。
4. 前端不编辑 `AlertSource.Params`、`EnabledPull` 或 `LastPullMillTS`，因此 provider pull 能力当前缺少可见配置入口。
5. 前端没有调用 `/api/alertinput/source/setup/webhook`，因此 webhook 自动安装能力没有明确 UI 入口。

## Permission Uncertainty

1. `/api/alertinput` 整组当前只观察到 `LicMiddleware()`，未观察到 `AuthMiddleware()`。
2. `/integration/alerts` 被 `AuthRouter` 包装，但 `AuthRouter` 中实际权限检查被注释。
3. `alertsIntegration` 菜单与 `告警接入` feature 有初始化映射，但 API 权限映射是否存在无法从当前文件确认。
4. 入站事件接口 `/api/alertinput/event/source` 和 `/api/alertinput/event/json` 与管理接口位于同一 route group，是否应该有不同认证策略需要确认。

## Data Model Uncertainty

1. `AlertSource2Cluster` 自动迁移存在，但未在本轮证据中确认其 DAO 读写是否完整覆盖 source 创建、更新和查询。
2. `Cluster` 模型复用于多个集成场景，本模块 cluster API 与 `/api/integration/cluster/*` 的边界需要后续统一确认。
3. 动态 schema table 使用 `alert_input_schema_` 前缀并执行 SQL 拼接，虽然有 `AllowSchema` 和 `util.ValidateSQL`，仍需关注不同数据库方言下的安全和兼容性。
4. `DeleteSchema` 会先删除关联 `AlertEnrichSchemaTarget`，但对 `AlertEnrichRule` 主记录的保留语义需要确认。
5. `service.New` 中 `difyRepo` 赋值时序位于初始化 provider pull 后，是否可能造成 provider goroutine 使用 nil `difyRepo` 需要确认。

## To Be Confirmed

- To Be Confirmed: `sourceId` 不存在时直接丢弃事件是否为故意设计。
- To Be Confirmed: `staticEnrich` 与 `schemaMapping` 是否为同一概念。
- To Be Confirmed: `datadog` 和 `pagerduty` 是否计划在 `/integration/alerts` 前端暴露。
- To Be Confirmed: `/api/alertinput` 是否需要与其他管理接口一样追加 `AuthMiddleware()`。
