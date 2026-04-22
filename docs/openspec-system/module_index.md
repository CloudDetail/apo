# Module Index

## Scope
本文档列出当前整编后的规范化模块清单，并把业务模块、共享基础设施和横切关注点明确分开。

## Business Modules

| Module | Category | Canonical Purpose | Detailed Docs |
|---|---|---|---|
| `service-observability` | Business | service 观察、依赖、错误实例、endpoint、RED 视图 | [module_overview](modules/service-observability/module_overview.md) |
| `logs-exploration-and-parsing` | Business | 日志查询、上下文、索引、解析规则与日志表管理 | [module_overview](modules/logs-exploration-and-parsing/module_overview.md) |
| `trace-exploration-and-profiling` | Business | trace 列表、详情、筛选和 profiling 能力 | [module_overview](modules/trace-exploration-and-profiling/module_overview.md) |
| `dashboards-and-health-monitoring` | Business | 各类 dashboard 与服务健康监控 | [module_overview](modules/dashboards-and-health-monitoring/module_overview.md) |
| `alert-rules-and-incident-events` | Business | 告警规则、通知、事件和事件报告读取 | [module_overview](modules/alert-rules-and-incident-events/module_overview.md) |
| `alert-sources-and-enrichment` | Business | 告警源接入、schema、cluster 与 enrich 规则 | [module_overview](modules/alert-sources-and-enrichment/module_overview.md) |
| `automated-diagnosis-and-workflows` | Business | 自动化诊断、workflow、异常贡献度与缺陷检测 | [module_overview](modules/automated-diagnosis-and-workflows/module_overview.md) |
| `conversational-diagnosis-and-reporting` | Business | 聊天式诊断、cause report、agent、自定义查询 | [module_overview](modules/conversational-diagnosis-and-reporting/module_overview.md) |
| `data-ingestion-and-oneagent` | Business | 数据接入、集群安装、配置生成与 OneAgent 入口 | [module_overview](modules/data-ingestion-and-oneagent/module_overview.md) |
| `dataplane-and-topology-governance` | Business | dataplane、服务名规则、自定义拓扑与数据面治理 | [module_overview](modules/dataplane-and-topology-governance/module_overview.md) |
| `identity-access-and-organization` | Business | 登录、用户、角色、菜单、权限、团队 | [module_overview](modules/identity-access-and-organization/module_overview.md) |
| `data-scope-governance` | Business | DataGroup、授权与数据范围治理 | [module_overview](modules/data-scope-governance/module_overview.md) |
| `platform-administration` | Business | TTL、license、logo、deploy mode 和平台设置 | [module_overview](modules/platform-administration/module_overview.md) |

## Shared Infrastructure

| Shared Module | Purpose |
|---|---|
| `frontend-app-shell` | 前端启动、布局、根路由与主容器。 |
| `frontend-state-request-and-i18n` | 请求层、全局状态、上下文、国际化。 |
| `backend-http-kernel` | HTTP 启动、路由注册、中间件、统一响应。 |
| `backend-persistence-and-connectors` | 数据库与外部观测系统连接。 |
| `configuration-logging-and-bootstrap` | 配置、日志、环境与启动。 |
| `deployment-assets-and-static-templates` | 静态资源、安装模板、初始化脚本。 |

## Cross-Cutting Concerns

| Cross-Cutting Concern | Purpose |
|---|---|
| `auth-and-session` | 登录、token、refresh 与会话恢复。 |
| `permission-menu-router-model` | 菜单、router、feature、页面插页元数据。 |
| `license-and-edition-gating` | 版本/许可证能力裁剪。 |
| `data-group-and-time-range-context` | 数据范围与时间范围的共同上下文。 |
| `external-service-boundaries` | 外部服务与本仓库实现的边界。 |

## Boundary Normalization Result
本次整编后的规范化结果如下：

- 没有发现两个不同目录名指向同一业务模块的重复文档目录。
- 统一采用业务能力命名，而不是 `core`、`oss`、`pro` 目录命名。
- 统一把 `health`、`metric`、`network`、`k8s` 这类支撑接口从业务模块中抽离为共享支撑层。
- 统一把 `/ai/api`、`/originx/api`、`/dify/console/api`、`/alert/api/v1` 视为外部或旁路边界。

## Source Governance Reference
模块目录解决的是业务能力边界，不解决源码归属边界。当前文档还必须额外区分：

- `upstream-common`
- `commercial-variant`
- `commercial-only`
- `upstream-edition-gated`

尤其是前端 `core` / `oss` 与后端共享包，不能因为它们承载商业能力就全部写成 Pro 独有，也不能因为路径是共享路径就默认等于上游。统一说明见 [source_governance.md](source_governance.md)。

## Recommended Reading Order

1. [system_overview.md](system_overview.md)
2. [domain_map.md](domain_map.md)
3. [frontend_backend_mapping.md](frontend_backend_mapping.md)
4. 目标模块的 `modules/<module>/module_overview.md`
5. 需要进一步对照时再看：
   - [route_to_module_map.md](route_to_module_map.md)
   - [api_to_module_map.md](api_to_module_map.md)
   - [page_inventory.md](page_inventory.md)
   - [api_inventory.md](api_inventory.md)
   - [model_inventory.md](model_inventory.md)
