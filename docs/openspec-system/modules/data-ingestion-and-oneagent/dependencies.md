# Data Ingestion And OneAgent Dependencies

## Upstream Dependencies

| Dependency | Source | Usage |
|---|---|---|
| Menu And Router Metadata | `init_menu_item.go`, `init_router.go`, `init_feature_mapping.go` | 定义 `/integration/data` 入口和 feature 映射。 |
| Request Utility | `frontend/src/core/utils/request.js` | `integration.ts` 和 `appConfig.js` 的请求基础。 |
| Notification Utility | `frontend/src/core/utils/notify` | 保存和删除成功提示。 |
| Integration Constants | `frontend/src/core/views/IntegrationCenter/constant.tsx` | `traceItems`、`portsDefault` 等元数据。 |
| AuthRouter | `frontend/src/core/components/AuthRouter.tsx` | 路由包装层。 |
| License Middleware | `backend/pkg/middleware/lic_middleware.go` | `/api/integration` API 准入控制。 |

## Downstream Dependencies

| Dependency | Direction | Usage |
|---|---|---|
| OneAgent Deployment | Backend -> `backend/static/integration-tmpl/*` | cluster integration 会被渲染成 Helm values 或 VM 安装脚本。 |
| Adapter Service | Backend -> `apo-apm-adapter-svc` | `TriggerAdapterUpdate` 推送最新 trace API 配置。 |
| OneAgent Submodules | Repo -> `one-agent/node-agent`, `preload`, `ilogtail`, `odigos` | 安装与部署资产边界。 |
| Slack App Config | Frontend -> `/ai/api/v1/app-config/*` | `BotManagement` 依赖外部或旁路 AI 配置接口。 |

## Shared Components

| Component | File | Notes |
|---|---|---|
| `BasicCard` | `src/core/components/Card/BasicCard` | cluster 列表、probe installer 等页面容器。 |
| `AuthRouter` | `src/core/components/AuthRouter.tsx` | route guard。 |
| `APOCollectorFormItem` | `src/core/components/Integration/APOCollectorFormItem/index.tsx` | `ProbeForm` 复用组件。 |
| `CopyPre` | `src/core/components/CopyPre` | 安装命令页面用于复制命令。 |

## Shared Utilities

| Utility | File | Notes |
|---|---|---|
| `JSONField[T]` | `backend/pkg/model/integration/config2json.go` | integration config 的 JSON 序列化、secret 替换和 secret 复用。 |
| `CompressMD` | `backend/pkg/model/integration/compress_md.go` | markdown 存储辅助。 |
| Template Rendering | `text/template` in `service_integration_config.go` | 渲染安装说明和配置文件。 |
| `portsDefault` | `frontend/src/core/views/IntegrationCenter/constant.tsx` | APO collector 端口缺省值。 |

## External Services

| Service | Evidence | Notes |
|---|---|---|
| ClickHouse | `GetStaticIntegration`, `forceSetupMetricLogAPI` | 作为默认 log 存储配置源。 |
| VictoriaMetrics | `GetStaticIntegration`, `forceSetupMetricLogAPI` | 作为默认 metric 存储配置源。 |
| APM Sidecar Providers | `TraceAPI` model | `skywalking`、`jaeger`、`nbs3`、`arms`、`huawei`、`elastic`、`pinpoint`。 |
| Adapter Service | `service_trigger_adapter_update.go` | `/trace/api/update` 外部调用目标。 |
| Slack OAuth | `BotManagement.tsx` | 打开 `https://slack.com/oauth/v2/authorize...` 弹窗。 |
| `/ai/api/v1/app-config/*` | `appConfig.js` | Slack 状态与 teamId 写入接口。 |

## Cross Module Links

| Module | Linkage |
|---|---|
| `alert-sources-and-enrichment` | 共用 `Cluster` 模型；删除 cluster 时还会删除 `alert_source2_clusters` 关系。 |
| `dataplane-and-topology-governance` | `probeInstaller` 第二步复用 `DataplaneRadioList` 和 `DataplaneConfigForm`。 |
| `identity-access-and-organization` | 普通入口 `dataIntegration` 菜单和 feature 可见性受权限配置影响。 |
| `platform-administration` | `LicMiddleware`、chart/deploy version 和 server `BaseURL` 配置影响本模块行为。 |

## Notes

- `probe-management` 没有单独后端 API，它复用同一组 `/api/integration/*`。
- 安装命令既依赖后端模板，也依赖前端本地拼接出的补充命令，两者都应作为模块实现的一部分记录。
- `BotManagement` 是本模块内的独立 tab，但其数据来源属于外部 AI 边界，不能和 `/api/integration` 混写。
