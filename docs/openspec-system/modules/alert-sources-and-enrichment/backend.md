# Alert Sources And Enrichment Backend

## API Summary

| Category | Method | API Path | Handler | Notes |
|---|---|---|---|---|
| Event Input | POST | `/api/alertinput/event/source` | `SourceHandler` | 按 query 中的 `sourceId`、`sourceName`、`sourceType` 处理原始 body。 |
| Event Input | POST | `/api/alertinput/event/json` | `JsonHandler` | 缺省 source type 为 `"unkonwn"`，缺省 source name 由 client IP 生成。 |
| Source | POST | `/api/alertinput/source/create` | `CreateAlertSource` | 创建 source，初始化默认 enrich rule，并尝试 setup provider。 |
| Source | POST | `/api/alertinput/source/update` | `UpdateAlertSource` | 更新 source，并重新 setup provider。 |
| Source | POST | `/api/alertinput/source/get` | `GetAlertSource` | 按 `sourceId` 获取 source。 |
| Source | POST | `/api/alertinput/source/delete` | `DeleteAlertSource` | 删除 source、provider、dispatcher enricher 和 enrich rule。 |
| Source | GET | `/api/alertinput/source/list` | `ListAlertSource` | 返回 `alertSources`。 |
| Provider | GET | `/api/alertinput/source/paramspec` | `GetAlertProviderParamsSpec` | 返回单个或全部 provider params spec。 |
| Provider | POST | `/api/alertinput/source/setup/webhook` | `SetupAlertProviderWebhook` | router 实际路径为该值，swagger 注释不一致。 |
| Enrich | POST | `/api/alertinput/source/enrich/update` | `UpdateAlertSourceEnrichRule` | 更新 source enrich rule。 |
| Enrich | POST | `/api/alertinput/source/enrich/get` | `GetAlertSourceEnrichRule` | router 为 POST，handler 注释写 GET。 |
| Enrich | GET | `/api/alertinput/enrich/tags/list` | `ListTargetTags` | 返回 `targetTags`。 |
| Cluster | POST | `/api/alertinput/cluster/create` | `CreateCluster` | 后端已暴露，当前主前端未接线。 |
| Cluster | GET | `/api/alertinput/cluster/list` | `ListCluster` | `ClusterSelector` 中存在调用，但 UI 被注释。 |
| Cluster | POST | `/api/alertinput/cluster/update` | `UpdateCluster` | 后端已暴露，当前主前端未接线。 |
| Cluster | POST | `/api/alertinput/cluster/delete` | `DeleteCluster` | 后端已暴露，当前主前端未接线。 |
| Schema | POST | `/api/alertinput/schema/create` | `CreateSchema` | 创建动态 schema table，可同时插入 `fullRows`。 |
| Schema | GET | `/api/alertinput/schema/delete` | `DeleteSchema` | 删除 schema table，并清理相关 schema target。 |
| Schema | GET | `/api/alertinput/schema/used/check` | `CheckSchemaIsUsed` | 检查 schema 是否被 enrich rule 使用。 |
| Schema | GET | `/api/alertinput/schema/list` | `ListSchema` | 返回 schema 名称列表。 |
| Schema | GET | `/api/alertinput/schema/listwithcolumns` | `ListSchemaWithColumns` | 返回 schema 到 columns 的 map。 |
| Schema | GET | `/api/alertinput/schema/column/get` | `GetSchemaColumns` | 返回单个 schema 的 columns。 |
| Schema | POST | `/api/alertinput/schema/data/update` | `UpdateSchemaData` | 清空、新增或更新 schema 数据。 |
| Schema | GET | `/api/alertinput/schema/data/get` | `GetSchemaData` | 返回 schema columns 和 rows。 |
| Default Rule | GET | `/api/alertinput/source/enrich/default/clear` | `ClearDefaultAlertEnrichRule` | 清理 source type 的默认 enrich rule。 |
| Default Rule | GET | `/api/alertinput/source/enrich/default/get` | `GetDefaultAlertEnrichRule` | router 实际路径为该值，swagger 注释不一致。 |
| Default Rule | POST | `/api/alertinput/source/enrich/default/set` | `SetDefaultAlertEnrichRule` | 设置 source type 默认 enrich rule。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 注册 `/api/alertinput` route group，整组使用 `LicMiddleware()`。 |
| `backend/pkg/api/alertinput/handler.go` | 定义 handler interface，并构造 `alertinput.New(promRepo, dbRepo, chRepo, difyRepo)`。 |
| `backend/pkg/api/alertinput/func_sourcehandler.go` | 接收按 source 配置推送的告警事件。 |
| `backend/pkg/api/alertinput/func_jsonhandler.go` | 接收通用 JSON 告警事件。 |
| `backend/pkg/api/alertinput/func_createalertsource.go` | 创建 source，并对 `ErrAlertSourceAlreadyExist` 使用专门错误码。 |
| `backend/pkg/api/alertinput/func_updatealertsource.go` | 更新 source，并对 `ErrAlertSourceAlreadyExist` 使用专门错误码。 |
| `backend/pkg/api/alertinput/func_updatealertsourceenrichrule.go` | 更新 enrich rule，并对 `ErrAlertSourceNotExist` 使用专门错误码。 |
| `backend/pkg/api/alertinput/func_getalertproviderparamsspec.go` | 查询 provider 参数规格。 |
| `backend/pkg/api/alertinput/func_setupproviderwebhook.go` | 设置 provider webhook，失败时返回 `SetupAlertProviderWebhookError`。 |

## Services

| Service | File | Behavior |
|---|---|---|
| `New` | `service.go` | 启动时加载已有 source 和 enrich rules，初始化 dispatcher，加载 target tags，启动 provider pull。 |
| `ProcessAlertEvents` | `service_process_alert_events.go` | 调用 dispatcher decode/enrich，提交 Dify，再写入 ClickHouse。 |
| `CreateAlertSource` | `service_create_alert_source.go` | 校验 source name 唯一性，生成或校验 uuid，写入 DB，初始化默认 enrich rule，setup provider。 |
| `UpdateAlertSource` | `service_alert_source.go` | 校验 source 存在和 source name 不冲突，更新 DB，重新 setup provider。 |
| `DeleteAlertSource` | `service_alert_source.go` | 删除 DB source、provider、dispatcher enricher、enrich rules、conditions、schema targets。 |
| `UpdateAlertEnrichRule` | `service_update_alert_enrich_rule.go` | 构建新的 enricher，对比旧 rule，按删除、修改和新增批量更新 DB，并刷新 dispatcher。 |
| `GetAlertEnrichRule` | `service_get_alert_enrich_rule.go` | 聚合 `AlertEnrichRule`、conditions 和 schema targets。 |
| `GetAlertEnrichRuleTags` | `service_get_alert_enrich_rule_tags.go` | 返回 `target_tags`。 |
| `CreateSchema` | `service_schema.go` | 创建动态 schema table，并可插入初始数据。 |
| `DeleteSchema` | `service_schema.go` | 删除 schema 前更新内存 enricher，再删除 table。 |
| `SetupAlertProvider` | `provider_manager.go` | 只有 `EnabledPull` 为 true 且 provider 支持 pull 时才启动 goroutine。 |
| `SetupProviderWebhook` | `service_setup_webhook.go` | 只有 provider 支持 `SupportWebhookInstall` 时才执行 webhook setup。 |
| `GetProviderParamsSpec` | `service_get_provider_params_spec.go` | 返回 registry 中 provider spec；未知 source type 使用 `BasicEncoder(sourceType)`。 |
| `GetDefaultAlertEnrichRule` | `service_default_alert_source.go` | source type 无默认规则时回退到 `json`。 |

## Models And Schemas

| Model | File | Mapping |
|---|---|---|
| `AlertSource` | `backend/pkg/model/integration/alert/alert_source.go` | `TableName() == "alert_sources"`。 |
| `AlertSource2Cluster` | `alert_source.go` | source 与 cluster 的关系模型。 |
| `SourceFrom` | `alert_source.go` | `source_id` 主键字段，query/json 均使用 `sourceId`。 |
| `SourceInfo` | `alert_source.go` | `source_name` 唯一，`source_type` 表示来源类型。 |
| `AlertEnrichRule` | `backend/pkg/model/integration/alert/alert_enrich.go` | `TableName() == "alert_enrich_rules"`。 |
| `AlertEnrichCondition` | `alert_enrich.go` | `enrich_rule_id` 和 `source_id` 建索引。 |
| `AlertEnrichSchemaTarget` | `alert_enrich.go` | 保存 schema field 到 target tag 的映射。 |
| `TargetTag` | `backend/pkg/model/integration/alert/target_tag.go` | `TableName() == "target_tags"`。 |
| `Cluster` | `backend/pkg/model/integration/cluster.go` | 默认 gorm 表名应为 `clusters`。 |
| `AlertEnrichRuleConfigRequest` | `backend/pkg/model/integration/alert/req.go` | 保存 source enrich rule 的请求结构。 |
| `CreateSchemaRequest` | `req.go` | 创建 schema 的请求结构。 |
| `UpdateSchemaDataRequest` | `req.go` | 更新 schema 数据的请求结构。 |

## Database Mapping

| Storage | Evidence | Notes |
|---|---|---|
| `alert_sources` | `AlertSource.TableName()` | source 基础信息、params、pull 状态。 |
| `target_tags` | `TargetTag.TableName()` | 标准化 target tag。 |
| `alert_enrich_rules` | `AlertEnrichRule.TableName()` | enrich rule 主表。 |
| `alert_enrich_conditions` | `AutoMigrate(&alert.AlertEnrichCondition{})` | gorm 默认表名，保存条件。 |
| `alert_enrich_schema_targets` | `AutoMigrate(&alert.AlertEnrichSchemaTarget{})` | gorm 默认表名，保存 schema target。 |
| `alert_source2_clusters` | `AutoMigrate(&alert.AlertSource2Cluster{})` | gorm 默认表名，保存 source-cluster 关系。 |
| `alert_input_schema_<schema>` | `SchemaPrefix = "alert_input_schema_"` | 动态 schema table。 |
| ClickHouse alert events | `ckRepo.InsertAlertEvent` | enrich 后的事件写入 ClickHouse，具体表结构归告警事件文档追踪。 |

## Validation And Rules

| Area | Implemented Rule |
|---|---|
| Source ID | `CreateAlertSource` 在未传 `sourceId` 时生成 uuid；传入时必须能被 `uuid.Parse` 解析。 |
| Source Name | `CreateAlertSource` 和 `UpdateAlertSource` 会通过 dispatcher map 检查 source name 冲突。 |
| Pull Provider | `SetupAlertProvider` 要求 `EnabledPull == true`、provider 存在且 `SupportPull == true`。 |
| Provider Params | 支持 pull 的 provider 会通过 `provider.ValidateJSON(source.Params.Obj, pType.ParamSpec)` 校验参数。 |
| Webhook Provider | `SetupProviderWebhook` 要求 provider 存在、`SupportWebhookInstall == true` 且 factory 能生成 provider。 |
| Schema Name | `AllowSchema = regexp.MustCompile("^[a-zA-Z0-9_-]{1,40}$")`。 |
| Schema Column | 动态 schema table 和 schema target column 也受 `AllowSchema` 限制。 |
| Enrich Rule | `createAlertSource` 会对 schemaMapping 的 schema、schemaSource、schemaTargets 做合法性校验。 |
| Default Rule | `GetDefaultAlertEnrichRule` 先找指定 `sourceType`，找不到时回退 `json`，再找不到返回空规则。 |

## Error Handling

| Area | Behavior |
|---|---|
| Request Binding | handler 使用 `ParamBindError` 返回 400。 |
| Source Duplicate | 创建或更新 source 时识别 `ErrAlertSourceAlreadyExist`，返回 `AlertSourceAlreadyExisted`。 |
| Source Missing | 更新 enrich rule 时识别 `ErrAlertSourceNotExist`，返回 `AlertSourceNotExisted`。 |
| Event Receive | body 读取失败返回 `AcceptAlertEventFailed`。 |
| Event Process | `ProcessAlertEvents` 返回错误时，handler 返回 `ProcessAlertEventFailed`。 |
| Schema Validation | schema/table/column 不匹配 `AllowSchema` 时返回 `ErrNotAllowSchema`，由 handler 对应接口包装为各自错误码。 |
| Provider Webhook | webhook setup 失败返回 HTTP 500 和 `SetupAlertProviderWebhookError`。 |

## Notes

- `service.New` 中 `difyRepo` 赋值发生在加载已有 source 和初始化 provider 之后；如果 provider pull 在初始化阶段立即使用 `s.difyRepo`，需要确认是否存在时序风险。
- `ProcessAlertEvents` 中如果 `dispatcher.DispatchEvents` 返回 `ErrAlertSourceNotExist` 且请求带 `sourceId`，代码直接返回 `nil`，调用方会得到成功响应但事件未写入。
- `ProcessAlertEvents` 中自动创建默认 source 的错误处理存在可疑分支：`initDefaultAlertSource` 后无论 `err` 是否为空都会继续返回 `fmt.Errorf("enrich alertEvent failed, err: %v", err)`。
- `func_listtargettags.go` 的 swagger 注释路径写成 `/api/alertsinput/enrich/tags/list`，实际 router 是 `/api/alertinput/enrich/tags/list`。
- provider registry 中 `prometheus`、`zabbix`、`json` 当前为 `BasicEncoder`，不支持 pull 或 webhook install。
