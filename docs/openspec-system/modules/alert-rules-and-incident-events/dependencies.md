# Alert Rules And Incident Events Dependencies

## Upstream Dependencies

| Dependency | Evidence | Notes |
|---|---|---|
| Permission Menu Config | `frontend/src/core/contexts/UserContext.js`, `backend/pkg/repository/database/init_menu_item.go` | 菜单和详情 route 是否可见依赖权限配置。 |
| License Middleware | `backend/pkg/router/router_api.go` | 所有 `/api/alerts` 接口先经过 `LicMiddleware`。 |
| Data Group Context | `dataGroupReducer`, `common.DataGroupStorage` | 前端传 `groupId`，后端事件列表用它展开子组。 |
| Time Range Context | `timeRangeReducer` | 告警事件列表和过滤元数据依赖全局时间范围。 |
| Group Label Bootstrap | `frontend/src/App.js`, `/api/alerts/rule/groups` | 告警规则 group 选择和编辑回填依赖启动时加载的 group label。 |
| Quick Metrics Table | `quick_alert_rule_metric` | PromQL 编辑器的快捷指标来自 DB。 |
| Target Tags Table | `target_tags` | 告警事件 tag display 的展示名来自 DB。 |
| Dify Config | `config.Get().Dify`, `difyRepo` | 事件分类、alert check id、cache minutes 依赖 Dify 配置和 repository。 |
| Alert Receiver Config | `config.Get().AlertReceiver` | 决定 receiver CRUD 和静默能力走内置 receiver 还是外部 AlertManager config。 |

## Downstream Dependencies

| Dependency | Evidence | Notes |
|---|---|---|
| Alert Sources And Enrichment | `/api/alerts/inputs/alertmanager` -> `alertinput.ProcessAlertEvents` | 本模块保留兼容输入入口，但实际事件接入处理属于告警来源接入域。 |
| Automated Diagnosis And Workflows | `WorkflowsIframe`, `/dify/app/...`, `alertanalyze` route 共用 `/api/alerts` namespace | 告警事件页只负责打开 workflow，工作流定义和异常分析主体应另文档化。 |
| Conversational Diagnosis And Reporting | `frontend/src/pro/views/report`, `/api/alerts/events/report/get` | 报告页面展示细节属于报告模块，本模块提供数据读取接口。 |
| Service Observability | `tryGetAlertService` | 后端为告警 workflow 参数推断 service/endpoint，会依赖 Prometheus 与 ClickHouse 中的服务拓扑数据。 |
| DingTalk External Receiver | `service_forward_to_dingtalk.go` | DingTalk webhook 是外部通知目标。 |
| External Alert Rule Status | `/alert/api/v1/rules`, `/alert/-/reload` | 前端 wrapper 调用外部 AlertManager/VMAlert 风格接口。 |

## Shared Components

| Component | File | Notes |
|---|---|---|
| `BasicTable` | `frontend/src/core/components/Table/basicTable` | 规则、通知、事件、历史事件都使用或复用类似表格模式。 |
| `BasicCard` | `frontend/src/core/components/Card/BasicCard` | 规则和通知页面的统一容器。 |
| `LoadingSpinner` | `frontend/src/core/components/Spinner` | 本模块各页面的 loading 显示。 |
| `Empty` | `frontend/src/core/components/Empty/Empty` | 过滤值或报告为空时使用。 |
| `MonacoEditorWrapper` | `frontend/src/core/components/Editor/MonacoEditor` | 告警规则 PromQL 编辑器。 |
| `DataGroupTreeSelector` | `frontend/src/core/components/DataGroup/DataGroupTreeSelector` | 规则表单内选择数据组。 |
| `WorkflowsIframe` | `frontend/src/oss/views/workflows/workflowsIframe` | 事件列表和详情中的 workflow 嵌入。 |
| Notification Forms | `frontend/src/oss/views/alerts/modal/compontents/*` | email、webhook、dingtalk、wechat 表单片段。 |

## Shared Utilities

| Utility | File | Notes |
|---|---|---|
| Request Wrapper | `frontend/src/core/utils/request.js` | 所有 `core/api/alerts.js` wrapper 共用。 |
| Notify Utility | `frontend/src/core/utils/notify` | 规则和通知保存/删除后的用户提示。 |
| Time Conversion | `frontend/src/core/utils/time` | 告警事件详情和历史事件时间展示。 |
| i18n | `frontend/src/i18n.js`, `frontend/public/locales` | 页面文字、tag、状态、tooltip 都依赖多语言资源。 |
| Redux Store | `groupLabelReducer`, `dataGroupReducer`, `timeRangeReducer`, `settingReducer` | 规则、事件、过滤、JSON 展示主题都读取共享状态。 |
| Backend Core Context | `backend/pkg/core` | handler 的 `ShouldBindJSON`、`Payload`、`AbortWithError`、语言读取都来自 core context。 |
| Backend Error Codes | `backend/pkg/code` | 参数错误和业务错误统一使用 code 包。 |
| Backend Query Builder | `backend/pkg/repository/clickhouse` | 告警事件 ClickHouse查询使用共享 query builder。 |

## External Services

| External Service | Evidence | Notes |
|---|---|---|
| AlertManager Or VMAlert | `/alert/api/v1/rules`, `/alert/-/reload`, `kubernetes.DefaultAlertRuleFile` | 规则状态和规则配置与外部告警引擎有关。 |
| Dify | `/dify/app/...`, `difyRepo.WorkflowsRun` | 分类 workflow、alert check logs 和 run-once 页面依赖 Dify。 |
| Kubernetes API | `kubernetes.Repo` | 外部模式下规则和 AlertManager config 存在 ConfigMap 中。 |
| ClickHouse | `clickhouse.Repo` | 告警事件、workflow record、通知 record、报告存储依赖 ClickHouse。 |
| Relational Database | `database.Repo` | DingTalk receiver、quick metrics、target tags、静默配置等依赖 DB。 |
| DingTalk | `notification.SendNotification` | DingTalk receiver 转发到外部 webhook。 |
| Prometheus | `prometheus.Repo` | `tryGetAlertService` 会通过 Prometheus 从告警 tag 反推 service/endpoint。 |

## Dependency Notes

1. 本模块不是独立 CRUD 系统，而是规则配置、事件读模型、workflow 状态和通知配置的组合模块。
2. `/api/alerts` namespace 中混入 `alertanalyze` 接口，后续分析自动诊断模块时必须复用同一 router 文件，但不要把它们写成本模块主体能力。
3. 告警事件输入路径当前横跨 `/api/alerts/inputs/alertmanager` 与 `/api/alertinput/*`，后者应由 `alert-sources-and-enrichment` 详细展开。
4. 报告读取接口在本模块后端，但报告页面结构属于 `conversational-diagnosis-and-reporting`。
