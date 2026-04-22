# Conversational Diagnosis And Reporting Open Questions

## Business Semantics That Cannot Be Fully Inferred From Code

1. `agentChatId` 在 `/report` 页面中的完整来源链路未在当前仓库内闭环。
2. `reportType` 的全部可选值和语义分支没有在当前仓库中集中声明。

## Legacy Or Dead-Code Suspicion

1. `frontend/src/pro/api/chatAI.ts` 中 `/api/chat`、`/api/alertchat` wrapper 当前未见主页面使用。
2. `PreUseInspection.tsx` 仍引用 `/#/agent-datasource`，像是旧数据平面入口残留。
3. `AGUIChat` 中 `ENABLE_SESSION_MANAGEMENT = false`，本地 session 管理逻辑当前整体处于关闭状态。

## Missing Frontend-Backend Alignment

1. 前端大量调用 `/ai/api/*`、`/originx/api/*`，但主 Go 后端未提供同名实现。
2. `frontend/src/core/api/app.js` 调用 `/api/app/deployMode`，但当前后端 router 只暴露 `/api/app/upload` 与 `/api/app/logo`。
3. `report.ts` 使用的路径写法是 `api/alerts/events/report/get`，缺少前导 `/`；虽然 axios `baseURL` 为空时通常可工作，但仍需确认是否存在环境差异。

## Permission Uncertainty

1. `AuthRouter` 的实际路由鉴权逻辑被注释，当前前端是否只依赖后端菜单过滤需要确认。
2. `/api/custom/*` 当前 router 只见 `LicMiddleware()`，用户身份校验是否在更上层或旁路完成需要确认。

## Data Model Uncertainty

1. `custom_query_service.querytype` 存储的是逗号拼接字符串，是否允许重复值、是否存在排序要求，代码未显式说明。
2. `alert_report` 表按标签和时间窗口取最近一条，但“最近”的精确业务定义仅体现为 `LIMIT 1`，未看到额外排序字段说明。

