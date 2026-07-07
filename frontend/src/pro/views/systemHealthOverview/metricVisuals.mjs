const STATUS_RATIOS = {
  red: 92,
  yellow: 64,
  green: 24,
  gray: 0,
}

const STATUS_TONES = {
  red: 'critical',
  yellow: 'warning',
  green: 'stable',
  gray: 'muted',
}

export const HEALTH_SCORE_CRITICAL_THRESHOLD = 60
export const HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD = 80
export const DEFAULT_ABNORMAL_SCORE_THRESHOLD = 80
export const DEFAULT_HEALTH_SCORE_THRESHOLDS = {
  severeAbnormalScoreThreshold: HEALTH_SCORE_CRITICAL_THRESHOLD,
  generalAbnormalScoreThreshold: HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD,
}

const normalizeScoreThreshold = (value, fallback) => {
  const number = Number(value)
  if (!Number.isFinite(number)) return fallback
  const threshold = Math.trunc(number)
  return threshold
}

export const normalizeHealthScoreThresholds = (config = DEFAULT_HEALTH_SCORE_THRESHOLDS) => {
  const source = config && typeof config === 'object' ? config : {}
  const hasSevere = source.severeAbnormalScoreThreshold !== undefined
  const hasGeneral =
    source.generalAbnormalScoreThreshold !== undefined ||
    source.abnormalScoreThreshold !== undefined ||
    typeof config === 'number'
  const generalSource =
    typeof config === 'number'
      ? config
      : source.generalAbnormalScoreThreshold ?? source.abnormalScoreThreshold
  const severeSource = source.severeAbnormalScoreThreshold
  const generalThreshold = normalizeScoreThreshold(
    generalSource,
    DEFAULT_HEALTH_SCORE_THRESHOLDS.generalAbnormalScoreThreshold,
  )
  let severeThreshold = normalizeScoreThreshold(
    severeSource,
    DEFAULT_HEALTH_SCORE_THRESHOLDS.severeAbnormalScoreThreshold,
  )

  if (
    generalThreshold < 1 ||
    generalThreshold > 100 ||
    severeThreshold < 1 ||
    severeThreshold >= 100
  ) {
    return { ...DEFAULT_HEALTH_SCORE_THRESHOLDS }
  }
  if (severeThreshold >= generalThreshold) {
    if (hasSevere && hasGeneral) {
      return { ...DEFAULT_HEALTH_SCORE_THRESHOLDS }
    }
    severeThreshold = Math.max(1, generalThreshold - 1)
  }
  return {
    severeAbnormalScoreThreshold: severeThreshold,
    generalAbnormalScoreThreshold: generalThreshold,
  }
}

export const normalizeAbnormalScoreThreshold = (value) =>
  normalizeHealthScoreThresholds(value).generalAbnormalScoreThreshold

export const abnormalSystemsHelpText = (config = DEFAULT_HEALTH_SCORE_THRESHOLDS) => {
  const thresholds = normalizeHealthScoreThresholds(config)
  return `系统健康度低于 ${thresholds.severeAbnormalScoreThreshold} 分，计为异常系统。`
}

export const attentionSystemsHelpText = (config = DEFAULT_HEALTH_SCORE_THRESHOLDS) => {
  const thresholds = normalizeHealthScoreThresholds(config)
  return `系统健康度 ${thresholds.severeAbnormalScoreThreshold}-${thresholds.generalAbnormalScoreThreshold - 1} 分之间，计为关注系统。`
}

export const healthScoreStatus = (score, config = DEFAULT_HEALTH_SCORE_THRESHOLDS) => {
  const number = Number(score)
  if (!Number.isFinite(number)) return 'gray'
  const thresholds = normalizeHealthScoreThresholds(config)
  if (number < thresholds.severeAbnormalScoreThreshold) return 'red'
  if (number < thresholds.generalAbnormalScoreThreshold) return 'yellow'
  return 'green'
}

export const healthScoreLegendItemsFor = (config = DEFAULT_HEALTH_SCORE_THRESHOLDS) => {
  const thresholds = normalizeHealthScoreThresholds(config)
  return [
    { status: 'green', range: `${thresholds.generalAbnormalScoreThreshold}-100分`, label: '健康' },
    {
      status: 'yellow',
      range: `${thresholds.severeAbnormalScoreThreshold}-${thresholds.generalAbnormalScoreThreshold - 1}分`,
      label: '关注（关注系统）',
    },
    {
      status: 'red',
      range: `0-${thresholds.severeAbnormalScoreThreshold - 1}分`,
      label: '异常（异常系统）',
    },
    { status: 'gray', range: '无可评分数据', label: '无数据' },
  ]
}

export const healthScoreLegendItems = healthScoreLegendItemsFor(DEFAULT_HEALTH_SCORE_THRESHOLDS)

const VALID_STATUSES = new Set(['red', 'yellow', 'green', 'gray'])
const TREND_COLOR_ROLES = {
  resourceShift: 'resourceShift',
  errorRate: 'errorRate',
  latency: 'latency',
  errorLog: 'logs',
}
const TREND_UNITS = {
  resourceShift: '%',
  errorRate: '%',
  latency: 'ms',
  errorLog: 'count',
}
const TREND_MARK_LINE_DISABLED_KEYS = new Set(['resourceShift', 'errorRate', 'latency'])
const AGGREGATE_RESOURCE_WATERLINE_NOTE = '系统资源水位聚合'
const AGGREGATE_RESOURCE_WATERLINE_BADGE_LABELS = new Set(['平均水位', 'P95水位'])
const METRIC_SOURCE_TITLES = {
  resource: '数据组内服务关联 Pod 或 Host 去重后的整体资源水位。',
  resourceShift: '数据组整体资源水位较 24 小时前同窗口的变化幅度。',
  errorRate: '数据组内总失败请求数占总请求数的比例。',
  latency: '数据组整体请求延时，当前使用请求数加权平均值。',
  errorLog: '数据组内可见服务 error、critical 和 exception 日志总数。',
  badRequests: '数据组内当前 firing 告警按级别汇总后的数量。',
  lockCost: '数据组内 futex 锁等待总耗时按总请求数均摊后的耗时。',
}

const clamp = (value, min = 0, max = 100) => Math.min(max, Math.max(min, value))

const normalizeStatus = (status) => (VALID_STATUSES.has(status) ? status : 'gray')

const safeText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

const firstNumber = (value) => {
  const match = safeText(value).match(/[+-]?\d+(?:\.\d+)?/)
  if (!match) return null
  const number = Number(match[0])
  return Number.isFinite(number) ? number : null
}

const percentFromValue = (value) => {
  if (!safeText(value).includes('%')) return null
  const number = firstNumber(value)
  if (number === null) return null
  return clamp(Math.round(number))
}

const sourceObjectName = (source) => {
  const parts = safeText(source).split('/').filter(Boolean)
  if (!parts.length) return ''
  return parts[parts.length - 1]
}

const resourceTypeFromNote = (note) => {
  const text = safeText(note)
  if (text.includes('内存')) return '内存'
  if (text.toUpperCase().includes('CPU')) return 'CPU'
  return '资源'
}

const compactResourceNote = (note, status, value) => {
  const text = safeText(note)
  if (text.includes(AGGREGATE_RESOURCE_WATERLINE_NOTE)) {
    if (status === 'gray') return '暂无资源水位数据'
    if (percentFromValue(value) === null) return text || '暂无资源水位数据'
    if (status === 'green') return '系统判定水位正常'
    return '系统判定水位过高'
  }
  const resourceType = resourceTypeFromNote(note)
  const resourceLabel = resourceType === 'CPU' ? ' CPU ' : resourceType
  if (status === 'gray') return '暂无资源水位数据'
  if (percentFromValue(value) === null) return safeText(note) || '暂无资源水位数据'
  if (status === 'green') return `当前${resourceLabel}水位正常`
  return `当前${resourceLabel}水位偏高`
}

const compactMetricNote = (key, note, status, value) => {
  if (key === 'resource') return compactResourceNote(note, status, value)
  return safeText(note) || '暂无指标说明'
}

const metricDisplayValue = (value) => {
  const text = safeText(value)
  return !text || text === '--' ? '暂无数据' : text
}

const compactLargeMillisecondsPerRequestValue = (value) => {
  const text = metricDisplayValue(value)
  const match = text.match(/^([+-]?\d+(?:\.\d+)?)\s+(ms\/请求)$/)
  if (!match) return text

  const number = Number(match[1])
  if (!Number.isFinite(number) || Math.abs(number) < 1000) return text
  return `${Math.round(number)} ${match[2]}`
}

const countDisplayValue = (value) => {
  const text = metricDisplayValue(value)
  return text === '暂无数据' || /\D/.test(text) ? text : `${text} 个`
}

const logCountDisplayValue = (value) => {
  const text = metricDisplayValue(value)
  return text === '暂无数据' || /\D/.test(text) ? text : `${text} 条`
}

const trendPointValue = (point) => {
  if (point?.missingSnapshot || point?.value === undefined || point?.value === null) return null
  const value = Number(point.value)
  return Number.isFinite(value) ? value : null
}

const normalizeSeries = (series) => {
  if (!Array.isArray(series)) return []
  return series
    .map((point) => {
      const timestamp = Number(point?.timestamp)
      if (!Number.isFinite(timestamp)) return null
      return { timestamp, value: trendPointValue(point) }
    })
    .filter(Boolean)
    .sort((left, right) => left.timestamp - right.timestamp)
}

const normalizeThresholds = (thresholds) => {
  if (!Array.isArray(thresholds)) return []
  return thresholds
    .map((threshold) => {
      const value = Number(threshold?.value)
      if (!Number.isFinite(value)) return null
      return {
        label: safeText(threshold?.label) || `${value}`,
        value,
        status: normalizeStatus(threshold?.status),
      }
    })
    .filter(Boolean)
}

const normalizeBadges = (badges) =>
  (Array.isArray(badges) ? badges : [])
    .map((badge) => {
      const label = safeText(badge?.label)
      const value = safeText(badge?.value)
      if (!label || !value) return null
      return {
        label,
        value,
        status: normalizeStatus(badge?.status),
      }
    })
    .filter(Boolean)

const isAggregateResourceWaterline = (detail = {}, badges = []) =>
  safeText(detail.note).includes(AGGREGATE_RESOURCE_WATERLINE_NOTE) ||
  badges.some((badge) => AGGREGATE_RESOURCE_WATERLINE_BADGE_LABELS.has(badge.label))

const metricChartFor = (key, detail = {}) => {
  const colorRole = TREND_COLOR_ROLES[key]
  const series = normalizeSeries(detail.series)
  if (!colorRole || !series.length) return undefined
  return {
    key,
    colorRole,
    unit: safeText(detail.seriesUnit) || safeText(detail.unit) || TREND_UNITS[key] || '',
    label: safeText(detail.seriesLabel),
    series,
    thresholds: normalizeThresholds(detail.thresholds),
  }
}

const formatCompactNumber = (value) => {
  const number = Number(value)
  if (!Number.isFinite(number)) return '--'
  if (Math.abs(number) >= 100) return `${Math.round(number)}`
  if (Math.abs(number) >= 10) return number.toFixed(1).replace(/\.0$/, '')
  return number.toFixed(2).replace(/\.?0+$/, '')
}

const formatAxisValue = (value, unit = '') => {
  const compact = formatCompactNumber(value)
  if (!unit || compact === '--') return compact
  if (unit === 'count') return compact
  return `${compact}${unit}`
}

const formatTimeLabel = (timestamp) => {
  const number = Number(timestamp)
  if (!Number.isFinite(number)) return ''
  const milliseconds = number > 1e12 ? Math.round(number / 1000) : number
  const date = new Date(milliseconds)
  if (Number.isNaN(date.getTime())) return ''
  return `${`${date.getHours()}`.padStart(2, '0')}:${`${date.getMinutes()}`.padStart(2, '0')}`
}

const thresholdColor = (status, colors = {}) => {
  if (status === 'red') return colors.criticalColor || '#f04438'
  if (status === 'yellow') return colors.warningColor || '#f79009'
  return colors.gridColor || '#98a2b3'
}

const seriesHasTrendData = (series) =>
  (Array.isArray(series) ? series : []).some((point) => Number.isFinite(point?.value))

const noDataTrendGraphic = () => ({
  type: 'text',
  left: 'center',
  top: 'middle',
  silent: true,
  style: {
    text: '无数据',
    fill: '#98a2b3',
    fontSize: 12,
    fontWeight: 600,
  },
})

export const metricSourceTitleFor = (key, detail = {}) => {
  const title = METRIC_SOURCE_TITLES[key] || '数据来源：系统健康指标'
  if (key !== 'resource') return title
  const sourceName = sourceObjectName(detail.source)
  if (!sourceName) return title
  return `${title}最高水位 Pod：${sourceName}`
}

export const safePercent = (numerator, denominator) => {
  const top = Number(numerator)
  const bottom = Number(denominator)
  if (!Number.isFinite(top) || !Number.isFinite(bottom) || bottom <= 0) return 0
  return clamp(Math.round((top / bottom) * 100))
}

export const statusRatio = (status) => STATUS_RATIOS[normalizeStatus(status)] || 0

const statusTone = (status) => STATUS_TONES[normalizeStatus(status)] || 'muted'

const summaryNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

const effectiveSystemStatus = (system = {}, scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS) => {
  if (!Array.isArray(system.services) || system.services.length === 0 || system.status === 'gray') {
    return 'gray'
  }
  return healthScoreStatus(system.score, scoreThresholds)
}

const countSystemsByScoreStatus = (
  systems = [],
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  const counts = { red: 0, yellow: 0, green: 0, gray: 0 }
  if (!Array.isArray(systems)) return counts

  systems.forEach((system) => {
    const status = effectiveSystemStatus(system, scoreThresholds)
    counts[status] += 1
  })

  return counts
}

export const buildOverviewSummaryCards = (
  summary = {},
  systems = [],
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
  options = {},
) => {
  const totalSystems = summaryNumber(summary.totalSystems) || systems.length
  const thresholds = normalizeHealthScoreThresholds(scoreThresholds)
  const scoreStatusCounts = countSystemsByScoreStatus(systems, thresholds)
  const hasSystems = Array.isArray(systems) && systems.length > 0
  const severeSystems = hasSystems ? scoreStatusCounts.red : summaryNumber(summary.lowScoreSystems)
  const attentionSystems = hasSystems
    ? scoreStatusCounts.yellow
    : summaryNumber(
        summary.attentionSystems ??
          summary.warningSystems ??
          summary.yellowSystems ??
          Math.max(
            summaryNumber(summary.abnormalSystems) - summaryNumber(summary.lowScoreSystems),
            0,
          ),
      )
  const pendingIssueTotal = summaryNumber(options.pendingIssueTotal ?? summary.pendingIssues)

  return [
    {
      key: 'abnormalSystems',
      label: '异常系统',
      help: abnormalSystemsHelpText(thresholds),
      value: `${severeSystems} / ${totalSystems}`,
      numerator: severeSystems,
      denominator: totalSystems,
      highlightNumerator: true,
      tone: 'critical',
      visual: 'ratio',
    },
    {
      key: 'attentionSystems',
      label: '关注系统',
      help: attentionSystemsHelpText(thresholds),
      value: `${attentionSystems} / ${totalSystems}`,
      numerator: attentionSystems,
      denominator: totalSystems,
      highlightNumerator: true,
      tone: 'warning',
      visual: 'ratio',
    },
    {
      key: 'pendingIssues',
      label: '待处理问题',
      value: pendingIssueTotal,
      tone: 'critical',
      visual: 'number',
    },
  ]
}

export const buildStatusDistribution = (
  systems = [],
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  const distribution = {
    total: systems.length,
    red: 0,
    yellow: 0,
    green: 0,
    gray: 0,
  }
  systems.forEach((system) => {
    distribution[effectiveSystemStatus(system, scoreThresholds)] += 1
  })
  return distribution
}

export const metricVisualFor = (key, detail = {}, fallbackStatus = 'gray') => {
  const status = normalizeStatus(detail.status || fallbackStatus)
  const value = metricDisplayValue(detail.value)
  const base = {
    status,
    tone: statusTone(status),
    value,
    note: compactMetricNote(key, detail.note, status, value),
  }
  const source = safeText(detail.source)
  const chart = metricChartFor(key, detail)
  const badges = normalizeBadges(detail.badges)
  if (badges.length) {
    base.badges = badges
  }
  if (source) {
    base.source = source
  }

  if (key === 'resource') {
    if (Array.isArray(detail.series) && detail.series.length) {
      base.chart = null
    }
    const isAggregate = isAggregateResourceWaterline(detail, badges)
    const currentPercent = percentFromValue(base.value)
    return {
      kind: 'waterline',
      ...base,
      percent: currentPercent ?? 0,
      helper: isAggregate ? '判定水位' : '当前水位',
      showThresholdMarker: status !== 'green' && status !== 'gray',
    }
  }

  if (key === 'resourceShift') {
    const shiftPercent = firstNumber(base.value)
    return {
      kind: 'deviation',
      ...base,
      ...(chart ? { chart } : {}),
      percent: shiftPercent === null ? statusRatio(status) : safePercent(shiftPercent, 60),
      helper: '较 24h 前',
    }
  }

  if (key === 'errorLog') {
    return {
      kind: 'counter',
      ...base,
      ...(chart ? { chart } : {}),
      value: logCountDisplayValue(detail.value),
      percent: statusRatio(status),
      helper: '窗口计数',
    }
  }

  if (key === 'badRequests') {
    return {
      kind: 'number',
      ...base,
      value: countDisplayValue(detail.value),
      helper: '活跃告警',
    }
  }

  if (key === 'latency') {
    return {
      kind: 'gauge',
      ...base,
      ...(chart ? { chart } : {}),
      percent: statusRatio(status),
      helper: '延时风险',
    }
  }

  if (key === 'lockCost') {
    return {
      kind: 'number',
      ...base,
      value: compactLargeMillisecondsPerRequestValue(detail.value),
      helper: '锁等待均摊',
    }
  }

  return {
    kind: 'gauge',
    ...base,
    ...(chart ? { chart } : {}),
    percent: statusRatio(status),
    helper: '状态风险',
  }
}

export const buildMetricTrendOption = (chart = {}, colors = {}) => {
  const series = normalizeSeries(chart.series)
  const thresholds = normalizeThresholds(chart.thresholds)
  const unit = safeText(chart.unit)
  const lineColor = colors.color || '#4992ff'
  const optionSeries = {
    type: 'line',
    data: series.map((point) => point.value),
    smooth: true,
    showSymbol: false,
    connectNulls: false,
    lineStyle: {
      width: 2,
      color: lineColor,
    },
    areaStyle: {
      color: lineColor,
      opacity: 0.08,
    },
    emphasis: {
      focus: 'series',
    },
  }

  if (!TREND_MARK_LINE_DISABLED_KEYS.has(chart.key) && thresholds.length) {
    optionSeries.markLine = {
      symbol: 'none',
      silent: true,
      label: {
        color: colors.textColor || '#667085',
        formatter: ({ name }) => name,
        fontSize: 10,
      },
      lineStyle: {
        type: 'dashed',
        width: 1,
      },
      data: thresholds.map((threshold) => ({
        name: threshold.label,
        yAxis: threshold.value,
        lineStyle: {
          color: thresholdColor(threshold.status, colors),
        },
      })),
    }
  }

  return {
    animation: false,
    grid: {
      top: 8,
      right: 8,
      bottom: 4,
      left: 6,
      containLabel: false,
    },
    tooltip: {
      trigger: 'axis',
      confine: true,
      valueFormatter: (value) => formatAxisValue(value, unit),
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: series.map((point) => formatTimeLabel(point.timestamp)),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { show: false },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      min: 0,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: {
        show: false,
        formatter: (value) => formatAxisValue(value, unit),
        color: colors.textColor || '#667085',
        fontSize: 10,
        margin: 2,
      },
      splitLine: {
        show: false,
        lineStyle: {
          color: colors.gridColor || 'rgba(102, 112, 133, 0.16)',
          type: 'dashed',
        },
      },
    },
    series: [optionSeries],
    ...(seriesHasTrendData(series) ? {} : { graphic: noDataTrendGraphic() }),
  }
}
