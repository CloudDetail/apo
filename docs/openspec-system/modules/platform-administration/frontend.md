# Platform Administration Frontend

## Pages

| Page | Route | Description |
|---|---|---|
| `ConfigPage` | `/config` | 当前只渲染 TTL 配置面板。 |
| `PlatformManage` | `/system/platform-manage` | 平台管理总页，当前只展开 License 管理。 |
| `LicenseManagementPage` | `/system/platform-manage` 内部 | 根据 license 状态渲染 `LicenseInfoCard` 或 `UnlicensedCard`。 |

## Routes

| File | Route | Component | Notes |
|---|---|---|---|
| `frontend/src/pro/routes.js` | `/system/platform-manage` | `PlatformManage` | 平台管理入口。 |
| `frontend/src/oss/views/config/index.jsx` | `/config` | `ConfigPage` | 当前系统配置入口。 |

## Menus

| Source | Key | Route | Notes |
|---|---|---|---|
| `backend/pkg/repository/database/init_menu_item.go` | `config` | `/config` | 当前 TTL 配置对应的系统菜单。 |
| `frontend/src/core/layout/DefaultLayout.js` | banner action | `/system/platform-manage` | license 告警可直接跳转到平台管理页。 |

## Key Components

| Component | File | Implemented Behavior |
|---|---|---|
| `ConfigTTL` | `frontend/src/oss/views/config/configTTL/ConfigTTL.jsx` | 调 `/api/config/getTTL` 拉取各类别表 TTL，并支持分类更新。 |
| `TTLTable` | `frontend/src/oss/views/config/configTTL/TTLTable.jsx` | 调 `/api/config/setSingleTableTTL` 修改单表 TTL。 |
| `PlatformManage` | `frontend/src/pro/views/platformManage/index.tsx` | 当前只包含一个 `Collapse` item，即 License management。 |
| `LicenseManagementPage` | `frontend/src/pro/views/LicenseManage/index.tsx` | 根据 `licenseInfo.active` 选择渲染 `LicenseInfoCard` 或 `UnlicensedCard`。 |
| `LicenseInfoCard` | `frontend/src/pro/views/LicenseManage/components/LicenseInfoCard.tsx` | 展示 license ID、active 状态、seat 使用量和有效期，并提供上传按钮。 |
| `UnlicensedCard` | `frontend/src/pro/views/LicenseManage/components/UnlicensedCard.tsx` | 在未激活或非 `active === 1` 时展示上传入口。 |
| `UploadModal` | `frontend/src/pro/views/LicenseManage/components/modals/UploadModal.tsx` | 解析 JSON license 文本，构造成 `output` 上传。 |
| `LicenseContext` | `frontend/src/pro/contexts/LicenseContext.tsx` | 在用户登录后自动拉取 `/api/license/current`。 |
| `DeployModeContext` | `frontend/src/core/contexts/DeployModeContext.js` | 读取部署模式并提供 `getLogo`、`getFavicon`、`getTitle`。 |
| `DefaultLayout` | `frontend/src/core/layout/DefaultLayout.js` | 根据 license 状态渲染 warning/error banner。 |

## State Management

| State | File | Notes |
|---|---|---|
| TTL Data | `ConfigTTL.jsx` | 本地维护各类别表 TTL 数据。 |
| License Global State | `LicenseContext.tsx` | `licenseInfo` 全局缓存。 |
| Deploy Mode Global State | `DeployModeContext.js` | `deployMode` 全局缓存，默认值是 `apo`。 |
| Upload Modal State | `LicenseInfoCard.tsx`, `UnlicensedCard.tsx` | 控制上传对话框开关。 |

## API Usage

| UI Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 拉 TTL 配置 | `getTTLApi` | `/api/config/getTTL` | `ConfigTTL` 使用。 |
| 按类别更新 TTL | `setTTLApi` | `/api/config/setTTL` | `CollapsePanelHeader` 使用。 |
| 更新单表 TTL | `setSingleTableTTLApi` | `/api/config/setSingleTableTTL` | `TTLTable` 使用。 |
| 拉当前 license | `getLicense` | `/api/license/current` | `LicenseContext`、`LicenseInfoCard` 使用。 |
| 上传 license | `uploadLicense` | `/api/license/upload` | `UploadModal` 使用。 |
| 拉 deploy mode | `getDeployModeApi` | `/api/app/deployMode` | `DeployModeContext` 使用；后端当前直接返回 `PRODUCT_MODE` 环境变量。 |

## User Interaction Flow

1. 用户登录后，`LicenseContext` 检测到 `user.userId` 存在，调用 `/api/license/current`。
2. `DefaultLayout` 根据 `licenseInfo.active` 决定是否展示 warning/error banner。
3. 用户进入 `/system/platform-manage`，页面当前只显示 License management 面板。
4. 如果 `licenseInfo.active === 1`，前端显示 `LicenseInfoCard`；否则显示 `UnlicensedCard`。
5. 用户打开 `UploadModal`，粘贴 license JSON 文本或上传 `.lic` 文件。
6. `UploadModal` 解析 JSON，抽出 `license`、`encryptPublicKey`、`privateKey/decryptPrivateKey`、`signPublicKey`。
7. 前端把整理后的 JSON 字符串放入 `output` 字段，调用 `/api/license/upload`。
8. 成功后刷新 `LicenseContext`。
9. `/config` 页面则单独通过 `/api/config/getTTL` 和两个写接口维护 TTL。

## Permissions

| Area | Evidence | Notes |
|---|---|---|
| License Fetch | `LicenseContext.tsx` | 只有 `user.userId` 存在时才拉取。 |
| Backend Middleware | `router_api.go` | `/api/config` 需 `AuthMiddleware()+LicMiddleware()`；`/api/license` 需 `AuthMiddleware()`；`/api/app/upload` 叠加 `LicMiddleware()`。 |
| Layout Visibility | `DefaultLayout.js` | 告警 banner 是全局显示，不受单页可见性限制。 |

## Notes

- `PlatformManage` 中 Logo and Title Customization 相关 UI 已注释，不能写成当前可用功能。
- `DeployModeContext` 已有前后端闭环，但后端只是返回 `PRODUCT_MODE`，并没有提供可视化配置入口。
- `UploadModal` 的上传格式是 JSON 文本再包一层 `output`，不是直接把 `.lic` 二进制传给后端。
