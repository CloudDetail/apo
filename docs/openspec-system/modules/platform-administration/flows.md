# Platform Administration Flows

## Main Flow 1

### TTL Configuration Flow

1. 用户进入 `/config`。
2. `ConfigTTL` 调 `/api/config/getTTL` 读取各类别表当前 TTL。
3. 前端把结果分成 `logs`、`trace`、`k8s`、`topology`、`other` 五类展示。
4. 用户在某一类头部修改天数时，前端调 `/api/config/setTTL`。
5. 用户在具体表行上修改天数时，前端调 `/api/config/setSingleTableTTL`。
6. 后端解析表的 TTL 表达式并更新 `toIntervalDay(...)`。
7. 前端刷新 TTL 数据。

## Main Flow 2

### License Upload Flow

1. 用户进入 `/system/platform-manage`。
2. `LicenseContext` 已提前通过 `/api/license/current` 拉取当前状态。
3. 用户打开 `UploadModal`。
4. 前端把 license JSON 文本解析为 `{ license, encryptPublicKey, privateKey/decryptPrivateKey, signPublicKey }`。
5. 前端将该对象再 `JSON.stringify` 后放入 `output` 字段，调用 `/api/license/upload`。
6. 后端解析 `output`，用提供的公私钥激活 license。
7. 后端校验 license 是否过期，并在首次上传时创建 `activation` 记录。
8. 后端把当前激活记录全部置为 `activated=false`，再写入新记录。
9. 前端刷新 `LicenseContext`，页面切换到最新状态。

## Main Flow 3

### Global License Banner Flow

1. 用户登录后，`LicenseContext` 在 `user.userId` 存在时调 `/api/license/current`。
2. `DefaultLayout` 读取 `licenseInfo.active`。
3. 若 `active === 1`，不显示 banner。
4. 若 `active === 0`、`2` 或 `-1`，显示 warning/error banner。
5. 用户点击 banner 上按钮会跳到 `/system/platform-manage`。

## Main Flow 4

### Logo Upload Flow

1. 管理员或其他调用方向 `/api/app/upload` 提交 `logoText` 和 `logoImage`。
2. 后端检查文件大小和 MIME 类型。
3. 如果识别到 SVG，则执行 `sanitizeSVG`。
4. 后端把 `LogoImage` 与 `LogoText` 写入 DB。
5. 读取 logo 时通过 `/api/app/logo` 返回最新记录。

## Edge Cases

| Case | Implemented Behavior |
|---|---|
| 当前无激活 license | `/api/license/current` 返回 `Active: licenseStatusNotActivated`。 |
| `license` 已过期 | 上传时会被拒绝。 |
| 非法 TTL 类别 | 后端直接报错。 |
| 非法 logo MIME | 后端直接拒绝。 |
| 大于 2MB 的 logo | 后端直接拒绝。 |
| `PlatformManage` 搜索框 | 当前整段注释，不参与主流程。 |

## Failure Paths

| Path | Failure Behavior |
|---|---|
| `/api/app/deployMode` 调用失败或返回空值 | 前端 `DeployModeContext` 回退为默认 `apo`。 |
| license JSON 格式错误 | `UploadModal` 前端提示 `invalidJson`。 |
| license 关键字段缺失 | `UploadModal` 前端提示 `invalidLicenseFormat`。 |
| license 激活失败 | 后端返回 `UploadLicenseError`。 |
| ClickHouse TTL 修改失败 | 前端刷新后仍保持旧值。 |

## Permission Constraints

| Layer | Constraint |
|---|---|
| `/api/config` | 需要 `AuthMiddleware()` 和 `LicMiddleware()`。 |
| `/api/license` | 需要 `AuthMiddleware()`。 |
| `/api/app/upload` | 需要 `AuthMiddleware()` 且额外叠 `LicMiddleware()`。 |
| Global Layout | 只要登录并获取到 licenseInfo，就会显示全局告警。 |

## To Be Confirmed

- `/api/app/deployMode` 当前只暴露 `PRODUCT_MODE`，是否需要与 `activation.mode` 或管理员配置联动，需要确认。
- License expired 场景下前端究竟应显示 `LicenseInfoCard` 还是 `UnlicensedCard`，当前代码存在重叠意图，需要确认。
- Logo and Title Customization UI 是暂时下线还是尚未交付，需要确认。
