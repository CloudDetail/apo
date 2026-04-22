# Analysis Order

## Goal
本文档定义后续轮次的推荐分析顺序。排序原则不是按目录遍历，而是按“建立全局正确性所需的依赖顺序”推进。

## Current Status
第一阶段已经完成并写入以下基线文档：

- `system/system-overview.md`
- `system/repository-map.md`
- `system/frontend-backend-linkage.md`
- `source-mapping/repository-to-module-map.md`
- `analysis/open-questions.md`

后续阶段应在此基础上继续展开。

## Recommended Order
| Phase | Focus | Primary Outputs | Why This Order Is Reasonable |
|---|---|---|---|
| Phase 1 | 系统边界与全局映射 | 已完成的基线文档 | 先固定系统边界、前后端根、业务域拆分与主要风险，否则后续模块文档容易相互冲突。 |
| Phase 2 | 平台底座与横切控制 | `runtime-topology.md`, `backend-architecture.md`, `frontend-architecture.md`, `permission-menu-router-model`, `configuration-logging-and-bootstrap` | 权限、启动、配置、共享 connector 会影响所有业务模块，是后续所有说明的公共前提。 |
| Phase 3 | 核心观测主域 | `service-observability`, `logs-exploration-and-parsing`, `trace-exploration-and-profiling`, `dashboards-and-health-monitoring` | 这些是系统核心价值路径，也是其他模块最常引用的数据消费域。 |
| Phase 4 | 告警与接入链路 | `alert-rules-and-incident-events`, `alert-sources-and-enrichment`, `data-ingestion-and-oneagent` | 告警治理、告警来源接入和 OneAgent 接入共同连接采集侧、事件侧和分析侧。 |
| Phase 5 | 自动化诊断与交互诊断 | `automated-diagnosis-and-workflows`, `conversational-diagnosis-and-reporting` | 这两块都依赖前面观测和告警基础，但它们的前端交互模型与后端边界明显不同。 |
| Phase 6 | 平台治理与管理域 | `dataplane-and-topology-governance`, `identity-access-and-organization`, `data-scope-governance`, `platform-administration` | 这些模块要么是治理平面，要么是平台与组织管理能力，适合在核心业务链路稳定后集中补齐。 |

## Detailed Priority List

### 1. System Baseline
优先继续补齐以下系统文档：

- `system/runtime-topology.md`
- `system/frontend-architecture.md`
- `system/backend-architecture.md`
- `system/external-dependencies.md`
- `system/deployment-and-configuration.md`

原因：

- 当前已经知道前后端入口，但还未把运行时拓扑画清。
- 前端与后端各自的内部结构还需要正式文档承接。
- 外部依赖边界是后续 AI、workflows、one-agent、dataplane 文档的前提。

### 2. Permission And Control Metadata
第二优先级是以下横切主题：

- `cross-cutting/permission-menu-router-model.md`
- `cross-cutting/license-and-edition-gating.md`
- `cross-cutting/data-group-and-time-range-context.md`

原因：

- 当前项目的页面可见性不是纯前端静态逻辑。
- 后端数据库初始化数据与前端导航直接耦合。
- DataGroup 与时间范围是多个观测模块的共享上下文。

### 3. Backend Shared Foundations
第三优先级是后端共享底座：

- `shared-infrastructure/backend-http-kernel.md`
- `shared-infrastructure/backend-persistence-and-connectors.md`
- `shared-infrastructure/configuration-logging-and-bootstrap.md`

原因：

- 后续每个模块都会引用 `/api -> service -> repository` 链路。
- 不先写共享底座，会导致模块文档重复解释 ClickHouse、Prometheus、database、middleware。

### 4. Frontend Shared Foundations
第四优先级是前端共享底座：

- `shared-infrastructure/frontend-app-shell.md`
- `shared-infrastructure/frontend-state-request-and-i18n.md`

原因：

- 几乎所有页面模块都共享这些上下文。
- token 刷新、i18n、sidebar、route render、deploy mode 会影响所有业务文档的解释方式。

### 5. Core Observability Domains
第五优先级开始进入主业务模块：

1. `modules/service-observability/`
2. `modules/logs-exploration-and-parsing/`
3. `modules/trace-exploration-and-profiling/`
4. `modules/dashboards-and-health-monitoring/`

原因：

- 这是整个系统最中心的数据消费主链。
- service detail、logs、trace 之间已有明显交叉引用关系。
- dashboard 与 service health 在用户心智上属于“聚合观测视图”，适合跟在主观测模块之后。

### 6. Alert And Ingestion Domains
第六优先级：

1. `modules/alert-rules-and-incident-events/`
2. `modules/alert-sources-and-enrichment/`
3. `modules/data-ingestion-and-oneagent/`

原因：

- 告警域一方面消费观测数据，一方面给 AI 分析提供输入。
- 接入与 OneAgent 决定数据是如何进入系统的。

### 7. Diagnosis And Automation Domains
第七优先级：

1. `modules/automated-diagnosis-and-workflows/`
2. `modules/conversational-diagnosis-and-reporting/`

原因：

- 自动化诊断和会话式诊断虽然都属于 AI 相关能力，但它们依赖的页面入口、API 边界和外部服务并不完全相同。
- 先拆开再写，可以避免把 workflow iframe、异常检测、chat、report 混成一个过宽模块。

### 8. Governance And Platform Domains
最后补齐管理域：

1. `modules/dataplane-and-topology-governance/`
2. `modules/identity-access-and-organization/`
3. `modules/data-scope-governance/`
4. `modules/platform-administration/`

原因：

- 这些模块横跨多个业务域，适合在主体观测链路稳定后集中整理。
- 其中 `data-scope-governance` 与 `platform-administration` 也可根据后续需求提前插队，因为它们是明显的横切主题。

## Suggested Next Round
如果下一轮继续推进，建议按以下顺序直接开写：

1. `system/runtime-topology.md`
2. `shared-infrastructure/backend-http-kernel.md`
3. `cross-cutting/permission-menu-router-model.md`
4. `shared-infrastructure/frontend-app-shell.md`

这样可以先把“系统怎么启动、怎么路由、怎么决定页面可见性、前端怎么装配”这一组关键骨架固定下来。

## Stop Conditions
在以下情况出现前，不建议进入更细的页面级或 API 级文档：

- 外部边界尚未分清。
- 前端 route 与后端 permission router 映射尚未核对。
- V1/V2 API 并存关系尚未说明。
- `one-agent` submodule 边界尚未确认。

否则后续细节文档很容易在错误边界上越写越多。
