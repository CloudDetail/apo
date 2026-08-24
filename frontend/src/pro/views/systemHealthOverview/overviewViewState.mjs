export const hasOverviewContent = ({ systems = [], pendingIssues = [], summary = {} }) =>
  systems.length > 0 || pendingIssues.length > 0 || Object.keys(summary || {}).length > 0

const normalizedText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

export const overviewSystemEmptyMessage = ({
  systems = [],
  systemPageTotal = 0,
  systemHealthFilter = '',
  systemKeyword = '',
  debouncedSystemKeyword = '',
  allFilterValue = 'all',
} = {}) => {
  if (Array.isArray(systems) && systems.length > 0) return ''

  const currentFilter = normalizedText(systemHealthFilter)
  const hasSystemFilter = currentFilter !== '' && currentFilter !== allFilterValue
  const hasSystemKeyword =
    normalizedText(systemKeyword) !== '' || normalizedText(debouncedSystemKeyword) !== ''
  if (hasSystemFilter || hasSystemKeyword) return '没有匹配的系统'

  const total = Number(systemPageTotal)
  if (Number.isFinite(total) && total > 0) return '没有匹配的系统'

  return '暂无系统健康度数据'
}

export const overviewServiceEmptyMessage = ({
  services = [],
  servicePageTotal = 0,
  serviceHealthFilter = '',
  serviceKeyword = '',
  debouncedServiceKeyword = '',
  allFilterValue = 'all',
} = {}) => {
  if (Array.isArray(services) && services.length > 0) return ''

  const currentFilter = normalizedText(serviceHealthFilter)
  const hasServiceFilter = currentFilter !== '' && currentFilter !== allFilterValue
  const hasServiceKeyword =
    normalizedText(serviceKeyword) !== '' || normalizedText(debouncedServiceKeyword) !== ''
  if (hasServiceFilter || hasServiceKeyword) return '没有匹配的服务'

  const total = Number(servicePageTotal)
  if (Number.isFinite(total) && total > 0) return '没有匹配的服务'

  return '暂无服务健康度数据'
}

export const shouldRenderInitialOverviewLoading = ({ overviewLoading }) => Boolean(overviewLoading)

export const isLatestOverviewRequest = (requestId, latestRequestId) => requestId === latestRequestId

export const buildOverviewRequestDecision = ({
  requestKey,
  lastRequestKey = '',
  overviewLoading,
}) => {
  if (!requestKey) {
    return { shouldLoad: false, nextLastRequestKey: lastRequestKey }
  }
  if (lastRequestKey === requestKey) {
    return { shouldLoad: false, nextLastRequestKey: lastRequestKey }
  }
  if (overviewLoading) {
    return { shouldLoad: false, nextLastRequestKey: lastRequestKey }
  }
  return { shouldLoad: true, nextLastRequestKey: requestKey }
}
