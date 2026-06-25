import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import PropTypes from 'prop-types'
import { Alert, Button, Input, Select, Space, Spin, Switch, Tag, Tooltip } from 'antd'
import ReactECharts from 'echarts-for-react'
import { ArrowLeft, HelpCircle, Search, Sparkles } from 'lucide-react'
import { useSelector } from 'react-redux'
import { useLocation } from 'react-router-dom'
import { MetricsLineChartColor } from 'src/constants'
import { ISOToTimestamp } from 'src/core/utils/time'
import {
  analyzeSystemHealthPendingIssueApi,
  getSystemHealthOverviewApi,
  getSystemHealthSystemDetailApi,
} from 'src/pro/api/serviceHealthy'
import {
  buildMetricTrendOption,
  buildOverviewSummaryCards,
  buildStatusDistribution,
  metricSourceTitleFor,
  metricVisualFor,
  safePercent,
} from './metricVisuals.mjs'
import { buildPendingIssueQueueRows } from './pendingIssueQueue.mjs'
import { describePendingIssue } from './pendingIssueDescription.mjs'
import { buildDetailRequestKey, shouldRenderDetailBody } from './detailViewState.mjs'
import {
  isLatestOverviewRequest,
  shouldRenderInitialOverviewLoading,
} from './overviewViewState.mjs'
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

const statusOrder = ['red', 'yellow', 'green', 'gray']

const metricTrendColors = {
  latency: MetricsLineChartColor.latency,
  errorRate: MetricsLineChartColor.errorRate,
  logs: MetricsLineChartColor.logs,
  resourceShift: '#ff8a45',
}

const severityValue = (status) => ({ red: 4, yellow: 3, gray: 2, green: 1 })[status] || 0

const safeText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

const positiveNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) && number > 0 ? number : 0
}

const ringColor = (score) => {
  if (score < 60) return '#f04438'
  if (score < 80) return '#f79009'
  return '#12b76a'
}

const normalizeSystem = (system = {}) => ({
  id: system.id ?? system.groupId,
  groupId: system.groupId ?? system.id,
  name: system.name || '未命名系统',
  score: Number.isFinite(Number(system.score)) ? Number(system.score) : 0,
  status: system.status || 'gray',
  metrics: normalizeMetrics(system.metrics),
  metricDetails: system.metricDetails || {},
  services: Array.isArray(system.services) ? system.services.map(normalizeService) : [],
  source: system.source || 'data_group',
})

const normalizePendingIssue = (issue = {}) => ({
  id: issue.id || `${issue.groupId || ''}:${issue.serviceName || ''}:${issue.metricKey || ''}`,
  groupId: issue.groupId,
  systemName: issue.systemName || '未命名系统',
  serviceName: issue.serviceName || '未命名服务',
  metricKey: issue.metricKey || '',
  metricLabel: issue.metricLabel || issue.metricKey || '未知指标',
  status: issue.status || 'gray',
  value: issue.value || '--',
  note: issue.note || '暂无指标说明',
})

const buildPendingIssueAnalysisKey = (issue = {}, params = {}) =>
  JSON.stringify([
    issue.groupId || '',
    issue.serviceName || '',
    issue.metricKey || '',
    issue.status || '',
    issue.value || '',
    issue.note || '',
    params.relativeTime || '',
    params.startTime || '',
    params.endTime || '',
  ])

const normalizeService = (service = {}) => ({
  name: service.name || '未命名服务',
  desc: service.desc || '',
  score: Number.isFinite(Number(service.score)) ? Number(service.score) : 0,
  status: service.status || 'gray',
  metrics: normalizeMetrics(service.metrics),
  metricDetails: service.metricDetails || {},
})

const normalizeMetrics = (metrics = {}) =>
  metricLabels.reduce((result, [key]) => {
    result[key] = metrics[key] || 'gray'
    return result
  }, {})

const normalizeUrlTimeParams = (searchParams) => {
  const relativeTime = searchParams.get('relativeTime')
  const startTime =
    safeText(searchParams.get('startTime')) || ISOToTimestamp(searchParams.get('from'))
  const endTime = safeText(searchParams.get('endTime')) || ISOToTimestamp(searchParams.get('to'))

  if (positiveNumber(startTime) > 0 && positiveNumber(endTime) > positiveNumber(startTime)) {
    return {
      startTime,
      endTime,
    }
  }

  if (relativeTime) {
    return { relativeTime }
  }

  return null
}

const buildStoreTimeParams = (timeRange = {}) => {
  const rangeTypeKey = safeText(timeRange.rangeTypeKey)
  const storeStartTime = positiveNumber(timeRange.startTime)
  const storeEndTime = positiveNumber(timeRange.endTime)

  if (rangeTypeKey) return { relativeTime: rangeTypeKey }
  if (storeStartTime > 0 && storeEndTime > storeStartTime) {
    return { startTime: storeStartTime, endTime: storeEndTime }
  }
  return null
}

const makeTimeRangeSignature = (timeRange = {}) =>
  [
    safeText(timeRange.rangeTypeKey),
    positiveNumber(timeRange.startTime) || '',
    positiveNumber(timeRange.endTime) || '',
  ].join(':')

const buildOverviewRequestParams = (search, timeRange = {}, options = {}) => {
  const urlTimeParams = normalizeUrlTimeParams(new URLSearchParams(search))
  const storeTimeParams = buildStoreTimeParams(timeRange)

  if (!options.preferStore && urlTimeParams) return urlTimeParams
  if (storeTimeParams) return storeTimeParams
  return { relativeTime: '15m' }
}

const scoreState = (system) => {
  if (!system || system.status === 'gray' || !system.services?.length) {
    return {
      className: 'gray',
      note: '该系统暂无可用于评分的服务健康数据',
      text: '暂无数据',
    }
  }
  if (system.score < 60) {
    return {
      className: 'red',
      note: '低于健康线，建议优先处理红灯指标',
      text: '严重异常',
    }
  }
  if (system.score < 80) {
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

const serviceSeverity = (service) =>
  Object.values(service.metrics).reduce((sum, status) => sum + severityValue(status), 0)

const isAbnormal = (system) =>
  system.status !== 'gray' &&
  system.services?.length > 0 &&
  (system.score < 80 ||
    Object.values(system.metrics).some((value) => value === 'red' || value === 'yellow'))

const statusCount = (metrics, status) =>
  Object.values(metrics).filter((value) => value === status).length

const statusTags = (metrics, status, emptyText) => {
  const tags = metricLabels
    .map(([key, label]) => ({ label, status: metrics[key] }))
    .filter((item) => item.status === status)

  if (!tags.length) return <Tag>{emptyText}</Tag>
  return tags.map((item) => (
    <Tag className={`sho-tag sho-tag-${status}`} key={item.label}>
      {item.label}
    </Tag>
  ))
}

const metricDetail = (target, key) => {
  const detail = target?.metricDetails?.[key]
  if (detail) return detail
  const status = target?.metrics?.[key] || 'gray'
  return {
    status,
    value: '--',
    note: status === 'gray' ? '暂无数据' : `${statusMeta[status]?.label || '未知'}状态`,
  }
}

const Lamp = ({ status }) => (
  <Tooltip title={statusMeta[status]?.label || '无数据'}>
    <span className={`sho-lamp sho-lamp-${status || 'gray'}`} />
  </Tooltip>
)

Lamp.propTypes = {
  status: PropTypes.string,
}

const StatusDistributionBar = ({ distribution }) => {
  const total = positiveNumber(distribution.total)

  return (
    <div className="sho-status-distribution" aria-label="系统状态分布">
      {statusOrder.map((status) => {
        const count = positiveNumber(distribution[status])
        if (!count) return null
        const meta = statusMeta[status]
        return (
          <Tooltip title={`${meta.label}：${count} 个系统`} key={status}>
            <span
              className={`sho-status-distribution-segment sho-status-distribution-${status}`}
              style={{ flexGrow: count, '--sho-segment-percent': `${safePercent(count, total)}%` }}
            />
          </Tooltip>
        )
      })}
      {!total && <span className="sho-status-distribution-empty" />}
    </div>
  )
}

StatusDistributionBar.propTypes = {
  distribution: PropTypes.shape({
    total: PropTypes.number,
    red: PropTypes.number,
    yellow: PropTypes.number,
    green: PropTypes.number,
    gray: PropTypes.number,
  }).isRequired,
}

const SummaryMetricCard = ({
  label,
  value,
  caption,
  percent,
  tone = 'neutral',
  visual = 'meter',
  children,
}) => {
  const shouldShowMeter = visual === 'meter' && !children
  const shouldShowRing = visual === 'ring'

  if (shouldShowRing) {
    return (
      <div className={`sho-summary-card sho-summary-card-${tone} sho-summary-card-${visual}`}>
        <div className="sho-summary-card-ring-copy">
          <div className="sho-summary-card-head">
            <span>{label}</span>
          </div>
          <small>{caption}</small>
        </div>
        <div
          className="sho-summary-ring"
          aria-label={`${label} ${value}`}
          style={{ '--sho-ring-fill': `${percent}%` }}
        >
          <strong>{value}</strong>
        </div>
      </div>
    )
  }

  return (
    <div className={`sho-summary-card sho-summary-card-${tone} sho-summary-card-${visual}`}>
      <div className="sho-summary-card-head">
        <span>{label}</span>
        <strong>{value}</strong>
      </div>
      {children}
      {shouldShowMeter ? (
        <div className="sho-summary-meter" aria-hidden="true">
          <span style={{ '--sho-fill': `${percent}%` }} />
        </div>
      ) : null}
      <small>{caption}</small>
    </div>
  )
}

SummaryMetricCard.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.node.isRequired,
  caption: PropTypes.string.isRequired,
  percent: PropTypes.number,
  tone: PropTypes.oneOf(['critical', 'warning', 'stable', 'neutral']),
  visual: PropTypes.oneOf(['meter', 'number', 'ring']),
  children: PropTypes.node,
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

const SystemMetricTrendChart = ({ chart }) => {
  const option = useMemo(
    () =>
      buildMetricTrendOption(chart, {
        color: metricTrendColors[chart.colorRole] || '#4992ff',
        warningColor: statusMeta.yellow.color,
        criticalColor: statusMeta.red.color,
        textColor: '#667085',
        gridColor: 'rgba(102, 112, 133, 0.16)',
      }),
    [chart],
  )

  return (
    <div className="sho-metric-trend" aria-label={`${chart.key}趋势图`}>
      <ReactECharts notMerge option={option} style={{ width: '100%', height: 72 }} />
    </div>
  )
}

SystemMetricTrendChart.propTypes = {
  chart: PropTypes.shape({
    key: PropTypes.string.isRequired,
    colorRole: PropTypes.string.isRequired,
    unit: PropTypes.string,
    series: PropTypes.arrayOf(
      PropTypes.shape({
        timestamp: PropTypes.number.isRequired,
        value: PropTypes.number.isRequired,
      }),
    ).isRequired,
    thresholds: PropTypes.arrayOf(
      PropTypes.shape({
        label: PropTypes.string,
        value: PropTypes.number,
        status: PropTypes.string,
      }),
    ),
  }).isRequired,
}

const splitMetricDisplayValue = (value) => {
  const text = safeText(value) || '--'
  const spacedUnit = text.match(/^([+-]?\d+(?:\.\d+)?)(?:\s+)(.+)$/)
  if (spacedUnit) {
    return { main: spacedUnit[1], unit: spacedUnit[2] }
  }

  const attachedUnit = text.match(/^([+-]?\d+(?:\.\d+)?)([a-zA-Z][\w/%]*)$/)
  if (attachedUnit) {
    return { main: attachedUnit[1], unit: attachedUnit[2] }
  }

  return { main: text, unit: '' }
}

const MetricValue = ({ value }) => {
  const displayValue = splitMetricDisplayValue(value)

  return (
    <b>
      <span className="sho-metric-value-main">{displayValue.main}</span>
      {displayValue.unit ? (
        <span className="sho-metric-value-unit">{displayValue.unit}</span>
      ) : null}
    </b>
  )
}

MetricValue.propTypes = {
  value: PropTypes.node.isRequired,
}

const MetricVisualTile = ({ metricKey, label, detail, status }) => {
  const visual = metricVisualFor(metricKey, detail, status)
  const sourceTitle = metricSourceTitleFor(metricKey, detail)

  return (
    <div className={`sho-metric-tile sho-metric-tile-${visual.tone}`}>
      <div className="sho-metric-tile-head">
        <Tooltip title={sourceTitle}>
          <strong>{label}</strong>
        </Tooltip>
        <Lamp status={visual.status} />
      </div>
      <div className={`sho-metric-visual sho-metric-visual-${visual.kind}`}>
        <div className="sho-metric-value-block">
          <MetricValue value={visual.value} />
          <span>{visual.helper}</span>
        </div>
        {visual.chart ? (
          <SystemMetricTrendChart chart={visual.chart} />
        ) : visual.kind === 'number' ? null : visual.kind === 'waterline' ? (
          <div className="sho-waterline" aria-label={`${label}${visual.helper}${visual.percent}%`}>
            <span className="sho-waterline-marker" />
            <span className="sho-waterline-fill" style={{ '--sho-fill': `${visual.percent}%` }} />
          </div>
        ) : (
          <div className="sho-risk-meter" aria-label={`${label}${visual.helper}${visual.percent}%`}>
            <span style={{ '--sho-fill': `${visual.percent}%` }} />
          </div>
        )}
      </div>
      <div className="sho-metric-foot">
        <small>{visual.note}</small>
        {!visual.chart && visual.marker ? <em>{visual.marker}</em> : null}
      </div>
    </div>
  )
}

MetricVisualTile.propTypes = {
  metricKey: PropTypes.string.isRequired,
  label: PropTypes.string.isRequired,
  detail: PropTypes.shape({
    status: PropTypes.string,
    value: PropTypes.string,
    note: PropTypes.string,
    unit: PropTypes.string,
    source: PropTypes.string,
    series: PropTypes.array,
    thresholds: PropTypes.array,
  }).isRequired,
  status: PropTypes.string,
}

const serviceMatchesFilter = (service, filter) => {
  const values = Object.values(service.metrics)
  if (filter === 'red') return values.includes('red')
  if (filter === 'yellow') return !values.includes('red') && values.includes('yellow')
  if (filter === 'green') return values.every((value) => value === 'green')
  return true
}

const OverviewInitialLoading = () => (
  <div className="sho-overview-loading" role="status" aria-live="polite">
    <div className="sho-loading-summary" aria-hidden="true">
      {[0, 1, 2, 3].map((item) => (
        <div className="sho-loading-card" key={item}>
          <span className="sho-loading-line sho-loading-line-title" />
          <span className="sho-loading-line sho-loading-line-value" />
          <span className="sho-loading-line sho-loading-line-meter" />
          <span className="sho-loading-line sho-loading-line-caption" />
        </div>
      ))}
    </div>

    <section className="sho-panel sho-overview-loading-panel">
      <div className="sho-panel-head">
        <div className="sho-panel-title">健康度</div>
        <span className="sho-muted">正在同步指标状态</span>
      </div>
      <div className="sho-overview-loading-body">
        <Spin />
        <div>
          <strong>正在加载健康度</strong>
          <span>获取系统、指标和待处理问题，完成后展示当前状态。</span>
        </div>
      </div>
      <div className="sho-loading-table" aria-hidden="true">
        {[0, 1, 2, 3].map((row) => (
          <div className="sho-loading-table-row" key={row}>
            <span />
            <span />
            <span />
            <span />
            <span />
          </div>
        ))}
      </div>
    </section>

    <section className="sho-panel sho-overview-loading-pending">
      <div className="sho-panel-head">
        <div className="sho-panel-title">待处理问题</div>
        <span className="sho-muted">等待风险排序</span>
      </div>
      <div className="sho-loading-pending-lines" aria-hidden="true">
        <span />
        <span />
      </div>
    </section>
  </div>
)

export default function SystemHealthOverviewPage() {
  const location = useLocation()
  const timeRange = useSelector((state) => state.timeRange)
  const pageLoadedAtRef = useRef(Date.now() * 1000)
  const initialTimeRangeSignatureRef = useRef(null)
  const overviewRequestSeqRef = useRef(0)
  const [systems, setSystems] = useState([])
  const [summary, setSummary] = useState({})
  const [pendingIssues, setPendingIssues] = useState([])
  const [issueAnalyses, setIssueAnalyses] = useState({})
  const [expandedIssueAnalyses, setExpandedIssueAnalyses] = useState({})
  const [overviewLoading, setOverviewLoading] = useState(false)
  const [overviewError, setOverviewError] = useState('')
  const [selectedSystemId, setSelectedSystemId] = useState('')
  const [selectedSystem, setSelectedSystem] = useState(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [detailRefreshToken, setDetailRefreshToken] = useState(0)
  const [onlyAbnormal, setOnlyAbnormal] = useState(true)
  const [systemKeyword, setSystemKeyword] = useState('')
  const [serviceKeyword, setServiceKeyword] = useState('')
  const [serviceFilter, setServiceFilter] = useState('all')

  const currentTimeRangeSignature = makeTimeRangeSignature(timeRange)
  if (initialTimeRangeSignatureRef.current === null) {
    initialTimeRangeSignatureRef.current = currentTimeRangeSignature
  }
  const timeRangeChangedAfterMount =
    currentTimeRangeSignature !== initialTimeRangeSignatureRef.current
  const timeRangeRefreshedAfterMount =
    positiveNumber(timeRange?.refreshTimestamp) >= pageLoadedAtRef.current
  const queryParams = useMemo(
    () =>
      buildOverviewRequestParams(location.search, timeRange, {
        preferStore: timeRangeChangedAfterMount || timeRangeRefreshedAfterMount,
      }),
    [
      location.search,
      timeRange?.rangeTypeKey,
      timeRange?.startTime,
      timeRange?.endTime,
      timeRange?.refreshTimestamp,
      timeRangeChangedAfterMount,
      timeRangeRefreshedAfterMount,
    ],
  )
  const detailRequestKey = useMemo(
    () => buildDetailRequestKey(selectedSystemId, queryParams),
    [selectedSystemId, queryParams],
  )

  const loadOverview = useCallback(async () => {
    const requestId = overviewRequestSeqRef.current + 1
    overviewRequestSeqRef.current = requestId
    setOverviewLoading(true)
    setOverviewError('')
    try {
      const resp = await getSystemHealthOverviewApi(queryParams)
      if (!isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) return
      const nextSystems = Array.isArray(resp?.systems) ? resp.systems.map(normalizeSystem) : []
      setSystems(nextSystems)
      setSummary(resp?.summary || {})
      setPendingIssues(
        Array.isArray(resp?.pendingIssues) ? resp.pendingIssues.map(normalizePendingIssue) : [],
      )
    } catch (error) {
      if (!isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) return
      setOverviewError(error?.message || '健康度加载失败')
      setSystems([])
      setSummary({})
      setPendingIssues([])
    } finally {
      if (isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) {
        setOverviewLoading(false)
      }
    }
  }, [queryParams])

  useEffect(() => {
    loadOverview()
  }, [loadOverview])

  useEffect(() => {
    if (!detailRequestKey || !selectedSystemId) return undefined

    let cancelled = false
    const loadDetail = async () => {
      setDetailLoading(true)
      setDetailError('')
      try {
        const resp = await getSystemHealthSystemDetailApi(selectedSystemId, queryParams)
        if (!cancelled) {
          setSelectedSystem(normalizeSystem(resp?.system))
        }
      } catch (error) {
        if (!cancelled) {
          setDetailError(error?.message || '系统详情加载失败')
        }
      } finally {
        if (!cancelled) {
          setDetailLoading(false)
        }
      }
    }

    loadDetail()
    return () => {
      cancelled = true
    }
  }, [detailRequestKey, detailRefreshToken, queryParams, selectedSystemId])

  const overviewSystems = useMemo(
    () =>
      systems
        .filter((system) => !onlyAbnormal || isAbnormal(system))
        .filter(
          (system) =>
            !systemKeyword.trim() ||
            system.name.toLowerCase().includes(systemKeyword.trim().toLowerCase()),
        )
        .sort((left, right) => left.score - right.score),
    [systems, onlyAbnormal, systemKeyword],
  )
  const pendingIssueRows = useMemo(() => buildPendingIssueQueueRows(pendingIssues), [pendingIssues])
  const pendingServiceCount = useMemo(
    () => new Set(pendingIssueRows.map((issue) => issue.serviceName)).size,
    [pendingIssueRows],
  )

  const filteredServices = useMemo(() => {
    if (!selectedSystem) return []
    const keyword = serviceKeyword.trim().toLowerCase()
    return [...selectedSystem.services]
      .filter(
        (service) =>
          !keyword || `${service.name} ${service.desc || ''}`.toLowerCase().includes(keyword),
      )
      .filter((service) => serviceMatchesFilter(service, serviceFilter))
      .sort((left, right) => serviceSeverity(right) - serviceSeverity(left))
  }, [selectedSystem, serviceKeyword, serviceFilter])

  const openDetail = (system) => {
    setSelectedSystemId(String(system.groupId))
    setSelectedSystem(system)
    setServiceKeyword('')
    setServiceFilter('all')
    setDetailError('')
    setDetailRefreshToken((current) => current + 1)
  }

  const closeDetail = () => {
    setSelectedSystemId('')
    setSelectedSystem(null)
    setDetailError('')
  }

  const toggleIssueAnalysis = (analysisKey) => {
    setExpandedIssueAnalyses((current) => ({
      ...current,
      [analysisKey]: !(current[analysisKey] !== false),
    }))
  }

  const analyzeIssue = async (issue, analysisKey) => {
    setIssueAnalyses((current) => ({
      ...current,
      [analysisKey]: {
        ...(current[analysisKey] || {}),
        loading: true,
        error: '',
      },
    }))
    try {
      const data = await analyzeSystemHealthPendingIssueApi({
        groupId: issue.groupId,
        serviceName: issue.serviceName,
        metricKey: issue.metricKey,
        ...queryParams,
      })
      setIssueAnalyses((current) => ({
        ...current,
        [analysisKey]: {
          data,
          error: '',
          loading: false,
        },
      }))
      if (data?.aiStatus === 'succeeded') {
        setExpandedIssueAnalyses((current) => ({
          ...current,
          [analysisKey]: true,
        }))
      }
    } catch (error) {
      setIssueAnalyses((current) => ({
        ...current,
        [analysisKey]: {
          ...(current[analysisKey] || {}),
          error: error?.message || 'AI 分析失败',
          loading: false,
        },
      }))
    }
  }

  const renderOverviewState = () => {
    if (overviewLoading) {
      return (
        <div className="sho-state-panel">
          <Spin />
          <span>正在加载健康度</span>
        </div>
      )
    }
    if (overviewError) {
      return (
        <Alert
          action={
            <Button size="small" onClick={loadOverview}>
              重试
            </Button>
          }
          message="健康度加载失败"
          description={overviewError}
          type="error"
          showIcon
        />
      )
    }
    if (!systems.length) {
      return <div className="sho-empty-block">暂无健康度数据</div>
    }
    if (!overviewSystems.length) {
      return <div className="sho-empty-block">没有匹配的系统</div>
    }
    return null
  }

  const renderPendingIssueAnalysis = (analysisState, isExpanded) => {
    if (!analysisState?.data && !analysisState?.error) return null
    if (analysisState.error) {
      return (
        <div className="sho-pending-analysis sho-pending-analysis-error">{analysisState.error}</div>
      )
    }

    const data = analysisState.data || {}
    if (data.aiStatus !== 'succeeded') {
      return (
        <div className="sho-pending-analysis sho-pending-analysis-error">
          {data.error || 'AI 分析暂不可用'}
        </div>
      )
    }
    if (!isExpanded) return null

    return (
      <div className="sho-pending-analysis">
        {data.analysis && (
          <div className="sho-pending-analysis-block">
            <strong>分析</strong>
            <span>{data.analysis}</span>
          </div>
        )}
        {data.impact && (
          <div className="sho-pending-analysis-block">
            <strong>影响</strong>
            <span>{data.impact}</span>
          </div>
        )}
        {Array.isArray(data.actionPlan) && data.actionPlan.length > 0 && (
          <div className="sho-pending-steps">
            {data.actionPlan.map((step, index) => (
              <div className="sho-pending-step" key={`${step.phase}-${index}`}>
                <b>{index + 1}</b>
                <div>
                  <strong>{step.phase}</strong>
                  {step.target && <small>{step.target}</small>}
                  <span>{step.action}</span>
                  {step.validation && <em>验证：{step.validation}</em>}
                  {step.risk && <em>风险：{step.risk}</em>}
                </div>
              </div>
            ))}
          </div>
        )}
        {data.dataBoundary && <small className="sho-pending-boundary">{data.dataBoundary}</small>}
      </div>
    )
  }

  const renderPendingIssues = () => (
    <section className="sho-panel sho-pending-panel">
      <div className="sho-panel-head">
        <div className="sho-panel-title">
          待处理问题
          <Tooltip title="当前用户有权访问的所有系统中，按风险排序的红灯和黄灯服务指标">
            <HelpCircle size={15} className="sho-help-icon" />
          </Tooltip>
        </div>
        <span className="sho-muted">
          {pendingServiceCount} 个服务 / {pendingIssueRows.length} 类问题
        </span>
      </div>
      {pendingIssueRows.length ? (
        <div className="sho-pending-table-wrap">
          <table className="sho-pending-table">
            <colgroup>
              <col className="sho-pending-col-service" />
              <col className="sho-pending-col-issue" />
              <col className="sho-pending-col-status" />
              <col className="sho-pending-col-system" />
              <col className="sho-pending-col-value" />
              <col className="sho-pending-col-action" />
            </colgroup>
            <thead>
              <tr>
                <th>服务</th>
                <th>问题</th>
                <th>状态</th>
                <th>系统</th>
                <th>当前值</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {pendingIssueRows.map((issue) => {
                const analysisKey = buildPendingIssueAnalysisKey(issue, queryParams)
                const analysisState = issueAnalyses[analysisKey] || {}
                const hasSucceededAnalysis = analysisState.data?.aiStatus === 'succeeded'
                const isAnalysisExpanded =
                  hasSucceededAnalysis && expandedIssueAnalyses[analysisKey] !== false
                const analysisNode = renderPendingIssueAnalysis(analysisState, isAnalysisExpanded)
                const statusLabel = statusMeta[issue.status]?.label || '未知'

                return (
                  <React.Fragment key={issue.queueId}>
                    <tr className="sho-pending-row">
                      <td className="sho-pending-service-cell">
                        <div className="sho-pending-service-inline">
                          <span className="sho-pending-service-name" title={issue.serviceName}>
                            {issue.serviceName}
                          </span>
                        </div>
                      </td>
                      <td className="sho-pending-issue-cell">
                        <div className="sho-pending-issue-content">
                          <div className="sho-pending-issue-title">
                            <Lamp status={issue.status} />
                            <strong>{issue.metricLabel}</strong>
                          </div>
                          <small title={issue.note}>{describePendingIssue(issue)}</small>
                        </div>
                      </td>
                      <td>
                        <Tag className={`sho-tag sho-tag-${issue.status}`}>{statusLabel}</Tag>
                      </td>
                      <td>
                        <span className="sho-pending-system" title={issue.systemName}>
                          {issue.systemName}
                        </span>
                      </td>
                      <td>
                        <span className="sho-pending-value">{issue.value}</span>
                      </td>
                      <td className="sho-pending-action-cell">
                        <Button
                          icon={<Sparkles size={13} />}
                          loading={analysisState.loading}
                          size="small"
                          type={hasSucceededAnalysis && isAnalysisExpanded ? 'primary' : 'default'}
                          onClick={() => {
                            if (hasSucceededAnalysis) {
                              toggleIssueAnalysis(analysisKey)
                              return
                            }
                            analyzeIssue(issue, analysisKey)
                          }}
                        >
                          {hasSucceededAnalysis
                            ? isAnalysisExpanded
                              ? '收起分析'
                              : '展开分析'
                            : 'AI 分析'}
                        </Button>
                      </td>
                    </tr>
                    {analysisNode && (
                      <tr className="sho-pending-analysis-row">
                        <td colSpan={6}>{analysisNode}</td>
                      </tr>
                    )}
                  </React.Fragment>
                )
              })}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="sho-empty-block">暂无需要优先处理的系统问题</div>
      )}
    </section>
  )

  const renderOverview = () => {
    const initialOverviewLoading = shouldRenderInitialOverviewLoading({
      overviewLoading,
      systems,
      pendingIssues,
      summary,
    })
    const overviewState = renderOverviewState()
    const statusDistribution = buildStatusDistribution(systems)
    const summaryCards = buildOverviewSummaryCards(summary, systems, metricLabels)

    return (
      <>
        <div className="sho-toolbar">
          <div>
            <div className="sho-title">健康度概览</div>
            <div className="sho-subtitle">
              默认优先展示异常系统，按健康度和指标状态快速定位问题系统。
            </div>
          </div>
        </div>

        {initialOverviewLoading ? (
          <OverviewInitialLoading />
        ) : (
          <>
            <div className="sho-summary">
              {summaryCards.map((card) => (
                <SummaryMetricCard
                  caption={card.caption}
                  key={card.key}
                  label={card.label}
                  percent={card.percent}
                  tone={card.tone}
                  value={card.value}
                  visual={card.visual}
                >
                  {card.distribution ? (
                    <StatusDistributionBar distribution={statusDistribution} />
                  ) : null}
                </SummaryMetricCard>
              ))}
            </div>

            <section className="sho-panel">
              <div className="sho-panel-head">
                <div className="sho-panel-title">
                  系统健康度
                  <Tooltip title="由资源水位、资源波动、错误率、响应延时、错误日志、异常请求、锁等待综合计算">
                    <HelpCircle size={15} className="sho-help-icon" />
                  </Tooltip>
                </div>
                <Space wrap size={8}>
                  <Space size={6}>
                    <span className="sho-muted">只看异常</span>
                    <Switch
                      checked={onlyAbnormal}
                      checkedChildren="开"
                      unCheckedChildren="关"
                      onChange={setOnlyAbnormal}
                    />
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
              {overviewState || (
                <HealthTable>
                  <thead>
                    <tr>
                      <th>系统</th>
                      <th>健康度</th>
                      {metricLabels.map(([, label]) => (
                        <th key={label}>{label}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {overviewSystems.map((system) => (
                      <tr key={system.groupId} onClick={() => openDetail(system)}>
                        <td>
                          <div className="sho-name-cell">
                            <Lamp
                              status={
                                statusCount(system.metrics, 'red') > 0
                                  ? 'red'
                                  : statusCount(system.metrics, 'yellow') > 0
                                    ? 'yellow'
                                    : system.status
                              }
                            />
                            <strong>{system.name}</strong>
                          </div>
                        </td>
                        <td>
                          <HealthRing score={system.score} />
                        </td>
                        {metricLabels.map(([key]) => (
                          <td key={key}>
                            <Lamp status={system.metrics[key]} />
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </HealthTable>
              )}
            </section>
            {renderPendingIssues()}
          </>
        )}
      </>
    )
  }

  const renderDetail = () => {
    if (!selectedSystem) return null

    const state = scoreState(selectedSystem)
    const shouldShowDetailBody = shouldRenderDetailBody({ selectedSystem, detailLoading })
    const keyServices = [...selectedSystem.services]
      .sort((left, right) => serviceSeverity(right) - serviceSeverity(left))
      .slice(0, 4)

    return (
      <>
        <div className="sho-toolbar">
          <Button icon={<ArrowLeft size={15} />} onClick={closeDetail}>
            返回
          </Button>
          <div className="sho-toolbar-main">
            <div className="sho-title">{selectedSystem.name}</div>
            <div className="sho-subtitle">按服务查看健康状态，优先处理红灯和持续黄灯指标。</div>
          </div>
        </div>

        {detailError && (
          <Alert
            className="sho-detail-alert"
            action={
              <Button size="small" onClick={() => setDetailRefreshToken((current) => current + 1)}>
                重试
              </Button>
            }
            message="系统详情加载失败"
            description={detailError}
            type="error"
            showIcon
          />
        )}

        {detailLoading && (
          <div className="sho-state-panel sho-detail-loading">
            <Spin />
            <span>正在加载系统详情</span>
          </div>
        )}

        {shouldShowDetailBody && (
          <>
            <div className="sho-detail-summary">
              <div className="sho-score-card">
                <HealthRing score={selectedSystem.score} size="large" />
                <div>
                  <span>健康度</span>
                  <strong className={`sho-score-state sho-score-state-${state.className}`}>
                    {state.text}
                  </strong>
                  <small>{state.note}</small>
                </div>
              </div>
              <div className="sho-stat">
                <span>主要异常</span>
                <div className="sho-tags">
                  {statusTags(selectedSystem.metrics, 'red', '暂无红灯指标')}
                </div>
              </div>
              <div className="sho-stat">
                <span>关注指标</span>
                <div className="sho-tags">
                  {statusTags(selectedSystem.metrics, 'yellow', '暂无黄灯指标')}
                </div>
              </div>
              <div className="sho-stat">
                <span>
                  系统关键服务
                  <Tooltip title="按健康风险排序的重点服务">
                    <HelpCircle size={14} className="sho-help-icon" />
                  </Tooltip>
                </span>
                <div className="sho-tags">
                  {keyServices.length ? (
                    keyServices.map((service) => (
                      <Tooltip title={service.name} key={service.name}>
                        <Tag className="sho-key-service-tag">{service.name}</Tag>
                      </Tooltip>
                    ))
                  ) : (
                    <Tag>暂无服务</Tag>
                  )}
                </div>
              </div>
            </div>

            <section className="sho-panel sho-system-metrics-panel">
              <div className="sho-panel-head">
                <div className="sho-panel-title">系统级指标</div>
                <Legend />
              </div>
              <div className="sho-metric-grid">
                {metricLabels.map(([key, label]) => {
                  const status = selectedSystem.metrics[key] || 'gray'
                  const detail = metricDetail(selectedSystem, key)
                  return (
                    <MetricVisualTile
                      detail={detail}
                      key={key}
                      label={label}
                      metricKey={key}
                      status={status}
                    />
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
                    {metricLabels.map(([, label]) => (
                      <th key={label}>{label}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {filteredServices.map((service) => (
                    <tr key={service.name}>
                      <td>
                        <div className="sho-service-name">
                          <Tooltip title={service.name}>
                            <strong>{service.name}</strong>
                          </Tooltip>
                          {service.desc ? <span>{service.desc}</span> : null}
                        </div>
                      </td>
                      <td>
                        <HealthRing score={service.score} />
                      </td>
                      {metricLabels.map(([key]) => (
                        <td key={key}>
                          <Lamp status={service.metrics[key]} />
                        </td>
                      ))}
                    </tr>
                  ))}
                  {!filteredServices.length && (
                    <tr>
                      <td className="sho-empty-cell" colSpan={9}>
                        没有匹配的服务
                      </td>
                    </tr>
                  )}
                </tbody>
              </HealthTable>
            </section>
          </>
        )}
      </>
    )
  }

  return (
    <div className="system-health-overview-page">
      {selectedSystemId ? renderDetail() : renderOverview()}
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
      <span className="sho-legend-item" key={key}>
        <Lamp status={key} />
        <b>{meta.label}</b>
      </span>
    ))}
  </div>
)
