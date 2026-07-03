import {
  DEFAULT_HEALTH_SCORE_THRESHOLDS,
  normalizeHealthScoreThresholds,
} from './metricVisuals.mjs'

export const DEFAULT_HEALTH_FILTER_VALUES = {
  abnormal: 'abnormal',
  attention: 'attention',
  healthy: 'healthy',
  noData: 'noData',
  all: 'all',
}

export const DEFAULT_HEALTH_FILTER_VALUE = DEFAULT_HEALTH_FILTER_VALUES.abnormal

export const buildDetailHeaderRequestKey = (selectedSystemId, queryParams = {}) => {
  if (!selectedSystemId) return ''
  return JSON.stringify([
    `${selectedSystemId}`,
    queryParams.relativeTime || '',
    queryParams.startTime || '',
    queryParams.endTime || '',
  ])
}

export const shouldSyncDetailHeaderFromRows = (
  selectedSystemFromRows,
  systemRowsRequestKey,
  overviewRequestKey,
) =>
  Boolean(
    selectedSystemFromRows &&
      systemRowsRequestKey &&
      overviewRequestKey &&
      systemRowsRequestKey === overviewRequestKey,
  )

export const buildAbnormalServiceRatio = (
  services,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  const thresholds = normalizeHealthScoreThresholds(scoreThresholds)
  const serviceList = Array.isArray(services) ? services : []
  const abnormal = serviceList.filter((service) => {
    if (!hasHealthScoreData(service)) return false
    const score = Number(service?.score)
    return Number.isFinite(score) && score < thresholds.severeAbnormalScoreThreshold
  }).length
  const attention = serviceList.filter((service) => {
    if (!hasHealthScoreData(service)) return false
    const score = Number(service?.score)
    return (
      Number.isFinite(score) &&
      score >= thresholds.severeAbnormalScoreThreshold &&
      score < thresholds.generalAbnormalScoreThreshold
    )
  }).length

  return {
    abnormal,
    attention,
    total: serviceList.length,
  }
}

export const hasHealthScoreData = (target) => {
  if (target?.status === 'gray') return false
  const score = Number(target?.score)
  return Number.isFinite(score)
}

export const targetMatchesHealthFilter = (
  target,
  selectedFilters,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  const filters = Array.isArray(selectedFilters)
    ? selectedFilters
    : selectedFilters
      ? [selectedFilters]
      : []
  if (!filters.length) return true
  if (filters.includes(DEFAULT_HEALTH_FILTER_VALUES.all)) return true

  if (!hasHealthScoreData(target)) {
    return filters.includes(DEFAULT_HEALTH_FILTER_VALUES.noData)
  }

  const score = Number(target?.score)

  const thresholds = normalizeHealthScoreThresholds(scoreThresholds)
  if (score < thresholds.severeAbnormalScoreThreshold) {
    return filters.includes(DEFAULT_HEALTH_FILTER_VALUES.abnormal)
  }
  if (score < thresholds.generalAbnormalScoreThreshold) {
    return filters.includes(DEFAULT_HEALTH_FILTER_VALUES.attention)
  }
  return filters.includes(DEFAULT_HEALTH_FILTER_VALUES.healthy)
}

export const serviceMatchesAbnormalAttentionFilter = (
  service,
  onlyAbnormalAttention,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  if (!onlyAbnormalAttention) return true
  return targetMatchesHealthFilter(
    service,
    [DEFAULT_HEALTH_FILTER_VALUES.abnormal, DEFAULT_HEALTH_FILTER_VALUES.attention],
    scoreThresholds,
  )
}

export const compareHealthScoreAscending = (left, right) => {
  const leftNoData = !hasHealthScoreData(left)
  const rightNoData = !hasHealthScoreData(right)
  if (leftNoData !== rightNoData) return leftNoData ? 1 : -1
  if (leftNoData && rightNoData) return 0
  const leftScore = Number(left?.score)
  const rightScore = Number(right?.score)

  return leftScore - rightScore
}

export const shouldRenderDetailBody = ({ selectedSystem, detailHeaderLoading }) =>
  Boolean(selectedSystem) && !detailHeaderLoading
