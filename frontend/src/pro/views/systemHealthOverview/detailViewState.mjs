import {
  DEFAULT_HEALTH_SCORE_THRESHOLDS,
  normalizeHealthScoreThresholds,
} from './metricVisuals.mjs'

export const buildDetailRequestKey = (selectedSystemId, queryParams = {}) => {
  if (!selectedSystemId) return ''
  return JSON.stringify([
    `${selectedSystemId}`,
    queryParams.relativeTime || '',
    queryParams.startTime || '',
    queryParams.endTime || '',
  ])
}

export const buildAbnormalServiceRatio = (
  services,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  const thresholds = normalizeHealthScoreThresholds(scoreThresholds)
  const serviceList = Array.isArray(services) ? services : []
  const abnormal = serviceList.filter((service) => {
    const score = Number(service?.score)
    return Number.isFinite(score) && score < thresholds.severeAbnormalScoreThreshold
  }).length
  const attention = serviceList.filter((service) => {
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

export const serviceMatchesAbnormalAttentionFilter = (
  service,
  onlyAbnormalAttention,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  if (!onlyAbnormalAttention) return true
  const score = Number(service?.score)
  if (!Number.isFinite(score)) return false
  return score < normalizeHealthScoreThresholds(scoreThresholds).generalAbnormalScoreThreshold
}

export const shouldRenderDetailBody = ({ selectedSystem, detailLoading }) =>
  Boolean(selectedSystem) && !detailLoading
