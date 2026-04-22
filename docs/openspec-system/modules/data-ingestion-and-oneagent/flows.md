# Data Ingestion And OneAgent Flows

## Main Flow 1

### Cluster Integration Creation Flow

1. 用户进入 `/integration/data`。
2. `ClusterTable` 调用 `/api/integration/cluster/list` 获取 cluster 列表。
3. 用户点击新增，前端跳转 `/integration/data/settings`。
4. `IntegrationSettings` 调用 `/api/integration/configuration` 获取默认 trace/metric/log 配置。
5. 用户填写 cluster 名称、cluster type、APO collector、trace、metric、log 表单。
6. `SettingsForm` 保存时补全 `portsDefault`，并调用 `/api/integration/cluster/create`。
7. 后端 `CreateCluster` 检查 cluster name 是否重复，生成 uuid，清洗 collector 地址。
8. 后端强制写入默认 VictoriaMetrics 和 ClickHouse 配置。
9. 后端保存 cluster 基础信息和 integration config。
10. 前端收到返回后把 `clusterId`、`clusterType`、`activeKey=install` 写入 query，切到安装页。

## Main Flow 2

### Installation Config And Command Flow

1. 用户在列表页点击某个 cluster 的 `installCmdTitle` 按钮，或在配置页切到 `install` tab。
2. 前端调用 `/api/integration/cluster/install/cmd` 获取 markdown 安装说明。
3. 用户点击下载配置按钮时，前端调用 `/api/integration/cluster/install/config`。
4. 后端 `GetIntegrationInstallConfigFile` 读取 cluster integration 配置。
5. 后端 `convert2DeployValues` 根据 trace/metric/log 组合生成 `_modes`、版本变量、collector 地址和 cluster id 等模板变量。
6. 若 `clusterType == "k8s"`，后端渲染 `apo-one-agent-values.yaml.tmpl`；若 `clusterType == "vm"`，渲染 `installCfg.sh.tmpl`。
7. 前端根据返回的 `Content-Disposition` 头下载 `agent-values.yaml` 或 `installCfg.sh`。
8. 用户在界面看到对应的 Helm 或 Docker Compose 安装命令。

## Main Flow 3

### Probe Management Minimal Flow

1. 商业版用户进入 `/probe-management`。
2. 前端仍复用 `ClusterTable`，但 `pathname === '/probe-management'` 时 `isMinimal` 为 true。
3. 在 EE 模式下，页面右侧额外展示 `BotManagement`。
4. 用户新增 probe 时跳转 `/probe-management/settings`。
5. `SettingsForm` 因 `shouldShow === false` 进入最小化模式，只保留更少的 collector 配置，并把 `isMinimal = true`。
6. 最小化模式如果用户没有填写 `collectorGatewayAddr`，前端先调用 `/api/integration/vars/serverAddr` 获取默认地址。
7. 保存后仍调用 `/api/integration/cluster/create` 或 `/api/integration/cluster/update`，随后进入安装页。
8. 读取安装命令时，后端对 minimal cluster 走 `install-cmd-minimal.tmpl`；下载配置时 k8s 走 `apo-one-agent-values-minimal.yaml.tmpl`，vm 走 `cfg.env.tmpl`。

## Main Flow 4

### Trace Adapter Update Flow

1. 外部调用 `/api/integration/adapter/update?lastUpdateTS=<ts>`。
2. handler 绑定 `TriggerAdapterUpdateRequest`。
3. service 调用 `dbRepo.GetLatestTraceAPIs(lastUpdateTS)`。
4. DAO 先找比 `lastUpdateTS` 更新的 trace integration，再执行 SQL 取每个 `apm_type` 最新的 `trace_api`。
5. service 把结果 POST 到 `adapterServiceAddress + "/trace/api/update"`。
6. 调用方始终得到 `"ok"`，具体失败只写日志。

## Edge Cases

| Case | Implemented Behavior |
|---|---|
| cluster 名称重复 | `CreateCluster` 返回 `cluster name already exists`。 |
| 新建 cluster 未传 `id` | 后端自动生成 uuid。 |
| `collectorGatewayAddr` 为空 | 安装配置生成时会使用 `serverAddr` 兜底。 |
| 最小化模式未填 `collectorGatewayAddr` | 前端会先请求 `/api/integration/vars/serverAddr` 并回填默认值。 |
| `clusterType` 为未知值 | `getIntegrationConfigFile` / `GetIntegrationInstallDoc` 返回 `unexpected clusterType` 错误。 |
| `trace.mode == sidecar` | `updateTraceIntegration` 会尝试保留旧 secret。 |
| `IndependentTraceAPI == false` | sidecar trace API 更新可能同步影响同 `apm_type` 的其他 trace integrations。 |
| `IndependentMetricDatasource == false` | metric 更新会同步影响相同 `dsType` 的其他 metric integrations。 |
| `IndependentLogDatabase == false` | log 更新会同步影响相同 `dbType` 的其他 log integrations。 |

## Failure Paths

| Path | Failure Behavior |
|---|---|
| cluster 创建失败 | handler 返回 `CreateClusterFailed`。 |
| cluster 删除失败 | handler 返回 `DeleteClusterFailed`。 |
| cluster 读取失败 | handler 返回 `GetClusterIntegrationFailed`。 |
| 安装 markdown 获取失败 | handler 返回 `GetIntegrationInstallDocFailed`。 |
| 安装配置文件获取失败 | handler 返回 `GetIntegrationInstallConfigFileFailed`。 |
| adapter POST 失败 | `TriggerAdapterUpdate` 只写日志，接口仍返回 `"ok"`。 |
| cluster 更新失败 | `func_updatecluster.go` 当前不会向前端返回错误，存在失败误报风险。 |

## Permission Constraints

| Layer | Constraint |
|---|---|
| Frontend Route | `/integration/data`、`/integration/data/settings`、`/probe-management`、`/probe-management/settings` 被 `AuthRouter` 包装。 |
| Backend API | `/api/integration` route group 当前只观察到 `LicMiddleware()`。 |
| Menu | 普通模式通过 `dataIntegration -> /integration/data` 暴露；`PRODUCT_MODE=syncause` 时，数据库初始化会把同一菜单切到 `/probe-management`，并为 `/probe-management/settings` 建立 feature-router 映射。 |

## To Be Confirmed

- To Be Confirmed: `probeInstaller` 页面为何存在但未接入 routes。
- To Be Confirmed: `func_updatecluster.go` 出错分支为空是否为遗漏。
- To Be Confirmed: `/probe-management` 与 `/integration/data` 的最终产品边界是否长期并存。
- To Be Confirmed: `getClusterInstallPackageApi` 是否对应未提交或外部实现。
