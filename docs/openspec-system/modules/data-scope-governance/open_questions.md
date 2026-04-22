# Data Scope Governance Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. 旧版 `/api/data/group*` 与 V2 `/api/v2/data/group*` 的最终替换计划在当前仓库中看不完整。
2. `protected` scope 的业务含义是“当前用户无权删除”还是“系统保护项”，需要进一步确认。

## Legacy Or Dead-Code Suspicion

1. 前端旧版 `creatDataGroupApi`、`updateDataGroupApi`、`deleteDataGroupApi` wrapper 仍在，但主页面已经不使用。
2. `DataGroupContext` 中出现 `know`/`known` 拼写差异，可能是历史残留。
3. `DataGroupStorage.Refresh` 中部分 ClickHouse 扫描逻辑被注释，说明 scope 来源路径经历过调整。

## Missing Frontend-Backend Alignment

1. 前端 V2 删除使用 `DELETE /api/v2/data/group/delete`，而 handler 注释和旧版接口仍让人容易误读为 POST 风格。
2. 主页面已完全使用 V2，而旧版接口仍保留在 API wrapper 与后端 handler 中，文档和调用者容易混淆。

## Permission Uncertainty

1. `DataGroupContext` 禁用节点逻辑来自 `permissionType`，但前端 route 层权限检查当前未真正执行。
2. `CleanExpiredDataScope` 的删除资格取决于当前用户可见 group 集合，和管理员/普通用户体验可能不同。

## Data Model Uncertainty

1. `APO_ALL_DATA` 与 `group_id = 0` 的 root group 是系统保底机制，但实际业务是否允许修改 root group 语义，代码中没有前端入口说明。
2. `DataScope` 的完整字段及其生成来源需要在后续 deeper round 联读 `dao_data_scope.go` 和观测 repository 才能完全说明。
