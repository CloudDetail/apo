export const hasOverviewContent = ({ systems = [], pendingIssues = [], summary = {} }) =>
  systems.length > 0 || pendingIssues.length > 0 || Object.keys(summary || {}).length > 0

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
    return { shouldLoad: false, nextLastRequestKey: requestKey }
  }
  return { shouldLoad: true, nextLastRequestKey: requestKey }
}
