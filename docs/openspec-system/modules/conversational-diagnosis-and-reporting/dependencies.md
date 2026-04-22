# Conversational Diagnosis And Reporting Dependencies

## Upstream Dependencies

| Dependency | Type | Why It Matters |
|---|---|---|
| `timeRangeReducer` | Frontend shared state | 决定 ChatAI、CauseReport、Report 的查询窗口。 |
| `serviceAnalyze` / `alertAnalyze` | Cross-module page | 会通过链接把用户导入本模块的报告和对话页面。 |
| `UserContext` | Frontend shared context | 决定 `agent` 等入口的菜单可见性。 |
| `LicenseContext` | Cross-cutting | license 状态会在布局层给出告警提示。 |

## Downstream Dependencies

| Dependency | Type | Why It Matters |
|---|---|---|
| `/api/custom/*` | Backend internal | ChatAI 自定义数据视图配置直接依赖。 |
| `/api/alerts/events/report/get` | Backend internal | `/report` 页面直接依赖。 |
| `/api/dataplane/data/check` | Cross-module backend | Agent 页前置能力检查。 |
| `/ai/api/*` | External backend | ChatAI/Agent 主能力承载方。 |
| `/originx/api/*` | External backend | Cause report 数据承载方。 |

## Shared Components

| Component | File | Use |
|---|---|---|
| `BasicCard` | `src/core/components/Card/BasicCard` | ChatAI、Agent、平台页通用容器。 |
| `LoadingSpinner` | `src/core/components/Spinner` | ChatAI、CauseReport、Report 通用 loading。 |
| `notify` | `src/core/utils/notify` | 保存模板、拉取 license、登录成功等提示。 |

## Shared Utilities

| Utility | File | Use |
|---|---|---|
| `request.js` | `frontend/src/core/utils/request.js` | HTTP 拦截、token 刷新、语言头。 |
| `streamFetch` | `frontend/src/pro/util/streamFetch` | ChatAI 相关 SSE 请求。 |
| `getStep` | `src/core/utils/step` | 根据时间范围计算查询步长。 |

## External Services

| External Service | Evidence | Role |
|---|---|---|
| AgentSRE / `/ai/api/v1/srechat` | `AGUIProvider.initialConfig.baseUrl` | `/agent` 主对话入口。 |
| AI chat endpoints | `ChatAICom` / `ChatAIContext` | `/ai/api/alertchat`、`/ai/api/chat`、`/ai/api/nodeinfo`。 |
| Model setting service | `modelSetting.ts` | `/ai/api/v1/model-setting/*`。 |
| Conversation history service | `conversation.ts` | `/ai/api/v1/conversations/*`。 |
| `originx` | `cause.ts` | 报告拓扑、日志和流程数据。 |

