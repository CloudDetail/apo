import React, { useMemo, useState } from 'react'
import PropTypes from 'prop-types'
import { Button, Input, Select, Space, Tag, Tooltip } from 'antd'
import { ArrowLeft, HelpCircle, Search, Sparkles } from 'lucide-react'
import './index.css'

const metricLabels = [
  ['resource', '资源水位'],
  ['resourceShift', '资源波动'],
  ['errorRate', '错误率'],
  ['latency', '响应延时'],
  ['errorLog', '错误日志'],
  ['badRequests', '异常请求'],
  ['lockCost', '锁等待'],
]

const statusMeta = {
  green: { label: '正常', color: '#12b76a' },
  yellow: { label: '关注', color: '#f79009' },
  red: { label: '异常', color: '#f04438' },
  gray: { label: '无数据', color: '#98a2b3' },
}

const metricSamples = {
  resource: {
    green: ['43%', 'CPU / 内存处于可用区间'],
    yellow: ['74%', '接近资源阈值'],
    red: ['91%', '资源使用已超过阈值'],
    gray: ['--', '暂无资源数据'],
  },
  resourceShift: {
    green: ['8%', '波动低于基线'],
    yellow: ['31%', '波动接近异常阈值'],
    red: ['57%', '短时间剧烈波动'],
    gray: ['--', '暂无波动数据'],
  },
  errorRate: {
    green: ['0.12%', '错误率稳定'],
    yellow: ['1.9%', '错误率高于基线'],
    red: ['5.7%', '错误率超过异常阈值'],
    gray: ['--', '暂无请求数据'],
  },
  latency: {
    green: ['128ms', 'P95 响应延时稳定'],
    yellow: ['420ms', 'P95 响应延时升高'],
    red: ['960ms', 'P95 响应延时明显异常'],
    gray: ['--', '暂无响应延时数据'],
  },
  errorLog: {
    green: ['12', '错误日志未异常增长'],
    yellow: ['146', '错误日志高于基线'],
    red: ['621', '错误日志激增'],
    gray: ['--', '日志源未上报'],
  },
  badRequests: {
    green: ['18', '异常请求稳定'],
    yellow: ['233', '异常请求升高'],
    red: ['874', '异常请求集中出现'],
    gray: ['--', '暂无请求明细'],
  },
  lockCost: {
    green: ['24ms', '锁等待稳定'],
    yellow: ['188ms', '锁等待升高'],
    red: ['740ms', '锁等待严重异常'],
    gray: ['--', '暂无锁等待数据'],
  },
}

const systems = [
  {
    id: 'order',
    name: '订单系统',
    score: 38,
    metrics: {
      resource: 'green',
      resourceShift: 'yellow',
      errorRate: 'red',
      latency: 'yellow',
      errorLog: 'red',
      badRequests: 'yellow',
      lockCost: 'green',
    },
    services: [
      {
        name: 'order-api',
        desc: '核心下单接口',
        score: 35,
        metrics: {
          resource: 'green',
          resourceShift: 'yellow',
          errorRate: 'red',
          latency: 'yellow',
          errorLog: 'red',
          badRequests: 'yellow',
          lockCost: 'green',
        },
      },
      {
        name: 'order-callback',
        desc: '支付结果回调',
        score: 46,
        metrics: {
          resource: 'green',
          resourceShift: 'green',
          errorRate: 'yellow',
          latency: 'green',
          errorLog: 'red',
          badRequests: 'yellow',
          lockCost: 'green',
        },
      },
      {
        name: 'order-scheduler',
        desc: '超时订单处理',
        score: 61,
        metrics: {
          resource: 'green',
          resourceShift: 'yellow',
          errorRate: 'green',
          latency: 'yellow',
          errorLog: 'green',
          badRequests: 'green',
          lockCost: 'red',
        },
      },
    ],
  },
  {
    id: 'payment',
    name: '支付系统',
    score: 52,
    metrics: {
      resource: 'red',
      resourceShift: 'yellow',
      errorRate: 'yellow',
      latency: 'red',
      errorLog: 'yellow',
      badRequests: 'green',
      lockCost: 'green',
    },
    services: [
      {
        name: 'payment-gateway',
        desc: '收银台与支付网关',
        score: 44,
        metrics: {
          resource: 'red',
          resourceShift: 'yellow',
          errorRate: 'yellow',
          latency: 'red',
          errorLog: 'yellow',
          badRequests: 'green',
          lockCost: 'green',
        },
      },
      {
        name: 'refund-worker',
        desc: '退款异步任务',
        score: 67,
        metrics: {
          resource: 'green',
          resourceShift: 'green',
          errorRate: 'green',
          latency: 'yellow',
          errorLog: 'green',
          badRequests: 'green',
          lockCost: 'yellow',
        },
      },
    ],
  },
  {
    id: 'promotion',
    name: '营销系统',
    score: 64,
    metrics: {
      resource: 'yellow',
      resourceShift: 'red',
      errorRate: 'green',
      latency: 'yellow',
      errorLog: 'green',
      badRequests: 'green',
      lockCost: 'green',
    },
    services: [
      {
        name: 'coupon-center',
        desc: '优惠券中心',
        score: 58,
        metrics: {
          resource: 'yellow',
          resourceShift: 'red',
          errorRate: 'green',
          latency: 'yellow',
          errorLog: 'green',
          badRequests: 'green',
          lockCost: 'green',
        },
      },
    ],
  },
  {
    id: 'member',
    name: '会员系统',
    score: 71,
    metrics: {
      resource: 'green',
      resourceShift: 'green',
      errorRate: 'green',
      latency: 'yellow',
      errorLog: 'yellow',
      badRequests: 'green',
      lockCost: 'red',
    },
    services: [
      {
        name: 'member-profile',
        desc: '会员资料服务',
        score: 69,
        metrics: {
          resource: 'green',
          resourceShift: 'green',
          errorRate: 'green',
          latency: 'yellow',
          errorLog: 'yellow',
          badRequests: 'green',
          lockCost: 'red',
        },
      },
    ],
  },
  {
    id: 'report',
    name: '报表系统',
    score: 79,
    metrics: {
      resource: 'green',
      resourceShift: 'green',
      errorRate: 'green',
      latency: 'green',
      errorLog: 'gray',
      badRequests: 'green',
      lockCost: 'yellow',
    },
    services: [
      {
        name: 'export-worker',
        desc: '报表导出任务',
        score: 72,
        metrics: {
          resource: 'green',
          resourceShift: 'green',
          errorRate: 'green',
          latency: 'green',
          errorLog: 'gray',
          badRequests: 'green',
          lockCost: 'yellow',
        },
      },
    ],
  },
  {
    id: 'inventory',
    name: '库存系统',
    score: 86,
    metrics: {
      resource: 'green',
      resourceShift: 'green',
      errorRate: 'green',
      latency: 'green',
      errorLog: 'green',
      badRequests: 'green',
      lockCost: 'green',
    },
    services: [
      {
        name: 'stock-query',
        desc: '库存查询接口',
        score: 88,
        metrics: {
          resource: 'green',
          resourceShift: 'green',
          errorRate: 'green',
          latency: 'green',
          errorLog: 'green',
          badRequests: 'green',
          lockCost: 'green',
        },
      },
    ],
  },
]

const severityValue = (status) => ({ red: 4, yellow: 3, gray: 2, green: 1 }[status] || 0)

const ringColor = (score) => {
  if (score < 60) return '#f04438'
  if (score < 80) return '#f79009'
  return '#12b76a'
}

const scoreState = (score) => {
  if (score < 60) {
    return {
      className: 'red',
      note: '低于健康线，建议优先处理红灯指标',
      text: '严重异常',
    }
  }
  if (score < 80) {
    return {
      className: 'yellow',
      note: '存在持续黄灯指标，需要关注变化',
      text: '需要关注',
    }
  }
  return {
    className: 'green',
    note: '核心指标处于稳定区间',
    text: '运行健康',
  }
}

const serviceSeverity = (service) => Object
  .values(service.metrics)
  .reduce((sum, status) => sum + severityValue(status), 0)

const isAbnormal = (system) => system.score < 80 ||
  Object.values(system.metrics).some((value) => value === 'red' || value === 'yellow')

const statusCount = (metrics, status) => Object.values(metrics).filter((value) => value === status).length

const statusTags = (metrics, status, emptyText) => {
  const tags = metricLabels
    .map(([key, label]) => ({ label, status: metrics[key] }))
    .filter((item) => item.status === status)

  if (!tags.length) return <Tag>{emptyText}</Tag>
  return tags.map((item) => (
    <Tag className={`sho-tag sho-tag-${status}`} key={item.label}>{item.label}</Tag>
  ))
}

const abnormalMetricTags = (metrics, limit = 3) => {
  const tags = metricLabels
    .map(([key, label]) => ({ key, label, status: metrics[key] }))
    .filter((item) => item.status === 'red' || item.status === 'yellow')
    .sort((left, right) => severityValue(right.status) - severityValue(left.status))
    .slice(0, limit)

  if (!tags.length) return <Tag>无明显异常</Tag>
  return tags.map((item) => (
    <Tag className={`sho-tag sho-tag-${item.status}`} key={item.key}>{item.label}</Tag>
  ))
}

const Lamp = ({ status }) => (
  <Tooltip title={statusMeta[status]?.label || '无数据'}>
    <span className={`sho-lamp sho-lamp-${status}`} />
  </Tooltip>
)

Lamp.propTypes = {
  status: PropTypes.string,
}

const HealthRing = ({ score, size = 'small' }) => (
  <span
    className={`sho-health-ring sho-health-ring-${size}`}
    style={{
      '--sho-score': score,
      '--sho-ring': ringColor(score),
    }}
  >
    {score}
  </span>
)

HealthRing.propTypes = {
  score: PropTypes.number,
  size: PropTypes.oneOf(['small', 'large']),
}

const serviceMatchesFilter = (service, filter) => {
  const values = Object.values(service.metrics)
  if (filter === 'red') return values.includes('red')
  if (filter === 'yellow') return !values.includes('red') && values.includes('yellow')
  if (filter === 'green') return values.every((value) => value === 'green')
  return true
}

const priorityItems = (system) => {
  const items = []
  system.services.forEach((service) => {
    metricLabels.forEach(([key, label]) => {
      const status = service.metrics[key]
      if (status === 'red' || status === 'yellow') {
        items.push({
          key,
          label,
          service: service.name,
          status,
          score: severityValue(status) * 10 + (100 - service.score),
        })
      }
    })
  })
  return items.sort((left, right) => right.score - left.score).slice(0, 6)
}

const priorityInsightMap = {
  resource: {
    analysis: '资源水位偏高，服务容量余量不足，继续上涨会放大响应延时和错误率。',
    action: '先确认突增流量、批任务或实例异常；必要时临时扩容，并回收异常任务占用的 CPU / 内存。',
  },
  resourceShift: {
    analysis: '资源波动超过近期基线，说明负载或实例状态不稳定，容易触发间歇性故障。',
    action: '对比发布时间、批处理窗口和节点重启记录，优先稳定实例分布，再观察波动是否回落。',
  },
  errorRate: {
    analysis: '错误率已经进入异常区间，用户请求正在直接失败，是当前优先级最高的问题之一。',
    action: '先看该服务最近 5 分钟错误 Top 接口和状态码；若集中在新版本或依赖调用，优先回滚或降级依赖。',
  },
  latency: {
    analysis: '响应延时持续升高，可能来自下游等待、慢查询、队列堆积或资源竞争。',
    action: '按 P95 延时拆解接口和下游依赖，先隔离最慢链路；必要时启用限流或降级非核心路径。',
  },
  errorLog: {
    analysis: '错误日志短时间集中出现，说明异常路径已经形成稳定复现信号。',
    action: '聚合错误堆栈和关键字段，先处理出现频次最高且关联请求失败的日志类型。',
  },
  badRequests: {
    analysis: '异常请求量升高，可能是参数、鉴权、依赖返回或客户端调用模式发生变化。',
    action: '按接口、来源和错误码聚合异常请求，确认是否需要拦截异常流量或修复调用方参数。',
  },
  lockCost: {
    analysis: '锁等待耗时过高，说明并发竞争或事务持有时间正在影响服务吞吐。',
    action: '定位最长锁等待的接口、SQL 或任务，优先缩短事务范围，并暂停高冲突的批处理任务。',
  },
}

const priorityInsight = (item) => {
  const fallback = {
    analysis: `${item.service} 的 ${item.label} 已偏离正常区间，可能继续影响核心链路稳定性。`,
    action: '先确认最近变更和依赖状态，再按影响面从红灯指标开始处理。',
  }
  const insight = priorityInsightMap[item.key] || fallback
  return {
    analysis: `${item.service}：${insight.analysis}`,
    action: insight.action,
  }
}

export default function SystemHealthOverviewPage() {
  const [selectedSystemId, setSelectedSystemId] = useState('')
  const [onlyAbnormal, setOnlyAbnormal] = useState(true)
  const [systemKeyword, setSystemKeyword] = useState('')
  const [serviceKeyword, setServiceKeyword] = useState('')
  const [serviceFilter, setServiceFilter] = useState('all')
  const [expandedPriorityKey, setExpandedPriorityKey] = useState('')

  const selectedSystem = useMemo(
    () => systems.find((system) => system.id === selectedSystemId),
    [selectedSystemId],
  )

  const overviewSystems = useMemo(() => systems
    .filter((system) => !onlyAbnormal || isAbnormal(system))
    .filter((system) => !systemKeyword.trim() ||
      system.name.toLowerCase().includes(systemKeyword.trim().toLowerCase()))
    .sort((left, right) => left.score - right.score), [onlyAbnormal, systemKeyword])

  const filteredServices = useMemo(() => {
    if (!selectedSystem) return []
    const keyword = serviceKeyword.trim().toLowerCase()
    return [...selectedSystem.services]
      .filter((service) => !keyword ||
        `${service.name} ${service.desc}`.toLowerCase().includes(keyword))
      .filter((service) => serviceMatchesFilter(service, serviceFilter))
      .sort((left, right) => serviceSeverity(right) - serviceSeverity(left))
  }, [selectedSystem, serviceKeyword, serviceFilter])

  const openDetail = (system) => {
    setSelectedSystemId(system.id)
    setServiceKeyword('')
    setServiceFilter('all')
    setExpandedPriorityKey('')
  }

  const renderOverview = () => (
    <>
      <div className="sho-toolbar">
        <div>
          <div className="sho-title">系统健康概览</div>
          <div className="sho-subtitle">默认优先展示异常系统，按健康度和指标状态快速定位问题系统。</div>
        </div>
      </div>

      <div className="sho-summary">
        <div className="sho-stat sho-stat-red">
          <span>异常系统</span>
          <strong>4 / 12</strong>
          <small>2 个系统健康度低于 60</small>
        </div>
        <div className="sho-stat sho-stat-yellow">
          <span>异常指标</span>
          <strong>31</strong>
          <small>错误率、响应延时、错误日志占比最高</small>
        </div>
        <div className="sho-stat sho-stat-red">
          <span>严重异常</span>
          <strong>7</strong>
          <small>集中在订单、支付、会员系统</small>
        </div>
        <div className="sho-stat">
          <span>新增异常</span>
          <strong>5</strong>
          <small>最近 10 分钟新增</small>
        </div>
      </div>

      <section className="sho-panel">
        <div className="sho-panel-head">
          <div className="sho-panel-title">
            系统健康评分
            <Tooltip title="由资源水位、资源波动、错误率、响应延时、错误日志、异常请求、锁等待综合计算">
              <HelpCircle size={15} className="sho-help-icon" />
            </Tooltip>
          </div>
          <Space wrap size={8}>
            <Space size={6}>
              <span className="sho-muted">只看异常</span>
              <Button
                size="small"
                type={onlyAbnormal ? 'primary' : 'default'}
                onClick={() => setOnlyAbnormal((value) => !value)}
              >
                {onlyAbnormal ? '开' : '关'}
              </Button>
            </Space>
            <Input
              allowClear
              className="sho-search"
              placeholder="搜索系统"
              prefix={<Search size={14} />}
              value={systemKeyword}
              onChange={(event) => setSystemKeyword(event.target.value)}
            />
            <Legend />
          </Space>
        </div>
        <HealthTable>
          <thead>
            <tr>
              <th>系统</th>
              <th>健康度</th>
              {metricLabels.map(([, label]) => <th key={label}>{label}</th>)}
            </tr>
          </thead>
          <tbody>
            {overviewSystems.map((system) => (
              <tr key={system.id} onClick={() => openDetail(system)}>
                <td>
                  <div className="sho-name-cell">
                    <Lamp status={statusCount(system.metrics, 'red') > 0 ? 'red' : statusCount(system.metrics, 'yellow') > 0 ? 'yellow' : 'green'} />
                    <strong>{system.name}</strong>
                  </div>
                </td>
                <td><HealthRing score={system.score} /></td>
                {metricLabels.map(([key]) => (
                  <td key={key}><Lamp status={system.metrics[key]} /></td>
                ))}
              </tr>
            ))}
          </tbody>
        </HealthTable>
      </section>
    </>
  )

  const renderDetail = () => {
    const state = scoreState(selectedSystem.score)
    const keyServices = [...selectedSystem.services]
      .sort((left, right) => serviceSeverity(right) - serviceSeverity(left))
      .slice(0, 4)
    const priorities = priorityItems(selectedSystem)

    return (
      <>
        <div className="sho-toolbar">
          <Button icon={<ArrowLeft size={15} />} onClick={() => setSelectedSystemId('')}>返回</Button>
          <div className="sho-toolbar-main">
            <div className="sho-title">{selectedSystem.name}</div>
            <div className="sho-subtitle">按服务查看健康状态，优先处理红灯和持续黄灯指标。</div>
          </div>
        </div>

        <div className="sho-detail-summary">
          <div className="sho-score-card">
            <HealthRing score={selectedSystem.score} size="large" />
            <div>
              <span>健康评分</span>
              <strong className={`sho-score-state sho-score-state-${state.className}`}>{state.text}</strong>
              <small>{state.note}</small>
            </div>
          </div>
          <div className="sho-stat">
            <span>主要异常</span>
            <div className="sho-tags">{statusTags(selectedSystem.metrics, 'red', '暂无红灯指标')}</div>
          </div>
          <div className="sho-stat">
            <span>关注指标</span>
            <div className="sho-tags">{statusTags(selectedSystem.metrics, 'yellow', '暂无黄灯指标')}</div>
          </div>
          <div className="sho-stat">
            <span>
              系统关键服务
              <Tooltip title="系统中被依赖最多的服务">
                <HelpCircle size={14} className="sho-help-icon" />
              </Tooltip>
            </span>
            <div className="sho-tags">
              {keyServices.map((service) => <Tag color="blue" key={service.name}>{service.name}</Tag>)}
            </div>
          </div>
        </div>

        <section className="sho-panel">
          <div className="sho-panel-head">
            <div className="sho-panel-title">系统级指标</div>
            <Legend />
          </div>
          <div className="sho-metric-grid">
            {metricLabels.map(([key, label]) => {
              const status = selectedSystem.metrics[key] || 'gray'
              const [value, note] = metricSamples[key]?.[status] || ['--', '暂无数据']
              return (
                <div className="sho-metric-tile" key={key}>
                  <div className="sho-metric-tile-head">
                    <strong>{label}</strong>
                    <Lamp status={status} />
                  </div>
                  <b>{value}</b>
                  <small>{note}</small>
                </div>
              )
            })}
          </div>
        </section>

        <section className="sho-panel">
          <div className="sho-panel-head">
            <div className="sho-panel-title">服务列表</div>
            <Space wrap size={8}>
              <Input
                allowClear
                className="sho-search"
                placeholder="搜索服务"
                prefix={<Search size={14} />}
                value={serviceKeyword}
                onChange={(event) => setServiceKeyword(event.target.value)}
              />
              <Select
                value={serviceFilter}
                onChange={setServiceFilter}
                style={{ width: 116 }}
                options={[
                  { value: 'all', label: '全部服务' },
                  { value: 'red', label: '只看异常' },
                  { value: 'yellow', label: '只看关注' },
                  { value: 'green', label: '只看正常' },
                ]}
              />
            </Space>
          </div>
          <HealthTable className="sho-service-table">
            <thead>
              <tr>
                <th>服务</th>
                <th>健康度</th>
                {metricLabels.map(([, label]) => <th key={label}>{label}</th>)}
                <th>主要异常</th>
              </tr>
            </thead>
            <tbody>
              {filteredServices.map((service) => (
                <tr key={service.name}>
                  <td>
                    <div className="sho-service-name">
                      <strong>{service.name}</strong>
                      <span>{service.desc}</span>
                    </div>
                  </td>
                  <td><HealthRing score={service.score} /></td>
                  {metricLabels.map(([key]) => (
                    <td key={key}><Lamp status={service.metrics[key]} /></td>
                  ))}
                  <td><div className="sho-tags">{abnormalMetricTags(service.metrics)}</div></td>
                </tr>
              ))}
              {!filteredServices.length && (
                <tr>
                  <td className="sho-empty-cell" colSpan={10}>没有匹配的服务</td>
                </tr>
              )}
            </tbody>
          </HealthTable>
        </section>

        <section className="sho-panel">
          <div className="sho-panel-head">
            <div className="sho-panel-title">优先处理</div>
          </div>
          <div className="sho-priority-list">
            {priorities.map((item) => {
              const itemKey = `${item.service}-${item.key}`
              const isExpanded = expandedPriorityKey === itemKey
              const insight = priorityInsight(item)

              return (
                <div
                  className={`sho-priority-item ${isExpanded ? 'sho-priority-item-expanded' : ''}`}
                  key={itemKey}
                >
                  <Lamp status={item.status} />
                  <div className="sho-priority-main">
                    <strong>{item.label}</strong>
                    <span>{item.service} · {item.status === 'red' ? '已超过异常阈值' : '接近异常阈值'}</span>
                  </div>
                  <Button
                    className="sho-ai-button"
                    icon={<Sparkles size={13} />}
                    size="small"
                    type={isExpanded ? 'primary' : 'default'}
                    onClick={() => setExpandedPriorityKey(isExpanded ? '' : itemKey)}
                  >
                    AI 分析
                  </Button>
                  {isExpanded && (
                    <div className="sho-priority-ai-detail">
                      <div>
                        <b>分析</b>
                        <span>{insight.analysis}</span>
                      </div>
                      <div>
                        <b>处理建议</b>
                        <span>{insight.action}</span>
                      </div>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </section>
      </>
    )
  }

  return (
    <div className="system-health-overview-page">
      {selectedSystem ? renderDetail() : renderOverview()}
    </div>
  )
}

const HealthTable = ({ children, className = '' }) => (
  <div className={`sho-table-wrap ${className}`}>
    <table>{children}</table>
  </div>
)

HealthTable.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
}

const Legend = () => (
  <div className="sho-legend">
    {Object.entries(statusMeta).map(([key, meta]) => (
      <span key={key}>
        <Lamp status={key} />
        {meta.label}
      </span>
    ))}
  </div>
)
