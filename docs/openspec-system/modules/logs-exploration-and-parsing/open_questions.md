# Logs Exploration And Parsing Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `To Be Confirmed`: `/logs/full` 是否有意作为“全局日志库”使用，而不是像 `/logs/fault-site` 一样受 `groupId` 约束。
2. `To Be Confirmed`: 解析规则中的 `serviceName` 字段到底是权限归属、路由辅助信息，还是仅供自动补充 `routeRule` 使用。
3. `To Be Confirmed`: `isStructured` 的真实业务语义是否表示“日志源已经是结构化 JSON”，还是仅表示“无需 regex 解析”。

## Legacy Or Dead-Code Suspicion

1. `To Be Confirmed`: `frontend/src/core/contexts/LogsContext.js` 中仍保留被注释掉的 `getLogRuleApi` 调用痕迹，说明页面状态流曾有旧版本实现。
2. `To Be Confirmed`: `frontend/src/oss/views/logs/component/FullLogs/component/SerarchBar/RawLogQuery/index.jsx` 里 `FullTextSearch` 与刷新按钮被注释，查询交互可能经历过未完全清理的演化。
3. `To Be Confirmed`: `frontend/src/core/components/AuthRouter.tsx` 中路由权限校验 `useEffect` 被整体注释，当前守卫组件更像路由包裹层而不是实际控制层。

## Missing Frontend-Backend Alignment

1. `To Be Confirmed`: `GetFaultLogContentRequest.ServiceName` 在后端 handler 中明确标注 unused，因此前端即使传了 service 也不会影响权限判断。
2. `To Be Confirmed`: `ConfigTableModal` 前端只收集 `dataBase`、`tableName`、`timeField`，但后端 `AddOtherTableRequest` 还支持 `cluster`、`instance`、`logField`，当前 UI 无法配置这些字段。
3. `To Be Confirmed`: `ConfigTableModal` 的表单项名是 `dataBase`，但保存时读取的是 `treeSelect`，实现上存在状态命名不一致，虽然可能可运行，但语义较脆弱。
4. `To Be Confirmed`: 上下文查询请求结构 `LogQueryContextRequest` 只接受 `tableName`、`dataBase`、`tags`、`timestamp`，没有暴露 `timeField` / `logField`；前端通过隐藏外部表的上下文按钮规避了这一点，但能力边界需要文档显式说明。

## Permission Uncertainty

1. `To Be Confirmed`: 当前只在 `POST /api/log/fault/pagelist` 观察到 `CheckGroupPermission`，其余日志查询与配置接口没有看到等价的数据范围校验。
2. `To Be Confirmed`: `POST /api/log/fault/content` 存在明确 TODO，说明当前故障现场内容接口缺少 service 级数据权限控制。
3. `To Be Confirmed`: `/logs/full` 页面虽然经过 `AuthRouter` 包装，但由于客户端校验逻辑被注释，实际权限边界更依赖后端路由中间件与菜单配置。
4. `To Be Confirmed`: 解析规则管理与外部表管理是否应该具备更细粒度的角色权限，当前代码未见按钮级或接口级专门控制。

## Data Model Uncertainty

1. `To Be Confirmed`: `GetServiceRoute` 通过最近 7 天活跃实例名拆分 pod 前缀来生成 `k8s.pod.name` 规则，这一推断高度依赖命名规范。
2. `To Be Confirmed`: `handler.New(...)` 启动时执行 `InitParseLogTable`，默认 `routeRule` 固定匹配 `apo`，这是否只适用于默认安装环境仍不明确。
3. `To Be Confirmed`: 解析规则新增、更新、删除跨 ClickHouse、Vector ConfigMap、数据库执行，但当前未见事务式补偿；若中途失败，系统可能进入部分成功状态。
4. `To Be Confirmed`: `service_get_log_chart.go` 在尾部补齐直方图桶时对 `To` 的赋值复用了 `firstFrom` 相关表达式，实现语义存在可疑点，需要后续代码核验。
