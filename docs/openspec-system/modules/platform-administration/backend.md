# Platform Administration Backend

## API Summary

| Category | Method | API Path | Handler | Notes |
|---|---|---|---|---|
| Config | GET | `/api/config/getTTL` | `GetTTL` | 返回 logs/trace/k8s/topology/other 五类 TTL 信息。 |
| Config | POST | `/api/config/setTTL` | `SetTTL` | 按类别更新 TTL。 |
| Config | POST | `/api/config/setSingleTableTTL` | `SetSingleTableTTL` | 按单表更新 TTL。 |
| License | POST | `/api/license/upload` | `UploadLicense` | 上传并激活 license。 |
| License | GET | `/api/license` | `GetLicenses` | 获取全部 license 记录。 |
| License | GET | `/api/license/current` | `GetCurrentLicense` | 获取当前激活 license。 |
| App | GET | `/api/app/deployMode` | `GetDeployMode` | 直接返回 `PRODUCT_MODE` 环境变量。 |
| App | POST | `/api/app/upload` | `UploadLogo` | 上传 logo。 |
| App | GET | `/api/app/logo` | `GetLogo` | 获取当前 logo。 |

## Controllers Or Routers

| File | Responsibility |
|---|---|
| `backend/pkg/router/router_api.go` | 注册 `/api/config`, `/api/license`, `/api/app`。 |
| `backend/pkg/api/config/handler.go` | TTL 配置接口。 |
| `backend/pkg/api/license/handler.go` | license 上传与查询接口。 |
| `backend/pkg/api/app/handler.go` | app 模块接口定义与注册，包括 logo 与 deploy mode。 |

## Services

| Service | File | Behavior |
|---|---|---|
| `GetTTL` | `backend/pkg/services/config/service_getttl.go` | 从 ClickHouse `CREATE TABLE` 语句中解析 TTL，并按类别分组。 |
| `SetTTL` | `backend/pkg/services/config/service_setttl.go` | 按 `dataType` 找到对应表集合，统一更新 TTL interval。 |
| `SetSingleTableTTL` | `service_setttl.go` | 按表名更新 TTL。 |
| `UploadLicense` | `backend/pkg/services/license/service_upload_license.go` | 解析 `output` JSON，激活 license，校验未过期后写 DB。 |
| `GetLicenses` | `backend/pkg/services/license/service_get_license.go` | 读取全部 license，并结合 Prometheus 节点数生成 seat 使用量。 |
| `GetCurrentLicense` | `service_get_license.go` | 返回当前激活 license；无记录时返回 `Active: licenseStatusNotActivated`。 |
| `CheckLicenseNodes` / `CheckLicenseDate` | `service_license_check.go` | 后台定时检查 license 节点数和到期状态。 |
| `UploadLogo` | `backend/pkg/services/app/logo.go` | 校验大小、MIME、SVG 安全后写入 DB。 |
| `GetLogo` | `backend/pkg/services/app/logo.go` | 返回当前 logo。 |

## Models And Schemas

| Model | File | Mapping |
|---|---|---|
| `License` | `backend/pkg/model/license/license.go` | `TableName() == "license"`。 |
| `Activation` | `backend/pkg/model/license/license.go` | `TableName() == "activation"`。 |
| `UploadLicenseRequest` | `backend/pkg/model/request/license_req.go` | 只有 `Output` 字段。 |
| `SetTTLRequest` / `SetSingleTTLRequest` | `backend/pkg/model/request/config_req.go` | TTL 更新请求。 |
| `UploadLogoRequest` | `backend/pkg/model/request/app_req.go` | logo 文本和文件上传。 |
| `LicenseResp` | `backend/pkg/model/response/license_resp.go` | license 查询结果。 |
| `GetLogoResp` | `backend/pkg/model/response/app_resp.go` | logo 查询返回。 |

## Database Mapping

| Storage | Evidence | Notes |
|---|---|---|
| `license` | `model/license/license.go` | 激活后的 license 记录。 |
| `activation` | `model/license/license.go` | 部署模式激活记录。 |
| Logo table | `backend/pkg/repository/database/app/logo.go` | 具体表名由 `model.Logo` 决定，当前 app repo 负责单条 logo 资源。 |
| ClickHouse tables | `service_getttl.go`, `service_setttl.go` | TTL 配置直接作用于 ClickHouse 表。 |

## Validation And Rules

| Area | Implemented Rule |
|---|---|
| TTL Category | `SetTTLRequest.DataType` 只允许 `logs trace k8s other topology`。 |
| TTL Day | `SetTTL` 与 `SetSingleTableTTL` 都要求 `day > 0`。 |
| Table Name | 单表 TTL 更新要求 `model.IsTableExists(req.Name)`。 |
| License Upload | `UploadLicense` 先解析 `output`，再通过 RSA key 激活并校验未过期。 |
| Current License | `CreateLicense` 写新 license 前会先把已激活记录全部置为 `activated=false`。 |
| Logo Size | logo 文件大小不能超过 2MB。 |
| Logo Type | 只允许 jpeg/png/svg+xml/x-icon/xml。 |
| SVG Safety | SVG 上传时会走 `sanitizeSVG`。 |
| Logo Text | `UploadLogoRequest.LogoText` 最长 15 个字符。 |

## Error Handling

| Area | Behavior |
|---|---|
| TTL Parse Failure | 读取表信息或修改 TTL 失败会直接返回 error。 |
| Invalid TTL Type | 未知 `dataType` 会返回 `type: <x> does not have tables`。 |
| License Activate Failure | 返回 `ActivateLicenseFailed`。 |
| License Expired | 返回 `LicenseExpired`。 |
| Invalid Logo Type | 返回 `LogoTypeError`。 |
| Logo Oversize | 返回 `GetLogoError: logo file size exceeds 2M`。 |

## Notes

- `/api/app/deployMode` 已注册，但它不经过独立 service 或数据库，只直接返回 `PRODUCT_MODE`。
- `GetCurrentLicense` 会把最近一天内 Prometheus 节点数作为 `seatsCurrentlyInUse`。
- `UploadLicenseRequest` 的字段名是 `output`，这解释了前端为什么先拼装 JSON 再传字符串。
