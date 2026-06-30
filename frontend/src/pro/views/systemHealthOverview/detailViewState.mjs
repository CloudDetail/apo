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
    return Number.isFinite(score) && score < thresholds.generalAbnormalScoreThreshold
  }).length

  return {
    abnormal,
    total: serviceList.length,
  }
}

export const shouldRenderDetailBody = ({ selectedSystem, detailLoading }) =>
  Boolean(selectedSystem) && !detailLoading
