# 系统健康中心版本B影响路径数据口径

本文档面向 `/healthy-service?groupId=1&relativeTime=15m` 的“版本B 影响路径”页面，逐行说明页面信息的正式数据来源和计算方式。

当前页面草稿在 `frontend/src/pro/views/serviceHealthy/index.js` 中使用 `MOCK_INFRASTRUCTURE`、`MOCK_SERVICES`、`MOCK_LOGS`、`MOCK_ALERTS`、`HEALTH_RISK_CLUSTERS` 等本地常量。这些常量只能作为版式样例，不能作为正式数据来源。正式实现不得沿用 `/api/service/ryglight` 的 `percentScore/status/scoreDetail` 作为新版系统健康评分标准；新版健康度以“业务入口受影响程度 + 告警严重度 + RED退化 + 下游传播 + 数据完整性”为主。

## 查询上下文

| 页面输入 | 来源 | 计算/使用方式 |
| --- | --- | --- |
| 时间窗口 | URL `relativeTime=15m` 或全局时间选择器 | 转换为 `startTime/endTime` 微秒时间戳；所有接口统一使用同一窗口。 |
| 数据组 | URL `groupId=1` 或全局数据组 | 所有接口传 `groupId`。注意：旧健康页前端曾传 `dataGroupId`，后端请求结构使用 `groupId`，新版必须统一为 `groupId`。 |
| 搜索关键词 | 页面输入框“过滤问题 / 对象” | 仅做前端过滤，匹配问题标题、影响对象、证据源、入口服务、入口 endpoint。 |

## 核心接口

| 数据域 | 接口 | 关键字段 | 说明 |
| --- | --- | --- | --- |
| 活跃告警 | `POST /api/alerts/alert/events` | `alertEvents[]`, `pagination.total`, `impactEndpoints`, `impactEntrys` | 查询当前窗口内 firing 告警，并补齐告警影响的服务和入口。 |
| 影响入口 | `POST /api/alerts/event/impact` | `data[].serviceName`, `data[].endpoint`, `data[].relatedAlertRate`, `data[].latency`, `data[].errorRate`, `data[].tps`, `data[].namespaces` | 查询受告警影响的业务入口，并由服务总览逻辑补齐 RED 指标。 |
| 调用关系 | `POST /api/service/relation` | `parents`, `current`, `childRelations`, `topologyLevels` | 对每个入口查询下游路径和拓扑层级。 |
| 下游贡献度 | `ANY /api/alerts/descendant/anormal/contribution` | `contributationMap.latency`, `contributationMap.errorRate`, `latencyContributationList` | 基于 PolarisAnalyzer 对下游节点按延迟/错误率相关性排序，最多取前三。 |
| 下游异常变化 | `POST /api/alerts/descendant/anormal/delta` | `originAnormalCounts`, `alertTriggeredCounts`, `finalAnormalCounts`, `deltaAnormalEvents`, `finalAnormalEvents` | 判断告警触发前后异常是否新增、持续或恢复。 |
| 下游异常趋势 | `POST /api/alerts/descendant/anormal/trend` | `anormalCount.chartData`, `anormalCount.value`, `anormalCount.ratio` | 展示下游异常趋势和趋势强度。 |
| 下游相关性 | `POST /api/service/descendant/relevance` | `serviceName`, `endpoint`, `distance`, `distanceType`, `REDStatus`, `AlertStatus`, `alertReason` | 作为传播风险和根因候选的辅助证据；建议传 `withAlert=true`。 |
| RED小图 | `POST /api/service/redcharts` | `latency`, `errorRate`, `tps` | 用于入口和下游节点的趋势展示；不直接作为最终健康度。 |

## 明确公式清单

本节给出可直接实现的公式。所有分值默认归一化到 `0-100`；风险分越高越危险，健康分越高越健康。

### 基础函数

```text
clamp(x, min, max) = min(max(x, min), max)

avg(values) =
  len(values) == 0 ? 0 : sum(values) / len(values)

top(values, n) =
  values 按数值降序排序后取前 n 个

safeRatio(numerator, denominator) =
  denominator <= 0 ? null : numerator / denominator
```

```text
ratioRisk(ratioPercent):
  ratioPercent == null      => null
  ratioPercent < 20         => 0
  20 <= ratioPercent < 50   => 35
  50 <= ratioPercent < 100  => 70
  ratioPercent >= 100       => 100
```

```text
statusRisk(status):
  status in [normal, healthy, green]       => 0
  status in [recovering, minor]            => 25
  status in [warning, yellow]              => 50
  status in [error, severe, red]           => 80
  status in [critical, outage, unavailable] => 100
  otherwise                                => 30
```

```text
severityRisk(severity):
  severity == critical => 100
  severity == error    => 80
  severity == warning  => 45
  severity == info     => 20
  otherwise            => 30
```

### 问题输入变量

对每个问题 `issue` 先形成以下变量：

```text
entries = issue 关联的 /api/alerts/event/impact.data[]
alerts = issue 关联的 /api/alerts/alert/events.alertEvents[]
relations = issue 入口对应的 /api/service/relation.childRelations
levels = issue 入口对应的 /api/service/relation.topologyLevels
contribution.latency = /api/alerts/descendant/anormal/contribution.contributationMap.latency
contribution.errorRate = /api/alerts/descendant/anormal/contribution.contributationMap.errorRate
descendantRelevance = /api/service/descendant/relevance 返回的下游节点列表
```

### 入口影响分

```text
entryRisk(entry) = clamp(entry.relatedAlertRate, 0, 100)

entryImpactScore(issue) =
  len(entries) == 0
    ? 0
    : 0.70 * max(entryRisk(entry) for entry in entries)
      + 0.30 * avg(top(entryRisk(entry) for entry in entries, 3))
```

含义：`relatedAlertRate` 已由后端按“该入口关联告警次数 / 当前窗口告警总次数 * 100”产出。多入口时用最大入口影响为主，Top3 平均避免单点噪声。

### 告警严重度分

```text
activeAlerts = [alert for alert in alerts if alert.status == "firing"]

maxSeverityScore =
  len(activeAlerts) == 0 ? 0 : max(severityRisk(alert.severity or alert.level) for alert in activeAlerts)

alertCountPressure = min(len(activeAlerts) / 5, 1) * 100

alertSeverityScore(issue) =
  0.70 * maxSeverityScore + 0.30 * alertCountPressure
```

### RED 退化分

对每个入口计算：

```text
latencyRisk(entry) = ratioRisk(entry.latency.ratio.value or entry.latency.ratio)
errorRisk(entry) = ratioRisk(entry.errorRate.ratio.value or entry.errorRate.ratio)
trafficRisk(entry) = ratioRisk(abs(entry.tps.ratio.value or entry.tps.ratio))

redRisk(entry) =
  weightedAvgIgnoreNull([
    (latencyRisk(entry), 0.45),
    (errorRisk(entry), 0.45),
    (trafficRisk(entry), 0.10)
  ])
```

```text
weightedAvgIgnoreNull(weightedValues):
  valid = [(value, weight) for value, weight in weightedValues if value != null]
  len(valid) == 0 ? 0 : sum(value * weight for valid) / sum(weight for valid)

redDegradationScore(issue) =
  len(entries) == 0 ? 0 : max(redRisk(entry) for entry in entries)
```

没有 `ratio` 时不硬算退化，只把该数据源记为缺失或低覆盖；有 `value` 时可展示但不进入该分项。

### 下游传播分

```text
maxAffectedDepth =
  len(levels) == 0 ? 0 : max(level.depth for level in levels where level.outOfGroup != true)

topologyDepthRisk = min(maxAffectedDepth / 4, 1) * 100
```

```text
contributionNodes = unique(contribution.latency + contribution.errorRate)

contributionRisk =
  len(contributionNodes) == 0 ? 0 :
  len(contributionNodes) == 1 ? 60 :
  len(contributionNodes) == 2 ? 80 : 100
```

```text
descendantAlertRisk =
  len(descendantRelevance) == 0
    ? 0
    : max(
        max(statusRisk(node.REDStatus), statusRisk(node.AlertStatusCH), statusRisk(node.AlertStatusPROM))
        for node in descendantRelevance
      )
```

```text
propagationScore(issue) =
  0.40 * topologyDepthRisk
  + 0.35 * contributionRisk
  + 0.25 * descendantAlertRisk
```

### 数据完整性惩罚

每个问题期望至少具备 5 类数据：`alerts`、`impactEntries`、`relations`、`redMetrics`、`descendantContribution`。

```text
expectedSources = ["alerts", "impactEntries", "relations", "redMetrics", "descendantContribution"]
availableSources = 本问题实际成功取得且参与计算的数据源集合
missingRequiredSourceCount = len(expectedSources) - len(intersect(expectedSources, availableSources))

dataCompletenessPenalty =
  min(missingRequiredSourceCount * 20, 100)
```

### 单条影响路径风险分

```text
pathRisk(issue) =
  0.35 * entryImpactScore(issue)
  + 0.25 * alertSeverityScore(issue)
  + 0.20 * redDegradationScore(issue)
  + 0.15 * propagationScore(issue)
  + 0.05 * dataCompletenessPenalty(issue)
```

### 问题健康扣分

用于页面“健康度解释”的每个扣分项。

```text
issueWeightInSystemRisk(issue):
  maxPart = issue.pathRisk == max(pathRisk(allIssues)) ? 0.55 : 0
  top3Part = issue in top(pathRisk(allIssues), 3) ? 0.30 / min(3, len(allIssues)) : 0
  avgPart = 0.15 / len(allIssues)
  return maxPart + top3Part + avgPart

scoreImpact(issue) =
  -round(pathRisk(issue) * issueWeightInSystemRisk(issue))
```

### 系统健康度

```text
risks = [pathRisk(issue) for issue in issues]

systemRisk =
  len(risks) == 0
    ? 0
    : 0.55 * max(risks)
      + 0.30 * avg(top(risks, 3))
      + 0.15 * avg(risks)

healthScore = round(100 - clamp(systemRisk, 0, 100))
```

### 系统状态

```text
healthStatus(healthScore):
  healthScore >= 90 => healthy
  healthScore >= 75 => minor
  healthScore >= 60 => warning
  healthScore >= 40 => severe
  otherwise         => outage
```

### 问题状态和优先级

```text
issueStatus(pathRisk):
  pathRisk >= 75 => severe
  pathRisk >= 50 => warning
  pathRisk >= 25 => minor
  otherwise      => healthy
```

```text
priority(issue):
  pathRisk(issue) >= 75
    or any(severityRisk(alert.severity) == 100 for alert in issue.alerts)
    => P0
  pathRisk(issue) >= 50 => P1
  pathRisk(issue) >= 25 => P2
  otherwise => hidden
```

### 内部数据质量评分

```text
sourceCoverage(issue) =
  len(availableSources(issue)) / len(expectedSources)
```

```text
evidenceConsistency(issue):
  sameTarget =
    top contribution node 与 highest RED risk entry/downstream 或 critical alert object 一致 ? 1 : 0
  sameWindow =
    关键告警时间与 RED 退化窗口重叠 ? 1 : 0
  noContradiction =
    不存在明确反证 ? 1 : 0
  return (sameTarget + sameWindow + noContradiction) / 3
```

```text
topologyCompleteness(issue):
  hasEntry = len(entries) > 0 ? 1 : 0
  hasRelation = len(relations) > 0 ? 1 : 0
  hasLevel = len(levels) > 0 ? 1 : 0
  return (hasEntry + hasRelation + hasLevel) / 3
```

```text
freshness(issue):
  latestTs = max(所有参与计算数据点 timestamp)
  lagSeconds = now - latestTs
  collectIntervalSeconds = 60
  return clamp(1 - lagSeconds / (5 * collectIntervalSeconds), 0, 1)
```

```text
issueDataQualityScore =
  round(100 * (
    0.45 * sourceCoverage
    + 0.25 * evidenceConsistency
    + 0.20 * topologyCompleteness
    + 0.10 * freshness
  ))
```

系统内部数据质量评分：

```text
systemDataQualityScore =
  len(issues) == 0 ? 0 : round(avg(issueDataQualityScore for issue in top(issues by pathRisk, 3)))
```

### 页面字段公式对照

| 页面字段 | 公式 |
| --- | --- |
| `AI 健康度` | `healthScore` |
| 状态标签 | `healthStatus(healthScore)` |
| `归并风险` / 问题数量 | `len(issues where priority != hidden)` |
| `活跃告警` | `sum(len(issue.activeAlerts) for issue in issues)` 去重后计数 |
| `N 条活跃告警被归并为 M 个独立问题` | `N = unique firing alerts count`, `M = visible issue count` |
| `X 告警` | `len(issue.activeAlerts)` |
| `Y 类证据` | `len(issue.availableSources)` |
| `健康度解释` 扣分 | `scoreImpact(issue)` |
| 优先级 | `priority(issue)` |
| 风险卡片状态 | `issueStatus(pathRisk(issue))` |
| 排名序号 | 按 `priority(P0>P1>P2)`, `pathRisk desc` 排序后的 index；内部数据质量评分只用于同风险兜底排序 |
| 影响对象 | `unique(entries.serviceName/endpoint + contributionNodes + alert object)` |
| 用户/业务表现 | 模板：`影响 {len(entries)} 个入口，最高告警影响占比 {max(entryRisk)}%，最高路径风险 {round(pathRisk)}` |
| `N 类数据源` | `len(issue.availableSources)` |
| 证据阶段 | `getEvidenceRole(evidence.type).stage` |
| 证据结果 | 接口字段值或上面公式的分项结果，不允许手写样例数值 |

## 健康度计算

### 单条影响路径风险分

每个问题以一个或多个入口路径为单位计算 `pathRisk`，范围为 `0-100`，值越高风险越大。

```text
pathRisk =
  0.35 * entryImpactScore
+ 0.25 * alertSeverityScore
+ 0.20 * redDegradationScore
+ 0.15 * propagationScore
+ 0.05 * dataCompletenessPenalty
```

| 分项 | 来源 | 计算方式 |
| --- | --- | --- |
| `entryImpactScore` | `/api/alerts/event/impact.data[].relatedAlertRate` | 单入口取 `clamp(relatedAlertRate, 0, 100)`；多入口问题取 `max * 0.7 + avg(top3) * 0.3`。 |
| `alertSeverityScore` | `/api/alerts/alert/events.alertEvents[]` | `maxSeverityScore * 0.7 + alertCountPressure * 0.3`；`alertCountPressure = min(activeAlertCount / 5, 1) * 100`。严重级别建议映射：critical=100, error=80, warning=45, info=20, unknown=30。 |
| `redDegradationScore` | `/api/alerts/event/impact.data[].latency/errorRate/tps` 或 `/api/service/redcharts` | `0.45 * latencyRisk + 0.45 * errorRisk + 0.10 * trafficRisk`。优先用 `ratio`；没有 `ratio` 时只展示数值并降低内部数据质量评分，不硬算退化。 |
| `propagationScore` | `/api/service/relation`, `/api/alerts/descendant/anormal/contribution`, `/api/service/descendant/relevance` | `0.40 * topologyDepthRisk + 0.35 * contributionRisk + 0.25 * descendantAlertRisk`。 |
| `dataCompletenessPenalty` | 本次聚合过程 | 必需源缺失时累计；每缺 1 类加 20，封顶 100。该项用于避免“无数据即健康”。 |

### 分项细则

```text
ratioRisk(ratio):
  ratio < 20%        => 0
  20% <= ratio < 50% => 35
  50% <= ratio < 100% => 70
  ratio >= 100%      => 100
```

```text
topologyDepthRisk = min(maxAffectedDepth / 4, 1) * 100
contributionRisk =
  0 个贡献节点 => 0
  1 个贡献节点 => 60
  2 个贡献节点 => 80
  >=3 个贡献节点 => 100
descendantAlertRisk = max(下游节点 REDStatus/AlertStatus 映射分)
```

建议状态映射：

| 状态 | 分值 |
| --- | --- |
| normal/healthy | 0 |
| recovering/minor | 25 |
| warning | 50 |
| error/severe | 80 |
| critical/outage | 100 |

### 系统健康度

系统健康度不能对所有服务简单平均，避免严重路径被大量正常节点稀释。

```text
systemRisk =
  0.55 * max(pathRisk)
+ 0.30 * avg(top3 pathRisk)
+ 0.15 * avg(all pathRisk)

healthScore = round(100 - systemRisk)
```

健康等级：

| 分数 | 等级 | 页面状态 |
| --- | --- | --- |
| 90-100 | healthy | 正常 |
| 75-89 | minor | 轻微异常 |
| 60-74 | warning | 警告 |
| 40-59 | severe | 严重 |
| 0-39 | outage | 故障 |

### 内部数据质量评分

内部数据质量评分和健康度分开计算；该评分用于排序兜底、数据边界和调试，不作为页面核心展示。

```text
dataQualityScore =
  0.45 * sourceCoverage
+ 0.25 * evidenceConsistency
+ 0.20 * topologyCompleteness
+ 0.10 * freshness
```

| 分项 | 计算方式 |
| --- | --- |
| `sourceCoverage` | 已获得的必需数据源数量 / 期望数据源数量。 |
| `evidenceConsistency` | 关键证据方向一致得高分；例如入口 RED、告警对象、调用关系、下游贡献节点指向同一链路。存在反证时降低。 |
| `topologyCompleteness` | `/api/service/relation` 返回入口、当前节点、下游边和层级越完整越高。 |
| `freshness` | 最新数据点距当前时间越短越高；超过一个采集周期开始降分。 |

## 页面逐行数据口径

### 页面头部

| 页面信息 | 正式来源 | 计算方式 | 当前草稿状态 |
| --- | --- | --- | --- |
| `AI 系统健康度分析报告` | 静态标题 | 固定文案。 | 静态。 |
| 状态标签，如“严重” | `healthScore` | 按健康等级表由 `healthScore` 映射。 | 来自 `report.status` mock。 |
| 副标题“融合 APM、eBPF、日志、Trace、基础设施与实时告警数据...” | 静态说明 + 实际可用数据源 | 固定说明可保留，但证据源列表必须按本次聚合实际可用源生成。 | 静态，包含未接入数据源。 |
| 时间窗口 `Last 15 minutes` | URL/global time range | `relativeTime=15m` 展示为 Last 15 minutes；绝对时间展示起止时间。 | 静态。 |
| 证据源 `APM / eBPF / Infra / Logs / Alerts` | 本次聚合的 `availableSources` | 从实际成功返回的接口和字段归并，不成功的不展示或标为缺失。 | 静态。 |
| 分析方式 `风险归并 / 根因定位 / 处置排序` | 静态说明 | 固定方法论文案。 | 静态。 |
| 过滤问题/对象 | 用户输入 + 问题列表 | 前端按 `title/scope/dataSources/entry` 过滤。 | 已实现本地过滤。 |

### 版本B顶部

| 页面信息 | 正式来源 | 计算方式 | 当前草稿状态 |
| --- | --- | --- | --- |
| `Business Impact View` | 静态标签 | 固定。 | 静态。 |
| `按影响面组织处置` | 静态标题 | 固定。 | 静态。 |
| `AI 健康度` 数值 | `healthScore` | 按系统健康度公式计算，不使用 `/api/service/ryglight`。 | 来自 `report.score=42` mock。 |

### “业务健康”子页

| 页面信息 | 正式来源 | 计算方式 | 当前草稿状态 |
| --- | --- | --- | --- |
| 仪表盘分数 | `healthScore` | `100 - systemRisk`。 | mock。 |
| 状态文本 | `healthScore` | 按健康等级表映射。 | mock。 |
| `业务健康判定` 主句 | 规则模板 | 根据最高优先级问题生成模板：`{核心入口/链路}已经{等级描述}，系统{整体状态描述}`。不得凭空生成具体业务对象。 | 静态 mock 文案。 |
| 判定说明段落 | 问题摘要 | 从 top path 的 `title`、`entryImpactScore`、`alertSeverityScore`、`propagationScore` 生成结构化摘要。 | 静态 mock 文案。 |
| `支付成功率/推荐推理/库存预占` 三个 SLI 行 | 正式问题 top3 | 每行显示 top3 问题的业务对象、状态、关键指标；对象和指标必须来自接口。没有对应数据则不显示该行。 | mock，不能直接使用。 |
| `健康度解释` 扣分项 | `pathRisk` top3 | 每个问题展示 `scoreImpact = -round(pathRisk * weightInSystemRisk)`；权重来自系统风险聚合公式。 | mock。 |

### “AI归并”子页

| 页面信息 | 正式来源 | 计算方式 | 当前草稿状态 |
| --- | --- | --- | --- |
| `N 条活跃告警被归并为 M 个独立问题` | `/api/alerts/alert/events`, 归并结果 | `N = firing alertEvents count`；`M = issues.length`。 | mock。 |
| 归并说明 | 静态方法论 + 归并证据 | 固定说明；实际归并依据逐簇显示。 | 静态。 |
| 问题标题 | 归并结果 | 优先取受影响入口服务/endpoint 或主要告警名生成。 | mock。 |
| 优先级 P0/P1/P2 | `pathRisk`, 业务影响 | P0: `pathRisk >= 75` 或 critical 告警影响核心入口；P1: `50-74`；P2: `25-49`；低于 25 不进入主列表。 | mock。 |
| 归并依据 | 告警、入口、拓扑、指标共振 | 按规则拼接：`{告警数}条活跃告警、{证据源数}类证据在同一时间窗指向{scope}`。 | mock。 |
| `X 告警` | `/api/alerts/alert/events` | 问题关联的 firing 告警数。 | mock。 |
| `Y 类证据` | 聚合过程 | 成功命中的证据源类型去重计数。 | mock。 |
| `主处置对象/独立跟踪` | 排序结果 | 排名第 1 且 P0 为主处置对象，其余为独立跟踪。 | mock。 |
| `scope` | `/api/alerts/event/impact`, `/api/service/relation` | 入口、关键下游、告警对象合并为 `service/endpoint` 或 `service -> downstream`。 | mock。 |

### “根因证据”子页

| 页面信息 | 正式来源 | 计算方式 | 当前草稿状态 |
| --- | --- | --- | --- |
| 根因候选标题 | 贡献度 + 拓扑 + 指标退化 | 若 top contribution 节点与 RED/告警对象一致，生成候选：`{节点/依赖} 指标退化`。不能生成数据库连接等待等未取到的具体结论。 | mock。 |
| 证据标签，例如 APM P99、timeout | RED/告警/异常事件 | 只展示实际字段存在的指标。 | mock。 |
| 反证：流量突增已弱化 | RED tps/qps ratio | 若 `tps.ratio` 未明显升高且 latency/error 升高，可标记“流量突增证据不足”。没有流量数据则显示“待确认”。 | mock。 |
| 反证：网络整体故障不支持 | 网络/拓扑相关证据 | 当前 APO 已有健康页代码没有统一 eBPF 网络聚合字段；无数据时只能标“未接入/待确认”。 | mock。 |
| 反证：发布或配置变更待核验 | 变更/工作流数据 | 当前可跳转 `/workflows`，但健康页没有现成变更事件接口。默认标“待核验”，不参与自动评分。 | mock。 |

### “处置优先级”子页

| 页面信息 | 正式来源 | 计算方式 | 当前草稿状态 |
| --- | --- | --- | --- |
| 排序说明 | 静态方法论 | 固定显示：业务影响、健康度扣分、证据完整性、扩散风险。 | 静态。 |
| 排名序号 | 排序结果 | 按 `priority`、`pathRisk` 排序；内部数据质量评分只用于同风险兜底排序。 | mock。 |
| P0/P1/P2 | 风险等级 | 同优先级规则。 | mock。 |
| 风险标题 | 归并结果 | 使用主入口/主告警生成。 | mock。 |
| SLA | 规则配置 | 由优先级映射默认 SLA：P0 30min，P1 60min，P2 观察窗口；没有配置时只显示优先级。 | mock。 |
| 当前目标 | playbook模板 | 根据最高风险分项生成：入口影响高 -> 降低业务影响；告警高 -> 处理告警对象；传播高 -> 阻断传播。 | mock。 |
| 验证信号 | 指标来源 | 与风险分项绑定：latency/errorRate/alert count/descendant relevance 回落。 | mock。 |
| 进入处置按钮 | 路由 | 跳转到告警页并携带风险对象参数。 | 已有跳转样例。 |

### “依赖健康”子页

| 页面信息 | 正式来源 | 计算方式 | 当前草稿状态 |
| --- | --- | --- | --- |
| 总览说明 | 问题 + 拓扑边界 | 根据 top 问题和调用关系生成。 | mock。 |
| 每条依赖 lane 的业务域 | `/api/alerts/event/impact.data[]` | 入口服务/endpoint 作为业务域。若没有业务域元数据，显示 service/endpoint。 | mock。 |
| 当前节点 | `/api/service/relation.current` 或入口 | 展示入口或被分析服务节点。 | mock。 |
| 下游节点 | `/api/service/relation.childRelations` | 展示贡献度最高或 RED/告警异常最高的下游节点。 | mock。 |
| lane 描述 | 贡献度/RED/告警 | 结构化模板：`{priority}: {主退化指标} + {传播/贡献证据}`。 | mock。 |

### 风险卡片列表

每个问题卡片必须由同一套 `SystemHealthIssue` 对象驱动。

| 页面行 | 正式来源 | 计算方式 | 当前草稿状态 |
| --- | --- | --- | --- |
| 优先级 | `pathRisk`, 告警严重度 | P0/P1/P2 映射。 | mock。 |
| 标题 | 影响入口/告警名/贡献节点 | 优先级：业务入口名称 > 聚合告警名 > 最高贡献节点。 | mock。 |
| 状态标签 | `pathRisk` 或 `healthScore` | 单问题状态按 `pathRisk` 映射：`>=75 severe`, `50-74 warning`, `25-49 minor/recovering`。 | mock。 |
| 活跃告警 | `/api/alerts/alert/events` | 问题关联 firing 告警数。 | mock。 |
| 开始时间 | 关联告警最早 `receivedTime/startTime` | 取问题关联 firing 告警最早时间；没有告警则取异常趋势首次异常点。 | mock。 |
| 影响对象 | `/api/alerts/event/impact`, `/api/service/relation` | 入口 + 关键下游 + 告警对象去重展示。 | mock。 |
| 用户/业务表现 | 入口影响分 + RED + 告警 | 结构化模板，不写无法证明的自然语言。例：`影响 {entryCount} 个入口，最高告警影响占比 {relatedAlertRate}%`。 | mock。 |
| `AI 推理证据` | 聚合结果 | 正式名称建议改为“推理证据”或“分析证据”；不要求 LLM。 | 静态。 |
| `N 类数据源` | `availableSources` | 对本问题命中的数据源去重计数。 | mock。 |
| 证据摘要 | 主要证据 | 从最高权重证据生成一句模板。 | mock。 |
| 数据源标签 | `availableSources` | 成功返回且用于本问题计算的数据源。 | mock。 |
| 证据卡片阶段 | `getEvidenceRole(type)` | 按证据类型映射：APM->指标画像，Trace->链路落点，日志->语义佐证，eBPF->根因收敛，其他->证据补充。 | 已有前端映射函数，输入来自 mock。 |
| 证据卡片类型 | 证据对象 `type` | 来自接口/聚合过程，例如 RED指标、告警事件、Trace链路、下游贡献度。 | mock。 |
| 证据结果 | 证据对象 `result` | 只填实际接口返回或可计算结果。 | mock。 |
| 证据作用说明 | 静态映射 | 可以保留现有 `getEvidenceRole` 文案。 | 静态。 |
| 同类证据合并 | `type/source/query` 相同的证据 | 后端优先合并同类告警明细：每个问题只保留 1 张“告警明细”卡，最多展示 3 条代表明细并标明总数；前端再按 `type/source/query` 做兜底合并并显示 `xN`。 | 已按告警事件口径实现。 |
| `AI 处置计划` 标题 | 静态 | 固定文案；建议改为“处置计划”。 | 静态。 |
| 处置排序说明 | 静态 + 排序依据 | 固定说明。 | 静态。 |
| 处置步骤编号 | playbook结果 | 从规则模板生成并排序。 | mock。 |
| 处置阶段 | playbook模板 | 根据主风险分项选择阶段，如风险控制、根因定位、恢复验证。 | mock。 |
| 处置动作 | playbook模板 + 对象参数 | 只能基于已确认对象生成，例如“查看关联告警”“观察 latency/errorRate 是否回落”。涉及回滚、扩容、清理进程等动作需要外部变更/资源数据支撑，否则标为建议核验项。 | mock。 |
| 查看关联告警按钮 | `/alerts/events` | 带 `objectType=risk`, `objectName=scope` 参数跳转。 | 已有跳转样例。 |

## 告警聚合原则

系统健康判断不得按告警数量直接汇总。聚合过程必须优先回答“哪些业务入口或关键依赖正在受影响、影响有多严重、是否持续和扩散、证据是否一致”，再输出健康分、健康等级和处置优先级。

### 聚合层级

告警事件应按以下层级逐步收敛：

```text
原始告警事件
  -> 去重后的告警实例
    -> 告警对象健康
      -> 服务健康
        -> 业务入口 / 业务链路健康
          -> 问题
            -> 系统整体健康
```

各层级职责如下：

| 层级 | 聚合目标 | 关键判断 |
| --- | --- | --- |
| 原始告警事件 | 保留原始来源和时间线 | 不直接参与健康结论，先完成标准化和去重。 |
| 告警实例 | 合并同一规则在短时间内的重复触发 | 记录首次触发、最近触发、重复次数、持续时间和恢复状态。 |
| 告警对象 | 判断单个 pod、node、database、queue、endpoint 等对象风险 | 以最高严重度、持续时间和对象关键性为主，不用低级告警数量覆盖高级告警。 |
| 服务 | 判断服务级受损程度 | 按服务实例、数据库、缓存、MQ、下游依赖等对象的重要性加权；核心依赖异常可直接提升服务风险。 |
| 业务入口 / 链路 | 判断用户可感知影响 | 以入口 `relatedAlertRate`、RED 退化、调用链传播和核心链路标识为主。 |
| 问题 | 把同一根因或同一影响路径的多条告警合并 | 用入口、拓扑、贡献节点、告警对象、时间窗口和证据方向一致性生成 `aggregationKey`。 |
| 系统整体健康 | 输出系统状态和健康分 | 用最高风险、Top3 风险和平均风险组合，避免严重路径被大量正常对象稀释。 |

### 事件标准化

进入聚合前，所有告警必须统一为内部字段，不同来源的字段名和枚举值不得直接进入评分逻辑。

| 字段 | 要求 |
| --- | --- |
| `eventId/source` | 保留来源系统和原始事件 ID，便于回溯。 |
| `status` | 统一为 `firing/recovered/inspection/unknown` 等内部状态。 |
| `severity` | 统一映射为 `critical/error/warning/info/unknown`，并进入 `severityRisk()`。 |
| `startedAt/updatedAt/recoveredAt` | 用于持续时间、新鲜度和时间窗重叠判断。 |
| `serviceName/endpoint` | 作为业务入口和服务聚合主键；缺失时只能降级到对象聚合。 |
| `objectType/objectName` | 标识 pod、node、database、mq、namespace、集群 等告警对象。 |
| `labels/tags` | 用于补齐数据组、环境、集群、namespace、owner 等上下文。 |
| `rawRef` | 保留原始告警引用，证据卡和排查跳转使用。 |

严重度映射必须单调稳定。未知级别不得默认当作健康，可按 `unknown=30` 进入风险分，并降低内部数据质量评分。

### 去重和实例化

同一个对象、同一个告警规则、同一个来源，在同一时间窗口内重复触发，应归并为一个告警实例：

```text
dedupKey = source + "::" + alertName + "::" + objectType + "::" + objectName
```

去重后保留以下聚合字段：

```text
firstSeenAt = min(event.startedAt)
lastSeenAt = max(event.updatedAt)
repeatCount = count(events)
durationSeconds = now - firstSeenAt
status = 任一事件 firing ? firing : recovered
maxSeverity = max(severityRisk(event.severity))
```

去重只消除重复噪声，不消除独立影响。若同一规则命中不同核心对象，应保留为多个实例，再由问题阶段判断是否属于同一影响路径。

### 对象和服务健康

对象级风险以最高影响为主，重复次数和持续时间只作为加权压力：

```text
objectRisk =
  0.70 * maxSeverityRisk
+ 0.15 * durationRisk
+ 0.15 * repeatPressure
```

服务级风险不能对所有对象简单平均，应按对象角色和关键性加权：

```text
serviceRisk =
  0.35 * apiInstanceRisk
+ 0.30 * statefulDependencyRisk
+ 0.15 * cacheRisk
+ 0.10 * queueRisk
+ 0.10 * downstreamRisk
```

如果数据库、主队列、注册中心、认证服务等核心依赖不可用，服务风险应至少提升到 `severe`；如果只有非核心对象低级告警且无入口影响，服务状态最多为 `minor` 或 `warning`。

### 业务影响优先

系统健康结论以业务入口和核心链路受影响程度为主，基础设施告警只有在能映射到服务、入口或关键依赖时才提升系统级风险。

判断优先级如下：

1. 核心业务入口是否出现 `relatedAlertRate` 升高、错误率升高、延迟退化或流量异常。
2. 告警对象是否位于入口服务、核心下游、数据库、缓存、MQ 等关键路径。
3. 调用关系和贡献度是否显示影响正在向下游传播。
4. 告警时间窗是否与 RED 退化、异常趋势、日志或 trace 证据重叠。
5. 是否有反证，例如 RED 正常、告警已恢复、拓扑不连通或对象不在当前数据组。

没有业务入口和拓扑证据时，可以输出对象级或服务级风险，但不得声称用户业务已经受损。

### 问题归并

问题用于把同一根因或同一影响路径下的告警压缩为一个可处置单元。归并时优先使用以下证据：

| 证据 | 归并含义 |
| --- | --- |
| 同一 `serviceName/endpoint` | 多条告警影响同一业务入口。 |
| 同一关键下游贡献节点 | 多个入口共同指向同一个数据库、缓存、MQ 或下游服务。 |
| 同一告警对象 | 多条告警命中同一 pod、node、database、queue 等对象。 |
| 调用关系连通 | 入口、当前服务、下游节点在同一调用路径上。 |
| 时间窗口重叠 | 告警触发、RED 退化、异常趋势在同一分析窗口内重叠。 |
| 证据方向一致 | 告警对象、RED 异常、贡献节点和日志/trace 证据指向同一链路。 |

只有告警名称相似、文案相似或同属于一个数据组，不足以归并为同一问题。

### 时间窗口和状态处理

健康判断必须绑定统一时间窗口。页面默认窗口为 `relativeTime=15m`，所有接口都应使用同一 `startTime/endTime`。

| 场景 | 处理原则 |
| --- | --- |
| 短暂抖动且已恢复 | 降低风险，但保留恢复事件和最近恢复时间。 |
| 长时间 firing | 随持续时间提高风险和处置优先级。 |
| 告警恢复但 RED 仍异常 | 不能直接判定健康，应保留为恢复中或继续观察。 |
| 告警仍 firing 但入口和 RED 正常 | 降低业务影响分，但对象风险仍保留。 |
| 数据源缺失 | 不推断为健康，进入 `dataCompletenessPenalty` 和内部数据质量扣分。 |

### 系统健康输出

最终输出必须同时包含健康结论、主要原因、影响范围和数据边界。建议结构如下：

```json
{
  "status": "severe",
  "healthScore": 42,
  "mainReason": "订单入口相关告警与下游数据库风险在同一窗口内重叠",
  "activeAlertCount": 18,
  "issueCount": 3,
  "affectedServices": ["order-service", "payment-service"],
  "affectedObjects": ["order-db-01"],
  "topIssueAggregationKey": "order-service::/api/order/create"
}
```

输出文案必须来自结构化字段或可计算公式。不得因为告警数量多就直接判定系统严重，也不得因为告警数量少就忽略核心链路上的高严重度风险。

## 问题生成方式

### 输入

1. 从 `/api/alerts/alert/events` 获取 firing 告警，得到告警事件、告警对象、影响服务和影响入口。
2. 从 `/api/alerts/event/impact` 获取受影响入口和 `relatedAlertRate`、RED摘要。
3. 对每个入口调用 `/api/service/relation` 获取下游边和拓扑层级。
4. 对每个入口调用 `/api/alerts/descendant/anormal/contribution` 获取延迟/错误率贡献节点。
5. 对每个入口调用 `/api/service/descendant/relevance?withAlert=true` 获取下游相关性和告警状态。
6. 可选调用 `/api/alerts/descendant/anormal/delta` 和 `/trend` 补充异常变化。

### 归并规则

问题按以下 key 合并：

```text
aggregationKey = primaryEntry.serviceName + "::" + primaryEntry.endpoint
```

若多个入口命中同一关键下游贡献节点，且告警时间窗重叠，可合并：

```text
sameContributionNode && overlap(windowA, windowB) && evidenceDirectionConsistent
```

禁止合并的情况：

1. 入口调用关系不连通。
2. 贡献节点不同且没有共同上游/下游。
3. 一个风险处于恢复中，另一个仍在恶化，且缺少共同证据。
4. 只有文案相似，没有共同告警对象或共同拓扑边。

## 建议前端聚合结构

```ts
type SystemHealthReport = {
  timeRange: { startTime: number; endTime: number; relativeTime?: string }
  groupId: number
  healthScore: number
  status: 'healthy' | 'minor' | 'warning' | 'severe' | 'outage'
  availableSources: string[]
  summary: {
    issueCount: number
    activeAlertCount: number
    affectedEntryCount: number
    maxPathRisk: number
  }
  scoreFactors: Array<{
    key: string
    label: string
    value: number
    riskScore: number
  }>
  issues: SystemHealthIssue[]
}

type SystemHealthIssue = {
  key: string
  priority: 'P0' | 'P1' | 'P2'
  title: string
  status: 'healthy' | 'minor' | 'warning' | 'severe' | 'outage' | 'recovering'
  scope: string
  activeAlerts: number
  pathRisk: number
  scoreImpact: number
  startedAt?: string
  updatedAt?: string
  entries: Array<{ serviceName: string; endpoint: string; relatedAlertRate: number }>
  affectedNodes: Array<{ serviceName: string; endpoint: string; group?: string }>
  riskBreakdown: {
    entryImpactScore: number
    alertSeverityScore: number
    redDegradationScore: number
    propagationScore: number
    dataCompletenessPenalty: number
  }
  availableSources: string[]
  evidence: Array<{
    type: string
    source: string
    result: string
    rawRef?: unknown
  }>
  actionPlan: Array<{
    phase: string
    target: string
    action: string
    validation: string
    risk?: string
  }>
}
```

## 当前实现状态

当前 worktree 已新增 `GET /api/system-health/report`，前端 `frontend/src/pro/views/serviceHealthy/index.js` 已改为读取该接口。该接口已经完成的真实数据链路如下：

| 已实现数据 | 当前来源 | 当前计算/生成方式 |
| --- | --- | --- |
| 时间窗口 | `relativeTime/startTime/endTime` | `resolveTimeRange()` 解析；默认 `15m`。 |
| 数据组过滤 | `groupId` | 查询 alert event 时展开子数据组并设置 `SubGroupIDs`。 |
| 真实事件输入 | ClickHouse `GetAlertEventWithWorkflowRecord` 中的告警事件字段 | 读取当前窗口内告警事件；当前页面可声明的有效数据源只有“真实告警事件”。 |
| 问题类型 | `event.Status`, `event.Validity`, `event.IsValid`, 事件文本 | 分为 `activeAlert`、`inspection`、`recoveredAlert` 三类。 |
| 问题标题 | 告警事件 `Name` | 巡检事件加 `巡检指标异常：` 前缀；恢复事件加 `已恢复告警：` 前缀。 |
| 影响对象 `scope` | 事件标签 | 依次取 `targetObj`、service、infra node、database URL、namespace、pod、group。 |
| 活跃告警数 | firing 且有效的 active alert | 对归并后的事件计数。 |
| 证据 | alert event | 只生成“告警事件”和“告警明细”证据卡；APM/Trace/日志/eBPF 不作为当前证据展示。 |
| 健康分 | 当前简化版 `pathRisk` 聚合 | 使用告警影响、告警严重度、巡检异常、传播简化值、数据完整性惩罚计算；缺失 APM/入口/拓扑/下游贡献度会进入完整性惩罚。 |
| 内部数据质量评分 | 当前简化版覆盖率 + 新鲜度 | 按 `alerts/impactEntries/relations/redMetrics/descendantContribution` 的期望源覆盖率计算；当前只有 `alerts` 命中。 |
| 处置计划 | 固定规则模板 | active/recovered/inspection 三类分别生成通用处置步骤。 |

## 当前页面未实现数据清单

本节按当前页面实际展示项列出“还未实现或仅部分实现”的数据。`阻塞程度` 中，`高` 表示页面当前文案会让用户误以为已有真实数据支撑；`中` 表示已可展示但口径较粗；`低` 表示主要是增强项。

| 页面位置 | 页面字段/文案 | 当前状态 | 缺少的数据 | 需要接入/计算方式 | 阻塞程度 |
| --- | --- | --- | --- | --- | --- |
| 页面头部 | `融合 APM、eBPF、日志、Trace、基础设施与实时告警数据` | 未实现，需删除 | 当前有效数据源只有真实告警事件；没有统一拉取 APM、eBPF、日志、Trace、基础设施指标 | 改为“基于当前已接入的真实告警事件归并风险”；其他源只在 `missingData` 中标记未接入 | 高 |
| 页面头部 | `证据源：真实告警事件 / 有效性校验` | 需修正 | 当前可声明的数据源只有真实告警事件；有效性校验不作为页面证据源展示 | 后端 `report.availableSources` 固定来自本次 issues 的真实参与评分源，当前只应为“真实告警事件” | 中 |
| 顶部指挥条 | `AI 健康度` | 部分实现 | 健康度已计算，但还不是完整影响路径公式：缺少入口影响、RED 退化、拓扑传播、下游贡献度 | 将 `entryImpactScore/redDegradationScore/propagationScore` 从真实接口补齐，而不是当前常量/简化值 | 高 |
| 顶部说明 | `P0/P1 来自有效进行中告警归纳，P2 来自巡检异常指标，P3 来自有效但已恢复告警` | 部分实现 | 当前 P2 只靠文本识别 `mutation/detect/inspection/巡检`，没有真实巡检指标值 | 接入异常检测执行记录或异常趋势数据，明确巡检对象、指标、阈值、当前值 | 中 |
| 顶部 KPI | `活跃问题` | 已实现但口径粗 | 当前等于过滤后的 issue 数，不区分真实影响路径数量 | 若版本B强调影响路径，应改为 `visible issues` 或 `affectedEntryCount` | 低 |
| 顶部 KPI | `对象` | 部分实现 | 当前前端按 `scope.split('/')` 粗略拆分，可能把普通字符串误拆成多个对象 | 后端返回结构化 `affectedObjects[]`，包含 `type/name/service/endpoint` | 中 |
| 评分条 | `scoreFactors[].reason` | 部分实现 | 当前 reason 只包含告警影响、告警严重度、巡检异常；缺少 RED、传播、完整性分项 | 后端返回 `riskBreakdown` 并生成完整原因：入口影响/告警严重度/RED/传播/完整性 | 中 |
| 风险卡片 | `影响对象` | 部分实现 | 当前只有单个 `scope` 字符串；没有入口、当前节点、下游节点、告警对象分层 | 从 `/api/alerts/event/impact`、`/api/service/relation`、告警标签合成 `entries[]`、`affectedNodes[]` | 高 |
| 风险卡片 | `用户/业务表现` | 部分实现 | 当前是基于事件类型的模板句，不能说明入口影响比例、RED 变化、业务成功率 | 用 `relatedAlertRate`、入口 RED 指标、告警数生成模板；没有数据时只说“影响对象为 X” | 高 |
| 风险卡片 | `AI 推理证据` | 部分实现 | 当前不是 LLM 推理，主要是规则模板；字段名会造成误解 | 改名“分析证据”，或接入 LLM 且传入证据对象和禁止编造约束 | 中 |
| 风险卡片 | `N 类数据源` | 已按实际修正 | 当前只应显示 1 类：真实告警事件；事件 `Source/Group` 只放入告警详情，不计为独立数据源 | 后续接入 APM/Trace/日志/eBPF 后再进入 `availableSources` | 中 |
| 证据卡片 | APM 指标画像 | 未实现 | 没有从 `/api/service/redcharts` 或 `/api/alerts/event/impact` 取 latency/error/tps 实值 | 对入口服务/endpoint 查询 RED 指标，生成 APM/RED 证据卡 | 高 |
| 证据卡片 | eBPF 根因收敛 | 未实现 | 仅当事件来源包含 eBPF 才显示泛化描述；没有 eBPF 网络/系统调用明细 | 接入 eBPF 数据源或专门接口，至少返回 tcp retrans/connect latency/syscall/run queue | 高 |
| 证据卡片 | Trace 链路落点 | 未实现 | 当前没有查询 Trace 慢 span、异常 span、调用链落点 | 接入 `/api/trace/*` 或异常 span 能力，按入口/告警对象查慢调用链 | 高 |
| 证据卡片 | 日志语义佐证 | 未实现 | 当前没有查询日志关键词、重复次数、错误模式 | 接入 `/api/log/query`、`/api/log/chart` 或故障现场日志接口，返回关键词、次数、趋势 | 高 |
| 证据卡片 | 基础设施/资源证据 | 未实现 | 只有 `infraAlerts` 按 group=infra 计数；没有 CPU、内存、磁盘、网络等指标 | 接入 metric/k8s/network 相关接口，按告警对象映射 node/pod/container | 中 |
| 证据卡片 | 数据库连接池、慢 SQL | 未实现 | 当前没有连接池、慢 SQL TopN、SQL 指标用于证据 | 核对并接入 `/api/service/sql/metrics` 或新增数据库指标来源 | 中 |
| 风险分项 | `entryImpactScore` | 部分实现 | 当前按事件类型和 activeAlerts 简化计算，不使用业务入口 `relatedAlertRate` | 从 `/api/alerts/event/impact.data[].relatedAlertRate` 计算 | 高 |
| 风险分项 | `redDegradationScore` | 部分实现 | 当前只有巡检事件固定 45，其余为 0 | 从 RED ratio 计算 latency/error/tps 风险 | 高 |
| 风险分项 | `propagationScore` | 部分实现 | 当前按同组事件数量估算；没有拓扑深度、贡献节点、下游告警 | 接入 `/api/service/relation`、`/api/alerts/descendant/anormal/contribution`、`/api/service/descendant/relevance` | 高 |
| 风险分项 | `dataCompletenessPenalty` | 已按实际修正 | 当前只有 `alerts`，其余 `impactEntries/relations/redMetrics/descendantContribution` 缺失 | expectedSources 使用 `alerts/impactEntries/relations/redMetrics/descendantContribution`；当前缺失 4 类 | 中 |
| 风险归并 | `MergeBasis` | 部分实现 | 当前按同一 group/name/scope 聚合，不是真正按入口、调用关系、指标共振归并 | 使用入口、贡献节点、告警对象、时间窗口重叠做归并 | 高 |
| 优先级 | `P0/P1/P2/P3` | 部分实现 | 当前 P0 只由 critical severity 决定，P1/P2 主要由简化风险分决定 | 按完整 `pathRisk`、核心入口影响、critical 告警、扩散风险共同决定 | 中 |
| 处置计划 | `AI 处置计划` | 部分实现 | 当前为三类通用模板，没有按根因、指标、对象生成具体动作 | 基于最高风险分项生成 playbook：影响控制、RED定位、拓扑传播、恢复验证 | 中 |
| 处置计划 | `Validation` | 部分实现 | 当前是通用“无新增有效告警/指标趋势” | 绑定实际指标：latency/errorRate/alert count/contribution node 状态回落 | 中 |
| 处置计划 | `Owner` | 未实现 | 后端返回空字符串 | 从数据组、服务归属、告警标签、团队配置或规则文件映射 owner | 中 |
| 处置计划 | `SLA` | 部分实现 | 当前按优先级固定默认 SLA | 后续应支持规则配置或告警级别/业务等级映射 | 低 |
| 按钮 | `查看关联告警` | 部分实现 | 当前只跳转并带 `objectType/objectName`，目标页是否消费这些参数未确认 | 目标告警页需要支持按对象/风险 scope 过滤，或传 alert IDs | 中 |
| 空态 | `未发现可归纳的有效告警、巡检异常或已恢复有效告警` | 已实现但口径需说明 | 当前只说明告警事件范围无问题，不代表 APM/Trace/Logs/Infra 健康 | 空态文案需加“已接入的告警事件范围内” | 中 |
| 响应字段 | `resourceUsage` | 未实现 | 后端固定 `0`，前端当前未展示 | 若未来展示资源使用率，需从基础设施指标计算 | 低 |
| 响应字段 | `logRiskCount` | 未实现 | 后端固定 `0`，前端当前未展示 | 若未来展示日志风险数，需从日志查询/日志告警计算 | 低 |
| 响应字段 | `infraAlerts` | 部分实现 | 只按 `event.Group == "infra"` 计数 | 若作为基础设施风险，需要映射 node/pod 和指标趋势 | 低 |

## 当前可删除或需改名的页面承诺

在未接入对应数据前，以下展示会造成“已经实现多源分析”的误解，建议先动态降级或改名：

| 当前文案 | 建议 |
| --- | --- |
| `融合 APM、eBPF、日志、Trace、基础设施与实时告警数据` | 改为 `基于当前已接入的真实告警事件归并风险、计算健康度并生成处置优先级；未接入的数据会标记为缺失，不参与推断。` |
| `AI 推理证据` | 改为 `分析证据`；只有接入 LLM 推理链路后再恢复 AI 表述。 |
| `APM 指标画像`、`Trace 链路落点`、`日志语义佐证`、`eBPF 根因收敛` 类阶段 | 当前不展示为证据阶段；统一在“未接入数据”中标明缺失。 |
| `有效性校验` 作为证据源 | 不作为当前页面证据源展示；当前 `availableSources` 只展示“真实告警事件”。 |
| `用户/业务表现` | 在没有业务入口影响数据前改为 `影响说明`。 |
| `按影响面组织处置` | 在未接入 `/api/alerts/event/impact` 前可保留为目标，但卡片中应标注“按告警对象归并”。 |

## 当前缺口

| 页面诉求 | 当前 APO 代码状态 | 建议处理 |
| --- | --- | --- |
| eBPF 网络、系统调用、运行队列作为健康中心统一证据 | 当前版本B草稿只有 mock；健康页没有统一 eBPF 证据接口。 | 不展示具体 eBPF 数值；只在真实接口接入后进入 `availableSources`。 |
| GPU 指标、推理队列、模型加载记录 | 当前代码未看到健康中心可直接使用的统一接口。 | 不把“推荐推理容量”作为固定示例；若无数据，页面不生成该问题。 |
| 数据库连接池、慢 SQL TopN | 当前草稿为 mock 文案；现有 SQL metric 能力需单独核对字段是否覆盖连接池/慢 SQL。 | 只展示已有字段；连接池/慢 SQL 作为待核验项。 |
| 发布/配置变更 | 当前页面只有跳转 `/workflows` 的示意。 | 标记“待核验”，不参与自动结论。 |
| 自然语言 AI 结论 | 当前没有后端 AI 聚合接口。 | 第一版使用确定性模板；如接入 LLM，必须传入证据对象和禁止编造的约束。 |

## 落地原则

1. `/api/service/ryglight` 不作为版本B健康评分来源；最多作为节点侧辅助标签，且默认不参与主分。
2. 页面中每个数字都必须能回溯到接口字段或本文公式。
3. 缺失数据不推断为健康，进入内部数据质量评分和数据完整性惩罚。
4. 没有真实来源的业务名、对象名、指标数值、处置动作不得固定展示。
5. 第一版优先用确定性规则生成标题、摘要、优先级和处置建议，避免 mock 文案进入生产。
