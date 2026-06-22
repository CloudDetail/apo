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
  resource: '数据来源：Pod 资源水位',
  resourceShift: '数据来源：Pod / 服务资源指标，对比 24h 基线',
  errorRate: '数据来源：Trace 请求指标，系统错误请求数 / 总请求数',
  latency: '数据来源：Trace 请求指标，系统总耗时 / 总请求数',
  errorLog: '数据来源：日志指标，error / critical / exception 总数',
  badRequests: '数据来源：告警事件，当前 firing 告警数',
  lockCost: '数据来源：eBPF 锁等待指标，按请求数均摊',
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

const compactResourceNote = (note, status) => {
  const resourceType = resourceTypeFromNote(note)
  const resourceLabel = resourceType === 'CPU' ? ' CPU ' : resourceType
  if (status === 'gray') return '暂无资源水位数据'
  if (status === 'green') return `当前${resourceLabel}水位正常`
  return `当前${resourceLabel}水位偏高`
}

const compactMetricNote = (key, note, status) => {
  if (key === 'resource') return compactResourceNote(note, status)
  return safeText(note) || '暂无指标说明'
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
  return `${title}，最高水位 Pod：${sourceName}`
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

const metricStatusLeaders = (systems = [], metricDefinitions = [], statuses = []) => {
  const statusSet = new Set(statuses)
  return metricDefinitions
    .map(([key, label], index) => ({
      key,
      label,
      index,
      count: systems.reduce(
        (total, system) => (statusSet.has(system?.metrics?.[key]) ? total + 1 : total),
        0,
      ),
    }))
    .filter((item) => item.count > 0)
    .sort((left, right) => {
      if (left.count !== right.count) return right.count - left.count
      return left.index - right.index
    })
}

const severeSystemLeaders = (systems = []) =>
  systems
    .map((system, index) => ({
      name: safeText(system?.name),
      index,
      count: Object.values(system?.metrics || {}).filter((status) => status === 'red').length,
    }))
    .filter((item) => item.name && item.count > 0)
    .sort((left, right) => {
      if (left.count !== right.count) return right.count - left.count
      return left.index - right.index
    })

export const buildOverviewSummaryCards = (summary = {}, systems = [], metricDefinitions = []) => {
  const totalSystems = summaryNumber(summary.totalSystems) || systems.length
  const abnormalSystems = summaryNumber(summary.abnormalSystems)
  const lowScoreSystems = summaryNumber(summary.lowScoreSystems)
  const abnormalMetrics = summaryNumber(summary.abnormalMetrics)
  const severeMetrics = summaryNumber(summary.severeMetrics)
  const newAbnormalMetrics = summaryNumber(
    summary.newAbnormalMetrics ?? summary.recentAbnormalMetrics,
  )
  const abnormalMetricLabels = metricStatusLeaders(systems, metricDefinitions, ['red', 'yellow'])
    .slice(0, 3)
    .map((item) => item.label)
  const severeSystemNames = severeSystemLeaders(systems)
    .slice(0, 3)
    .map((item) => item.name)

  return [
    {
      key: 'abnormalSystems',
      label: '异常系统',
      value: `${abnormalSystems} / ${totalSystems}`,
      caption: `${lowScoreSystems} 个系统健康度低于 60`,
      percent: safePercent(abnormalSystems, totalSystems),
      tone: 'critical',
      visual: 'ring',
    },
    {
      key: 'abnormalMetrics',
      label: '异常指标',
      value: abnormalMetrics,
      caption: abnormalMetricLabels.length
        ? `${abnormalMetricLabels.join('、')}占比最高`
        : '暂无异常指标',
      tone: 'warning',
      visual: 'number',
    },
    {
      key: 'severeMetrics',
      label: '严重异常',
      value: severeMetrics,
      caption: severeSystemNames.length ? `集中在${severeSystemNames.join('、')}` : '暂无红灯指标',
      tone: 'critical',
      visual: 'number',
    },
    {
      key: 'newAbnormalMetrics',
      label: '新增异常',
      value: newAbnormalMetrics,
      caption: '最近 10 分钟新增',
      tone: newAbnormalMetrics > 0 ? 'warning' : 'neutral',
      visual: 'number',
    },
  ]
}

const effectiveSystemStatus = (system = {}) => {
  if (!Array.isArray(system.services) || system.services.length === 0 || system.status === 'gray') {
    return 'gray'
  }
  if (Number(system.score) < 60 || hasStatus(system.metrics, 'red')) return 'red'
  if (Number(system.score) < 80 || hasStatus(system.metrics, 'yellow')) return 'yellow'
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
  const base = {
    status,
    tone: statusTone(status),
    value: safeText(detail.value) || '--',
    note: compactMetricNote(key, detail.note, status),
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
    return {
      kind: 'waterline',
      ...base,
      percent: percentFromValue(base.value) ?? statusRatio(status),
      helper: '当前水位',
      marker: '90%',
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
      marker: '60%',
    }
  }

  if (key === 'errorLog') {
    return {
      kind: 'counter',
      ...base,
      ...(chart ? { chart } : {}),
      percent: statusRatio(status),
      helper: '窗口计数',
      marker: '严重度',
    }
  }

  if (key === 'badRequests') {
    return {
      kind: 'number',
      ...base,
      helper: '活动告警',
      marker: '当前值',
    }
  }

  if (key === 'latency') {
    return {
      kind: 'gauge',
      ...base,
      ...(chart ? { chart } : {}),
      percent: statusRatio(status),
      helper: '延时风险',
      marker: '状态',
    }
  }

  if (key === 'lockCost') {
    return {
      kind: 'number',
      ...base,
      helper: '锁等待均摊',
      marker: '当前值',
    }
  }

  return {
    kind: 'gauge',
    ...base,
    ...(chart ? { chart } : {}),
    percent: statusRatio(status),
    helper: '状态风险',
    marker: '状态',
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

  if (thresholds.length) {
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
