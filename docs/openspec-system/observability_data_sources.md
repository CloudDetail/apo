# Observability Data Sources

## Scope
本文档用于把 APO、apo-dataplane、AgentSRE 和当前集群接入链路中的“后台数据从哪里来、有哪些数据、谁负责查询”统一讲清楚。

读者可以把它当成面向产品、交付、实施和研发协作的总览：

- APO 后台不是所有数据的生产者，而是查询、整理和转发入口。
- OneAgent / 外部接入系统负责采集现场数据。
- ClickHouse、Prometheus / VictoriaMetrics、Jaeger、Kubernetes、关系型数据库等负责保存或提供不同类型的数据。
- apo-dataplane 把不同厂商和不同存储的数据源包装成统一能力。
- AgentSRE 通过 APO 后台、apo-dataplane、VictoriaMetrics 等接口拿数据，再生成诊断过程、证据和结论。

## Mental Model
可以先用一个类比理解：

| 系统角色 | 概念类比 | 主要职责 |
|---|---|---|
| APO Frontend | 操作台 / 大屏 | 展示服务、日志、链路、指标、告警和诊断结果。 |
| APO Backend | 调度中心 | 接收前端请求，做权限、license、数据范围过滤，再去不同地方查数据。 |
| apo-dataplane | 数据源适配器 / 翻译官 | 把 APO、Datadog、ARMS、SkyWalking、Tempo、Jaeger、Loki、Elastic 等数据源统一成服务、实例、拓扑、日志、链路、指标接口。 |
| OneAgent | 现场采集员 | 采集 trace、log、metric、eBPF / Polaris、网络和进程相关数据。 |
| ClickHouse | 大型明细仓库 | 保存日志、trace 明细、告警事件、拓扑、profiling、workflow record 等明细数据。 |
| Prometheus / VictoriaMetrics | 时序指标仓库 | 保存 CPU、内存、请求量、错误率、延迟、RED、eBPF 等时间序列指标。 |
| Jaeger | Trace 详情系统 | 提供完整 trace 查询和 trace UI。 |
| Kubernetes API | 集群现场接口 | 提供 namespace、pod、configmap、告警规则配置、事件等集群元数据。 |
| AgentSRE | 智能分析师 | 查询日志、trace、metric、拓扑和实例数据，汇总成 evidence / findings / root cause。 |

## Runtime Component Inventory
从当前 APO 仓、apo-dataplane 仓、AgentSRE 仓和部署模板综合看，系统组件不只包括“前端、后端、Dataplane、AgentSRE、数据库”。更完整的清单应按运行层分组理解。

### Core APO Runtime

| 组件 | 代码 / 配置证据 | 角色 |
|---|---|---|
| APO Frontend | `frontend/`、`frontend/apo-front-deploy.yml` | React 页面、路由、状态、图表和交互入口。 |
| Frontend Nginx | `frontend/nginx.conf`、`frontend/apo-front-deploy.yml` | 承载静态前端，并反向代理 `/api/`、`/jaeger/`、`/grafana/`、`/alert`。 |
| APO Backend | `backend/`、`backend/pkg/router/router.go` | 核心 API 服务，初始化 DB、ClickHouse、Prometheus、Kubernetes、Jaeger、Dify、Dataplane 等连接。 |
| Embedded Metadata Server | `backend/pkg/router/extra_router.go`、`meta_server.*` | 可选内嵌 `/metadata` 接口，由 `github.com/CloudDetail/metadata` 提供 handlers。 |
| In-Memory Token Cache | `backend/pkg/repository/cache/cache.go` | 后端进程内 token blacklist，不是外部 Redis。 |
| License Manager Library | `backend/pkg/services/license/*`、`backend/go.mod` | 商业 license 解析、激活和校验，当前表现为代码库依赖，不是独立 HTTP 组件。 |

### Observability Storage And Query Backends

| 组件 | 代码 / 配置证据 | 角色 |
|---|---|---|
| ClickHouse | `clickhouse.*`、`backend/pkg/repository/clickhouse/` | 日志、Trace 明细、告警事件、拓扑、profiling、workflow record 等明细仓库。 |
| Prometheus / VictoriaMetrics | `promethues.*`、`backend/pkg/repository/prometheus/` | 时序指标查询；当前配置同时支持 Prometheus 和 VM 兼容模式。 |
| Jaeger | `jaeger.address`、`frontend/nginx.conf`、`backend/pkg/repository/jaeger/` | Trace UI 和单条 trace HTTP API。 |
| Kubernetes API | `kubernetes.*`、`backend/pkg/repository/kubernetes/` | Namespace、Pod、ConfigMap、告警规则、Vector 配置等集群元数据。 |
| DeepFlow Server / ClickHouse | `deepflow.*`、`backend/pkg/services/network/` | 网络 pod map 和 span segments；可配置独立 DeepFlow ClickHouse，也可复用 APO ClickHouse。 |
| Grafana | `frontend/grafana-deploy.yaml`、`frontend/nginx.conf` | Dashboard iframe 和 `/grafana/` 子路径展示。 |
| VMAlert / AlertManager API | `frontend/nginx.conf`、`frontend/src/core/api/alerts.js` | `/alert/api/v1/rules` 和 `/alert/-/reload` 等告警规则状态/重载边界。 |

### Data Collection Components

| 组件 | 代码 / 配置证据 | 角色 |
|---|---|---|
| OneAgent Aggregation | `one-agent/README.md`、`backend/static/integration-tmpl/` | 采集侧聚合入口和安装资产。 |
| eBPF Agent / Preload | `apoOneAgent`、`.gitmodules` `one-agent/preload` | eBPF 与 preload 注入相关采集能力。 |
| node-agent | `.gitmodules`、`apoNodeAgent` | 网络延迟、进程状态等节点侧采集。 |
| ilogtail | `.gitmodules`、`apoIlogtail` | 故障现场日志和日志链路采集。 |
| Grafana Alloy | `grafanaAlloy` | 指标采集与转发。 |
| Grafana Beyla | `grafanaBeyla` | 非侵入式 Go 应用 trace 采集。 |
| OpenTelemetry Collector | `apo-otel-collector`、`apoOtelCollectorAgent` | trace / log / metric 汇聚、转发和网关。 |
| Odigos / odiglet | `.gitmodules`、`odigos`、`odiglet` | Kubernetes 自动注入和 OpenTelemetry instrumentation 管理。 |
| apo-nginx-proxy | `apo-nginx-proxy` | 采集侧代理入口，把 agent 流量转发到 APO collector。 |
| originx-gc-agent | `originxGcAgent` | 商业采集侧组件，具体语义需结合外部组件仓继续确认。 |
| profileAgent | `apo-one-agent-values-minimal.yaml.tmpl`、`cfg.env.tmpl` | profiling 相关采集组件开关，当前模板中可禁用。 |
| vmagent | `cfg.env.tmpl` | Docker Compose 场景下的 VM 指标采集/转发组件开关。 |

### AI, Diagnosis And External Interaction

| 组件 | 代码 / 配置证据 | 角色 |
|---|---|---|
| AgentSRE Server | `AgentSRE: server.py` | 智能分析服务，提供 `/api/v1/srechat`、conversation、findings、model-setting、app-config 等接口。 |
| AgentSRE DB | AgentSRE `database.*` | 保存会话、消息、模型设置、app_config、findings 运行时数据。 |
| LLM Provider | AgentSRE `llm.*`、model setting controller | OpenAI-compatible / DashScope / Anthropic / DeepSeek 等模型调用边界。 |
| Grafana MCP | AgentSRE `grafana_mcp.*` | 通过 `mcp-grafana` 查询 Grafana / Prometheus / Loki 等工具能力。 |
| Slack WebSocket Gateway | AgentSRE `websocket.*`、`bot_websocket_client.py` | Slack bot 事件接入和双向消息通道。 |
| RAG Retrieval Service | AgentSRE `rag_config.*` | 外部知识检索入口。 |
| Braintrust | AgentSRE `braintrust.*` | LLM 调用、成本和实验追踪。 |
| Dify | APO `dify.*`、`/dify/console/api/*` | workflow 运行、告警检查、事件分类和 Dify console 边界。 |
| `/ai/api/*` Boundary | `frontend/src/pro/api/*`、AgentSRE routers | 前端 Agent / conversation / model-setting / app-config 的外部 AI 服务边界。 |
| `/originx/api/*` Boundary | `frontend/src/pro/api/cause.ts` | Cause report / mutated trace / flow 分析边界，当前主 APO 后端未注册同名路由。 |

### Dataplane Provider Ecosystem

| 组件 | 代码 / 配置证据 | 角色 |
|---|---|---|
| apo-dataplane | `apo-dataplane` | provider manager 和统一数据查询 API。 |
| APO Provider | apo-dataplane `pkg/datasource/apo` | 通过 APO Prometheus、ClickHouse、Jaeger、APO Backend 提供默认观测能力。 |
| External Providers | apo-dataplane `providerFactories` | 支持 Datadog、ARMS、Elastic、SkyWalking、Tempo、Signoz、Jaeger、Sofa、Loki、NginxLog、PrometheusMetric、GrafanaPrometheus、ExternalAPI 等。 |

### Config-Only Or Historical Signals

| 项 | 判断 |
|---|---|
| `auth_provider` / OIDC | 只在 `backend/config/apo-dev.yaml` 中出现，当前 `backend/config/config.go` 和 router 未观察到对应启用代码，不应按已实现运行组件写入主链路。 |
| `multi_tenant` | 只在 dev 配置中出现，当前未观察到主配置结构和 route/service 使用。 |
| `incident` | 只在 dev 配置中出现，当前未观察到主配置结构和 route/service 使用。 |

## High-Level Data Flow
整体链路可以分成采集、存储、查询、分析、展示五步：

```text
应用 / Node / Kubernetes / 外部告警源
  -> OneAgent / 外部 Provider / AlertManager
  -> ClickHouse / Prometheus or VictoriaMetrics / Jaeger / Kubernetes / DB
  -> APO Backend repositories / apo-dataplane provider APIs
  -> APO Frontend pages / AgentSRE tools
  -> 图表、表格、日志、Trace、证据树、诊断报告
```

更具体一点：

```text
Trace / Log / Metric 采集侧
  -> OpenTelemetry Collector / Alloy / ilogtail / Beyla / eBPF / node-agent
  -> ClickHouse + Prometheus/VictoriaMetrics + Jaeger
  -> APO Backend / apo-dataplane
  -> 服务观测、日志检索、Trace 检索、告警分析、AgentSRE
```

当前 APO 后端初始化时会连接这些主要数据访问对象：

| 后端对象 | 配置入口 | 数据来源类型 | 典型用途 |
|---|---|---|---|
| `database.Repo` | `database.*` | SQLite / MySQL / Postgres | 用户、角色、权限、菜单、数据组、接入配置、dataplane 配置、receiver、license、logo 等控制面数据。 |
| `clickhouse.Repo` | `clickhouse.*` | ClickHouse | 日志、trace、K8s 事件、告警事件、拓扑、profiling、flame graph、workflow record 等明细数据。 |
| `prometheus.Repo` | `promethues.*` | Prometheus / VictoriaMetrics | RED、资源、实例、PQL、指标查询、告警规则候选指标等时序数据。 |
| `kubernetes.Repo` | `kubernetes.*` | Kubernetes API | namespace、pod、告警规则 ConfigMap、AlertManager ConfigMap、Vector ConfigMap。 |
| `jaeger.JaegerRepo` | `jaeger.address` | Jaeger HTTP API / UI | 单条 trace 详情、Jaeger iframe 页面。 |
| `polarisanalyzer.Repo` | 内部 client | Polaris Analyzer | 下游相关性、贡献度、推断分析。 |
| `dify.DifyRepo` | `dify.*` | Dify workflow API | 告警检查、workflow 运行记录和自动分析。 |
| `dataplane.DataplaneRepo` | `dataplane.address` | apo-dataplane + 本地 CH/DB | 服务、实例、拓扑、RED、数据源健康、服务名规则。 |

## Data Source Types

### 1. Control Plane Data
控制面数据是“系统自己维护的账本”，一般存在关系型数据库里。

| 数据 | 从哪来 | APO 中怎么用 |
|---|---|---|
| 用户、团队、角色、权限 | `database.Repo` | 登录、菜单权限、API 权限、组织管理。 |
| feature、menu、router | 初始化数据 + DB | 决定前端能看到哪些菜单和页面。 |
| license、logo、deploy mode | 商业后端包 + DB / 环境配置 | 控制商业能力是否可用、平台展示信息。 |
| DataGroup / 数据范围 | DB + Prometheus + ClickHouse + Kubernetes | 限定用户能查询哪些 cluster、namespace、service、实例。 |
| integration cluster 配置 | DB | 保存 trace / metric / log 接入配置和安装参数。 |
| dataplane 配置 | DB | 保存外部数据源类型、连接参数、服务名规则、自定义拓扑。 |
| 日志解析规则 | DB + Vector ConfigMap | 控制日志如何路由、解析和生成结构化表。 |
| 告警接入、schema、enrich 规则 | DB + ClickHouse | 控制外部告警如何进入 APO 并补充标签。 |
| receiver、静默、quick metric、target tags | DB / Kubernetes ConfigMap | 告警通知、告警规则和事件展示。 |
| AgentSRE 对话、消息、findings | AgentSRE DB | 保存智能分析过程、证据节点、工具调用和最终结果。 |

通俗解释：这些不是机器运行时产生的观测数据，而是“平台怎么管理、谁能看什么、接了哪些数据源”的配置和业务记录。

### 2. Metric Data
Metric 是“随时间变化的数字”，适合画曲线和做异常检测。

| 类型 | 例子 | 主要来源 | APO / AgentSRE 用途 |
|---|---|---|---|
| RED 指标 | 请求数、错误数、耗时 | Prometheus / VictoriaMetrics，典型指标包括 `kindling_span_trace_duration_nanoseconds_count` 和 `kindling_span_trace_duration_nanoseconds_sum` | 服务列表、RED 趋势、错误率、平均延迟、告警影响面。 |
| 容器指标 | CPU、内存、磁盘、网络、文件描述符 | Prometheus / VictoriaMetrics | 实例详情、资源瓶颈分析、AgentSRE resource analysis。 |
| 主机指标 | CPU busy、内存、磁盘、网络、打开文件数 | Prometheus / VictoriaMetrics | Node 级故障分析、宿主机资源判断。 |
| 进程指标 | 进程 FD、进程状态 | Prometheus / VictoriaMetrics / node-agent | 进程异常和资源耗尽判断。 |
| eBPF / Polaris 指标 | epoll、file、net、cpu、runq、futex、thread polaris | OneAgent / eBPF / Polaris -> Prometheus / VictoriaMetrics | 线程级耗时、锁等待、网络等待、CPU 等底层原因分析。 |
| 数据库 / MQ / 外部调用指标 | DB duration bucket、MQ duration bucket、external duration bucket | Prometheus / VictoriaMetrics | 下游依赖慢、数据库慢、外部接口慢判断。 |
| 自定义通用指标 | 用户选择的 metric name + labels | apo-dataplane provider 或 APO `/api/metric/query` | AgentSRE 和前端通用曲线查询。 |

APO 后台有两类指标入口：

| 入口 | 面向谁 | 后端来源 |
|---|---|---|
| `/api/metric/list`、`/api/metric/query`、`/api/metric/queryPods` | 通用指标查询、Pod 元数据辅助查询 | `prometheus.Repo`。 |
| `/api/service/*`、`/api/dataplane/redcharts`、`/api/dataplane/redcharts/list` | 服务观测、RED、实例、拓扑相关页面 | `prometheus.Repo` + `clickhouse.Repo` + `dataplaneRepo`。 |
| apo-dataplane `/datasource/queryMetrics` | AgentSRE 和外部统一查询 | Provider 的 `Metrics` 能力，可能落到 APO、Prometheus、Grafana Prometheus、Datadog、Elastic、Signoz 等。 |

通俗解释：Metric 像体温计、血压计和心电图，用来判断“有没有异常、什么时候开始异常、异常有多严重”。

### 3. Log Data
Log 是“程序自己写下来的文字记录”，适合解释具体报错原因。

| 类型 | 例子 | 主要来源 | 存储 / 查询 |
|---|---|---|---|
| 原始日志 | 容器 stdout / stderr、应用文本日志 | OneAgent / ilogtail / 日志采集器 | ClickHouse 默认日志表，如 `raw_logs`、`all_logs`。 |
| 结构化日志 | 解析出 level、service、pod、namespace、字段后的日志 | Vector 解析规则 + ClickHouse | `logs_<parseName>` 等解析后表。 |
| 外部日志表 | 用户登记的其他日志表 | 外部 ClickHouse 表或类似来源 | `otherlogtable` 元数据 + ClickHouse 查询。 |
| 故障现场日志 | 围绕 traceId、service、instance、时间窗口的日志片段 | ClickHouse | `/logs/fault-site`、`/api/log/fault/*`。 |
| 上下文日志 | 某条日志前后若干行 | ClickHouse | `/api/log/context`。 |
| AgentSRE 日志证据 | pod 维度日志、错误模板、错误数量 | apo-dataplane `/datasource/queryLogs` | AgentSRE `LogAnalysisTool` 用 Drain3 聚类错误模板。 |

APO 日志入口：

| 入口 | 查询内容 | 来源 |
|---|---|---|
| `/api/log/query` | 全量日志分页查询 | ClickHouse。 |
| `/api/log/chart` | 日志数量直方图 / 趋势 | ClickHouse。 |
| `/api/log/index` | 日志字段索引 | ClickHouse。 |
| `/api/log/context` | 单条日志上下文 | ClickHouse。 |
| `/api/log/rule/*` | 日志解析规则 | DB + Kubernetes Vector ConfigMap。 |
| apo-dataplane `/datasource/queryLogs` | provider 统一日志查询 | `LogProvider`，支持 APO、Datadog、Elastic、Signoz、Loki、ExternalAPI 等。 |

需要特别区分：

- `backend` 自己的 `logger.file_path` 是 APO 后台进程日志，只记录 APO 后台运行情况。
- `/api/log/*` 查的是被观测应用、容器或外部日志表里的业务日志。

通俗解释：Log 像机器的日记。Metric 告诉你“发烧了”，Log 告诉你“它自己说为什么不舒服”。

### 4. Trace Data
Trace 是“一次请求的完整路线图”，适合看请求经过哪些服务、卡在哪一段。

| 类型 | 核心字段 | 主要来源 | APO / AgentSRE 用途 |
|---|---|---|---|
| span 列表 | `traceId`、`spanId`、`parentSpanId`、service、operation、duration、error | OneAgent / OpenTelemetry / Beyla / Jaeger / ClickHouse | `/trace/fault-site` 列表、AgentSRE slow/error trace 分析。 |
| 完整 trace | 一个 `traceId` 下的所有 span | Jaeger 或 provider trace backend | Jaeger 详情、调用链路分析。 |
| trace-derived topology | 从 trace 推导父子服务关系 | apo-dataplane TraceProvider | 外部 provider 没有直接拓扑时用于生成拓扑。 |
| profiling event | on/off CPU、线程执行消耗 | ClickHouse `profiling_event` | `/api/trace/onoffcpu`。 |
| flame graph | flamebearer 数据 | ClickHouse `flame_graph` | `/api/trace/flame`、`/api/trace/flame/process`。 |

APO Trace 入口：

| 入口 | 查询内容 | 来源 |
|---|---|---|
| `/api/trace/pagelist` | 故障现场 trace 列表 | ClickHouse `span_trace`。 |
| `/api/trace/pagelist/filters` | trace 筛选项 | ClickHouse。 |
| `/api/trace/info` | 单条 trace 详情 | Jaeger HTTP API。 |
| `/trace/full` | Jaeger 搜索 UI | 前端 iframe 到 Jaeger。 |
| `/api/trace/onoffcpu` | on/off CPU profiling | ClickHouse `profiling_event`。 |
| `/api/trace/flame`、`/api/trace/flame/process` | flame graph | ClickHouse `flame_graph`。 |
| apo-dataplane `/datasource/querySpans`、`/datasource/queryTraces` | provider 统一 span / trace 查询 | APO、ARMS、Datadog、Elastic、SkyWalking、Tempo、Signoz、Jaeger、Sofa、ExternalAPI 等。 |

通俗解释：Trace 像请求路线图。它能告诉你“请求从 A 到 B 到 C，最后卡在 C，耗时 2 秒”。

### 5. Kubernetes Resource And Event Data
Kubernetes 数据是“集群现场状态”。

| 数据 | 来源 | APO 用途 |
|---|---|---|
| namespace 列表 / 详情 | Kubernetes API | 筛选器、数据范围、K8s 页面。 |
| pod 列表 / 详情 | Kubernetes API | 实例详情、pod 元数据、AgentSRE pod fallback。 |
| ConfigMap | Kubernetes API | 告警规则、AlertManager 配置、Vector 日志解析配置。 |
| K8s event | Kubernetes API 或采集后入 ClickHouse | service 详情里的 K8s 事件、事件数量、故障辅助证据。 |
| Pod 状态 | Kubernetes API + Prometheus + ClickHouse | Ready、重启、所在 Node、namespace、podName 等实例信息。 |

APO 入口：

| 入口 | 查询内容 | 来源 |
|---|---|---|
| `/api/k8s/namespaces`、`/api/k8s/namespace/info` | namespace | Kubernetes API。 |
| `/api/k8s/pods`、`/api/k8s/pod/info` | pod | Kubernetes API。 |
| `/api/service/k8s/events/count` | service 相关 K8s event 计数 | ClickHouse / Kubernetes 事件读模型。 |
| `/api/service/instance*`、`/api/dataplane/instances` | 服务实例与 pod 归属 | Prometheus + ClickHouse + Kubernetes + fallback。 |

通俗解释：K8s 资源像现场设备清单，K8s event 像值班记录。

### 6. Topology And Network Data
拓扑是“谁调用谁”的关系图。

| 数据 | 来源 | APO 用途 |
|---|---|---|
| 服务拓扑 | ClickHouse service relationship / apo-dataplane provider | 服务地图、上下游、影响面。 |
| 自定义拓扑 | DB 配置 | 人工补充或修正服务关系。 |
| trace-derived topology | TraceProvider 解析 trace | 外部 trace 系统可推导服务依赖。 |
| endpoint 关系 | ClickHouse + Prometheus | service endpoint 详情、入口/依赖关系。 |
| network pod map / segments | DeepFlow ClickHouse 或默认 ClickHouse + Prometheus + K8s | 网络调用、span segment、pod 关系。 |

入口：

| 入口 | 查询内容 | 来源 |
|---|---|---|
| `/api/service/topology`、`/api/service/relation` | 服务与 endpoint 关系 | ClickHouse + Prometheus。 |
| `/api/dataplane/topology` | provider 统一拓扑 | apo-dataplane / ClickHouse / 外部 provider。 |
| `/api/network/podmap`、`/api/network/segments` | 网络拓扑和 span segments | DeepFlow ClickHouse 或 APO ClickHouse。 |

通俗解释：Topology 像服务关系图，告诉你 A 服务上游是谁、下游是谁、故障会影响谁。

### 7. Alert And Incident Data
告警数据是“报警器响了”的入口。

| 数据 | 来源 | APO 用途 |
|---|---|---|
| 告警事件 | AlertManager webhook、外部 alert source、APO receiver | 告警列表、详情、分类、分析入口。 |
| 告警规则 | Kubernetes ConfigMap / AlertManager / VMAlert 风格接口 / DB 辅助数据 | 规则配置、规则状态、PromQL 编辑。 |
| 告警标签与 schema | DB + ClickHouse | 过滤、展示、enrich。 |
| 告警报告 | ClickHouse / workflow record / report 接口 | 事件报告、历史分析记录。 |
| 通知记录 | ClickHouse + receiver 配置 | DingTalk / webhook 等通知链路。 |
| 静默配置 | DB 或 AlertManager config 边界 | 告警静默。 |

入口：

| 入口 | 查询内容 | 来源 |
|---|---|---|
| `/api/alerts/event/list`、`/api/alerts/event/detail` | 告警事件 | ClickHouse。 |
| `/api/alerts/inputs/alertmanager` | AlertManager 输入兼容入口 | alert input service。 |
| `/api/alertinput/*` | 告警源、schema、enrich、cluster | DB + ClickHouse + Dify。 |
| `/api/alerts/rule/*` | 告警规则 | Kubernetes ConfigMap / DB / Prometheus。 |
| `/api/alerts/events/report/*` | 事件报告 | ClickHouse。 |
| `/api/alerts/alert/events`、`/api/alerts/event/impact` 等 | 告警分析 | ClickHouse + Prometheus + PolarisAnalyzer + Kubernetes。 |

通俗解释：Alert 不是根因，只是“报警铃”。系统还要查 metric、log、trace、event、topology 才能解释为什么报警。

### 8. Diagnosis, Workflow And AgentSRE Data
分析数据是“看完现场数据后产出的结论”。

| 数据 | 来源 | 用途 |
|---|---|---|
| Dify workflow record | Dify workflow API + ClickHouse | 告警自动检查、workflow 历史、分类和报告。 |
| Polaris 结果 | PolarisAnalyzer | 下游相关性、贡献度、异常推断。 |
| AgentSRE conversations | AgentSRE DB | 对话历史、消息、用户问题。 |
| AgentSRE findings | AgentSRE DB + 工具调用结果 | 证据树、工具输出、根因结论、展示数据。 |
| AgentSRE tool displays | AgentSRE 运行时事件 | 前端展示日志、trace、metric、拓扑等工具证据。 |

AgentSRE 主要读取这些接口：

| AgentSRE 工具 | 查询入口 | 拿到什么 |
|---|---|---|
| `LogAnalysisTool` | APO `/api/dataplane/instances` + apo-dataplane `/datasource/queryLogs` | pod 日志、错误数量、错误模板。 |
| `TracesAnalysisTool` | apo-dataplane `/datasource/queryTraces`、`/datasource/querySpans` | slow / error span、历史 span 对比。 |
| `ResourceQueryTool` | apo-dataplane `/datasource/queryMetrics` | 容器、主机、网络、磁盘、FD 等资源曲线。 |
| `ServiceTpsTool` | APO `/api/dataplane/redcharts` | 请求量、错误、延迟。 |
| topology helpers | APO `/api/dataplane/topology` | 服务上下游关系。 |
| instance helpers | APO `/api/dataplane/instances` + `/api/metric/queryPods` fallback | service 对应 pod、namespace、node、pid。 |
| eBPF helpers | APO `/api/metric/query` + VictoriaMetrics query_range | thread polaris、RTT、runqueue、futex 等底层指标。 |

通俗解释：AgentSRE 本身不是采集系统，它更像专家：先问 APO / Dataplane / VM 要资料，再组织成诊断报告。

## apo-dataplane Provider Model
apo-dataplane 的核心价值是把不同数据源翻译成同一套能力。

### Capabilities

| Capability | 中文含义 | 典型返回 |
|---|---|---|
| `RedMetric` | 服务 RED 指标 | service list、请求数、错误数、耗时。 |
| `ServiceInstance` | 服务实例 | pod、namespace、node、container、pid、ip。 |
| `Topology` | 拓扑 | parent service、child service、类型。 |
| `Log` | 日志 | timestamp、level、message、attributes。 |
| `Trace` | 链路 | trace、span、duration、error、attributes。 |
| `Metrics` | 通用指标 | title、unit、timeseries、labels、chartData。 |

### Provider Types
当前代码中的 provider 类型包括：

| Provider | 主要能力 |
|---|---|
| `apo` | RED、实例、拓扑、Trace、Log、Metrics。 |
| `arms` | RED、实例、拓扑、Trace。 |
| `datadog` | RED、实例、拓扑、Trace、Log、Metrics。 |
| `elastic` | Log、Trace、Metrics、Topology、实例、RED。 |
| `skywalking` | RED、实例、拓扑、Trace。 |
| `tempo` | RED、实例、拓扑、Trace。 |
| `signoz` | RED、实例、拓扑、Trace、Log、Metrics。 |
| `jaeger` | Trace、Topology。 |
| `sofa` | Trace、Topology。 |
| `loki` | Log。 |
| `nginxLog` | RED。 |
| `prometheusmetric` | Metrics。 |
| `grafanaPrometheus` | Metrics。 |
| `externalapi` | Topology、Trace、Log、Metrics。 |

### Query APIs

| apo-dataplane API | 统一能力 | 下游可能来源 |
|---|---|---|
| `/datasource/queryProviderList` | provider 清单 | provider manager。 |
| `/datasource/checkProvider` | 数据源连通性检查 | provider checker。 |
| `/datasource/queryServiceList` | `RedMetric` | APO Prometheus、Datadog、ARMS 等。 |
| `/datasource/queryServiceInstances` | `ServiceInstance` | APO Prometheus、外部 APM / 指标源等。 |
| `/datasource/queryServiceRedCharts` | `RedMetric` | 时序指标源。 |
| `/datasource/queryServiceRedValue` | `RedMetric` | 时序指标源。 |
| `/datasource/queryTopology` | `Topology` | ClickHouse、trace-derived topology、外部 APM。 |
| `/datasource/querySpans` | `Trace` | Jaeger、Tempo、SkyWalking、APO trace 等。 |
| `/datasource/queryTraces` | `Trace` | Jaeger、Tempo、SkyWalking、APO trace 等。 |
| `/datasource/queryLogs` | `Log` | APO ClickHouse、Loki、Elastic、Datadog 等。 |
| `/datasource/queryMetrics` | `Metrics` | APO `/api/metric/query`、Prometheus、Grafana Prometheus、Datadog 等。 |

## Data Category Cheat Sheet

| 名称 | 看什么 | 最常见来源 | 典型问题 |
|---|---|---|---|
| Metric | 数字曲线 | Prometheus / VictoriaMetrics | CPU 高不高？错误率何时升高？延迟是不是变慢？ |
| Log | 文本日志 | ClickHouse / Loki / Elastic | 程序报了什么错？有没有 exception / timeout / failed？ |
| Trace | 请求路线 | ClickHouse / Jaeger / Tempo / SkyWalking | 请求卡在哪个服务？哪个 span 最慢？哪里报错？ |
| Event | 平台事件 | Kubernetes API / ClickHouse | Pod 有没有重启？有没有 OOMKilled？镜像是否拉取失败？ |
| Alert | 报警记录 | AlertManager / alert input / ClickHouse | 哪个告警触发？标签是什么？影响哪个服务？ |
| Topology | 服务关系 | ClickHouse / trace / external APM | 上游是谁？下游是谁？故障影响范围多大？ |
| Resource | 容器 / 主机资源 | Prometheus / VictoriaMetrics | CPU、内存、磁盘、网络是否异常？ |
| Profiling | 线程 / 函数耗时 | ClickHouse / eBPF / Polaris | 卡在 CPU、IO、网络、锁，还是 runqueue？ |
| Config | 平台配置 | DB / ConfigMap / YAML | 接了哪些集群？哪些数据源？谁有权限？ |
| Findings | 诊断证据 | AgentSRE DB + 工具输出 | 根因是什么？证据有哪些？建议怎么处理？ |

## Example: One Alert Analysis
以 `order-service error rate high` 为例：

1. 告警进入 APO：
   - 来源可能是 AlertManager webhook、外部 alert source 或 APO receiver。
   - 告警事件落到 ClickHouse，标签里可能包含 `cluster`、`namespace`、`service`、`severity`、`alertname`。
2. APO 后台展示告警：
   - `/api/alerts/event/list` 查事件列表。
   - `/api/alerts/event/detail` 查事件详情。
   - DB 用来补充 target tags、receiver、静默、权限等控制面信息。
3. 系统定位服务和实例：
   - `/api/dataplane/instances` 或 `/api/service/instances` 找 `order-service` 对应的 pod。
   - 如果实例目录缺数据，AgentSRE 可能通过 `/api/metric/query` 或 `/api/metric/queryPods` 用 pod 名兜底发现。
4. 查 Metric：
   - RED 看错误率、请求量、延迟。
   - 资源指标看 CPU、内存、网络、磁盘。
   - eBPF / Polaris 看线程是否卡在网络、文件、锁、runqueue。
5. 查 Log：
   - 按 pod、namespace、时间窗口查日志。
   - AgentSRE 会统计 `error`、`exception`、`failed` 等错误模板。
6. 查 Trace：
   - 查 error trace 或 slow trace。
   - 找到失败 span、最慢 span、下游 DB / HTTP / MQ 调用。
7. 查 Event：
   - 看 Pod 是否重启、OOM、调度失败、Node 异常。
8. 查 Topology：
   - 看 `order-service` 上下游。
   - 判断是自身问题、上游流量问题，还是下游依赖问题。
9. AgentSRE 输出 findings：
   - 把 metric、log、trace、event、topology 组织成证据节点。
   - 生成根因候选、理由和建议动作。

## Common Misunderstandings

| 误解 | 正确理解 |
|---|---|
| APO 后台自己生成所有数据 | APO 后台主要负责查询、聚合和整理，原始数据大多来自采集器、K8s、时序库、ClickHouse、Jaeger 或外部 provider。 |
| Alert 就是根因 | Alert 只是入口；根因要结合 metric、log、trace、event、topology 判断。 |
| Log 和 Trace 是一回事 | Log 是文本日记；Trace 是一次请求的路径。两者可通过 `traceId` 关联，但不是同一种数据。 |
| Metric 能解释所有问题 | Metric 适合看趋势和异常点，但具体错误原因通常要看 log / trace / event。 |
| Dataplane 是一个新存储 | apo-dataplane 主要是适配层，不等同于数据库；它背后仍然查 APO、Prometheus、ClickHouse、Jaeger、Loki、Datadog 等。 |
| AgentSRE 直接连所有底层系统 | AgentSRE 当前主要通过 APO 后台、apo-dataplane、VictoriaMetrics 和自身 DB 拿数据。 |
| `logger.file_path` 是业务日志来源 | 这是 APO 后台自己的运行日志，不是 `/api/log/query` 查的业务日志。 |

## Source Map
后续查代码时可以按下面路径定位：

| 主题 | 关键路径 |
|---|---|
| APO 后台初始化数据源 | `backend/pkg/router/router.go` |
| APO API 路由 | `backend/pkg/router/router_api.go` |
| APO 配置结构 | `backend/config/config.go` |
| APO 默认配置示例 | `backend/config/apo.yml` |
| 后端 ClickHouse 查询 | `backend/pkg/repository/clickhouse/` |
| 后端 Prometheus / VM 查询 | `backend/pkg/repository/prometheus/` |
| 后端 Kubernetes 查询 | `backend/pkg/repository/kubernetes/` |
| 后端 Jaeger 查询 | `backend/pkg/repository/jaeger/` |
| 后端 dataplane 聚合 | `backend/pkg/repository/dataplane/`、`backend/pkg/services/dataplane/` |
| 日志模块文档 | `docs/openspec-system/modules/logs-exploration-and-parsing/module_overview.md` |
| Trace 模块文档 | `docs/openspec-system/modules/trace-exploration-and-profiling/module_overview.md` |
| 服务观测模块文档 | `docs/openspec-system/modules/service-observability/module_overview.md` |
| 告警模块文档 | `docs/openspec-system/modules/alert-rules-and-incident-events/module_overview.md` |
| OneAgent 采集说明 | `one-agent/README.md` |
| apo-dataplane provider 机制 | `apo-dataplane: README.md` |
| apo-dataplane HTTP API | `apo-dataplane: pkg/httpserver/http_server.go` |
| apo-dataplane provider 能力 | `apo-dataplane: pkg/datasource/provider/` |
| AgentSRE 工具入口 | `AgentSRE: api/tools/` |
| AgentSRE APO / Dataplane 调用 | `AgentSRE: api/tools/apo/apo_dataplane.py` |

## Notes And Boundaries

- 本文档描述“已在当前代码和文档中看到的读路径与集成边界”，不是承诺每个部署环境都启用了所有 provider。
- 具体数据源地址由运行环境配置决定，本文档只记录配置键和职责，不记录密钥。
- `promethues.storage` 可配置为 Prometheus 或 VictoriaMetrics 兼容模式，代码中名称沿用历史拼写。
- `deepflow.ch_address` 未配置时，网络相关查询会默认复用 APO ClickHouse 配置。
- `dataplane.address` 为空时，APO 后台不会初始化 dataplane repository。
- 外部边界如 Dify、Grafana、AlertManager、VMAlert、Datadog、ARMS、SkyWalking、Tempo、Loki、Elastic 的真实可用性取决于实际部署和 provider 配置。
