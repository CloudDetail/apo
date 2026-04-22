# Platform Administration Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `/api/app/deployMode` 当前直接返回 `PRODUCT_MODE`；它与 `activation.mode`、license 或管理员配置之间是否应该存在更稳定的映射关系，代码中仍不清晰。
2. branding 配置最终是否应允许管理员修改标题、favicon 等，在现有代码中只看到了部分注释掉的 UI。

## Legacy Or Dead-Code Suspicion

1. `PlatformManage` 中 Logo and Title Customization 相关区块被整段注释。
2. `UnlicensedCard` 中 trial 申请相关 UI 被整段注释。
3. `UploadModal` 的文件读取逻辑有重复 `FileReader` 痕迹，可能是历史实现残留。

## Missing Frontend-Backend Alignment

1. 前端 `DeployModeContext` 已能调到 `/api/app/deployMode`，但后端只返回环境变量，缺少更完整的配置模型。
2. 前端对 expired license 的展示意图与 `LicenseInfoCard` / `UnlicensedCard` 的实际分支存在部分重叠。

## Permission Uncertainty

1. `/api/license` 只要求登录，但上传 license 是否还需要更细粒度管理员权限，当前代码中未体现。
2. `/api/app/logo` 读取接口没有叠 `LicMiddleware()`，这是否是预期的公开读取边界，需要确认。

## Data Model Uncertainty

1. logo 表的具体结构在 `model.Logo` 中，当前模块文档只能确认 repository/app 会以单条最新记录方式处理。
2. `activation.mode` 与前端 `deployMode` 是否一一对应，当前仓库中没有直接映射证据。
