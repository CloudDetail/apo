# Data Ingestion And OneAgent Module Overview

## Module Goal
本模块负责配置观测数据接入入口与 OneAgent 安装资产。前端通过 `/integration/data` 和 `/integration/data/settings` 管理 cluster 级 trace、metric、log 接入配置；后端通过 `/api/integration/*` 维护 cluster 与 integration config，并生成 Kubernetes 与 VM 场景下的安装命令和配置文件。仓库中的 `one-agent` 目录和安装模板共同构成该模块的部署资产边界。

## Scope
当前代码证据支持以下范围：

- cluster 列表查看、新增、编辑、删除。
- cluster 级 trace、metric、log 接入配置表单。
- APO collector 地址和端口配置。
- 基于 cluster type 的安装命令展示与安装配置文件下载。
- 最小化接入场景下的安装命令与配置模板渲染。
- Kubernetes `apo-one-agent-values.yaml` 与 VM `installCfg.sh` 模板渲染。
- 默认 chart version 与 deploy version 输出，以及通过环境变量覆盖。
- trace adapter 更新触发。
- `probe-management` 最小化接入流程和 EE 模式下的 `BotManagement` 侧栏。
- 仓库内 `one-agent` README 与 git submodule 边界。

当前代码不应纳入本模块主体范围的能力：

- `alert-sources-and-enrichment`：`/integration/alerts` 和 `/api/alertinput/*`。
- `dataplane-and-topology-governance`：`/api/dataplane/*` 与 dataplane 类型配置主体。
- `conversational-diagnosis-and-reporting`：`/agent` 页面上的 AGUI chat 逻辑。
- `identity-access-and-organization`：菜单、角色、团队、权限治理。
- `platform-administration`：license、deploy mode、branding 和平台配置。

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Route | `frontend/src/routes.js` | `/integration/data`、`/integration/data/settings`。 |
| Route | `frontend/src/pro/routes.js` | `/probe-management`、`/probe-management/settings`。 |
| Page | `frontend/src/core/views/IntegrationCenter/DataIntegration/index.tsx` | cluster 列表页。 |
| Page | `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/index.tsx` | 接入配置页。 |
| Page | `frontend/src/pro/views/probeInstaller/index.tsx` | 步骤式安装向导，当前未观察到路由接入。 |
| API Wrapper | `frontend/src/core/api/integration.ts` | 本模块主要前端 API 封装。 |
| External Wrapper | `frontend/src/core/api/appConfig.js` | `BotManagement` 使用的外部 AI 配置边界。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Router Group | `backend/pkg/router/router_api.go` | `/api/integration` route group，当前整组使用 `LicMiddleware()`。 |
| Handler Package | `backend/pkg/api/integration` | 本模块主要 handler。 |
| Service Package | `backend/pkg/services/integration` | cluster、静态配置、模板渲染、adapter 更新。 |
| Model Package | `backend/pkg/model/integration` | cluster、trace/metric/log integration、请求响应结构和 JSON helper。 |
| Repository | `backend/pkg/repository/database/integration` | cluster 和 integration config 持久化。 |
| Asset Templates | `backend/static/integration-tmpl` | 安装说明和安装配置生成模板，包含普通模式与 minimal 模式模板。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| `Cluster` | 观测接入的管理单位，包含 `id`、`name`、`clusterType`、`APOCollector` 和 `isMinimal`。 |
| `ClusterIntegration` | cluster + trace + metric + log 的组合配置对象。 |
| `APOCollector` | 安装和接入流程使用的 collector 地址与端口配置。 |
| `TraceIntegration` | trace 接入配置，支持 `self-collector`、`sidecar`、`collector` 等模式。 |
| `MetricIntegration` | metric 接入配置，当前服务端会强制使用默认 VM 配置。 |
| `LogIntegration` | log 接入配置，当前服务端会强制使用默认 ClickHouse 配置。 |
| `isMinimal` | 最小化模式标志，`/probe-management` 保存时会设置为 `true`。 |
| `agentCollectorMode` | 由 `convert2DeployValues` 派生出的安装模板模式列表。 |
| `apoChartVersion` / `apoComposeDeployVersion` | 安装模板使用的 chart/deploy 版本，可由环境变量覆盖。 |
| OneAgent | 仓库文档中定义的多组件部署集合，包含 `preload`、`ilogtail`、`odigos`、`node-agent` 等子模块。 |

## Related Files

- `frontend/src/routes.js`
- `frontend/src/pro/routes.js`
- `frontend/src/core/api/integration.ts`
- `frontend/src/core/api/appConfig.js`
- `frontend/src/core/views/IntegrationCenter/DataIntegration/index.tsx`
- `frontend/src/core/views/IntegrationCenter/DataIntegration/ClusterTable.tsx`
- `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/index.tsx`
- `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/SettingsForm/index.tsx`
- `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/InstallCmd/index.tsx`
- `frontend/src/core/views/IntegrationCenter/DataIntegration/BotManagement.tsx`
- `frontend/src/pro/views/probeInstaller/index.tsx`
- `frontend/src/pro/views/probeInstaller/ProbeForm/index.tsx`
- `backend/pkg/router/router_api.go`
- `backend/pkg/api/integration/handler.go`
- `backend/pkg/services/integration/service.go`
- `backend/pkg/services/integration/service_static_integration.go`
- `backend/pkg/services/integration/service_create_cluster.go`
- `backend/pkg/services/integration/service_update_cluster_integration.go`
- `backend/pkg/services/integration/service_integration_config.go`
- `backend/pkg/services/integration/service_integration_doc.go`
- `backend/pkg/services/integration/service_trigger_adapter_update.go`
- `backend/pkg/model/integration/cluster.go`
- `backend/pkg/model/integration/integration_config.go`
- `backend/pkg/repository/database/integration/dao_integration_config.go`
- `backend/static/integration-tmpl/kubernetes/apo-one-agent-values.yaml.tmpl`
- `backend/static/integration-tmpl/dockercompose/installCfg.sh.tmpl`
- `one-agent/README.md`
- `.gitmodules`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `AuthRouter` | 所有本模块路由都被包装，但前端实际 route 权限检查仍为注释状态。 |
| Frontend Shared | `portsDefault`, `traceItems`, `metricsItems`, `logsItems` | 配置表单和安装命令 UI 的基础元数据。 |
| Frontend Shared | `notify` | 保存和删除成功提示。 |
| Backend Shared | `LicMiddleware` | `/api/integration` 整组受 license 控制。 |
| Backend Repo | `database.Repo` | cluster 和 integration config 的主持久化边界。 |
| Backend Asset | `static/integration-tmpl/*` | 安装说明与配置文件输出依赖模板。 |
| External Boundary | `/ai/api/v1/app-config/*` | `BotManagement` 使用的外部 AI 配置接口。 |
| Cross Module | `dataplane-and-topology-governance` | `probeInstaller` 第二步复用 dataplane 组件。 |

## Risks And Unknowns

1. `probeInstaller` 页面存在完整实现，但当前未观察到前端路由接入，可能是未启用或遗留方案。
2. `createDataIntegrationApi` 和 `updateDataIntegrationApi` 后端会强制 metric/log 指向默认 VM/CH 配置，前端表单可编辑范围与真实持久化范围并不完全一致。
3. `/api/integration` 当前只观察到 `LicMiddleware()`，未观察到 `AuthMiddleware()`。
4. `func_updatecluster.go` 中出错分支为空，更新失败时 handler 仍会 `c.Payload("ok")`，需要重点确认。
5. `getClusterInstallPackageApi` 前端 wrapper 指向的路径在主后端路由中未见实现。
6. `BotManagement` 依赖 `/ai/api/v1/app-config/*` 外部接口，主后端中未见对应实现。
