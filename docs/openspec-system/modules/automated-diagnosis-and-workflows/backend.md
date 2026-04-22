# Automated Diagnosis And Workflows Backend

## API Summary

| Category | Method | API Path | Handler | Notes |
|---|---|---|---|---|
| Alert Events | POST | `/api/alerts/alert/events` | `GetAlertEvents` | 告警分析页的事件列表。 |
| Impact | POST | `/api/alerts/event/impact` | `GetAlertImpact` | 入口影响面。 |
| Descendant | ANY | `/api/alerts/descendant/anormal/contribution` | `GetDescendantContribution` | 下游贡献度排序。 |
| Descendant | ANY | `/api/alerts/descendant/anormal/list` | `GetDescendantAnormal` | 下游异常列表。 |
| Descendant | ANY | `/api/alerts/descendant/anormal/trend` | `GetAnormalTrendByEntry` | 下游异常趋势。 |
| Descendant | POST | `/api/alerts/descendant/anormal/delta` | `GetDescendantAnormalDelta` | 下游异常 delta。 |
| Anomaly Span | POST | `/api/alerts/anomaly-span/list` | `GetAnomalySpan` | 故障报告用异常 span 列表。 |
| Mutation | GET | `/api/alerts/detect/mutation/metrics` | `GetPredefinedDetectExpr` | 预定义检测表达式。 |
| Mutation | POST | `/api/alerts/detect/mutation/add` | `AddDefectsDetect` | 执行突变检测。 |
| Mutation | GET | `/api/alerts/detect/mutation/exec-list` | `GetDefectDetectExecList` | 执行记录。 |
| Mutation | GET | `/api/alerts/detect/mutation/rule-list` | `GetDefectDetectRuleList` | 规则记录。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 在 `/api/alerts` group 下注册 `alertanalyze` 相关 routes，并位于 `AuthMiddleware()` 之后。 |
| `backend/pkg/api/alertanalyze/handler.go` | 构造 `alertanalyzeService`、`serviceoverviewService`、`serviceInfoService`。 |
| `func_getalertevents.go` | 告警分析页告警事件接口。 |
| `func_getalertimpact.go` | 影响面分析接口。 |
| `func_getdescendantcontribution.go` | 下游故障贡献度接口。 |
| `func_getdescendantanormal.go` | 下游异常列表接口。 |
| `func_getanormaltrendbyentry.go` | 趋势接口。 |
| `func_getdescendantanormaldelta.go` | delta 接口。 |
| `func_getanomalyspan.go` | 异常 span 接口。 |
| `func_getpredefineddetectexpr.go`, `func_adddefectsdetect.go`, `func_getdefectdetectexeclist.go`, `func_getdefectdetectrulelist.go` | 突变检测接口。 |

## Services

| Service | File | Behavior |
|---|---|---|
| `GetAlertEventsSvc` | `service_get_alert_events_svc.go` | 查询 firing + mutation 告警，推导 `impactEndpoints` 和 `impactEntrys`。 |
| `GetAlertsImpact` | `service_get_alerts_impact.go` | 聚合入口与告警的关联强度，按 `relatedAlertRate` 排序。 |
| `GetDescendantContribution` | `service_get_descendant_contributaion.go` | 列出下游节点，调用 Polaris `SortDescendantByRelevance` 按 `sortBy` 排序。 |
| `GetDescendantAnormal` | `service_get_descendant_anormal.go` | 查询下游节点告警事件，结合实例和数据库实例过滤。 |
| `GetAnormalTrendByEntry` | `service_get_descendant_anormal_trend.go` | 汇总 error 与 alert 两类异常事件，生成 chart 数据。 |
| `SearchAnormalDeltaByEntry` | `service_get_descendant_anormal_delta.go` | 生成 `OriginAnormalCounts`、`AlertTriggeredCounts`、`FinalAnormalCounts` 及事件记录。 |
| `GetAnomalySpan` | `service_get_anomaly_span.go` | 根据 service、endpoint、reason、慢/错条件查询异常 span。 |
| `DetectDefectsOptions` | `service_predefined_detect_expr_part.go` | 从数据库 quick mutation metrics 生成预定义指标。 |
| `DetectDefects` | `service_defects_detect.go` | 执行 mutation check，写入 ClickHouse，并可同步创建告警规则。 |
| `GetDefectDetectExecList` | `service_defects_exec_list.go` | 查询突变检测执行记录。 |
| `GetDefectDetectRuleList` | `service_defects_rule_list.go` | 查询规则记录，并检查是否已同步到告警规则。 |
| `EndpointsMap` | `endpoints_map.go` | 根据 `serviceName`、数据库 URL、pod/namespace、node/pid、infra node 等线索映射服务。 |

## Models And Schemas

| Model | File | Mapping |
|---|---|---|
| `GetAlertEventsSvcReq` | `backend/pkg/model/request/alert_analyze.go` | 告警分析页事件列表请求。 |
| `AlertImpactRequest` | `alert_analyze.go` | 入口影响面分析请求。 |
| `GetDescendantAnormal*Request` | `backend/pkg/model/request/service_req.go` | 下游异常列表、趋势、delta 请求。 |
| `GetAnomalySpanRequest` | `service_req.go` | 异常 span 查询请求。 |
| `GetAlertEventsSvcResp` | `backend/pkg/model/response/service_resp.go` | 告警分析事件返回。 |
| `AlertRelatedEntry` | `service_resp.go` | 受影响入口返回结构。 |
| `GetDescendantAlertContributationResponse` | `service_resp.go` | 贡献度结果。 |
| `GetDescendantAnormalEventResponse` | `backend/pkg/model/response/alert_input.go` | 下游异常列表。 |
| `GetDescendantDeltaAnormalEventResponse` | `alert_input.go` | 趋势 / delta / 事件记录结构。 |
| `WorkflowRecord` | `backend/pkg/model/workflow_records.go` | workflow 运行记录。 |
| `AlertEventClassify` | `workflow_records.go` | `alertGroup + alertName` 到 workflow 的映射结果。 |
| `DetectMutation` | `backend/pkg/model/mutation_detect.go` | 突变检测执行/规则记录。 |
| `DetectExprPart` | `mutation_detect.go` | 预定义检测表达式拼装模型。 |

## Database Mapping

| Storage | Evidence | Notes |
|---|---|---|
| ClickHouse alert events | `GetAlertEventsSvc`, `GetDescendantAnormal`, `GetAnormalTrendByEntry`, `SearchAnormalDeltaByEntry` | 告警分析和下游异常查询主来源。 |
| ClickHouse detect mutation records | `DetectDefects`, `GetDefectDetectExecList`, `GetDefectDetectRuleList` | 突变检测执行记录。 |
| Database quick mutation metric | `DetectDefectsOptions` | 预定义指标来源。 |
| Workflow records | `WorkflowRecord` model + `repository/dify` async worker | workflow 执行结果的持久化结构。 |

## Validation And Rules

| Area | Implemented Rule |
|---|---|
| Data Group | 多个 service 会根据 `groupId` 展开子数据组并限制服务范围。 |
| Default Pagination | `GetAlertEventsSvc` 和 `GetDescendantAnormal` 在无分页参数时会注入默认 `PageParam`。 |
| Descendant Types | 趋势和 delta 查询通过 `AnormalTypes` 选择 `alert`、`error`、`middleware`、`mutation-*` 等事件类型。 |
| Contribution Sort | `GetDescendantContribution` 默认 `sortBy = "latency,errorRate"`。 |
| Mutation Group | `DetectDefects` 若 `MutationCheck.Group` 为空，则默认使用 `kubernetes.MutationCustomLabelVal`。 |
| Alert Rule Sync | `DetectDefects` 在 `SynchronizeToAlertRules == true` 时会调用 `k8sRepo.AddAlertRule`。 |
| Dify Cache | async alert check 对 `alertGroup + alertName` 使用 LRU 缓存分类结果。 |
| Dify Auto Analyze | `workflow_worker.go` 在 `AutoAnalyze == true` 且 alert check 输出为 `"false"` 时触发 analyze workflow。 |

## Error Handling

| Area | Behavior |
|---|---|
| Request Binding | handler 使用 `ParamBindError`。 |
| Impact / Descendant Query | service 层错误会直接回传 handler，由 handler 返回对应错误码。 |
| Mutation Rule Sync | 同步告警规则失败时返回 `DetectDefectsCreatAlertError`。 |
| Mutation Exec Record | 插入执行记录失败时返回 `AddExecRecordError`。 |
| Dify Workflow | `repository/dify` 中 workflow API 失败会写入 `Output` 或 `AnalyzeErr` 的失败描述。 |

## Notes

- `GetAlertEventsSvc` 不是直接返回原始告警事件，而是先关联服务、影响 endpoint、入口 endpoint，再回填 `source` 到 `EnrichTags` 与 `Tags`。
- `EndpointsMap.SearchAndAddEndpointsForAlert` 会依次尝试按 `serviceName`、数据库、k8s pod、VM process、infra node 推断服务。
- `SearchAnormalDeltaByEntry` 会同时处理 error 事件和 alert 事件，并把 mutation 相关事件类型作为 `mutation-<type>` 纳入过滤。
- `repository/dify` 不直接暴露本地 HTTP route，而是作为外部工作流引擎的 client/repo 层被其他模块和本模块共同复用。
