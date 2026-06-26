import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import PropTypes from 'prop-types'
import {
  Alert,
  Button,
  Input,
  InputNumber,
  message,
  Pagination,
  Popover,
  Select,
  Space,
  Spin,
  Statistic,
  Switch,
  Tag,
  Tooltip,
} from 'antd'
import ReactECharts from 'echarts-for-react'
import {
  ArrowLeft,
  Check,
  CircleAlert,
  HelpCircle,
  Search,
  Settings,
  Sparkles,
  TriangleAlert,
} from 'lucide-react'
import { useSelector } from 'react-redux'
import { useLocation } from 'react-router-dom'
import { MetricsLineChartColor } from 'src/constants'
import {
  analyzeSystemHealthPendingIssueApi,
  getSystemHealthOverviewApi,
  getSystemHealthSystemDetailApi,
  setSystemHealthConfigApi,
} from 'src/pro/api/serviceHealthy'
import {
  buildMetricTrendOption,
  buildOverviewSummaryCards,
  buildStatusDistribution,
  DEFAULT_ABNORMAL_SCORE_THRESHOLD,
  HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD,
  HEALTH_SCORE_CRITICAL_THRESHOLD,
  healthScoreLegendItems,
  metricSourceTitleFor,
  metricVisualFor,
  normalizeAbnormalScoreThreshold,
  safePercent,
} from './metricVisuals.mjs'
import { buildPendingIssueQueueRows } from './pendingIssueQueue.mjs'
import { describePendingIssue } from './pendingIssueDescription.mjs'
import {
  buildPendingIssuesFromSystems,
  DEFAULT_PENDING_ISSUE_PAGE,
  DEFAULT_PENDING_ISSUE_PAGE_SIZE,
  paginatePendingIssues,
} from './pendingIssuePagination.mjs'
import { buildDetailRequestKey, shouldRenderDetailBody } from './detailViewState.mjs'
import {
  buildOverviewRequestDecision,
  isLatestOverviewRequest,
  shouldRenderInitialOverviewLoading,
} from './overviewViewState.mjs'
import {
  buildOverviewRequestKey,
  buildOverviewRequestParams,
  makeTimeRangeSignature,
} from './timeParams.mjs'
import './index.css'

const metricLabels = [
  ['resource', '资源水位'],
  ['resourceShift', '资源波动'],
  ['errorRate', '错误率'],
  ['latency', '响应延时'],
  ['errorLog', '错误日志'],
  ['badRequests', '活跃告警'],
  ['lockCost', '锁等待'],
]

const statusMeta = {
  green: { label: '正常', color: '#12b76a' },
  yellow: { label: '关注', color: '#f79009' },
  red: { label: '异常', color: '#f04438' },
  gray: { label: '无数据', color: '#98a2b3' },
}

const statusOrder = ['red', 'yellow', 'green', 'gray']
const OVERVIEW_REQUEST_DEBOUNCE_MS = 150

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
  if (score < HEALTH_SCORE_CRITICAL_THRESHOLD) return '#f04438'
  if (score < HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD) return '#f79009'
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

const serviceSeverity = (service) =>
  Object.values(service.metrics).reduce((sum, status) => sum + severityValue(status), 0)

const isAbnormal = (system, threshold = DEFAULT_ABNORMAL_SCORE_THRESHOLD) =>
  system.status !== 'gray' &&
  system.services?.length > 0 &&
  system.score < normalizeAbnormalScoreThreshold(threshold)

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
  help,
  value,
  icon,
  numerator,
  denominator,
  highlightNumerator = false,
  percent,
  tone = 'neutral',
  visual = 'meter',
  children,
}) => {
  const shouldShowMeter = visual === 'meter' && !children
  const renderedValue =
    visual === 'ratio' ? (
      <span className="sho-summary-value" aria-label={`${label} ${value}`}>
        <span className={highlightNumerator ? 'sho-summary-value-critical' : undefined}>
          {numerator}
        </span>
        <span> / {denominator}</span>
      </span>
    ) : (
      value
    )
  const statisticFormatter = visual === 'ratio' ? () => renderedValue : undefined

  return (
    <div className={`sho-summary-card sho-summary-card-${tone} sho-summary-card-${visual}`}>
      <div className="sho-summary-card-head">
        <span>
          {icon}
          <span>{label}</span>
          {help ? (
            <Tooltip title={help}>
              <HelpCircle size={15} className="sho-help-icon sho-summary-help" />
            </Tooltip>
          ) : null}
        </span>
        <Statistic className="sho-summary-statistic" formatter={statisticFormatter} value={value} />
      </div>
      {children}
      {shouldShowMeter ? (
        <div className="sho-summary-meter" aria-hidden="true">
          <span style={{ '--sho-fill': `${percent}%` }} />
        </div>
      ) : null}
    </div>
  )
}

SummaryMetricCard.propTypes = {
  label: PropTypes.string.isRequired,
  help: PropTypes.node,
  value: PropTypes.node.isRequired,
  icon: PropTypes.node,
  numerator: PropTypes.number,
  denominator: PropTypes.number,
  highlightNumerator: PropTypes.bool,
  percent: PropTypes.number,
  tone: PropTypes.oneOf(['critical', 'warning', 'stable', 'neutral']),
  visual: PropTypes.oneOf(['meter', 'number', 'ratio']),
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
  size: PropTypes.oneOf(['mini', 'small']),
}

const HealthScoreTooltip = () => (
  <div className="sho-health-score-help">
    <div className="sho-health-score-help-copy">
      由资源水位、资源波动、错误率、响应延时、错误日志、活跃告警、锁等待综合计算
    </div>
    <div className="sho-health-score-help-list">
      {healthScoreLegendItems.map((item) => (
        <div className="sho-health-score-help-item" key={item.status}>
          <span
            className={`sho-health-score-help-ring sho-health-score-help-ring-${item.status}`}
            style={{ '--sho-ring': statusMeta[item.status]?.color || statusMeta.gray.color }}
            aria-hidden="true"
          />
          <span className="sho-health-score-help-range">{item.range}</span>
          <span
            className={`sho-legend-item sho-health-score-help-tag sho-health-score-help-tag-${item.status}`}
          >
            <span className={`sho-lamp sho-lamp-${item.status}`} />
            <b>{item.label}</b>
          </span>
        </div>
      ))}
    </div>
  </div>
)

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
        </div>
        {visual.chart ? (
          <SystemMetricTrendChart chart={visual.chart} />
        ) : visual.kind === 'number' ? null : visual.kind === 'waterline' ? (
          <div className="sho-waterline" aria-label={`${label}${visual.helper}${visual.percent}%`}>
            {visual.showThresholdMarker ? <span className="sho-waterline-marker" /> : null}
            <span className="sho-waterline-fill" style={{ '--sho-fill': `${visual.percent}%` }} />
          </div>
        ) : (
          <div className="sho-risk-meter" aria-label={`${label}${visual.helper}${visual.percent}%`}>
            <span style={{ '--sho-fill': `${visual.percent}%` }} />
          </div>
        )}
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

const DetailLoadingStatus = ({ className = '' }) => (
  <div
    className={['sho-detail-loading-status', className].filter(Boolean).join(' ')}
    role="status"
    aria-live="polite"
  >
    <span className="sho-detail-loading-pulse" aria-hidden="true">
      <span />
      <span />
      <span />
    </span>
    <span>正在加载系统详情</span>
  </div>
)

DetailLoadingStatus.propTypes = {
  className: PropTypes.string,
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
      {[0, 1, 2].map((item) => (
        <div className="sho-loading-card" key={item}>
          <span className="sho-loading-line sho-loading-line-title" />
          <span className="sho-loading-line sho-loading-line-value" />
        </div>
      ))}
    </div>

    <section className="sho-panel sho-overview-loading-panel">
      <div className="sho-panel-head">
        <div className="sho-panel-title">系统健康度</div>
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
            <span className="sho-loading-table-cell sho-loading-table-cell-system" />
            <span className="sho-loading-table-cell sho-loading-table-cell-score" />
            {metricLabels.map(([key]) => (
              <span className="sho-loading-table-cell sho-loading-table-cell-metric" key={key} />
            ))}
          </div>
        ))}
      </div>
    </section>

    <section className="sho-panel sho-overview-loading-pending">
      <div className="sho-panel-head">
        <div className="sho-panel-title">待处理问题</div>
        <span className="sho-muted">等待风险排序</span>
      </div>
      <div className="sho-loading-pending-table" aria-hidden="true">
        {[0, 1, 2].map((row) => (
          <div className="sho-loading-pending-row" key={row}>
            <span />
            <span />
            <span />
            <span />
            <span />
            <span />
          </div>
        ))}
      </div>
    </section>
  </div>
)

const DetailLoadingPanel = () => (
  <section className="sho-panel sho-detail-loading">
    <div className="sho-panel-head">
      <div className="sho-panel-title">健康指标</div>
      <DetailLoadingStatus className="sho-detail-loading-status-muted" />
    </div>
    <div className="sho-detail-loading-body" aria-hidden="true">
      <div className="sho-detail-loading-metrics">
        {metricLabels.map(([key]) => (
          <div className="sho-detail-loading-metric" key={key}>
            <div className="sho-detail-loading-metric-head">
              <span className="sho-loading-line sho-detail-loading-title-line" />
              <span className="sho-detail-loading-dot" />
            </div>
            <span className="sho-loading-line sho-detail-loading-value-line" />
            <span className="sho-loading-line sho-detail-loading-meter-line" />
          </div>
        ))}
      </div>
      <div className="sho-detail-loading-table">
        {[0, 1, 2, 3].map((row) => (
          <div className="sho-detail-loading-row" key={row}>
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
            <span className="sho-loading-line" />
          </div>
        ))}
      </div>
    </div>
  </section>
)

export default function SystemHealthOverviewPage() {
  const location = useLocation()
  const timeRange = useSelector((state) => state.timeRange)
  const pageLoadedAtRef = useRef(Date.now() * 1000)
  const initialTimeRangeSignatureRef = useRef(null)
  const overviewRequestSeqRef = useRef(0)
  const lastOverviewRequestKeyRef = useRef('')
  const overviewRequestTimerRef = useRef(null)
  const [systems, setSystems] = useState([])
  const [summary, setSummary] = useState({})
  const [abnormalScoreThreshold, setAbnormalScoreThreshold] = useState(
    DEFAULT_ABNORMAL_SCORE_THRESHOLD,
  )
  const [thresholdDraft, setThresholdDraft] = useState(DEFAULT_ABNORMAL_SCORE_THRESHOLD)
  const [thresholdSaving, setThresholdSaving] = useState(false)
  const [pendingIssuePage, setPendingIssuePage] = useState(DEFAULT_PENDING_ISSUE_PAGE)
  const [pendingIssuePageSize, setPendingIssuePageSize] = useState(DEFAULT_PENDING_ISSUE_PAGE_SIZE)
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
        requireResolvedTimeRange: true,
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
  const overviewRequestKey = useMemo(() => buildOverviewRequestKey(queryParams), [queryParams])
  const detailRequestKey = useMemo(
    () => buildDetailRequestKey(selectedSystemId, queryParams),
    [selectedSystemId, queryParams],
  )

  const loadOverview = useCallback(async () => {
    if (!queryParams) return
    const requestId = overviewRequestSeqRef.current + 1
    overviewRequestSeqRef.current = requestId
    setOverviewLoading(true)
    setOverviewError('')
    try {
      const resp = await getSystemHealthOverviewApi({
        ...queryParams,
        skipNewAbnormalMetrics: true,
      })
      if (!isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) return
      const nextSystems = Array.isArray(resp?.systems) ? resp.systems.map(normalizeSystem) : []
      const nextThreshold = normalizeAbnormalScoreThreshold(resp?.config?.abnormalScoreThreshold)
      setSystems(nextSystems)
      setAbnormalScoreThreshold(nextThreshold)
      setThresholdDraft(nextThreshold)
      setSummary(resp?.summary || {})
    } catch (error) {
      if (!isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) return
      setOverviewError(error?.message || '健康度加载失败')
      setSystems([])
      setSummary({})
    } finally {
      if (isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) {
        setOverviewLoading(false)
      }
    }
  }, [queryParams])

  useEffect(() => {
    const decision = buildOverviewRequestDecision({
      requestKey: overviewRequestKey,
      lastRequestKey: lastOverviewRequestKeyRef.current,
      overviewLoading,
    })
    if (!decision.shouldLoad) {
      lastOverviewRequestKeyRef.current = decision.nextLastRequestKey
      return
    }
    if (overviewRequestTimerRef.current) {
      window.clearTimeout(overviewRequestTimerRef.current)
    }
    overviewRequestTimerRef.current = window.setTimeout(() => {
      overviewRequestTimerRef.current = null
      lastOverviewRequestKeyRef.current = overviewRequestKey
      loadOverview()
    }, OVERVIEW_REQUEST_DEBOUNCE_MS)
    return () => {
      if (overviewRequestTimerRef.current) {
        window.clearTimeout(overviewRequestTimerRef.current)
        overviewRequestTimerRef.current = null
      }
    }
  }, [loadOverview, overviewLoading, overviewRequestKey])

  useEffect(() => {
    setPendingIssuePage(DEFAULT_PENDING_ISSUE_PAGE)
  }, [queryParams])

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
        .filter((system) => !onlyAbnormal || isAbnormal(system, abnormalScoreThreshold))
        .filter(
          (system) =>
            !systemKeyword.trim() ||
            system.name.toLowerCase().includes(systemKeyword.trim().toLowerCase()),
        )
        .sort((left, right) => left.score - right.score),
    [systems, onlyAbnormal, systemKeyword, abnormalScoreThreshold],
  )
  const allPendingIssues = useMemo(
    () => buildPendingIssuesFromSystems(systems).map(normalizePendingIssue),
    [systems],
  )
  const pendingIssuePageData = useMemo(
    () => paginatePendingIssues(allPendingIssues, pendingIssuePage, pendingIssuePageSize),
    [allPendingIssues, pendingIssuePage, pendingIssuePageSize],
  )
  const pendingIssues = pendingIssuePageData.issues
  const pendingIssueTotal = pendingIssuePageData.total
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

  const handlePendingIssuePageChange = (page, pageSize) => {
    setPendingIssuePage(page)
    setPendingIssuePageSize(pageSize)
  }

  const saveAbnormalScoreThreshold = async () => {
    const nextThreshold = normalizeAbnormalScoreThreshold(thresholdDraft)
    setThresholdDraft(nextThreshold)
    if (nextThreshold === abnormalScoreThreshold) return

    setThresholdSaving(true)
    try {
      const resp = await setSystemHealthConfigApi({
        abnormalScoreThreshold: nextThreshold,
      })
      const savedThreshold = normalizeAbnormalScoreThreshold(
        resp?.abnormalScoreThreshold ?? nextThreshold,
      )
      setAbnormalScoreThreshold(savedThreshold)
      setThresholdDraft(savedThreshold)
      message.success('异常阈值已更新')
      await loadOverview()
    } catch (error) {
      message.error(error?.message || '异常阈值保存失败')
      setThresholdDraft(abnormalScoreThreshold)
    } finally {
      setThresholdSaving(false)
    }
  }

  const renderThresholdSettings = () => {
    const hasThresholdChange =
      normalizeAbnormalScoreThreshold(thresholdDraft) !== abnormalScoreThreshold

    return (
      <div className="sho-threshold-popover">
        <div className="sho-threshold-popover-title">异常标准</div>
        <div className="sho-threshold-rule">
          健康度低于 <b>{abnormalScoreThreshold}</b> 分，就算异常。
        </div>
        <div className="sho-threshold-note">所有人使用同一个设置。</div>
        <Space className="sho-threshold-editor" size={6}>
          <span className="sho-muted">健康度低于</span>
          <InputNumber
            max={100}
            min={1}
            precision={0}
            size="small"
            value={thresholdDraft}
            onChange={(value) => setThresholdDraft(value)}
            onPressEnter={saveAbnormalScoreThreshold}
          />
          <span className="sho-muted">分</span>
          <Tooltip title={hasThresholdChange ? '保存阈值' : '阈值未变化'}>
            <span className="sho-threshold-save-wrap">
              <Button
                aria-label="保存异常阈值"
                className="sho-threshold-save"
                disabled={!hasThresholdChange}
                icon={<Check size={14} />}
                loading={thresholdSaving}
                size="small"
                type="primary"
                onClick={saveAbnormalScoreThreshold}
              />
            </span>
          </Tooltip>
        </Space>
      </div>
    )
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
    if (!queryParams || overviewLoading) {
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
          {pendingServiceCount} 个服务 / 本页 {pendingIssueRows.length} 类问题 / 共{' '}
          {pendingIssueTotal} 条
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
      {pendingIssueTotal > pendingIssuePageSize && (
        <div className="sho-pending-pagination">
          <Pagination
            current={pendingIssuePage}
            pageSize={pendingIssuePageSize}
            pageSizeOptions={['10', '20', '50', '100']}
            showSizeChanger
            showTotal={(total) => `共 ${total} 条`}
            size="small"
            total={pendingIssueTotal}
            onChange={handlePendingIssuePageChange}
          />
        </div>
      )}
    </section>
  )

  const renderOverview = () => {
    const effectiveOverviewLoading = !queryParams || overviewLoading
    const initialOverviewLoading = shouldRenderInitialOverviewLoading({
      overviewLoading: effectiveOverviewLoading,
      systems,
      pendingIssues,
      summary,
    })
    const overviewState = renderOverviewState()
    const statusDistribution = buildStatusDistribution(systems)
    const summaryCards = buildOverviewSummaryCards(summary, systems, abnormalScoreThreshold)

    return (
      <>
        <div className="sho-toolbar">
          <div>
            <div className="sho-title">健康度概览</div>
          </div>
        </div>

        {initialOverviewLoading ? (
          <OverviewInitialLoading />
        ) : (
          <>
            <div className="sho-summary">
              {summaryCards.map((card) => (
                <SummaryMetricCard
                  denominator={card.denominator}
                  highlightNumerator={card.highlightNumerator}
                  icon={
                    card.key === 'abnormalSystems' ? (
                      <Popover
                        content={renderThresholdSettings()}
                        placement="bottomLeft"
                        trigger="click"
                      >
                        <Button
                          aria-label="异常判定设置"
                          className="sho-threshold-trigger"
                          icon={<Settings size={15} />}
                          size="small"
                          type="text"
                        />
                      </Popover>
                    ) : card.key === 'severeMetrics' ? (
                      <TriangleAlert
                        className="sho-summary-title-icon sho-summary-title-icon-critical"
                        size={17}
                      />
                    ) : card.key === 'generalMetrics' ? (
                      <CircleAlert
                        className="sho-summary-title-icon sho-summary-title-icon-warning"
                        size={17}
                      />
                    ) : null
                  }
                  key={card.key}
                  label={card.label}
                  help={card.key === 'abnormalSystems' ? null : card.help}
                  numerator={card.numerator}
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
                  <Tooltip
                    title={<HealthScoreTooltip />}
                    overlayClassName="sho-health-score-popover"
                  >
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

    const shouldShowDetailBody = shouldRenderDetailBody({ selectedSystem, detailLoading })

    return (
      <>
        <div className="sho-detail-compact-header">
          <div className="sho-detail-compact-inner">
            <div className="sho-detail-compact-system">
              <Button
                className="sho-detail-compact-back"
                icon={<ArrowLeft size={12} />}
                size="small"
                type="text"
                onClick={closeDetail}
              >
                返回
              </Button>
              <div className="sho-detail-compact-title">
                <strong>{selectedSystem.name}</strong>
              </div>
            </div>
            {detailLoading ? (
              <>
                <span className="sho-detail-compact-divider" aria-hidden="true" />
                <DetailLoadingStatus className="sho-detail-compact-loading" />
              </>
            ) : (
              <>
                <span className="sho-detail-compact-divider" aria-hidden="true" />
                <div className="sho-detail-compact-score">
                  <HealthRing score={selectedSystem.score} size="mini" />
                  <div className="sho-detail-compact-score-label">
                    <span>健康度</span>
                  </div>
                </div>
                <span className="sho-detail-compact-divider" aria-hidden="true" />
                <div className="sho-detail-compact-exceptions">
                  <div className="sho-detail-compact-exception-label">
                    <TriangleAlert size={18} />
                    <span>主要异常</span>
                  </div>
                  <div className="sho-tags sho-detail-compact-tags">
                    {statusTags(selectedSystem.metrics, 'red', '暂无红灯指标')}
                  </div>
                </div>
              </>
            )}
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

        {detailLoading && <DetailLoadingPanel />}

        {shouldShowDetailBody && (
          <>
            <section className="sho-panel sho-system-metrics-panel">
              <div className="sho-panel-head">
                <div className="sho-panel-title">健康指标</div>
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
                <div className="sho-panel-title">服务健康度</div>
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
    <div
      className={`system-health-overview-page${selectedSystemId ? ' system-health-overview-page-detail' : ''}`}
    >
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
