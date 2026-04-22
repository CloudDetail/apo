# Automated Diagnosis And Workflows Dependencies

## Upstream Dependencies

| Dependency | Source | Usage |
|---|---|---|
| Service APIs | `core/api/service` | 告警分析和服务分析都复用 service 关系与告警事件接口。 |
| Redux Time Range | `timeRangeReducer` | 控制分析窗口。 |
| Redux Data Group | `dataGroupReducer` | 控制查询数据范围。 |
| Charts Context | `src/core/contexts/ChartsContext` | 影响面列表内嵌图。 |
| Data Source Filter | `src/core/components/Filter/DataSourceFilter` | 影响面列表过滤器。 |
| BasicTable / Table UI | `src/core/components/Table/basicTable` | 告警事件、影响面和记录列表使用。 |

## Downstream Dependencies

| Dependency | Direction | Usage |
|---|---|---|
| Chat Root Cause Page | Frontend -> `/alert-analyze/chatAI/:service/:endpoint` | 影响面列表的 root cause analysis 按钮跳转。 |
| Alert Event Governance | This Module -> `alert-rules-and-incident-events` | 被分析的 alert events 来源于告警事件域。 |
| Report And Conversation | This Module -> `conversational-diagnosis-and-reporting` | analyze 页面和 workflow 结果会引导到 chat/report 侧。 |

## Shared Components

| Component | File | Notes |
|---|---|---|
| `WorkflowsIframe` | `src/oss/views/workflows/workflowsIframe.tsx` | 告警事件页和工作流页都复用该 iframe 方案。 |
| `ServiceEndpointTag` | `src/pro/components/ServiceEndpointTag` | 告警分析和影响面列表展示 service + endpoint。 |
| `ChartTempCell` | `src/core/components/Chart/ChartTempCell` | 影响面列表内联小图。 |
| `LoadingSpinner` | `src/core/components/Spinner` | 多个分析页面通用 loading。 |

## Shared Utilities

| Utility | File | Notes |
|---|---|---|
| `getStep` | `src/core/utils/step` | 计算 step，用于趋势和 delta 查询。 |
| `convertUTCToLocal` | `src/core/utils/time` | 告警事件时间展示。 |
| `notify` | `src/core/utils/notify` | 突变检测保存成功提示。 |
| `lodash.debounce` | `AlertAnalyzeContext`, `ServiceAnalyzeContext` | 避免频繁触发 delta 请求。 |

## External Services

| Service | Evidence | Notes |
|---|---|---|
| Dify Console | `frontend/src/core/api/workflows.ts` | `/dify/console/api/*` 登录和控制台边界。 |
| Dify Workflow API | `backend/pkg/repository/dify/dao.go`, `client.go` | `/v1/workflows/run`。 |
| Polaris Analyzer | `service_get_descendant_contributaion.go` | 下游贡献度排序。 |
| Prometheus | `service_defects_detect.go`, `endpoints_map.go` | mutation check、服务实例、下游数据库查询。 |
| Kubernetes | `service_defects_detect.go` | 同步创建告警规则。 |

## Cross Module Links

| Module | Linkage |
|---|---|
| `alert-rules-and-incident-events` | 告警事件来源、workflow 按钮和 workflow record 展示与本模块互相依赖。 |
| `service-observability` | 共享 service relation、service alert events 和 endpoint 关联逻辑。 |
| `trace-exploration-and-profiling` | `GetAnomalySpan` 返回的 trace/span 信息会被故障报告或 trace 页面继续消费。 |
| `conversational-diagnosis-and-reporting` | 影响面列表可直接打开 chat 根因分析页面。 |

## Notes

- 本模块并不拥有 Dify console 的全部实现，只拥有前端入口与后端 Dify client/repo 边界。
- 影响面分析和服务分析强依赖 service / topology 领域，因此文档中必须明确跨模块引用，而不能写成完全独立的 AI 子系统。
