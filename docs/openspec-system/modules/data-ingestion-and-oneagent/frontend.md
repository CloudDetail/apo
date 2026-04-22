# Data Ingestion And OneAgent Frontend

## Pages

| Page | Route | Description |
|---|---|---|
| `DataIntegrationPage` | `/integration/data` | cluster 列表页，渲染 `ClusterTable`。 |
| `IntegrationSettings` | `/integration/data/settings` | 普通接入配置页，保存 cluster 后可切换到安装命令 tab。 |
| `DataIntegrationPage` | `/probe-management` | 复用 cluster 列表页，但按 pathname 进入最小化模式；在 EE 版本时显示 `BotManagement`。 |
| `IntegrationSettings` | `/probe-management/settings` | 复用接入配置页，但按 pathname 进入最小化模式。 |
| `ProbeInstallerPage` | 未观察到路由接入 | 步骤式安装向导，代码存在但当前 routes 未使用。 |

## Routes

| File | Route | Component | Notes |
|---|---|---|---|
| `frontend/src/routes.js` | `/integration/data` | `DataIntegrationPage` | `hideSystemTimeRangePicker: true`。 |
| `frontend/src/routes.js` | `/integration/data/settings` | `IntegrationSettings` | 普通配置模式。 |
| `frontend/src/pro/routes.js` | `/probe-management` | `ProbeManagementPage` | 实际导向同一个 `DataIntegration` 页面。 |
| `frontend/src/pro/routes.js` | `/probe-management/settings` | `IntegrationSettings` | 最小化配置模式。 |

## Menus

| Source | Key | Route | Notes |
|---|---|---|---|
| `backend/pkg/repository/database/init_menu_item.go` | `dataIntegration` | `/integration/data` | 父级菜单为 `integration`。 |
| `backend/pkg/repository/database/init_feature_mapping.go` | `数据接入` | `dataIntegration` | feature 到菜单存在初始化映射。 |
| `backend/pkg/repository/database/init_router.go` | `/integration/data`, `/integration/data/settings` | `HideTimeSelector: true` | 说明该模块页面不依赖全局时间范围。 |

## Key Components

| Component | File | Implemented Behavior |
|---|---|---|
| `ClusterTable` | `frontend/src/core/views/IntegrationCenter/DataIntegration/ClusterTable.tsx` | 读取 cluster 列表，支持编辑、删除和安装命令弹窗。 |
| `InstallCmd` | `Integration/InstallCmd/index.tsx` | 根据 `clusterType` 展示 k8s 或 vm 安装说明，并支持下载配置文件。 |
| `IntegrationSettings` | `Integration/index.tsx` | 按 query 中的 `clusterId` 决定编辑现有 cluster 还是新建 cluster。 |
| `SettingsForm` | `Integration/SettingsForm/index.tsx` | 保存 cluster 名称、clusterType、APO collector、trace、metric、log 配置。 |
| `APOCollectorFormItem` | `SettingsForm/APOCollectorFormItem/index.tsx` | 配置 collector 地址和高级端口信息；在最小化模式下文案会去掉 `APO`。 |
| `TraceFormItem` | `SettingsForm/TraceFormItem/index.tsx` | 根据 `apmType` 和 `mode` 动态渲染 sidecar/self-collector 字段。 |
| `MetricsFormItem` | `SettingsForm/MetricsFormItem/index.tsx` | 仅暴露 `DsTypeRadio` 等轻量配置。 |
| `LogsFormItem` | `SettingsForm/LogsFormItem/index.tsx` | 暴露 log 自采模式 `full` / `sample`。 |
| `BotManagement` | `DataIntegration/BotManagement.tsx` | `probe-management` 页面右侧 Slack bot 管理，包含 Team ID 输入和状态轮询。 |
| `ProbeInstallerPage` | `frontend/src/pro/views/probeInstaller/index.tsx` | 三步式引导：probe 基础配置、dataplane 配置、安装完成。 |
| `ProbeForm` | `probeInstaller/ProbeForm/index.tsx` | 向导第一步，收集 cluster 和 APO collector 信息。 |
| `DataplaneStep` | `probeInstaller/DataplaneStep/index.tsx` | 复用 dataplane 组件，作为向导第二步。 |

## State Management

| State | File | Notes |
|---|---|---|
| URL Query | `Integration/index.tsx` | `clusterId`、`clusterType`、`activeKey` 决定编辑对象和 tab。 |
| Local State | `ClusterTable.tsx` | 保存列表数据、安装弹窗状态和当前 clusterInfo。 |
| Local Form | `SettingsForm/index.tsx` | 使用 Ant Design `Form` 管理 cluster integration 配置。 |
| Local State | `InstallCmd/index.tsx` | 保存安装 markdown、处理配置下载。 |
| Local State | `BotManagement.tsx` | 保存 Slack 状态、连接中状态、倒计时和轮询定时器。 |
| Wizard State | `probeInstaller/index.tsx` | 保存 `current` 步骤、`probeParams`、`installStep` 和 `readonly`。 |

## API Usage

| UI Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 列表读取 | `getIntegrationClusterListApi` | `/api/integration/cluster/list` | cluster 列表页主数据源。 |
| 新建 cluster | `createDataIntegrationApi` | `/api/integration/cluster/create` | `SettingsForm` 和 `ProbeInstallerPage` 都会调用。 |
| 更新 cluster | `updateDataIntegrationApi` | `/api/integration/cluster/update` | 配置页编辑保存。 |
| 删除 cluster | `deleteClusterIntegrationApi` | `/api/integration/cluster/delete` | 列表页删除动作。 |
| 读取静态配置 | `getIntegrationConfigApi` | `/api/integration/configuration` | 新建配置页用于回填 trace/metric/log 默认配置。 |
| 读取 cluster integration | `getClusterIntegrationInfoApi` | `/api/integration/cluster/get` | 编辑配置页用于加载已有 cluster integration。 |
| 读取安装说明 | `getClusterInstallCmdApi` | `/api/integration/cluster/install/cmd` | 安装命令弹窗和安装 tab 使用。 |
| 下载安装配置 | `getClusterInstallConfigApi` | `/api/integration/cluster/install/config` | 下载 `agent-values.yaml` 或 `installCfg.sh`。 |
| 读取环境变量 | `getIntegrationVarApi` | `/api/integration/vars/:variable` | 最小化模式下回填 `serverAddr`。 |
| Slack Team ID upsert | `upsertAppConfigApi` | `/ai/api/v1/app-config/upsert` | 外部 AI 边界。 |
| Slack 状态查询 | `getSlackStatusApi` | `/ai/api/v1/app-config/slack-status` | 外部 AI 边界。 |

## User Interaction Flow

1. 用户进入 `/integration/data`，`ClusterTable` 调用 `/api/integration/cluster/list`。
2. 用户点击新增，跳转 `/integration/data/settings`。
3. `IntegrationSettings` 调用 `/api/integration/configuration` 读取静态 trace/metric/log 配置。
4. 用户填写 cluster 名称、cluster type 和 APO collector，必要时配置 trace sidecar/self-collector、log 采集模式等。
5. 保存时 `SettingsForm` 调用 `/api/integration/cluster/create` 或 `/api/integration/cluster/update`。
6. 保存成功后前端把 `clusterId`、`clusterType` 和 `activeKey=install` 写入 query，并切换到安装命令页。
7. 安装命令页调用 `/api/integration/cluster/install/cmd` 获取 markdown，并调用 `/api/integration/cluster/install/config` 下载配置文件。
8. 在 `/probe-management` 路径下，页面复用相同列表和配置逻辑，但保存时会把 `isMinimal` 置为 `true`，并在 EE 模式下显示 `BotManagement`。

## Permissions

| Area | Evidence | Notes |
|---|---|---|
| Route Guard | `AuthRouter` | 相关 route 都被包装，但 `AuthRouter` 实际权限检查 effect 仍被注释。 |
| Menu Visibility | `init_menu_item.go`, `init_feature_mapping.go` | `dataIntegration` 菜单项与 `数据接入` feature 存在初始化映射。 |
| Backend Middleware | `router_api.go` | `/api/integration` 使用 `LicMiddleware()`。 |

## Notes

- `ProbeInstallerPage` 虽然实现了多步骤向导，但当前 routes 未引用，不能写成现网主入口。
- `SettingsForm` 在普通 `/integration/data/settings` 路径会展示 trace、metric、log 配置；在 `/probe-management/settings` 最小化路径仅展示更少字段。
- `BotManagement` 依赖 `/ai/api/v1/app-config/*`，不属于当前主后端 `/api/integration` 能力。
- 安装命令页中 `deployVersion` 前端默认写死为 `v1.11.000`，而后端模板也维护 `apoComposeDeployVersion`；两侧版本来源需要重点对照。
