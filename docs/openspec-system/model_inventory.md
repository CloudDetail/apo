# Model Inventory

## Scope
本文档盘点当前 `backend/pkg/model` 里的主要模型家族，并按业务模块、共享基础设施和跨模块共享模型进行归类。

## Business-Owned Model Families

| Module | Model Families | Notes |
|---|---|---|
| `identity-access-and-organization` | `backend/pkg/model/profile/*.go`, `backend/pkg/model/user.go`, `backend/pkg/model/request/user_req.go`, `role_req.go`, `team_req.go`, `backend/pkg/model/response/user_resp.go`, `team_resp.go` | 用户、角色、团队和 feature/profile 相关模型。 |
| `data-scope-governance` | `backend/pkg/model/datagroup/*.go`, `backend/pkg/model/request/data_req.go`, `backend/pkg/model/response/data_resp.go` | DataGroup、data scope、group2scope 与 V2 请求响应模型。 |
| `service-observability` | `backend/pkg/model/request/service_req.go`, `backend/pkg/model/response/service_resp.go`, `backend/pkg/model/service_instance.go`, `backend/pkg/model/related_instance.go`, `backend/pkg/model/topology.go` | service 观察和拓扑相关模型。 |
| `logs-exploration-and-parsing` | `backend/pkg/model/request/log_req.go`, `log_info.go`, `log_query.go`, `log_parse.go`, `log_table.go`, `backend/pkg/model/response/log_resp.go`, `log_info.go`, `log_query.go`, `log_parse.go`, `log_table.go`, `log_chart.go`, `log_index.go` | 日志查询、上下文、解析规则和表结构模型。 |
| `trace-exploration-and-profiling` | `backend/pkg/model/request/trace_req.go`, `span_trace.go`, `backend/pkg/model/response/trace_resp.go`, `span_trace.go`, `backend/pkg/model/flame_graph.go`, `backend/pkg/model/flame_graph_test.go` | trace 查询与 flame graph 数据结构。 |
| `alert-rules-and-incident-events` | `backend/pkg/model/alert_event.go`, `alert_event_tags.go`, `alert_report.go`, `alert_status.go`, `backend/pkg/model/amconfig/*`, `backend/pkg/model/request/alert_event.go`, `request/alert_slience.go`, `backend/pkg/model/response/alert_event.go`, `response/alert_slience.go` | 告警规则、告警事件、静默配置和 Alertmanager 配置模型。 |
| `alert-sources-and-enrichment` | `backend/pkg/model/request/alert_input.go` 中 source/schema/cluster/enrich 相关字段，`backend/pkg/model/response/alert_input.go` 中 provider/schema 相关返回 | 告警输入接入与 enrich 相关模型与响应。 |
| `automated-diagnosis-and-workflows` | `backend/pkg/model/request/alert_analyze.go`, `backend/pkg/model/mutation_detect.go`, `backend/pkg/model/anormal_event.go`, `backend/pkg/model/service_anomaly_span.go`, `backend/pkg/model/workflow_records.go` | 自动化分析、异常、缺陷检测和 workflow 记录。 |
| `conversational-diagnosis-and-reporting` | `backend/pkg/model/request/custom_req.go`, `backend/pkg/model/response/custom_resp.go` | 自定义查询与会话式页面的本仓库模型。 |
| `data-ingestion-and-oneagent` | `backend/pkg/model/integration/cluster.go`, `integration_config.go`, `req.go`, `resp.go` | 数据接入、集群配置和安装参数。 |
| `dataplane-and-topology-governance` | `backend/pkg/model/dataplane.go`, `backend/pkg/model/request/data_plane_req.go`, `backend/pkg/model/response/dataplane_resp.go`, `backend/pkg/model/integration/data_plane.go`, `data_plane_type.go` | 数据平面、拓扑和 datasource 能力模型。 |
| `platform-administration` | `backend/pkg/model/app.go`, `config_table.go`, `backend/pkg/model/license/license.go`, `backend/pkg/model/request/app_req.go`, `config_req.go`, `license_req.go`, `backend/pkg/model/response/app_resp.go`, `config_resp.go`, `license_resp.go` | app/logo、TTL、license 相关模型。 |

## Shared Infrastructure Model Families

| Shared Ownership | Model Families | Notes |
|---|---|---|
| Shared observability support | `backend/pkg/model/metric.go`, `backend/pkg/model/request/metric_req.go`, `backend/pkg/model/response/metric_resp.go`, `backend/pkg/model/request/network.go`, `backend/pkg/model/response/network.go`, `backend/pkg/model/request/k8s_req.go`, `backend/pkg/model/response/k8s_resp.go` | 这些模型为 metric/network/k8s 支撑接口服务，被多个业务模块消费。 |
| Shared utility / generic | `backend/pkg/model/constants.go`, `backend/pkg/model/model.go`, `backend/pkg/model/minheap.go`, `backend/pkg/model/middleware_instance.go` | 更适合视为基础模型或工具结构，不绑定单一业务。 |

## Cross-Module Shared Model Families

| Model Family | Shared By | Why It Should Not Be Forced Into One Module |
|---|---|---|
| `backend/pkg/model/request/alert_input.go` | `alert-rules-and-incident-events`, `alert-sources-and-enrichment`, `automated-diagnosis-and-workflows` | 文件名虽叫 `alert_input`，但其中既有告警规则、接收器、detect 表达式，也有 provider 参数结构。 |
| `backend/pkg/model/response/alert_input.go` | 同上 | 返回类型既服务告警规则，又服务自动分析和接入配置。 |
| `backend/pkg/model/workflow_records.go` | `automated-diagnosis-and-workflows`, `alert-rules-and-incident-events`, `conversational-diagnosis-and-reporting` | workflow、告警通知记录和报告记录交叉使用。 |
| `backend/pkg/model/topology.go` | `service-observability`, `dataplane-and-topology-governance` | 同时被 service 拓扑和 dataplane 自定义拓扑使用。 |

## Coverage Check

- 没有发现整个 `backend/pkg/model` 子目录完全无归属。
- 真正需要避免误挂到业务模块的，是 generic root 文件与多模块复用的 alert/observability 模型家族。
- 后续补写详细模型文档时，建议优先按“模型家族”而不是“单个 struct”建索引，否则重复度会很高。
