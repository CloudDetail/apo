export const buildPendingIssueAnalysisKey = (issue = {}, params = {}) =>
  JSON.stringify([
    issue.groupId || '',
    issue.serviceName || '',
    issue.metricKey || '',
    issue.status || '',
    issue.value || '',
    issue.note || '',
    params.relativeTime || '',
    params.startTime || '',
    params.endTime || '',
  ])
