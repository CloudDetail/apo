# Alert Sources And Enrichment Dependencies

## Upstream Dependencies

| Dependency | Source | Usage |
|---|---|---|
| Menu And Router Metadata | `backend/pkg/repository/database/init_menu_item.go`, `init_router.go` | 提供 `/integration/alerts` 的菜单项、父级菜单和时间选择器隐藏设置。 |
| Route Permission Shell | `frontend/src/core/components/AuthRouter.tsx` | 页面被包装，但当前前端权限检查未实际执行。 |
| Request Utility | `frontend/src/core/utils/request.js` | `alertInput.ts` 统一使用 `get`、`post`。 |
| Notification Utility | `frontend/src/core/utils/notify` | 前端保存和删除后的成功提示。 |
| Ant Design Form/Table/Drawer | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/*` | 页面交互主要依赖 Ant Design 组件。 |
| i18n Namespace | `useTranslation('core/alertsIntegration')` | 页面文案通过 `core/alertsIntegration` namespace 读取。 |
| License Middleware | `backend/pkg/middleware/lic_middleware.go` | `/api/alertinput` route group 的后端准入控制。 |

## Downstream Dependencies

| Dependency | Direction | Usage |
|---|---|---|
| ClickHouse Alert Event Storage | Backend -> `clickhouse.Repo` | `ProcessAlertEvents` 和 provider pull 会调用 `InsertAlertEvent`。 |
| Dify Alert Event Submission | Backend -> `dify.DifyRepo` | 入站和 pull 事件会调用 `SubmitAlertEvents`。 |
| Alert Event Listing | This Module -> `alert-rules-and-incident-events` | 本模块写入的事件会被告警事件列表和详情读取。 |
| Automated Diagnosis | This Module -> `automated-diagnosis-and-workflows` | 本模块提交给 Dify 的事件会进入后续 workflow/diagnosis 边界。 |
| Provider Webhook External Systems | Backend -> `datadog` / `pagerduty` provider | `SetupProviderWebhook` 调用 provider 的 webhook setup。 |
| Provider Pull External Systems | Backend -> `datadog` provider | `EnabledPull` 开启时，支持 pull 的 provider 会拉取事件。 |

## Shared Components

| Component | File | Notes |
|---|---|---|
| `BasicCard` | `frontend/src/core/components/Card/BasicCard` | 页面容器使用。 |
| `DatasourceItem` | `frontend/src/core/views/IntegrationCenter/components/DatasourceItem` | 左侧 source 类型卡片复用数据接入中心组件。 |
| `LoadingSpinner` | `frontend/src/core/components/Spinner` | enrich rule 读取和保存时展示 loading。 |
| `AuthRouter` | `frontend/src/core/components/AuthRouter.tsx` | route 包装层。 |
| `AlertIntegrationProvider` | `frontend/src/core/contexts/AlertIntegrationContext.tsx` | 本模块页面级共享状态，不是全局 redux。 |

## Shared Utilities

| Utility | File | Notes |
|---|---|---|
| `get` / `post` | `frontend/src/core/utils/request.js` | 前端 API wrapper 的基础请求函数。 |
| `notify` | `frontend/src/core/utils/notify` | 成功提示。 |
| `uuidv4` | `uuid` | 前端新建 source 时生成 `sourceId`。 |
| `provider.ValidateJSON` | `backend/pkg/services/integration/alert/provider` | provider pull 启动前校验 `Params`。 |
| `util.ValidateSQL` | `backend/pkg/util` | schema table 创建 SQL 执行前校验。 |
| `driver.InitSQL` | `backend/pkg/repository/database/driver` | 初始化 `target_tags`、`alert_sources`、`alert_enrich_rules` 的种子数据。 |

## External Services

| Service | Evidence | Notes |
|---|---|---|
| ClickHouse | `ckRepo.InsertAlertEvent` | enrich 后的 alert events 落库。 |
| Dify | `difyRepo.SubmitAlertEvents` | 事件提交给 Dify，具体 workflow 处理属于其他模块。 |
| Datadog | `provider/datadog.go` | registry 显示支持 pull 和 webhook install。 |
| PagerDuty | `provider/pagerduty.go` | registry 显示支持 webhook install，但 TODO pull support。 |
| Prometheus | `provider.BasicEncoder("Prometheus")` | 当前不支持 pull 或 webhook install。 |
| Zabbix | `provider.BasicEncoder("Zabbix")` | 当前不支持 pull 或 webhook install。 |
| JSON | `provider.BasicEncoder("JSON")` | 当前不支持 pull 或 webhook install。 |

## Cross Module Links

| Module | Linkage |
|---|---|
| `alert-rules-and-incident-events` | `/api/alerts/inputs/alertmanager` 入口实际调用本模块 `ProcessAlertEvents`；本模块写入的事件由告警事件列表和详情消费。 |
| `data-ingestion-and-oneagent` | 前端共用 IntegrationCenter 的来源卡片、图标资源和接入中心视觉结构，但业务数据接入属于另一个模块。 |
| `identity-access-and-organization` | 菜单、feature 和 route permission 配置影响 `/integration/alerts` 可见性。 |
| `platform-administration` | license middleware 控制本模块后端 API 可用性。 |

## Notes

- 前端本模块没有使用全局 `dataGroupReducer` 或 `timeRangeReducer`，这是与告警事件列表模块的明显区别。
- schema 管理后端能力会影响 enrich rule 表单可选项，但当前主前端没有 schema CRUD 页面。
- cluster 关联模型存在于后端和类型中，但当前主前端不暴露选择器。
