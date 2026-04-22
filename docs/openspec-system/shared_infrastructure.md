# Shared Infrastructure

## Scope
本文档把全项目中不应归入单一业务模块的共享设施单独列出，并明确哪些内容属于“共享基础设施”，哪些内容属于“横切关注点”。

## Shared Infrastructure Modules

| Shared Module | Primary Sources | Main Responsibility |
|---|---|---|
| `frontend-app-shell` | `frontend/src/index.js`, `frontend/src/App.js`, `frontend/src/routes.js`, `frontend/src/oss/routes.js`, `frontend/src/pro/routes.js`, `frontend/src/core/layout/DefaultLayout.js`, `frontend/src/core/components/AppSidebar.js`, `frontend/src/core/components/AppContent.js` | 装配 Provider、布局、根路由、侧边栏和页面容器。 |
| `frontend-state-request-and-i18n` | `frontend/src/core/utils/request.js`, `frontend/src/core/contexts/*`, `frontend/src/core/store/*`, `frontend/src/i18n.js`, `frontend/public/locales/*` | 统一请求、token 注入、错误处理、全局状态、国际化和文案。 |
| `backend-http-kernel` | `backend/main.go`, `backend/pkg/router/*`, `backend/pkg/core/*`, `backend/pkg/middleware/*` | 启动 HTTP 服务、注册路由、处理中间件、统一返回和错误。 |
| `backend-persistence-and-connectors` | `backend/pkg/repository/database`, `clickhouse`, `prometheus`, `jaeger`, `kubernetes`, `dify`, `cache`, `dataplane`, `polarisanalyzer` | 连接数据库、缓存、观测系统和外部分析服务。 |
| `configuration-logging-and-bootstrap` | `backend/config/*`, `backend/pkg/logger/*`, `backend/Dockerfile` | 配置加载、日志初始化、进程启动和镜像打包。 |
| `deployment-assets-and-static-templates` | `backend/static/*`, `backend/sqlscripts/*`, `frontend/public/*`, `one-agent/README.md` | 集群安装模板、初始化数据脚本、静态资源和安装说明。 |

## Shared Service And API Support Areas

| Source Area | Recommended Classification | Reason |
|---|---|---|
| `backend/pkg/services/common/*` | 共享服务支撑 | 提供 data group tree、filter 等通用能力，被多个模块复用。 |
| `backend/pkg/services/utils/filter.go` | 共享工具 | 不承载独立业务能力。 |
| `backend/pkg/api/health` | 共享运行时支撑 | 健康检查接口，不对应独立业务页面。 |
| `backend/pkg/api/k8s` | 共享观测元数据接口 | 供观测能力拉取 namespace/pod 元数据，支撑多个模块。 |
| `backend/pkg/api/metric` | 共享指标查询接口 | 被 dashboard、service、诊断等多个模块间接消费。 |
| `backend/pkg/api/network` | 共享网络/拓扑支撑接口 | 供健康监控和拓扑相关视图复用。 |

## Source Ownership Notes
共享基础设施除了有“职责边界”，还有“源码归属边界”。当前最需要额外标注的是以下几组：

| Shared Area | Key Files | Ownership Status | Notes |
|---|---|---|---|
| `frontend-app-shell` | `frontend/src/App.js`, `frontend/src/index.js`, `frontend/src/routes.js`, `frontend/src/_nav.js`, `frontend/src/core/layout/DefaultLayout.js`, `frontend/src/core/components/AppHeader.js`, `frontend/src/core/components/AppSidebar.js` | `commercial-variant on shared paths` | 这些文件属于共享壳层，但已经注入 deploy mode、license、导航裁剪等商业增强。 |
| `frontend-state-request-and-i18n` | `frontend/src/constants.js`, `frontend/src/i18n.js`, `frontend/src/core/utils/request.js`, `frontend/src/core/store/reducers/timeRangeReducer.js`, `frontend/src/core/components/DateTime/*` | `mixed shared-path drift` | 差异不全是 Pro 功能，其中不少是共享层本地分叉。 |
| `backend-http-kernel` | `backend/pkg/router/router_api.go`, `backend/pkg/router/router.go`, `backend/pkg/middleware/middleware.go`, `backend/pkg/middleware/lic_middleware.go` | `mixed shared-package + commercial extension` | 商业仓既新增了 `LicMiddleware`，也在共享 router / middleware 上继续扩展。 |
| `backend-persistence-and-connectors` | `backend/pkg/repository/database/*`, `backend/pkg/repository/prometheus/*`, `backend/pkg/repository/clickhouse/*` | `mixed shared-package drift` | repository 层大部分目录与上游同名，但内部已有 license、app、mutation、dataplane 等本地扩展。 |

详细归属规则见 [source_governance.md](source_governance.md)。

## Cross-Cutting Concerns
以下内容重要，但不属于“共享基础设施模块”。它们更适合作为横切关注点对待：

| Concern | Primary Sources | Why It Is Cross-Cutting Rather Than Infrastructure |
|---|---|---|
| `auth-and-session` | Login 页面、`request.js`、`AuthMiddleware`、JWT 工具 | 认证与会话状态作用于所有业务请求和页面访问。 |
| `permission-menu-router-model` | `UserContext.js`、`AppSidebar.js`、`backend/pkg/services/permission/*`、`backend/pkg/repository/database/init_*.go` | 决定页面可见性和菜单结构，是全站访问控制模型。 |
| `license-and-edition-gating` | `LicenseContext.tsx`、`DefaultLayout.js`、`LicMiddleware` | 决定功能是否对当前用户/实例开放。 |
| `data-group-and-time-range-context` | `DataGroupContext.tsx`、Redux time range reducer、`backend/pkg/services/data/*` | 共同影响多个观测查询。 |
| `external-service-boundaries` | `frontend/src/pro/api/*`、`frontend/src/core/api/workflows.ts`、`backend/config/apo.yml` | 用于划清本仓库实现和外部服务边界。 |

## Separation Rule
在模块文档中遇到以下内容时，应优先引用本文件而不是重复展开：

- 根路由装配、布局和菜单渲染。
- token 注入、全局 request 错误处理和国际化。
- 后端 HTTP 启动、middleware 和统一响应包装。
- repository connector 的通用接入方式。
- 健康检查、Kubernetes 元数据、通用指标查询等支撑接口。

## Current Coverage Result
按当前扫描结果：

- 共享基础设施已经可以覆盖主壳层、请求层、后端内核和 repository connector。
- `backend/pkg/api/health`、`metric`、`network`、`k8s` 这类接口不再需要被强行挂到单一业务模块。
- 尚未细写的 `shared-infrastructure/` 子目录专题文档，应在后续补充时参考本文件的边界定义，而不是重新切分模块。
