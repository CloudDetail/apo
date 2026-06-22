const metricOrderKeys = [
  'resource',
  'resourceShift',
  'errorRate',
  'latency',
  'errorLog',
  'badRequests',
  'lockCost',
]

export const pendingIssueSeverityValue = (status) =>
  ({
    red: 4,
    yellow: 3,
    gray: 2,
    green: 1,
  })[status] || 0

const metricOrder = (metricKey) => {
  const index = metricOrderKeys.indexOf(metricKey)
  return index >= 0 ? index : metricOrderKeys.length
}

const textValue = (value) => (value === undefined || value === null ? '' : `${value}`)

const comparePendingIssues = (left = {}, right = {}) => {
  const severityDelta =
    pendingIssueSeverityValue(right.status) - pendingIssueSeverityValue(left.status)
  if (severityDelta !== 0) return severityDelta

  const metricDelta = metricOrder(left.metricKey) - metricOrder(right.metricKey)
  if (metricDelta !== 0) return metricDelta

  return (
    [
      textValue(left.systemName).localeCompare(textValue(right.systemName)),
      textValue(left.serviceName).localeCompare(textValue(right.serviceName)),
      textValue(left.metricLabel).localeCompare(textValue(right.metricLabel)),
    ].find((delta) => delta !== 0) || 0
  )
}

const uniqueValues = (issues = []) => {
  const values = []
  issues.forEach((issue) => {
    const value = textValue(issue.value).trim()
    if (value && value !== '--' && !values.includes(value)) {
      values.push(value)
    }
  })
  return values.length ? values.join('、') : '--'
}

const uniqueTexts = (issues = [], key) => {
  const values = []
  issues.forEach((issue) => {
    const value = textValue(issue[key]).trim()
    if (value && !values.includes(value)) {
      values.push(value)
    }
  })
  return values.join('、')
}

export const buildPendingIssueQueueRows = (issues = []) => {
  const groups = new Map()

  issues.forEach((issue) => {
    const queueId = [textValue(issue.serviceName), textValue(issue.metricKey)].join(':')

    if (!groups.has(queueId)) {
      groups.set(queueId, {
        queueId,
        issues: [],
      })
    }
    groups.get(queueId).issues.push(issue)
  })

  return Array.from(groups.values())
    .map((group) => {
      const sortedIssues = [...group.issues].sort(comparePendingIssues)
      const leadIssue = sortedIssues[0] || {}

      return {
        ...leadIssue,
        queueId: group.queueId,
        issueCount: group.issues.length,
        systemName: uniqueTexts(sortedIssues, 'systemName') || leadIssue.systemName,
        value: uniqueValues(sortedIssues),
      }
    })
    .sort((left, right) => {
      const severityDelta =
        pendingIssueSeverityValue(right.status) - pendingIssueSeverityValue(left.status)
      if (severityDelta !== 0) return severityDelta

      const metricDelta = metricOrder(left.metricKey) - metricOrder(right.metricKey)
      if (metricDelta !== 0) return metricDelta

      return (
        [
          textValue(left.systemName).localeCompare(textValue(right.systemName)),
          textValue(left.serviceName).localeCompare(textValue(right.serviceName)),
          textValue(left.metricLabel).localeCompare(textValue(right.metricLabel)),
        ].find((delta) => delta !== 0) || 0
      )
    })
}
