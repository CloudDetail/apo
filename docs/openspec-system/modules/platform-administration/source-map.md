# Platform Administration Source Map

## Scope
本文档固定 `platform-administration` 模块在当前仓库中的真实代码证据。该模块覆盖 TTL 配置、license 查询与上传、logo 上传与读取、部署模式品牌切换，以及这些能力在全局布局中的可见告警。需要特别区分“已实现的 license / TTL / deploy mode 读取能力”和“已注释或未开放的 branding 配置 UI”。

## Relevant Source Directories

| Layer | Path | Notes |
|---|---|---|
| Frontend | `frontend/src/oss/views/config` | `/config` 页面，当前只展示 TTL 配置。 |
| Frontend | `frontend/src/pro/views/platformManage` | `/system/platform-manage` 平台管理页，当前只承载 License 面板。 |
| Frontend | `frontend/src/pro/views/LicenseManage` | License 查询与上传 UI。 |
| Frontend | `frontend/src/pro/api/license.ts` | `/api/license/current` 与 `/api/license/upload` wrapper。 |
| Frontend | `frontend/src/core/api/config.js` | `/api/config/getTTL|setTTL|setSingleTableTTL` wrapper。 |
| Frontend | `frontend/src/core/api/app.js` | `/api/app/deployMode` wrapper。 |
| Frontend | `frontend/src/core/contexts/DeployModeContext.js` | 应用品牌名、logo、favicon 选择。 |
| Frontend | `frontend/src/pro/contexts/LicenseContext.tsx` | 全局 license 状态。 |
| Frontend | `frontend/src/core/layout/DefaultLayout.js` | license 状态 banner 的全局展示。 |
| Backend | `backend/pkg/api/config` | TTL 配置接口。 |
| Backend | `backend/pkg/api/license` | license 上传、查询接口。 |
| Backend | `backend/pkg/api/app` | logo 上传、读取和 deploy mode 查询接口。 |
| Backend | `backend/pkg/services/config` | TTL 读写逻辑。 |
| Backend | `backend/pkg/services/license` | license 激活、查询、校验。 |
| Backend | `backend/pkg/services/app/logo.go` | logo 文件校验、清洗和读取；deploy mode 不走该 service。 |
| Backend | `backend/pkg/repository/database/license` | `license` 与 `activation` 持久化。 |
| Backend | `backend/pkg/repository/database/app` | logo 持久化。 |

## Route Files

| Side | File | Observed Routes |
|---|---|---|
| Frontend | `frontend/src/pro/routes.js` | `/system/platform-manage`。 |
| Frontend | `frontend/src/routes.js` | `/demo/service-config` 与本模块无关；平台管理主入口在 pro routes。 |
| Frontend | `frontend/src/oss/routes.js` 或主路由合并 | `/config`。 |
| Backend | `backend/pkg/router/router_api.go` | `/api/config/getTTL`, `/api/config/setTTL`, `/api/config/setSingleTableTTL`, `/api/license/upload`, `/api/license`, `/api/license/current`, `/api/app/upload`, `/api/app/logo`, `/api/app/deployMode`。 |

## Page Files

| Type | File | Notes |
|---|---|---|
| Config Page | `frontend/src/oss/views/config/index.jsx` | 当前只渲染 `ConfigTTL`。 |
| TTL Panel | `frontend/src/oss/views/config/configTTL/ConfigTTL.jsx` | 读取 TTL 分组，支持批量修改类别 TTL。 |
| TTL Table | `frontend/src/oss/views/config/configTTL/TTLTable.jsx` | 支持单表 TTL 更新。 |
| Platform Manage | `frontend/src/pro/views/platformManage/index.tsx` | 当前只展开 License management；Logo/Title 区块整段注释。 |
| License Manage | `frontend/src/pro/views/LicenseManage/index.tsx` | 根据 `licenseInfo.active` 在 `LicenseInfoCard` 和 `UnlicensedCard` 间切换。 |
| Upload Modal | `frontend/src/pro/views/LicenseManage/components/modals/UploadModal.tsx` | 接收 JSON license 文本，转为 `output` 字段上传。 |
| License Banner | `frontend/src/core/layout/DefaultLayout.js` | 根据 `licenseInfo.active` 显示 warning / error banner。 |

## API Wrapper Files

| File | Wrapper | API Path | Notes |
|---|---|---|---|
| `frontend/src/core/api/config.js` | `getTTLApi` | `/api/config/getTTL` | 读取 TTL 配置。 |
| `frontend/src/core/api/config.js` | `setTTLApi` | `/api/config/setTTL` | 按类别更新 TTL。 |
| `frontend/src/core/api/config.js` | `setSingleTableTTLApi` | `/api/config/setSingleTableTTL` | 单表 TTL 更新。 |
| `frontend/src/pro/api/license.ts` | `getLicense` | `/api/license/current` | 当前激活 license。 |
| `frontend/src/pro/api/license.ts` | `uploadLicense` | `/api/license/upload` | 上传 license。 |
| `frontend/src/core/api/app.js` | `getDeployModeApi` | `/api/app/deployMode` | 前端使用；后端 `GetDeployMode` 直接返回 `PRODUCT_MODE`。 |

## Controller Or Router Files

| Type | File | Notes |
|---|---|---|
| Config Router | `backend/pkg/api/config/handler.go` | TTL 读写接口。 |
| License Router | `backend/pkg/api/license/handler.go` | license 上传和查询接口。 |
| App Router | `backend/pkg/api/app/handler.go` | app 模块接口定义，包括 logo 与 deploy mode。 |
| Deploy Mode Handler | `backend/pkg/api/app/func_getdeploymode.go` | 直接读取 `os.Getenv("PRODUCT_MODE")` 并返回。 |

## Service Files

| Service Area | Files | Notes |
|---|---|---|
| TTL Config | `backend/pkg/services/config/service_getttl.go`, `service_setttl.go` | TTL 读取与更新。 |
| License | `backend/pkg/services/license/service_upload_license.go`, `service_get_license.go`, `service_license_check.go` | 激活、查询和定时检查。 |
| App Branding | `backend/pkg/services/app/logo.go` | 上传 logo 并做 MIME / SVG 清洗。 |

## Model Entity Schema Files

| Type | File | Notes |
|---|---|---|
| `License` / `Activation` | `backend/pkg/model/license/license.go` | `license` 与 `activation` 表。 |
| License Request | `backend/pkg/model/request/license_req.go` | `UploadLicenseRequest`，字段名为 `output`。 |
| Config Request | `backend/pkg/model/request/config_req.go` | TTL 修改请求。 |
| App Request | `backend/pkg/model/request/app_req.go` | `UploadLogoRequest`，要求 `logoText` 和 `logoImage`。 |
| License Response | `backend/pkg/model/response/license_resp.go` | license 查询返回结构。 |
| App Response | `backend/pkg/model/response/app_resp.go` | `GetLogoResp`。 |

## Important Shared Dependencies

| Dependency | File | Notes |
|---|---|---|
| `LicenseContext` | `frontend/src/pro/contexts/LicenseContext.tsx` | 平台级 license 状态来源。 |
| `DeployModeContext` | `frontend/src/core/contexts/DeployModeContext.js` | 平台级品牌切换来源。 |
| `DefaultLayout` | `frontend/src/core/layout/DefaultLayout.js` | 在全局布局中消费 license 状态。 |
| ClickHouse Repo | `backend/pkg/services/config` | TTL 能力依赖 ClickHouse。 |
| Prometheus Repo | `backend/pkg/services/license` | 计算 license 节点使用量。 |

## Evidence Notes

1. `/system/platform-manage` 当前实际只承载 `LicenseManagementPage`，logo/title 配置 UI 已整段注释。
2. `DeployModeContext` 前端会请求 `/api/app/deployMode`，当前后端已实现该接口，但逻辑只是读取 `PRODUCT_MODE` 环境变量。
3. `/config` 页面当前只渲染 `ConfigTTL`，没有其他系统配置项。
4. `UploadModal` 期望用户粘贴 JSON license 文本，并最终把整理后的 JSON 字符串放入 `output` 字段上传。
5. `LicenseContext` 和 `DefaultLayout` 说明 license 状态是平台级横切信息，而不是只在平台管理页内部显示。
