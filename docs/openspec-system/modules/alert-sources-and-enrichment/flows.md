# Alert Sources And Enrichment Flows

## Main Flow 1

### Source Creation And Push URL Flow

1. 用户进入 `/integration/alerts`。
2. `AlertsIntegrationTable` 调用 `/api/alertinput/source/list` 获取 `alertSources`。
3. 用户在 `AlertsDatasourceList` 中选择 `json`、`zabbix` 或 `prometheus`。
4. 前端设置 URL query `sourceId=add&sourceType=<type>` 并打开 `IntegrationDrawer`。
5. `BaseInfoContent` 进入编辑态，前端生成 `uuid` 并展示 `/api/alertinput/event/source?sourceId=<uuid>`。
6. 用户填写 `sourceName` 后保存，前端调用 `/api/alertinput/source/create`。
7. 后端 `CreateAlertSource` 检查 source name 是否冲突，校验或生成 uuid，写入 `alert_sources`。
8. 后端调用 `initDefaultAlertSource`，为该 source 创建内存 enricher 和默认 enrich rule。
9. 后端调用 `SetupAlertProvider`，仅当 `EnabledPull == true` 且 provider 支持 pull 时启动 pull。
10. 前端收到返回的 `sourceId` 后更新 URL query，抽屉进入已有 source 的配置模式。

## Main Flow 2

### Source Enrich Rule Editing Flow

1. 用户打开已有 source 的配置抽屉。
2. `IntegrationDrawer` 调用 `/api/alertinput/source/get` 读取 `sourceName`、`sourceType` 和 `clusters`。
3. `AlertIntegrationProvider` 在抽屉打开时调用 `/api/alertinput/schema/listwithcolumns` 与 `/api/alertinput/enrich/tags/list`。
4. `TagContactRule` 调用 `/api/alertinput/source/enrich/get` 读取 `enrichRuleConfigs`。
5. 前端将后端返回的 `tagMapping` rule 转为带 `targetTag` 的表单结构。
6. 前端将非 `tagMapping` rule 转为 `schemaObject`、`schemaTargets` 和 `schemaFields` 表单结构。
7. 用户编辑 conditions、`fromField`、`fromRegex`、target tag、schema source 或 schema target。
8. 保存时前端组装 `sourceId` 和 `enrichRuleConfigs`，调用 `/api/alertinput/source/enrich/update`。
9. 后端 `UpdateAlertEnrichRule` 从 dispatcher 读取旧 enricher，构造新 enricher。
10. 后端对比旧 rule 和新 rule，删除不存在的 rule，重写被修改的 conditions、schema targets 和 rule 主记录，新增新 rule。
11. 后端用 `AddOrUpdateAlertSourceRule` 刷新 dispatcher 中的内存 enricher。

## Main Flow 3

### Alert Event Ingestion Flow

1. 外部系统向 `/api/alertinput/event/source?sourceId=<sourceId>` 或 `/api/alertinput/event/json` POST 原始 body。
2. handler 从 query 绑定 `SourceFrom`，再读取 request body。
3. handler 调用 `ProcessAlertEvents`。
4. `Dispatcher.DispatchEvents` 先通过 `sourceId` 查找 `EnricherMap`；若没有 `sourceId`，则通过 `sourceName` 查找 `SourceName2EnricherMap`。
5. dispatcher 调用 `decoder.Decode(*source, data)`，把 payload 转成 `[]AlertEvent`。
6. dispatcher 调用 source 对应的 `AlertEnricher.Enrich(events)`。
7. service 调用 `difyRepo.SubmitAlertEvents(events)`。
8. service 调用 `ckRepo.InsertAlertEvent(ctx, events, source)` 写入 ClickHouse。
9. handler 返回 `"ok"`。

## Main Flow 4

### Provider Pull Flow

1. 后端启动或 source 创建/更新后调用 `SetupAlertProvider`。
2. 如果 `source.EnabledPull` 为 false，provider 不会启动；如果已存在 provider instance，则会取消并删除。
3. 如果 `source.EnabledPull` 为 true，后端检查 `ProviderRegistry[source.SourceType]` 是否存在且 `SupportPull` 为 true。
4. 后端校验 `source.Params.Obj` 是否符合 provider `ParamSpec`。
5. 后端创建 provider instance 并启动 `keepPullAlert` goroutine。
6. pull 线程按 `last_pull_mill_ts` 和 interval 计算拉取窗口，最多把过早的 last pull 限制到最近 15 天。
7. provider `PullAlerts` 返回事件后，后端调用 `DispatchDecodedEvents` 执行 enrich。
8. 后端提交 Dify，写入 ClickHouse，并更新 `last_pull_mill_ts`。

## Edge Cases

| Case | Implemented Behavior |
|---|---|
| 新建 source 未传 `sourceId` | 后端自动生成 uuid。 |
| 新建 source 传入非法 uuid | 后端返回 `sourceID is not a valid uuid`。 |
| 新建或更新 source name 冲突 | 后端返回 `ErrAlertSourceAlreadyExist`，handler 转成 `AlertSourceAlreadyExisted`。 |
| enrich rule 保存时 source 不存在 | `UpdateAlertEnrichRule` 返回 `ErrAlertSourceNotExist`。 |
| schema/table/column 名称不合法 | DAO 或 service 返回 `ErrNotAllowSchema`。 |
| schema 被删除 | service 会遍历内存 enricher 并调用 `RemoveRuleByDeletedSchema(schema)`。 |
| provider 不支持 pull | `SetupAlertProvider` 返回 `provider <sourceType> not support pull`。 |
| provider 不支持 webhook | `SetupProviderWebhook` 返回 `alert source not support setup webhook now`。 |
| `/api/alertinput/event/json` 未传 source type | 后端使用 `"unkonwn"`。 |
| `/api/alertinput/event/json` 未传 source name | 后端使用 `<sourceType>-(<clientIP>)`。 |

## Failure Paths

| Path | Failure Behavior |
|---|---|
| body 读取失败 | handler 返回 `AcceptAlertEventFailed`。 |
| decode 失败 | dispatcher 返回 `decode alertEvent failed` 包装错误，handler 返回 `ProcessAlertEventFailed`。 |
| source 不存在且请求带 `sourceId` | `ProcessAlertEvents` 直接返回 `nil`，事件被丢弃但接口表现为成功。 |
| source 不存在且请求不带 `sourceId` | service 尝试 `initDefaultAlertSource`，但随后仍返回 `enrich alertEvent failed` 包装错误。 |
| ClickHouse 写入失败 | `ProcessAlertEvents` 返回 `InsertAlertEvent` 的错误。 |
| provider pull 失败 | goroutine 记录日志并等待下一轮，不直接反馈到前端。 |
| provider 拉取后的 ClickHouse 写入失败 | 记录日志并返回，不更新 `lastPullTime`。 |

## Permission Constraints

| Layer | Constraint |
|---|---|
| Frontend Route | `/integration/alerts` 被 `AuthRouter` 包装，但当前 `AuthRouter` 权限检查 effect 被注释。 |
| Backend API | `/api/alertinput` route group 当前只观察到 `LicMiddleware()`。 |
| Menu | `alertsIntegration` 菜单受后端菜单和 feature 初始化配置影响。 |
| Event Input | `/api/alertinput/event/source` 和 `/api/alertinput/event/json` 位于同一 `LicMiddleware` group 下，未观察到额外 auth。 |

## To Be Confirmed

- `ProcessAlertEvents` 对带 `sourceId` 但 source 未创建的事件直接返回成功并丢弃，是否符合业务预期需要确认。
- `schemaMapping` 与前端 `staticEnrich` 命名差异是否由 enricher 明确兼容，需要进一步跟踪。
- provider pull 在 `service.New` 里可能早于 `s.difyRepo = difyRepo` 启动，是否存在 nil 使用风险需要确认。
