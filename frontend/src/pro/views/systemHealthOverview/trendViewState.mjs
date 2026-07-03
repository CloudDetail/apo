import {
  DEFAULT_HEALTH_SCORE_THRESHOLDS,
  normalizeHealthScoreThresholds,
} from './metricVisuals.mjs'

const safeNumber = (value, fallback = 0) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : fallback
}

const statLastValue = (stat = {}) => safeNumber(stat.last)

const statAverageValue = (stat = {}) => {
  const avg = Number(stat.avg)
  if (Number.isFinite(avg)) return avg
  return statLastValue(stat)
}

const detailServiceDistributionMode = (granularity = '') =>
  granularity && granularity !== 'hour' ? 'average' : 'snapshot'

const emptyTrendStats = () => ({
  avg: null,
  peak: null,
  min: null,
  last: null,
  ratio: 0,
  sampleCount: 0,
})

export const DEFAULT_TREND_RANGE = 'last7d'
export const SYSTEM_HEALTH_TREND_RANGE_STORAGE_KEY = 'apo.systemHealth.trendRange'

export const trendRangeOptions = [
  { key: 'today', label: '今天' },
  { key: 'last24h', label: '近24小时' },
  { key: 'thisWeek', label: '本周' },
  { key: 'last7d', label: '过去7天' },
  { key: 'last15d', label: '过去15天' },
  { key: 'last30d', label: '过去一个月' },
  { key: 'last6mo', label: '过去半年' },
  { key: 'last1y', label: '过去一年' },
]

const supportedTrendRanges = new Set(trendRangeOptions.map((option) => option.key))

export const normalizeTrendRange = (range, fallback = DEFAULT_TREND_RANGE) => {
  const nextRange = typeof range === 'string' ? range : ''
  const nextFallback = supportedTrendRanges.has(fallback) ? fallback : DEFAULT_TREND_RANGE
  return supportedTrendRanges.has(nextRange) ? nextRange : nextFallback
}

const browserLocalStorage = () => {
  try {
    return globalThis?.localStorage || null
  } catch (error) {
    return null
  }
}

export const readStoredTrendRange = (storage = browserLocalStorage()) => {
  if (!storage || typeof storage.getItem !== 'function') return DEFAULT_TREND_RANGE
  try {
    return normalizeTrendRange(storage.getItem(SYSTEM_HEALTH_TREND_RANGE_STORAGE_KEY))
  } catch (error) {
    return DEFAULT_TREND_RANGE
  }
}

export const writeStoredTrendRange = (range, storage = browserLocalStorage()) => {
  const nextRange = normalizeTrendRange(range)
  if (!storage || typeof storage.setItem !== 'function') return nextRange
  try {
    storage.setItem(SYSTEM_HEALTH_TREND_RANGE_STORAGE_KEY, nextRange)
  } catch (error) {
    // Ignore storage failures so trend selection still updates in memory.
  }
  return nextRange
}

export const trendMetricGroups = [
  {
    key: 'abnormalSystems',
    label: '异常系统',
    chartLabel: '异常系统趋势图',
    tone: 'critical',
    metrics: [
      {
        key: 'abnormalSystems',
        label: '数量趋势图',
        seriesLabel: '数量',
        tone: 'critical',
        primary: 'peak',
      },
      {
        key: 'abnormalSystemRatio',
        label: '占比趋势图',
        seriesLabel: '占比',
        tone: 'critical',
        primary: 'peak',
        valueType: 'percent',
      },
    ],
  },
  {
    key: 'attentionSystems',
    label: '关注系统',
    chartLabel: '关注系统趋势图',
    tone: 'warning',
    metrics: [
      {
        key: 'attentionSystems',
        label: '数量趋势图',
        seriesLabel: '数量',
        tone: 'warning',
        primary: 'peak',
      },
      {
        key: 'attentionSystemRatio',
        label: '占比趋势图',
        seriesLabel: '占比',
        tone: 'warning',
        primary: 'peak',
        valueType: 'percent',
      },
    ],
  },
]

export const trendMetricConfigs = trendMetricGroups.flatMap((group) => group.metrics)

export const buildHealthScoreBaselineItems = (
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  const thresholds = normalizeHealthScoreThresholds(scoreThresholds)
  return [
    {
      key: 'attentionBaseline',
      label: '关注基线',
      value: thresholds.generalAbnormalScoreThreshold,
      tone: 'warning',
    },
    {
      key: 'abnormalBaseline',
      label: '异常基线',
      value: thresholds.severeAbnormalScoreThreshold,
      tone: 'critical',
    },
  ]
}

export const trendPrimaryValue = (point = {}, metricKey = '') => {
  if (point?.missingSnapshot) return null
  const stat = point[metricKey] || {}
  if (metricKey === 'systemHealthScore') {
    const sampleCount = Number(stat.sampleCount)
    if (Number.isFinite(sampleCount) && sampleCount <= 0) return null
    const min = Number(stat.min)
    return Number.isFinite(min) ? min : null
  }
  return safeNumber(stat.peak)
}

export const trendTooltipStats = (point = {}, metricKey = '') => {
  if (point?.missingSnapshot) return emptyTrendStats()
  const stat = point[metricKey] || {}
  if (metricKey === 'systemHealthScore') {
    const sampleCount = safeNumber(stat.sampleCount)
    if (sampleCount <= 0) {
      return {
        avg: null,
        min: null,
        last: null,
        ratio: 0,
        sampleCount: 0,
      }
    }
    return {
      avg: safeNumber(stat.avg),
      min: safeNumber(stat.min),
      last: safeNumber(stat.last),
      ratio: safeNumber(stat.belowThresholdRatio),
      sampleCount,
    }
  }
  return {
    avg: safeNumber(stat.avg),
    peak: safeNumber(stat.peak),
    min: safeNumber(stat.min),
    last: safeNumber(stat.last),
    ratio: safeNumber(stat.activeRatio),
    sampleCount: safeNumber(stat.sampleCount),
  }
}

export const buildTrendChartRows = (points = [], metricKey = '') =>
  (Array.isArray(points) ? points : [])
    .map((point) => {
      const timestamp = safeNumber(point?.timestamp)
      if (!timestamp) return null
      return {
        timestamp,
        ...(point?.missingSnapshot ? { missingSnapshot: true } : {}),
        value: trendPrimaryValue(point, metricKey),
        stats: trendTooltipStats(point, metricKey),
      }
    })
    .filter(Boolean)
    .sort((left, right) => left.timestamp - right.timestamp)

export const buildDetailTrendRows = (points = [], granularity = '') =>
  (Array.isArray(points) ? points : [])
    .map((point) => {
      const timestamp = safeNumber(point?.timestamp)
      if (!timestamp) return null
      const serviceDistributionMode = detailServiceDistributionMode(granularity)
      const serviceStatValue =
        serviceDistributionMode === 'average' ? statAverageValue : statLastValue

      if (point?.missingSnapshot) {
        return {
          timestamp,
          missingSnapshot: true,
          score: null,
          serviceCounts: { abnormal: null, attention: null, healthy: null, noData: null },
          serviceRatios: { abnormal: null, attention: null, healthy: null, noData: null },
        }
      }

      const serviceCounts = {
        abnormal: serviceStatValue(point?.abnormalServices),
        attention: serviceStatValue(point?.attentionServices),
        healthy: serviceStatValue(point?.healthyServices),
        noData: serviceStatValue(point?.noDataServices),
      }
      const serviceTotal =
        serviceCounts.abnormal +
        serviceCounts.attention +
        serviceCounts.healthy +
        serviceCounts.noData
      const serviceRatios =
        serviceTotal > 0
          ? {
              abnormal: serviceCounts.abnormal / serviceTotal,
              attention: serviceCounts.attention / serviceTotal,
              healthy: serviceCounts.healthy / serviceTotal,
              noData: serviceCounts.noData / serviceTotal,
            }
          : { abnormal: 0, attention: 0, healthy: 0, noData: 0 }

      return {
        timestamp,
        score: trendPrimaryValue(point, 'systemHealthScore'),
        serviceDistributionMode,
        serviceCounts,
        serviceRatios,
      }
    })
    .filter(Boolean)
    .sort((left, right) => left.timestamp - right.timestamp)

export const isLongTrendGranularity = (granularity = '') =>
  granularity === 'day' || granularity === 'week'

const DAY_INTERVAL_MS = 24 * 60 * 60 * 1000

export const trendAxisUsesDateOnly = (rows = [], granularity = '') => {
  if (isLongTrendGranularity(granularity)) return true

  const timestamps = (Array.isArray(rows) ? rows : [])
    .map((row) => safeNumber(row?.timestamp, Number.NaN))
    .filter(Number.isFinite)
    .sort((left, right) => left - right)

  if (timestamps.length < 2) return false

  let minInterval = Number.POSITIVE_INFINITY
  for (let index = 1; index < timestamps.length; index += 1) {
    const interval = timestamps[index] - timestamps[index - 1]
    if (interval > 0) {
      minInterval = Math.min(minInterval, interval)
    }
  }

  return Number.isFinite(minInterval) && minInterval >= DAY_INTERVAL_MS
}
