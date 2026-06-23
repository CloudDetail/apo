export const hasOverviewContent = ({ systems = [], pendingIssues = [], summary = {} }) =>
  systems.length > 0 || pendingIssues.length > 0 || Object.keys(summary || {}).length > 0

export const shouldRenderInitialOverviewLoading = ({
  overviewLoading,
  systems = [],
  pendingIssues = [],
  summary = {},
}) => Boolean(overviewLoading) && !hasOverviewContent({ systems, pendingIssues, summary })

export const isLatestOverviewRequest = (requestId, latestRequestId) => requestId === latestRequestId
