export const buildDetailRequestKey = (selectedSystemId, queryParams = {}) => {
  if (!selectedSystemId) return ''
  return JSON.stringify([
    `${selectedSystemId}`,
    queryParams.relativeTime || '',
    queryParams.startTime || '',
    queryParams.endTime || '',
  ])
}

export const shouldRenderDetailBody = ({ selectedSystem, detailLoading }) =>
  Boolean(selectedSystem) && !detailLoading
