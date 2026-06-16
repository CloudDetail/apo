import React, { useMemo, useState } from 'react'
import PropTypes from 'prop-types'
import { Button, Progress, Segmented, Space, Table, Tag } from 'antd'
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  Bell,
  CircuitBoard,
  Clock3,
  Database,
  FileSearch,
  Gauge,
  GitBranch,
  ListFilter,
  Network,
  ShieldCheck,
} from 'lucide-react'
import './index.css'

const severityWeight = {
  critical: 16,
  error: 10,
  warning: 5,
  info: 2,
}

const statusMeta = {
  healthy: { label: '健康', color: 'success', dot: 'green' },
  warning: { label: '注意', color: 'warning', dot: 'yellow' },
  degraded: { label: '风险', color: 'orange', dot: 'orange' },
  critical: { label: '严重', color: 'error', dot: 'red' },
  unknown: { label: 'Unknown', color: 'default', dot: 'gray' },
}

const services = [
  {
    name: 'payment-api',
    domain: '交易域',
    availability: 99.6,
    availabilityTarget: 99.95,
    p95: 890,
    p95Target: 500,
    errorRate: 0.42,
    errorTarget: 0.1,
    tps: 188.3,
    logErrors: 37,
    deployRisk: 18,
    lastAbnormal: '14:23',
    trend: [91, 89, 86, 82, 77, 72, 58],
  },
  {
    name: 'order-service',
    domain: '订单域',
    availability: 99.84,
    availabilityTarget: 99.9,
    p95: 430,
    p95Target: 420,
    errorRate: 0.18,
    errorTarget: 0.1,
    tps: 475.46,
    logErrors: 14,
    deployRisk: 6,
    lastAbnormal: '14:19',
    trend: [88, 84, 82, 79, 76, 73, 72],
  },
  {
    name: 'ts-travel-service',
    domain: '出行域',
    availability: 99.98,
    availabilityTarget: 99.9,
    p95: 252.89,
    p95Target: 500,
    errorRate: 0,
    errorTarget: 0.1,
    tps: 89.06,
    logErrors: 1,
    deployRisk: 2,
    lastAbnormal: '-',
    trend: [94, 94, 95, 94, 96, 95, 95],
  },
  {
    name: 'user-service',
    domain: '用户域',
    availability: 99.99,
    availabilityTarget: 99.9,
    p95: 120,
    p95Target: 400,
    errorRate: 0,
    errorTarget: 0.1,
    tps: 320.12,
    logErrors: 0,
    deployRisk: 0,
    lastAbnormal: '-',
    trend: [97, 97, 98, 97, 98, 98, 98],
  },
  {
    name: 'mq-settlement',
    domain: '结算域',
    availability: 99.9,
    availabilityTarget: 99.9,
    p95: 620,
    p95Target: 700,
    errorRate: 0.03,
    errorTarget: 0.1,
    tps: 42.2,
    logErrors: 8,
    deployRisk: 4,
    lastAbnormal: '14:31',
    trend: [86, 84, 82, 80, 79, 78, 79],
  },
]

const alerts = [
  {
    id: 'INC-1401',
    service: 'payment-api',
    level: 'critical',
    title: 'P99 延迟超过基线 2.4 倍',
    duration: '18m',
    status: 'firing',
    impact: '支付链路 3.2% 请求变慢',
    source: 'APM / Trace',
  },
  {
    id: 'INC-1402',
    service: 'payment-api',
    level: 'error',
    title: '5xx 错误率高于 SLO 阈值',
    duration: '14m',
    status: 'firing',
    impact: '支付成功率下降 0.9%',
    source: '指标 / 告警',
  },
  {
    id: 'INC-1403',
    service: 'order-service',
    level: 'warning',
    title: '订单接口错误率同比上升',
    duration: '22m',
    status: 'firing',
    impact: '订单查询轻微抖动',
    source: '指标',
  },
  {
    id: 'INC-1404',
    service: 'mq-settlement',
    level: 'warning',
    title: '结算队列积压恢复中',
    duration: '31m',
    status: 'resolved',
    impact: '结算延迟风险降低',
    source: 'MQ / 日志',
  },
]

const clamp = (value, min = 0, max = 100) => Math.max(min, Math.min(max, value))

const scoreByRatio = (actual, target, direction = 'lower') => {
  if (actual === null || actual === undefined || target === null || target === undefined) return null
  if (direction === 'higher') {
    if (actual >= target) return 100
    return clamp(100 - ((target - actual) / target) * 120)
  }
  if (actual <= target) return 100
  return clamp(100 - ((actual - target) / target) * 120)
}

const getLevel = (score) => {
  if (score === null || score === undefined) return 'unknown'
  if (score < 60) return 'critical'
  if (score < 75) return 'degraded'
  if (score < 90) return 'warning'
  return 'healthy'
}

const computeServiceHealth = (service, serviceAlerts) => {
  const availabilityScore = scoreByRatio(
    service.availability,
    service.availabilityTarget,
    'higher',
  )
  const latencyScore = scoreByRatio(service.p95, service.p95Target, 'lower')
  const errorScore = scoreByRatio(service.errorRate, service.errorTarget, 'lower')
  const alertPenalty = serviceAlerts.reduce(
    (sum, alert) => sum + (severityWeight[alert.level] ?? 0),
    0,
  )
  const alertScore = clamp(100 - alertPenalty)
  const logScore = clamp(100 - service.logErrors * 2)
  const deployScore = clamp(100 - service.deployRisk)

  const dimensions = {
    availability: availabilityScore,
    latency: latencyScore,
    errorRate: errorScore,
    alerts: alertScore,
    logs: logScore,
    deploy: deployScore,
  }

  const score = Math.round(
    availabilityScore * 0.24 +
      latencyScore * 0.24 +
      errorScore * 0.2 +
      alertScore * 0.18 +
      logScore * 0.08 +
      deployScore * 0.06,
  )

  const reasons = Object.entries(dimensions)
    .map(([key, value]) => ({ key, value }))
    .sort((a, b) => a.value - b.value)
    .slice(0, 3)

  return {
    score,
    level: getLevel(score),
    dimensions,
    reasons,
    activeAlerts: serviceAlerts.filter((alert) => alert.status === 'firing').length,
  }
}

const StatusPill = ({ level }) => {
  const meta = statusMeta[level] ?? statusMeta.unknown
  return (
    <span className="hl-status-pill">
      <span className={`hl-dot hl-dot-${meta.dot}`} />
      <Tag color={meta.color} bordered={false}>
        {meta.label}
      </Tag>
    </span>
  )
}

StatusPill.propTypes = {
  level: PropTypes.string,
}

const Panel = ({ title, subtitle, icon, children, className = '' }) => (
  <section className={`hl-panel ${className}`}>
    <div className="hl-panel-head">
      <div className="hl-panel-title">
        {icon}
        <span>{title}</span>
      </div>
      {subtitle && <span className="hl-panel-subtitle">{subtitle}</span>}
    </div>
    <div className="hl-panel-body">{children}</div>
  </section>
)

Panel.propTypes = {
  title: PropTypes.string.isRequired,
  subtitle: PropTypes.string,
  icon: PropTypes.node,
  children: PropTypes.node,
  className: PropTypes.string,
}

const Sparkline = ({ values, tone = 'blue' }) => {
  const width = 120
  const height = 32
  const min = Math.min(...values)
  const max = Math.max(...values)
  const range = max - min || 1
  const points = values
    .map((value, index) => {
      const x = (index / (values.length - 1)) * width
      const y = height - ((value - min) / range) * (height - 6) - 3
      return `${x},${y}`
    })
    .join(' ')

  return (
    <svg className={`hl-sparkline hl-sparkline-${tone}`} viewBox={`0 0 ${width} ${height}`}>
      <polyline points={points} fill="none" strokeWidth="2" strokeLinecap="round" />
    </svg>
  )
}

Sparkline.propTypes = {
  values: PropTypes.arrayOf(PropTypes.number).isRequired,
  tone: PropTypes.string,
}

const DimensionBar = ({ label, value }) => {
  const level = getLevel(value)
  return (
    <div className="hl-dimension-row">
      <span>{label}</span>
      <Progress
        percent={Math.round(value)}
        showInfo={false}
        size="small"
        strokeColor={
          level === 'critical'
            ? '#f04438'
            : level === 'degraded'
              ? '#f79009'
              : level === 'warning'
                ? '#fdb022'
                : '#12b76a'
        }
      />
      <b>{Math.round(value)}</b>
    </div>
  )
}

DimensionBar.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.number.isRequired,
}

const formatPct = (value) => `${value.toFixed(value >= 10 ? 2 : 3)}%`

export default function HealthLabPage() {
  const [scope, setScope] = useState('服务与告警')
  const showServicePanels = scope !== '告警视角'
  const showAlertPanels = scope !== '服务视角'

  const serviceModels = useMemo(
    () =>
      services.map((service) => ({
        ...service,
        health: computeServiceHealth(
          service,
          alerts.filter((alert) => alert.service === service.name),
        ),
      })),
    [],
  )

  const summary = useMemo(() => {
    const score = Math.round(
      serviceModels.reduce((sum, service) => sum + service.health.score, 0) /
        serviceModels.length,
    )
    const abnormal = serviceModels.filter((service) => service.health.level !== 'healthy')
    const activeAlerts = alerts.filter((alert) => alert.status === 'firing')
    const criticalAlerts = activeAlerts.filter((alert) => alert.level === 'critical').length

    return {
      score,
      level: getLevel(score),
      abnormalCount: abnormal.length,
      activeAlerts: activeAlerts.length,
      criticalAlerts,
      topService: abnormal.sort((a, b) => a.health.score - b.health.score)[0],
    }
  }, [serviceModels])

  const serviceColumns = [
    {
      title: '服务',
      dataIndex: 'name',
      width: 180,
      fixed: 'left',
      render: (value, record) => (
        <div>
          <Button type="link" size="small" className="hl-link">
            {value}
            <ArrowRight size={13} />
          </Button>
          <div className="hl-table-subtext">{record.domain}</div>
        </div>
      ),
    },
    {
      title: '状态',
      width: 100,
      render: (_, record) => <StatusPill level={record.health.level} />,
    },
    {
      title: '健康分',
      width: 110,
      sorter: (a, b) => a.health.score - b.health.score,
      render: (_, record) => (
        <div className="hl-score-cell">
          <Progress
            percent={record.health.score}
            size="small"
            strokeColor={record.health.score < 75 ? '#f79009' : '#12b76a'}
          />
          <span>{record.health.score}</span>
        </div>
      ),
    },
    {
      title: '可用率',
      width: 120,
      render: (_, record) => (
        <MetricCell
          value={formatPct(record.availability)}
          day={record.availability >= record.availabilityTarget ? '+0.02%' : '-0.18%'}
          week={record.availability >= record.availabilityTarget ? '+0.04%' : '-0.31%'}
        />
      ),
    },
    {
      title: 'P95 延迟',
      width: 130,
      render: (_, record) => (
        <MetricCell
          value={`${record.p95}ms`}
          day={record.p95 > record.p95Target ? '+8.04%' : '-0.28%'}
          week={record.p95 > record.p95Target ? '+12.88%' : '+1.77%'}
          danger={record.p95 > record.p95Target}
        />
      ),
    },
    {
      title: '错误率',
      width: 120,
      render: (_, record) => (
        <MetricCell
          value={formatPct(record.errorRate)}
          day={record.errorRate > record.errorTarget ? '+42.1%' : '0%'}
          week={record.errorRate > record.errorTarget ? '+88.7%' : '0%'}
          danger={record.errorRate > record.errorTarget}
        />
      ),
    },
    {
      title: '告警',
      width: 90,
      render: (_, record) => (
        <Tag color={record.health.activeAlerts ? 'error' : 'success'} bordered={false}>
          {record.health.activeAlerts}
        </Tag>
      ),
    },
    {
      title: '趋势',
      width: 140,
      render: (_, record) => (
        <Sparkline
          values={record.trend}
          tone={record.health.level === 'healthy' ? 'green' : 'orange'}
        />
      ),
    },
    {
      title: '最近异常',
      width: 110,
      dataIndex: 'lastAbnormal',
    },
  ]

  const alertColumns = [
    {
      title: '等级',
      dataIndex: 'level',
      width: 92,
      render: (value) => (
        <Tag color={value === 'critical' ? 'error' : value === 'error' ? 'volcano' : 'orange'}>
          {value}
        </Tag>
      ),
    },
    {
      title: '事件',
      dataIndex: 'title',
      render: (value, record) => (
        <div>
          <Button type="link" size="small" className="hl-link">
            {value}
            <ArrowRight size={13} />
          </Button>
          <div className="hl-table-subtext">
            {record.id} / {record.source}
          </div>
        </div>
      ),
    },
    {
      title: '服务',
      dataIndex: 'service',
      width: 140,
    },
    {
      title: '持续',
      dataIndex: 'duration',
      width: 80,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (value) => <Tag color={value === 'firing' ? 'error' : 'success'}>{value}</Tag>,
    },
    {
      title: '影响',
      dataIndex: 'impact',
    },
  ]

  return (
    <div className="hl-page">
      <div className="hl-toolbar">
        <div>
          <div className="hl-title">系统健康度组件实验室</div>
          <div className="hl-subtitle">
            面向 APO 服务概览和告警事件的基础组件，聚焦可解释健康分、服务状态和告警影响。
          </div>
        </div>
        <Space wrap>
          <Segmented
            value={scope}
            onChange={setScope}
            options={['服务与告警', '服务视角', '告警视角']}
          />
          <Tag color="blue">Last 15 minutes</Tag>
          <Tag>数据组 ALL</Tag>
        </Space>
      </div>

      <div className="hl-grid hl-grid-top">
        <Panel title="系统健康总览" subtitle="规则计算，不由 LLM 打分" icon={<Gauge size={16} />}>
          <div className="hl-overview">
            <div className={`hl-score-ring hl-score-ring-${summary.level}`}>
              <span>{summary.score}</span>
              <small>/100</small>
            </div>
            <div className="hl-overview-content">
              <StatusPill level={summary.level} />
              <div className="hl-overview-line">
                <span>异常服务</span>
                <b>{summary.abnormalCount}</b>
              </div>
              <div className="hl-overview-line">
                <span>活跃告警</span>
                <b>{summary.activeAlerts}</b>
              </div>
              <div className="hl-overview-line">
                <span>Critical</span>
                <b>{summary.criticalAlerts}</b>
              </div>
            </div>
          </div>
        </Panel>

        <Panel title="健康分解释" subtitle="TopN 扣分维度" icon={<ShieldCheck size={16} />}>
          {summary.topService?.health.reasons.map((reason) => (
            <DimensionBar
              key={reason.key}
              label={reason.key}
              value={reason.value}
            />
          ))}
          <div className="hl-note">
            当前主风险来自 <b>{summary.topService?.name}</b>。解释输入只取异常服务 TopN 和扣分维度
            TopN。
          </div>
        </Panel>

        {showAlertPanels && (
          <Panel title="告警事件流" subtitle="服务关联事件" icon={<Bell size={16} />}>
            <div className="hl-alert-stream">
              {alerts.slice(0, 4).map((alert) => (
                <div className="hl-alert-card" key={alert.id}>
                  <span className={`hl-dot hl-dot-${alert.level === 'critical' ? 'red' : 'orange'}`} />
                  <div>
                    <div className="hl-alert-title">{alert.title}</div>
                    <div className="hl-table-subtext">
                      {alert.service} / {alert.duration} / {alert.impact}
                    </div>
                  </div>
                  <Tag color={alert.status === 'firing' ? 'error' : 'success'}>{alert.status}</Tag>
                </div>
              ))}
            </div>
          </Panel>
        )}
      </div>

      <div className="hl-grid hl-grid-main">
        {showServicePanels && (
          <Panel
            className="hl-wide"
            title="服务健康度表格"
            subtitle="贴近 APO 服务页密度"
            icon={<CircuitBoard size={16} />}
          >
            <Table
              rowKey="name"
              columns={serviceColumns}
              dataSource={serviceModels}
              size="small"
              pagination={false}
              scroll={{ x: 1040 }}
            />
          </Panel>
        )}

        {showAlertPanels && (
          <Panel title="告警等级分布" subtitle="结构判断，不替代列表" icon={<AlertTriangle size={16} />}>
            <Distribution
              items={[
                { label: 'Critical', value: 1, tone: 'red' },
                { label: 'Error', value: 1, tone: 'volcano' },
                { label: 'Warning', value: 2, tone: 'orange' },
                { label: 'Resolved', value: 1, tone: 'green' },
              ]}
            />
          </Panel>
        )}

        {showServicePanels && (
          <Panel title="异常服务影响摘要" subtitle="服务-接口-依赖" icon={<Network size={16} />}>
            <ImpactList />
          </Panel>
        )}

        {showAlertPanels && (
          <Panel
            className="hl-wide"
            title="告警事件明细"
            subtitle="事件可回跳服务和证据"
            icon={<ListFilter size={16} />}
          >
            <Table
              rowKey="id"
              columns={alertColumns}
              dataSource={alerts}
              size="small"
              pagination={false}
            />
          </Panel>
        )}

        <Panel title="数据与 AI 边界" subtitle="可解释约束" icon={<Database size={16} />}>
          <div className="hl-boundary-list">
            <BoundaryItem title="健康分来源" text="规则计算：可用率、延迟、错误率、告警、日志错误、变更风险。" />
            <BoundaryItem title="LLM 输入" text="仅允许异常服务 TopN、告警聚合 TopN、Trace 样本 TopN、扣分维度 TopN。" />
            <BoundaryItem title="缺失数据" text="显示 Unknown / 数据不足，不推断为健康，不生成原因。" />
          </div>
        </Panel>

        <Panel title="建议动作入口" subtitle="处理闭环" icon={<FileSearch size={16} />}>
          <div className="hl-action-list">
            <ActionItem priority="P0" text="查看 14:20 payment-api 发布记录" action="查看变更" icon={<GitBranch size={14} />} />
            <ActionItem priority="P0" text="检查 order-db 慢 SQL Top5" action="查看 SQL" icon={<Database size={14} />} />
            <ActionItem priority="P1" text="打开支付链路异常 Trace" action="查看 Trace" icon={<Activity size={14} />} />
            <ActionItem priority="P1" text="创建事件并通知负责人" action="创建事件" icon={<Clock3 size={14} />} />
          </div>
        </Panel>
      </div>
    </div>
  )
}

const MetricCell = ({ value, day, week, danger = false }) => (
  <div className="hl-metric-cell">
    <b className={danger ? 'hl-danger-text' : ''}>{value}</b>
    <span>日同比 {day}</span>
    <span>周同比 {week}</span>
  </div>
)

MetricCell.propTypes = {
  value: PropTypes.string.isRequired,
  day: PropTypes.string.isRequired,
  week: PropTypes.string.isRequired,
  danger: PropTypes.bool,
}

const Distribution = ({ items }) => {
  const total = items.reduce((sum, item) => sum + item.value, 0)
  return (
    <div className="hl-distribution">
      <div className="hl-stack">
        {items.map((item) => (
          <span
            key={item.label}
            className={`hl-stack-part hl-stack-${item.tone}`}
            style={{ width: `${(item.value / total) * 100}%` }}
          />
        ))}
      </div>
      {items.map((item) => (
        <div className="hl-dist-row" key={item.label}>
          <span>
            <span className={`hl-dot hl-dot-${item.tone === 'red' ? 'red' : item.tone === 'green' ? 'green' : 'orange'}`} />
            {item.label}
          </span>
          <b>{item.value}</b>
        </div>
      ))}
    </div>
  )
}

Distribution.propTypes = {
  items: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string.isRequired,
      value: PropTypes.number.isRequired,
      tone: PropTypes.string.isRequired,
    }),
  ).isRequired,
}

const ImpactList = () => (
  <div className="hl-impact-list">
    <div className="hl-impact-row">
      <Tag color="error">接口</Tag>
      <span>/api/v1/pay</span>
      <Button type="link" size="small">Trace</Button>
    </div>
    <div className="hl-impact-row">
      <Tag color="warning">实例</Tag>
      <span>payment-api-7d9</span>
      <Button type="link" size="small">日志</Button>
    </div>
    <div className="hl-impact-row">
      <Tag color="warning">依赖</Tag>
      <span>order-db</span>
      <Button type="link" size="small">SQL</Button>
    </div>
    <div className="hl-impact-row">
      <Tag color="blue">数据组</Tag>
      <span>production / ALL</span>
      <Button type="link" size="small">筛选</Button>
    </div>
  </div>
)

const BoundaryItem = ({ title, text }) => (
  <div className="hl-boundary-item">
    <b>{title}</b>
    <span>{text}</span>
  </div>
)

BoundaryItem.propTypes = {
  title: PropTypes.string.isRequired,
  text: PropTypes.string.isRequired,
}

const ActionItem = ({ priority, text, action, icon }) => (
  <div className="hl-action-item">
    <Tag color={priority === 'P0' ? 'error' : 'warning'}>{priority}</Tag>
    <span>{text}</span>
    <Button size="small" icon={icon}>
      {action}
    </Button>
  </div>
)

ActionItem.propTypes = {
  priority: PropTypes.string.isRequired,
  text: PropTypes.string.isRequired,
  action: PropTypes.string.isRequired,
  icon: PropTypes.node,
}
