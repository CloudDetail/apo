# Alert Sources And Enrichment Module Overview

## Module Goal
本模块负责把外部告警事件接入 APO，并在进入告警事件存储与后续诊断流程前执行来源识别、payload decode、tag enrich、schema mapping 和 provider pull/webhook 相关处理。前端当前提供 `/integration/alerts` 告警接入配置页，后端提供 `/api/alertinput/*` 接口组以及被 `/api/alerts/inputs/alertmanager` 复用的事件处理能力。

## Scope
当前代码证据支持以下范围：

- 告警来源列表查看、新增、编辑、删除。
- 告警来源 push URL 展示，URL 形态为 `/api/alertinput/event/source?sourceId=...`。
- 按 source 维护 enrich rule，包含 `tagMapping` 和 schema mapping 表单结构。
- 使用 target tags 作为标准化目标标签，并允许前端通过 id 为 `0` 的方式提交自定义 tag。
- 使用动态 schema table 支持 schema mapping。
- 后端支持 schema 创建、删除、列查询、数据读写和使用检测。
- 后端支持 cluster CRUD 和 source-cluster 关系模型，但当前主前端 cluster selector 未启用。
- 后端支持 provider params spec、provider webhook setup、provider pull 和默认 enrich rule。
- 后端支持 `/api/alertinput/event/source` 与 `/api/alertinput/event/json` 入站事件处理。
- 入站或 pull 得到的事件会经 dispatcher、decoder、enricher 处理后提交 Dify 并写入 ClickHouse。

当前代码不应纳入本模块主体范围的能力：

- `alert-rules-and-incident-events`：告警规则、AlertManager receiver、告警事件列表、事件详情、静默和 resolve。
- `automated-diagnosis-and-workflows`：`/alert-analyze`、突变检测、异常贡献度和 workflow 配置主体。
- `data-ingestion-and-oneagent`：APM、trace、metric、log 的数据接入与 OneAgent 安装引导。
- `dataplane-and-topology-governance`：服务、endpoint、instance、red chart 和 topology 查询。
- `conversational-diagnosis-and-reporting`：报告页、chat、agent、custom query 展示。

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/routes.js` | 注册 `/integration/alerts`，使用 `AuthRouter` 包装。 |
| Page | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/index.tsx` | 告警接入页面容器。 |
| Source List | `AlertsDatasourceList/index.tsx` | 提供新增 `json`、`zabbix`、`prometheus` 告警来源的入口。 |
| Source Table | `AlertsIntegrationTable/index.tsx` | 读取 `/api/alertinput/source/list`，支持编辑和删除。 |
| Drawer | `ConfigDrawer/index.tsx` | 由 URL query 控制新增/编辑抽屉状态。 |
| Base Info | `ConfigDrawer/BaseInfoContent/index.tsx` | 新增或更新 source 基础信息，展示 push URL。 |
| Enrich Rules | `ConfigDrawer/TagContent/TagContactRule/index.tsx` | 读取和保存 source enrich rule。 |
| Context | `frontend/src/core/contexts/AlertIntegrationContext.tsx` | 抽屉打开时加载 schema 列表和 target tags。 |
| API Wrapper | `frontend/src/core/api/alertInput.ts` | 前端本模块主要 API 封装。 |
| Types | `frontend/src/core/types/alertIntegration.ts` | `AlertKey`、`AlertInputSourceParams`、`SaveAlertEnrichParams` 等类型。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | `/api/alertinput` route group，当前整组使用 `LicMiddleware()`。 |
| Handler Package | `backend/pkg/api/alertinput` | 本模块主要 handler 包。 |
| Service Package | `backend/pkg/services/integration/alert` | 本模块主要业务逻辑。 |
| Decoder Package | `backend/pkg/services/integration/alert/decoder` | 负责不同来源 payload decode。 |
| Enrich Package | `backend/pkg/services/integration/alert/enrich` | 负责 tag enrich 和 schema mapping。 |
| Provider Package | `backend/pkg/services/integration/alert/provider` | provider registry、params spec、pull/webhook 能力声明。 |
| Models | `backend/pkg/model/integration/alert` | source、target tag、enrich rule、schema request/response 和错误类型。 |
| Repository | `backend/pkg/repository/database/integration/alert` | source、enrich、target tag、schema、source-cluster 关系持久化。 |
| Cross Entry | `backend/pkg/api/alerts/func_inputalertmanager.go` | `/api/alerts/inputs/alertmanager` 实际复用 `alertinput.ProcessAlertEvents`。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `AlertSource` | 告警来源配置，表名为 `alert_sources`，包含 `sourceId`、`sourceName`、`sourceType`、`params`、`enabledPull` 和 `lastPullMillTS`。 |
| `SourceFrom` | 入站事件处理时携带的来源标识，支持通过 `sourceId` 或 `sourceName` 查找 enricher。 |
| `AlertEnrichRule` | enrich rule 持久化模型，表名为 `alert_enrich_rules`。 |
| `AlertEnrichCondition` | enrich rule 的条件列表，当前前端只暴露 `match` 和 `notMatch`，后端注释还列出 `gt`、`lt`、`ge`、`le`、`eq`。 |
| `AlertEnrichSchemaTarget` | schema mapping 的目标字段和 target tag 关系。 |
| `target_tags` | 标准化目标标签表，前端会过滤掉 `id === 0` 的记录，并用 `id = 0` 表示自定义 tag。 |
| `tagMapping` | 前端和后端均识别的 rule 类型，从事件字段和可选 regex 直接映射到 target tag。 |
| `schemaMapping` | 后端模型中的 schema mapping 类型，前端 UI 文案选项值当前使用 `staticEnrich`，保存时走非 `tagMapping` 分支并组装 `schema`、`schemaSource`、`schemaTargets`。 |
| `alert_input_schema_` | 动态 schema table 前缀，schema 名和 column 名必须匹配 `^[a-zA-Z0-9_-]{1,40}$`。 |
| `ProviderRegistry` | provider 能力注册表，当前包含 `datadog`、`pagerduty`、`prometheus`、`zabbix`、`json`。 |
| `EnabledPull` | source 开启 pull 后，后端才尝试为支持 pull 的 provider 启动定时拉取 goroutine。 |
| `defaultEnrichRules` | 按 `sourceType` 保存默认 enrich rule，不存在时回落到 `json`。 |

## Related Files

- `frontend/src/routes.js`
- `frontend/src/core/api/alertInput.ts`
- `frontend/src/core/contexts/AlertIntegrationContext.tsx`
- `frontend/src/core/types/alertIntegration.ts`
- `frontend/src/core/views/IntegrationCenter/constant.tsx`
- `frontend/src/core/views/IntegrationCenter/AlertsIntegration/index.tsx`
- `frontend/src/core/views/IntegrationCenter/AlertsIntegration/AlertsDatasourceList/index.tsx`
- `frontend/src/core/views/IntegrationCenter/AlertsIntegration/AlertsIntegrationTable/index.tsx`
- `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/index.tsx`
- `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/BaseInfoContent/index.tsx`
- `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/TagContent/TagContactRule/index.tsx`
- `backend/pkg/router/router_api.go`
- `backend/pkg/api/alertinput/handler.go`
- `backend/pkg/services/integration/alert/service.go`
- `backend/pkg/services/integration/alert/service_process_alert_events.go`
- `backend/pkg/services/integration/alert/service_alert_source.go`
- `backend/pkg/services/integration/alert/service_create_alert_source.go`
- `backend/pkg/services/integration/alert/service_update_alert_enrich_rule.go`
- `backend/pkg/services/integration/alert/service_default_alert_source.go`
- `backend/pkg/services/integration/alert/provider/provider.go`
- `backend/pkg/services/integration/alert/provider_manager.go`
- `backend/pkg/model/integration/alert/alert_source.go`
- `backend/pkg/model/integration/alert/alert_enrich.go`
- `backend/pkg/model/integration/alert/target_tag.go`
- `backend/pkg/repository/database/integration/alert/dao.go`
- `backend/pkg/repository/database/integration/alert/dao_manage_schema.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `AuthRouter` | `/integration/alerts` 被包装，但当前实际 route 权限检查逻辑被注释。 |
| Frontend Shared | `UserContext` + `AppSidebar` | 菜单可见性来自后端权限配置，`alertsIntegration` 由数据库初始化。 |
| Frontend Shared | `request.js` | 本模块 API wrapper 统一通过共享 request 工具发送请求。 |
| Frontend Shared | `notify` | 创建、更新、删除、保存 enrich rule 后展示结果。 |
| Backend Shared | `LicMiddleware` | `/api/alertinput` 整组受 license 控制。 |
| Backend Repo | `database.Repo` | source、enrich rule、target tags、schema 和 cluster 的主要持久化入口。 |
| Backend Repo | `clickhouse.Repo` | enrich 后的事件写入 ClickHouse。 |
| Backend Repo | `dify.DifyRepo` | enrich 后事件提交给 Dify。 |
| Cross Module | `alert-rules-and-incident-events` | `/api/alerts/inputs/alertmanager` 入口实际转到本模块处理告警事件。 |

## Risks And Unknowns

1. `/api/alertinput` route group 当前只观察到 `LicMiddleware()`，未观察到 `AuthMiddleware()`，权限边界需要确认。
2. 前端只暴露 `json`、`zabbix`、`prometheus`，但后端 provider registry 支持 `datadog` 和 `pagerduty`，存在“后端能力大于当前 UI”的差异。
3. `ClusterSelector` 已实现但被注释，source-cluster 关系是否为未来功能或临时关闭需要确认。
4. 前端没有直接暴露 provider params、webhook setup、schema CRUD、cluster CRUD 和 default enrich rule 配置，不能把这些能力写成当前用户可见主流程。
5. `JsonHandler` 默认 `sourceType` 为 `"unkonwn"`，拼写可疑但属于现有契约。
6. `ProcessAlertEvents` 在带 `sourceId` 但 source 不存在时直接返回 `nil` 丢弃事件，需要确认是否符合业务预期。
7. `schemaMapping` 后端类型和前端 UI 选项值 `staticEnrich` 之间存在命名差异，需要确认是否由 service/enricher 兼容处理。
