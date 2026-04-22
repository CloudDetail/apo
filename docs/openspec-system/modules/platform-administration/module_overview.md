# Platform Administration Module Overview

## Module Goal
本模块负责平台级配置与治理，包括数据保留周期、license 生命周期、品牌 logo 资源，以及部署模式驱动的品牌切换。它既包含面向管理员的配置页，也包含全局布局中的 license 告警和品牌资源消费。

## Scope
当前代码证据支持以下范围：

- `/config` 页面中的 TTL 分类配置与单表 TTL 调整。
- `/system/platform-manage` 页面中的当前 license 查询与上传。
- 全局 `LicenseContext` 对 license 状态的读取和缓存。
- 全局布局中基于 `licenseInfo.active` 的 banner 告警。
- `/api/app/upload`、`/api/app/logo` 与 `/api/app/deployMode` 的平台品牌资源读取能力。
- `DeployModeContext` 基于 deploy mode 的前端品牌切换逻辑。

当前代码不应直接写入本模块主范围的内容：

- 可视化 branding 配置页，因为相关 UI 在 `platformManage/index.tsx` 中已注释。
- 任何未在当前页面出现的 trial 申请流程；`UnlicensedCard` 中 trial 部分仍被注释。
- 通过数据库或管理员页面持久化 deploy mode；当前后端只直接返回 `PRODUCT_MODE` 环境变量。

## Frontend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Config Route | `frontend/src/oss/views/config/index.jsx` | `/config`。 |
| Platform Route | `frontend/src/pro/routes.js` | `/system/platform-manage`。 |
| Global Context | `frontend/src/pro/contexts/LicenseContext.tsx` | 全局 license 状态。 |
| Global Context | `frontend/src/core/contexts/DeployModeContext.js` | 全局品牌状态。 |
| Global Layout | `frontend/src/core/layout/DefaultLayout.js` | layout 级 license 提示。 |

## Backend Entry

| Entry Type | Evidence | Notes |
|---|---|---|
| Config API | `backend/pkg/api/config` | TTL 读写。 |
| License API | `backend/pkg/api/license` | license 上传和读取。 |
| App API | `backend/pkg/api/app` | logo 上传、读取，以及 `deployMode` 返回。 |
| License Service | `backend/pkg/services/license` | 激活、解析、定时检查。 |
| App Service | `backend/pkg/services/app/logo.go` | 文件校验、SVG 清洗。 |

## Core Concepts

| Concept | Explanation |
|---|---|
| TTL | ClickHouse 数据保留周期，既可按数据类别改，也可按单表改。 |
| `license.active` | license 当前状态标识，前端根据它决定告警和页面展示。 |
| `LicenseContext` | 当前登录用户下的 license 全局缓存。 |
| `DeployModeContext` | 根据 `/api/app/deployMode` 返回的部署模式在 APO / Syncause 之间切换 logo、favicon、标题。 |
| `UploadLicenseRequest.output` | 前端上传 license 时发送给后端的字段。 |
| `Logo` | 存储在 DB 的平台 logo 资源。 |

## Related Files

- `frontend/src/oss/views/config/index.jsx`
- `frontend/src/oss/views/config/configTTL/ConfigTTL.jsx`
- `frontend/src/oss/views/config/configTTL/TTLTable.jsx`
- `frontend/src/pro/views/platformManage/index.tsx`
- `frontend/src/pro/views/LicenseManage/index.tsx`
- `frontend/src/pro/views/LicenseManage/components/LicenseInfoCard.tsx`
- `frontend/src/pro/views/LicenseManage/components/UnlicensedCard.tsx`
- `frontend/src/pro/views/LicenseManage/components/modals/UploadModal.tsx`
- `frontend/src/core/api/config.js`
- `frontend/src/pro/api/license.ts`
- `frontend/src/core/api/app.js`
- `frontend/src/pro/contexts/LicenseContext.tsx`
- `frontend/src/core/contexts/DeployModeContext.js`
- `frontend/src/core/layout/DefaultLayout.js`
- `backend/pkg/api/config/handler.go`
- `backend/pkg/api/license/handler.go`
- `backend/pkg/api/app/handler.go`
- `backend/pkg/api/app/func_getdeploymode.go`
- `backend/pkg/services/config/service_getttl.go`
- `backend/pkg/services/config/service_setttl.go`
- `backend/pkg/services/license/service_upload_license.go`
- `backend/pkg/services/license/service_get_license.go`
- `backend/pkg/services/license/service_license_check.go`
- `backend/pkg/services/app/logo.go`
- `backend/pkg/model/license/license.go`
- `backend/pkg/model/request/license_req.go`
- `backend/pkg/model/request/config_req.go`
- `backend/pkg/model/request/app_req.go`
- `backend/pkg/model/response/license_resp.go`

## Dependencies

| Dependency Type | Dependency | Why It Matters |
|---|---|---|
| Frontend Shared | `request.js` | 统一认证和请求处理。 |
| Frontend Shared | `DefaultLayout` | banner 告警显示平台状态。 |
| Cross Module | `identity-access-and-organization` | 只有登录后 `LicenseContext` 才会拉取当前 license。 |
| Backend Shared | ClickHouse | TTL 能力直接依赖表结构与 `CREATE TABLE` 语句。 |
| Backend Shared | Prometheus | license 节点使用量统计依赖 Prometheus。 |

## Risks And Unknowns

1. `/api/app/deployMode` 已实现，但只是直接返回 `PRODUCT_MODE`，并没有体现管理员可配或数据库持久化的 deploy mode 模型。
2. `platformManage/index.tsx` 中 Logo/Title 配置 UI 被整段注释，说明 branding 交付状态不完整。
3. `LicenseManagementPage` 只把 `active === 1` 视为有 license，其他状态都落到 `UnlicensedCard`，与 `LicenseInfoCard` 中的过期分支形成部分重叠。
4. `UploadModal` 的文件读取逻辑存在双重 `FileReader` 调用痕迹，需要在后续回合确认是否稳定。
