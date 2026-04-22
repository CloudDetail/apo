# Repository Map

## Scope
本文档用于描述当前仓库的目录边界、前后端根目录、主要入口文件和关键实现层次，为后续模块文档提供统一坐标系。

## Top-Level Directories
当前仓库顶层结构中，与系统实现直接相关的目录如下：

| Path | Role | Notes |
|---|---|---|
| `frontend/` | 前端应用根目录 | 含构建配置、源代码、测试、静态资源。 |
| `backend/` | 后端应用根目录 | 含 Go 入口、配置、repository、service、静态模板与 SQL 脚本。 |
| `one-agent/` | 采集侧子系统聚合目录 | 当前可见 README，具体采集实现主要通过 submodule 挂入。 |
| `docs/` | 仓库已有文档目录 | 目前主要是宣传或说明图片资源。 |
| `openspec/` | 仓库级 OpenSpec 根目录 | 含 `changes/` 与 `specs/`。 |
| `.github/` | CI / workflow 辅助目录 | 不属于业务实现主体。 |
| `.codex/` | 本地代理与技能配置目录 | 不属于产品运行时实现。 |

## Frontend Root
前端根目录为 `frontend/`。当前可见的关键文件与关键子目录如下：

| Path | Role |
|---|---|
| `frontend/package.json` | 前端依赖与脚本入口。 |
| `frontend/vite.config.mjs` | Vite 构建配置，含 `/api` 与 `/ai/api` 开发代理。 |
| `frontend/src/index.js` | React 根入口，装配全局 Provider。 |
| `frontend/src/App.js` | 根 Router 与主布局入口。 |
| `frontend/src/routes.js` | 基础路由组合层。 |
| `frontend/src/oss/routes.js` | OSS 功能路由。 |
| `frontend/src/pro/routes.js` | Pro 功能路由。 |
| `frontend/src/core/` | 前端共享基础设施与系统管理相关页面。 |
| `frontend/src/oss/` | OSS 观测功能页面与组件。 |
| `frontend/src/pro/` | Pro 能力、AI、dataplane、license 等页面与组件。 |
| `frontend/public/locales/` | 国际化资源。 |

### Frontend Internal Segmentation
从当前代码组织方式看，前端内部存在三层明显结构：

| Segment | Observed Role |
|---|---|
| `src/core/` | 共享组件、上下文、请求封装、全局 store、系统管理页面、接入中心页面。 |
| `src/oss/` | 观测主域页面，包括 service、logs、trace、dashboard、alerts、workflows。 |
| `src/pro/` | 商业版或扩展能力页面，包括 alert analyze、cause report、chat AI、dataplane、agent、license。 |

需要注意的是，这种分目录方式并不等同于最终文档模块边界。很多业务域实际跨越了 `core`、`oss`、`pro` 三个目录。

### Frontend Source Ownership
当前前端目录还需要叠加“源码归属”视角，而不是只看目录名：

| Segment | Source Ownership | Current Status |
|---|---|---|
| `frontend/src/core/` | 上游开源共享路径 | 商业仓当前 302 个文件与上游一致，20 个为同路径变体，15 个为商业仓独有文件。 |
| `frontend/src/oss/` | 上游开源共享路径 | 当前 124 个文件与上游一致，1 个文件已漂移。 |
| `frontend/src/pro/` | 商业仓主导 | 当前前端 Pro 功能主体。 |
| `frontend/public/locales/*/pro/*` | 商业仓主导 | Pro 页面语言资源。 |
| `frontend/src` 根层壳文件 | 共享壳层但已商业分叉 | `App.js`、`index.js`、`routes.js`、`_nav.js`、`i18n.js` 等已与上游不同。 |

因此：

- `core` 与 `oss` 不能直接写成“商业仓本地前端代码”。
- 即使是非商业路径，也可能已经和上游不同。
- 模块文档必须额外记录前端源码归属。

## Refined Capability Groups
为了避免按目录直接写文档导致同一能力被拆散，当前文档体系使用以下能力分组作为主导航：

- `service-observability`
- `logs-exploration-and-parsing`
- `trace-exploration-and-profiling`
- `dashboards-and-health-monitoring`
- `alert-rules-and-incident-events`
- `alert-sources-and-enrichment`
- `automated-diagnosis-and-workflows`
- `conversational-diagnosis-and-reporting`
- `data-ingestion-and-oneagent`
- `dataplane-and-topology-governance`
- `identity-access-and-organization`
- `data-scope-governance`
- `platform-administration`

这组能力边界比单纯使用 `oss`、`pro`、`core` 更接近实际业务能力，也更利于建立稳定的前后端映射。

## Backend Root
后端根目录为 `backend/`。当前可见的关键文件与关键子目录如下：

| Path | Role |
|---|---|
| `backend/go.mod` | Go module 根与依赖清单。 |
| `backend/main.go` | 后端进程入口。 |
| `backend/config/` | 配置结构、默认配置、初始化 SQL。 |
| `backend/pkg/router/` | HTTP 服务器装配与 API route group 注册。 |
| `backend/pkg/core/` | gin 封装、context、错误模型、统一响应。 |
| `backend/pkg/middleware/` | `AuthMiddleware`、`LicMiddleware`、权限相关中间件。 |
| `backend/pkg/api/` | 各业务域 HTTP handler。 |
| `backend/pkg/services/` | 业务 service 层。 |
| `backend/pkg/repository/` | 数据访问与外部系统 connector。 |
| `backend/pkg/model/` | 请求、响应、领域模型。 |
| `backend/sqlscripts/` | 初始化或升级脚本、API/feature 相关数据。 |
| `backend/static/` | 接入模板、预定义 metrics、静态资源。 |
| `backend/Dockerfile` | 后端镜像构建入口。 |

### Backend Layering
从当前实现可见，后端主要按以下结构组织：

```text
main
  -> router
  -> middleware
  -> api handlers
  -> services
  -> repositories
  -> databases / external observability systems
```

`pkg/api/`、`pkg/services/`、`pkg/repository/` 的命名维度大体一致，说明后端是按业务域切 API 与 service，再按存储/外部系统切 repository。

### Backend Source Ownership
后端当前没有 `pro/` 这种显式目录分层，因此需要按包级对比：

| Area | Source Ownership | Current Status |
|---|---|---|
| `backend/pkg/api` | 共享包 + 商业独有包 | `alertanalyze`、`app`、`custom`、`license` 仅在商业仓存在，另有若干共享 API 包已被改写。 |
| `backend/pkg/services` | 共享包 + 商业独有包 | `alertanalyze`、`app`、`custom`、`license` 为商业独有，`alerts`、`dataplane`、`integration` 等共享包也已扩展。 |
| `backend/pkg/model` | 共享包 + 商业独有模型 | `license` 目录仅存在于商业仓，另有 alert analyze / custom / app 相关请求响应模型。 |
| `backend/pkg/repository` | 共享 repository + 本地扩展 | `database/app`、`database/license` 只在商业仓出现，其他 repository 目录也有同路径漂移。 |
| `backend/pkg/router`, `backend/pkg/middleware` | 共享内核 + 商业门控扩展 | `router_api.go` 与 `middleware.go` 被改写，`lic_middleware.go` 为商业仓独有。 |

因此，后端文档不能只按目录深度或文件夹名判断 OSS / Pro。更可靠的方式是看：

1. `router_api.go` 当前注册了哪些 route group。
2. `pkg/api` / `pkg/services` 是否存在只在商业仓出现的包。
3. 共享包是否已被本地扩展。

统一归属规则见 [../source_governance.md](../source_governance.md)。

## OneAgent Root
`one-agent/` 当前在仓库中主要承担采集侧子系统聚合说明的角色。

| Path | Role |
|---|---|
| `one-agent/README.md` | 说明 OneAgent 的功能边界与集成组件。 |
| `.gitmodules` | 声明 `node-agent`、`preload`、`ilogtail`、`odigos` 等 submodule。 |

从当前 checkout 可见，`one-agent/` 并不是普通前后端目录的一部分，而是采集侧技术栈入口。后续如果需要写 OneAgent 内部实现，必须先确认 submodule 是否已完整拉取。

## OpenSpec And Documentation Roots

| Path | Role |
|---|---|
| `openspec/changes/` | 仓库级变更提案与归档。 |
| `openspec/specs/` | 仓库级规范沉淀。 |
| `docs/openspec-system/` | 本轮新增的系统逆向文档目录。 |

## Primary Entry Point Map
当前仓库中最重要的入口文件如下：

| Area | File | Reason |
|---|---|---|
| Frontend runtime entry | `frontend/src/index.js` | 应用启动、Provider 装配、主题、i18n、上下文入口。 |
| Frontend router entry | `frontend/src/App.js` | `HashRouter`、`Login`、主布局入口。 |
| Frontend route assembly | `frontend/src/routes.js` | 组合 `baseRoutes`、`ossRoutes`、`proRoutes`。 |
| Backend process entry | `backend/main.go` | logger、HTTP server、shutdown。 |
| Backend server assembly | `backend/pkg/router/router.go` | 初始化 repository、service 依赖与 `setApiRouter`。 |
| Backend API registry | `backend/pkg/router/router_api.go` | 注册全部主要 API route group。 |
| Backend HTTP kernel | `backend/pkg/core/core.go` | 统一 context、错误与响应处理。 |

## Supporting Runtime Assets
当前仓库里还有几类不属于“业务代码主体”，但对系统运行非常关键的资产：

| Asset Area | Paths | Role |
|---|---|---|
| Configuration | `backend/config/apo.yml`, `backend/config/config.go` | 系统连接、端口、logger、database、ClickHouse、Prometheus、Dify 等默认配置。 |
| Static install templates | `backend/static/integration-tmpl/` | 集群接入、OneAgent 安装、docker compose / k8s 模板。 |
| Database initialization | `backend/pkg/repository/database/init_*.go`, `backend/sqlscripts/` | 初始化 feature、router、menu、permission、translation 等控制数据。 |
| Localization | `frontend/public/locales/`, `frontend/src/i18n.js` | 页面文案、路由名称、模块说明国际化。 |

## Observed Structural Characteristics
当前仓库有几个结构特征，会直接影响后续文档组织：

1. 前端菜单与页面可见性不完全由前端本地决定，而是由后端返回的 `menuItem` 与 `routes` 决定。
2. 后端在数据库初始化阶段主动维护 `feature`、`router`、`menu`、`feature_mapping` 等权限控制元数据。
3. `frontend/src/pro/` 中有多项能力依赖外部服务边界，不能简单视为 `backend` 的内部子模块。
4. `one-agent/` 是系统重要边界，但当前主仓库内只保留聚合入口和说明，不保留完整实现源码。

## Documentation Implication
基于当前目录结构，后续文档需要避免两个常见误区：

- 不能只按目录写文档，否则同一个业务域会被拆散到 `core`、`oss`、`pro`、`backend/pkg/api`、`backend/pkg/services` 多处。
- 不能把前端调用到的所有路径都默认记为本仓库后端接口，必须先确认其是否在 `backend/pkg/router/router_api.go` 中注册或是否属于外部边界。
