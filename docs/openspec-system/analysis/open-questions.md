# Open Questions

## Scope
本文档记录当前第一阶段扫描后仍未完全闭合的问题、证据、影响和后续建议。这里不做猜测性结论，只记录当前仓库代码无法完全解释的点。

## Decomposition Notes
在当前轮次中，模块拆分已进一步收敛为以下原则：

- 将“告警治理”和“告警来源接入”分开。
- 将“自动化诊断/workflows”和“会话式诊断/报告”分开。
- 将 `one-agent` 与接入链路并入同一模块，而不是单独保持一个仅有边界说明的碎片模块。

下面的开放问题会继续影响这套分解是否还需要微调。

## High Priority Questions
| ID | Topic | Observed Evidence | Impact | Recommended Follow-Up |
|---|---|---|---|---|
| OQ-01 | Frontend / backend route drift | 前端主路由与 `init_router.go` 现已对齐到 `/dataplane`、`/probe-management`、`/agent`，但 `service_check_router.go` 白名单仍保留 `/agent-dataplane`、`probe-installer` | 会影响权限模型、菜单可见性与文档链路准确性 | 核对现网权限校验到底依赖数据库 router 结果、白名单，还是两者并存。 |
| OQ-02 | Frontend wrapper drift against backend registration | 前端仍存在 `/api/permission/roles` 调用包装，但当前主后端注册的是 `/api/role/roles` | 会导致“页面实际依赖接口”与“仓库可见后端接口”不一致 | 继续核对该 wrapper 是否仍被实际调用，或应统一收敛到 `/api/role/roles`。 |
| OQ-03 | External AI boundaries are not fully implemented inside this repo | 前端直接使用 `/ai/api`、`/originx/api`、`/dify/console/api`，当前主后端中未见完整对应 route group | AI 模块文档如果处理不当，会把外部服务误写成本仓库实现 | 在 AI 模块文档中明确拆分“本仓库实现”和“外部依赖能力”。 |
| OQ-04 | Workflows frontend path and Dify integration boundary | 前端 `core/api/workflows.ts` 使用 `/dify/console/api/*`；后端可见的是 `repository/dify` 用于 alert check / analyze | `workflows` 业务链可能跨越后台服务与外部 Dify console | 单独建立 workflows 边界说明，避免与 `backend/pkg/repository/dify` 混写。 |
| OQ-05 | OneAgent implementation visibility is incomplete | `one-agent/README.md` 描述完整采集体系，但当前仓库主要靠 `.gitmodules` 引用 `node-agent`、`preload`、`ilogtail`、`odigos` | 无法仅凭当前 checkout 完整解释采集子系统内部实现 | 确认 submodule 是否已拉取，并决定文档是只写边界还是纳入子仓源码分析。 |

## Medium Priority Questions
| ID | Topic | Observed Evidence | Impact | Recommended Follow-Up |
|---|---|---|---|---|
| OQ-06 | V1 and V2 DataGroup APIs coexist | 前端 `core/api/dataGroup.ts` 同时调用 `/api/data/*` 与 `/api/v2/data/*`；后端主注册更偏向 V2 | 如果不说明，会造成数据组文档内接口混乱 | 单独梳理 DataGroup API 演进关系，标出仍在前端调用中的旧路径。 |
| OQ-07 | AuthRouter permission check is currently disabled | `frontend/src/core/components/AuthRouter.tsx` 中实际权限检查逻辑被注释 | 页面真实访问行为可能与理论权限模型不一致 | 需要区分“理论权限链路”和“当前前端实际执行逻辑”。 |
| OQ-08 | Edition boundary is not fully clean | `frontend/src/routes.js` 在非 `pro` 分支下仍拼接 `proRoutes`；`AppSidebar` 又受 `VITE_APP_CODE_VERSION`、license、菜单返回共同影响 | 目录名、环境变量、license、后端菜单模型可能共同决定实际暴露能力 | 补一份 edition / license / menu 的综合说明。 |
| OQ-09 | Deploy mode behavior is only partially visible | 前端 `DeployModeContext` 通过 `/api/app/deployMode` 切换 APO / Syncause branding；后端当前直接返回 `PRODUCT_MODE` | branding 文档虽然已有前后端闭环，但 deploy mode 与 `activation.mode`、管理员配置之间的关系仍不清晰 | 继续确认 deploy mode 是否只应视为环境注入，还是未来需要更完整的持久化/配置模型。 |
| OQ-10 | Alert rule API path split | 前端 `core/api/alerts.js` 同时存在 `/alert/api/v1/rules` 与 `/api/alerts/*` | 说明告警规则能力可能跨越两个后端边界 | 后续告警文档中把“仓库后端能力”和“外部 alert API”分开。 |
| OQ-14 | Shared frontend paths have already diverged from upstream repo | `frontend/src/core` 已有 20 个同路径变体和 15 个商业独有文件，`frontend/src/oss` 也已出现差异 | 如果文档不标注源码归属，后续很容易把改动落在错误仓库，或把共享路径误写成纯 Pro 功能 | 对 `core` / `oss` 相关文档统一补 `Source Ownership` 说明，并在改动前固定做 upstream 对比。 |
| OQ-15 | Backend differentiation is package-based, not directory-based | 后端没有 `pro/` 类似目录，但 `alertanalyze`、`license`、`app`、`custom` 等包只存在于商业仓，同时大量共享包又被改写 | 如果文档仍按目录粗粒度判断，会把后端共享包上的商业扩展遗漏掉 | 以 `router_api.go`、`pkg/api`、`pkg/services`、`pkg/model`、`pkg/repository` 的包级差异建立后端归属规则。 |

## Low Priority But Important Questions
| ID | Topic | Observed Evidence | Impact | Recommended Follow-Up |
|---|---|---|---|---|
| OQ-11 | Database-initialized router names may not fully cover frontend route file reality | `backend/pkg/repository/database/init_router.go` 的 `validRouters` 列表明显比前端路由更保守，且存在命名不一致 | 会影响权限、插页文案、菜单映射说明 | 后续可以导出现网 `router` 表或本地初始化结果比对。 |
| OQ-12 | Some features are implied through static templates and docs rather than direct runtime code | `backend/static/integration-tmpl/`、`frontend` 中接入说明文本较多 | 容易把“安装资产”与“运行时代码”混为一谈 | 文档中应区分“运行时业务能力”和“交付/安装资产”。 |
| OQ-13 | Network and dashboard capability boundaries are still fuzzy | `dashboard` 页面、`metric` API、`network` API、`service health` 页面之间存在能力交叠 | 如果拆分不谨慎，会出现模块重复说明 | 在写 dashboard 模块前先补 `runtime-topology` 与 metrics 来源说明。 |

## Question Clusters

### Cluster A: Route, Menu, Permission
这一组问题必须优先处理，因为它们决定文档里“页面属于谁、谁能看到谁”的基础边界。

相关问题：

- `OQ-01`
- `OQ-06`
- `OQ-07`
- `OQ-08`
- `OQ-11`

### Cluster B: External Service Boundaries
这一组问题决定系统哪些能力属于本仓库实现，哪些只是被调用或被代理。

相关问题：

- `OQ-02`
- `OQ-03`
- `OQ-04`
- `OQ-10`

### Cluster C: Collector And Deployment Side
这一组问题影响接入、安装、采集链路文档的边界。

相关问题：

- `OQ-05`
- `OQ-12`

### Cluster D: Platform Runtime Semantics
这一组问题影响 branding、license、dashboard 分类等高层解释方式。

相关问题：

- `OQ-09`
- `OQ-13`

### Cluster E: Source Governance
这一组问题影响共享路径代码该在哪个仓库修改，以及文档如何标注上游基线、共享变体与商业独有实现。

相关问题：

- `OQ-14`
- `OQ-15`

## Guidance For Later Documentation
后续写模块文档时，遇到上述开放问题应遵守以下原则：

1. 如果某能力在前端有调用、但在当前主后端未观察到实现，必须标注为“外部边界待确认”。
2. 如果前端 route 与后端 router 元数据不一致，必须同时记录两者，不允许强行统一成单一说法。
3. 如果同一能力同时存在旧接口与新接口，必须在文档中明确“当前前端实际仍在调用哪一组”。
4. 如果 `one-agent` 具体实现不在当前 checkout 中，就只记录本仓库可见边界与安装衔接，不写内部行为细节。
5. 如果共享路径或共享包已经偏离上游，必须额外记录源码归属，不允许只写业务模块边界。

## Suggested Closure Order
建议按以下顺序逐步消化开放问题：

1. `OQ-01` 路由漂移
2. `OQ-02` 未直接观察到的内部接口
3. `OQ-03` 与 `OQ-04` 外部 AI / workflow 边界
4. `OQ-06` DataGroup V1/V2 并存
5. `OQ-05` OneAgent submodule 可见性
6. `OQ-08` edition / license 组合控制
7. `OQ-14` 前端共享路径与上游分叉治理
8. `OQ-15` 后端共享包与商业扩展的区分规则
9. 其余问题在相关模块文档中边写边收敛
