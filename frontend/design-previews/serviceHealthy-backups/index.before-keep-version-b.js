import React, { useMemo, useState } from 'react'
import PropTypes from 'prop-types'
import {
  Button,
  Card,
  Col,
  Descriptions,
  Input,
  Progress,
  Row,
  Space,
  Tabs,
  Tag,
} from 'antd'
import {
  Bot,
  Layers,
  Network,
  Sparkles,
} from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import './index.css'

const LEVEL_META = {
  healthy: { text: '正常', color: '#12b76a', tag: 'success', score: 92 },
  minor: { text: '轻微异常', color: '#84cc16', tag: 'lime', score: 82 },
  warning: { text: '警告', color: '#f79009', tag: 'warning', score: 68 },
  severe: { text: '严重', color: '#f04438', tag: 'error', score: 42 },
  outage: { text: '故障', color: '#b42318', tag: 'error', score: 18 },
  recovering: { text: '恢复中', color: '#2e90fa', tag: 'processing', score: 74 },
}

const jumpTargets = {
  service: '/service',
  logs: '/logs/fault-site',
  trace: '/trace/fault-site',
  alerts: '/alerts/events',
  changes: '/workflows',
}

const MOCK_INFRASTRUCTURE = [
  {
    key: 'node-a',
    node: 'prod-node-a',
    role: 'checkout / payment',
    status: 'warning',
    cpu: 82,
    load: '11.4 / 16 cores',
    memory: 76,
    availableMemory: '14.6 GB',
    swap: 8,
    disk: 69,
    diskIo: '320 MB/s',
    diskLatency: '18 ms',
    network: '1.8 Gbps',
    packetLoss: '0.6%',
    connections: '18.2k',
    gpu: 36,
    gpuMemory: 41,
    gpuTemp: 63,
    process: '1,284 proc / 4 zombie',
    recentAlerts: 3,
    aiFinding: 'CPU load 与网络连接数同步升高，支付链路告警集中落在该节点。',
  },
  {
    key: 'node-b',
    node: 'prod-node-b',
    role: 'order / inventory',
    status: 'recovering',
    cpu: 61,
    load: '7.1 / 16 cores',
    memory: 68,
    availableMemory: '22.3 GB',
    swap: 3,
    disk: 72,
    diskIo: '190 MB/s',
    diskLatency: '12 ms',
    network: '920 Mbps',
    packetLoss: '0.1%',
    connections: '9.6k',
    gpu: 0,
    gpuMemory: 0,
    gpuTemp: '-',
    process: '912 proc / 0 zombie',
    recentAlerts: 1,
    aiFinding: '库存服务慢调用已回落，但磁盘写延迟仍高于过去 7 日基线。',
  },
  {
    key: 'node-c',
    node: 'ai-infer-node-01',
    role: 'recommendation inference',
    status: 'warning',
    cpu: 54,
    load: '6.2 / 24 cores',
    memory: 71,
    availableMemory: '38.5 GB',
    swap: 0,
    disk: 63,
    diskIo: '90 MB/s',
    diskLatency: '8 ms',
    network: '740 Mbps',
    packetLoss: '0.0%',
    connections: '5.1k',
    gpu: 88,
    gpuMemory: 84,
    gpuTemp: 78,
    process: '688 proc / 0 zombie',
    recentAlerts: 2,
    aiFinding: 'GPU 显存和温度持续接近阈值，推理服务扩容风险升高。',
  },
]

const MOCK_SERVICES = [
  {
    key: 'payment-api',
    service: 'payment-api',
    status: 'severe',
    score: 38,
    alive: '存活',
    availability: '99.20%',
    successRate: '93.2%',
    error4xx: '1.4%',
    error5xx: '5.4%',
    avgLatency: '420 ms',
    p95: '1.2 s',
    p99: '2.8 s',
    qps: '1,228',
    tps: '380',
    concurrency: 430,
    restarts: 3,
    recentAlerts: 5,
    ebpf: {
      tcpRetrans: '+68%',
      syscallLatency: 'connect 180ms',
      dnsLatency: '32 ms',
      runQueue: '11.4',
      networkFinding: '支付到数据库连接重传升高，connect 系统调用长尾明显。',
    },
    aiFinding: 'P99 延迟和 5xx 错误率共同上升，疑似支付网关到数据库连接池耗尽。',
  },
  {
    key: 'order-service',
    service: 'order-service',
    status: 'warning',
    score: 66,
    alive: '存活',
    availability: '99.81%',
    successRate: '98.1%',
    error4xx: '0.8%',
    error5xx: '1.1%',
    avgLatency: '260 ms',
    p95: '760 ms',
    p99: '1.5 s',
    qps: '2,086',
    tps: '710',
    concurrency: 260,
    restarts: 1,
    recentAlerts: 2,
    ebpf: {
      tcpRetrans: '+24%',
      syscallLatency: 'recvmsg 96ms',
      dnsLatency: '18 ms',
      runQueue: '7.1',
      networkFinding: '上游重试导致连接等待时间增加，服务进程调度延迟轻微升高。',
    },
    aiFinding: '错误率缓慢爬升，受 payment-api 超时重试影响较明显。',
  },
  {
    key: 'inventory-service',
    service: 'inventory-service',
    status: 'recovering',
    score: 74,
    alive: '存活',
    availability: '99.90%',
    successRate: '98.8%',
    error4xx: '0.4%',
    error5xx: '0.8%',
    avgLatency: '180 ms',
    p95: '540 ms',
    p99: '920 ms',
    qps: '860',
    tps: '240',
    concurrency: 112,
    restarts: 0,
    recentAlerts: 1,
    ebpf: {
      tcpRetrans: '+8%',
      syscallLatency: 'read 42ms',
      dnsLatency: '12 ms',
      runQueue: '4.6',
      networkFinding: '网络重传已回落，磁盘读系统调用仍高于恢复期基线。',
    },
    aiFinding: '库存预占慢调用已恢复，但仍需要观察磁盘 I/O 对查询耗时的影响。',
  },
  {
    key: 'user-service',
    service: 'user-service',
    status: 'healthy',
    score: 94,
    alive: '存活',
    availability: '99.99%',
    successRate: '99.9%',
    error4xx: '0.1%',
    error5xx: '0.0%',
    avgLatency: '90 ms',
    p95: '160 ms',
    p99: '240 ms',
    qps: '1,642',
    tps: '520',
    concurrency: 80,
    restarts: 0,
    recentAlerts: 0,
    ebpf: {
      tcpRetrans: '稳定',
      syscallLatency: 'accept 12ms',
      dnsLatency: '9 ms',
      runQueue: '2.1',
      networkFinding: '内核网络、系统调用和进程调度信号稳定。',
    },
    aiFinding: '核心指标稳定，暂未发现服务级风险。',
  },
]

const MOCK_LOGS = [
  {
    key: 'timeout',
    scope: '应用日志',
    pattern: 'timeout / deadline exceeded',
    volume: '+210%',
    repeat: '1,284 次',
    firstSeen: '历史存在，今日突增',
    affected: 'payment-api, order-service',
    severity: 'warning',
    aiFinding: '超时日志与数据库慢查询时间窗口重叠，建议优先核验连接池与 SQL 计划。',
  },
  {
    key: 'oom',
    scope: '系统日志',
    pattern: 'OutOfMemory / memory pressure',
    volume: '+38%',
    repeat: '12 次',
    firstSeen: '新出现',
    affected: 'payment-api instance-7c8',
    severity: 'severe',
    aiFinding: '支付服务重启前出现内存压力，可能导致短时 5xx 和重试风暴。',
  },
  {
    key: 'audit',
    scope: '审计日志',
    pattern: 'config changed',
    volume: '稳定',
    repeat: '3 次',
    firstSeen: '历史存在',
    affected: 'gateway routing',
    severity: 'minor',
    aiFinding: '变更量不高，但与网关 499 增长接近，建议纳入复盘证据。',
  },
]

const MOCK_ALERTS = [
  {
    key: 'a1',
    name: '支付接口 P99 延迟超过基线 2.4 倍',
    source: 'APM',
    level: 'critical',
    status: 'firing',
    object: 'payment-api',
    startedAt: '2026-06-15 10:18:42',
    grouped: '支付链路性能下降',
  },
  {
    key: 'a2',
    name: '数据库连接池使用率超过 90%',
    source: 'Prometheus',
    level: 'warning',
    status: 'firing',
    object: 'mysql-primary',
    startedAt: '2026-06-15 10:20:11',
    grouped: '支付链路性能下降',
  },
  {
    key: 'a3',
    name: 'AI 推理节点 GPU 显存超过 80%',
    source: 'Node Exporter',
    level: 'warning',
    status: 'firing',
    object: 'ai-infer-node-01',
    startedAt: '2026-06-15 10:23:03',
    grouped: '推理容量风险',
  },
  {
    key: 'a4',
    name: '库存预占接口慢调用恢复中',
    source: 'Trace',
    level: 'warning',
    status: 'recovering',
    object: 'inventory-service',
    startedAt: '2026-06-15 10:02:30',
    grouped: '库存服务恢复观察',
  },
]

const ANALYSIS_EVIDENCE = [
  { key: 'apm', label: 'APM 服务指标', value: 'payment-api P99 2.8s，5xx 5.4%', tone: 'severe' },
  { key: 'ebpf-network', label: 'eBPF 网络', value: 'TCP 重传 +68%，数据库连接建立长尾', tone: 'severe' },
  { key: 'ebpf-kernel', label: 'eBPF 内核', value: 'connect 180ms，运行队列 11.4', tone: 'warning' },
  { key: 'logs', label: '日志语义', value: 'timeout 今日突增 +210%，OOM 新增 12 次', tone: 'warning' },
  { key: 'alerts', label: '实时告警', value: '支付链路性能下降聚合告警持续 firing', tone: 'severe' },
]

const SCORE_FACTORS = [
  { key: 'payment', label: 'P0 支付链路', value: -34, tone: 'severe' },
  { key: 'inference', label: 'P1 推理容量', value: -8, tone: 'warning' },
  { key: 'inventory', label: 'P2 恢复观察', value: -4, tone: 'minor' },
]

const HEALTH_RISK_CLUSTERS = [
  {
    key: 'payment-degradation',
    priority: 'P0',
    title: '支付链路性能下降',
    status: 'severe',
    scope: 'payment-api / mysql-primary / order-service',
    activeAlerts: 2,
    confidence: 93,
    owner: '支付值班组',
    sla: '30 分钟内恢复核心支付成功率',
    workflowStatus: '处理中',
    progress: '根因定位',
    scoreImpact: -34,
    updatedAt: '2026-06-15 10:31',
    startedAt: '2026-06-15 10:18',
    mergeBasis: '2 条活跃告警、4 类关键证据在 10:18-10:31 时间窗内同时指向支付写库链路。',
    confidenceBreakdown: ['APM 异常强度高', 'eBPF 根因指向强', 'Trace 与日志时间窗一致', '数据库指标待二次确认'],
    dataQuality: ['数据库慢查询明细延迟约 2 分钟', '未发现与该风险冲突的反证'],
    ebpfFinding: {
      observed: 'payment-api 到 mysql-primary 的 TCP 重传 +68%，connect P99 180ms。',
      excluded: 'QPS 未明显上升，运行队列异常集中在支付节点，暂不支持“单纯流量突增”判断。',
      narrowed: '根因范围从应用慢收敛到数据库连接建立长尾与网络重传。',
    },
    systemProblem:
      '支付接口长尾延迟和 5xx 同时升高，数据库连接池占用超过 90%，订单服务已出现重试传导。当前不是孤立告警，而是支付链路退化问题。',
    aiConclusion:
      'AI 将 APM、数据库指标、日志语义、Trace 和 eBPF 网络/系统调用数据对齐后，判断主矛盾更接近“数据库连接等待 + 慢查询”造成的连接池耗尽，并由支付服务重试放大到订单链路。',
    ebpfRole:
      'eBPF 直接观测到 payment-api 到 mysql-primary 的 TCP 重传 +68%、connect P99 180ms、运行队列 11.4；这些内核侧信号把根因范围从“应用慢”收敛到“数据库连接建立与网络重传长尾”。',
    impact: '影响支付等待、订单状态同步和少量失败重试；若不处理，可能扩散为订单积压和用户支付失败。',
    dataSources: ['APM', 'eBPF 网络', 'eBPF 系统调用', '日志语义', 'Trace', '数据库指标', '实时告警'],
    actions: [
      '优先检查 mysql-primary 当前连接数、等待连接数、连接池借还耗时和慢 SQL TopN。',
      '临时降低 payment-api 与 order-service 的重试放大，保护数据库连接池。',
      '核验 10:10 后支付写库、订单状态更新相关发布或配置变更，必要时回滚可疑变更。',
      '持续观察 eBPF connect 长尾、TCP 重传和 APM P99 是否同步回落，作为恢复确认信号。',
    ],
    actionPlan: [
      {
        phase: '风险控制',
        target: '阻止支付重试继续放大数据库连接等待。',
        action: '临时降低 payment-api 与 order-service 的重试倍数，限制非核心支付重试。',
        validation: '5xx 低于 2%，连接池等待数连续 5 分钟下降。',
        risk: '过度降级会影响少量边缘支付重试体验。',
      },
      {
        phase: '根因定位',
        target: '确认数据库连接池与慢 SQL 是否为主因。',
        action: '检查 mysql-primary 当前连接数、等待连接数、连接池借还耗时和慢 SQL TopN。',
        validation: '慢 SQL 与 connect 长尾在同一时间窗内同步变化。',
        risk: '数据库指标存在 2 分钟采集延迟。',
      },
      {
        phase: '变更核验',
        target: '排除发布或配置变更造成的连接池退化。',
        action: '核验 10:10 后支付写库、订单状态更新相关发布或配置变更。',
        validation: '变更回滚或配置恢复后 P99 与 TCP 重传同步回落。',
        risk: '回滚需确认订单状态一致性。',
      },
      {
        phase: '恢复验证',
        target: '确认核心支付链路恢复且下游重试消退。',
        action: '持续观察 eBPF connect 长尾、TCP 重传、APM P99 和订单同步失败率。',
        validation: '上述指标连续 10 分钟回到基线 1.2 倍以内。',
        risk: '只看应用延迟可能漏掉网络重传残留。',
      },
    ],
    evidence: [
      { type: 'APM 指标', query: 'service=payment-api latency_p99, error_5xx, qps by 1m', result: '服务画像：P99 2.8s，5xx 5.4%，QPS 未明显上升，说明不是流量突增导致的整体拥塞' },
      { type: 'eBPF 网络', query: 'src=payment-api dst=mysql-primary tcp_retrans, connect_latency', result: 'TCP 重传 +68%，connect P99 180ms' },
      { type: '日志查询', query: 'service=payment-api keyword=(timeout OR deadline exceeded)', result: '15 分钟内 1,284 次，较基线 +210%' },
      { type: 'Trace 链路', query: 'trace where root=payment-api and duration>1s group by slow_span', result: '链路落点：慢 span 集中在 payment-api -> mysql-primary 的支付写库与订单状态更新段' },
    ],
  },
  {
    key: 'inference-capacity',
    priority: 'P1',
    title: 'AI 推理容量余量下降',
    level: 'warning',
    status: 'warning',
    scope: 'ai-infer-node-01 / recommendation inference',
    activeAlerts: 1,
    confidence: 86,
    owner: 'AI 平台组',
    sla: '60 分钟内恢复推理容量余量',
    workflowStatus: '待处理',
    progress: '容量确认',
    scoreImpact: -8,
    updatedAt: '2026-06-15 10:29',
    startedAt: '2026-06-15 10:23',
    mergeBasis: 'GPU 指标、推理队列和 eBPF 进程观测共同指向容量余量下降。',
    confidenceBreakdown: ['GPU 指标异常稳定', 'eBPF 排除网络异常', '业务影响尚未越过严重阈值'],
    dataQuality: ['模型版本变更记录待接入', '无明显网络层反证'],
    ebpfFinding: {
      observed: '线程数稳定，运行队列 4.6，TCP 重传和丢包稳定。',
      excluded: '未观察到网络丢包或连接建立长尾，暂不支持网络链路异常判断。',
      narrowed: '处理方向收敛到 GPU 显存、批处理队列和模型加载。',
    },
    systemProblem:
      '推理节点 GPU 显存和温度持续接近阈值，容量余量下降，但尚未与支付链路退化形成同一根因。',
    aiConclusion:
      'AI 通过资源指标、推理服务指标和 eBPF 进程/网络数据交叉判断：该风险主要来自推理节点资源余量，而不是网络丢包或系统调用阻塞。',
    ebpfRole:
      'eBPF 未观察到明显 TCP 重传和连接建立异常，线程数与运行队列稳定，因此帮助排除了网络链路异常，把处理方向定位到 GPU 显存、模型加载和批处理队列。',
    impact: '如果流量继续升高，推荐推理服务可能出现排队、长尾延迟或推理超时。',
    dataSources: ['GPU 指标', '推理服务指标', 'eBPF 进程', 'eBPF 网络', '实时告警'],
    actions: [
      '确认 recommendation inference 的请求量、批处理队列、GPU 显存曲线和模型加载记录。',
      '清理异常显存占用进程，检查批大小或模型版本变更。',
      '准备迁移或扩容推理实例，避免容量风险升级为服务退化。',
    ],
    actionPlan: [
      {
        phase: '容量确认',
        target: '确认推理容量余量是否继续下降。',
        action: '检查请求量、批处理队列、GPU 显存曲线和模型加载记录。',
        validation: '队列深度停止上升，P99 未越过严重阈值。',
        risk: '容量继续下降会导致推荐推理超时。',
      },
      {
        phase: '资源治理',
        target: '释放异常显存占用并降低排队风险。',
        action: '清理异常显存占用进程，检查批大小或模型版本变更。',
        validation: 'GPU 显存回落到 75% 以下，温度低于 75。',
        risk: '清理进程前需确认不会中断在线推理。',
      },
      {
        phase: '扩容准备',
        target: '在容量风险升级前准备承接实例。',
        action: '准备迁移或扩容推理实例。',
        validation: '新增实例接入后队列深度回落。',
        risk: '模型预热不足会造成短时延迟抖动。',
      },
    ],
    evidence: [
      { type: 'GPU 指标', query: 'node=ai-infer-node-01 gpu_util, gpu_mem, gpu_temp by 1m', result: 'GPU 88%，显存 84%，温度 78' },
      { type: '推理服务', query: 'service=recommendation-infer qps, queue_depth, p99', result: '队列轻微上升，P99 未越过严重阈值' },
      { type: 'eBPF 进程', query: 'node=ai-infer-node-01 process_thread, run_queue', result: '线程数稳定，运行队列 4.6' },
      { type: '网络观测', query: 'node=ai-infer-node-01 tcp_retrans, packet_loss', result: '重传稳定，丢包 0.0%' },
    ],
  },
  {
    key: 'inventory-recovery',
    priority: 'P2',
    title: '库存服务恢复期观察',
    status: 'recovering',
    scope: 'inventory-service / prod-node-b',
    activeAlerts: 1,
    confidence: 81,
    owner: '订单履约组',
    sla: '观察 30 分钟无复发后关闭',
    workflowStatus: '恢复观察',
    progress: '恢复验证',
    scoreImpact: -4,
    updatedAt: '2026-06-15 10:28',
    startedAt: '2026-06-15 10:02',
    mergeBasis: 'Trace 慢调用回落、系统日志错误减少，但磁盘 I/O 与 eBPF read 长尾仍高于恢复期基线。',
    confidenceBreakdown: ['Trace 已显示恢复趋势', 'eBPF read 长尾仍在', '业务不可用证据不足'],
    dataQuality: ['磁盘 I/O 基线来自近 7 日同时间段', '未发现新错误类型'],
    ebpfFinding: {
      observed: 'read 系统调用长尾 42ms，网络重传已回落。',
      excluded: '网络重传不再是主要因素，暂不支持继续按网络故障处理。',
      narrowed: '剩余风险收敛到磁盘读取和热点查询。',
    },
    systemProblem:
      '库存预占慢调用已回落，整体处于恢复观察；当前风险低于支付链路，但磁盘 I/O 仍高于恢复期基线。',
    aiConclusion:
      'AI 对比 Trace 慢调用、磁盘延迟、系统日志和 eBPF read 系统调用后，判断该风险更像恢复期尾部波动，不应与支付链路合并为同一个故障。',
    ebpfRole:
      'eBPF 显示网络重传已回落，但 read 系统调用耗时仍高于恢复期基线，说明剩余风险更偏向磁盘读取和查询耗时。',
    impact: '短期可能造成库存预占接口 P95 波动，暂未看到核心业务不可用迹象。',
    dataSources: ['Trace', '磁盘 I/O', 'eBPF 系统调用', '系统日志', '实时告警'],
    actions: [
      '继续观察库存预占接口 P95/P99 与磁盘读延迟是否同步回落。',
      '保留恢复期告警，不与支付链路主风险混合处理。',
      '如果 read 系统调用长尾再次升高，检查热点查询、索引和磁盘队列。',
    ],
    actionPlan: [
      {
        phase: '恢复观察',
        target: '确认库存预占慢调用不再复发。',
        action: '观察库存预占接口 P95/P99 与磁盘读延迟是否同步回落。',
        validation: 'P99 连续 30 分钟稳定在 1s 以下。',
        risk: '提前关闭会漏掉恢复期尾部波动。',
      },
      {
        phase: '风险隔离',
        target: '避免把恢复期波动误归并到支付主风险。',
        action: '保留恢复期告警，独立跟踪库存健康问题。',
        validation: '无新增订单履约失败告警。',
        risk: '风险合并会干扰 P0 支付链路处置。',
      },
      {
        phase: '复发定位',
        target: 'read 长尾再次升高时快速定位磁盘或查询问题。',
        action: '检查热点查询、索引和磁盘队列。',
        validation: 'read 长尾与磁盘队列同步回落。',
        risk: '只看 Trace 可能低估系统调用残留长尾。',
      },
    ],
    evidence: [
      { type: 'Trace 查询', query: 'service=inventory-service span=reserve_stock latency_p95,p99', result: 'P99 920ms，较 10:05 峰值回落 41%' },
      { type: '磁盘 I/O', query: 'node=prod-node-b disk_io, disk_latency by 1m', result: '写延迟 12ms，仍高于 7 日基线' },
      { type: 'eBPF 系统调用', query: 'process=inventory-service syscall=read latency', result: 'read 长尾 42ms，网络重传已回落' },
      { type: '日志查询', query: 'service=inventory-service keyword=(timeout OR slow)', result: '重复错误频率下降，未出现新错误类型' },
    ],
  },
]

const StatusTag = ({ status }) => {
  const meta = LEVEL_META[status] ?? LEVEL_META.warning
  return <Tag color={meta.tag}>{meta.text}</Tag>
}

StatusTag.propTypes = {
  status: PropTypes.string,
}

const SectionTitle = ({ icon, title, extra }) => (
  <div className="health-section-title">
    <Space size={8}>
      {icon}
      <span>{title}</span>
    </Space>
    {extra}
  </div>
)

SectionTitle.propTypes = {
  icon: PropTypes.node,
  title: PropTypes.string.isRequired,
  extra: PropTypes.node,
}

const CompactEvidence = ({ label, value, tone }) => (
  <div className={`health-compact-evidence health-compact-evidence-${tone}`}>
    <span>{label}</span>
    <b>{value}</b>
  </div>
)

CompactEvidence.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.string.isRequired,
  tone: PropTypes.string,
}

const avg = (rows, key) => Math.round(rows.reduce((sum, item) => sum + Number(item[key] || 0), 0) / rows.length)

const getShortConclusion = (text) => text.split('，并由')[0].replace('AI 将 ', '')

const getEvidenceRole = (type) => {
  if (type.includes('APM')) {
    return {
      stage: '指标画像',
      effect: '回答“服务是否真的异常、异常强度多大、是否由流量突增触发”，用于确认风险成立。',
      kind: 'metric',
    }
  }
  if (type.includes('GPU') || type.includes('推理服务')) {
    return {
      stage: '资源画像',
      effect: '确认资源或服务指标已经偏离基线，作为容量风险成立的入口证据。',
      kind: 'metric',
    }
  }
  if (type.includes('Trace')) {
    return {
      stage: '链路落点',
      effect: '回答“慢在哪里发生、沿哪条调用链传播”，用于把服务级异常定位到具体调用段。',
      kind: 'trace',
    }
  }
  if (type.includes('eBPF')) {
    return {
      stage: '根因收敛',
      effect: '从内核侧观测网络、系统调用或进程调度信号，缩小应用慢、网络慢和资源慢之间的根因范围。',
      kind: 'ebpf',
    }
  }
  if (type.includes('日志')) {
    return {
      stage: '语义佐证',
      effect: '用错误语义和重复频率验证异常是否真实发生，并识别新出现的错误模式。',
      kind: 'log',
    }
  }
  if (type.includes('磁盘') || type.includes('网络')) {
    return {
      stage: '环境佐证',
      effect: '补充资源或链路层信号，帮助判断是否需要扩大处置范围。',
      kind: 'infra',
    }
  }
  return {
    stage: '证据补充',
    effect: '补充 AI 判断所需的观测信息，提升结论置信度。',
    kind: 'default',
  }
}

export default function ServiceHealthyPage() {
  const navigate = useNavigate()
  const [keyword, setKeyword] = useState('')
  const [selectedRiskKey, setSelectedRiskKey] = useState(HEALTH_RISK_CLUSTERS[0].key)
  const lowerKeyword = keyword.trim().toLowerCase()

  const filteredRiskClusters = useMemo(() => {
    if (!lowerKeyword) return HEALTH_RISK_CLUSTERS
    return HEALTH_RISK_CLUSTERS.filter(
      (item) =>
        item.title.toLowerCase().includes(lowerKeyword) ||
        item.scope.toLowerCase().includes(lowerKeyword) ||
        item.dataSources.join(' ').toLowerCase().includes(lowerKeyword),
    )
  }, [lowerKeyword])

  const selectedRisk = useMemo(() => {
    return filteredRiskClusters.find((item) => item.key === selectedRiskKey) ?? filteredRiskClusters[0] ?? null
  }, [filteredRiskClusters, selectedRiskKey])

  const report = useMemo(() => {
    const resourceUsage = Math.round(
      (avg(MOCK_INFRASTRUCTURE, 'cpu') + avg(MOCK_INFRASTRUCTURE, 'memory') + avg(MOCK_INFRASTRUCTURE, 'disk')) / 3,
    )
    const highRiskServices = MOCK_SERVICES.filter((item) => ['warning', 'severe', 'outage'].includes(item.status))
    const activeAlerts = MOCK_ALERTS.filter((item) => item.status === 'firing')
    return {
      status: 'severe',
      score: 42,
      confidence: 91,
      resourceUsage,
      infraAlerts: MOCK_INFRASTRUCTURE.reduce((sum, item) => sum + item.recentAlerts, 0),
      serviceRiskCount: highRiskServices.length,
      activeAlertCount: activeAlerts.length,
      logRiskCount: MOCK_LOGS.filter((item) => item.severity !== 'minor').length,
      summary:
        'AI 判断当前系统健康度为 42 分，处于严重退化但未故障状态。风险重心集中在 payment-api 支付链路：APM 看到 P99 延迟和 5xx 同时升高，数据库指标显示连接池接近耗尽，日志语义出现 timeout 突增，Trace 慢 span 集中在写库链路；eBPF 进一步确认 TCP 重传、connect 系统调用长尾和运行队列同步升高。处置应先恢复支付链路，再处理推理容量风险，最后观察库存恢复期波动。',
    }
  }, [])

  const handleAlertTriage = (objectType, objectName) => {
    const params = new URLSearchParams({
      source: 'ai-system-report',
      objectType,
      objectName,
    })
    navigate(`${jumpTargets.alerts}?${params.toString()}`)
  }

  const handleEvidenceDetail = (objectName) => {
    const params = new URLSearchParams({
      source: 'ai-system-report',
      objectName,
    })
    navigate(`${jumpTargets.trace}?${params.toString()}`)
  }

  const currentVersion = (
    <>
      <Row gutter={[12, 12]}>
        <Col span={24}>
          <Card className="health-card health-analysis-card">
            <SectionTitle
              icon={<Sparkles size={17} />}
              title="系统健康度综合分析"
              extra={
                <Space size={6} wrap>
                  <Tag color="error">健康度 {report.score}</Tag>
                  <Tag color="blue">AI 置信度 {report.confidence}%</Tag>
                  <Tag color="blue">eBPF 根因增强</Tag>
                </Space>
              }
            />
            <div className="health-alert-summary-layout">
              <div className="health-analysis-score">
                <Progress
                  type="dashboard"
                  percent={report.score}
                  strokeColor={LEVEL_META[report.status].color}
                  format={() => report.score}
                />
                <div className="health-status-block health-status-block-compact">
                  <span className="health-status-dot" style={{ backgroundColor: LEVEL_META[report.status].color }} />
                  <span className="health-status-text">{LEVEL_META[report.status].text}</span>
                </div>
                <Descriptions size="small" column={1} className="health-descriptions">
                  <Descriptions.Item label="归并问题">{HEALTH_RISK_CLUSTERS.length}</Descriptions.Item>
                  <Descriptions.Item label="活跃告警">{report.activeAlertCount}</Descriptions.Item>
                  <Descriptions.Item label="分析置信度">{report.confidence}%</Descriptions.Item>
                </Descriptions>
              </div>
              <div className="health-alert-summary-body">
                <div className="health-analysis-conclusion">
                  支付链路严重退化，推理容量接近阈值，库存服务处于恢复观察；健康度下降的主因是支付到数据库链路的连接等待和慢查询放大。
                </div>
                <div className="health-alert-summary-grid">
                  <CompactEvidence label="核心问题" value="payment-api P99 与 5xx 同升，数据库连接池接近耗尽" tone="severe" />
                  <CompactEvidence label="eBPF 定位" value="TCP 重传 +68%，connect P99 180ms，指向连接建立长尾" tone="severe" />
                  <CompactEvidence label="数据融合" value="APM、Trace、日志、告警、基础设施与 eBPF 证据互相印证" tone="warning" />
                  <CompactEvidence label="处置顺序" value="P0 支付链路，P1 推理容量，P2 库存恢复观察" tone="minor" />
                </div>
                <div className="health-analysis-paragraph">{report.summary}</div>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[12, 12]} className="health-row">
        <Col span={24}>
          <Card className="health-card">
            <SectionTitle icon={<Layers size={17} />} title="融合数据源对健康度的贡献" />
            <div className="health-analysis-evidence">
              {ANALYSIS_EVIDENCE.map((item) => (
                <CompactEvidence key={item.key} label={item.label} value={item.value} tone={item.tone} />
              ))}
            </div>
            <div className="health-ebpf-highlight">
              <Network size={16} />
              <span>
                eBPF 数据提供了应用指标之外的内核侧证据：连接建立耗时、TCP 重传、系统调用长尾和运行队列，让 AI 能区分“服务自身慢”“数据库连接等待”“网络重传”和“资源调度压力”。
              </span>
            </div>
          </Card>
        </Col>
      </Row>

      <SectionTitle
        icon={<Network size={17} />}
        title="并行风险处置"
        extra={<Tag color="blue">{filteredRiskClusters.length} 个独立问题</Tag>}
      />
      <div className="health-issue-board">
        {filteredRiskClusters.map((cluster) => (
          <Card className={`health-card health-issue-card health-issue-${cluster.priority.toLowerCase()}`} key={cluster.key}>
            <div className="health-issue-head">
              <div>
                <div className="health-issue-title">
                  <span className={`health-risk-priority health-risk-priority-${cluster.priority.toLowerCase()}`}>{cluster.priority}</span>
                  <span>{cluster.title}</span>
                  <StatusTag status={cluster.status} />
                </div>
                <div className="health-alert-event-meta">
                  <span>{cluster.scope}</span>
                  <span>活跃告警 {cluster.activeAlerts}</span>
                  <span>AI 置信度 {cluster.confidence}%</span>
                  <span>开始 {cluster.startedAt}</span>
                </div>
              </div>
              <Button size="small" type={cluster.priority === 'P0' ? 'primary' : 'default'} onClick={() => handleAlertTriage('risk', cluster.scope)}>
                立即处理
              </Button>
            </div>

            <div className="health-issue-ai">
              <div className="health-alert-ai-block">
                <Sparkles size={15} />
                <span>{cluster.aiConclusion}</span>
              </div>
              <div className="health-risk-ebpf">
                <Network size={15} />
                <span>{cluster.ebpfRole}</span>
              </div>
            </div>

            <div className="health-issue-workspace">
              <div className="health-alert-panel">
                <div className="health-alert-panel-title">风险信号</div>
                <div className="health-alert-impact">{cluster.systemProblem}</div>
                <div className="health-alert-impact">{cluster.impact}</div>
                <div className="health-data-source-strip">
                  {cluster.dataSources.map((source) => (
                    <Tag key={source} bordered={false} color={source.includes('eBPF') ? 'blue' : 'default'}>
                      {source}
                    </Tag>
                  ))}
                </div>
              </div>
              <div className="health-alert-panel health-issue-action-panel">
                <div className="health-alert-panel-title">处置措施</div>
                <div className="health-alert-steps">
                  {cluster.actions.map((step, stepIndex) => (
                    <div className="health-alert-step" key={step}>
                      <b>{stepIndex + 1}</b>
                      <span>{step}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="health-alert-query-section">
              <div className="health-alert-panel-title">关联查询数据</div>
              <div className="health-issue-query-grid">
                {cluster.evidence.map((item) => (
                  <div className="health-alert-query-card" key={`${cluster.key}-${item.type}`}>
                    <div className="health-alert-query-type">{item.type}</div>
                    <code>{item.query}</code>
                    <span>{item.result}</span>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        ))}
      </div>
    </>
  )

  const impactSummary = {
    p0Count: filteredRiskClusters.filter((cluster) => cluster.priority === 'P0').length,
    p1Count: filteredRiskClusters.filter((cluster) => cluster.priority === 'P1').length,
    affectedObjectCount: new Set(filteredRiskClusters.flatMap((cluster) => cluster.scope.split('/').map((item) => item.trim()))).size,
  }

  const impactServices = [
    {
      key: 'payment',
      name: '支付交易',
      status: 'severe',
      score: 38,
      service: 'payment-api',
      signal: 'P99 2.8s / 5xx 5.4%',
      relation: '连接等待传导至订单同步',
      sources: ['APM', 'Trace', '日志', 'eBPF'],
    },
    {
      key: 'order',
      name: '订单履约',
      status: 'warning',
      score: 64,
      service: 'order-service',
      signal: '重试 +32% / 同步失败 0.8%',
      relation: '受支付重试放大影响',
      sources: ['APM', 'Trace', '告警'],
    },
    {
      key: 'recommendation',
      name: '推荐推理',
      status: 'warning',
      score: 71,
      service: 'recommendation-infer',
      signal: 'GPU 88% / 队列轻微上升',
      relation: '独立容量问题',
      sources: ['GPU', '进程', 'eBPF'],
    },
    {
      key: 'inventory',
      name: '库存预占',
      status: 'recovering',
      score: 76,
      service: 'inventory-service',
      signal: 'P99 920ms / read 42ms',
      relation: '恢复期观察',
      sources: ['Trace', '磁盘', 'eBPF'],
    },
  ]

  const selectedServiceFlow = selectedRisk?.key === 'payment-degradation'
    ? ['用户请求', 'checkout-web', 'payment-api', 'mysql-primary', 'order-service']
    : selectedRisk?.key === 'inference-capacity'
      ? ['推荐请求', 'recommendation-api', 'recommendation-infer', 'ai-infer-node-01', '模型缓存']
      : ['订单请求', 'order-service', 'inventory-service', 'prod-node-b', '磁盘读写']

  const selectedEvidenceGroups = selectedRisk
    ? [
        { label: '症状确认', value: selectedRisk.evidence[0]?.result ?? selectedRisk.systemProblem },
        { label: '链路定位', value: selectedRisk.evidence.find((item) => item.type.includes('Trace'))?.result ?? selectedRisk.scope },
        { label: 'eBPF 收敛', value: selectedRisk.ebpfFinding.narrowed },
        { label: '排除项', value: selectedRisk.ebpfFinding.excluded },
      ]
    : []

  const impactCommandOverview = (
    <div className="health-impact-command">
      <div className="health-impact-command-strip">
        <div className="health-impact-command-status">
          <Progress
            type="circle"
            percent={report.score}
            strokeColor={LEVEL_META[report.status].color}
            width={74}
            format={() => report.score}
          />
          <div>
            <span>AI 健康度</span>
            <b>{LEVEL_META[report.status].text}</b>
            <small>置信度 {report.confidence}% / 最近 15 分钟</small>
          </div>
        </div>
        <div className="health-impact-command-readout">
          <strong>支付交易链路拉低整体健康度，推荐推理容量与库存恢复期问题独立处置。</strong>
          <div>
            {SCORE_FACTORS.map((item) => (
              <span className={`health-compact-evidence-${item.tone}`} key={`score-strip-${item.key}`}>
                {item.label}<b>{item.value}</b>
              </span>
            ))}
          </div>
        </div>
        <div className="health-impact-command-kpis">
          <span>活跃问题 <b>{filteredRiskClusters.length}</b></span>
          <span>P0/P1 <b>{impactSummary.p0Count}/{impactSummary.p1Count}</b></span>
          <span>告警 <b>{report.activeAlertCount}</b></span>
          <span>对象 <b>{impactSummary.affectedObjectCount}</b></span>
        </div>
      </div>

      <div className="health-impact-service-wall">
        {impactServices.map((service) => (
          <div className={`health-impact-service-tile health-issue-${service.status}`} key={`impact-service-${service.key}`}>
            <div>
              <span className="health-status-dot" style={{ backgroundColor: LEVEL_META[service.status].color }} />
              <b>{service.name}</b>
              <strong>{service.score}</strong>
            </div>
            <small>{service.service}</small>
            <p>{service.signal}</p>
            <em>{service.relation}</em>
            <div>
              {service.sources.map((source) => (
                <Tag key={`${service.key}-${source}`} bordered={false} color={source === 'eBPF' ? 'blue' : 'default'}>
                  {source}
                </Tag>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="health-impact-command-main">
        <div className="health-impact-risk-queue">
          <div className="health-impact-command-section-head">
            <div>
              <span>Problem Feed</span>
              <b>活跃健康问题</b>
            </div>
            <Tag color="blue">AI 已归并</Tag>
          </div>
          {filteredRiskClusters.map((cluster, index) => (
            <div
              className={`health-impact-risk-row health-issue-${cluster.priority.toLowerCase()} ${
                selectedRisk?.key === cluster.key ? 'health-impact-risk-row-active' : ''
              }`}
              key={`command-risk-${cluster.key}`}
              onClick={() => setSelectedRiskKey(cluster.key)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' || event.key === ' ') setSelectedRiskKey(cluster.key)
              }}
              role="button"
              tabIndex={0}
            >
              <b>{index + 1}</b>
              <span className={`health-risk-priority health-risk-priority-${cluster.priority.toLowerCase()}`}>{cluster.priority}</span>
              <div className="health-impact-risk-main">
                <strong>{cluster.title}</strong>
                <small>{cluster.scope}</small>
              </div>
              <div><span>影响</span><strong>{cluster.scoreImpact}</strong></div>
              <div><span>阶段</span><strong>{cluster.progress}</strong></div>
              <div><span>数据</span><strong>{cluster.dataSources.slice(0, 3).join(' / ')}</strong></div>
              <Button
                size="small"
                type={cluster.priority === 'P0' ? 'primary' : 'default'}
                onClick={(event) => {
                  event.stopPropagation()
                  handleAlertTriage('risk', cluster.scope)
                }}
              >
                处置
              </Button>
            </div>
          ))}
        </div>

        <div className="health-impact-evidence-panel">
          {selectedRisk ? (
            <>
              <div className="health-impact-command-section-head">
                <div>
                  <span>AI Analysis</span>
                  <b>{selectedRisk.title}</b>
                </div>
                <Tag color="blue">置信度 {selectedRisk.confidence}%</Tag>
              </div>
              <div className="health-impact-evidence-summary-block">
                <strong>{getShortConclusion(selectedRisk.aiConclusion)}</strong>
                <span>{selectedRisk.mergeBasis}</span>
              </div>
              <div className="health-impact-analysis-chain">
                {selectedEvidenceGroups.map((item) => (
                  <div key={`selected-evidence-${item.label}`}>
                    <span>{item.label}</span>
                    <b>{item.value}</b>
                  </div>
                ))}
              </div>
              <div className="health-impact-query-strip">
                {selectedRisk.evidence.map((item) => {
                  const role = getEvidenceRole(item.type)
                  return (
                    <div key={`command-evidence-${selectedRisk.key}-${item.type}`}>
                      <Tag bordered={false} color={item.type.includes('eBPF') ? 'blue' : 'default'}>
                        {role.stage}
                      </Tag>
                      <strong>{item.type}</strong>
                      <span>{item.result}</span>
                    </div>
                  )
                })}
              </div>
            </>
          ) : (
            <div className="health-impact-evidence-summary-block">
              <strong>暂无匹配风险</strong>
              <span>调整过滤条件后查看 AI 归并问题和证据链。</span>
            </div>
          )}
        </div>
      </div>

      <div className="health-impact-boundary-map">
        <div className="health-impact-command-section-head">
          <div>
            <span>Dependency Path</span>
            <b>调用链与影响边界</b>
          </div>
          <Tag color="default">非 K8s 视角</Tag>
        </div>
        <div className="health-impact-path-board">
          <div className="health-impact-path-line">
            {selectedServiceFlow.map((node, index) => (
              <React.Fragment key={`flow-${selectedRisk?.key}-${node}`}>
                <div className={index === 2 ? 'health-impact-path-node health-impact-path-node-hot' : 'health-impact-path-node'}>
                  <span>{index === 0 ? '入口' : index === selectedServiceFlow.length - 1 ? '依赖' : '服务'}</span>
                  <b>{node}</b>
                </div>
                {index < selectedServiceFlow.length - 1 && (
                  <div className="health-impact-path-edge">
                    <i />
                    <small>{selectedRisk?.key === 'payment-degradation' && index === 2 ? 'eBPF: TCP 重传 +68%' : '调用'}</small>
                  </div>
                )}
              </React.Fragment>
            ))}
          </div>
          <div className="health-impact-path-notes">
            <div>
              <span>影响说明</span>
              <b>{selectedRisk?.impact}</b>
            </div>
            <div>
              <span>eBPF 作用</span>
              <b>{selectedRisk?.ebpfFinding.observed}</b>
            </div>
          </div>
        </div>
      </div>
    </div>
  )

  const impactPathVersion = (
    <div className="health-version-impact">
      <div className="health-impact-head">
        <div>
          <span>Business Impact View</span>
          <b>按影响面组织处置</b>
        </div>
        <div className="health-impact-score">
          <b>{report.score}</b>
          <span>AI 健康度</span>
        </div>
      </div>
      {impactCommandOverview}
      <div className="health-impact-detail-title">
        <span>Issue Details</span>
        <b>问题详情</b>
      </div>
      <div className="health-impact-grid">
        {filteredRiskClusters.map((cluster) => (
          <div className={`health-impact-card health-issue-${cluster.priority.toLowerCase()}`} key={cluster.key}>
            <div className="health-impact-card-top">
              <span className={`health-risk-priority health-risk-priority-${cluster.priority.toLowerCase()}`}>{cluster.priority}</span>
              <b>{cluster.title}</b>
              <StatusTag status={cluster.status} />
              <div className="health-impact-card-meta">
                <span>活跃告警 {cluster.activeAlerts}</span>
                <span>置信度 {cluster.confidence}%</span>
                <span>开始 {cluster.startedAt}</span>
              </div>
            </div>
            <div className="health-impact-route">
              <div>
                <strong>影响对象</strong>
                <span>{cluster.scope}</span>
              </div>
              <div>
                <strong>用户/业务表现</strong>
                <span>{cluster.impact}</span>
              </div>
            </div>
            <div className="health-impact-evidence">
              <div className="health-impact-evidence-head">
                <b>AI 推理证据</b>
                <span>{cluster.dataSources.length} 类数据源</span>
              </div>
              <div className="health-impact-evidence-summary">
                <div>
                  <span>{getShortConclusion(cluster.aiConclusion)}</span>
                </div>
                <div className="health-impact-source-row">
                  {cluster.dataSources.map((source) => (
                    <Tag key={`${cluster.key}-${source}`} bordered={false} color={source.includes('eBPF') ? 'blue' : 'default'}>
                      {source}
                    </Tag>
                  ))}
                </div>
              </div>
              <div className="health-impact-evidence-grid">
                {cluster.evidence.map((item) => {
                  const role = getEvidenceRole(item.type)
                  return (
                    <div
                      className={`health-impact-evidence-item ${
                        item.type.includes('eBPF') ? 'health-impact-evidence-ebpf' : ''
                      } health-impact-evidence-${role.kind}`}
                      key={`${cluster.key}-${item.type}`}
                    >
                      <div className="health-impact-evidence-title">
                        <Tag bordered={false} color={item.type.includes('eBPF') ? 'blue' : 'default'}>
                          {role.stage}
                        </Tag>
                        <strong>{item.type}</strong>
                      </div>
                      <div className="health-impact-evidence-result">{item.result}</div>
                      <span>{role.effect}</span>
                    </div>
                  )
                })}
              </div>
            </div>
            <div className="health-impact-actions">
              <div className="health-impact-actions-head">
                <div>
                  <div className="health-impact-actions-title">AI 处置计划</div>
                  <span>按影响范围、根因置信度和恢复验证顺序排序</span>
                </div>
                <Tag bordered={false} color={cluster.priority === 'P0' ? 'red' : 'blue'}>
                  {cluster.priority}
                </Tag>
              </div>
              {cluster.actionPlan.map((step, index) => (
                <div className="health-impact-action-step" key={`${cluster.key}-legacy-${step.phase}`}>
                  <b>{index + 1}</b>
                  <div>
                    <strong>{step.phase}</strong>
                    <span>{step.action}</span>
                  </div>
                </div>
              ))}
              <div className="health-impact-execute">
                <Button size="small" type={cluster.priority === 'P0' ? 'primary' : 'default'} onClick={() => handleAlertTriage('risk', cluster.scope)}>
                  查看关联告警
                </Button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )

  const riskOpsVersion = (
    <div className="health-version-impact">
      <div className="health-impact-head">
        <div>
          <span>值班入口</span>
          <b>系统健康处置</b>
        </div>
        <div className="health-impact-score">
          <b>{report.score}</b>
          <span>AI 健康度</span>
        </div>
      </div>

      <div className="health-ops-summary">
        <div className="health-ops-score">
          <div className="health-impact-overview-status">
            <span className="health-status-dot" style={{ backgroundColor: LEVEL_META[report.status].color }} />
            <b>{LEVEL_META[report.status].text}</b>
            <Tag color="error">健康度 {report.score}</Tag>
            <Tag color="blue">置信度 {report.confidence}%</Tag>
          </div>
          <div className="health-ops-audit">
            <span>分析窗口 Last 15 minutes</span>
            <span>更新时间 {selectedRisk?.updatedAt ?? '2026-06-15 10:31'}</span>
            <span>数据新鲜度 90s 内</span>
          </div>
        </div>
        <div className="health-ops-score-factors">
          <div className="health-impact-overview-title">健康度扣分构成</div>
          {SCORE_FACTORS.map((item) => (
            <div className={`health-ops-score-factor health-compact-evidence-${item.tone}`} key={item.key}>
              <span>{item.label}</span>
              <b>{item.value}</b>
            </div>
          ))}
        </div>
        <div className="health-ops-summary-text">
          <div className="health-impact-overview-title">AI 全局判断</div>
          <span>{report.summary}</span>
        </div>
      </div>

      {selectedRisk ? (
        <div className="health-ops-workbench">
          <div className="health-ops-queue">
            <div className="health-impact-overview-title">风险优先级</div>
            {filteredRiskClusters.map((cluster) => (
              <button
                className={`health-ops-queue-item ${selectedRisk.key === cluster.key ? 'health-ops-queue-item-active' : ''}`}
                key={cluster.key}
                type="button"
                onClick={() => setSelectedRiskKey(cluster.key)}
              >
                <span className={`health-risk-priority health-risk-priority-${cluster.priority.toLowerCase()}`}>{cluster.priority}</span>
                <div>
                  <b>{cluster.title}</b>
                  <small>{cluster.workflowStatus} / {cluster.owner}</small>
                </div>
                <strong>{cluster.scoreImpact}</strong>
              </button>
            ))}
          </div>

          <div className="health-ops-detail">
            <div className="health-ops-detail-head">
              <div>
                <div className="health-impact-primary-title">
                  <span className={`health-risk-priority health-risk-priority-${selectedRisk.priority.toLowerCase()}`}>
                    {selectedRisk.priority}
                  </span>
                  <b>{selectedRisk.title}</b>
                  <StatusTag status={selectedRisk.status} />
                </div>
                <div className="health-alert-event-meta">
                  <span>{selectedRisk.scope}</span>
                  <span>{selectedRisk.activeAlerts} 条活跃告警</span>
                  <span>开始 {selectedRisk.startedAt}</span>
                </div>
              </div>
              <Tag bordered={false} color="blue">AI 置信度 {selectedRisk.confidence}%</Tag>
            </div>

            <div className="health-ops-merge">
              <div>
                <strong>归并依据</strong>
                <span>{selectedRisk.mergeBasis}</span>
              </div>
              <div>
                <strong>业务影响</strong>
                <span>{selectedRisk.impact}</span>
              </div>
            </div>

            <div className="health-ops-confidence">
              <div className="health-impact-overview-title">置信度拆解</div>
              {selectedRisk.confidenceBreakdown.map((item) => (
                <span key={item}>{item}</span>
              ))}
            </div>

            <div className="health-ops-ebpf">
              <div className="health-impact-overview-title">eBPF 根因收敛</div>
              <div><strong>内核观测</strong><span>{selectedRisk.ebpfFinding.observed}</span></div>
              <div><strong>排除项</strong><span>{selectedRisk.ebpfFinding.excluded}</span></div>
              <div><strong>根因指向</strong><span>{selectedRisk.ebpfFinding.narrowed}</span></div>
            </div>

            <div className="health-ops-evidence">
              <div className="health-impact-evidence-head">
                <b>关键证据</b>
                <span>引用 {selectedRisk.dataSources.length} 类数据源</span>
              </div>
              <div className="health-impact-evidence-grid">
                {selectedRisk.evidence.map((item) => {
                  const role = getEvidenceRole(item.type)
                  return (
                    <div
                      className={`health-impact-evidence-item ${
                        item.type.includes('eBPF') ? 'health-impact-evidence-ebpf' : ''
                      } health-impact-evidence-${role.kind}`}
                      key={`${selectedRisk.key}-${item.type}`}
                    >
                      <div className="health-impact-evidence-title">
                        <Tag bordered={false} color={item.type.includes('eBPF') ? 'blue' : 'default'}>
                          {role.stage}
                        </Tag>
                        <strong>{item.type}</strong>
                      </div>
                      <div className="health-impact-evidence-result">{item.result}</div>
                      <span>{role.effect}</span>
                    </div>
                  )
                })}
              </div>
            </div>

            <div className="health-ops-data-quality">
              <div className="health-impact-overview-title">数据边界</div>
              {selectedRisk.dataQuality.map((item) => (
                <span key={item}>{item}</span>
              ))}
            </div>
          </div>

          <div className="health-ops-actions">
            <div className="health-impact-actions-head">
              <div>
                <div className="health-impact-actions-title">处置闭环</div>
                <span>{selectedRisk.owner} / {selectedRisk.sla}</span>
              </div>
              <Tag bordered={false} color={selectedRisk.status === 'recovering' ? 'blue' : 'red'}>
                {selectedRisk.workflowStatus}
              </Tag>
            </div>
            <div className="health-ops-progress">
              <span>当前阶段</span>
              <b>{selectedRisk.progress}</b>
            </div>
            {selectedRisk.actionPlan.map((step, index) => (
              <div className="health-ops-action-step" key={`${selectedRisk.key}-${step.phase}`}>
                <b>{index + 1}</b>
                <div>
                  <strong>{step.phase}</strong>
                  <span>{step.target}</span>
                  <p>{step.action}</p>
                  <small>验证：{step.validation}</small>
                  <small>风险：{step.risk}</small>
                </div>
              </div>
            ))}
            <div className="health-ops-action-buttons">
              <Button size="small" type={selectedRisk.priority === 'P0' ? 'primary' : 'default'} onClick={() => handleAlertTriage('risk', selectedRisk.scope)}>
                查看关联告警
              </Button>
              <Button size="small" onClick={() => handleEvidenceDetail(selectedRisk.scope)}>
                查看证据详情
              </Button>
            </div>
          </div>
        </div>
      ) : (
        <Card className="health-card">未找到匹配健康问题</Card>
      )}
    </div>
  )

  const consultRoomVersion = (
    <div className="health-version-consult">
      <div className="health-consult-main">
        <div className="health-consult-avatar">
          <Bot size={28} />
          <b>AI 会诊室</b>
          <span>多源观测联合推理</span>
        </div>
        <div className="health-consult-dialog">
          <div className="health-consult-message health-consult-message-user">
            <b>值班视角</b>
            <span>现在最需要确认哪些问题，先处理什么？</span>
          </div>
          <div className="health-consult-message health-consult-message-ai">
            <b>AI 判断</b>
            <span>{report.summary}</span>
          </div>
          {filteredRiskClusters.map((cluster) => (
            <div className={`health-consult-diagnosis health-issue-${cluster.priority.toLowerCase()}`} key={cluster.key}>
              <div>
                <span className={`health-risk-priority health-risk-priority-${cluster.priority.toLowerCase()}`}>{cluster.priority}</span>
                <b>{cluster.title}</b>
                <StatusTag status={cluster.status} />
              </div>
              <p>{cluster.aiConclusion}</p>
              <div className="health-consult-proof">
                {cluster.evidence.slice(0, 3).map((item) => (
                  <span key={item.type}>{item.type}: {item.result}</span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
      <div className="health-consult-playbook">
        <b>Playbook</b>
        {filteredRiskClusters.map((cluster) => (
          <div key={cluster.key}>
            <strong>{cluster.priority} {cluster.title}</strong>
            <span>{cluster.actions[0]}</span>
          </div>
        ))}
      </div>
    </div>
  )

  const warRoomVersion = (
    <div className="health-version-warroom">
      <div className="health-warroom-hero">
        <div>
          <span>War Room</span>
          <b>{report.score}</b>
          <p>系统处于严重退化，核心风险集中在支付链路，另有推理容量与库存恢复两个独立作战面。</p>
        </div>
        <div className="health-warroom-kpis">
          <span>健康问题 <b>{filteredRiskClusters.length}</b></span>
          <span>活跃告警 <b>{report.activeAlertCount}</b></span>
          <span>AI 置信度 <b>{report.confidence}%</b></span>
        </div>
      </div>
      <div className="health-warroom-grid">
        <div className="health-warroom-panel health-warroom-panel-wide">
          <b>Live Incident Board</b>
          {filteredRiskClusters.map((cluster) => (
            <div className={`health-warroom-row health-issue-${cluster.priority.toLowerCase()}`} key={cluster.key}>
              <span className={`health-risk-priority health-risk-priority-${cluster.priority.toLowerCase()}`}>{cluster.priority}</span>
              <strong>{cluster.title}</strong>
              <span>{cluster.scope}</span>
              <Button size="small" type={cluster.priority === 'P0' ? 'primary' : 'default'} onClick={() => handleAlertTriage('risk', cluster.scope)}>
                处理
              </Button>
            </div>
          ))}
        </div>
        <div className="health-warroom-panel">
          <b>eBPF Signals</b>
          <span>TCP 重传 +68%</span>
          <span>connect P99 180ms</span>
          <span>read 长尾 42ms</span>
          <span>run queue 11.4</span>
        </div>
        <div className="health-warroom-panel">
          <b>Next Actions</b>
          {filteredRiskClusters.map((cluster) => (
            <span key={cluster.key}>{cluster.priority}: {cluster.actions[0]}</span>
          ))}
        </div>
      </div>
    </div>
  )

  const tabs = [
    { key: 'current', label: '版本A 综合报告', children: currentVersion },
    { key: 'impact-path', label: '版本B 影响路径', children: impactPathVersion },
    { key: 'risk-ops', label: '版本C 系统健康处置', children: riskOpsVersion },
    { key: 'consult-room', label: '版本D AI会诊室', children: consultRoomVersion },
    { key: 'war-room', label: '版本E 战情屏', children: warRoomVersion },
  ]

  return (
    <div className="system-health-page">
      <div className="health-report-header">
        <div className="health-report-main">
          <div className="health-eyebrow">
            <Bot size={15} />
            <span>AI Ops Report</span>
          </div>
          <div className="health-title-row">
            <div className="health-title">AI 系统健康度分析报告</div>
            <StatusTag status={report.status} />
          </div>
          <div className="health-subtitle">
            融合 APM、eBPF、日志、Trace、基础设施与实时告警数据，归并当前风险并给出系统健康判断和运维优先级。
          </div>
          <div className="health-report-meta">
            <span>时间窗口：Last 15 minutes</span>
            <span>证据源：APM / eBPF / Infra / Logs / Alerts</span>
            <span>分析方式：问题归并 / 根因定位 / 处置排序</span>
          </div>
        </div>
        <Space wrap className="health-report-actions">
          <Input.Search
            allowClear
            placeholder="过滤问题 / 对象"
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            style={{ width: 220 }}
          />
        </Space>
      </div>

      <Tabs className="health-version-tabs" items={tabs} destroyInactiveTabPane={false} />
    </div>
  )
}
