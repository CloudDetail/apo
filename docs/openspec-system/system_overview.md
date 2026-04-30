# System Overview

## Scope
本文档是当前 `docs/openspec-system/` 的统一总入口。目标是用一份稳定的系统级视图，把仓库中的前端、后端、采集侧聚合目录、业务模块、共享基础设施和横切关注点连接起来。

## System Boundary

| Layer | Primary Source Roots | Description |
|---|---|---|
| Web UI | `frontend/` | React + Vite 单页应用，负责页面路由、菜单渲染、状态上下文、请求封装与业务页面。 |
| Core API Server | `backend/` | Go HTTP 服务，负责认证、license 校验、业务 API、数据访问、接入治理与平台管理。 |
| Collector Aggregation | `one-agent/` | 当前仓库中主要表现为 README 与 submodule 聚合入口，说明采集侧存在独立实现边界。 |
| External Runtime Dependencies | ClickHouse, Prometheus or VictoriaMetrics, Kubernetes, Jaeger, Dify, `/ai/api`, `/originx/api`, `/dify/console/api`, `/alert/api/v1` | 前五类在 `backend/pkg/repository` 中有直接访问实现，后四类主要表现为前端调用或旁路服务边界。 |

## Repository Roots

| Root | Role |
|---|---|
| `frontend/` | 前端根目录。 |
| `backend/` | 后端根目录。 |
| `one-agent/` | 采集侧聚合目录。 |
| `docs/openspec-system/` | 当前整编后的 OpenSpec 文档根目录。 |
| `openspec/` | 仓库级变更与规格目录。 |

## Runtime Structure

```text
Browser
  -> frontend app shell
  -> route/page
  -> frontend API wrapper
  -> backend route group
  -> service layer
  -> repository / database / observability storage / external service
  -> response model
  -> frontend state and UI
```

观测类主链路还会再向前补一层：

```text
OneAgent / external alert source / external metrics source
  -> ClickHouse / Prometheus / Jaeger / database / Kubernetes metadata
  -> backend repositories and services
  -> frontend dashboards / logs / trace / topology / diagnosis views
```

## Canonical Business Domains

| Module | Primary Capability |
|---|---|
| `service-observability` | service 列表、service detail、依赖、endpoint、错误实例与 RED 观察。 |
| `logs-exploration-and-parsing` | 故障现场日志、全量日志、上下文、索引字段、解析规则与其他日志表管理。 |
| `trace-exploration-and-profiling` | trace 列表、trace 详情、筛选器、flame graph、on/off CPU。 |
| `dashboards-and-health-monitoring` | system/basic/application/middleware/mysql dashboard 与服务健康监控。 |
| `alert-rules-and-incident-events` | 告警规则、通知渠道、事件列表、静默、手工恢复、事件报告读取。 |
| `alert-sources-and-enrichment` | 告警源接入、cluster、schema、enrich 规则与来源治理。 |
| `automated-diagnosis-and-workflows` | workflow、异常贡献度分析、缺陷检测与自动化分析链路。 |
| `conversational-diagnosis-and-reporting` | 聊天式分析、cause report、report、agent UI、自定义查询与对话接口。 |
| `data-ingestion-and-oneagent` | 数据接入、集群安装、安装配置生成、OneAgent 相关入口。 |
| `dataplane-and-topology-governance` | dataplane、服务名规则、自定义拓扑与数据面治理。 |
| `identity-access-and-organization` | 登录、用户、角色、菜单、权限、团队与组织控制。 |
| `data-scope-governance` | DataGroup、授权、数据源范围与数据可见域治理。 |
| `platform-administration` | TTL 配置、license、logo、deploy mode 与平台级设置。 |

## Shared Infrastructure

| Shared Area | Description |
|---|---|
| `frontend-app-shell` | `frontend/src/index.js`、`App.js`、`DefaultLayout`、`AppSidebar`、`AppContent` 等前端装配层。 |
| `frontend-state-request-and-i18n` | `request.js`、Redux、contexts、toast、国际化。 |
| `backend-http-kernel` | `main.go`、`pkg/router`、`pkg/core`、`pkg/middleware`。 |
| `backend-persistence-and-connectors` | `pkg/repository/database`、`clickhouse`、`prometheus`、`jaeger`、`kubernetes`、`dify`、`cache`、`dataplane`。 |
| `configuration-logging-and-bootstrap` | `backend/config`、`pkg/logger`、构建镜像与启动配置。 |
| `deployment-assets-and-static-templates` | `backend/static`、`backend/sqlscripts`、`frontend/public`、`one-agent/README.md`。 |

## Cross-Cutting Concerns

| Concern | Description |
|---|---|
| `auth-and-session` | 登录、token、refresh、`AuthMiddleware`、前端 request 注入。 |
| `permission-menu-router-model` | `feature`、`menu`、`router`、`feature_mapping`、`router_insert_page` 与菜单路由装配。 |
| `license-and-edition-gating` | `LicenseContext`、`LicMiddleware`、页面与 API 的能力开关。 |
| `data-group-and-time-range-context` | `DataGroupContext`、time range reducer 对多个观测模块的共同影响。 |
| `observability-data-sources` | Trace、Log、Metric、Event、Alert、Topology、AgentSRE findings 等后台数据来源、存储和查询入口。 |
| `external-service-boundaries` | `/ai/api`、`/originx/api`、`/dify/console/api`、`/alert/api/v1` 等旁路或外部边界。 |

## Source Ownership Snapshot
当前系统文档除了写“业务模块边界”，还必须写“源码归属边界”：

| Area | Ownership View | Current Status |
|---|---|---|
| `frontend/src/core` | 上游共享路径 | 商业仓 337 个文件中，302 个与上游一致，20 个为同路径变体，15 个为商业仓独有文件。 |
| `frontend/src/oss` | 上游共享路径 | 当前 124 个文件与上游一致，1 个文件已漂移。 |
| `frontend/src/pro` | 商业仓主导 | 当前前端 Pro 主体。 |
| `backend/pkg` | 共享包 + 商业扩展 | 723 个同路径文件中，89 个已漂移，另有 102 个运行时代码文件仅存在于商业仓。 |

这意味着：

- `core` / `oss` 不能被简单写成“商业仓本地源码”。
- 后端也不能只按目录名区分 OSS / Pro，而要按包级扩展和路由注册来判断。
- 文档里必须同时记录业务模块归属与源码归属。

统一说明见 [source_governance.md](source_governance.md)。

## Canonical Navigation Through The Docs

1. 先读 [module_index.md](module_index.md) 确认模块边界。
2. 再读 [frontend_backend_mapping.md](frontend_backend_mapping.md) 建立前后端对应关系。
3. 如果要理解后台数据从哪来，读 [observability_data_sources.md](observability_data_sources.md)。
4. 如果要查路由、页面、API、模型清单，则继续看：
   - [route_to_module_map.md](route_to_module_map.md)
   - [page_inventory.md](page_inventory.md)
   - [api_to_module_map.md](api_to_module_map.md)
   - [api_inventory.md](api_inventory.md)
   - [model_inventory.md](model_inventory.md)
5. 如果要进入模块细节，则从 `modules/<module>/module_overview.md` 继续。

## Coverage Statement
当前整编目标覆盖整个已实现系统，而不是单模块说明。按当前扫描结果：

- 13 个主要业务模块已形成独立详细文档目录。
- 全部主业务前端路由都可以归属到某个业务模块。
- 全部已注册后端 route group 都可归属到业务模块或共享基础设施。
- 仍有少量前端页面文件、前端 API wrapper 路径和共享模型文件存在“历史残留、外部边界或实现漂移”问题，已在各全局索引文档中单独标注。
