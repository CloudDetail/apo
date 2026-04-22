# Data Ingestion And OneAgent Source Map

## Scope
本文档固定 `data-ingestion-and-oneagent` 模块在当前仓库中的代码落点。该模块覆盖观测数据接入配置、cluster 级接入管理、安装命令与安装配置生成、OneAgent 部署模板，以及商业版 `probe-management` 的最小化接入流程。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/core/views/IntegrationCenter/DataIntegration` | `/integration/data` 与 `/integration/data/settings` 的主页面目录。 |
| Frontend | `frontend/src/core/api/integration.ts` | `/api/integration/*` 的主要 API wrapper。 |
| Frontend | `frontend/src/core/api/appConfig.js` | `BotManagement` 使用的 `/ai/api/v1/app-config/*` 外部边界。 |
| Frontend | `frontend/src/pro/views/probeInstaller` | 商业版 probe 安装向导页面，当前仓库中存在但未观察到路由接入。 |
| Frontend | `frontend/src/pro/views/agent` | `ProbeInstallerPage` 完成后会跳转到 `/agent`，当前该页面更偏 AI chat 边界。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/integration` route group 注册入口。 |
| Backend | `backend/pkg/api/integration` | 本模块 controller/handler 包。 |
| Backend | `backend/pkg/services/integration` | cluster 管理、静态配置、安装文档和安装配置文件生成。 |
| Backend | `backend/pkg/model/integration` | cluster、trace/metric/log integration、请求响应结构、JSONField 和压缩 markdown 模型。 |
| Backend | `backend/pkg/repository/database/integration` | observability input 持久化 DAO，包含 cluster 和 integration config 管理。 |
| Backend Asset | `backend/static/integration-tmpl` | Kubernetes 与 Docker Compose 的安装文档模板和配置模板，包含 minimal 模式变体。 |
| Repo Asset | `one-agent/README.md` | OneAgent 组成说明。 |
| Repo Asset | `.gitmodules` | `one-agent/node-agent`、`preload`、`ilogtail`、`odigos` 子模块声明。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/routes.js` | `/integration/data`, `/integration/data/settings` |
| Frontend | `frontend/src/pro/routes.js` | `/probe-management`, `/probe-management/settings` |
| Backend | `backend/pkg/router/router_api.go` | `/api/integration/configuration`, `/api/integration/vars/:variable`, `/api/integration/cluster/list`, `/api/integration/cluster/get`, `/api/integration/cluster/create`, `/api/integration/cluster/update`, `/api/integration/cluster/delete`, `/api/integration/cluster/install/config`, `/api/integration/cluster/install/cmd`, `/api/integration/adapter/update` |
| External Or Adjacent | `frontend/src/core/api/appConfig.js` | `/ai/api/v1/app-config/upsert`, `/ai/api/v1/app-config/slack-status` |

## Page Files

| Type | File | Notes |
|---|---|---|
| Route Page | `frontend/src/core/views/IntegrationCenter/DataIntegration/index.tsx` | `/integration/data` 页面入口，当前直接渲染 `ClusterTable`。 |
| Cluster Table | `frontend/src/core/views/IntegrationCenter/DataIntegration/ClusterTable.tsx` | cluster 列表、编辑、删除、安装命令弹窗；在 EE 且 `/probe-management` 时额外展示 `BotManagement`。 |
| Settings Page | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/index.tsx` | `/integration/data/settings` 和 `/probe-management/settings` 共用页面，根据 pathname 切换普通模式与最小化模式。 |
| Settings Form | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/SettingsForm/index.tsx` | cluster 基本信息、APO collector、trace、metric、log 接入配置。 |
| Install Commands | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/InstallCmd/index.tsx` | 根据 cluster type 与 `isMinimal` 生成普通或 minimal 的安装说明和配置下载动作。 |
| Probe Installer | `frontend/src/pro/views/probeInstaller/index.tsx` | 多步骤安装向导页面，当前仓库中存在但未观察到 route 使用。 |
| Probe Form | `frontend/src/pro/views/probeInstaller/ProbeForm/index.tsx` | 向导第一步，采集 cluster 名称、cluster type 和 APO collector 配置。 |
| Dataplane Step | `frontend/src/pro/views/probeInstaller/DataplaneStep/index.tsx` | 向导第二步，复用 dataplane 组件进行数据源配置。 |
| Bot Management | `frontend/src/core/views/IntegrationCenter/DataIntegration/BotManagement.tsx` | `probe-management` 页右侧 Slack bot 管理，依赖外部 `/ai/api/v1/app-config/*`。 |
| Empty Guide | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/Document/index.tsx` | 空白 guide 卡片，当前未见实际内容。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/integration.ts` | `getIntegrationConfigApi` | `/api/integration/configuration` | 读取静态接入配置。 |
| `frontend/src/core/api/integration.ts` | `createDataIntegrationApi` | `/api/integration/cluster/create` | 新建 cluster integration。 |
| `frontend/src/core/api/integration.ts` | `getIntegrationClusterListApi` | `/api/integration/cluster/list` | 列表页读取 cluster。 |
| `frontend/src/core/api/integration.ts` | `getClusterInstallCmdApi` | `/api/integration/cluster/install/cmd` | 读取 markdown 安装说明。 |
| `frontend/src/core/api/integration.ts` | `getClusterInstallConfigApi` | `/api/integration/cluster/install/config` | 下载 `agent-values.yaml`、`installCfg.sh`，或 minimal 模式下的 `cfg.env`。 |
| `frontend/src/core/api/integration.ts` | `getClusterInstallPackageApi` | `/cluster/integration/cluster/install/package/:type` | wrapper 存在，当前主后端 route 未观察到实现。 |
| `frontend/src/core/api/integration.ts` | `getClusterIntegrationInfoApi` | `/api/integration/cluster/get` | 读取 cluster integration 配置。 |
| `frontend/src/core/api/integration.ts` | `updateDataIntegrationApi` | `/api/integration/cluster/update` | 更新 cluster integration。 |
| `frontend/src/core/api/integration.ts` | `deleteClusterIntegrationApi` | `/api/integration/cluster/delete` | 删除 cluster integration。 |
| `frontend/src/core/api/integration.ts` | `getIntegrationVarApi` | `/api/integration/vars/:variable` | 读取 `serverAddr` 或 `baseUrl`。 |
| `frontend/src/core/api/appConfig.js` | `upsertAppConfigApi` | `/ai/api/v1/app-config/upsert` | Slack bot teamId 提交，属于外部或旁路 AI 边界。 |
| `frontend/src/core/api/appConfig.js` | `getSlackStatusApi` | `/ai/api/v1/app-config/slack-status` | Slack bot 状态查询。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Router | `backend/pkg/router/router_api.go` | `/api/integration` 整组使用 `LicMiddleware()`。 |
| Handler Factory | `backend/pkg/api/integration/handler.go` | 创建 `integration.New(database)` service。 |
| Config Handler | `backend/pkg/api/integration/func_getstaticintegration.go` | 输出静态 trace/metric/log 配置和 chart/deploy version。 |
| Var Handler | `backend/pkg/api/integration/func_getintegrationvar.go` | 输出 `baseUrl` 或 `serverAddr`。 |
| Cluster Handler | `backend/pkg/api/integration/func_createcluster.go`, `func_getcluster.go`, `func_updatecluster.go`, `func_deletecluster.go`, `func_listcluster.go` | cluster integration CRUD。 |
| Install Handler | `backend/pkg/api/integration/func_getintegrationinstalldoc.go`, `func_getintegrationinstallconfigfile.go` | 安装 markdown 和安装配置文件，minimal cluster 会切到专用模板。 |
| Adapter Handler | `backend/pkg/api/integration/func_triggeradapterupdate.go` | 触发 trace adapter 配置刷新。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| Service Construction | `backend/pkg/services/integration/service.go` | 初始化 `baseURL`、`serverAddr`，供安装配置和 var 查询使用。 |
| Static Config | `backend/pkg/services/integration/service_static_integration.go` | 从 config 和数据库组装默认 trace/metric/log 配置，并输出 `chartVersion` 和 `deployVersion`。 |
| Cluster Create | `backend/pkg/services/integration/service_create_cluster.go` | 创建 cluster，强制 metric/log 走默认 VM/CH 配置。 |
| Cluster Update | `backend/pkg/services/integration/service_update_cluster_integration.go` | 更新 cluster integration，继续强制 metric/log 指向默认 VM/CH。 |
| Cluster Read | `backend/pkg/services/integration/service_get_cluster_integration.go` | 读取 cluster integration 并隐藏 secret。 |
| Cluster List/Delete | `backend/pkg/services/integration/service_clusters.go` | 查询 cluster 列表和删除 cluster 配置。 |
| Install Config | `backend/pkg/services/integration/service_integration_config.go` | 读取 cluster integration，按普通或 minimal 模式套用 k8s/vm 模板生成配置文件。 |
| Install Markdown | `backend/pkg/services/integration/service_integration_doc.go` | 根据 cluster type 生成安装说明 markdown；minimal cluster 走 `install-cmd-minimal.tmpl`，文件注释已标记 `Deprecated`。 |
| Adapter Update | `backend/pkg/services/integration/service_trigger_adapter_update.go` | 把最新 trace API 配置 POST 到 adapter 服务。 |
| Basic Getters | `backend/pkg/services/integration/service_config.go` | 输出 `GetBaseURL`、`GetServerAddr`。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| Cluster Model | `backend/pkg/model/integration/cluster.go` | cluster 基础信息、`clusterType` 和 `APOCollector`。 |
| Integration Config | `backend/pkg/model/integration/integration_config.go` | `TraceIntegration`、`MetricIntegration`、`LogIntegration`、`TraceAPI`、`APOCollector`。 |
| Request Model | `backend/pkg/model/integration/req.go` | `GetCInstallRequest`、`TriggerAdapterUpdateRequest`、`GetIntegrationVarRequest`。 |
| Response Model | `backend/pkg/model/integration/resp.go` | `ListClusterResponse`、`GetCInstallDocResponse`、`GetCInstallConfigResponse`。 |
| JSON Helper | `backend/pkg/model/integration/config2json.go` | `JSONField[T]`、secret masking、secret value 接受逻辑。 |
| Markdown Helper | `backend/pkg/model/integration/compress_md.go` | 压缩 markdown 换行。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| Route Guard | `frontend/src/core/components/AuthRouter.tsx` | `/integration/data`、`/integration/data/settings`、`/probe-management`、`/probe-management/settings` 都经 `AuthRouter` 包装。 |
| Menu Metadata | `backend/pkg/repository/database/init_menu_item.go`, `init_feature.go`, `init_feature_mapping.go`, `init_router.go` | 默认模式初始化 `/integration/data`；`PRODUCT_MODE=syncause` 时会把 `dataIntegration` 菜单切到 `/probe-management`，并加入 `/probe-management`、`/probe-management/settings` 路由与 feature 映射。 |
| Integration Constants | `frontend/src/core/views/IntegrationCenter/constant.tsx` | `traceItems`、`metricsItems`、`logsItems`、`portsDefault` 用于表单和安装命令 UI。 |
| Shared APO Collector Form | `frontend/src/core/components/Integration/APOCollectorFormItem/index.tsx` | `ProbeForm` 复用的 APO collector 配置组件。 |
| Static Templates | `backend/static/integration-tmpl/*` | 安装说明和配置文件由模板渲染，不是硬编码字符串。 |
| OneAgent Assets | `one-agent/README.md`, `.gitmodules` | OneAgent 说明和实际子模块边界。 |
| Database Repo | `backend/pkg/repository/database/integration` | cluster 与 integration config 的主要持久化入口。 |
| External Adapter | `service_trigger_adapter_update.go` | POST 到 `http://apo-apm-adapter-svc:8079/trace/api/update` 或环境变量指定地址。 |

## Evidence Notes

1. `/probe-management` 与 `/probe-management/settings` 并不是独立前端实现，而是复用 `DataIntegration` 和 `Integration` 页面，通过 pathname 进入最小化模式。
2. `PRODUCT_MODE=syncause` 时，数据库初始化会把 `dataIntegration` 菜单指向 `/probe-management`，并加入 `/probe-management`、`/probe-management/settings` 路由与 feature 映射。
3. minimal cluster 读取安装命令时会走 `install-cmd-minimal.tmpl`；配置下载时 k8s 走 `apo-one-agent-values-minimal.yaml.tmpl`，vm 走 `cfg.env.tmpl`。
4. `frontend/src/pro/views/probeInstaller/index.tsx` 存在完整步骤式安装向导，但当前路由文件未观察到接入。
5. `createDataIntegrationApi` / `updateDataIntegrationApi` 保存时，后端会强制 metric/log 指向默认 VM/CH 配置，因此前端表单中相关高级配置并不是完全自由输入。
6. `getClusterInstallPackageApi` wrapper 指向 `/cluster/integration/cluster/install/package/:type`，当前主后端 route 中未见对应实现。
7. `BotManagement` 使用 `/ai/api/v1/app-config/*`，这不是当前主后端 `/api/integration` 实现的一部分，应作为外部边界记录。
