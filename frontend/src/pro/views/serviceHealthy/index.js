import React, { useEffect, useMemo, useRef, useState } from 'react'
import PropTypes from 'prop-types'
import {
  Button,
  Drawer,
  Empty,
  Input,
  List,
  Progress,
  Select,
  Space,
  Spin,
  Tag,
} from 'antd'
import ReactECharts from 'echarts-for-react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { ChartColorList } from 'src/constants'
import { ISOToTimestamp } from 'src/core/utils/time'
import {
  getSystemHealthReportApi,
  getSystemHealthReportRecordApi,
  getSystemHealthReportRecordsApi,
} from 'src/pro/api/serviceHealthy'
import './index.css'

const LEVEL_META = {
  healthy: { text: '正常', color: '#12b76a', tag: 'success', score: 92 },
  minor: { text: '轻微异常', color: '#84cc16', tag: 'lime', score: 82 },
  warning: { text: '警告', color: '#f79009', tag: 'warning', score: 68 },
  severe: { text: '严重', color: '#f04438', tag: 'error', score: 42 },
  outage: { text: '故障', color: '#b42318', tag: 'error', score: 18 },
  recovering: { text: '恢复中', color: '#2e90fa', tag: 'processing', score: 74 },
  unknown: { text: '分析中', color: '#667085', tag: 'default', score: 0 },
}

const jumpTargets = {
  alerts: '/alerts/events',
}

const EMPTY_REPORT = {
  status: 'unknown',
  score: 0,
  availableSources: [],
  resourceUsage: 0,
  infraAlerts: 0,
  serviceRiskCount: 0,
  activeAlertCount: 0,
  logRiskCount: 0,
  summary: '正在读取真实环境数据。',
}

const REPORT_STORAGE_PREFIX = 'apo:system-health-report:'
const REPORT_STORAGE_MAX_AGE = 30 * 60 * 1000
const ENABLE_REPORT_STORAGE = false

const StatusTag = ({ status, className, hiddenStatuses = [] }) => {
  if (hiddenStatuses.includes(status)) return null
  const meta = LEVEL_META[status] ?? LEVEL_META.warning
  return <Tag className={className} color={meta.tag}>{meta.text}</Tag>
}

StatusTag.propTypes = {
  className: PropTypes.string,
  hiddenStatuses: PropTypes.arrayOf(PropTypes.string),
  status: PropTypes.string,
}

const ExpandableText = ({
  className = '',
  collapseLabel = '收起',
  expandLabel = '展开',
  text = '',
  threshold = 140,
}) => {
  const content = safeText(text)
  const contentRef = useRef(null)
  const [expanded, setExpanded] = useState(false)
  const [overflowing, setOverflowing] = useState(false)
  useEffect(() => {
    const element = contentRef.current
    if (!element || expanded) return undefined
    let frame = 0
    const measure = () => {
      window.cancelAnimationFrame(frame)
      frame = window.requestAnimationFrame(() => {
        setOverflowing(element.scrollHeight > element.clientHeight + 1)
      })
    }
    measure()
    window.addEventListener('resize', measure)
    return () => {
      window.cancelAnimationFrame(frame)
      window.removeEventListener('resize', measure)
    }
  }, [content, expanded])
  if (!content) return null
  const expandable = overflowing || content.length > threshold
  return (
    <div
      className={[
        'health-expandable-text',
        !expanded ? 'health-expandable-text-collapsed' : '',
        className,
      ].filter(Boolean).join(' ')}
    >
      <span className="health-expandable-text-content" ref={contentRef}>{content}</span>
      {expandable && (
        <button
          type="button"
          className="health-expandable-text-toggle"
          aria-expanded={expanded}
          onClick={() => setExpanded((value) => !value)}
        >
          {expanded ? collapseLabel : expandLabel}
        </button>
      )}
    </div>
  )
}

ExpandableText.propTypes = {
  className: PropTypes.string,
  collapseLabel: PropTypes.string,
  expandLabel: PropTypes.string,
  text: PropTypes.string,
  threshold: PropTypes.number,
}

const safeList = (value) => (Array.isArray(value) ? value : [])

const safeText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

const firstText = (...values) => values.map(safeText).find(Boolean) || ''

const normalizePriority = (priority) => firstText(priority, 'P3')

const normalizeStatus = (status, fallback = 'warning') => firstText(status, fallback)

const positiveNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) && number > 0 ? number : 0
}

const aiAnalysisMeta = (analysis = {}) => {
  switch (analysis.status) {
    case 'succeeded':
      return { text: '智能摘要', color: 'processing' }
    case 'partial':
      return { text: '智能分析已部分生成', color: 'warning' }
    case 'failed':
      return { text: '智能摘要暂不可用，已展示基础分析', color: 'warning' }
    case 'skipped':
      return { text: '当前无需智能摘要', color: 'default' }
    case 'not_configured':
      return { text: '基础健康分析', color: 'default' }
    default:
      return null
  }
}

const makeReportCacheKey = (params) => {
  const normalized = Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== null && value !== '')
    .sort(([leftKey], [rightKey]) => leftKey.localeCompare(rightKey))
    .map(([key, value]) => `${key}=${value}`)
    .join('&')
  return `${REPORT_STORAGE_PREFIX}${normalized}`
}

const normalizeUrlTimeParams = (searchParams) => {
  const relativeTime = searchParams.get('relativeTime')
  const startTime = firstText(searchParams.get('startTime'), ISOToTimestamp(searchParams.get('from')))
  const endTime = firstText(searchParams.get('endTime'), ISOToTimestamp(searchParams.get('to')))

  if (positiveNumber(startTime) > 0 && positiveNumber(endTime) > positiveNumber(startTime)) {
    return {
      relativeTime: undefined,
      startTime,
      endTime,
    }
  }

  if (relativeTime) {
    return {
      relativeTime,
      startTime: undefined,
      endTime: undefined,
    }
  }

  return null
}

const buildStoreTimeParams = (timeRange = {}) => {
  const rangeTypeKey = safeText(timeRange.rangeTypeKey)
  const storeStartTime = positiveNumber(timeRange.startTime)
  const storeEndTime = positiveNumber(timeRange.endTime)

  if (rangeTypeKey) {
    return {
      relativeTime: rangeTypeKey,
      startTime: undefined,
      endTime: undefined,
    }
  }

  if (storeStartTime > 0 && storeEndTime > storeStartTime) {
    return {
      relativeTime: undefined,
      startTime: storeStartTime,
      endTime: storeEndTime,
    }
  }

  return null
}

const makeTimeRangeSignature = (timeRange = {}) => [
  safeText(timeRange.rangeTypeKey),
  positiveNumber(timeRange.startTime) || '',
  positiveNumber(timeRange.endTime) || '',
].join(':')

const buildReportRequestParams = (search, timeRange = {}, options = {}) => {
  const searchParams = new URLSearchParams(search)
  const groupId = searchParams.get('groupId') || 1
  const urlTimeParams = normalizeUrlTimeParams(searchParams)
  const storeTimeParams = buildStoreTimeParams(timeRange)

  if (!options.preferStore && urlTimeParams) {
    return {
      groupId,
      ...urlTimeParams,
    }
  }

  if (storeTimeParams) {
    return {
      groupId,
      ...storeTimeParams,
    }
  }

  return {
    groupId,
    relativeTime: '15m',
    startTime: undefined,
    endTime: undefined,
  }
}

const readPersistedReport = (cacheKey) => {
  try {
    if (!ENABLE_REPORT_STORAGE) return null
    if (typeof window === 'undefined' || !window.localStorage) return null
    const raw = window.localStorage.getItem(cacheKey)
    if (!raw) return null
    const cached = JSON.parse(raw)
    if (!cached?.data?.report || !Array.isArray(cached?.data?.issues)) return null
    if (Date.now() - Number(cached.savedAt) > REPORT_STORAGE_MAX_AGE) return null
    return cached
  } catch {
    return null
  }
}

const persistReport = (cacheKey, data) => {
  try {
    if (!ENABLE_REPORT_STORAGE) return
    if (typeof window === 'undefined' || !window.localStorage) return
    window.localStorage.setItem(cacheKey, JSON.stringify({
      savedAt: Date.now(),
      data,
    }))
  } catch {
    // Browser storage can be disabled or full; the live request result is still usable.
  }
}

const formatRecordTime = (value) => {
  const text = safeText(value)
  if (!text) return '-'
  return text.replace('T', ' ').replace(/\+\d\d:\d\d$/, '')
}

const recordWindowText = (record = {}) => {
  if (positiveNumber(record.startTime) > 0 && positiveNumber(record.endTime) > positiveNumber(record.startTime)) {
    const start = new Date(Number(record.startTime) / 1000)
    const end = new Date(Number(record.endTime) / 1000)
    if (!Number.isNaN(start.getTime()) && !Number.isNaN(end.getTime())) {
      const pad = (value) => `${value}`.padStart(2, '0')
      const date = `${start.getFullYear()}-${pad(start.getMonth() + 1)}-${pad(start.getDate())}`
      const startTime = `${pad(start.getHours())}:${pad(start.getMinutes())}`
      const endTime = `${pad(end.getHours())}:${pad(end.getMinutes())}`
      return `${date} ${startTime}-${endTime}`
    }
  }
  return safeText(record.window) || '当前窗口'
}

const formatElapsedSeconds = (seconds) => {
  const value = Math.max(0, Math.floor(Number(seconds) || 0))
  if (value < 60) return `${value} 秒`
  const minutes = Math.floor(value / 60)
  const rest = value % 60
  return rest > 0 ? `${minutes} 分 ${rest} 秒` : `${minutes} 分钟`
}

const reportWaitingBlocks = [
  {
    key: 'range',
    title: '分析范围',
    description: '读取当前数据组和时间窗口，确定本次系统健康报告的分析边界。',
  },
  {
    key: 'alerts',
    title: '告警事件',
    description: '按服务、入口、实例和告警状态归并真实告警，识别重复触发和共同风险对象。',
  },
  {
    key: 'evidence',
    title: '证据链',
    description: '融合指标趋势、请求北极星、日志分析和告警事件方向报告，形成关键依据。',
  },
  {
    key: 'actions',
    title: '处置建议',
    description: '结合历史故障报告中的建议行动，生成面向系统健康中心的处置步骤。',
  },
  {
    key: 'summary',
    title: '报告摘要',
    description: '汇总系统状态、主要事件和需要关注的问题；智能分析超时时会返回基础报告。',
  },
]

const truncateText = (text = '', limit = 220) => {
  if (text.length <= limit) return text
  return `${text.slice(0, limit)}...`
}

const mergeEvidenceItems = (items = []) => {
  const grouped = new Map()
  safeList(items).forEach((item) => {
    const key = [item.type || '证据补充', item.source || ''].join('::')
    const existing = grouped.get(key)
    if (!existing) {
      grouped.set(key, {
        ...item,
        resultList: item.result ? [item.result] : [],
        count: 1,
      })
      return
    }
    existing.count += 1
    if (item.result) {
      existing.resultList.push(item.result)
    }
    if (!existing.severity && item.severity) {
      existing.severity = item.severity
    }
  })
  return Array.from(grouped.values()).map((item) => {
    const uniqueResults = Array.from(new Set(item.resultList)).filter(Boolean)
    if (item.count <= 1) {
      return {
        ...item,
        result: uniqueResults[0] || item.result || '',
      }
    }
    return {
      ...item,
      result: truncateText(`${item.count} 条同类证据已合并：${uniqueResults.slice(0, 3).join('；')}`, 260),
    }
  })
}

const summarizeExistingAlertAnalysis = (analysis = {}) => {
  const alertName = safeText(analysis.alertName)
  const alertDirection = safeText(analysis.alertDirection)
  const checkOutput = safeText(analysis.checkOutput)
  const analyzeError = safeText(analysis.analyzeError)
  const validity = safeText(analysis.validity)
  const subject = alertName ? `告警 ${alertName}` : '关联告警'
  const validityText = {
    valid: 'AI校验为有效告警',
    invalid: 'AI校验为无效告警',
    failed: '告警分析失败',
    skipped: '已跳过分析',
    unknown: '等待分析确认',
  }[validity] || ''

  if (alertDirection) {
    return `${subject}：${truncateText(alertDirection, 180)}`
  }
  if (analyzeError) {
    return `${subject}分析失败：${truncateText(analyzeError, 140)}`
  }
  if (validityText) {
    return `${subject}：${validityText}`
  }
  if (checkOutput && !['true', 'false'].includes(checkOutput.toLowerCase())) {
    return `${subject}已有校验输出：${truncateText(checkOutput, 140)}`
  }
  return ''
}

const buildExistingAlertAnalysisText = (issue = {}) => {
  const summaries = safeList(issue.existingAlertAnalyses)
    .map(summarizeExistingAlertAnalysis)
    .filter(Boolean)
  return truncateText(Array.from(new Set(summaries)).join('；'), 300)
}

const getEvidenceRole = (type) => {
  const evidenceType = safeText(type)
  if (evidenceType.includes('请求北极星')) {
    return {
      stage: '北极星耗时分解',
      effect: '基于单次请求在网络、调度、CPU 等资源上的平均耗时变化判断延时方向。',
      kind: 'metric',
    }
  }
  if (evidenceType.includes('告警事件')) {
    return {
      stage: evidenceType.includes('方向报告') ? '告警分析' : '告警汇总',
      effect: '基于当前窗口内真实告警事件确认影响目标、触发状态和告警级别。',
      kind: 'alert',
    }
  }
  if (evidenceType.includes('告警明细')) {
    return {
      stage: '告警明细',
      effect: '展示告警事件携带的详情文本，不补充告警之外的数据推断。',
      kind: 'alert',
    }
  }
  if (evidenceType.includes('告警指标趋势') || evidenceType.includes('指标趋势')) {
    return {
      stage: '指标趋势',
      effect: evidenceType.includes('方向报告') ? '来自告警事件分析报告中的指标趋势判断。' : '根据告警标签定位对应指标，并展示当前窗口变化。',
      kind: 'alert',
    }
  }
  if (evidenceType.includes('Trace链路样本') || evidenceType.includes('APM链路')) {
    return {
      stage: 'APM链路',
      effect: evidenceType.includes('方向报告') ? '来自告警事件分析报告中的链路定位。' : '结合当前窗口链路样本补充影响范围。',
      kind: 'trace',
    }
  }
  if (evidenceType.includes('日志分析')) {
    return {
      stage: '日志分析',
      effect: '来自告警事件分析报告中的日志侧结论。',
      kind: 'log',
    }
  }
  if (evidenceType.includes('数据边界')) {
    return {
      stage: '数据边界',
      effect: '说明当前结论没有使用尚未接入的数据源，避免把缺失数据推断成事实。',
      kind: 'boundary',
    }
  }
  if (evidenceType.includes('接入能力边界')) {
    return {
      stage: '接入能力',
      effect: '展示当前已配置的数据接入能力和缺口；仅作为边界说明，不直接生成根因推断。',
      kind: 'boundary',
    }
  }
  return {
    stage: evidenceType || '补充依据',
    effect: '来自当前已接入告警事件的补充信息。',
    kind: 'default',
  }
}

const getEvidenceByType = (items, keyword, options = {}) => {
  const exclude = safeList(options.exclude)
  return safeList(items).find((item) => {
    const type = item.type || ''
    return type.includes(keyword) && !exclude.some((excluded) => type.includes(excluded))
  })
}

const objectTypeText = {
  endpoint: '请求路径',
  target: '对象',
  service: '服务',
  host: '主机',
  node: '节点',
  database: '数据库',
  namespace: '命名空间',
  pod: 'Pod',
  group: '类型',
}

const formatMetricNumber = (value) => {
  if (value === undefined || value === null || Number.isNaN(Number(value))) {
    return '-'
  }
  const number = Number(value)
  if (Math.abs(number) >= 1000) {
    return number.toLocaleString(undefined, { maximumFractionDigits: 1 })
  }
  if (Math.abs(number) >= 10) {
    return number.toLocaleString(undefined, { maximumFractionDigits: 2 })
  }
  return number.toLocaleString(undefined, { maximumFractionDigits: 4 })
}

const formatPercent = (value) => {
  if (value === undefined || value === null || Number.isNaN(Number(value))) {
    return '无法计算'
  }
  return `${Number(value).toLocaleString(undefined, { maximumFractionDigits: 2 })}%`
}

const normalizeMetricTimestamp = (timestamp) => {
  const number = Number(timestamp)
  if (!Number.isFinite(number)) return null
  if (number > 100000000000000000) return Math.floor(number / 1000000)
  if (number > 100000000000000) return Math.floor(number / 1000)
  return number < 1000000000000 ? number * 1000 : number
}

const normalizeMetricPoint = (point = {}) => {
  if (!point || typeof point !== 'object') return null
  const timestamp = normalizeMetricTimestamp(firstText(point.timestamp, point.time, point.ts))
  if (timestamp === null) return null
  if (point.value === undefined || point.value === null || point.value === '') {
    return { timestamp, value: null }
  }
  const value = Number(point.value)
  return { timestamp, value: Number.isFinite(value) ? value : null }
}

const metricChartPoints = (metric = {}) =>
  safeList(metric.series).flatMap((series) => safeList(series.points)
    .map(normalizeMetricPoint)
    .filter((point) => point && point.value !== null)
    .map((point) => [point.timestamp, point.value])
    .filter(Boolean))

const metricDistinctTimestampCount = (metric = {}) => new Set(metricChartPoints(metric).map((point) => point[0])).size

const metricSampleCount = (metric = {}) => positiveNumber(metric.pointCount) || metricChartPoints(metric).length

const isLowConfidenceMetric = (metric = {}) => metricSampleCount(metric) < 3 || metricDistinctTimestampCount(metric) < 3

const metricTrendWord = (metric = {}) => {
  const delta = Number(metric.delta)
  const change = metric.changePercent === undefined || metric.changePercent === null ? null : Number(metric.changePercent)
  if (!Number.isFinite(delta)) return '无法判断'
  if (change !== null && Number.isFinite(change) && Math.abs(change) < 5) return '基本持平'
  if (delta > 0) return '有所升高'
  if (delta < 0) return '有所回落'
  return '基本持平'
}

const ruleMetricNameMappings = [
  { pattern: /平均请求延时.*相比昨天.*升高.*%.*以上/, name: '请求延时日同比' },
  { pattern: /请求错误率.*相比昨天.*升高.*%.*以上/, name: '请求错误率日同比' },
  { pattern: /平均请求延时.*超过\s*\d+(?:\.\d+)?\s*(?:s|秒|ms|毫秒)?/i, name: '平均请求延时' },
  { pattern: /请求错误率.*超过\s*\d+(?:\.\d+)?\s*%?/, name: '请求错误率' },
]

const inferMetricNameFromRuleName = (ruleName = '') => {
  const text = safeText(ruleName)
  if (!text) return ''
  return ruleMetricNameMappings.find(({ pattern }) => pattern.test(text))?.name || ''
}

const inferMetricNameFromRuleTexts = (...texts) => texts.map(inferMetricNameFromRuleName).find(Boolean) || ''

const extractRuleNameFromText = (text = '') => {
  const normalized = safeText(text)
  if (!normalized) return ''
  const matches = Array.from(normalized.matchAll(/(?:告警规则|规则|ruleName|rule_name|alertname|告警)[:=：]\s*([^；，。\n]+)/gi))
  const candidates = matches.flatMap((match) => match[1].split('/').map(safeText).filter(Boolean).reverse())
  return candidates.find((candidate) => inferMetricNameFromRuleName(candidate)) || firstText(...candidates)
}

const metricDisplayName = (metric = {}, fallbackText = '') => {
  const explicit = firstText(metric.metricName, metric.metric_name, metric.displayName, metric.display_name, metric.name)
  if (explicit) return explicit
  const ruleMetricName = inferMetricNameFromRuleTexts(metric.ruleName, metric.rule_name, extractRuleNameFromText(fallbackText))
  return ruleMetricName || '告警指标'
}

const metricBackendDisplayName = (metric = {}) => firstText(
  metric.metricName,
  metric.metric_name,
  metric.displayName,
  metric.display_name,
  metric.name,
  metric.prometheusMetricName,
  metric.prometheus_metric_name,
) || inferMetricNameFromRuleTexts(metric.ruleName, metric.rule_name) || '相关指标'

const ensureMetricNameInText = (metric = {}, text = '', fallbackText = '') => {
  const content = safeText(text)
  if (!content) return ''
  const name = metricDisplayName(metric, fallbackText)
  if (!name || content.includes(name)) {
    return content
  }
  return `${name}：${content}`
}

const isMixedMetricAnalysis = (metric = {}, text = '') => {
  const content = safeText(text)
  if (!content) return false
  const metricName = metricDisplayName(metric)
  return content.includes('请求北极星')
    || (metricName !== '请求平均耗时' && content.includes('请求平均耗时'))
    || (metricName !== '请求吞吐量' && (content.includes('TPS') || content.includes('吞吐量')))
}

const metricEvidenceDisplayText = (metricEvidence = {}, issue = {}) => {
  const metric = metricEvidence.metric
  const aiMetricAnalysis = safeText(issue.metricAnalysis)
  if (!metric) return aiMetricAnalysis

  const metricName = metricDisplayName(metric, metricEvidence.result)
  const sampleCount = metricSampleCount(metric)
  const timestampCount = metricDistinctTimestampCount(metric)
  if (!metric.eventScoped) {
    return `${metricName}未能按该告警事件的标签定位到目标序列，只能用于检查告警规则口径，不能直接说明当前影响目标的业务退化。`
  }
  if (isLowConfidenceMetric(metric)) {
    const countText = sampleCount > 0 ? `${sampleCount} 个采样点` : '少量采样点'
    return `${metricName}仅返回${countText}${timestampCount > 0 ? `、${timestampCount} 个有效时间点` : ''}，样本不足以可靠判断升高或回落；该指标只能作为复发观察信号，需要结合告警事件和日志确认影响是否持续。`
  }
  if (aiMetricAnalysis && !isMixedMetricAnalysis(metric, aiMetricAnalysis)) {
    return ensureMetricNameInText(metric, aiMetricAnalysis, metricEvidence.result)
  }

  const direction = metricTrendWord(metric)
  const rangeText = `区间 ${formatMetricNumber(metric.minValue)} 到 ${formatMetricNumber(metric.maxValue)}`
  const changeText = metric.changePercent === undefined || metric.changePercent === null
    ? ''
    : `，变化 ${formatPercent(metric.changePercent)}`
  const stateText = positiveNumber(issue.activeAlerts) > 0
    ? '当前仍有告警触发，需继续确认指标是否还在高位。'
    : '当前告警已恢复，指标更适合作为恢复稳定性和复发观察依据。'
  return `${metricName}在当前窗口${direction}${changeText}，${rangeText}。${stateText}`
}

const metricSeriesLabel = (labels = {}, index = 0) => {
  const candidates = ['svc_name', 'serviceName', 'service', 'pod', 'pod_name', 'namespace', 'content_key', 'endpoint', 'container', 'instance']
  const parts = candidates
    .filter((key) => labels[key])
    .slice(0, 3)
    .map((key) => labels[key])
  return parts.length > 0 ? parts.join(' / ') : `指标序列 ${index + 1}`
}

const metricChartSeries = (metric = {}) =>
  (() => {
    const rawSeries = safeList(metric.series)
    .map((series, index) => {
      const pointMap = new Map()
      safeList(series.points)
        .map(normalizeMetricPoint)
        .filter(Boolean)
        .forEach((point) => {
          pointMap.set(point.timestamp, point.value === null ? 0 : point.value)
        })
      const data = Array.from(pointMap.entries())
        .map(([timestamp, value]) => [timestamp, value])
        .sort((left, right) => left[0] - right[0])
      return {
        hasData: data.length > 0,
        name: metricSeriesLabel(series.labels, index),
        pointMap,
        data,
      }
    })
    .filter((series) => series.hasData)
    const allTimestamps = Array.from(new Set(rawSeries.flatMap((series) => series.data.map((point) => point[0]))))
      .sort((left, right) => left - right)
    const alignPointLimit = positiveNumber(metric.pointLimit) || 160
    const shouldAlignGaps = allTimestamps.length > 0
      && allTimestamps.length <= alignPointLimit
      && rawSeries.some((series) => series.pointMap.size < allTimestamps.length)
    if (!shouldAlignGaps) {
      return rawSeries.map(({ name, data }) => ({ name, data }))
    }
    return rawSeries.map(({ name, pointMap }) => ({
      name,
      data: allTimestamps.map((timestamp) => [
        timestamp,
        pointMap.has(timestamp) ? pointMap.get(timestamp) : 0,
      ]),
    }))
  })()

const hasMetricChartData = (metric) => metricChartSeries(metric).length > 0

const MetricTrendChart = ({ metric, compact = false }) => {
  const series = metricChartSeries(metric)
  if (series.length === 0) return null
  const allTimestamps = series.flatMap((item) => item.data.map((point) => point[0]))
  const minTime = Math.min(...allTimestamps)
  const maxTime = Math.max(...allTimestamps)
  const rangeMs = Number.isFinite(minTime) && Number.isFinite(maxTime) ? maxTime - minTime : 0
  const lowConfidence = isLowConfidenceMetric(metric)
  const metricNotes = [
    lowConfidence && '采样点较少，趋势仅作参考',
    metric.pointsTruncated && '已按窗口抽样展示',
  ].filter(Boolean)
  const timeLabelFormatter = (value) => {
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ''
    const timeText = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    if (rangeMs >= 24 * 60 * 60 * 1000) {
      return `${date.toLocaleDateString([], { month: '2-digit', day: '2-digit' })} ${timeText}`
    }
    return timeText
  }
  const option = {
    color: ChartColorList,
    backgroundColor: 'transparent',
    animation: false,
    grid: {
      left: 8,
      right: 8,
      top: compact ? 12 : 18,
      bottom: compact ? 18 : 28,
      containLabel: true,
    },
    tooltip: {
      trigger: 'axis',
      confine: true,
      valueFormatter: (value) => formatMetricNumber(value),
    },
    legend: {
      show: !compact && series.length > 1,
      type: 'scroll',
      bottom: 0,
      icon: 'roundRect',
      itemWidth: 10,
      itemHeight: 8,
      textStyle: {
        color: 'var(--ant-color-text-secondary)',
        fontSize: 11,
      },
    },
    xAxis: {
      type: 'time',
      boundaryGap: false,
      axisLine: { lineStyle: { color: 'var(--ant-color-border)' } },
      axisTick: { show: false },
      axisLabel: {
        color: 'var(--ant-color-text-tertiary)',
        fontSize: 11,
        hideOverlap: true,
        formatter: timeLabelFormatter,
      },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      scale: true,
      axisLabel: {
        color: 'var(--ant-color-text-tertiary)',
        fontSize: 11,
        formatter: (value) => formatMetricNumber(value),
      },
      splitLine: { lineStyle: { color: 'var(--ant-color-border-secondary)', type: 'dashed' } },
    },
    series: series.map((item) => ({
      name: item.name,
      type: 'line',
      smooth: !lowConfidence,
      showSymbol: lowConfidence || item.data.length <= 8,
      symbolSize: 4,
      connectNulls: true,
      lineStyle: { width: compact ? 2 : 2.5 },
      emphasis: { focus: 'series' },
      data: item.data,
    })),
  }

  return (
    <div className={compact ? 'health-metric-trend-chart health-metric-trend-chart-compact' : 'health-metric-trend-chart'}>
      <ReactECharts option={option} style={{ width: '100%', height: compact ? 120 : 180 }} />
      {metricNotes.length > 0 && (
        <div className="health-metric-trend-notes">
          {metricNotes.map((note) => (
            <span
              className={note.includes('缺失') ? 'health-metric-trend-note health-metric-trend-note-secondary' : 'health-metric-trend-note'}
              key={note}
            >
              {note}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

MetricTrendChart.propTypes = {
  metric: PropTypes.object,
  compact: PropTypes.bool,
}

const stripTechnicalEvidenceText = (text = '') =>
  truncateText(
    text
      .replace(/^方向报告中的\s*/g, '')
      .replace(/^APM\s*链路分析显示[:：]\s*/g, '')
      .replace(/^日志分析显示[:：]\s*/g, '')
      .replace(/^指标趋势显示[:：]\s*/g, '')
      .replace(/^告警事件分析显示[:：]\s*/g, '')
      .replace(/\d+\s*条真实告警事件，?/g, '')
      .replace(/\d+\s*个去重告警实例，?/g, '')
      .replace(/\d+\s*类告警，?/g, '')
      .replace(/状态=[^，。；]+[，。；]?/g, '')
      .replace(/聚合对象=[^，。；]+[，。；]?/g, '')
      .replace(/告警类型=[^，。；]+[，。；]?/g, '')
      .replace(/重复触发=\d+次[，。；]?/g, '')
      .replace(/最长持续=[^，。；]+[，。；]?/g, '')
      .replace(/PQL=[^；]+；?/g, '')
      .replace(/规则=[^；]+；?/g, '')
      .replace(/匹配标签=[^；]+；?/g, '')
      .replace(/通过 [^；]+?定位当前触发 (series|序列)；?/g, '')
      .replace(/告警=[^；]+；?/g, '')
      .replace(/raw_tags/g, '告警标签')
      .replace(/Prometheus series/g, '指标序列')
      .replace(/series/g, '序列')
      .replace(/alertName \+ 告警标签/g, '告警标签')
      .replace(/alertName \+ raw_tags/g, '告警标签')
      .replace(/span_trace/g, '链路样本'),
    150,
  )

const summarizeMetricEvidenceText = (text = '') => {
  const normalized = text
    .replace(/^方向报告中的指标趋势显示[:：]\s*/g, '')
    .replace(/series/g, '序列')
  if (normalized.includes('没有命中告警事件对应的 序列') || normalized.includes('没有命中告警事件对应的序列')) {
    return ''
  }
  if (normalized.includes('未找到告警规则')) {
    return ''
  }
  if (normalized.includes('查询窗口趋势失败')) {
    return ''
  }
  const matched = normalized.match(/匹配当前 序列=([^；]+)|匹配 序列=([^；]+)|匹配序列=([^；]+)/)
  const change = normalized.match(/变化=([^；]+?)(?:（([^）]+)）)?(?:；|$)/)
  const parts = ['已命中告警对应指标']
  if (matched) {
    parts.push('已按事件标签定位')
  }
  if (change) {
    parts.push(`窗口变化 ${change[2] || change[1]}`)
  }
  if (parts.length > 1) {
    return `${parts.join('，')}。`
  }
  return stripTechnicalEvidenceText(text)
}

const metricEvidenceUnavailableText = (text = '') => {
  const normalized = safeText(text).replace(/series/g, '序列')
  if (!normalized) return ''
  if (normalized.includes('查询窗口趋势失败')) {
    return '指标趋势暂不可用：当前无法连接指标查询服务，本次未使用指标趋势判断该风险。'
  }
  if (normalized.includes('未找到告警规则')) {
    return '指标趋势暂不可用：未找到该告警对应的规则表达式，无法回看指标趋势。'
  }
  if (normalized.includes('规则未配置 PQL') || normalized.includes('未配置 PQL')) {
    return '指标趋势暂不可用：告警规则未配置指标查询表达式，无法回看指标趋势。'
  }
  if (normalized.includes('没有命中告警事件对应的 序列') || normalized.includes('没有命中告警事件对应的序列')) {
    return '指标趋势未能按当前告警标签定位到目标序列，暂不作为该风险的趋势依据。'
  }
  if (normalized.includes('规则查询未返回可用时序点')) {
    return '指标趋势暂不可用：规则查询未返回可用时序点，本次未使用指标趋势判断该风险。'
  }
  return ''
}

const summarizeTraceEvidenceText = (text = '') =>
  truncateText(
    text
      .replace(/^方向报告中的\s*APM\s*链路分析显示[:：]\s*/g, '')
      .replace(/^按告警.*?反查\s*/, '')
      .replace(/span_trace[:：]?/g, '链路样本：')
      .replace(/service=[^；]+；?/g, '')
      .replace(/endpoint=[^；]+；?/g, '')
      .replace(/span 样本/g, '链路样本')
      .replace(/本页\d+条样本中\s*/g, '')
      .replace(/error=/g, '错误=')
      .replace(/slow=/g, '慢调用=')
      .replace(/、代表trace=[^；。]+/g, '')
      .replace(/trace=[^；。]+/g, '')
      .replace(/。该证据.*$/g, '。')
      .replace(/\s{2,}/g, ' ')
      .trim(),
    150,
  )

const summarizeRequestPolarisEvidenceText = (item = {}) => {
  const metric = item.metric
  const metricName = metricBackendDisplayName(metric)
  const text = safeText(item.result)
    .replace(/请求北极星全局视角/g, '请求北极星')
    .replace(/调用链慢 span/g, '慢调用样本')
    .replace(/慢 span/g, '慢调用样本')
    .replace(/span/g, '调用样本')
    .replace(/req\/s/g, '次/秒')
    .replace(/\s{2,}/g, ' ')
    .trim()
  const summary = productReportSummary(stripTechnicalEvidenceText(text), false, 180)
  if (!metricName || summary.includes(metricName)) return summary
  return truncateText(`${metricName}：${summary}`, 200)
}

const productStructuredSummary = (text = '', limit = 260) =>
  truncateText(text
    .replace(/^AI\s*/, '')
    .replace(/AI\s*/g, '')
    .replace(/影响对象/g, '故障目标')
    .replace(/、?\d+\s*类已接入数据源/g, '')
    .replace(/，?计算路径风险\s*[\d.]+/g, '')
    .replace(/\s*未接入的.*?不参与推断。?/g, '')
    .replace(/\s{2,}/g, ' ')
    .trim(), limit)

const productReportSummary = (text = '', aiGenerated = false, limit = 260) => {
  const summary = safeText(text).trim()
  if (aiGenerated) return truncateText(summary, limit)
  return productStructuredSummary(summary, limit)
}

const productIssueConclusion = (issue = {}, aiGenerated = false, limit = 260) =>
  productReportSummary(firstText(issue.aiConclusion, issue.systemProblem, issue.impact, issue.mergeBasis), aiGenerated, limit)

const issueSymptomText = (issue = {}) => {
  const text = [
    issue.title,
    issue.scope,
    issue.rootCause,
    issue.aiConclusion,
    issue.systemProblem,
    issue.impact,
    issue.mergeBasis,
    safeList(issue.evidence).map((item) => item.result).join(' '),
  ].join(' ').toLowerCase()
  if (/(latency|duration|timeout|slow|p95|p99|耗时|延时|延迟|慢调用|超时)/.test(text)) return '请求延时'
  if (/(error|exception|fail|失败|错误|异常|5xx|4xx)/.test(text)) return '错误异常'
  if (/(cpu|memory|oom|磁盘|内存|负载|资源)/.test(text)) return '资源压力'
  if (/(unavailable|down|不可用|宕机|断连)/.test(text)) return '可用性'
  if (/(规则|阈值|表达式|无效告警|invalid)/.test(text)) return '告警规则'
  return '健康风险'
}

const topCountLabel = (items = []) => {
  const counts = new Map()
  items.forEach((item) => counts.set(item, (counts.get(item) || 0) + 1))
  const sorted = Array.from(counts.entries()).sort((left, right) => right[1] - left[1])
  return sorted[0]?.[0] || ''
}

const reportScopePatternLabel = (issues = []) => {
  const types = new Set(issues.flatMap((issue) => safeList(issue.affectedObjects).map((object) => safeText(object.type))))
  const issueCount = issues.length
  if (types.has('endpoint') && types.has('service') && issueCount > 3) return '多入口/多服务'
  if (types.has('endpoint')) return issueCount > 1 ? '多入口' : '单入口'
  if (types.has('service')) return issueCount > 1 ? '多服务' : '单服务'
  if (types.has('host')) return issueCount > 1 ? '多主机' : '单主机'
  if (types.has('node')) return issueCount > 1 ? '多节点' : '单节点'
  if (types.has('pod')) return issueCount > 1 ? '多实例' : '单实例'
  return issueCount > 1 ? '多目标' : '单目标'
}

const isRuleDiagnosticIssue = (issue = {}) => {
  const text = [
    issue.priority,
    issue.title,
    issue.rootCause,
    issue.aiConclusion,
    issue.systemProblem,
    issue.impact,
    issue.mergeBasis,
    safeList(issue.evidence).map((item) => item.result).join(' '),
  ].join(' ').toLowerCase()
  return normalizePriority(issue.priority) === 'P4' || /(invalidalert|无效告警|告警规则诊断|规则有效性|red 指标背离|red指标背离)/.test(text)
}

const reportEventDescriptionText = (issues = [], activeCount = 0, recoveringCount = 0) => {
  const totalEvents = issues.reduce((sum, issue) => sum + positiveNumber(issue.rawEventCount), 0)
  const totalInstances = issues.reduce((sum, issue) => sum + positiveNumber(issue.deduplicatedAlerts), 0)
  const repeatSeen = totalEvents >= 5 || (totalInstances > 0 && totalEvents > totalInstances)
  const symptom = topCountLabel(issues.map(issueSymptomText))
  const scopePattern = reportScopePatternLabel(issues)
  const eventPrefix = totalEvents > 0 ? '窗口内发生过' : '当前窗口识别到'
  const repeatText = repeatSeen ? '，并出现反复触发' : ''
  if (activeCount > 0) {
    return `${eventPrefix}${scopePattern}${symptom}告警${repeatText}，目前仍有告警未恢复`
  }
  if (recoveringCount > 0) {
    return `${eventPrefix}${scopePattern}${symptom}告警${repeatText}，目前告警已清零，相关影响目标处于恢复观察`
  }
  return `当前窗口未发现持续触发的核心告警，系统处于观察状态`
}

const reportExecutiveSummary = (report = {}, issues = [], fallback = '') => {
  if (issues.length === 0) return productReportSummary(fallback)
  const activeCount = positiveNumber(report.activeAlertCount)
  const recoveringCount = issues.filter((issue) => normalizeStatus(issue.status) === 'recovering' || positiveNumber(issue.activeAlerts) === 0).length
  const ruleDiagnosticCount = issues.filter(isRuleDiagnosticIssue).length
  const healthText = LEVEL_META[report.status]?.text || report.status || '异常'
  const eventDescription = reportEventDescriptionText(issues, activeCount, recoveringCount)
  const focus = ruleDiagnosticCount > 0
    ? '需要关注告警规则口径是否与服务 RED 指标背离，避免把无效告警当作服务故障处理'
    : activeCount > 0
      ? '需要关注告警影响是否继续扩大，并优先确认入口、下游和关键指标的恢复情况'
      : '需要关注告警是否复发、服务 RED 指标是否退化，以及告警分析是否指向共同根因'
  return `系统当前处于${healthText}状态。${eventDescription}。${focus}。`
}

const issueMetaItems = (issue = {}) => [
  positiveNumber(issue.rawEventCount) > 0 && `事件 ${issue.rawEventCount}`,
  positiveNumber(issue.deduplicatedAlerts) > 0 && `实例 ${issue.deduplicatedAlerts}`,
  positiveNumber(issue.activeAlerts) > 0 && `触发中 ${issue.activeAlerts}`,
  issue.startedAt && `开始 ${issue.startedAt}`,
].filter(Boolean)

const actionPlanSourceText = (steps = []) => {
  const sources = new Set(safeList(steps).map((step) => safeText(step.source)).filter(Boolean))
  if (sources.has('alertReport')) return '来自告警事件分析报告'
  if (sources.has('ai')) return '来自智能分析'
  if (sources.has('system')) return '基于当前告警状态生成'
  return '基于当前风险生成'
}

const isInternalIssueKeyText = (text = '', issue = {}) => {
  const normalized = safeText(text).toLowerCase()
  if (!normalized) return false
  if (normalized === safeText(issue.key).toLowerCase()) return true
  return /^(activealert|recoveredalert|invalidalert|inspection)-(entry|service|object|group)-[a-z0-9-]+$/.test(normalized)
}

const displayActionTarget = (target = '', issue = {}) => {
  const text = safeText(target)
  if (!text) return ''
  if (isInternalIssueKeyText(text, issue)) {
    return firstText(issue.scope, issue.title)
  }
  return text
}

const shouldHideDefaultEvidence = (item = {}) => {
  const type = safeText(item.type)
  const result = safeText(item.result)
  return type.includes('接入能力边界')
    || type.includes('数据边界')
    || result.includes('查询边界')
    || result.includes('查询失败')
}

const visibleObjectTags = (issue = {}) => {
  const scope = firstText(issue.scope, issue.title)
  const normalizedScope = scope.toLowerCase()
  const seen = new Set()
  return safeList(issue.affectedObjects).filter((object) => {
    const name = safeText(object.name)
    const type = safeText(object.type)
    if (!name || !type) return false
    const key = `${type}:${name}`
    if (seen.has(key)) return false
    seen.add(key)
    if (type === 'group') return false
    if (type === 'target') return false
    if (type === 'service' && normalizedScope.includes(name.toLowerCase())) return false
    if (type === 'endpoint' && normalizedScope.includes(name.toLowerCase())) return false
    return true
  }).slice(0, 3)
}

const HTTP_METHOD_PATTERN = /^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+\//i

const buildScopeImpactObjects = (scope = '', title = '', affectedObjects = []) => {
  const scopeText = safeText(scope)
  if (!scopeText || scopeText === safeText(title)) return []

  const parts = scopeText.split(/\s+\/\s+/).map(safeText).filter(Boolean)
  if (parts.length >= 2) {
    const [serviceName, ...rest] = parts
    const targetName = rest.join(' / ')
    const objectTypes = new Set(safeList(affectedObjects).map((object) => safeText(object.type)))
    const looksLikeServiceRoute = objectTypes.has('service')
      || objectTypes.has('endpoint')
      || HTTP_METHOD_PATTERN.test(targetName)
    if (!looksLikeServiceRoute) {
      return [{ type: 'scope', label: '范围', name: scopeText }]
    }
    const items = []
    if (serviceName) {
      items.push({ type: 'service', label: objectTypeText.service, name: serviceName })
    }
    if (targetName) {
      items.push({
        type: HTTP_METHOD_PATTERN.test(targetName) ? 'endpoint' : 'target',
        label: HTTP_METHOD_PATTERN.test(targetName) ? objectTypeText.endpoint : objectTypeText.target,
        name: targetName,
      })
    }
    return items
  }

  return [{ type: 'scope', label: '范围', name: scopeText }]
}

const buildImpactObjects = (displayTitle = '', scope = '', visibleAffectedObjects = [], affectedObjects = visibleAffectedObjects) => {
  const seen = new Set()
  const objects = []
  const addObject = (object = {}) => {
    const type = safeText(object.type)
    const name = safeText(object.name)
    if (!name) return
    const label = safeText(object.label) || objectTypeText[type] || type || '对象'
    const key = `${label}:${name}`.toLowerCase()
    if (seen.has(key)) return
    seen.add(key)
    objects.push({ type, label, name })
  }

  buildScopeImpactObjects(scope, displayTitle, affectedObjects).forEach(addObject)
  safeList(visibleAffectedObjects).forEach((object) => {
    addObject({
      ...object,
      label: objectTypeText[object.type] || object.type,
    })
  })

  const maxVisibleObjects = 4
  return {
    items: objects.slice(0, maxVisibleObjects),
    overflowCount: Math.max(0, objects.length - maxVisibleObjects),
  }
}

const summarizeEvidenceItem = (item = {}) => {
  const type = safeText(item.type)
  if (type.includes('请求北极星')) {
    return summarizeRequestPolarisEvidenceText(item)
  }
  if (type.includes('告警指标趋势') || type.includes('指标趋势')) {
    return summarizeMetricEvidenceText(item.result || '')
  }
  if (type.includes('Trace链路样本') || type.includes('APM链路')) {
    return summarizeTraceEvidenceText(item.result || '')
  }
  if (type.includes('日志分析')) {
    return productReportSummary(stripTechnicalEvidenceText((item.result || '').replace(/^方向报告中的日志分析显示[:：]\s*/g, '')))
  }
  if (type.includes('告警事件分析')) {
    return productReportSummary(stripTechnicalEvidenceText((item.result || '').replace(/^方向报告中的告警事件分析显示[:：]\s*/g, '')))
  }
  if (type.includes('告警事件') || type.includes('告警明细')) {
    return stripTechnicalEvidenceText(item.result || '')
  }
  return productReportSummary(stripTechnicalEvidenceText(item.result || ''))
}

const rawLogCandidateKeys = [
  'rawLogs',
  'rawLogLines',
  'logLines',
  'logs',
  'samples',
  'rawSamples',
  'entries',
]

const formatRawLogObject = (value = {}) => [
  firstText(value.timestamp, value.time, value.ts, value.datetime),
  firstText(value.level, value.severity),
  firstText(value.service, value.serviceName, value.svc_name),
  firstText(value.pod, value.podName, value.instance, value.node),
  firstText(value.message, value.content, value.log, value.body, value.text, value.stack),
].filter(Boolean).join(' ')

const normalizeRawLogLines = (value) => {
  if (!value) return []
  if (Array.isArray(value)) {
    return value.flatMap(normalizeRawLogLines)
  }
  if (typeof value === 'object') {
    const formatted = formatRawLogObject(value)
    return formatted ? [formatted] : []
  }
  return `${value}`
    .split(/\r?\n/)
    .map((line) => line.trimEnd())
    .filter((line) => line.trim())
}

const extractRawLogLines = (item = {}) => {
  const lines = rawLogCandidateKeys.flatMap((key) => normalizeRawLogLines(item[key]))
  return Array.from(new Set(lines.map((line) => truncateText(line, 500)))).slice(0, 80)
}

const inferRiskSymptom = (issue = {}, alertEvidence = {}, detailEvidence = {}) => {
  const evidenceText = [
    issue.title,
    issue.scope,
    issue.systemProblem,
    issue.aiConclusion,
    issue.impact,
    issue.mergeBasis,
    alertEvidence.result,
    detailEvidence.result,
    safeList(alertEvidence.alertEvents).map((event) => [
      event.name,
      event.summary,
      safeList(Object.values(event.labels || {})).join(' '),
    ].join(' ')).join(' '),
  ].join(' ').toLowerCase()

  if (/(latency|duration|timeout|slow|p95|p99|耗时|延时|延迟|慢调用|超时)/.test(evidenceText)) {
    return '延时类'
  }
  if (/(error|exception|fail|失败|错误|异常|5xx|4xx)/.test(evidenceText)) {
    return '错误类'
  }
  if (/(cpu|memory|oom|磁盘|内存|负载|资源)/.test(evidenceText)) {
    return '资源类'
  }
  if (/(unavailable|down|不可用|宕机|断连)/.test(evidenceText)) {
    return '可用性类'
  }
  return '健康风险'
}

const issueObjectLabel = (issue = {}) => {
  const objectTypes = new Set(safeList(issue.affectedObjects).map((object) => safeText(object.type)))
  const scope = firstText(issue.scope, issue.title)
  if (objectTypes.has('endpoint') || /\b(GET|POST|PUT|DELETE|PATCH)\b/i.test(scope)) {
    return '请求路径'
  }
  if (objectTypes.has('service')) {
    return '服务'
  }
  if (objectTypes.has('host')) {
    return '主机'
  }
  if (objectTypes.has('node')) {
    return '节点'
  }
  if (objectTypes.has('pod')) {
    return 'Pod'
  }
  return '影响目标'
}

const ALERT_TREND_WINDOW_MS = 3 * 24 * 60 * 60 * 1000
const ALERT_TREND_DAY_MS = 24 * 60 * 60 * 1000
const ALERT_TREND_HOUR_MS = 60 * 60 * 1000
const ALERT_TREND_SHORT_MS = 15 * 60 * 1000
const ALERT_TREND_CLUSTER_GAP_MS = 90 * 60 * 1000
const alertTrendCandidateKeys = [
  'serviceAlertTrend',
  'alertTrend',
  'alertHistory',
  'recentAlertTrend',
  'serviceAlertHistory',
]

const alertActivityLabel = ({ eventCount = 0, instanceCount = 0, alertClassCount = 0 }) => {
  if (eventCount >= 10) return '高频触发'
  if (eventCount >= 3) return '多次触发'
  if (eventCount >= 2) return '重复触发'
  if (eventCount === 1) return '单次触发'
  if (alertClassCount > 1 || instanceCount > 1) return '存在多个告警对象'
  return '触发情况待确认'
}

const priorityRank = {
  P0: 0,
  P1: 1,
  P2: 2,
  P3: 3,
  P4: 4,
}

const issueServiceName = (issue = {}) => {
  const serviceObject = safeList(issue.affectedObjects).find((object) => safeText(object.type) === 'service')
  if (serviceObject?.name) return safeText(serviceObject.name)
  const explicitName = firstText(issue.service, issue.serviceName, issue.resourceName, issue.objectName, issue.targetName)
  if (explicitName) return explicitName
  const primaryObject = safeList(issue.affectedObjects).find((object) => ['host', 'node', 'pod', 'namespace'].includes(safeText(object.type)))
  if (primaryObject?.name) return safeText(primaryObject.name)
  const scope = safeText(issue.scope)
  return firstText(scope.split(/\s+\/\s+/)[0], issue.title)
}

const issueFocusTarget = (issue = {}) => {
  const affectedObjects = safeList(issue.affectedObjects)
  const serviceObject = affectedObjects.find((object) => safeText(object.type) === 'service' && safeText(object.name))
  if (serviceObject) {
    return { label: '服务', name: safeText(serviceObject.name) }
  }
  const objectTypes = new Set(affectedObjects.map((object) => safeText(object.type)))
  const explicitName = firstText(issue.service, issue.serviceName, issue.resourceName, issue.objectName, issue.targetName)
  if (explicitName) {
    return {
      label: objectTypes.has('host') ? '主机池' : objectTypes.has('node') ? '节点组' : '对象',
      name: explicitName,
    }
  }
  const primaryObject = affectedObjects.find(
    (object) => ['host', 'node', 'pod', 'database', 'namespace', 'target'].includes(safeText(object.type)) && safeText(object.name),
  )
  if (primaryObject) {
    return {
      label: objectTypeText[safeText(primaryObject.type)] || '对象',
      name: safeText(primaryObject.name),
    }
  }
  const fallback = issueServiceName(issue)
  return fallback ? { label: '对象', name: fallback } : null
}

const focusTargetKey = (target = {}) => (target?.name ? `${target.label}:${target.name}` : '')

const buildTopFocusTargets = (issues = []) => {
  const groups = new Map()
  safeList(issues).forEach((issue) => {
    const target = issueFocusTarget(issue)
    if (!target?.name) return
    const priority = normalizePriority(issue.priority)
    const key = focusTargetKey(target)
    const item = groups.get(key) || {
      activeAlerts: 0,
      eventCount: 0,
      issueCount: 0,
      label: target.label,
      name: target.name,
      priority,
      priorityOrder: priorityRank[priority] ?? 9,
      scoreImpact: 0,
    }
    const order = priorityRank[priority] ?? 9
    item.issueCount += 1
    item.activeAlerts += positiveNumber(issue.activeAlerts)
    item.eventCount += positiveNumber(issue.rawEventCount)
    item.scoreImpact += Math.abs(Number(issue.scoreImpact) || 0)
    if (order < item.priorityOrder) {
      item.priority = priority
      item.priorityOrder = order
    }
    groups.set(key, item)
  })

  return Array.from(groups.values())
    .sort((left, right) => (
      left.priorityOrder - right.priorityOrder
      || right.activeAlerts - left.activeAlerts
      || right.scoreImpact - left.scoreImpact
      || right.eventCount - left.eventCount
      || right.issueCount - left.issueCount
      || left.name.localeCompare(right.name)
    ))
    .slice(0, 3)
}

const parseAlertTime = (value) => {
  const text = safeText(value)
  if (!text) return null
  const numeric = Number(text)
  if (Number.isFinite(numeric) && numeric > 0) {
    return numeric < 1000000000000 ? numeric * 1000 : numeric
  }
  const normalized = text.includes('T') ? text : text.replace(' ', 'T')
  const timestamp = Date.parse(normalized)
  return Number.isNaN(timestamp) ? null : timestamp
}

const isActiveTrendSegment = (segment = {}) => {
  if (segment.isActive === true || segment.active === true || segment.ongoing === true) return true
  const statusText = firstText(segment.status, segment.state, segment.lifecycle).toLowerCase()
  return /(active|firing|triggering|unresolved|open|触发中|未恢复|活跃)/.test(statusText)
}

const formatDurationBrief = (seconds) => {
  const value = Math.max(0, Math.floor(Number(seconds) || 0))
  if (value <= 0) return ''
  if (value < 60) return `${value} 秒`
  const minutes = Math.floor(value / 60)
  if (minutes < 60) return `${minutes} 分钟`
  const hours = Math.floor(minutes / 60)
  const restMinutes = minutes % 60
  if (hours < 24) return restMinutes > 0 ? `${hours} 小时 ${restMinutes} 分钟` : `${hours} 小时`
  const days = Math.floor(hours / 24)
  const restHours = hours % 24
  return restHours > 0 ? `${days} 天 ${restHours} 小时` : `${days} 天`
}

const alertTrendSource = (issue = {}) => alertTrendCandidateKeys
  .map((key) => issue[key])
  .find((value) => value && typeof value === 'object') || null

const alertTrendSegmentType = (segment = {}) => {
  const text = [
    segment.category,
    segment.type,
    segment.alertType,
    segment.kind,
    segment.name,
    segment.alertName,
    segment.ruleName,
    segment.title,
    segment.summary,
  ].map(safeText).join(' ').toLowerCase()
  if (/(latency|duration|timeout|slow|p95|p99|延时|延迟|慢调用|超时)/.test(text)) return '延迟'
  if (/(error|exception|fail|5xx|4xx|错误|异常|失败)/.test(text)) return '错误率'
  if (/(log|日志)/.test(text)) return '日志异常'
  if (/(cpu|memory|oom|disk|load|资源|内存|磁盘|负载)/.test(text)) return '资源压力'
  if (/(available|unavailable|down|不可用|宕机|断连)/.test(text)) return '可用性'
  return firstText(segment.category, segment.type, segment.alertType) || '告警信号'
}

const alertTrendRawSegments = (trend = {}) => [
  ...safeList(trend.segments),
  ...safeList(trend.windows),
  ...safeList(trend.intervals),
  ...safeList(trend.activityWindows),
  ...safeList(trend.events),
  ...safeList(trend.alerts),
]

const normalizeAlertTrendSegments = (trend = {}) => alertTrendRawSegments(trend)
  .map((segment, index) => {
    const start = parseAlertTime(firstText(segment.startTime, segment.startedAt, segment.start, segment.from, segment.time, segment.timestamp))
    const explicitDuration = positiveNumber(segment.durationSeconds || segment.duration || segment.activeDurationSeconds)
    const end = parseAlertTime(firstText(segment.endTime, segment.endedAt, segment.resolvedAt, segment.end, segment.to))
    const trendEnd = parseAlertTime(firstText(trend.endTime, trend.to)) || Date.now()
    const isActive = isActiveTrendSegment(segment)
    const safeStart = start || (end && explicitDuration ? end - explicitDuration * 1000 : null)
    const safeEnd = end || (isActive ? trendEnd : safeStart && explicitDuration ? safeStart + explicitDuration * 1000 : null)
    if (!safeStart && !safeEnd) return null
    const segmentStart = safeStart || safeEnd - 5 * 60 * 1000
    const segmentEnd = Math.max(segmentStart + 60 * 1000, safeEnd || segmentStart + Math.max(explicitDuration, 5 * 60) * 1000)
    return {
      key: `${segmentStart}-${segmentEnd}-${index}`,
      start: segmentStart,
      end: segmentEnd,
      count: positiveNumber(segment.count || segment.eventCount || segment.alertCount) || 1,
      title: firstText(segment.summary, segment.title, segment.alertName, segment.name, segment.ruleName, alertTrendSegmentType(segment)),
      type: alertTrendSegmentType(segment),
      severity: safeText(segment.severity, segment.priority, segment.level),
      isActive,
    }
  })
  .filter(Boolean)
  .sort((left, right) => left.start - right.start)

const alertTrendWindow = (trend = {}, segments = []) => {
  const explicitEnd = parseAlertTime(firstText(trend.endTime, trend.to))
  const explicitStart = parseAlertTime(firstText(trend.startTime, trend.from))
  const segmentEnd = segments.length > 0 ? Math.max(...segments.map((segment) => segment.end)) : 0
  const end = explicitEnd || segmentEnd || Date.now()
  return {
    end,
    start: explicitStart || end - ALERT_TREND_WINDOW_MS,
    text: safeText(trend.window, trend.windowText, trend.rangeText) || '最近 3 天',
  }
}

const alertTypeListText = (trend = {}, segments = [], fallback = '') => {
  const types = [
    ...safeList(trend.alertTypes).map((item) => safeText(item.name, item.type, item.label, item)),
    ...segments.map((segment) => segment.type),
    fallback,
  ].filter(Boolean)
  return Array.from(new Set(types)).slice(0, 3).join('、')
}

const buildAlertTrendEpisodes = (segments = [], trend = {}) => {
  const gap = positiveNumber(trend.clusterGapSeconds) > 0
    ? positiveNumber(trend.clusterGapSeconds) * 1000
    : ALERT_TREND_CLUSTER_GAP_MS
  const groups = []
  safeList(segments).forEach((segment) => {
    const last = groups[groups.length - 1]
    if (!last || segment.start - last.end > gap) {
      groups.push({
        start: segment.start,
        end: segment.end,
        segments: [segment],
      })
      return
    }
    last.end = Math.max(last.end, segment.end)
    last.segments.push(segment)
  })

  return groups.map((group, index) => {
    const count = group.segments.reduce((sum, segment) => sum + (segment.count || 1), 0)
    const types = Array.from(new Set(group.segments.map((segment) => segment.type).filter(Boolean)))
    const duration = Math.max(0, group.end - group.start)
    const active = group.segments.some((segment) => segment.isActive)
    const kind = count > 1 || group.segments.length > 1
      ? 'cluster'
      : duration <= ALERT_TREND_SHORT_MS
        ? 'short'
        : 'sustained'
    return {
      key: `${group.start}-${group.end}-${index}`,
      start: group.start,
      end: group.end,
      count,
      duration,
      kind,
      active,
      title: group.segments.map((segment) => segment.title).slice(0, 3).join('；'),
      type: types[0] || '告警信号',
      types,
    }
  })
}

const alertTrendActivityText = (trend = {}, segments = []) => {
  const count = positiveNumber(
    trend.occurrenceCount
      || trend.eventCount
      || trend.alertCount
      || trend.totalCount
      || trend.rawEventCount,
  ) || segments.length
  const explicit = firstText(trend.activityText, trend.activitySummary)
  if (explicit) {
    return explicit
  }
  if (count >= 8 || segments.length >= 5) return '多次集中出现'
  if (count >= 3 || segments.length >= 3) return '间歇性反复'
  if (count > 0 || segments.length > 0) return '少量告警信号'
  return '缺少历史告警活动数据'
}

const alertTrendDurationText = (trend = {}, segments = []) => {
  const explicitDuration = positiveNumber(
    trend.longestDurationSeconds
      || trend.maxDurationSeconds
      || trend.longestActiveSeconds
      || trend.durationSeconds,
  )
  const segmentDuration = segments.length > 0
    ? Math.max(...segments.map((segment) => Math.max(0, Math.floor((segment.end - segment.start) / 1000))))
    : 0
  const duration = formatDurationBrief(explicitDuration || segmentDuration)
  if (!duration) return '持续特征待补充'
  return `最长活跃约 ${duration}`
}

const alertTrendProblemText = ({ activeCount = 0, activityText = '', durationText = '', typeText = '', trend = {} }) => {
  const explicit = firstText(trend.problem, trend.reflectedProblem, trend.conclusion, trend.insight)
  if (explicit) return explicit
  const longRunning = /小时|天/.test(durationText)
  const repeated = /多次|反复|集中/.test(activityText)
  if (activeCount > 0 && repeated && longRunning) {
    return '更像服务稳定性持续波动，而不是单条规则的孤立触发。'
  }
  if (activeCount > 0 && repeated) {
    return '相关告警仍在活跃，应从服务整体确认影响面是否扩大。'
  }
  if (repeated) {
    return '近期存在复发迹象，恢复后仍需要观察同类告警是否再次集中出现。'
  }
  if (typeText.includes('错误率') && typeText.includes('延迟')) {
    return '更接近入口服务稳定性波动，需要结合业务入口表现判断影响。'
  }
  return '目前更像局部告警信号，需要更多历史窗口判断是否形成稳定性趋势。'
}

const buildAlertBehaviorAnalysis = (issue = {}, alertEvidence = {}, detailEvidence = {}) => {
  const eventCount = positiveNumber(issue.rawEventCount)
  const instanceCount = positiveNumber(issue.deduplicatedAlerts)
  const activeCount = positiveNumber(issue.activeAlerts)
  const symptom = inferRiskSymptom(issue, alertEvidence, detailEvidence)
  const objectLabel = issueObjectLabel(issue)
  const activityText = alertActivityLabel({ eventCount, instanceCount })
  const validAnalyses = safeList(issue.existingAlertAnalyses).filter((item) => safeText(item.validity) === 'valid').length
  const invalidAnalyses = safeList(issue.existingAlertAnalyses).filter((item) => safeText(item.validity) === 'invalid').length
  const stateText = activeCount > 0
    ? `当前仍有 ${activeCount} 条告警触发`
    : '当前已恢复，进入恢复观察'
  const scaleText = [
    eventCount > 0 && `${eventCount} 次触发`,
    instanceCount > 1 && `${instanceCount} 个告警对象`,
  ].filter(Boolean).join('、')
  const validityText = validAnalyses > 0
    ? `已有分析认为其中 ${validAnalyses} 类告警有效`
    : invalidAnalyses > 0
      ? `已有分析认为其中 ${invalidAnalyses} 类告警可能无效`
      : ''

  return [
    `${objectLabel}出现${symptom}告警${activityText ? `，${activityText}` : ''}${scaleText ? `，涉及 ${scaleText}` : ''}`,
    stateText,
    validityText,
  ].filter(Boolean).join('；') + '。'
}

const buildAlertAnalysisDisplay = (issue = {}, alertEvidence = {}, detailEvidence = {}) => {
  const eventCount = positiveNumber(issue.rawEventCount)
  const instanceCount = positiveNumber(issue.deduplicatedAlerts)
  const activeCount = positiveNumber(issue.activeAlerts)
  const analyses = safeList(issue.existingAlertAnalyses)
  const alertClassCount = new Set(analyses.map((analysis) => safeText(analysis.alertName)).filter(Boolean)).size
  const validAnalyses = analyses.filter((item) => safeText(item.validity) === 'valid').length
  const invalidAnalyses = analyses.filter((item) => safeText(item.validity) === 'invalid').length
  const symptom = inferRiskSymptom(issue, alertEvidence, detailEvidence)
  const objectLabel = issueObjectLabel(issue)
  const validityText = validAnalyses > 0
    ? '告警有效'
    : invalidAnalyses > 0
      ? '需要校验'
      : '等待分析'
  const serviceName = issueServiceName(issue)
  const trend = alertTrendSource(issue)

  if (trend) {
    const segments = normalizeAlertTrendSegments(trend)
    const window = alertTrendWindow(trend, segments)
    const episodes = buildAlertTrendEpisodes(segments, trend)
    const typeText = alertTypeListText(trend, segments, symptom)
    const activityText = alertTrendActivityText(trend, segments)
    const durationText = alertTrendDurationText(trend, segments)
    const statusText = activeCount > 0 ? '仍有未恢复告警' : '恢复观察'
    const summary = firstText(trend.summary, trend.narrative)
      || `${window.text}，${serviceName} 相关告警${activityText}${typeText ? `，主要围绕 ${typeText}` : ''}。${durationText}，${statusText}。`
    const problem = alertTrendProblemText({
      activeCount,
      activityText,
      durationText,
      typeText,
      trend,
    })

    return {
      mode: 'history',
      serviceName,
      title: '近 3 天告警态势',
      window,
      summary: truncateText(summary, 260),
      problem: truncateText(problem, 180),
      stateTone: activeCount > 0 ? 'active' : 'stable',
      stateText: activeCount > 0 ? '仍在触发' : '恢复观察',
      validityText,
      segments,
      episodes,
    }
  }

  const activityText = alertActivityLabel({ eventCount, instanceCount, alertClassCount })
  const scaleText = [
    eventCount > 0 && `触发 ${eventCount} 次`,
    alertClassCount > 1 && `涉及 ${alertClassCount} 类告警`,
    instanceCount > 1 && `覆盖 ${instanceCount} 个告警对象`,
  ].filter(Boolean).join('，')
  const summary = `${serviceName} 在本次报告窗口内${activityText}${scaleText ? `，${scaleText}` : ''}。由于缺少最近 3 天服务级告警历史，暂不能判断是否长期复发。`
  const problem = activeCount > 0
    ? '当前仍有告警触发，先作为当前窗口态势处理；服务级历史趋势需要后端补充后再判断。'
    : '当前已恢复，暂不应仅凭当前窗口判断该服务是否存在长期稳定性问题。'

  return {
    mode: 'current',
    serviceName,
    title: '当前窗口告警态势',
    window: null,
    summary,
    problem,
    stateTone: activeCount > 0 ? 'active' : 'stable',
    stateText: activeCount > 0 ? '仍在触发' : '恢复观察',
    validityText,
    segments: [],
    episodes: [],
    emptyText: `${objectLabel}侧 ${symptom}，缺少最近 3 天服务告警历史。`,
  }
}

const mergeAlertAnalysisSupplement = (analysis = {}, supplementText = '') => {
  const supplement = truncateText(safeText(supplementText), 140)
  if (!supplement) return analysis
  const existingText = [
    analysis.summary,
    analysis.problem,
    analysis.emptyText,
  ].map(safeText).join(' ')
  const summaryText = safeText(analysis.summary)
  if (existingText.includes(supplement) || (summaryText && supplement.includes(summaryText))) return analysis
  return {
    ...analysis,
    supplement,
  }
}

const buildIssueAnalysisText = (issue = {}, aiGenerated = false) => {
  const base = productIssueConclusion(issue, aiGenerated, Number.POSITIVE_INFINITY)
  const metricAnalysis = safeText(issue.metricAnalysis)
  const rootCause = safeText(issue.rootCause)
  const direction = metricAnalysis || rootCause
  if (!direction) return base
  const directionLabel = metricAnalysis ? '指标方向' : '报告判断'
  const directionText = `${directionLabel}：${direction}`
  if (!base) return directionText
  if (base.includes(direction) || base.includes(directionLabel)) return base
  return `${base} ${directionText}`
}

const buildKeyEvidenceItems = (evidence = [], issue = {}) => {
  const findEvidence = (...keywords) => safeList(evidence).find((item) => {
    const type = safeText(item.type)
    return keywords.some((keyword) => type.includes(keyword))
  })
  const alertEvidence = getEvidenceByType(evidence, '告警事件', { exclude: ['告警事件分析'] })
  const reportAlertAnalysisEvidence = findEvidence('方向报告-告警事件分析')
  const metricEvidenceCandidates = safeList(evidence).filter((item) => {
    const type = safeText(item.type)
    return !type.includes('请求北极星') && ['方向报告-指标趋势', '告警指标趋势', '指标趋势'].some((keyword) => type.includes(keyword))
  })
  const metricEvidence = metricEvidenceCandidates.find((item) => safeText(item.type).includes('告警指标趋势') && item.metric)
    || metricEvidenceCandidates.find((item) => item.metric)
    || metricEvidenceCandidates[0]
  const requestPolarisEvidence = findEvidence('请求北极星')
  const logEvidence = findEvidence('方向报告-日志分析', '日志分析')
  const detailEvidence = getEvidenceByType(evidence, '告警明细')
  const missingData = safeList(issue.missingData)
  const supplementalEvidence = safeList(evidence).filter(
    (item) => !['告警事件', '告警指标趋势', '指标趋势', '请求北极星', 'Trace链路样本', 'APM链路', '日志分析', '告警明细'].some((keyword) => (item.type || '').includes(keyword)),
  )
  const items = []
  const existingAlertAnalysisText = buildExistingAlertAnalysisText(issue)
  const alertBehaviorText = buildAlertBehaviorAnalysis(issue, alertEvidence, detailEvidence)
  const alertAnalysisText = [alertBehaviorText, existingAlertAnalysisText && `已有告警分析：${existingAlertAnalysisText}`]
    .filter(Boolean)
    .join(' ')
  const reportAlertText = reportAlertAnalysisEvidence?.result
    ? summarizeEvidenceItem(reportAlertAnalysisEvidence)
    : ''
  const rootCauseText = safeText(issue.rootCause)

  if (metricEvidence?.metric?.eventScoped && metricSampleCount(metricEvidence.metric) > 0) {
    const metric = metricEvidence.metric
    const matched = metric.seriesMatched !== undefined && metric.seriesReturned !== undefined
      ? `命中 ${metric.seriesMatched} 条事件指标`
      : '已返回指标趋势'
    items.push({
      id: 'metric-trend',
      label: '指标变化',
      text: metricEvidenceDisplayText(metricEvidence, issue) || `${matched}，窗口变化 ${formatPercent(metric.changePercent)}。`,
      tone: metric.eventScoped ? 'metric' : 'warning',
      metric,
      metricName: metricBackendDisplayName(metric),
    })
  } else if (safeText(issue.metricAnalysis)) {
    items.push({
      id: 'metric-trend',
      label: '指标变化',
      text: ensureMetricNameInText(metricEvidence?.metric, issue.metricAnalysis, metricEvidence?.result),
      tone: 'metric',
    })
  } else if (metricEvidence?.result) {
    const metricSummary = summarizeMetricEvidenceText(metricEvidence.result)
    const unavailableText = metricEvidenceUnavailableText(metricEvidence.result)
    if (metricSummary || unavailableText) {
      items.push({
        id: 'metric-trend',
        label: '指标变化',
        text: ensureMetricNameInText(metricEvidence?.metric, metricSummary || unavailableText, metricEvidence.result),
        tone: metricSummary ? 'metric' : 'warning',
      })
    }
  }

  if (requestPolarisEvidence?.result) {
    const requestPolarisText = summarizeRequestPolarisEvidenceText(requestPolarisEvidence)
    if (requestPolarisText) {
      items.push({
        id: 'request-polaris',
        label: '北极星耗时分解',
        text: requestPolarisText,
        tone: 'metric',
        metric: requestPolarisEvidence.metric,
        metricName: metricBackendDisplayName(requestPolarisEvidence.metric),
      })
    }
  }

  if (logEvidence?.result) {
    items.push({
      id: 'log-analysis',
      label: '日志分析',
      text: summarizeEvidenceItem(logEvidence),
      tone: 'log',
      rawLogs: extractRawLogLines(logEvidence),
    })
  }

  if (rootCauseText) {
    items.push({
      id: 'root-cause',
      label: '根因判断',
      text: rootCauseText,
      tone: 'rootcause',
    })
  }

  if (alertAnalysisText || reportAlertText || alertEvidence?.result || detailEvidence?.result) {
    const alertAnalysis = mergeAlertAnalysisSupplement(
      buildAlertAnalysisDisplay(issue, alertEvidence, detailEvidence),
      reportAlertText,
    )
    const fallbackAlertText = firstText(
      alertAnalysis.summary,
      reportAlertText,
      alertAnalysisText,
      stripTechnicalEvidenceText(detailEvidence?.result || ''),
    )
    items.push({
      id: 'alert-state',
      label: '告警态势',
      text: truncateText(fallbackAlertText, 220),
      tone: 'analysis',
      alertAnalysis,
    })
  }

  if (!alertEvidence?.result && !reportAlertText && !alertAnalysisText && detailEvidence?.result) {
    items.push({
      id: 'alert-detail',
      label: '告警明细',
      text: stripTechnicalEvidenceText(detailEvidence.result),
      tone: 'alert',
    })
  }

  supplementalEvidence.forEach((item) => {
    if (shouldHideDefaultEvidence(item)) return
    const text = summarizeEvidenceItem(item)
    if (!text) return
    const role = getEvidenceRole(item.type)
    items.push({
      id: `supplemental-${safeText(item.type) || 'evidence'}`,
      label: role.stage,
      text,
      tone: role.kind === 'boundary' ? 'warning' : role.kind,
    })
  })

  if (items.length === 0 && missingData.length > 0) {
    items.push({
      id: 'data-boundary',
      label: '数据边界',
      text: `${missingData.join('、')} 暂不可用，相关趋势或运行状态未参与当前结论。`,
      tone: 'warning',
    })
  }

  const seen = new Set()
  return items.filter((item) => {
    const key = `${item.id || item.label}:${item.text}`
    if (!item.text || seen.has(key)) return false
    seen.add(key)
    return true
  }).slice(0, 8)
}

const formatTrendTimelineTime = (value) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return [
    `${date.getMonth() + 1}/${date.getDate()}`,
    `${`${date.getHours()}`.padStart(2, '0')}:${`${date.getMinutes()}`.padStart(2, '0')}`,
  ].join(' ')
}

const startOfLocalDay = (value) => {
  const date = new Date(value)
  date.setHours(0, 0, 0, 0)
  return date.getTime()
}

const formatTrendHeatmapDate = (value) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const alertTrendHeatmapLevel = (cell = {}) => {
  if (cell.minutes <= 0) return 0
  if (cell.minutes >= 45) return 4
  if (cell.minutes >= 30) return 3
  if (cell.minutes >= 10) return 2
  return 1
}

const mergeAlertTrendIntervals = (intervals = []) => {
  const sorted = safeList(intervals)
    .filter((interval) => interval.end > interval.start)
    .sort((left, right) => left.start - right.start)
  const merged = []
  sorted.forEach((interval) => {
    const last = merged[merged.length - 1]
    if (!last || interval.start > last.end) {
      merged.push({ ...interval })
      return
    }
    last.end = Math.max(last.end, interval.end)
  })
  return merged
}

const buildAlertTrendHeatmap = (segments = [], window = {}) => {
  const end = window.end || Date.now()
  const start = startOfLocalDay(end) - 2 * ALERT_TREND_DAY_MS
  return Array.from({ length: 3 }, (_, rowIndex) => {
    const rowStart = start + rowIndex * ALERT_TREND_DAY_MS
    const rowEnd = rowStart + ALERT_TREND_DAY_MS
    const cells = Array.from({ length: 24 }, (_, hour) => {
      const cellStart = rowStart + hour * ALERT_TREND_HOUR_MS
      const cellEnd = cellStart + ALERT_TREND_HOUR_MS
      const cell = {
        key: `${rowIndex}-${hour}`,
        start: cellStart,
        end: cellEnd,
        events: 0,
        minutes: 0,
        intervals: [],
        titles: [],
        types: [],
      }
      safeList(segments).forEach((segment) => {
        const overlapStart = Math.max(segment.start, cellStart)
        const overlapEnd = Math.min(segment.end, cellEnd)
        const overlap = Math.max(0, overlapEnd - overlapStart)
        if (overlap <= 0) return
        cell.events += segment.count || 1
        cell.intervals.push({ start: overlapStart, end: overlapEnd })
        if (segment.title) cell.titles.push(segment.title)
        if (segment.type) cell.types.push(segment.type)
      })
      const activeMs = mergeAlertTrendIntervals(cell.intervals)
        .reduce((sum, interval) => sum + interval.end - interval.start, 0)
      cell.minutes = Math.min(60, Math.ceil(activeMs / 60000))
      const level = alertTrendHeatmapLevel(cell)
      const types = Array.from(new Set(cell.types)).slice(0, 3)
      const titles = Array.from(new Set(cell.titles)).slice(0, 2)
      return {
        ...cell,
        intervals: undefined,
        level,
        title: level > 0
          ? [
            `${formatTrendTimelineTime(cellStart)} - ${formatTrendTimelineTime(cellEnd)} 出现告警`,
            cell.events > 1 ? `告警信号 ${cell.events} 条` : '告警信号 1 条',
            cell.minutes > 0 ? `持续约 ${formatDurationBrief(cell.minutes * 60)}` : '',
            types.length > 0 ? `类型：${types.join('、')}` : '',
            titles.length > 0 ? titles.join('；') : '',
          ].filter(Boolean).join('；')
          : `${formatTrendTimelineTime(cellStart)} - ${formatTrendTimelineTime(cellEnd)} 未发现相关告警`,
      }
    })
    return {
      key: `row-${rowIndex}`,
      start: rowStart,
      end: rowEnd,
      label: formatTrendHeatmapDate(rowStart),
      rangeText: `${formatTrendTimelineTime(rowStart)} - ${formatTrendTimelineTime(rowEnd)}`,
      cells,
    }
  })
}

const AlertTrendHeatmapLegend = () => (
  <div className="health-alert-trend-legend">
    <span>颜色越深，告警持续时间越长</span>
    <em>Less</em>
    <i className="health-alert-trend-guide-empty" />
    <i className="health-alert-trend-guide-hot health-alert-trend-guide-hot-1" />
    <i className="health-alert-trend-guide-hot health-alert-trend-guide-hot-2" />
    <i className="health-alert-trend-guide-hot health-alert-trend-guide-hot-3" />
    <i className="health-alert-trend-guide-hot health-alert-trend-guide-hot-4" />
    <em>More</em>
  </div>
)

const alertTrendHotCellCount = (analysis = {}) => {
  const segments = safeList(analysis.segments)
  if (segments.length === 0) return 0
  const window = analysis.window || alertTrendWindow({}, segments)
  return buildAlertTrendHeatmap(segments, window)
    .reduce((sum, row) => sum + row.cells.filter((cell) => cell.level > 0).length, 0)
}

const alertTrendJudgementText = (analysis = {}) => {
  const problem = safeText(analysis.problem)
  const hotCellCount = alertTrendHotCellCount(analysis)
  if (hotCellCount <= 0) return problem
  return `最近 3 天共有 ${hotCellCount} 个小时出现告警活动。${problem}`
}

const AlertTrendTimeline = ({ analysis }) => {
  const segments = safeList(analysis.segments)
  if (segments.length === 0) {
    return (
      <div className="health-alert-trend-empty">
        {analysis.emptyText || '最近 3 天服务告警历史暂未返回。'}
      </div>
    )
  }
  const window = analysis.window || alertTrendWindow({}, segments)
  const rows = buildAlertTrendHeatmap(segments, window)
  return (
    <div className="health-alert-trend-timeline">
      <div className="health-alert-trend-scale">
        <span>最近3天告警情况</span>
        <AlertTrendHeatmapLegend />
      </div>
      <div className="health-alert-trend-heatmap">
        <div className="health-alert-trend-heatmap-axis" aria-hidden="true">
          <span />
          {[0, 6, 12, 18, 24].map((hour) => (
            <b
              key={hour}
              style={{ gridColumn: hour === 24 ? '25 / 26' : `${hour + 2} / span 2` }}
            >
              {hour}
              h
            </b>
          ))}
        </div>
        {rows.map((row) => (
          <div className="health-alert-trend-heatmap-row" key={row.key}>
            <div className="health-alert-trend-heatmap-day" title={row.rangeText}>
              <span>{row.label}</span>
            </div>
            <div className="health-alert-trend-heatmap-cells">
              {row.cells.map((cell) => (
                <span
                  aria-label={cell.title}
                  className={`health-alert-trend-heatmap-cell health-alert-trend-heatmap-cell-${cell.level}`}
                  key={cell.key}
                  title={cell.title}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

AlertTrendTimeline.propTypes = {
  analysis: PropTypes.object,
}

const ServiceAlertTrendEvidence = ({ analysis }) => (
  <div className="health-alert-trend">
    <div className="health-alert-trend-head">
      <div>
        <span>{analysis.title}</span>
        <b>{analysis.serviceName}</b>
      </div>
      <div className="health-alert-analysis-tags">
        <em className={`health-alert-analysis-state health-alert-analysis-state-${analysis.stateTone}`}>
          {analysis.stateText}
        </em>
        <em className="health-alert-analysis-validity">
          {analysis.validityText}
        </em>
      </div>
    </div>
    <ExpandableText
      className="health-alert-trend-summary"
      text={analysis.summary}
      threshold={130}
    />
    {analysis.supplement && (
      <ExpandableText
        className="health-alert-trend-supplement"
        text={`报告补充：${analysis.supplement}`}
        threshold={100}
      />
    )}
    <AlertTrendTimeline analysis={analysis} />
    <div className="health-alert-trend-problem">
      <span>趋势判断</span>
      <ExpandableText
        className="health-alert-trend-problem-text"
        text={alertTrendJudgementText(analysis)}
        threshold={120}
      />
    </div>
  </div>
)

ServiceAlertTrendEvidence.propTypes = {
  analysis: PropTypes.object,
}

export default function ServiceHealthyPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const timeRange = useSelector((state) => state.timeRange)
  const pageLoadedAtRef = useRef(Date.now() * 1000)
  const initialTimeRangeSignatureRef = useRef(null)
  const [keyword, setKeyword] = useState('')
  const [priorityFilter, setPriorityFilter] = useState('all')
  const [serviceFilter, setServiceFilter] = useState('all')
  const [healthReportData, setHealthReportData] = useState(null)
  const [reportError, setReportError] = useState('')
  const [isReportRefreshing, setIsReportRefreshing] = useState(false)
  const [reportGeneratingStartedAt, setReportGeneratingStartedAt] = useState(0)
  const [reportGeneratingNow, setReportGeneratingNow] = useState(Date.now())
  const [historyOpen, setHistoryOpen] = useState(false)
  const [historyRecords, setHistoryRecords] = useState([])
  const [historyTotal, setHistoryTotal] = useState(0)
  const [historyLoading, setHistoryLoading] = useState(false)
  const [historyError, setHistoryError] = useState('')
  const [selectedRecord, setSelectedRecord] = useState(null)
  const [recordLoadingId, setRecordLoadingId] = useState('')
  const selectedRecordRef = useRef(null)
  const lowerKeyword = keyword.trim().toLowerCase()
  const currentTimeRangeSignature = makeTimeRangeSignature(timeRange)
  if (initialTimeRangeSignatureRef.current === null) {
    initialTimeRangeSignatureRef.current = currentTimeRangeSignature
  }
  const timeRangeChangedAfterMount = currentTimeRangeSignature !== initialTimeRangeSignatureRef.current
  const timeRangeRefreshedAfterMount = positiveNumber(timeRange?.refreshTimestamp) >= pageLoadedAtRef.current
  const requestParams = useMemo(
    () => buildReportRequestParams(location.search, timeRange, {
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

  useEffect(() => {
    selectedRecordRef.current = selectedRecord
  }, [selectedRecord])

  useEffect(() => {
    if (!isReportRefreshing || healthReportData) return undefined
    const timer = window.setInterval(() => {
      setReportGeneratingNow(Date.now())
    }, 1000)
    return () => window.clearInterval(timer)
  }, [isReportRefreshing, healthReportData])

  const loadHistoryRecords = () => {
    setHistoryLoading(true)
    setHistoryError('')
    getSystemHealthReportRecordsApi({
      groupId: requestParams.groupId,
      page: 1,
      pageSize: 30,
    })
      .then((data) => {
        setHistoryRecords(safeList(data?.records))
        setHistoryTotal(positiveNumber(data?.total))
      })
      .catch(() => {
        setHistoryError('历史报告读取失败，请稍后重试。')
      })
      .finally(() => {
        setHistoryLoading(false)
      })
  }

  const openHistory = () => {
    setHistoryOpen(true)
    loadHistoryRecords()
  }

  const loadRecordDetail = (record) => {
    const recordId = safeText(record?.recordId)
    if (!recordId) return
    setRecordLoadingId(recordId)
    setHistoryError('')
    getSystemHealthReportRecordApi(recordId)
      .then((data) => {
        if (data?.data?.report && Array.isArray(data?.data?.issues)) {
          setHealthReportData(data.data)
          setSelectedRecord(data.record || record)
          setReportError('')
          setHistoryOpen(false)
        }
      })
      .catch(() => {
        setHistoryError('历史报告详情读取失败，请稍后重试。')
      })
      .finally(() => {
        setRecordLoadingId('')
      })
  }

  useEffect(() => {
    let ignore = false
    selectedRecordRef.current = null
    setSelectedRecord(null)
    const cacheKey = makeReportCacheKey(requestParams)
    const cached = readPersistedReport(cacheKey)
    if (cached) {
      setHealthReportData(cached.data)
      setReportError('')
    } else {
      setHealthReportData(null)
    }
    setReportGeneratingStartedAt(Date.now())
    setReportGeneratingNow(Date.now())
    setIsReportRefreshing(true)
    getSystemHealthReportApi(requestParams)
      .then((data) => {
        if (!ignore && data?.report && Array.isArray(data?.issues)) {
          if (selectedRecordRef.current) return
          setHealthReportData(data)
          setSelectedRecord(null)
          setReportError('')
          persistReport(cacheKey, data)
        }
      })
      .catch(() => {
        if (!ignore) {
          if (cached) {
            setReportError('实时数据刷新失败，正在展示最近一次系统健康报告。')
          } else {
            setHealthReportData(null)
            setReportError('暂时无法读取系统健康数据，请稍后刷新或确认数据服务状态。')
          }
        }
      })
      .finally(() => {
        if (!ignore) {
          setIsReportRefreshing(false)
        }
      })
    return () => {
      ignore = true
    }
  }, [requestParams])

  const issues = healthReportData?.issues ?? []
  const priorityFilterOptions = useMemo(() => {
    const counts = new Map()
    issues.forEach((issue) => {
      const priority = normalizePriority(issue.priority)
      counts.set(priority, (counts.get(priority) || 0) + 1)
    })
    return [
      { value: 'P0', label: `P0 故障${counts.has('P0') ? ` (${counts.get('P0')})` : ''}` },
      { value: 'P1', label: `P1 严重${counts.has('P1') ? ` (${counts.get('P1')})` : ''}` },
      { value: 'P2', label: `P2 警告${counts.has('P2') ? ` (${counts.get('P2')})` : ''}` },
      { value: 'P3', label: `P3 观察${counts.has('P3') ? ` (${counts.get('P3')})` : ''}` },
      { value: 'P4', label: `P4 提示${counts.has('P4') ? ` (${counts.get('P4')})` : ''}` },
    ].filter((item) => counts.has(item.value))
  }, [issues])
  const serviceFilterOptions = useMemo(() => {
    const options = new Map()
    issues.forEach((issue) => {
      const target = issueFocusTarget(issue)
      const key = focusTargetKey(target)
      if (!key || options.has(key)) return
      options.set(key, {
        value: key,
        label: target.label === '服务' ? target.name : `${target.name} (${target.label})`,
      })
    })
    return Array.from(options.values()).sort((left, right) => left.label.localeCompare(right.label))
  }, [issues])
  const hasIssueFilters = Boolean(lowerKeyword || priorityFilter !== 'all' || serviceFilter !== 'all')

  const filteredIssues = useMemo(() => {
    return issues.filter(
      (item) => {
        const matchesKeyword = !lowerKeyword || [
          item.title,
          item.scope,
          item.systemProblem,
          item.rootCause,
          item.aiConclusion,
          item.impact,
          item.mergeBasis,
          safeList(item.dataSources).join(' '),
          safeList(item.actions).join(' '),
          safeList(item.existingAlertAnalyses).map((analysis) => [
            analysis.alertName,
            analysis.alertDirection,
            analysis.checkOutput,
            analysis.analyzeError,
            analysis.validity,
          ].join(' ')).join(' '),
          safeList(item.affectedObjects).map((object) => object.name).join(' '),
        ].join(' ').toLowerCase().includes(lowerKeyword)
        const matchesPriority = priorityFilter === 'all' || normalizePriority(item.priority) === priorityFilter
        const matchesService = serviceFilter === 'all' || focusTargetKey(issueFocusTarget(item)) === serviceFilter
        return matchesKeyword && matchesPriority && matchesService
      },
    )
  }, [lowerKeyword, priorityFilter, serviceFilter, issues])

  const report = healthReportData?.report ?? {
    ...EMPTY_REPORT,
    summary: reportError || EMPTY_REPORT.summary,
  }
  const reportStatusMeta = LEVEL_META[report.status] ?? LEVEL_META.warning
  const reportScore = Number.isFinite(Number(report.score)) ? Number(report.score) : 0
  const aiAnalysis = report.aiAnalysis || {}
  const isAISummaryReady = ['succeeded', 'partial'].includes(aiAnalysis.status)
  const aiMeta = aiAnalysisMeta(aiAnalysis)

  const handleAlertTriage = (objectType, objectName) => {
    const currentParams = new URLSearchParams(location.search)
    const groupId = currentParams.get('groupId') || '1'
    const params = new URLSearchParams({
      source: 'ai-system-report',
      objectType,
      objectName,
      groupId,
    })
    navigate(`${jumpTargets.alerts}?${params.toString()}`)
  }

  const impactSummary = {
    p0Count: filteredIssues.filter((issue) => issue.priority === 'P0').length,
    p1Count: filteredIssues.filter((issue) => issue.priority === 'P1').length,
    p2Count: filteredIssues.filter((issue) => issue.priority === 'P2').length,
    p3Count: filteredIssues.filter((issue) => issue.priority === 'P3').length,
    p4Count: filteredIssues.filter((issue) => issue.priority === 'P4').length,
  }
  const recoveringIssueCount = filteredIssues.filter((issue) => issue.status === 'recovering' && positiveNumber(issue.activeAlerts) === 0).length
  const topFocusTargets = buildTopFocusTargets(filteredIssues)
  const priorityDistribution = [
    { key: 'P0', label: 'P0', name: '故障', count: impactSummary.p0Count },
    { key: 'P1', label: 'P1', name: '严重', count: impactSummary.p1Count },
    { key: 'P2', label: 'P2', name: '警告', count: impactSummary.p2Count },
    { key: 'P3', label: 'P3', name: '观察', count: impactSummary.p3Count },
    { key: 'P4', label: 'P4', name: '提示', count: impactSummary.p4Count },
  ]
  const priorityTotal = Math.max(1, priorityDistribution.reduce((sum, item) => sum + item.count, 0))
  const highestPriority = priorityDistribution.find((item) => item.count > 0)
  const urgentIssues = filteredIssues.filter((issue) => ['P0', 'P1'].includes(normalizePriority(issue.priority)) && positiveNumber(issue.activeAlerts) > 0)
  const urgentAlertCount = urgentIssues.reduce((sum, issue) => sum + positiveNumber(issue.activeAlerts), 0)
  const ruleDiagnosticIssueCount = filteredIssues.filter(isRuleDiagnosticIssue).length
  const watchIssueCount = Math.max(0, filteredIssues.length - urgentIssues.length - ruleDiagnosticIssueCount)
  const triageLeadText = highestPriority
    ? `${highestPriority.label} ${highestPriority.name}为当前最高风险`
    : '暂无待处理风险'
  const commandStats = [
    { label: hasIssueFilters ? '筛选结果' : '风险项', value: filteredIssues.length, show: true },
    { label: '触发中告警', value: report.activeAlertCount || 0, show: true },
    { label: '恢复观察', value: recoveringIssueCount, show: true },
  ].filter((item) => item.show)
  const reportSummaryText = reportExecutiveSummary(report, issues, productReportSummary(report.summary || '', isAISummaryReady))
  const generationElapsedSeconds = reportGeneratingStartedAt > 0
    ? Math.max(0, Math.floor((reportGeneratingNow - reportGeneratingStartedAt) / 1000))
    : 0
  const historyDrawer = (
    <Drawer
      title="历史报告"
      open={historyOpen}
      onClose={() => setHistoryOpen(false)}
      width={420}
      destroyOnClose={false}
    >
      <div className="health-history-toolbar">
        <span>已保存 {historyTotal || historyRecords.length} 条记录</span>
        <Button size="small" onClick={loadHistoryRecords} loading={historyLoading}>刷新</Button>
      </div>
      {historyError && <div className="health-history-error">{historyError}</div>}
      <Spin spinning={historyLoading}>
        <List
          dataSource={historyRecords}
          locale={{ emptyText: '暂无历史报告记录' }}
          renderItem={(record) => (
            <List.Item
              className={[
                'health-history-item',
                selectedRecord?.recordId === record.recordId ? 'health-history-item-active' : '',
              ].filter(Boolean).join(' ')}
              onClick={() => loadRecordDetail(record)}
            >
              <div className="health-history-item-main">
                <div className="health-history-item-title">
                  <StatusTag status={record.status} />
                  <strong>{recordWindowText(record)}</strong>
                </div>
                <span>{truncateText(record.summary || '系统健康报告快照', 120)}</span>
                <small>
                  生成 {formatRecordTime(record.createdAt)}
                  {' · '}
                  风险项 {record.issueCount}
                  {' · '}
                  健康分 {record.score}
                </small>
              </div>
              <Button
                size="small"
                type={selectedRecord?.recordId === record.recordId ? 'primary' : 'default'}
                loading={recordLoadingId === record.recordId}
              >
                查看
              </Button>
            </List.Item>
          )}
        />
      </Spin>
    </Drawer>
  )

  if (!healthReportData) {
    const waitingDescription = reportError || '正在汇总当前时间窗口内的真实告警、指标趋势、请求北极星、日志分析和已生成的告警分析报告。'
    return (
      <>
        <div className="system-health-page">
          <div className="health-report-header">
            <div className="health-report-main">
              <div className="health-title-row">
                <div className="health-title">系统健康报告</div>
                {isReportRefreshing && !selectedRecord && <Tag color="processing">刷新中</Tag>}
                {selectedRecord && <Tag color="geekblue">历史快照</Tag>}
              </div>
              {selectedRecord && (
                <div className="health-report-meta">
                  <span>记录时间：{formatRecordTime(selectedRecord.createdAt)}</span>
                  <span>快照窗口：{recordWindowText(selectedRecord)}</span>
                </div>
              )}
            </div>
            <Space wrap className="health-report-actions">
              <Button onClick={openHistory}>历史记录</Button>
            </Space>
          </div>
          <div className="health-impact-empty">
            <div className="health-report-generating">
              <Spin spinning={isReportRefreshing} />
              <div className="health-report-generating-body">
                <strong>{reportError ? '暂时无法生成系统健康报告' : '正在生成系统健康报告'}</strong>
                <span>{waitingDescription}</span>
                <div className="health-report-generating-blocks">
                  {reportWaitingBlocks.map((block) => (
                    <div className="health-report-generating-block" key={block.key}>
                      <div className="health-report-generating-block-head">
                        <b>{block.title}</b>
                      </div>
                      <span>{block.description}</span>
                    </div>
                  ))}
                </div>
                <small>
                  已等待 {formatElapsedSeconds(generationElapsedSeconds)}
                  {!reportError ? '，分析完成后会自动刷新页面。' : '。'}
                  {' '}
                  可先查看历史记录中的已生成快照。
                </small>
                <Space wrap>
                  <Button onClick={openHistory}>查看历史记录</Button>
                </Space>
              </div>
            </div>
          </div>
        </div>
        {historyDrawer}
      </>
    )
  }

  const impactCommandOverview = (
    <div className="health-impact-command">
      <div className="health-impact-command-strip">
        <div className="health-impact-command-status">
          <div className="health-impact-command-status-main">
            <Progress
              type="circle"
              percent={reportScore}
              strokeColor={reportStatusMeta.color}
              width={74}
              format={() => reportScore}
            />
            <div>
              <span>健康度</span>
              <b>{reportStatusMeta.text}</b>
              <small>{report.window || '当前窗口'}</small>
            </div>
          </div>
          <div className="health-impact-command-stats">
            {commandStats.map((item) => (
              <span
                className={item.wide ? 'health-impact-command-stat-wide' : ''}
                key={item.label}
              >
                {item.label}
                <b>{item.value}</b>
              </span>
            ))}
          </div>
        </div>
        <div className="health-impact-command-readout">
          <div className="health-impact-command-lead">
            <span>当前判断</span>
            <strong>{reportSummaryText}</strong>
            {aiMeta && aiAnalysis.status !== 'not_configured' && (
              <small>{aiMeta.text}</small>
            )}
          </div>
          {topFocusTargets.length > 0 && (
            <div className="health-impact-command-focus">
              <div className="health-impact-command-panel-head">
                <strong>优先排查对象</strong>
                <span>按风险影响排序</span>
              </div>
              <div className="health-impact-focus-list">
                {topFocusTargets.map((target, index) => {
                  const metaText = [
                    target.label !== '服务' && target.label,
                    target.activeAlerts > 0 ? `触发中 ${target.activeAlerts}` : '恢复观察',
                    target.eventCount > 0 && `事件 ${target.eventCount}`,
                    target.issueCount > 1 && `${target.issueCount} 项风险`,
                  ].filter(Boolean).join(' · ')
                  return (
                    <div className="health-impact-focus-item" key={`${target.label}-${target.name}`}>
                      <span className="health-impact-focus-rank">{index + 1}</span>
                      <span className={`health-risk-priority health-risk-priority-${target.priority.toLowerCase()}`}>{target.priority}</span>
                      <div>
                        <b title={target.name}>{target.name}</b>
                        <small>{metaText}</small>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
        <div className="health-impact-command-priority">
          <div className="health-impact-command-panel-head">
            <strong>处置队列</strong>
            <span>{triageLeadText}</span>
          </div>
          <div className="health-impact-priority-stack" aria-label="风险构成">
            {priorityDistribution.map((item) => (
              item.count > 0 && (
                <span
                  className={`health-impact-priority-stack-item health-impact-priority-${item.key.toLowerCase()}`}
                  key={item.key}
                  style={{ width: `${Math.max(8, (item.count / priorityTotal) * 100)}%` }}
                  title={`${item.label} ${item.name}: ${item.count} 项`}
                />
              )
            ))}
          </div>
          <div className="health-impact-triage-list">
            <div className="health-impact-triage-item health-impact-triage-urgent">
              <span>立即处理</span>
              <strong>{urgentIssues.length} 项</strong>
              <small>{urgentAlertCount > 0 ? `P0/P1 未恢复，触发中 ${urgentAlertCount} 条` : '暂无未恢复高优先级风险'}</small>
            </div>
            <div className="health-impact-triage-item health-impact-triage-watch">
              <span>继续观察</span>
              <strong>{watchIssueCount} 项</strong>
              <small>关注恢复稳定性和复发迹象</small>
            </div>
            <div className="health-impact-triage-item health-impact-triage-rule">
              <span>规则校验</span>
              <strong>{ruleDiagnosticIssueCount} 项</strong>
              <small>避免无效告警进入故障处置</small>
            </div>
          </div>
        </div>
      </div>
    </div>
  )

  const impactPathVersion = (
    <div className="health-version-impact">
      {impactCommandOverview}
      <div className="health-impact-detail-title">
        <b>系统风险项</b>
        <div className="health-impact-detail-tools">
          <Input.Search
            allowClear
            className="health-impact-filter-search"
            placeholder="搜索问题 / 服务 / 接口"
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
          />
          <Select
            className="health-impact-filter-select"
            options={priorityFilterOptions}
            placeholder="风险级别"
            popupMatchSelectWidth={false}
            value={priorityFilter === 'all' ? undefined : priorityFilter}
            allowClear
            onChange={(value) => setPriorityFilter(value || 'all')}
          />
          <Select
            className="health-impact-filter-select health-impact-filter-service"
            options={serviceFilterOptions}
            placeholder="服务"
            popupMatchSelectWidth={false}
            showSearch
            optionFilterProp="label"
            value={serviceFilter === 'all' ? undefined : serviceFilter}
            allowClear
            onChange={(value) => setServiceFilter(value || 'all')}
          />
          <span className="health-impact-detail-count">{filteredIssues.length} 项</span>
        </div>
      </div>
      <div className="health-impact-grid">
        {filteredIssues.length === 0 && (
          <div className="health-impact-empty">
            {reportSummaryText ? (
              <div className="health-impact-empty-summary">
                <strong>系统风险项</strong>
                <span>{reportSummaryText}</span>
              </div>
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={reportError || '暂无需要关注的问题'}
              />
            )}
          </div>
        )}
        {filteredIssues.map((issue) => {
          const priority = normalizePriority(issue.priority)
          const status = normalizeStatus(issue.status, report.status)
          const title = firstText(issue.title, issue.systemProblem, issue.scope, '未命名问题')
          const scope = firstText(issue.scope, title)
          const displayTitle = title
          const metaItems = issueMetaItems(issue)
          const evidence = mergeEvidenceItems(issue.evidence)
          const actionPlan = safeList(issue.actionPlan)
          const actionSteps = actionPlan.length > 0
            ? actionPlan
            : safeList(issue.actions).map((action, index) => ({
              phase: `建议 ${index + 1}`,
              action,
              source: 'ai',
            }))
          const actionSourceText = actionPlanSourceText(actionSteps)
          const visibleAffectedObjects = visibleObjectTags(issue)
          const impactObjects = buildImpactObjects(displayTitle, scope, visibleAffectedObjects, issue.affectedObjects)
          const keyEvidence = buildKeyEvidenceItems(evidence, issue)
          const issueAnalysis = buildIssueAnalysisText(issue, isAISummaryReady)
          const shouldShowRoute = impactObjects.items.length > 0
          return (
            <div
              className={[
                'health-impact-card',
                `health-issue-${priority.toLowerCase()}`,
                actionSteps.length === 0 ? 'health-impact-card-no-actions' : '',
                !shouldShowRoute ? 'health-impact-card-no-route' : '',
              ].filter(Boolean).join(' ')}
              key={issue.key || `${priority}-${scope}`}
            >
              <div className="health-impact-card-top">
                <span className={`health-risk-priority health-risk-priority-${priority.toLowerCase()}`}>{priority}</span>
                <b>{displayTitle}</b>
                <div className="health-impact-card-meta">
                  <StatusTag
                    className="health-impact-status-tag"
                    hiddenStatuses={['recovering']}
                    status={status}
                  />
                  {metaItems.map((item) => <span key={`${issue.key}-${item}`}>{item}</span>)}
                </div>
                {issueAnalysis && (
                  <div className="health-impact-analysis">
                    <strong>分析结论</strong>
                    <ExpandableText
                      className="health-impact-analysis-text"
                      expandLabel="展开完整结论"
                      text={issueAnalysis}
                      threshold={150}
                    />
                  </div>
                )}
                {shouldShowRoute && (
                  <div className="health-impact-route">
                    <div className="health-impact-route-title">
                      <strong>影响对象</strong>
                      {impactObjects.overflowCount > 0 && <span>另有 {impactObjects.overflowCount} 项</span>}
                    </div>
                    <div className="health-impact-object-row">
                      {impactObjects.items.map((object) => (
                        <div
                          className="health-impact-object-item"
                          key={`${issue.key}-${object.label}-${object.name}`}
                          title={`${object.label}: ${object.name}`}
                        >
                          <span className="health-impact-object-label">{object.label}</span>
                          <span className="health-impact-object-value">{object.name}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
              <div className="health-impact-evidence">
                {keyEvidence.length > 0 && (
                  <>
                    <div className="health-impact-evidence-head">
                      <b>关键依据</b>
                    </div>
                    <div className="health-impact-key-evidence">
                      {keyEvidence.map((item) => (
                        <div className={`health-impact-key-evidence-item health-impact-key-evidence-${item.tone}`} key={`${issue.key}-${item.id || item.label}`}>
                          <strong>{item.label}</strong>
                          {item.metricName && (
                            <div className="health-metric-name-row" title={item.metricName}>
                              <span>指标名</span>
                              <b>{truncateText(item.metricName, 96)}</b>
                            </div>
                          )}
                          {item.alertAnalysis ? (
                            <ServiceAlertTrendEvidence analysis={item.alertAnalysis} />
                          ) : (
                            <ExpandableText
                              className="health-impact-key-evidence-text"
                              text={item.text}
                              threshold={120}
                            />
                          )}
                          {item.rawLogs?.length > 0 && (
                            <details className="health-raw-log-panel">
                              <summary>
                                <span>原始异常日志</span>
                                <small>{item.rawLogs.length} 行</small>
                              </summary>
                              <div className="health-raw-log-scroll">
                                <ol>
                                  {item.rawLogs.map((line, index) => (
                                    <li key={`${issue.key}-${item.id || item.label}-log-${index}`}>
                                      <code>{line}</code>
                                    </li>
                                  ))}
                                </ol>
                              </div>
                            </details>
                          )}
                          {item.metric && hasMetricChartData(item.metric) && (
                            <MetricTrendChart metric={item.metric} compact />
                          )}
                        </div>
                      ))}
                    </div>
                  </>
                )}
              </div>
              {actionSteps.length > 0 && (
                <div className="health-impact-actions">
                  <div className="health-impact-actions-head">
                    <div>
                      <div className="health-impact-actions-title">处置建议</div>
                      <span>{actionSourceText}</span>
                    </div>
                  </div>
                  {actionSteps.map((step, index) => (
                    <div className="health-impact-action-step" key={`${issue.key}-legacy-${step.phase}-${index}`}>
                      <b>{index + 1}</b>
                      <div>
                        <strong>{step.phase}</strong>
                        {displayActionTarget(step.target, issue) && <small>{displayActionTarget(step.target, issue)}</small>}
                        <span>{step.action}</span>
                        {step.validation && <em>验证：{step.validation}</em>}
                        {step.risk && <em>风险：{step.risk}</em>}
                      </div>
                    </div>
                  ))}
                  <div className="health-impact-execute">
                    <Button size="small" type={priority === 'P0' ? 'primary' : 'default'} onClick={() => handleAlertTriage('risk', scope)}>
                      查看关联告警
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )

  return (
    <>
      <div className="system-health-page">
        <div className="health-report-header">
          <div className="health-report-main">
            <div className="health-title-row">
              <div className="health-title">系统健康报告</div>
              <StatusTag status={report.status} />
              {isReportRefreshing && !selectedRecord && <Tag color="processing">刷新中</Tag>}
              {selectedRecord && <Tag color="geekblue">历史快照</Tag>}
              {reportError && healthReportData && <Tag color="warning">已显示缓存</Tag>}
              {aiMeta && aiAnalysis.status !== 'not_configured' && <Tag color={aiMeta.color}>{aiMeta.text}</Tag>}
            </div>
            <div className="health-report-meta">
              <span>时间窗口：{selectedRecord ? recordWindowText(selectedRecord) : (report.window || '当前窗口')}</span>
              {report.generatedAt && <span>分析时间：{report.generatedAt}</span>}
              {selectedRecord && <span>记录时间：{formatRecordTime(selectedRecord.createdAt)}</span>}
            </div>
          </div>
          <Space wrap className="health-report-actions">
            <Button onClick={openHistory}>历史记录</Button>
          </Space>
        </div>

        {impactPathVersion}
      </div>
      {historyDrawer}
    </>
  )
}
