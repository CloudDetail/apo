export type ReportType = 'slow' | 'error' | 'infra' | string
export type ReportStatus = 'active' | 'recovering' | 'resolved' | 'snapshot' | 'unknown'
export type SuggestLevel = 'high' | 'medium' | 'low' | string
export type CandidateState = 'primary' | 'secondary' | 'excluded' | 'unknown'
export type ReportSummaryCardTone = 'symptom' | 'rootCause' | 'excluded'

export interface AlertEventReportResponse {
  reportType: ReportType
  reportData: {
    overview: {
      timestamp: number
      detail: string
      reason: string
      tags: Record<string, string>
    }
    topology?: ReportTopology
    rootCauseAnalysis: {
      cause: string
      other: RootCauseCandidate[]
    }
    suggest: ReportSuggest[]
    analyzeProcess?: string
    evidence: ReportEvidence[]
    reportView?: ReportViewContract | null
  }
}

export interface AlertEventReportViewResponse {
  snapshot: AlertEventReportResponse
  view?: {
    runtimeStatus?: ReportRuntimeStatus
    dataSources?: Record<string, string>
  }
}

export interface ReportEvidenceRef {
  evidenceIndex: number
  title?: string
  summary?: string
}

export interface ReportViewContract {
  version?: string
  headline?: string
  conclusionSummary?: string
  faultLocation?: {
    label?: string
    target?: string
    summary?: string
  } | null
  summaryCards?: Array<{
    type?: 'symptom' | 'rootCause' | 'excluded' | string
    title?: string
    summary?: string
  }>
  rootCauseEvidenceRefs?: ReportEvidenceRef[]
  rootCauseMetricRefs?: ReportEvidenceRef[]
  businessImpact?: {
    summary?: string
    endpoints?: string[]
    entryServices?: string[]
    affectedServices?: string[]
  } | null
  timeline?: ReportTimelineItem[]
  analysisSteps?: ReportTimelineItem[]
  actionTitles?: Array<{
    suggestIndex: number
    title: string
  }>
  evidenceHighlights?: ReportEvidenceRef[]
}

export interface ReportTopology {
  current?: ReportTopologyNode
  parents?: ReportTopologyNode[]
  children?: ReportTopologyNode[]
}

export interface ReportTopologyNode {
  service: string
  description?: string
  metrics?: Record<string, { value: number | string }>
}

export interface RootCauseCandidate {
  level: 'major' | 'minor' | 'normal' | string
  type: string
  reason: string
}

export interface ReportSuggest {
  action: string
  summary?: string
  level?: SuggestLevel
  category?: string
  verificationSummary?: string
  verification?: string
  trigger?: string
  risk?: string
}

export interface ReportActionDetail {
  label: string
  text: string
}

export interface ReportAction extends ReportSuggest {
  displayTitle: string
  details: ReportActionDetail[]
}

export interface ReportEvidence {
  name: string
  type: 'log' | 'red' | 'metrics' | 'endpoints' | string
  summary?: string
  description?: string
  unit?: string
  data: any
}

export interface ReportRuntimeStatus {
  state: ReportStatus
  label: string
  summary: string
  updatedAt?: string | number
}

export interface ReportContext {
  id?: string
  startTime?: number
  endTime?: number
  dataSourceLabel?: string
  runtimeStatus?: ReportRuntimeStatus
}

export interface DiagnosisCandidate {
  rank: number
  direction: string
  judgement: string
  evidence: string
  conclusion: string
  state: CandidateState
}

export interface EvidenceSummaryItem {
  evidenceIndex: number
  name: string
  type: string
  typeLabel: string
  description: string
  valueSummary: string
}

export interface CoreMetricChartSeries {
  legend: string
  unit?: string
  focus?: boolean
  data: Array<[number, number | string]>
}

export interface CoreMetricSummaryItem {
  name: string
  valueSummary: string
  description: string
  unit?: string
  chartData?: CoreMetricChartSeries[]
}

export interface RootCauseEvidenceItem extends CoreMetricSummaryItem {
  type: string
  typeLabel: string
}

export interface FaultLocationSummary {
  label: string
  service: string
  description: string
  path: string[]
  basis: string
}

export interface BusinessImpactSummary {
  summary: string
  endpoints: string[]
  entryServices: string[]
  affectedServices: string[]
}

export interface ReportImpactGroup {
  label: string
  values: string[]
}

export interface ReportTimelineItem {
  timestamp?: number | string
  time?: string
  title: string
  summary?: string
  detail?: string
}

export interface ReportSummaryCard {
  label: string
  title: string
  description: string
  tone: ReportSummaryCardTone
}

export interface ReportViewModel {
  id: string
  reportType: ReportType
  typeLabel: string
  generatedAt: string
  analysisWindow: string
  dataSourceLabel: string
  status: ReportRuntimeStatus
  headline: string
  analysisSummary: string
  rootCause: string
  overview: string
  reason: string
  tags: Array<{ key: string; value: string }>
  candidates: DiagnosisCandidate[]
  checkedCount: number
  primaryCount: number
  secondaryCount: number
  excludedCount: number
  primaryAction?: ReportAction
  actions: ReportAction[]
  coreMetrics: CoreMetricSummaryItem[]
  rootCauseEvidence: RootCauseEvidenceItem[]
  evidenceSummary: EvidenceSummaryItem[]
  evidence: ReportEvidence[]
  topology?: ReportTopology
  topologySummary: string[]
  faultLocation: FaultLocationSummary
  businessImpact: BusinessImpactSummary
  impactGroups: ReportImpactGroup[]
  summaryCards: ReportSummaryCard[]
  timeline: ReportTimelineItem[]
  analyzeProcess?: string
  startTime: number
  endTime: number
}

const reportTypeLabels: Record<string, string> = {
  slow: '慢延时报告',
  error: '错误报告',
  infra: '基础设施报告',
}

const rootCauseTypeLabels: Record<string, string> = {
  cpu: 'CPU 异常',
  code: '代码异常',
  log: '日志异常',
  retry: '重试异常',
  net: '网络异常',
  storage: '存储异常',
  runtime: '运行时异常',
  infra: '基础设施异常',
  downstream: '下游依赖',
  dependency: '下游依赖',
  traffic: '流量异常',
  thirdparty: '第三方异常',
  third_party: '第三方异常',
  'third-party': '第三方异常',
}

const evidenceTypeLabels: Record<string, string> = {
  log: '日志',
  red: 'RED 指标',
  metrics: '指标',
  trace: 'Trace',
  traces: 'Trace',
  endpoints: '拓扑',
}

const actionLevelOrder: Record<string, number> = {
  high: 0,
  medium: 1,
  low: 2,
}

const rootCauseMetricKeywords: Record<string, string[]> = {
  cpu: ['cpu', '处理器', 'throttling', 'load', '错误日志', '日志', 'retry', '重试', 'gc'],
  code: ['异常', 'exception', 'error', '错误码', 'panic', 'stack', '日志', 'retry', '重试', 'cpu'],
  log: ['日志', 'log', '错误模板', 'template', 'error'],
  retry: ['retry', '重试', '重复调用', '下游调用次数', '调用次数'],
  net: ['rtt', '网络', '连接失败', '连接超时', 'timeout', 'tcp', '重传', '丢包', 'dns'],
  storage: ['db', '数据库', 'sql', '慢 sql', '磁盘', 'io', 'i/o', '连接池', '存储'],
  runtime: ['gc', 'heap', '堆', '线程', 'goroutine', 'fd', 'oom', '内存', 'cpu'],
  infra: ['pod', 'restart', 'oomkill', 'node', '节点', '容器', 'k8s', 'event', 'cpu', '内存', '磁盘'],
  downstream: ['下游', '依赖', 'latency', '延迟', 'error', '错误率', 'timeout', '超时', '5xx', '接口', '调用', 'mysql', 'db', 'database', '数据库', 'sql', '查询', '锁', '连接池', '耗时', '变慢'],
  dependency: ['下游', '依赖', 'latency', '延迟', 'error', '错误率', 'timeout', '超时', '5xx', '接口', '调用', 'mysql', 'db', 'database', '数据库', 'sql', '查询', '锁', '连接池', '耗时', '变慢'],
  traffic: ['qps', 'rps', '请求', '请求量', '并发', '队列', 'traffic', '流量', '吞吐'],
  thirdparty: ['第三方', '供应商', 'vendor', 'provider', 'gateway', 'timeout', 'error', '5xx'],
  third_party: ['第三方', '供应商', 'vendor', 'provider', 'gateway', 'timeout', 'error', '5xx'],
  'third-party': ['第三方', '供应商', 'vendor', 'provider', 'gateway', 'timeout', 'error', '5xx'],
}

const redAllowedRootCauseTypes = new Set(['downstream', 'dependency', 'traffic', 'thirdparty', 'third_party', 'third-party'])
const rootCauseEvidenceTypes = new Set(['metrics', 'red', 'log', 'trace', 'traces'])

const normalEvidencePatterns = [
  '表现正常',
  '整体正常',
  '运行正常',
  '状态正常',
  '未见异常',
  '未发现异常',
  '未检测到异常',
  '未检测到 rtt 异常',
  '未检测到rtt异常',
  '无明显异常',
  '无报错',
  '未见报错',
  '错误率为 0',
  '错误率为0',
  '错误率始终为 0',
  '错误率始终为0',
  '未采集到任何有效日志',
  '暂时无法判断',
  '不支持定位',
  '暂不支持',
]

const strongSupportPatterns = [
  'read timed out',
  'readtimedout',
  'sockettimeoutexception',
  'resourceaccessexception',
  'timed out',
  'timeout',
  '超时',
  '慢',
  '耗时',
  '升高',
  '高于',
  '5xx',
  '不可用',
  '失败',
  'exception',
  '异常增长',
  '异常升高',
]

export const defaultReportStatus: ReportRuntimeStatus = {
  state: 'snapshot',
  label: '诊断快照',
  summary: '报告内容表示生成时刻的诊断结果。',
}

function cleanText(value?: string | number) {
  if (value === undefined || value === null) return ''
  return String(value)
    .replace(/```[a-zA-Z0-9_-]*\s*/g, '')
    .replace(/```/g, '')
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/\*\*(.*?)\*\*/g, '$1')
    .replace(/__(.*?)__/g, '$1')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/[*_~]/g, '')
    .split(/\n+/)
    .map((line) => line.trim().replace(/^[-*+]\s+/, ''))
    .filter(Boolean)
    .join(' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function normalizeRootCauseType(type?: string) {
  const value = cleanText(type).toLowerCase()
  if (value === 'third-party') return 'thirdparty'
  if (value === 'third_party') return 'thirdparty'
  return value
}

function normalizeEvidenceType(type?: string) {
  return cleanText(type).toLowerCase()
}

function uniqueFilled(values: Array<string | undefined>) {
  return [...new Set(values.map((value) => cleanText(value)).filter(Boolean))]
}

function primaryRootCauseType(candidates: RootCauseCandidate[]) {
  const major = candidates.find((candidate) => candidate.level === 'major')
  return normalizeRootCauseType(major?.type || candidates[0]?.type)
}

function candidateState(level: string): CandidateState {
  if (level === 'major') return 'primary'
  if (level === 'minor') return 'secondary'
  if (level === 'normal') return 'excluded'
  return 'unknown'
}

function buildCandidates(candidates: RootCauseCandidate[]): DiagnosisCandidate[] {
  return candidates.map((candidate, index) => {
    const state = candidateState(candidate.level)
    return {
      rank: index + 1,
      direction: rootCauseTypeLabels[normalizeRootCauseType(candidate.type)] || candidate.type || '未分类方向',
      judgement:
        state === 'primary'
          ? '高度怀疑'
          : state === 'secondary'
            ? '待验证'
            : state === 'excluded'
              ? '已排除'
              : '已检查',
      evidence: cleanText(candidate.reason),
      conclusion:
        state === 'primary'
          ? '当前首要根因'
          : state === 'secondary'
            ? '需要继续验证'
            : state === 'excluded'
              ? '非根因'
              : '已纳入排查',
      state,
    }
  })
}

function normalizeTimestamp(timestamp?: number) {
  if (!timestamp) return 0
  return timestamp > 10_000_000_000_000 ? timestamp / 1000 : timestamp
}

export function formatReportTime(timestamp?: number) {
  const normalized = normalizeTimestamp(timestamp)
  if (!normalized) return '-'
  return new Intl.DateTimeFormat('zh-CN', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  })
    .format(new Date(normalized))
    .replace(/\//g, '-')
}

function formatMetricValue(value: number | string, unit?: string) {
  if (typeof value === 'string') return value
  const normalizedUnit = cleanText(unit).toLowerCase()
  if (normalizedUnit === 'percent' || normalizedUnit === '%') return `${Number(value.toFixed(2))}%`
  if (normalizedUnit === 's' || normalizedUnit === 'second' || normalizedUnit === 'seconds') {
    if (Math.abs(value) < 1) return `${Number((value * 1000).toFixed(2))} ms`
    return `${Number(value.toFixed(2))} s`
  }
  if (normalizedUnit === 'bytes') return `${Number((value / 1024 / 1024).toFixed(2))} MB`
  if (normalizedUnit && normalizedUnit !== 'count') return `${Number(value.toFixed(2))} ${unit}`
  return `${Number(value.toFixed(2))}`
}

function chartDataFromSeries(series: any) {
  const rawChartData = series?.chart?.chartData || series?.chartData
  if (!rawChartData || typeof rawChartData !== 'object') return []
  return Object.entries(rawChartData)
    .map(([timestamp, value]) => [Number(timestamp) / 1000, value] as [number, number | string])
    .filter(([timestamp, value]) => Number.isFinite(timestamp) && value !== undefined && value !== null)
    .sort(([left], [right]) => left - right)
}

function normalizeSeriesMatchText(text: string) {
  return cleanText(text).toLowerCase().replace(/\s+/g, '')
}

function getSeriesFocusScore(series: any, ref?: ReportEvidenceRef) {
  const refText = normalizeSeriesMatchText(`${ref?.title || ''} ${ref?.summary || ''}`)
  if (!refText) return 0

  const seriesText = normalizeSeriesMatchText(
    [series?.legend, series?.name, series?.type].filter(Boolean).join(' '),
  )
  if (!seriesText) return 0
  if (refText.includes(seriesText) || seriesText.includes(refText)) return 100

  const ipMatches = seriesText.match(/\b\d{1,3}(?:\.\d{1,3}){3}\b/g) || []
  if (ipMatches.some((ip) => refText.includes(ip))) return 80

  const podLikeTokens = seriesText.match(/[a-z0-9]+(?:-[a-z0-9]+){2,}/g) || []
  if (podLikeTokens.some((token) => token.length >= 12 && refText.includes(token))) return 60

  return 0
}

function metricChartData(evidence: ReportEvidence, ref?: ReportEvidenceRef): CoreMetricChartSeries[] {
  if (!Array.isArray(evidence.data)) return []
  const chartSeries = evidence.data
    .map((series): { series: CoreMetricChartSeries; focusScore: number } | undefined => {
      const data = chartDataFromSeries(series)
      if (!data.length) return undefined
      const unit = cleanText(series?.unit) || cleanText(evidence.unit)
      return {
        series: {
          legend: cleanText(series?.legend || series?.name || series?.type) || evidence.name,
          unit,
          data,
        },
        focusScore: getSeriesFocusScore(series, ref),
      }
    })
    .filter(
      (item): item is { series: CoreMetricChartSeries; focusScore: number } => Boolean(item),
    )
  const hasReferencedFocus = chartSeries.some((item) => item.focusScore > 0)

  return chartSeries.map((item, index) => ({
    ...item.series,
    legend: item.series.legend || `series-${index + 1}`,
    focus: hasReferencedFocus ? item.focusScore > 0 : index === 0,
  }))
}

function metricBoundarySummary(evidence: ReportEvidence, fallback?: string) {
  if (fallback) return cleanText(fallback)
  const firstSeries = metricChartData(evidence)[0]
  if (!firstSeries?.data.length) return ''
  const first = firstSeries.data[0][1]
  const last = firstSeries.data[firstSeries.data.length - 1][1]
  const firstValue = typeof first === 'number' ? formatMetricValue(first, evidence.unit) : String(first)
  const lastValue = typeof last === 'number' ? formatMetricValue(last, evidence.unit) : String(last)
  if (!firstValue || !lastValue || firstValue === lastValue) return ''
  return `${firstValue} -> ${lastValue}`
}

function isCompactMetricValueSummary(summary?: string) {
  const text = cleanText(summary)
  if (!text || text.length > 42) return false
  if (!/[0-9]/.test(text)) return false
  return !/[。！？!?；;]/.test(text)
}

function evidenceDataSearchText(data: any) {
  try {
    if (typeof data === 'string') return data.slice(0, 2000)
    if (Array.isArray(data)) {
      return data
        .slice(0, 5)
        .map((item) => (typeof item === 'string' ? item : JSON.stringify(item)))
        .join(' ')
        .slice(0, 2000)
    }
    if (data && typeof data === 'object') return JSON.stringify(data).slice(0, 2000)
  } catch {
    return ''
  }
  return ''
}

function evidenceSearchText(evidence: ReportEvidence, ref?: ReportEvidenceRef) {
  const seriesText = Array.isArray(evidence.data)
    ? evidence.data
        .map((series) => [series?.legend, series?.name, series?.type, series?.unit].filter(Boolean).join(' '))
        .join(' ')
    : ''
  return cleanText(
    `${ref?.title || ''} ${ref?.summary || ''} ${evidence.name} ${evidence.summary || ''} ${evidence.description || ''} ${evidence.unit || ''} ${seriesText} ${evidenceDataSearchText(evidence.data)}`,
  ).toLowerCase()
}

function isClearlyNormalEvidence(evidence: ReportEvidence, ref?: ReportEvidenceRef) {
  const searchText = evidenceSearchText(evidence, ref)
  const compactText = searchText.replace(/\s+/g, '')
  const hasNormalPattern = normalEvidencePatterns.some((pattern) => {
    const normalizedPattern = pattern.toLowerCase()
    return searchText.includes(normalizedPattern) || compactText.includes(normalizedPattern.replace(/\s+/g, ''))
  })

  if (!hasNormalPattern) return false

  return !strongSupportPatterns.some((pattern) => {
    const normalizedPattern = pattern.toLowerCase()
    return searchText.includes(normalizedPattern) || compactText.includes(normalizedPattern.replace(/\s+/g, ''))
  })
}

function evidenceSupportsRootCause(
  evidence: ReportEvidence,
  rootCauseType: string,
  ref?: ReportEvidenceRef,
  requireKeyword = true,
) {
  const normalizedType = normalizeRootCauseType(rootCauseType)
  const evidenceType = normalizeEvidenceType(evidence.type)
  if (!normalizedType || !rootCauseEvidenceTypes.has(evidenceType)) return false
  if (isClearlyNormalEvidence(evidence, ref)) return false

  if (evidenceType === 'red' && !redAllowedRootCauseTypes.has(normalizedType)) return false
  if ((evidenceType === 'metrics' || evidenceType === 'red') && !metricChartData(evidence).length) {
    return false
  }

  if (!requireKeyword) return true

  const keywords = rootCauseMetricKeywords[normalizedType] || [normalizedType]
  const searchText = evidenceSearchText(evidence, ref)
  return keywords.some((keyword) => searchText.includes(keyword.toLowerCase()))
}

function evidenceMatchesRootCause(
  evidence: ReportEvidence,
  rootCauseType: string,
  ref?: ReportEvidenceRef,
  requireKeyword = true,
) {
  const evidenceType = normalizeEvidenceType(evidence.type)
  if (evidenceType !== 'metrics' && evidenceType !== 'red') return false
  return evidenceSupportsRootCause(evidence, rootCauseType, ref, requireKeyword)
}

function buildMetricFromEvidence(
  evidence: ReportEvidence,
  ref?: ReportEvidenceRef,
): CoreMetricSummaryItem {
  const refSummary = cleanText(ref?.summary)
  const refValueSummary = isCompactMetricValueSummary(refSummary) ? refSummary : undefined
  const valueSummary = refValueSummary
    ? metricBoundarySummary(evidence, refValueSummary)
    : refSummary
      ? ''
      : metricBoundarySummary(evidence)

  return {
    name: cleanText(ref?.title) || evidence.name,
    valueSummary,
    description: refSummary && !refValueSummary ? refSummary : cleanText(evidence.description || evidence.summary),
    unit: evidence.unit,
    chartData: metricChartData(evidence, ref),
  }
}

function buildRootCauseEvidenceItem(
  evidence: ReportEvidence,
  ref?: ReportEvidenceRef,
): RootCauseEvidenceItem {
  const metric = buildMetricFromEvidence(evidence, ref)
  const evidenceType = normalizeEvidenceType(evidence.type)
  const isMetricEvidence = evidenceType === 'metrics' || evidenceType === 'red'
  return {
    ...metric,
    type: evidenceType || evidence.type || 'evidence',
    typeLabel: evidenceTypeLabels[evidenceType] || evidence.type || '证据',
    valueSummary:
      isMetricEvidence
        ? metric.valueSummary
        : Array.isArray(evidence.data) && evidence.data.length
          ? `${evidence.data.length} 条样本`
          : '',
    description: isMetricEvidence ? metric.description : cleanText(ref?.summary) || metric.description,
    chartData: isMetricEvidence ? metric.chartData : undefined,
  }
}

function buildRootCauseEvidence(
  evidence: ReportEvidence[],
  reportView: ReportViewContract | null | undefined,
  rootCauseType: string,
) {
  const selected = new Map<number, RootCauseEvidenceItem>()

  const selectReferencedEvidence = (refs: ReportEvidenceRef[], metricOnly: boolean) => {
    refs.forEach((ref) => {
      const evidenceIndex = Number(ref.evidenceIndex)
      const item = evidence[evidenceIndex]
      if (!item) return
      if (metricOnly && !evidenceMatchesRootCause(item, rootCauseType, ref, false)) return
      if (!metricOnly && !evidenceSupportsRootCause(item, rootCauseType, ref, false)) return
      selected.set(evidenceIndex, buildRootCauseEvidenceItem(item, ref))
    })
  }

  if (Array.isArray(reportView?.rootCauseEvidenceRefs)) {
    selectReferencedEvidence(reportView.rootCauseEvidenceRefs, false)
    return Array.from(selected.values()).slice(0, 3)
  }

  if (Array.isArray(reportView?.rootCauseMetricRefs)) {
    selectReferencedEvidence(reportView.rootCauseMetricRefs, true)
    return Array.from(selected.values()).slice(0, 3)
  }

  evidence.forEach((item, index) => {
    if (selected.size >= 3 || selected.has(index)) return
    if (!evidenceMatchesRootCause(item, rootCauseType)) return
    selected.set(index, buildRootCauseEvidenceItem(item))
  })

  return Array.from(selected.values()).slice(0, 3)
}

function buildCoreMetrics(rootCauseEvidence: RootCauseEvidenceItem[]): CoreMetricSummaryItem[] {
  return rootCauseEvidence
    .filter((item) => item.chartData?.length)
    .map(({ name, valueSummary, description, unit, chartData }) => ({
      name,
      valueSummary,
      description,
      unit,
      chartData,
    }))
}

function buildActions(suggest: ReportSuggest[], reportView?: ReportViewContract | null): ReportAction[] {
  const titleByIndex = new Map(
    (reportView?.actionTitles || [])
      .filter((item) => Number.isInteger(item.suggestIndex) && cleanText(item.title))
      .map((item) => [item.suggestIndex, cleanText(item.title)]),
  )

  return suggest
    .map((action, index) => {
      const displayTitle = titleByIndex.get(index) || cleanText(action.summary) || cleanText(action.action) || '未返回处置动作'
      const details: ReportActionDetail[] = [
        cleanText(action.action) && cleanText(action.action) !== displayTitle
          ? { label: '操作', text: cleanText(action.action) }
          : undefined,
        cleanText(action.verification || action.verificationSummary)
          ? { label: '验证', text: cleanText(action.verification || action.verificationSummary) }
          : undefined,
        cleanText(action.trigger) ? { label: '触发', text: cleanText(action.trigger) } : undefined,
        cleanText(action.risk) ? { label: '风险', text: cleanText(action.risk) } : undefined,
      ].filter((item): item is ReportActionDetail => Boolean(item))

      return {
        ...action,
        action: cleanText(action.action),
        displayTitle,
        details,
        originalIndex: index,
        priorityOrder: actionLevelOrder[action.level || ''] ?? 3,
      }
    })
    .sort((a, b) => a.priorityOrder - b.priorityOrder || a.originalIndex - b.originalIndex)
    .map(({ originalIndex, priorityOrder, ...action }) => action)
}

function buildBusinessImpact(reportView?: ReportViewContract | null): BusinessImpactSummary {
  const impact = reportView?.businessImpact
  if (!impact) {
    return { summary: '', endpoints: [], entryServices: [], affectedServices: [] }
  }
  return {
    summary: cleanText(impact.summary),
    endpoints: uniqueFilled(impact.endpoints || []),
    entryServices: uniqueFilled(impact.entryServices || []),
    affectedServices: uniqueFilled(impact.affectedServices || []),
  }
}

function buildImpactGroups(impact: BusinessImpactSummary): ReportImpactGroup[] {
  return [
    { label: 'Endpoint', values: impact.endpoints },
    { label: '入口服务', values: impact.entryServices },
    { label: '影响服务', values: impact.affectedServices },
  ].filter((group) => group.values.length > 0)
}

function buildSummaryCards(response: AlertEventReportResponse, reportView?: ReportViewContract | null): ReportSummaryCard[] {
  const fromView = (reportView?.summaryCards || [])
    .map((card) => {
      const tone =
        card.type === 'rootCause'
          ? 'rootCause'
          : card.type === 'excluded'
            ? 'excluded'
            : 'symptom'
      const label =
        tone === 'rootCause' ? '快速判断原因' : tone === 'excluded' ? '已排除项' : '告警症状'
      return {
        label,
        title: cleanText(card.title),
        description: cleanText(card.summary),
        tone,
      } as ReportSummaryCard
    })
    .filter((card) => card.title || card.description)

  if (fromView.length) return fromView

  const candidates = response.reportData.rootCauseAnalysis.other || []
  const primary = candidates.find((item) => item.level === 'major')
  const excluded = candidates.filter((item) => item.level === 'normal')
  return [
    {
      label: '告警症状',
      title: cleanText(response.reportData.overview.detail) || '未返回告警症状',
      description: '来自报告概述',
      tone: 'symptom',
    },
    {
      label: '快速判断原因',
      title: cleanText(response.reportData.rootCauseAnalysis.cause) || '未返回根因结论',
      description: cleanText(primary?.reason || response.reportData.overview.reason),
      tone: 'rootCause',
    },
    {
      label: '已排除项',
      title: excluded.length ? `已排除${excluded.map((item) => rootCauseTypeLabels[normalizeRootCauseType(item.type)] || item.type).join('、')}` : '暂无已排除方向',
      description: excluded.map((item) => cleanText(item.reason)).filter(Boolean).join('；'),
      tone: 'excluded',
    },
  ]
}

function buildEvidenceSummary(evidence: ReportEvidence[], reportView?: ReportViewContract | null): EvidenceSummaryItem[] {
  const highlighted = (reportView?.evidenceHighlights || [])
    .map((highlight) => {
      const evidenceIndex = Number(highlight.evidenceIndex)
      const item = evidence[evidenceIndex]
      if (!Number.isInteger(evidenceIndex)) return undefined
      if (!item || item.type === 'endpoints') return undefined
      return {
        evidenceIndex,
        name: cleanText(highlight.title) || item.name,
        type: item.type,
        typeLabel: evidenceTypeLabels[item.type] || item.type || '证据',
        description: cleanText(highlight.summary) || cleanText(item.description || item.summary),
        valueSummary: metricBoundarySummary(item),
      }
    })
    .filter((item): item is EvidenceSummaryItem => Boolean(item))

  if (highlighted.length) return highlighted

  return evidence
    .map((item, evidenceIndex) => ({ item, evidenceIndex }))
    .filter(({ item }) => item.type !== 'endpoints')
    .map(({ item, evidenceIndex }) => ({
      evidenceIndex,
      name: item.name,
      type: item.type,
      typeLabel: evidenceTypeLabels[item.type] || item.type || '证据',
      description: cleanText(item.description || item.summary),
      valueSummary:
        item.type === 'metrics' || item.type === 'red'
          ? metricBoundarySummary(item)
          : Array.isArray(item.data)
            ? `${item.data.length} 条样本`
            : '',
    }))
}

function buildFaultLocation(reportView?: ReportViewContract | null): FaultLocationSummary {
  const location = reportView?.faultLocation
  return {
    label: cleanText(location?.label),
    service: cleanText(location?.target),
    description: cleanText(location?.summary),
    path: uniqueFilled([location?.target]),
    basis: cleanText(location?.summary),
  }
}

function timelineTimestampToMilliseconds(value?: number | string) {
  if (value === undefined || value === null || value === '') return undefined
  const timestamp = typeof value === 'number' ? value : Number(cleanText(value))
  if (!Number.isFinite(timestamp)) return undefined

  const absolute = Math.abs(timestamp)
  if (absolute >= 1_000_000_000_000_000) return timestamp / 1000
  if (absolute >= 1_000_000_000_000) return timestamp
  if (absolute >= 1_000_000_000) return timestamp * 1000
  return undefined
}

function getLocalTimeZone() {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || undefined
  } catch {
    return undefined
  }
}

function formatTimelineTimestamp(value?: number | string) {
  const milliseconds = timelineTimestampToMilliseconds(value)
  if (milliseconds === undefined) return ''

  const date = new Date(milliseconds)
  if (Number.isNaN(date.getTime())) return ''

  const timeZone = getLocalTimeZone()
  return date.toLocaleTimeString('zh-CN', {
    ...(timeZone ? { timeZone } : {}),
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function buildTimeline(reportView?: ReportViewContract | null): ReportTimelineItem[] {
  const items = reportView?.timeline || []
  return items
    .map((item) => ({
      time: cleanText(item.time) || formatTimelineTimestamp(item.timestamp),
      title: cleanText(item.title),
      detail: cleanText(item.detail) || cleanText(item.summary),
    }))
    .filter((item) => item.title)
}

function buildTopologySummary(topology?: ReportTopology) {
  if (!topology?.current?.service) return []
  const parts = [`中心服务：${topology.current.service}`]
  if (topology.parents?.length) parts.push(`上游 ${topology.parents.length} 个`)
  if (topology.children?.length) parts.push(`下游 ${topology.children.length} 个`)
  if (topology.current.description) parts.push(topology.current.description)
  return parts
}

export function createReportViewModel(
  response: AlertEventReportResponse,
  context: ReportContext = {},
): ReportViewModel {
  const reportView = response.reportData.reportView
  const evidence = response.reportData.evidence || []
  const candidates = buildCandidates(response.reportData.rootCauseAnalysis.other || [])
  const rootCauseType = primaryRootCauseType(response.reportData.rootCauseAnalysis.other || [])
  const actions = buildActions(response.reportData.suggest || [], reportView)
  const businessImpact = buildBusinessImpact(reportView)
  const rootCauseEvidence = buildRootCauseEvidence(evidence, reportView, rootCauseType)
  const startTime = context.startTime || response.reportData.overview.timestamp
  const endTime = context.endTime || response.reportData.overview.timestamp

  return {
    id: context.id || '',
    reportType: response.reportType,
    typeLabel: reportTypeLabels[response.reportType] || '诊断报告',
    generatedAt: formatReportTime(response.reportData.overview.timestamp),
    analysisWindow: startTime && endTime ? `${formatReportTime(startTime)} - ${formatReportTime(endTime)}` : '',
    dataSourceLabel: context.dataSourceLabel || '报告快照',
    status: context.runtimeStatus || defaultReportStatus,
    headline: cleanText(reportView?.headline) || cleanText(response.reportData.overview.detail) || '未返回诊断标题',
    analysisSummary: cleanText(reportView?.conclusionSummary) || cleanText(response.reportData.overview.reason),
    rootCause: cleanText(response.reportData.rootCauseAnalysis.cause),
    overview: cleanText(response.reportData.overview.detail),
    reason: cleanText(response.reportData.overview.reason),
    tags: Object.entries(response.reportData.overview.tags || {}).map(([key, value]) => ({ key, value })),
    candidates,
    checkedCount: candidates.length,
    primaryCount: candidates.filter((item) => item.state === 'primary').length,
    secondaryCount: candidates.filter((item) => item.state === 'secondary').length,
    excludedCount: candidates.filter((item) => item.state === 'excluded').length,
    primaryAction: actions[0],
    actions,
    coreMetrics: buildCoreMetrics(rootCauseEvidence),
    rootCauseEvidence,
    evidenceSummary: buildEvidenceSummary(evidence, reportView),
    evidence,
    topology: response.reportData.topology,
    topologySummary: buildTopologySummary(response.reportData.topology),
    faultLocation: buildFaultLocation(reportView),
    businessImpact,
    impactGroups: buildImpactGroups(businessImpact),
    summaryCards: buildSummaryCards(response, reportView),
    timeline: buildTimeline(reportView),
    analyzeProcess: response.reportData.analyzeProcess,
    startTime: startTime || 0,
    endTime: endTime || 0,
  }
}

export function createReportViewModelFromViewResponse(
  response: AlertEventReportViewResponse,
  fallbackContext: Partial<ReportContext> = {},
): ReportViewModel {
  return createReportViewModel(response.snapshot, {
    ...fallbackContext,
    runtimeStatus: response.view?.runtimeStatus || fallbackContext.runtimeStatus || defaultReportStatus,
  })
}
