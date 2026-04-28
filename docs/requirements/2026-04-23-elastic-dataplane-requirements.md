# Elastic / ELK Dataplane 接入需求文档

## 一、文档目的

本文档用于定义商业版 APO 在 `/dataplane` 页面接入 `Elastic / ELK` Log-only 数据平面的当前需求效果、用户可见行为、能力边界与验收标准。

本次需求目标已收窄为 **Elastic Log-only 接入**：先把 `Elastic` 接入为 APO dataplane 体系中的正式类型，让 AgentSRE 可以通过现有 dataplane 日志接口消费 Elastic 日志。

完整 Elastic Observability / APM 能力，包括 `Trace`、`Metrics`、`Topology`、`ServiceInstance`、`RedMetric`，不作为本次验收目标，后续单独立项。

本文档面向产品、研发、测试和评审人员阅读，重点定义“最终要做成什么样”，不展开到具体代码任务拆解。

## 二、需求摘要

本期目标体验如下：

- 用户可以在 `/dataplane` 页面中看到新的 `Elastic` 数据平面类型
- `Elastic` 作为与 `Datadog`、`SigNoz` 同级的数据平面类型存在，而不是旧版接入中心里的孤立日志选项
- 当前主流程支持 `Elastic Cloud` 和自建 `Elastic Stack` 的 Elasticsearch endpoint
- 当前主流程只开放 `Log` capability
- 用户只需填写名称、URL 和 API Key，系统默认使用 `authType=apiKey`、`logIndexPattern=logs-*`
- 测试连接只验证 Elastic Log，不因为缺少 APM / metrics / topology 数据而失败
- 接入完成后，AgentSRE 可以继续通过 `/datasource/queryLogs` 查询 Elastic 日志

## 三、背景现状

### 3.1 当前 dataplane 主路径

当前商业版 APO 已有一套独立的 dataplane 治理入口：

- 前端主入口是 `/dataplane`
- 左侧展示可接入数据平面类型
- 右侧展示已配置数据源
- 表单支持“测试连接”和“连接”
- Agent 数据状态和 capability 汇总也依赖这套 dataplane 体系

当前这套体系里已经存在 `datadog`、`signoz`、`loki`、`jaeger`、`skywalking`、`arms` 等类型。

### 3.2 当前 ELK 的位置不正确

仓内现状表明：

- 旧版 `IntegrationCenter` 元数据里存在 `elastic` / `ELK` 文案
- 但当前正式 `/dataplane` 类型列表中并没有 `elastic`
- `apo-dataplane` 也没有现成的 `elastic` provider

因此，当前“ELK 已经支持”的结论并不成立。它只在旧版接入常量里出现，并没有真正进入当前 dataplane 主流程。

### 3.3 当前先不处理 Elastic 的 full APM 形态

当前 `/dataplane` 表单保存逻辑默认会把某个类型的 `capabilitySpec` 当成该实例的完整启用能力，这对于 `datadog`、`signoz` 这类天然全能力 provider 没问题，但不适用于 Elastic：

- 有些用户只有 Elasticsearch 日志索引，没有 APM / metrics / topology 数据
- 有些用户接入了 Elastic Observability / APM，可以提供 Trace、Service Map、APM metrics 等更多能力

本次只交付第一类日志场景。第二类 full APM 场景不作为本次验收目标。

## 四、核心问题

当前 APO 缺少一个正式可用的 `Elastic` 数据平面类型，导致：

- 已经把日志沉淀在 Elastic 体系中的用户，无法通过当前 `/dataplane` 主入口完成接入
- “只有日志索引”的 Elastic 用户无法低成本接入日志能力
- AgentSRE 无法通过 dataplane 主链路消费 Elastic 中已有的服务日志

## 五、目标与范围

### 5.1 一期目标

1. 在 `/dataplane` 中新增 `Elastic` 数据平面类型
2. 让 `Elastic` 作为正式 dataplane provider 接入 APO，而不是停留在旧版接入元数据
3. 支持 `Elastic Cloud` 和自建 `Elastic Stack` 的 Elasticsearch endpoint
4. 支持只有日志索引的 Log-only 上游形态
5. 默认按 `ECS / OTel / Kubernetes` 常见日志字段约定工作
6. 测试连接逻辑只校验当前 Log-only 实例，不校验未启用的 APM capability
7. 接入后 AgentSRE 可通过现有 dataplane 日志接口查询 Elastic 日志

### 5.2 非目标

以下内容不属于本期范围：

- 在一期支持完全任意、无限制的字段映射系统
- 不在本次开放字段映射 UI
- 不在本次开放 Basic Auth UI
- 不在本次支持 `Trace`、`Metrics`、`Topology`、`ServiceInstance`、`RedMetric`
- 不在本次实现完整实例级 capability 选择、coverage 聚合治理和结构化 check result UI
- 在前台嵌入 Kibana 页面或跳转式 Kibana 深链治理
- 一期实现日志规则管理、外部表登记等旧 `/logs/full` 体系的能力迁移
- 一期强制支持所有历史 Elastic 版本差异
- 一期实现 Alert / Incident 接入
- 一期实现与 Elastic ML 异常检测的深度联动
- 一期做多租户权限模型重构

## 六、目标用户与典型场景

### 6.1 目标用户

- 已经使用 `Elastic Cloud` 作为统一观测存储的团队
- 自建 `Elastic Stack`，把日志、链路或指标写入 Elasticsearch 的团队
- 当前只想先把日志接入 APO 的团队
- 已经使用 Elastic APM，希望在 APO 中统一看到 Trace / Topology / RED / Metrics 的团队不属于本次目标用户，应进入后续 full APM change

### 6.2 典型场景

#### 场景 A：只有日志索引

用户只有 Elasticsearch 日志索引，日志字段基本遵循 ECS，暂时没有 APM 数据。

最终效果应为：

- 用户可以创建一个 `Elastic` 数据平面
- 系统保存时只启用 `Log`
- 测试连接可以通过
- 该实例在已配置列表中展示为 `Log`
- 该实例不会在本次验收中声明支持 `Trace`、`Topology`、`ServiceInstance`、`RedMetric`、`Metrics`

#### 场景 B：完整 Observability / APM

该场景不属于本次验收范围，后续单独立项。

## 七、最终实现效果

### 7.1 用户可见效果

最终上线后，用户在 `/dataplane` 页面会获得以下体验：

- 左侧数据平面类型列表出现 `Elastic`
- 打开 `Elastic` 配置抽屉后，只需要填写名称、URL 和 API Key
- 系统默认保存为 `Log` capability，默认日志索引为 `logs-*`
- 测试连接只验证 Elastic Log
- 保存后，右侧“已配置数据源”会出现该 Elastic 实例，并展示 capability 为 `Log`
- 配置指引说明 Elastic Cloud、自建 Elastic、API Key、默认 `logs-*` 和 AgentSRE 需要的 pod / namespace 日志字段

### 7.2 系统效果

最终系统行为必须满足：

- APO 主仓可以识别 `elastic` 作为正式数据平面类型
- `apo-dataplane` 可以创建并注册 Elastic provider 实例
- Elastic provider 的 Log check 和 Log query 进入现有 dataplane 主链路
- AgentSRE 不需要改代码，继续通过 `/datasource/queryLogs` 查询日志
- 如果日志可以查询但缺少 pod / namespace 字段，测试返回 warning，不阻塞保存

## 八、功能需求

### 8.1 数据平面类型

系统必须新增一个正式的 `Elastic` dataplane type：

- 类型对外名称显示为 `Elastic`
- 允许产品或文案在说明中使用 `Elastic / ELK`
- 代码与持久化层统一使用 `elastic`
- 类型模板可以保留未来能力上限，但本次主流程只启用 `Log`

### 8.2 Log-only 配置

本次主流程要求如下：

- 用户填写名称、Elasticsearch URL 和 API Key
- 前端保存时固定 `capability=["Log"]`
- 前端保存时默认补齐 `authType=apiKey`
- 前端保存时默认补齐 `logIndexPattern=logs-*`
- Basic Auth、自定义日志索引、environment 和字段映射 UI 不属于本次范围

### 8.3 测试连接

测试连接必须满足：

- 只校验 Elastic endpoint、认证和 Log 查询能力
- 不校验 `Trace`、`Metrics`、`Topology`、`ServiceInstance`、`RedMetric`
- 当前时间窗无日志时返回 warning，不直接 hard fail
- 有日志但缺少 AgentSRE 服务过滤所需的 pod / namespace 字段时返回 warning，不阻塞保存
- 保持 APO 现有 `success/message` 兼容语义

### 8.4 日志字段约定

本次默认按 `ECS / OTel / Kubernetes` 常见日志字段约定工作，至少覆盖：

- `service.name`
- `host.name`
- `container.id`
- `kubernetes.namespace` / `kubernetes.namespace.name` / `k8s.namespace.name`
- `kubernetes.pod.name` / `k8s.pod.name`
- `log.level`
- `message` / `body`

### 8.5 AgentSRE 行为

AgentSRE 继续通过现有 dataplane 日志接口查询：

- URL：`/datasource/queryLogs`
- 过滤字段：`podName`、`k8sNamespace`
- 返回结果：现有 dataplane log response

如果 Elastic 日志缺少 pod / namespace 字段，AgentSRE 可能无法按服务稳定查到日志，这是本次的已知 warning 场景。

## 九、交互与页面需求

### 9.1 `/dataplane` 类型列表

`Elastic` 必须出现在 `/dataplane` 左侧类型列表中，并遵循现有 dataplane 类型展示方式。

### 9.2 配置表单

`Elastic` 的配置表单必须满足：

- 支持名称、URL、API Key 输入
- 支持测试连接
- 支持保存
- 不展示 full APM capability 选择
- 不展示 Trace / Metrics 索引配置
- 不展示字段映射配置

### 9.3 配置指引

配置抽屉右侧应提供 `Elastic` 接入说明，至少覆盖：

- Elastic Cloud 接入路径
- 自建 Elastic Stack 接入路径
- API Key 获取方式
- 默认 `logs-*` 日志索引说明
- AgentSRE 服务日志分析依赖的 pod / namespace 字段

## 十、跨仓职责边界

### 10.1 APO 商业仓职责

APO 商业仓应负责：

- `/dataplane` 页面类型展示
- Log-only 配置表单
- `elastic` 类型种子与升级脚本
- 测试 warning 的基本展示

### 10.2 `apo-dataplane` 仓职责

`apo-dataplane` 应负责：

- 注册 `elastic` provider
- 实现 provider 创建与检查
- 实现日志查询能力
- 支持 AgentSRE 的 pod / namespace 日志过滤

### 10.3 共享后端治理说明

本次不改 APO 共享后端候选路径。后续如果恢复完整 capability governance，再单独确认是否回 OSS 以及目标分支。

## 十一、兼容性要求

本次兼容性要求如下：

- 不破坏现有 `datadog`、`signoz`、`loki` 等已存在 dataplane 类型
- 不改变 `/api/dataplane/capabilities` 的接口语义
- 不要求旧版 `IntegrationCenter` 的 `ELK` 常量承担新主流程职责
- 不要求 AgentSRE 改代码

## 十二、验收标准

### 12.1 日志专用场景验收

当用户只有 Elasticsearch 日志索引时：

1. Given：用户填写有效 Elasticsearch URL 和 API Key
2. When：用户点击 `Test`
3. Then：系统只校验 `Log`，不校验未启用的 `Trace`、`Metrics`、`Topology`、`ServiceInstance`、`RedMetric`
4. Then：若日志查询链路成立但当前时间窗无日志，可返回 warning，不得因为未启用能力而失败
5. Then：若日志存在但缺少 pod / namespace 字段，可返回 warning，不阻塞保存
6. When：用户点击保存
7. Then：实例创建成功，已配置列表显示该 `Elastic` 实例且 capability 为 `Log`
8. Then：AgentSRE 可继续调用 `/datasource/queryLogs` 查询 Elastic 日志

### 12.2 配置体验验收

1. `Elastic` 出现在 `/dataplane` 类型列表中
2. 配置表单只要求名称、URL 和 API Key
3. 保存 payload 默认包含 `authType=apiKey`、`logIndexPattern=logs-*`、`capability=["Log"]`
4. 测试连接 warning 有可读提示
5. 已配置列表可正常显示、编辑、删除 Elastic 实例
6. 配置指引中同时包含 Elastic Cloud 与自建路径

## 十三、后续迭代方向

以下内容可以作为后续版本扩展：

- 更丰富的字段映射系统
- Basic Auth、自定义日志索引、environment 和字段映射 UI
- 完整实例级 capability 选择、coverage 聚合和结构化 check result UI
- Elastic Trace / Metrics / Topology / ServiceInstance / RedMetric 能力接入
- 自动发现索引 / data stream
- 更细粒度的权限与 API key 最小授权说明
- 与 Elastic 异常检测、告警、变更分析的深度联动
- 更精细的环境、空间、租户隔离模型

## 十四、参考信息

以下官方资料用于约束本需求中的上游事实边界：

- Elastic API 认证方式支持 Basic Auth 与 API Key Auth  
  https://www.elastic.co/docs/api/doc/elasticsearch/authentication
- Elastic APM Traces 说明  
  https://www.elastic.co/docs/solutions/observability/apm/traces
- Elastic APM Metrics 说明  
  https://www.elastic.co/docs/solutions/observability/apm/metrics/
- Elastic Service Map 说明  
  https://www.elastic.co/docs/solutions/observability/apm/service-map/
- Elastic Logs in APM / Observability 说明  
  https://www.elastic.co/docs/solutions/observability/apm/logs
- ECS `service.*` 字段说明  
  https://www.elastic.co/docs/reference/ecs/ecs-service
- ECS `host.*` 字段说明  
  https://www.elastic.co/docs/reference/ecs/ecs-host
- ECS `container.id` 字段说明  
  https://www.elastic.co/guide/en/ecs/current/ecs-container.html
