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

export const serviceAiAssistantActionLabel = (status) =>
  ({
    red: '异常分析',
    yellow: '关注调优',
    green: '持续保持',
    gray: '无数据',
  })[status] || '无数据'

export const serviceAiAssistantActionFor = (service = {}) => {
  const statuses = [service?.status, ...Object.values(service?.metrics || {})].filter(Boolean)

  if (statuses.includes('red')) {
    return { label: serviceAiAssistantActionLabel('red'), actionable: true, status: 'red' }
  }
  if (statuses.includes('yellow')) {
    return { label: serviceAiAssistantActionLabel('yellow'), actionable: true, status: 'yellow' }
  }
  if (statuses.length && statuses.every((status) => status === 'gray')) {
    return { label: serviceAiAssistantActionLabel('gray'), actionable: false, status: 'gray' }
  }
  if (statuses.includes('green')) {
    return { label: serviceAiAssistantActionLabel('green'), actionable: false, status: 'green' }
  }
  return { label: serviceAiAssistantActionLabel('gray'), actionable: false, status: 'gray' }
}

const metricIssueSeverity = (status) =>
  ({
    red: 2,
    yellow: 1,
  })[status] || 0

const textValue = (value, fallback = '') =>
  value === undefined || value === null || value === '' ? fallback : `${value}`

export const buildServiceHealthAssistantIssues = (
  service = {},
  selectedSystem = {},
  metricLabels = [],
) =>
  metricLabels
    .map(([metricKey, metricLabel], index) => {
      const status = service?.metrics?.[metricKey] || 'gray'
      const detail = service?.metricDetails?.[metricKey] || {}
      return {
        id: [textValue(selectedSystem?.groupId), textValue(service?.name), metricKey].join(':'),
        groupId: selectedSystem?.groupId,
        systemName: textValue(selectedSystem?.name, '未命名系统'),
        serviceName: textValue(service?.name, '未命名服务'),
        metricKey,
        metricLabel,
        status,
        value: textValue(detail.value, '--'),
        note: textValue(detail.note, '暂无指标说明'),
        serviceScore: Number.isFinite(Number(service?.score)) ? Number(service.score) : 0,
        systemScore: Number.isFinite(Number(selectedSystem?.score))
          ? Number(selectedSystem.score)
          : 0,
        metricIndex: index,
      }
    })
    .filter((issue) => issue.status === 'red' || issue.status === 'yellow')
    .sort((left, right) => {
      const severityDelta = metricIssueSeverity(right.status) - metricIssueSeverity(left.status)
      if (severityDelta !== 0) return severityDelta
      return left.metricIndex - right.metricIndex
    })
    .map(({ metricIndex, ...issue }) => issue)

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
