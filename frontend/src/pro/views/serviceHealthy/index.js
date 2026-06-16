import React, { useEffect, useMemo, useState } from 'react'
import PropTypes from 'prop-types'
import {
  Button,
  Empty,
  Input,
  Progress,
  Space,
  Tag,
} from 'antd'
import ReactECharts from 'echarts-for-react'
import { useLocation, useNavigate } from 'react-router-dom'
import { ChartColorList } from 'src/constants'
import { getSystemHealthReportApi } from 'src/pro/api/serviceHealthy'
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

const StatusTag = ({ status }) => {
  const meta = LEVEL_META[status] ?? LEVEL_META.warning
  return <Tag color={meta.tag}>{meta.text}</Tag>
}

StatusTag.propTypes = {
  status: PropTypes.string,
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
      return { text: 'AI增强', color: 'processing' }
    case 'failed':
      return { text: 'AI失败，展示结构化分析', color: 'warning' }
    case 'skipped':
      return { text: '无问题，跳过AI', color: 'default' }
    case 'not_configured':
      return { text: '结构化分析', color: 'default' }
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
  if (evidenceType.includes('告警事件')) {
    return {
      stage: '告警汇总',
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
  if (evidenceType.includes('告警指标趋势')) {
    return {
      stage: '指标趋势',
      effect: '根据告警标签定位对应指标，并展示当前窗口变化。',
      kind: 'alert',
    }
  }
  if (evidenceType.includes('Trace链路样本')) {
    return {
      stage: 'Trace样本',
      effect: '结合当前窗口链路样本补充影响范围。',
      kind: 'trace',
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
    stage: '证据补充',
    effect: '来自当前已接入告警事件的补充信息。',
    kind: 'default',
  }
}

const getEvidenceByType = (items, keyword) => safeList(items).find((item) => (item.type || '').includes(keyword))

const objectTypeText = {
  endpoint: '入口',
  target: '影响目标',
  service: '服务',
  node: '节点',
  database: '数据库',
  namespace: '命名空间',
  pod: 'Pod',
  group: '类型',
}

const summarizeAffectedObjectScope = (issues = []) => {
  const byType = new Map()
  const all = new Set()
  safeList(issues).forEach((issue) => {
    const affectedObjects = safeList(issue.affectedObjects)
    if (affectedObjects.length > 0) {
      affectedObjects.forEach((item) => {
        const type = safeText(item.type) || 'unknown'
        const name = safeText(item.name)
        if (!name) return
        const key = `${type}:${name}`
        all.add(key)
        if (!byType.has(type)) byType.set(type, new Set())
        byType.get(type).add(name)
      })
      return
    }
    safeText(issue.scope).split('/').map((item) => item.trim()).filter(Boolean).forEach((name) => {
      all.add(`scope:${name}`)
      if (!byType.has('scope')) byType.set('scope', new Set())
      byType.get('scope').add(name)
    })
  })
  const priorityTypes = ['endpoint', 'service', 'pod', 'node', 'database', 'namespace', 'target']
  const parts = priorityTypes
    .map((type) => {
      const count = byType.get(type)?.size || 0
      return count > 0 ? `${objectTypeText[type] || type} ${count}` : ''
    })
    .filter(Boolean)
    .slice(0, 3)
  if (parts.length > 0) {
    return {
      count: all.size,
      label: '影响范围',
      value: parts.join(' / '),
    }
  }
  return {
    count: all.size,
    label: '影响范围',
    value: all.size,
  }
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
  return number < 1000000000000 ? number * 1000 : number
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
  safeList(metric.series)
    .map((series, index) => {
      const data = safeList(series.points)
        .map((point) => {
          const timestamp = normalizeMetricTimestamp(point.timestamp)
          const value = Number(point.value)
          if (timestamp === null || !Number.isFinite(value)) return null
          return [timestamp, value]
        })
        .filter(Boolean)
        .sort((left, right) => left[0] - right[0])
      return {
        name: metricSeriesLabel(series.labels, index),
        data,
      }
    })
    .filter((series) => series.data.length > 0)

const hasMetricChartData = (metric) => metricChartSeries(metric).length > 0

const MetricTrendChart = ({ metric, compact = false }) => {
  const series = metricChartSeries(metric)
  if (series.length === 0) return null
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
        formatter: (value) => new Date(value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
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
      smooth: true,
      showSymbol: false,
      symbolSize: 4,
      lineStyle: { width: compact ? 2 : 2.5 },
      emphasis: { focus: 'series' },
      data: item.data,
    })),
  }

  return (
    <div className={compact ? 'health-metric-trend-chart health-metric-trend-chart-compact' : 'health-metric-trend-chart'}>
      <ReactECharts option={option} style={{ width: '100%', height: compact ? 120 : 180 }} />
      {metric.pointsTruncated && <span className="health-metric-trend-note">已按窗口抽样展示</span>}
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
  const normalized = text.replace(/series/g, '序列')
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

const summarizeTraceEvidenceText = (text = '') =>
  truncateText(
    text
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

const productStructuredSummary = (text = '') =>
  truncateText(text
    .replace(/^AI\s*/, '')
    .replace(/AI\s*/g, '')
    .replace(/影响对象/g, '故障目标')
    .replace(/、?\d+\s*类已接入数据源/g, '')
    .replace(/，?计算路径风险\s*[\d.]+/g, '')
    .replace(/\s*未接入的.*?不参与推断。?/g, '')
    .replace(/\s{2,}/g, ' ')
    .trim(), 260)

const productReportSummary = (text = '', aiGenerated = false) => {
  const summary = safeText(text).trim()
  if (aiGenerated) return truncateText(summary, 260)
  return productStructuredSummary(summary)
}

const productIssueConclusion = (issue = {}, aiGenerated = false) =>
  productReportSummary(firstText(issue.aiConclusion, issue.systemProblem, issue.impact, issue.mergeBasis), aiGenerated)

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

const summarizeEvidenceItem = (item = {}) => {
  const type = safeText(item.type)
  if (type.includes('告警指标趋势')) {
    return summarizeMetricEvidenceText(item.result || '')
  }
  if (type.includes('Trace链路样本')) {
    return summarizeTraceEvidenceText(item.result || '')
  }
  if (type.includes('告警事件') || type.includes('告警明细')) {
    return stripTechnicalEvidenceText(item.result || '')
  }
  return productReportSummary(stripTechnicalEvidenceText(item.result || ''))
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
    return '入口'
  }
  if (objectTypes.has('service')) {
    return '服务'
  }
  if (objectTypes.has('node')) {
    return '节点'
  }
  if (objectTypes.has('pod')) {
    return 'Pod'
  }
  return '影响目标'
}

const alertFrequencyLabel = (eventCount) => {
  if (eventCount >= 5) return '频率较高'
  if (eventCount >= 2) return '反复出现'
  if (eventCount === 1) return '单次偶发'
  return ''
}

const extractAlertPatternSummary = (alertEvidence = {}, issue = {}) => {
  const rawResult = safeText(alertEvidence.result)
  const patternMatch = rawResult.match(/事件规律=([^，。]+)/)
  if (patternMatch?.[1]) {
    return truncateText(patternMatch[1], 180) + '。'
  }

  const eventCount = positiveNumber(issue.rawEventCount)
  const instanceCount = positiveNumber(issue.deduplicatedAlerts)
  const activeCount = positiveNumber(issue.activeAlerts)
  if (eventCount <= 1) return ''

  const parts = []
  if (eventCount > instanceCount && instanceCount > 0) {
    parts.push('同一告警实例存在重复触发')
  }
  if (instanceCount > 1) {
    parts.push(`涉及 ${instanceCount} 个去重实例`)
  }
  if (activeCount === 0) {
    parts.push('事件已恢复但在窗口内反复出现')
  }
  return parts.length > 0 ? `${parts.join('，')}。` : ''
}

const buildMacroEvidenceSummary = (issue = {}, alertEvidence = {}, detailEvidence = {}) => {
  const eventCount = positiveNumber(issue.rawEventCount)
  if (eventCount === 0 && !alertEvidence?.result && !detailEvidence?.result) return ''

  const activeCount = positiveNumber(issue.activeAlerts)
  const instanceCount = positiveNumber(issue.deduplicatedAlerts)
  const symptom = inferRiskSymptom(issue, alertEvidence, detailEvidence)
  const objectLabel = issueObjectLabel(issue)
  const frequency = alertFrequencyLabel(eventCount)
  const eventText = eventCount > 0 ? `${eventCount} 次${symptom}告警` : `${symptom}告警`
  const currentState = activeCount > 0
    ? `当前仍有 ${activeCount} 条告警触发`
    : '当前未持续触发'
  const instanceText = instanceCount > 1 ? `，涉及 ${instanceCount} 个实例` : ''
  const observation = activeCount > 0
    ? '需要优先确认影响是否仍在扩大'
    : eventCount >= 3
      ? '虽然不处于持续告警状态，但告警频率偏高，需要纳入恢复观察'
      : '更像短时抖动，建议结合后续窗口继续观察'

  return `该${objectLabel}在当前窗口发生过 ${eventText}${instanceText}，${currentState}；${frequency ? `告警呈${frequency}，` : ''}${observation}。`
}

const extractAlertEvidenceSummary = (alertEvidence = {}, detailEvidence = {}, issue = {}) => {
  const eventCount = positiveNumber(issue.rawEventCount)
  const instanceCount = positiveNumber(issue.deduplicatedAlerts)
  const activeCount = positiveNumber(issue.activeAlerts)
  const parts = []
  if (eventCount > 0) {
    parts.push(`${eventCount} 条告警事件`)
  }
  if (instanceCount > 0) {
    parts.push(`${instanceCount} 个实例`)
  }
  if (activeCount > 0) {
    parts.push(`${activeCount} 条仍在触发`)
  } else if (eventCount > 0) {
    parts.push('当前未持续触发')
  }
  safeList(alertEvidence.alertEvents).slice(0, 2).forEach((event) => {
    const name = firstText(event.name, event.alertName)
    const status = safeText(event.status)
    const severity = safeText(event.severity)
    if (name) {
      parts.push(`${name}${severity ? ` ${severity}` : ''}${status ? ` ${status}` : ''}`)
    }
  })
  const fallbackText = firstText(alertEvidence.result, detailEvidence.result)
  if (parts.length === 0 && fallbackText) {
    return `告警侧：${stripTechnicalEvidenceText(fallbackText)}`
  }
  return parts.length > 0 ? `告警侧：${Array.from(new Set(parts)).join('，')}。` : ''
}

const extractMetricEvidenceSummary = (metricEvidence = {}, issue = {}) => {
  const metricAnalysis = safeText(issue.metricAnalysis)
  if (metricAnalysis) {
    return `指标侧：${truncateText(metricAnalysis, 220)}`
  }
  const metric = metricEvidence.metric
  if (metric?.eventScoped && positiveNumber(metric.pointCount) > 0) {
    const parts = []
    if (metric.seriesMatched !== undefined) {
      parts.push(`命中 ${metric.seriesMatched} 条事件指标`)
    } else {
      parts.push('已命中事件指标')
    }
    if (metric.ruleName) {
      parts.push(`规则 ${metric.ruleName}`)
    }
    parts.push(`变化 ${formatPercent(metric.changePercent)}`)
    parts.push(`区间 ${formatMetricNumber(metric.minValue)} 到 ${formatMetricNumber(metric.maxValue)}`)
    return `指标侧：${parts.join('，')}。`
  }
  const summary = summarizeMetricEvidenceText(metricEvidence.result || '')
  if (summary) {
    return `指标侧：${summary}`
  }
  if (metricEvidence?.result) {
    return `指标侧：未能按事件标签形成可用趋势。`
  }
  return ''
}

const extractTraceEvidenceSummary = (traceEvidence = {}) => {
  const summary = summarizeTraceEvidenceText(traceEvidence.result || '')
  return summary ? `APM链路侧：${summary}` : ''
}

const extractLogEvidenceSummary = (evidence = []) => {
  const logEvidence = safeList(evidence).find((item) => {
    const type = safeText(item.type)
    const source = safeText(item.source)
    return type.includes('日志') || source.includes('日志')
  })
  if (logEvidence?.result && !shouldHideDefaultEvidence(logEvidence)) {
    return `日志分析侧：${summarizeEvidenceItem(logEvidence)}`
  }
  return ''
}

const extractObjectStatusSummary = (evidence = []) => {
  const objectEvidence = safeList(evidence).find((item) => safeText(item.type).includes('对象状态'))
  const statuses = safeList(objectEvidence?.objectStatuses)
    .map((item) => {
      const name = firstText(item.name, item.namespace)
      const status = safeText(item.status)
      if (!name || !status) return ''
      return `${name} ${status}`
    })
    .filter(Boolean)
    .slice(0, 2)
  if (statuses.length > 0) {
    return `运行目标侧：${statuses.join('，')}。`
  }
  if (objectEvidence?.result && !shouldHideDefaultEvidence(objectEvidence)) {
    return `运行目标侧：${summarizeEvidenceItem(objectEvidence)}`
  }
  return ''
}

const buildIntegratedEvidenceText = ({
  evidence,
  issue,
  alertEvidence,
  metricEvidence,
  traceEvidence,
  detailEvidence,
}) => {
  const parts = [
    extractAlertPatternSummary(alertEvidence, issue),
    buildMacroEvidenceSummary(issue, alertEvidence, detailEvidence),
    extractAlertEvidenceSummary(alertEvidence, detailEvidence, issue),
    extractMetricEvidenceSummary(metricEvidence, issue),
    extractTraceEvidenceSummary(traceEvidence),
    extractLogEvidenceSummary(evidence),
    extractObjectStatusSummary(evidence),
  ].filter(Boolean)

  if (parts.length === 0) return ''
  return truncateText(parts.join(' '), 420)
}

const buildAlertBehaviorAnalysis = (issue = {}, alertEvidence = {}, detailEvidence = {}) => {
  const eventCount = positiveNumber(issue.rawEventCount)
  const instanceCount = positiveNumber(issue.deduplicatedAlerts)
  const activeCount = positiveNumber(issue.activeAlerts)
  const symptom = inferRiskSymptom(issue, alertEvidence, detailEvidence)
  const objectLabel = issueObjectLabel(issue)
  const frequency = alertFrequencyLabel(eventCount)
  const validAnalyses = safeList(issue.existingAlertAnalyses).filter((item) => safeText(item.validity) === 'valid').length
  const invalidAnalyses = safeList(issue.existingAlertAnalyses).filter((item) => safeText(item.validity) === 'invalid').length
  const stateText = activeCount > 0
    ? `当前仍有 ${activeCount} 条告警触发`
    : '当前已恢复，进入恢复观察'
  const scaleText = [
    eventCount > 0 && `${eventCount} 次告警`,
    instanceCount > 0 && `${instanceCount} 个去重实例`,
  ].filter(Boolean).join('、')
  const validityText = validAnalyses > 0
    ? `已有分析认为其中 ${validAnalyses} 类告警有效`
    : invalidAnalyses > 0
      ? `已有分析认为其中 ${invalidAnalyses} 类告警可能无效`
      : ''

  return [
    `${objectLabel}侧表现为${symptom}告警${frequency ? ` ${frequency}` : ''}出现${scaleText ? `，涉及 ${scaleText}` : ''}`,
    stateText,
    validityText,
  ].filter(Boolean).join('；') + '。'
}

const buildIssueAnalysisText = (issue = {}, aiGenerated = false) => {
  const base = productIssueConclusion(issue, aiGenerated)
  const metricAnalysis = safeText(issue.metricAnalysis)
  const rootCause = safeText(issue.rootCause)
  const direction = metricAnalysis || rootCause
  if (!direction) return base
  const directionText = `指标方向：${truncateText(direction, 180)}`
  if (!base) return directionText
  if (base.includes(direction) || base.includes('指标方向')) return base
  return truncateText(`${base} ${directionText}`, 360)
}

const buildKeyEvidenceItems = (evidence = [], issue = {}) => {
  const alertEvidence = getEvidenceByType(evidence, '告警事件')
  const metricEvidence = getEvidenceByType(evidence, '告警指标趋势')
  const traceEvidence = getEvidenceByType(evidence, 'Trace链路样本')
  const detailEvidence = getEvidenceByType(evidence, '告警明细')
  const missingData = safeList(issue.missingData)
  const supplementalEvidence = safeList(evidence).filter(
    (item) => !['告警事件', '告警指标趋势', 'Trace链路样本', '告警明细'].some((keyword) => (item.type || '').includes(keyword)),
  )
  const items = []
  const existingAlertAnalysisText = buildExistingAlertAnalysisText(issue)
  const alertBehaviorText = buildAlertBehaviorAnalysis(issue, alertEvidence, detailEvidence)
  const alertAnalysisText = [alertBehaviorText, existingAlertAnalysisText && `已有告警分析：${existingAlertAnalysisText}`]
    .filter(Boolean)
    .join(' ')
  const rootCauseText = safeText(issue.rootCause)
  const integratedEvidenceText = buildIntegratedEvidenceText({
    evidence,
    issue,
    alertEvidence,
    metricEvidence,
    traceEvidence,
    detailEvidence,
  })

  if (integratedEvidenceText) {
    items.push({
      label: '证据汇总',
      text: integratedEvidenceText,
      tone: 'summary',
    })
  }

  if (rootCauseText) {
    items.push({
      label: '根因判断',
      text: truncateText(rootCauseText, 260),
      tone: 'rootcause',
    })
  }

  if (alertAnalysisText) {
    items.push({
      label: '告警分析',
      text: truncateText(alertAnalysisText, 360),
      tone: 'analysis',
    })
  }

  const metricAnalysis = safeText(issue.metricAnalysis)
  if (metricEvidence?.metric?.eventScoped && metricEvidence.metric.pointCount > 0) {
    const metric = metricEvidence.metric
    const matched = metric.seriesMatched !== undefined && metric.seriesReturned !== undefined
      ? `命中 ${metric.seriesMatched} 条事件指标`
      : '已返回指标趋势'
    items.push({
      label: '指标变化',
      text: metricAnalysis || `${matched}，窗口变化 ${formatPercent(metric.changePercent)}。`,
      tone: metric.eventScoped ? 'metric' : 'warning',
      metric,
    })
  } else if (metricAnalysis) {
    items.push({
      label: '指标变化',
      text: metricAnalysis,
      tone: 'metric',
    })
  } else if (metricEvidence?.result) {
    const metricSummary = summarizeMetricEvidenceText(metricEvidence.result)
    if (metricSummary) {
      items.push({
        label: '指标变化',
        text: metricSummary,
        tone: 'metric',
      })
    }
  }

  if (traceEvidence?.result) {
    items.push({
      label: '链路样本',
      text: summarizeTraceEvidenceText(traceEvidence.result),
      tone: 'trace',
    })
  }

  if (!alertEvidence?.result && detailEvidence?.result) {
    items.push({
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
      label: role.stage,
      text,
      tone: role.kind === 'boundary' ? 'warning' : role.kind,
    })
  })

  if (items.length === 0 && missingData.length > 0) {
    items.push({
      label: '数据边界',
      text: `${missingData.join('、')} 暂不可用，相关趋势或运行状态未参与当前结论。`,
      tone: 'warning',
    })
  }

  const seen = new Set()
  return items.filter((item) => {
    const key = `${item.label}:${item.text}`
    if (!item.text || seen.has(key)) return false
    seen.add(key)
    return true
  }).slice(0, 6)
}

export default function ServiceHealthyPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const [keyword, setKeyword] = useState('')
  const [healthReportData, setHealthReportData] = useState(null)
  const [reportError, setReportError] = useState('')
  const [isReportRefreshing, setIsReportRefreshing] = useState(false)
  const lowerKeyword = keyword.trim().toLowerCase()

  useEffect(() => {
    let ignore = false
    const searchParams = new URLSearchParams(location.search)
    const requestParams = {
      groupId: searchParams.get('groupId') || 1,
      relativeTime: searchParams.get('relativeTime') || '15m',
      startTime: searchParams.get('startTime') || undefined,
      endTime: searchParams.get('endTime') || undefined,
    }
    const cacheKey = makeReportCacheKey(requestParams)
    const cached = readPersistedReport(cacheKey)
    if (cached) {
      setHealthReportData(cached.data)
      setReportError('')
    } else {
      setHealthReportData(null)
    }
    setIsReportRefreshing(true)
    getSystemHealthReportApi(requestParams)
      .then((data) => {
        if (!ignore && data?.report && Array.isArray(data?.issues)) {
          setHealthReportData(data)
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
            setReportError('真实环境数据读取失败，请确认后端 /api/system-health/report 已启动并可访问。')
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
  }, [location.search])

  const issues = healthReportData?.issues ?? []

  const filteredIssues = useMemo(() => {
    if (!lowerKeyword) return issues
    return issues.filter(
      (item) => [
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
      ].join(' ').toLowerCase().includes(lowerKeyword),
    )
  }, [lowerKeyword, issues])

  const report = healthReportData?.report ?? {
    ...EMPTY_REPORT,
    summary: reportError || EMPTY_REPORT.summary,
  }
  const reportStatusMeta = LEVEL_META[report.status] ?? LEVEL_META.warning
  const reportScore = Number.isFinite(Number(report.score)) ? Number(report.score) : 0
  const aiAnalysis = report.aiAnalysis || {}
  const isAISummaryReady = aiAnalysis.status === 'succeeded'
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

  const affectedScopeSummary = summarizeAffectedObjectScope(filteredIssues)
  const impactSummary = {
    p0Count: filteredIssues.filter((issue) => issue.priority === 'P0').length,
    p1Count: filteredIssues.filter((issue) => issue.priority === 'P1').length,
    p2Count: filteredIssues.filter((issue) => issue.priority === 'P2').length,
    p3Count: filteredIssues.filter((issue) => issue.priority === 'P3').length,
    p4Count: filteredIssues.filter((issue) => issue.priority === 'P4').length,
    affectedObjectCount: affectedScopeSummary.count,
  }
  const visibleKpis = [
    { label: lowerKeyword ? '筛选结果' : '风险项', value: filteredIssues.length, show: true },
    { label: 'P0/P1', value: `${impactSummary.p0Count}/${impactSummary.p1Count}`, show: impactSummary.p0Count + impactSummary.p1Count > 0 },
    { label: 'P2/P3', value: `${impactSummary.p2Count}/${impactSummary.p3Count}`, show: impactSummary.p2Count + impactSummary.p3Count > 0 },
    { label: 'P4', value: impactSummary.p4Count, show: impactSummary.p4Count > 0 },
    { label: '触发中告警', value: report.activeAlertCount, show: report.activeAlertCount > 0 },
    {
      label: '恢复观察',
      value: filteredIssues.filter((issue) => issue.status === 'recovering' && positiveNumber(issue.activeAlerts) === 0).length,
      show: filteredIssues.some((issue) => issue.status === 'recovering' && positiveNumber(issue.activeAlerts) === 0),
    },
    { label: affectedScopeSummary.label, value: affectedScopeSummary.value, show: impactSummary.affectedObjectCount > 0 },
  ].filter((item) => item.show)
  const reportSummaryText = reportExecutiveSummary(report, issues, productReportSummary(report.summary || '', isAISummaryReady))

  if (!healthReportData) {
    return (
      <div className="system-health-page">
        <div className="health-report-header">
          <div className="health-report-main">
            <div className="health-title-row">
              <div className="health-title">系统健康报告</div>
              {isReportRefreshing && <Tag color="processing">刷新中</Tag>}
            </div>
          </div>
        </div>
        <div className="health-impact-empty">
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={reportError || '正在生成后端分析'}
          />
        </div>
      </div>
    )
  }

  const impactCommandOverview = (
    <div className="health-impact-command">
      <div className="health-impact-command-strip">
        <div className="health-impact-command-status">
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
        <div className="health-impact-command-readout">
          <strong>{reportSummaryText}</strong>
          {aiMeta && aiAnalysis.status !== 'not_configured' && (
            <small>{aiMeta.text}</small>
          )}
        </div>
        <div className="health-impact-command-kpis">
          {visibleKpis.map((item) => (
            <span key={item.label}>{item.label} <b>{item.value}</b></span>
          ))}
        </div>
      </div>
    </div>
  )

  const impactPathVersion = (
    <div className="health-version-impact">
      {impactCommandOverview}
      <div className="health-impact-detail-title">
        <b>健康风险明细</b>
        <span>{filteredIssues.length} 项</span>
      </div>
      <div className="health-impact-grid">
        {filteredIssues.length === 0 && (
          <div className="health-impact-empty">
            {reportSummaryText ? (
              <div className="health-impact-empty-summary">
                <strong>健康风险明细</strong>
                <span>{reportSummaryText}</span>
              </div>
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={reportError || '暂无健康风险明细'}
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
            }))
          const visibleAffectedObjects = visibleObjectTags(issue)
          const keyEvidence = buildKeyEvidenceItems(evidence, issue)
          const issueAnalysis = buildIssueAnalysisText(issue, isAISummaryReady)
          const shouldShowRoute = displayTitle !== scope || visibleAffectedObjects.length > 0
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
                <StatusTag status={status} />
                <div className="health-impact-card-meta">
                  {metaItems.map((item) => <span key={`${issue.key}-${item}`}>{item}</span>)}
                </div>
              </div>
              {shouldShowRoute && (
                <div className="health-impact-route">
                  <div>
                    <strong>影响范围</strong>
                    {displayTitle !== scope && <span>{scope}</span>}
                    {visibleAffectedObjects.length > 0 && (
                      <div className="health-impact-object-row">
                        {visibleAffectedObjects.map((object) => (
                          <Tag bordered={false} key={`${issue.key}-${object.type}-${object.name}`}>
                            {objectTypeText[object.type] || object.type}: {object.name}
                          </Tag>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )}
              <div className="health-impact-evidence">
                {issueAnalysis && (
                  <div className="health-impact-analysis">
                    <strong>分析结论</strong>
                    <span>{issueAnalysis}</span>
                  </div>
                )}
                {keyEvidence.length > 0 && (
                  <>
                    <div className="health-impact-evidence-head">
                      <b>关键依据</b>
                    </div>
                    <div className="health-impact-key-evidence">
                      {keyEvidence.map((item) => (
                        <div className={`health-impact-key-evidence-item health-impact-key-evidence-${item.tone}`} key={`${issue.key}-${item.label}`}>
                          <strong>{item.label}</strong>
                          <span>{item.text}</span>
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
                    </div>
                  </div>
                  {actionSteps.map((step, index) => (
                    <div className="health-impact-action-step" key={`${issue.key}-legacy-${step.phase}-${index}`}>
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
    <div className="system-health-page">
      <div className="health-report-header">
        <div className="health-report-main">
          <div className="health-title-row">
            <div className="health-title">系统健康报告</div>
            <StatusTag status={report.status} />
            {isReportRefreshing && <Tag color="processing">刷新中</Tag>}
            {reportError && healthReportData && <Tag color="warning">已显示缓存</Tag>}
            {aiMeta && aiAnalysis.status !== 'not_configured' && <Tag color={aiMeta.color}>{aiMeta.text}</Tag>}
          </div>
          <div className="health-report-meta">
            <span>时间窗口：{report.window || '当前窗口'}</span>
            {report.generatedAt && <span>分析时间：{report.generatedAt}</span>}
          </div>
        </div>
        <Space wrap className="health-report-actions">
          <Input.Search
            allowClear
            placeholder="过滤风险 / 服务 / 接口"
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            style={{ width: 220 }}
          />
        </Space>
      </div>

      {impactPathVersion}
    </div>
  )
}
