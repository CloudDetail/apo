# Data Ingestion And OneAgent Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `/integration/data` 与 `/probe-management` 的产品定位差异只能从 pathname 分支和 `isMinimal` 推断，业务上是否分别面向不同版本或不同用户角色需要确认。
2. `GetStaticIntegration` 返回的 trace/metric/log 默认配置混合了 config 文件和数据库中的最新 trace API，这种“静态配置”命名是否准确需要确认。
3. `BotManagement` 出现在 `probe-management` 页面右侧，但其业务目标是 Slack bot 管理还是安装辅助，目前只能从 UI 文案推断。

## Legacy Or Dead-Code Suspicion

1. `frontend/src/pro/views/probeInstaller/index.tsx` 存在完整实现，但当前 routes 未接入。
2. `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/Document/index.tsx` 是空白卡片，疑似未完成页面。
3. `getClusterInstallPackageApi` wrapper 存在，但主后端路由中未见对应接口。
4. `service_integration_doc.go` 和 `/api/integration/cluster/install/cmd` 已标记 `Deprecated`，但前端仍在使用。

## Missing Frontend-Backend Alignment

1. 前端 `InstallCmd` 中 `deployVersion` 固定为 `v1.11.000`，后端模板版本由 `apoComposeDeployVersion` 和环境变量控制，两侧来源并未完全统一。
2. 前端允许配置 trace、metric、log 表单，但后端创建和更新时会强制覆盖 metric/log 为默认 VM/CH 配置。
3. 前端最小化模式通过 pathname 推断 `isMinimal`，后端只接收普通 cluster integration 请求，没有显式区分专用 API。
4. `ProbeInstallerPage` 第二步依赖 dataplane 组件，而当前主 routes 并未使用该页面。

## Permission Uncertainty

1. `/api/integration` 整组当前只观察到 `LicMiddleware()`，未观察到 `AuthMiddleware()`。
2. `AuthRouter` 当前未实际执行 route 权限检查，因此前端页面保护主要依赖菜单入口和后端接口层。
3. `syncause` 模式下已经存在 `/probe-management` 菜单、router 和 `/probe-management/settings` 的 feature 初始化映射，但列表页与 settings 页的最终权限链是否完全依赖这些初始化数据，仍需要确认。

## Data Model Uncertainty

1. `DeleteCluster` 会删除 `alert_source2_clusters` 关系，说明 cluster 是跨模块共享模型，但 source-cluster 关联在接入页中没有明显 UI 展示。
2. `IndependentTraceAPI`、`IndependentMetricDatasource`、`IndependentLogDatabase` 这三个布尔开关默认值为 false，但当前仓库中未观察到配置入口。
3. `convert2DeployValues` 里 `_deploy_version` 的赋值会被后续覆盖，clusterType 专属版本与全局版本的真实优先级需要确认。
4. `AdapterAPIConfig` 只取每个 `apm_type` 最新一条 trace API，并拼接 `apmList`，adapter 侧实际消费契约需要进一步核对。

## To Be Confirmed

- To Be Confirmed: `probeInstaller` 是否计划替代当前 `/probe-management` 列表+表单模式。
- To Be Confirmed: `func_updatecluster.go` 的空错误分支是否会造成静默失败。
- To Be Confirmed: `getClusterInstallPackageApi` 指向的下载接口是否在其他服务中实现。
- To Be Confirmed: `BotManagement` 是否应独立为外部 AI 集成模块，而不是长期留在数据接入页面内。
