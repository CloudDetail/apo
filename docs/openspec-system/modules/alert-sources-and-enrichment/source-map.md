# Alert Sources And Enrichment Source Map

## Scope
本文档只固定 `alert-sources-and-enrichment` 模块在当前仓库中的真实代码证据，作为后续详细文档的依据。该模块覆盖告警来源接入、告警事件入站、告警来源基础信息、告警 enrich rule、target tag、schema mapping、cluster 关联以及 provider pull/webhook 的后端能力。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/core/views/IntegrationCenter/AlertsIntegration` | `/integration/alerts` 页面主目录，包含告警接入列表、接入类型入口、配置抽屉、基础信息、enrich rule 和接入文档展示。 |
| Frontend | `frontend/src/core/api/alertInput.ts` | 前端已接线的 `/api/alertinput/*` API wrapper。 |
| Frontend | `frontend/src/core/contexts/AlertIntegrationContext.tsx` | 配置抽屉可见性、schema 列表和 target tag 列表的页面级 context。 |
| Frontend | `frontend/src/core/types/alertIntegration.ts` | 前端告警接入 source、enrich rule、schema target 的 TypeScript 类型。 |
| Frontend | `frontend/src/core/views/IntegrationCenter/constant.tsx` | `alertItems` 当前只暴露 `json`、`zabbix`、`prometheus` 三类告警来源。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/alertinput` route group 注册入口，当前整组只观察到 `LicMiddleware()`。 |
| Backend | `backend/pkg/api/alertinput` | 本模块 controller/handler 包。 |
| Backend | `backend/pkg/services/integration/alert` | 告警来源、事件入站、dispatcher、enricher、provider、schema、cluster 和默认 enrich rule 的主要 service 包。 |
| Backend | `backend/pkg/services/integration/alert/decoder` | `datadog`、`pagerduty`、`prometheus`、`zabbix`、`json` 等入站 payload decoder。 |
| Backend | `backend/pkg/services/integration/alert/enrich` | `AlertEnricher` 和 `TagEnricher`，负责按 rule 对事件补充 tags。 |
| Backend | `backend/pkg/services/integration/alert/provider` | provider registry、pull/webhook 能力声明和 provider 参数校验。 |
| Backend | `backend/pkg/model/integration/alert` | 告警 source、enrich rule、target tag、schema request/response、错误类型。 |
| Backend | `backend/pkg/model/integration/cluster.go` | `Cluster` 模型由本模块 cluster API 复用。 |
| Backend | `backend/pkg/repository/database/integration/alert` | `alert_sources`、`target_tags`、`alert_enrich_rules`、动态 schema table 和 source-cluster 关系的数据库 DAO。 |
| Backend | `backend/pkg/repository/clickhouse` | enrich 后的告警事件最终写入 ClickHouse。 |
| Backend | `backend/pkg/repository/dify` | 入站或 pull 得到的告警事件会提交给 Dify。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/routes.js` | `/integration/alerts` |
| Backend | `backend/pkg/router/router_api.go` | `/api/alertinput/event/source`, `/api/alertinput/event/json`, `/api/alertinput/source/create`, `/api/alertinput/source/paramspec`, `/api/alertinput/source/setup/webhook`, `/api/alertinput/source/update`, `/api/alertinput/source/get`, `/api/alertinput/source/delete`, `/api/alertinput/source/list`, `/api/alertinput/source/enrich/update`, `/api/alertinput/source/enrich/get`, `/api/alertinput/enrich/tags/list`, `/api/alertinput/cluster/*`, `/api/alertinput/schema/*`, `/api/alertinput/source/enrich/default/*` |
| Backend Adjacent | `backend/pkg/router/router_api.go` | `/api/alerts/inputs/alertmanager` 由 `alerts` handler 暴露，但实际转到 `alertinput.ProcessAlertEvents`，因此是本模块的跨模块输入边界。 |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/index.tsx` | `/integration/alerts` 页面，组合 `AlertsDatasourceList`、`AlertsIntegrationTable` 和 `IntegrationDrawer`。 |
| Source Type List | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/AlertsDatasourceList/index.tsx` | 左侧新增入口，当前从 `alertItems` 暴露 `json`、`zabbix`、`prometheus`。 |
| Source Table | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/AlertsIntegrationTable/index.tsx` | 读取 source 列表，支持编辑和删除。 |
| Config Drawer | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/index.tsx` | 通过 URL query `sourceId`、`sourceType` 控制新增或编辑抽屉。 |
| Base Info | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/BaseInfoContent/index.tsx` | 创建或更新 `sourceName`、`sourceType`、`clusters`，并展示 push URL。 |
| Base Info View | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/BaseInfoContent/BaseInfoDescriptions.tsx` | 只展示 `sourceName` 和 `/api/alertinput/event/source?sourceId=...`。 |
| Cluster UI | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/BaseInfoContent/ClusterSelector.tsx` | cluster selector 已实现，但在 `BaseInfoContent/index.tsx` 中被注释掉。 |
| Enrich Rule | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/TagContent/index.tsx` | 告警来源已存在时展示 enrich rule 编辑区。 |
| Enrich Rule Form | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/TagContent/TagContactRule/index.tsx` | 读取和保存 `enrichRuleConfigs`。 |
| Rule Item | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/TagContent/TagContactRule/TagRuleFormListCom.tsx` | rule 类型、条件、from field、tagMapping 或 schemaMapping 表单主体。 |
| Rule Conditions | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/TagContent/TagContactRule/ConditionsFormList.tsx` | 条件操作当前 UI 暴露 `match` 和 `notMatch`。 |
| Schema Mapping | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/TagContent/TagContactRule/SchemaSource.tsx` | 从 context 中的 `schemas` 选择 schema 和 source column。 |
| Schema Targets | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/TagContent/TagContactRule/SchemaFormList.tsx` | 选择 schema target fields 并绑定 target tag。 |
| Target Tag | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/TagContent/TagContactRule/TargetTagSelector.tsx` | 使用 `targetTags` 列表，也允许 id 为 `0` 的自定义 tag。 |
| Source Docs | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/ConfigDrawer/SourceTypeInfo/*` | 按 `sourceType` 展示 `PrometheusInfo`、`ZabbixInfo` 或 `JsonInfo` 文档说明。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/alertInput.ts` | `getAlertInputSourceListApi` | `/api/alertinput/source/list` | 页面表格读取 source 列表。 |
| `frontend/src/core/api/alertInput.ts` | `creatAlertInputSourceApi` | `/api/alertinput/source/create` | 新增 source。函数名当前拼写为 `creatAlertInputSourceApi`。 |
| `frontend/src/core/api/alertInput.ts` | `updateAlertsIntegrationApi` | `/api/alertinput/source/update` | 更新 source 基础信息。 |
| `frontend/src/core/api/alertInput.ts` | `getAlertInputBaseInfoApi` | `/api/alertinput/source/get` | 抽屉打开时读取单个 source。 |
| `frontend/src/core/api/alertInput.ts` | `getClusterListApi` | `/api/alertinput/cluster/list` | 只在已注释的 `ClusterSelector` 中使用。 |
| `frontend/src/core/api/alertInput.ts` | `getTargetTagsListApi` | `/api/alertinput/enrich/tags/list` | 抽屉打开时写入 context。 |
| `frontend/src/core/api/alertInput.ts` | `getSchemaListApi` | `/api/alertinput/schema/list` | wrapper 存在，当前主页面未观察到直接调用。 |
| `frontend/src/core/api/alertInput.ts` | `getSchemaColumnsApi` | `/api/alertinput/schema/column/get` | wrapper 存在，当前主页面未观察到直接调用。 |
| `frontend/src/core/api/alertInput.ts` | `saveAlertEnrichApi` | `/api/alertinput/source/enrich/update` | 保存 source enrich rule。 |
| `frontend/src/core/api/alertInput.ts` | `getAlertEnrichApi` | `/api/alertinput/source/enrich/get` | 读取 source enrich rule。 |
| `frontend/src/core/api/alertInput.ts` | `getAllSchemaApi` | `/api/alertinput/schema/listwithcolumns` | 抽屉打开时写入 context。 |
| `frontend/src/core/api/alertInput.ts` | `deleteAlertIntegrationApi` | `/api/alertinput/source/delete` | 删除 source。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router | `backend/pkg/router/router_api.go` | `/api/alertinput` 整组使用 `LicMiddleware()`，未观察到本组追加 `AuthMiddleware()`。 |
| Handler Factory | `backend/pkg/api/alertinput/handler.go` | 创建 `alertinput.Service`，接口声明包含 source、event、cluster、schema、default enrich rule 和 provider webhook。 |
| Event Handler | `backend/pkg/api/alertinput/func_sourcehandler.go` | `/api/alertinput/event/source`，按 query 中的 `SourceFrom` 读取 body 并调用 `ProcessAlertEvents`。 |
| Event Handler | `backend/pkg/api/alertinput/func_jsonhandler.go` | `/api/alertinput/event/json`，缺省 `sourceType` 填 `"unkonwn"`，缺省 `sourceName` 用 client IP 派生。 |
| Source Handler | `backend/pkg/api/alertinput/func_createalertsource.go`, `func_updatealertsource.go`, `func_getalertsource.go`, `func_deletealertsource.go`, `func_listalertsource.go` | 告警来源 CRUD。 |
| Provider Handler | `backend/pkg/api/alertinput/func_getalertproviderparamsspec.go`, `func_setupproviderwebhook.go` | provider params spec 与 webhook setup。实际注册 webhook 路径为 `/api/alertinput/source/setup/webhook`。 |
| Enrich Handler | `backend/pkg/api/alertinput/func_updatealertsourceenrichrule.go`, `func_getalertsourceenrichrule.go`, `func_listtargettags.go` | source enrich rule 和 target tag。 |
| Cluster Handler | `backend/pkg/api/alertinput/func_createcluster.go`, `func_listcluster.go`, `func_updatecluster.go`, `func_deletecluster.go` | cluster CRUD 后端已暴露。 |
| Schema Handler | `backend/pkg/api/alertinput/func_createschema.go`, `func_deleteschema.go`, `func_listschema.go`, `func_listschemawithcolumns.go`, `func_getschemacolumns.go`, `func_updateschemadata.go`, `func_getschemadata.go`, `func_checkschemaisused.go` | schema table、schema columns、schema data 和使用状态后端接口。 |
| Default Rule Handler | `backend/pkg/api/alertinput/func_getdefaultalertenrichrule.go`, `func_setdefaultalertenrichrule.go`, `func_cleardefaultalertenrichrule.go` | 默认 enrich rule 后端接口。实际 router 中读取路径为 `/api/alertinput/source/enrich/default/get`。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Service Construction | `backend/pkg/services/integration/alert/service.go` | 初始化已存在 source/enrich rule、target tags、provider pull 线程和 default enrich rules。 |
| Event Processing | `backend/pkg/services/integration/alert/service_process_alert_events.go` | 入站事件经过 dispatcher decode/enrich 后提交 Dify 并写入 ClickHouse。 |
| Source CRUD | `backend/pkg/services/integration/alert/service_alert_source.go`, `service_create_alert_source.go` | 创建、更新、删除 source；维护 dispatcher、provider 和 enrich rule 存储。 |
| Dispatcher | `backend/pkg/services/integration/alert/dispatch.go` | 用 `SourceID` 或 `SourceName` 找到 `AlertEnricher`，再调用 decoder 和 enricher。 |
| Enrich Rule | `backend/pkg/services/integration/alert/service_update_alert_enrich_rule.go`, `service_get_alert_enrich_rule.go`, `service_get_alert_enrich_rule_tags.go` | 更新和读取 source enrich rule，读取 target tags。 |
| Default Rule | `backend/pkg/services/integration/alert/service_default_alert_source.go` | 按 `sourceType` 维护默认 enrich rule，缺省回落到 `json`。 |
| Schema | `backend/pkg/services/integration/alert/service_schema.go` | schema 创建、列表、列、数据读写、删除和使用检测。 |
| Cluster | `backend/pkg/services/integration/alert/service_clusters.go` | cluster CRUD，直接委托 database repo。 |
| Provider | `backend/pkg/services/integration/alert/provider_manager.go`, `service_setup_webhook.go`, `service_get_provider_params_spec.go` | provider pull 线程、webhook 安装和 params spec。 |
| Provider Registry | `backend/pkg/services/integration/alert/provider/provider.go`, `provider/datadog.go`, `provider/pagerduty.go` | `datadog` 支持 pull 和 webhook install；`pagerduty` 支持 webhook install；`prometheus`、`zabbix`、`json` 当前为 `BasicEncoder`。 |
| Decoder | `backend/pkg/services/integration/alert/decoder/*.go` | 不同来源 payload 到 `AlertEvent` 的解析。 |
| Enricher | `backend/pkg/services/integration/alert/enrich/*.go` | rule 执行、tagMapping、schemaMapping 和条件匹配。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Source Model | `backend/pkg/model/integration/alert/alert_source.go` | `AlertSource.TableName()` 为 `alert_sources`，包含 `SourceFrom`、`Params`、`EnabledPull`、`LastPullMillTS`。 |
| Source Cluster | `backend/pkg/model/integration/alert/alert_source.go` | `AlertSource2Cluster` 用于 source 与 cluster 的关联。 |
| Enrich Model | `backend/pkg/model/integration/alert/alert_enrich.go` | `AlertEnrichRule.TableName()` 为 `alert_enrich_rules`，并定义 conditions 和 schema targets。 |
| Target Tag | `backend/pkg/model/integration/alert/target_tag.go` | `TargetTag.TableName()` 为 `target_tags`。 |
| Request Model | `backend/pkg/model/integration/alert/req.go` | `AlertEnrichRuleConfigRequest`、schema request、default enrich request。 |
| Response Model | `backend/pkg/model/integration/alert/resp.go` | `ListAlertSourceResponse`、`GetTargetTagsResponse`、schema response、`GetAlertEnrichRuleResponse`。 |
| Error Model | `backend/pkg/model/integration/alert/errors.go` | `ErrAlertSourceAlreadyExist`、`ErrAlertSourceNotExist`、`ErrNotAllowSchema`、`ErrIllegalAlertRule`。 |
| Cluster Model | `backend/pkg/model/integration/cluster.go` | `Cluster` 包含 `id`、`name`、`clusterType`、`apoCollector`、`isMinimal`。 |
| Schema DAO | `backend/pkg/repository/database/integration/alert/dao_manage_schema.go` | 动态 schema table 前缀为 `alert_input_schema_`，schema/table/column 名称受 `AllowSchema` 限制。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Route Guard | `frontend/src/core/components/AuthRouter.tsx` | `/integration/alerts` 被 `AuthRouter` 包装，但 `checkRoutePermission` 的 `useEffect` 当前被注释。 |
| Menu Permission | `backend/pkg/repository/database/init_menu_item.go` | `alertsIntegration` menu item 绑定 `/integration/alerts`，父级为 `integration`。 |
| Router Metadata | `backend/pkg/repository/database/init_router.go` | `/integration/alerts` 的 `HideTimeSelector` 为 `true`。 |
| Request Utility | `frontend/src/core/utils/request.js` | `alertInput.ts` 通过共用 `get`、`post` 发起请求。 |
| Notification Utility | `frontend/src/core/utils/notify` | source 创建、更新、删除和 enrich rule 保存成功后提示。 |
| License Middleware | `backend/pkg/middleware/lic_middleware.go` | `/api/alertinput` route group 当前只有 license 控制证据。 |
| Database Repo | `backend/pkg/repository/database/integration/alert` | source、target tags、enrich rule、schema 和 cluster 关系的主要持久化边界。 |
| ClickHouse Repo | `backend/pkg/repository/clickhouse` | `ProcessAlertEvents` 和 provider pull 成功后写入告警事件。 |
| Dify Repo | `backend/pkg/repository/dify` | 入站和 pull 得到的事件会调用 `SubmitAlertEvents`。 |
| Cross Module | `alert-rules-and-incident-events` | `/api/alerts/inputs/alertmanager` 作为 AlertManager 输入入口，实际复用本模块 service。 |

## Evidence Notes

1. 前端只暴露 `json`、`zabbix`、`prometheus` 三类 `AlertKey`，但后端 `ProviderRegistry` 还包含 `datadog` 和 `pagerduty`。
2. `ClusterSelector.tsx` 已实现，且 `AlertInputSourceParams` 支持 `clusters`，但 `BaseInfoContent/index.tsx` 中该组件被注释，当前主 UI 不暴露 cluster 选择。
3. 前端当前未观察到直接调用 `/api/alertinput/source/paramspec`、`/api/alertinput/source/setup/webhook`、cluster CRUD、schema CRUD/data、default enrich rule 接口。
4. `func_setupproviderwebhook.go` 的 swagger 注释路径是 `/api/alertinput/source/webhook`，但实际 router 注册为 `/api/alertinput/source/setup/webhook`。
5. `func_getdefaultalertenrichrule.go` 的 swagger 注释路径是 `/api/alertinput/source/enrich/default`，但实际 router 注册为 `/api/alertinput/source/enrich/default/get`。
6. `JsonHandler` 默认 source type 拼写为 `"unkonwn"`，这是当前代码事实，文档不做自动纠正。
