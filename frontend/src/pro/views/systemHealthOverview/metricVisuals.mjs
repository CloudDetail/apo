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
export const abnormalSystemsHelpText = `健康度低于 ${HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD} 分，或存在关注/异常指标时计入异常系统。`

export const healthScoreLegendItems = [
  { status: 'green', range: `${HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD}-100分`, label: '正常' },
  {
    status: 'yellow',
    range: `${HEALTH_SCORE_CRITICAL_THRESHOLD}-${HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD - 1}分`,
    label: '关注',
  },
  { status: 'red', range: `0-${HEALTH_SCORE_CRITICAL_THRESHOLD - 1}分`, label: '异常' },
  { status: 'gray', range: '无可评分数据', label: '无数据' },
]

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
const METRIC_SOURCE_TITLES = {
  resource: '当前窗口内 CPU 或内存压力最高的 Pod 水位。',
  resourceShift: '当前窗口的 CPU 或内存用量，相比 24 小时前同一时间段的最大变化。',
  errorRate: '当前窗口内失败请求占全部请求的比例。',
  latency: '当前窗口内所有请求的平均响应时间。',
  errorLog: '当前窗口内 error、critical 和 exception 日志总数。',
  badRequests: '当前仍在触发中的告警数量。',
  lockCost: '当前窗口内每个请求平均分摊到的锁等待耗时。',
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

const countDisplayValue = (value) => {
  const text = metricDisplayValue(value)
  return text === '暂无数据' || /\D/.test(text) ? text : `${text} 个`
}

const logCountDisplayValue = (value) => {
  const text = metricDisplayValue(value)
  return text === '暂无数据' || /\D/.test(text) ? text : `${text} 条`
}

const normalizeSeries = (series) => {
  if (!Array.isArray(series)) return []
  return series
    .map((point) => {
      const timestamp = Number(point?.timestamp)
      const value = Number(point?.value)
      if (!Number.isFinite(timestamp) || !Number.isFinite(value)) return null
      return { timestamp, value }
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

const metricChartFor = (key, detail = {}) => {
  const colorRole = TREND_COLOR_ROLES[key]
  const series = normalizeSeries(detail.series)
  if (!colorRole || !series.length) return undefined
  return {
    key,
    colorRole,
    unit: safeText(detail.unit) || TREND_UNITS[key] || '',
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

const hasStatus = (metrics, status) => Object.values(metrics || {}).includes(status)

const summaryNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

export const buildOverviewSummaryCards = (summary = {}, systems = []) => {
  const totalSystems = summaryNumber(summary.totalSystems) || systems.length
  const abnormalSystems = summaryNumber(summary.abnormalSystems)
  const abnormalMetrics = summaryNumber(summary.abnormalMetrics)
  const severeMetrics = summaryNumber(summary.severeMetrics)
  const generalMetrics = summaryNumber(
    summary.generalMetrics ??
      summary.warningMetrics ??
      summary.yellowMetrics ??
      Math.max(abnormalMetrics - severeMetrics, 0),
  )

  return [
    {
      key: 'abnormalSystems',
      label: '异常系统',
      help: abnormalSystemsHelpText,
      value: `${abnormalSystems} / ${totalSystems}`,
      numerator: abnormalSystems,
      denominator: totalSystems,
      highlightNumerator: abnormalSystems > 0,
      tone: 'critical',
      visual: 'ratio',
    },
    {
      key: 'severeMetrics',
      label: '严重异常',
      value: severeMetrics,
      tone: 'critical',
      visual: 'number',
    },
    {
      key: 'generalMetrics',
      label: '一般异常',
      value: generalMetrics,
      tone: 'warning',
      visual: 'number',
    },
  ]
}

const effectiveSystemStatus = (system = {}) => {
  if (!Array.isArray(system.services) || system.services.length === 0 || system.status === 'gray') {
    return 'gray'
  }
  if (Number(system.score) < HEALTH_SCORE_CRITICAL_THRESHOLD || hasStatus(system.metrics, 'red')) {
    return 'red'
  }
  if (
    Number(system.score) < HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD ||
    hasStatus(system.metrics, 'yellow')
  ) {
    return 'yellow'
  }
  return 'green'
}

export const buildStatusDistribution = (systems = []) => {
  const distribution = {
    total: systems.length,
    red: 0,
    yellow: 0,
    green: 0,
    gray: 0,
  }
  systems.forEach((system) => {
    distribution[effectiveSystemStatus(system)] += 1
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
  if (source) {
    base.source = source
  }

  if (key === 'resource') {
    if (Array.isArray(detail.series) && detail.series.length) {
      base.chart = null
    }
    const currentPercent = percentFromValue(base.value)
    return {
      kind: 'waterline',
      ...base,
      percent: currentPercent ?? 0,
      helper: '当前水位',
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
      helper: '相对 24h 基线',
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
    connectNulls: true,
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

  if (chart.key !== 'resourceShift' && thresholds.length) {
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
  }
}
