# System Overview

## Scope
本文档基于当前仓库静态扫描结果编写，范围覆盖：

- `frontend/`
- `backend/`
- `one-agent/`

本文档的目标不是解释每个页面或每个 API 的细节，而是先固定系统边界、主运行链路、主要业务域和后续文档拆分方式。

## System Boundary
从当前代码可见，系统主体由三个层次组成：

| Layer | Observed Implementation | 说明 |
|---|---|---|
| Web UI | `frontend/` | React + Vite 单页前端，负责页面路由、全局上下文、权限菜单渲染、观测视图和平台管理界面。 |
| Core API Server | `backend/` | Go HTTP 服务，负责认证、license 校验、业务 API、数据库初始化、观测查询、接入管理和部分 AI 分析支撑。 |
| Collector / Agent Side | `one-agent/` | 当前仓库中主要可见 README 与 submodule 声明，说明其是采集侧子系统聚合入口，而不是普通静态资源目录。 |

此外，系统还显式依赖多个外部运行时或外部服务：

- `ClickHouse`
- `Prometheus` 或 `VictoriaMetrics`
- `Kubernetes`
- `Jaeger`
- `Dify`
- `/ai/api`
- `/originx/api`
- `/dify/console/api`
- `/alert/api/v1`

其中前五项在 `backend` 配置与 repository 层有直接实现；后四项在当前仓库主要体现为前端调用边界，不能直接等同为 `backend` 内部实现。

## Repository Roots
当前仓库的主要系统根目录如下：

| Root | Role |
|---|---|
| `frontend/` | 前端应用根目录 |
| `backend/` | 后端应用根目录 |
| `one-agent/` | 采集侧子系统聚合目录 |
| `openspec/` | 仓库级 OpenSpec 目录 |
| `docs/` | 现有文档与图片资源目录 |

## Primary Entry Points
从当前代码可见，主入口与主装配点如下：

| Area | Entry Points | 说明 |
|---|---|---|
| Frontend bootstrap | `frontend/src/index.js`, `frontend/src/App.js` | 负责 Provider 装配、全局样式、Router、布局入口。 |
| Frontend route assembly | `frontend/src/routes.js`, `frontend/src/oss/routes.js`, `frontend/src/pro/routes.js` | 负责基础路由、OSS 路由、Pro 路由的组合。 |
| Backend bootstrap | `backend/main.go` | 负责 logger、HTTP server、graceful shutdown。 |
| Backend server assembly | `backend/pkg/router/router.go`, `backend/pkg/router/router_api.go` | 负责初始化资源、注册 API route group、装配 middleware。 |
| Backend core HTTP kernel | `backend/pkg/core/core.go` | 负责 gin 封装、统一返回、panic 与错误处理、swagger 挂载。 |

## Frontend Architecture Snapshot
从当前前端实现可见：

- 使用 `HashRouter` 作为路由容器。
- 使用 `DefaultLayout` 作为默认主布局。
- 通过 `Provider`、`ToastProvider`、`DeployModeProvider`、`UserProvider`、`DataGroupProvider`、`LicenseProvider` 等全局 provider 装配运行上下文。
- `AppSidebar` 与 `AppContent` 不是纯静态导航，而是依赖后端返回的 `menuItems` 与 `routes`。
- 路由定义按 `baseRoutes`、`ossRoutes`、`proRoutes` 三层组合。
- 前端请求层由统一的 `request.js` 封装，承担 token 注入、token 刷新、语言头设置和错误通知。

## Source Ownership Snapshot
当前系统文档除了写“模块边界”，还必须写“源码归属边界”：

| Area | Ownership View | Current Status |
|---|---|---|
| `frontend/src/core` | 上游共享路径 | 商业仓 337 个文件中，302 个与上游一致，20 个为同路径变体，15 个为商业仓独有文件。 |
| `frontend/src/oss` | 上游共享路径 | 当前 124 个文件与上游一致，1 个文件已漂移。 |
| `frontend/src/pro` | 商业仓主导 | 当前前端 Pro 主体。 |
| `backend/pkg` | 共享包 + 商业扩展 | 723 个同路径文件中，89 个已发生内容漂移，另有 102 个运行时代码文件仅存在于商业仓。 |

这意味着：

- `core` / `oss` 不能被简单写成“商业仓本地源码”。
- 后端也不能只按目录名区分 OSS / Pro，而要按包级扩展和路由注册来判断。
- 文档里必须同时记录业务模块归属与源码归属。

统一说明已收口到 [../source_governance.md](../source_governance.md)。

## Backend Architecture Snapshot
从当前后端实现可见：

- 使用 `router.NewHTTPServer` 初始化完整 HTTP 服务。
- 在启动期初始化数据库、ClickHouse、Prometheus、PolarisAnalyzer、cache、Kubernetes API、Jaeger、Dify、Dataplane。
- 通过 `router_api.go` 挂载主业务 API。
- 通过 `middleware.AuthMiddleware()` 与 `middleware.LicMiddleware()` 对大部分业务接口做认证与 license 限制。
- 使用 `repository -> services -> api handler` 的分层模式组织业务代码。
- 在数据库初始化阶段自动创建或更新 `feature`、`router`、`menu`、`feature_mapping`、`router_insert_page` 等控制数据。

## Major Business Domains
基于当前实现，当前更合理的业务域拆分如下。相比上一版，主要做了三类调整：

- 合并了边界重叠的模块：将原 `integration-and-oneagent-onboarding` 与 `oneagent-collector-subsystem` 收敛为同一个接入模块。
- 拆开了过宽模块：将原 `alerting-and-event-management` 拆为“告警治理”和“告警来源接入”，将原 `ai-workflows-and-root-cause` 拆为“自动化诊断”与“会话/报告”两块。
- 收紧了命名语义：模块名优先表达用户可感知能力，而不是目录名称或技术组件名称。

| Domain | 说明 |
|---|---|
| `service-observability` | 以 service 列表、service detail、endpoint、dependency、topology、RED 视角分析为核心的主观测域。 |
| `logs-exploration-and-parsing` | fault-site logs、full logs、log context、索引字段、日志解析规则与日志表管理。 |
| `trace-exploration-and-profiling` | full trace、fault-site trace、trace filters、trace detail、on/off CPU 与 flame graph。 |
| `dashboards-and-health-monitoring` | system/basic/application/middleware/mysql dashboard，以及服务健康监控与 RYG 状态展示。 |
| `alert-rules-and-incident-events` | 告警规则、通知渠道、告警事件、静默、手工恢复、事件过滤与事件报告读取。 |
| `alert-sources-and-enrichment` | 告警来源接入、provider schema、cluster 绑定、enrich 规则与告警接入管理界面。 |
| `automated-diagnosis-and-workflows` | workflows、alertAnalyze、serviceAnalyze、异常贡献度分析、缺陷检测与 Dify 驱动自动分析。 |
| `conversational-diagnosis-and-reporting` | causeReport、report、chatAI、agent UI、自定义数据视图、会话与模型设置。 |
| `data-ingestion-and-oneagent` | 数据接入、集群安装、安装配置生成、probeInstaller、OneAgent 接入说明与安装资产。 |
| `dataplane-and-topology-governance` | dataplane、服务名规则、自定义拓扑、datasource 能力暴露与数据面治理。 |
| `identity-access-and-organization` | login、user center、user、role、menu、permission、team 等身份与组织管理。 |
| `data-scope-governance` | DataGroup、scope、group 授权关系与数据可见范围治理。 |
| `platform-administration` | TTL 配置、license、deploy mode、logo、branding、平台管理页等平台级管理能力。 |

### Why This Refinement Is More Reasonable

| Change | Why Reasonable |
|---|---|
| `alert-rules-and-incident-events` 与 `alert-sources-and-enrichment` 分离 | 前者围绕“告警发生后的治理与事件处理”，后者围绕“告警如何被接入与预处理”，前后端 route group 也分别落在 `/api/alerts` 与 `/api/alertinput`。 |
| `automated-diagnosis-and-workflows` 与 `conversational-diagnosis-and-reporting` 分离 | 前者更偏自动化分析和 workflow 驱动，后者更偏聊天、报告、证据呈现与交互式诊断，两组前端入口和后端依赖边界不同。 |
| `data-ingestion-and-oneagent` 合并 OneAgent 子系统边界 | 当前主仓库内可见的 OneAgent 主要以安装模板、接入流程和 submodule 边界出现，和接入链路文档放在一起更符合实际可见实现。 |
| `platform-administration` 吞并 retention/config 能力 | `/config` 的 TTL 配置和 `/system/platform-manage` 的 license/branding 都属于平台级管理，不值得再拆成过细单元。 |

## Shared Infrastructure
业务域之外，当前仓库还存在一组明显复用的共享基础设施：

| Infrastructure | 说明 |
|---|---|
| `frontend-app-shell` | 前端启动、布局、菜单、主路由渲染。 |
| `frontend-state-request-and-i18n` | 请求封装、Redux、contexts、国际化、通知。 |
| `backend-http-kernel` | HTTP 启动、gin 封装、middleware、统一响应模型。 |
| `backend-persistence-and-connectors` | database、ClickHouse、Prometheus、Kubernetes、Jaeger、Dify、cache。 |
| `configuration-logging-and-bootstrap` | 配置、logger、环境变量、服务启动初始化。 |
| `deployment-assets-and-static-templates` | Dockerfile、`apo.yml`、`sqlscripts`、`static/integration-tmpl`。 |

## End-To-End Runtime Flow
当前实现中，最典型的端到端链路可以概括为以下结构：

```text
Browser
  -> frontend app shell
  -> route/view
  -> frontend API client
  -> backend route group
  -> service layer
  -> repository / database / observability storage
  -> response model
  -> frontend component tree
```

如果展开到观测场景，则还应补上采集与数据源侧：

```text
OneAgent / external observability inputs
  -> ClickHouse / Prometheus / Jaeger / Kubernetes metadata
  -> backend repositories and services
  -> frontend dashboards / lists / topology / detail views
```

## Documentation Strategy
后续文档不建议按 `frontend/src/oss`、`frontend/src/pro`、`backend/pkg/api` 这类技术目录逐个落笔，而应遵守以下策略：

1. 先以业务域为主进行拆分。
2. 每个业务模块都单独维护前后端链路文档。
3. 共享能力放入 `shared-infrastructure/`，避免在业务文档中重复解释。
4. 横切主题单独放入 `cross-cutting/`，尤其是权限、license、数据范围和外部依赖边界。
5. 对于未在本仓库内完整实现的能力，明确标注为外部边界，而不是补写内部细节。

## Current Risks
当前阶段已经识别出若干系统级高风险点：

- 前端与后端的 route 命名存在漂移。
- 前端调用的部分接口未在当前 `backend` 路由注册中直接观察到。
- 权限模型包含数据库初始化数据，不能只从前端 route 文件推导。
- AI 相关功能明显跨越本仓库与外部服务。
- `one-agent` 当前 checkout 不是完整实现体，而是聚合入口。

这些问题将在 `analysis/open-questions.md` 中继续展开。
