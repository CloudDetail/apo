# Data Ingestion And OneAgent Backend

## API Summary

| Category | Method | API Path | Handler | Notes |
|---|---|---|---|---|
| Static Config | GET | `/api/integration/configuration` | `GetStaticIntegration` | 返回 trace/metric/log 默认配置和版本信息。 |
| Runtime Var | GET | `/api/integration/vars/:variable` | `GetIntegrationVar` | 当前只支持 `baseUrl` 和 `serverAddr`。 |
| Cluster | GET | `/api/integration/cluster/list` | `ListCluster` | 返回 cluster 列表。 |
| Cluster | GET | `/api/integration/cluster/get` | `GetCluster` | 返回 cluster integration 及版本信息。 |
| Cluster | POST | `/api/integration/cluster/create` | `CreateCluster` | 新建 cluster integration。 |
| Cluster | POST | `/api/integration/cluster/update` | `UpdateCluster` | 更新 cluster integration。 |
| Cluster | GET | `/api/integration/cluster/delete` | `DeleteCluster` | 删除 cluster 及其 integration config。 |
| Install | GET | `/api/integration/cluster/install/config` | `GetIntegrationInstallConfigFile` | 下载安装配置文件。 |
| Install | GET | `/api/integration/cluster/install/cmd` | `GetIntegrationInstallDoc` | 返回安装 markdown，service 已标记 `Deprecated`。 |
| Adapter | GET | `/api/integration/adapter/update` | `TriggerAdapterUpdate` | 触发 adapter trace API 刷新。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 注册 `/api/integration` route group。 |
| `backend/pkg/api/integration/handler.go` | 声明 handler interface 并创建 service。 |
| `backend/pkg/api/integration/func_createcluster.go` | 创建 cluster integration。 |
| `backend/pkg/api/integration/func_getcluster.go` | 读取 cluster integration。 |
| `backend/pkg/api/integration/func_updatecluster.go` | 更新 cluster integration。当前错误分支为空。 |
| `backend/pkg/api/integration/func_deletecluster.go` | 删除 cluster integration。 |
| `backend/pkg/api/integration/func_getintegrationinstallconfigfile.go` | 输出文件下载响应头和配置文件内容。 |
| `backend/pkg/api/integration/func_getintegrationinstalldoc.go` | 返回安装 markdown。 |
| `backend/pkg/api/integration/func_getintegrationvar.go` | 返回 `baseUrl` 或 `serverAddr`。 |
| `backend/pkg/api/integration/func_triggeradapterupdate.go` | 触发 adapter 更新。 |

## Services

| Service | File | Behavior |
|---|---|---|
| `New` | `service.go` | 根据 `config.Get().Server.BaseURL` 初始化 `baseURL` 与 `serverAddr`。 |
| `GetStaticIntegration` | `service_static_integration.go` | 从配置读取默认 ClickHouse/VictoriaMetrics，并从数据库读取最新 trace API。 |
| `CreateCluster` | `service_create_cluster.go` | 检查 cluster name 是否重复，生成 uuid，清洗 collector 地址，写入 cluster 与 integration config。 |
| `UpdateClusterIntegration` | `service_update_cluster_integration.go` | 清洗 collector 地址，更新 cluster 和 integration config。 |
| `GetClusterIntegration` | `service_get_cluster_integration.go` | 读取 cluster integration，调用 `RemoveSecret()` 隐藏 secret 字段。 |
| `ListCluster` / `DeleteCluster` | `service_clusters.go` | 查询 cluster 列表，删除 cluster 及其逻辑删除的 integration config。 |
| `GetIntegrationInstallConfigFile` | `service_integration_config.go` | 根据 cluster type 渲染 `apo-one-agent-values.yaml.tmpl` 或 `installCfg.sh.tmpl`。 |
| `GetIntegrationInstallDoc` | `service_integration_doc.go` | 渲染 `install.md.tmpl`；函数注释已标记 `Deprecated`。 |
| `TriggerAdapterUpdate` | `service_trigger_adapter_update.go` | 读取最新 trace APIs，POST 到 adapter 服务。 |

## Models And Schemas

| Model | File | Mapping |
|---|---|---|
| `Cluster` | `backend/pkg/model/integration/cluster.go` | 基础 cluster 表记录。 |
| `ClusterIntegration` | `cluster.go` + `integration_config.go` | 聚合 cluster、trace、metric、log。 |
| `TraceIntegration` | `integration_config.go` | 表 `trace_integrations`，包含 `mode`、`apm_type`、`trace_api`、`self_collect_config`。 |
| `MetricIntegration` | `integration_config.go` | 表 `metric_integrations`，包含 `ds_type`、`metric_api`。 |
| `LogIntegration` | `integration_config.go` | 表 `log_integrations`，包含 `db_type`、`log_api`、`self_collect_config`。 |
| `APOCollector` | `integration_config.go` | 安装配置中使用的 collector 地址和端口结构。 |
| `AdapterAPIConfig` | `integration_config.go` | adapter 更新接口需要的 APIs + timeout + lastUpdateTS。 |
| `GetCInstallRequest` | `req.go` | 安装文档和配置文件请求。 |
| `GetCInstallDocResponse` | `resp.go` | 安装 markdown 返回体。 |
| `GetCInstallConfigResponse` | `resp.go` | 配置文件下载响应模型。 |

## Database Mapping

| Storage | Evidence | Notes |
|---|---|---|
| `clusters` | `AutoMigrate(&integration.Cluster{})` | cluster 基础表。 |
| `trace_integrations` | `AutoMigrate(&integration.TraceIntegration{})` | trace 接入配置表。 |
| `metric_integrations` | `AutoMigrate(&integration.MetricIntegration{})` | metric 接入配置表。 |
| `log_integrations` | `AutoMigrate(&integration.LogIntegration{})` | log 接入配置表。 |
| `alert_source2_clusters` | `dao_manage_cluster.go` 删除 cluster 时先删 `alert.AlertSource2Cluster` | 说明 cluster 还被告警来源模块复用。 |

## Validation And Rules

| Area | Implemented Rule |
|---|---|
| Cluster Name | `CreateCluster` 先调用 `CheckClusterNameExisted`，重复则返回 `cluster name already exists`。 |
| Cluster ID | 新建 cluster 若未传 `id`，服务端自动生成 uuid。 |
| APOCollector Address | `CreateCluster` 和 `UpdateClusterIntegration` 都会调用 `RemoveHttpPrefix()` 去掉 `http://` 和尾部 `/`。 |
| Metric Integration | `CreateCluster` 与 `UpdateClusterIntegration` 会强制设置 `Metric.DSType = "self-collector"`、`Name = "APO-DEFAULT-VM"`、`Mode = "pql"` 并写入 VM 地址。 |
| Log Integration | `CreateCluster` 与 `UpdateClusterIntegration` 会强制设置 `Log.DBType = "self-collector"`、`Name = "APO-DEFAULT-CK"`、`Mode = "sql"` 并写入 ClickHouse 地址。 |
| Trace API Secret | `GetClusterIntegration` 会调用 `RemoveSecret()`；`updateTraceIntegration` 在 sidecar 更新时会 `AcceptExistedSecret()` 保留旧 secret。 |
| Integration Propagation | 当 `IndependentTraceAPI`、`IndependentMetricDatasource`、`IndependentLogDatabase` 为 false 时，部分更新会同步影响相同类型的其他 integration。 |

## Error Handling

| Area | Behavior |
|---|---|
| Request Binding | handler 使用 `ParamBindError` 返回 400。 |
| Create Cluster | 创建失败返回 `CreateClusterFailed`。 |
| Get Cluster | 读取失败返回 `GetClusterIntegrationFailed`。 |
| Delete Cluster | 删除失败返回 `DeleteClusterFailed`。 |
| Install Doc | 安装 markdown 获取失败返回 `GetIntegrationInstallDocFailed`。 |
| Install Config | 安装配置获取失败返回 `GetIntegrationInstallConfigFileFailed`。 |
| Update Cluster | `func_updatecluster.go` 当前没有在 error 分支里返回错误，存在成功误报风险。 |

## Notes

- `GetStaticIntegration` 从 `config.Get()` 读取默认 ClickHouse/VictoriaMetrics 配置，同时从数据库读取最新 trace API，这意味着“静态配置”其实是“配置文件 + 数据库”的混合输出。
- `GetIntegrationInstallConfigFile` 和 `GetIntegrationInstallDoc` 都会在 `collectorGatewayAddr` 为空时使用 `s.serverAddr` 兜底。
- `convert2DeployValues` 会根据 trace/metric/log 配置派生 `_modes`，并生成 `_java_agent_type`、`_cluster_id`、`_is_minimal`、`_base_url_` 等模板变量。
- `convert2DeployValues` 先按 cluster type 写入一组 `_deploy_version` / `_app_version`，随后又统一用 `apoComposeDeployVersion` 覆盖 `_deploy_version`。
- `TriggerAdapterUpdate` 的目标地址默认是 `http://apo-apm-adapter-svc:8079/trace/api/update`，可以被 `ADAPTER_SERVICE_ADDRESS` 覆盖。
