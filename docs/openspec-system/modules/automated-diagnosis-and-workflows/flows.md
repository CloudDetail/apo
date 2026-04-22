# Automated Diagnosis And Workflows Flows

## Main Flow 1

### Dify Workflows Console Flow

1. 用户进入 `/workflows`。
2. 页面渲染 `WorkflowsIframe src="/dify/apps"`。
3. 组件读取本地 `difyToken` 和 `difyRefreshToken`。
4. 如果当前用户是 `anonymous` 且本地没有 token，前端调用 `/dify/console/api/login/apo/anonymous`。
5. 登录成功后前端把 `access_token`、`refresh_token` 写入 localStorage。
6. iframe URL 会自动拼接 `access_token` 和 `refresh_token`。
7. 当 iframe 通过 `postMessage` 发送 `refresh_token` 时，父页面会更新本地 token。

## Main Flow 2

### Alert Analyze Flow

1. 用户进入 `/alert-analyze`。
2. `AlertAnalyzeContext.fetchData()` 调用 `getServiceAlertEventsApi` 拉取分析候选告警事件。
3. 页面左侧 `AlertEventTable` 展示告警事件、关联服务和影响入口。
4. 用户点击一键分析，context 调用 `/api/alerts/event/impact` 获取受影响入口列表。
5. `ImpactList` 渲染入口列表，并为当前页入口获取 latency / errorRate / tps 图。
6. 用户可从入口列表继续查看故障贡献度、异常列表、打开 `/alert-analyze/info` 或跳转 `/alert-analyze/chatAI/:service/:endpoint`。

## Main Flow 3

### Descendant Service Analyze Flow

1. 用户通过 `#/alert-analyze/info?service=...&endpoint=...` 打开下游异常分析页面。
2. `ServiceAnalyzeContext` 从 query 读取 `service`、`endpoint`、`clusterIds`。
3. context 调用 `getServiceRelationApi` 获取拓扑关系。
4. context 调用 `/api/alerts/descendant/anormal/trend` 获取异常趋势。
5. context 调用 `/api/alerts/descendant/anormal/delta` 获取 `origin`、`alertTriggered`、`final` 异常计数与事件记录。
6. 页面展示拓扑图、趋势图、异常类型过滤和时间片播放条。
7. 用户点击节点后，右侧 `NodeInfo` 基于当前节点展示异常详情与故障报告入口。

## Main Flow 4

### Mutation Detection Flow

1. 用户在 `/alert-analyze` 页面打开 `AlertRuleQuickSet`。
2. 前端调用 `/api/alerts/detect/mutation/metrics` 获取预定义指标模板。
3. 用户填写 `name`、`for`、检测表达式、上下阈值，并可勾选 `synchronizeToAlertRules`。
4. 前端把表单值转换为 `DetectMutationRequest`，补充 `startTime`、`endTime`、`step`。
5. 后端 `DetectDefects` 执行 Prometheus mutation check。
6. 后端把生成的突变告警事件批量写入 ClickHouse。
7. 后端把执行记录写入 detect mutation 记录表。
8. 如果勾选同步规则，后端额外调用 `k8sRepo.AddAlertRule` 创建告警规则。
9. 用户可在 `AnomalyDetectionRecord` 弹窗中查看执行记录和规则记录。

## Main Flow 5

### Async Alert Check And Auto Analyze Flow

1. 其他告警模块把事件提交给 `difyRepo.SubmitAlertEvents(events)`。
2. `async_alert_check` worker 基于 `alertGroup + alertName` 获取或缓存 `AlertEventClassify`。
3. worker 执行 alert check workflow，写出 `WorkflowRecord`。
4. 如果 `AutoAnalyze == true` 且 alert check 输出为 `"false"`，worker 会进一步构造 `WorkflowParams` 并触发 analyze workflow。
5. analyze workflow 的结果写回 `WorkflowRecord.AlertDirection`、`AnalyzeRunID`、`AnalyzeErr`。

## Edge Cases

| Case | Implemented Behavior |
|---|---|
| 未传分页参数 | 多个接口会默认注入 `PageParam`。 |
| `sortBy` 为空 | 贡献度排序默认使用 `latency,errorRate`。 |
| `MutationCheck.Group` 为空 | 突变检测默认使用 `MutationCustomLabelVal`。 |
| 选中的异常类型不含 `middleware` | 下游异常查询不会额外查数据库实例。 |
| alert check 过期 | `workflow_worker.go` 会生成 `Output = "failed: alert check too late..."` 的记录。 |
| Dify workflow API 调用失败 | `WorkflowRecord.Output` 或 `AnalyzeErr` 记录失败原因。 |
| 无法从告警 tag 推断服务 | `EndpointsMap` 会继续尝试其他 tag 组合。 |

## Failure Paths

| Path | Failure Behavior |
|---|---|
| Dify 匿名登录失败 | `WorkflowsIframe` 显示 warning `Result`。 |
| 影响面或下游异常查询失败 | 前端 context 会清空部分状态并停止 loading。 |
| 突变检测执行失败 | 后端返回错误，前端不会提示保存成功。 |
| 同步告警规则失败 | `DetectDefects` 返回 `DetectDefectsCreatAlertError`。 |
| 点击异常列表 wrapper 路径不一致 | `getAlertDescendantApi` 当前可能访问不存在的 `/api/alerts/descendant/anormal` 路径。 |

## Permission Constraints

| Layer | Constraint |
|---|---|
| Frontend Route | `/workflows`、`/alert-analyze`、`/alert-analyze/info` 被 `AuthRouter` 包装。 |
| Backend API | `alertanalyze` routes 在 `/api/alerts` group 的 `AuthMiddleware()` 之后。 |
| Data Scope | 大多数分析请求还会叠加 `groupId` / `clusterIds` 限制。 |

## To Be Confirmed

- To Be Confirmed: `getAlertDescendantApi` 的路径漂移是遗留 wrapper 还是路由兼容缺失。
- To Be Confirmed: 非匿名用户的 Dify token 刷新流程在当前前端如何保证。
- To Be Confirmed: `repository/dify` 中 alert check / analyze 记录最终由哪个仓储写入 ClickHouse，需要结合其他模块继续核对。
