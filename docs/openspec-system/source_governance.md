# Source Governance

## Scope
本文档用于说明当前商业仓与开源仓之间的源码归属关系，以及文档应如何标注共享基线、商业变体与 Pro 独有能力。它同时覆盖前端与后端，但前后端的区分方式并不相同：

- 前端更接近“目录级分层 + 共享路径变体”。
- 后端更接近“共享包 + 商业扩展包 + 共享包上的商业改写”。

## Repositories

| Repository | Path | Role |
|---|---|---|
| Commercial Repo | `/Users/keting/workspace/product/commercial/apo` | 当前业务文档和商业实现所在仓库。 |
| Upstream OSS Repo | `/Users/keting/workspace/githubProject/apo` | 当前已确认的开源基线仓库。 |

## Classification Model

| Classification | Meaning | Documentation Implication |
|---|---|---|
| `upstream-common` | 两个仓库同路径文件存在且内容一致。 | 可以写成共享基线，但要注明源码主导权不在商业仓。 |
| `commercial-variant` | 两个仓库同路径文件都存在，但商业仓内容已经偏离开源仓。 | 文档必须写出“共享路径上的商业变体”，不能误写成纯 OSS 或纯 Pro。 |
| `commercial-only` | 只存在于商业仓。 | 可以作为 Pro 或商业特性直接归档。 |
| `upstream-edition-gated` | 开源仓自身已经带有 `CE/EE` 或版本分支。 | 不能把所有版本门控都误记为商业仓新增。 |

## Governance Rule

1. 任何位于共享路径或共享包中的代码，都不应在文档里默认写成“商业仓完全本地维护”。
2. 对 `commercial-variant` 文件，文档必须同时记录：
   - 业务模块归属
   - 共享路径/共享包归属
   - 与上游的偏离点
3. 对 `commercial-only` 文件，文档必须明确标记其属于 Pro 或商业增强，不要混入共享基线叙述。
4. 如果未来需要修改共享路径，应先对照上游仓，再决定是“先上游提交再回合并”，还是“只保留在商业仓的本地增强”。

## Frontend Ownership

### Frontend Governance Rule

1. `frontend/src/core` 与 `frontend/src/oss` 视为上游共享路径。
2. 这些路径如果要改代码，应优先在开源仓修改，再合并回当前仓。
3. `frontend/src/pro` 与 `frontend/public/locales/*/pro/*` 可视为商业仓主导的 Pro 前端区域。
4. 即使文件位于 `core` 或 `oss`，也可能已经是 `commercial-variant` 或 `commercial-only`，不能仅凭目录名判断。

### Frontend Comparison Snapshot
以下统计均排除了 `.DS_Store`：

| Area | Commercial Count | Upstream Count | Same Content | Changed Common Files | Commercial Only | Notes |
|---|---:|---:|---:|---:|---:|---|
| `frontend/src/core` | 337 | 322 | 302 | 20 | 15 | 共享路径中存在明显商业变体与新增商业文件。 |
| `frontend/src/oss` | 125 | 125 | 124 | 1 | 0 | 大部分仍与上游一致，但不能假设完全一致。 |
| `frontend/src/pro` | 322 | 0 | 0 | 0 | 322 | 当前前端 Pro 主体。 |
| `frontend/src` root comparable files | 7 | 7 | 0 | 7 | 0 | `App.js`、`index.js`、`routes.js`、`_nav.js` 等共享壳层文件均已偏离上游。 |

### Frontend Shared Path Variants

#### App Shell And Branding

| Source Area | Classification | Module / Concern | Observed Difference |
|---|---|---|---|
| `frontend/src/App.js` | `commercial-variant` | `frontend-app-shell`, `platform-administration` | 动态标题、favicon 与 deploy mode 逻辑。 |
| `frontend/src/index.js` | `commercial-variant` | `frontend-app-shell`, `license-and-edition-gating` | 注入 `DeployModeProvider` 与 `LicenseProvider`。 |
| `frontend/src/routes.js` | `commercial-variant` | `frontend-app-shell`, `data-scope-governance` | 引入 `proRoutes`、新增 `/system/data-group`、按版本拼接路由。 |
| `frontend/src/_nav.js` | `commercial-variant` | `frontend-app-shell` | 新增 MySQL 大盘、服务健康、Agent 等导航项。 |
| `frontend/src/core/layout/DefaultLayout.js` | `commercial-variant` | `platform-administration` | layout 级 logo、license banner、品牌切换。 |
| `frontend/src/core/components/AppHeader.js` | `commercial-variant` | `platform-administration`, `data-scope-governance` | 动态品牌标题，`DataGroupSelector` 改为隐藏式控制。 |
| `frontend/src/core/components/AppSidebar.js` | `commercial-variant` | `permission-menu-router-model` | 菜单裁剪叠加 `VITE_APP_CODE_VERSION === 'EE'` 逻辑。 |
| `frontend/src/core/views/Login/Login.jsx` | `commercial-variant` | `identity-access-and-organization`, `platform-administration` | 登录页品牌切换，并在 `syncause` 模式下跳过 `workflowLoginApi`。 |

#### Shared UI, Request, Time Range

| Source Area | Classification | Module / Concern | Observed Difference |
|---|---|---|---|
| `frontend/src/constants.js` | `commercial-variant` | `frontend-state-request-and-i18n` | 新增按 edition 切换的 `contentHeight`。 |
| `frontend/src/i18n.js` | `commercial-variant` | `frontend-state-request-and-i18n` | 新增 `core/dateTime` 与整组 `pro/*` namespace。 |
| `frontend/src/core/utils/request.js` | `commercial-variant` | `frontend-state-request-and-i18n` | `baseURL` 从固定地址改为空串。 |
| `frontend/src/core/store/reducers/timeRangeReducer.js` | `commercial-variant` | `data-group-and-time-range-context` | 修正 `thisWeek` 并新增 `last7Days`。 |
| `frontend/src/core/components/DateTime/*` | `commercial-variant` | `data-group-and-time-range-context` | 时间范围 UI 与默认值、只读能力增强。 |
| `frontend/src/core/components/HighlightCode/HighlightCode.tsx` | `commercial-variant` | `frontend-state-request-and-i18n` | 搜索占位符切换为 i18n 文本。 |
| `frontend/src/core/components/Chart/DelayLineChart.jsx` | `commercial-variant` | `dashboards-and-health-monitoring` | 增加 resize 兜底。 |
| `frontend/src/core/scss/theme/dark.scss`, `light.scss` | `commercial-variant` | `frontend-app-shell` | 新增容器背景 CSS 变量。 |

#### Shared Paths With Commercial Injection

| Source Area | Classification | Module | Observed Difference |
|---|---|---|---|
| `frontend/src/core/api/app.js` | `commercial-only` | `platform-administration` | 新增 `/api/app/deployMode` wrapper。 |
| `frontend/src/core/contexts/DeployModeContext.js` | `commercial-only` | `platform-administration` | 新增 APO / Syncause 品牌上下文。 |
| `frontend/src/core/api/appConfig.js` | `commercial-only` | `data-ingestion-and-oneagent` | 新增 `/ai/api/v1/app-config/*` 外部边界。 |
| `frontend/src/core/views/IntegrationCenter/DataIntegration/BotManagement.tsx` | `commercial-only` | `data-ingestion-and-oneagent` | Slack Bot 管理 UI。 |
| `frontend/src/core/components/Integration/*` | `commercial-only` | `data-ingestion-and-oneagent` | 更细粒度的采集接入表单组件。 |
| `frontend/src/core/views/IntegrationCenter/DataIntegration/ClusterTable.tsx` | `commercial-variant` | `data-ingestion-and-oneagent` | EE 模式下引入 `BotManagement`。 |
| `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/SettingsForm/*` | `commercial-variant` | `data-ingestion-and-oneagent` | Probe / minimal 流程增强。 |
| `frontend/src/core/views/IntegrationCenter/DataIntegration/Integration/InstallCmd/index.tsx` | `commercial-variant` | `data-ingestion-and-oneagent` | 商业文档链接与安装说明调整。 |
| `frontend/src/oss/views/dashboard/MiddlewareDashboard.jsx` | `commercial-variant` | `dashboards-and-health-monitoring` | 即使是 `oss` 路径，也已与上游不同。 |

### Frontend Upstream Edition Gating
以下文件在开源仓本身已经存在版本门控：

- `frontend/src/core/components/DataGroup/DataGroupSelector.tsx`
- `frontend/src/core/components/Card/BasicCard.tsx`
- `frontend/src/core/views/UserManage/components/UserTable.tsx`
- `frontend/src/core/views/Team/index.tsx`
- `frontend/src/oss/views/alertEvents/index.tsx`

因此，文档不能把所有 `VITE_APP_CODE_VERSION` 判断都写成商业仓新增逻辑。

### Frontend Module Summary

| Module | Frontend Ownership Summary |
|---|---|
| `service-observability` | 以 `upstream-common` 为主，但运行时依赖已商业分叉的共享壳层。 |
| `logs-exploration-and-parsing` | 以 `upstream-common` 为主，周边时间范围与请求层有本地分叉。 |
| `trace-exploration-and-profiling` | 以 `upstream-common` 为主，周边共享组件有分叉。 |
| `dashboards-and-health-monitoring` | `mixed`，同时包含共享 dashboard、`MiddlewareDashboard` 变体和 Pro 大盘页。 |
| `alert-rules-and-incident-events` | `mixed`，告警治理主体共享，报告消费和增强入口部分在 Pro。 |
| `alert-sources-and-enrichment` | 目前更接近 `upstream-common`。 |
| `automated-diagnosis-and-workflows` | `mixed`，`/workflows` 偏共享，自动分析页面位于 Pro。 |
| `conversational-diagnosis-and-reporting` | 以 `commercial-only` 为主。 |
| `data-ingestion-and-oneagent` | `mixed`，基础接入流程共享，Slack Bot 与 Probe 管理为商业增强。 |
| `dataplane-and-topology-governance` | 以 `commercial-only` 为主。 |
| `identity-access-and-organization` | `mixed`，用户/团队页多为共享，登录入口壳层已分叉。 |
| `data-scope-governance` | `mixed`，共享基线之上又叠加了本地路由与隐藏控制。 |
| `platform-administration` | `mixed` 且商业特征最强。 |

## Backend Ownership

### Backend Governance Rule
后端当前没有像前端那样的 `core/oss/pro` 目录分层，因此需要按“包级归属”识别：

1. 先看 `backend/pkg/router/router_api.go` 当前注册了哪些商业仓独有 route group。
2. 再看 `backend/pkg/api`、`backend/pkg/services`、`backend/pkg/model`、`backend/pkg/repository` 中是否存在只在商业仓出现的包。
3. 对两边都存在但内容不同的共享包，标记为 `commercial-variant`。
4. 对商业仓独有 API/service/model/repository 包，标记为 `commercial-only`。

### Backend Comparison Snapshot
以下是当前结构级扫描结果：

| Area | Commercial Count | Upstream Count | Same Path Files | Changed Common Files | Commercial Only | Notes |
|---|---:|---:|---:|---:|---:|---|
| `backend/` all files | 882 | 765 | 765 | 99 | 117 | 含工具、日志、本地数据库、生成物等非运行时文件。 |
| `backend/pkg` | 825 | 723 | 723 | 89 | 102 | 运行时代码层已经存在明显商业扩展。 |
| `backend/pkg/api` | 250 | 213 | 213 | 12 | 37 | API 层已出现多个商业仓独有包。 |
| `backend/pkg/services` | 236 | 203 | 203 | 23 | 33 | service 层商业扩展清晰。 |
| `backend/pkg/model` | 115 | 99 | 99 | 13 | 16 | 请求响应模型和 license 领域模型已分叉。 |
| `backend/pkg/repository` | 183 | 170 | 170 | 34 | 13 | shared repository 上存在较多本地扩展。 |

### Backend Commercial-Only Runtime Areas

| Runtime Area | Classification | Main Evidence | Likely Module Ownership |
|---|---|---|---|
| `backend/pkg/api/alertanalyze`, `backend/pkg/services/alertanalyze` | `commercial-only` | 异常贡献度、mutation detect、anomaly span 等 API / service | `automated-diagnosis-and-workflows` |
| `backend/pkg/api/license`, `backend/pkg/services/license`, `backend/pkg/model/license`, `backend/pkg/repository/database/license` | `commercial-only` | license 查询、上传、检查、激活与 DB 存储 | `platform-administration` |
| `backend/pkg/api/app`, `backend/pkg/services/app`, `backend/pkg/repository/database/app` | `commercial-only` | logo 上传/读取 | `platform-administration` |
| `backend/pkg/api/custom`, `backend/pkg/services/custom` | `commercial-only` | custom query CRUD | `conversational-diagnosis-and-reporting` |
| `backend/pkg/middleware/lic_middleware.go` | `commercial-only` | license 功能门控中间件 | `license-and-edition-gating` |

### Backend Commercial Variants On Shared Packages

| Shared Package | Classification | Observed Difference |
|---|---|---|
| `backend/pkg/router/router_api.go` | `commercial-variant` | 新增 `alertanalyze`、`license`、`app`、`custom` route group，并在多个既有 route group 上挂载 `LicMiddleware()`。 |
| `backend/pkg/router/router.go` | `commercial-variant` | Dify 初始化时增加 `CH` 注入。 |
| `backend/pkg/middleware/middleware.go` | `commercial-variant` | 中间件工厂新增 `licenseService`，`New()` 签名增加 `prometheus.Repo` 与 logger。 |
| `backend/pkg/api/alerts`, `backend/pkg/services/alerts` | `commercial-variant` | 新增事件报告接口与实现。 |
| `backend/pkg/api/dataplane`, `backend/pkg/services/dataplane`, `backend/pkg/model/*dataplane*`, `backend/pkg/repository/database/dao_dataplane.go` | `commercial-variant` + `commercial-only` | 共享 dataplane 包上增加更多治理接口、类型与存储对象。 |
| `backend/pkg/services/integration`, `backend/pkg/model/integration/*`, `backend/pkg/repository/prometheus/*` | `commercial-variant` | 支撑 minimal / probe / mutation / agent 检测等增强行为。 |
| `backend/pkg/repository/database/init_*.go` | `commercial-variant` | feature / router / menu / permission 初始化数据随商业能力扩展而变化。 |

### Backend Non-Source Or Local Artifacts
以下差异目前更像环境、工具或本地产物，不应被当作业务源码归属：

- `backend/.codex/*`
- `backend/logs/apo.log`
- `backend/database-apo.db`
- `backend/__debug_bin*`
- `backend/.gitlab-ci.yml`
- `backend/.gitlab-arm-ci.yml`

文档中应忽略这些内容，不把它们归入业务模块。

## Documentation Rule
后续所有系统文档和模块文档应增加以下约束：

1. 业务模块边界与源码归属边界分开写，不允许用目录名替代源码归属判断。
2. 只要前端入口位于 `core` / `oss`，就要标出它是 `upstream-common` 还是 `commercial-variant`。
3. 只要后端入口位于共享包中，但商业仓额外挂了 route、middleware、service、repository 增强，也要标记为 `commercial-variant`。
4. 对 `commercial-only` 的后端包，文档要明确它们归属于哪个业务模块或横切关注点。
5. 对调用外部服务的增强能力，要同时记录“本仓实现”和“外部边界”，不要因为是商业特性就默认全部在本仓实现。

## Instruction Precedence
本文件用于解释源码归属与分支治理规则，但真正的执行层约束以仓库根目录 [AGENTS.md](/Users/keting/workspace/product/commercial/apo/AGENTS.md) 为准。

如果 `AGENTS.md` 与本文件表述不一致，应优先遵守 `AGENTS.md`，再回头修正文档，避免解释性文档和执行规则漂移。

## Change Workflow

1. 先判断目标代码是否在共享路径或共享包中。
2. 对共享路径/共享包，先与上游仓同路径文件比较。
3. 把需求拆分为：
   - 应先回到上游的共享改动
   - 只保留在商业仓的本地增强
4. 修改完成后，回到文档里更新对应文件的源码归属标记。

## Related Documents

- [system_overview.md](system_overview.md)
- [domain_map.md](domain_map.md)
- [shared_infrastructure.md](shared_infrastructure.md)
- [frontend_backend_mapping.md](frontend_backend_mapping.md)
- [file_to_module_map.md](file_to_module_map.md)
- [analysis/open-questions.md](analysis/open-questions.md)
