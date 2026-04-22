# Dataplane And Topology Governance Dependencies

## Upstream Dependencies

| Dependency | Evidence | Why It Matters |
|---|---|---|
| `timeRangeReducer` | `frontend/src/pro/views/agentDataplane/index.tsx` | `serviceMatcher` 依赖时间范围查询服务和实例。 |
| `request.js` | `frontend/src/pro/api/dataplane.ts` | 为本模块所有 `/api/dataplane/*` 请求统一注入认证与语言头。 |
| `integration.CheckInvalid` | `backend/pkg/services/dataplane/service_data_plane.go` | 数据平面参数校验依赖集成模型定义。 |
| ClickHouse Repo | `backend/pkg/services/dataplane/service_get_service.go` 等 | 服务、实例、拓扑、已解析应用等核心数据都来自 ClickHouse。 |
| Prometheus Repo | `service_data_plane.go` | `agentReady` 判断依赖 Prometheus。 |
| DB Repo | `dao_dataplane.go`, `dao_service_name_rule*.go` | 治理配置和规则持久化基础。 |

## Downstream Dependencies

| Consumer | Evidence | Dependency Relationship |
|---|---|---|
| `conversational-diagnosis-and-reporting` | `frontend/src/pro/contexts/AGUIContext.tsx`, `DataStatus.tsx` | Agent 页面依赖本模块的 `data/check` 和数据状态能力。 |
| `data-ingestion-and-oneagent` | `frontend/src/pro/views/probeInstaller/DataplaneStep/index.tsx` | 安装流程复用数据平面类型和配置表单。 |
| 观测主页面 | `service_get_service.go`, `service_get_service_topology.go` | 服务、实例、拓扑查询结果会影响后续观测页面可见的服务归因结构。 |

## Shared Components

| Shared Component | File | Role In This Module |
|---|---|---|
| `BasicCard` | `src/core/components/Card/BasicCard` | `/dataplane` 页面主容器。 |
| `DateTimeRangePickerCom` | `frontend/src/pro/views/agentDataplane/components/DateTimeRangePickerCom` | `serviceMatcher` 时间选择器。 |
| `notify` | `src/core/utils/notify` | 配置成功、删除成功、测试结果提示。 |
| `SetupGuide` | `frontend/src/pro/components/Dataplane/DataplaneGuide` | 连接指导文档视图。 |

## Shared Utilities

| Utility | File | Role |
|---|---|---|
| `getTimestampRange` | `timeRangeReducer` | 默认时间范围计算。 |
| `useDebounce` | `react-use` | 限制服务列表请求频率。 |
| `determineInstanceSource` | `backend/pkg/services/dataplane/service_get_service.go` | 基于 labels 计算实例归因来源。 |
| `ServiceNameRuleCondition.Match` | `dao_service_name_rule_condition.go` | 规则条件匹配基础逻辑。 |

## External Services

| Service | Evidence | Current Role |
|---|---|---|
| ClickHouse | `backend/pkg/repository/dataplane/*`, `chRepo.QueryServices` 等 | 查询服务、实例、拓扑、待归因应用。 |
| Prometheus | `promRepo.GetMinimalAgentExisted` | 计算 `agentReady`。 |
| 外部数据平面系统 | `data_plane.params` | 各 `typeName` 对应的外部平台连接配置对象，本仓库只保存配置并做基本测试。 |

## Notes

- 本模块既被 Agent 入口消费，也被安装流程消费，因此它处在“接入治理”和“诊断入口前置条件”之间的枢纽位置。
- `dataplaneRepo` 与 `dbRepo` 同时存在，说明模块同时依赖治理侧配置库和观测侧查询库，文档中不能只写其中一层。
