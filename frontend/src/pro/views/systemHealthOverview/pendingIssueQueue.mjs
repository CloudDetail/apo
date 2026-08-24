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

const pendingIssueSystemKey = (issue = {}) => {
  const groupId = textValue(issue.groupId).trim()
  if (groupId) return `group:${groupId}`
  const systemName = textValue(issue.systemName).trim()
  return systemName ? `name:${systemName}` : ''
}

const pendingIssueSystems = (issues = []) => {
  const seen = new Set()
  const systems = []

  issues.forEach((issue) => {
    const key = pendingIssueSystemKey(issue)
    if (!key || seen.has(key)) return
    seen.add(key)
    systems.push({
      groupId: issue.groupId,
      name: textValue(issue.systemName).trim() || '未命名系统',
      score: Number.isFinite(Number(issue.systemScore)) ? Number(issue.systemScore) : 0,
      status: issue.systemStatus || 'gray',
    })
  })

  return systems
}

const hasPercentValue = (value) => textValue(value).includes('%')

const isResourceAvailabilityIssue = (issue = {}) => {
  if (issue.metricKey !== 'resource' || hasPercentValue(issue.value)) return false
  return /副本|实例|可用|replica/i.test(textValue(issue.note))
}

const pendingIssueMetricLabel = (issue = {}) => {
  if (isResourceAvailabilityIssue(issue)) return '资源可用性'
  return textValue(issue.metricLabel) || textValue(issue.metricKey) || '未知指标'
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
        metricLabel: pendingIssueMetricLabel(leadIssue),
        systemName: uniqueTexts(sortedIssues, 'systemName') || leadIssue.systemName,
        systems: pendingIssueSystems(sortedIssues),
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
