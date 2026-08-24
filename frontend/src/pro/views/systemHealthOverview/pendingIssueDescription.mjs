const STATUS_DESCRIPTIONS = {
  red: '已处于异常状态',
  yellow: '已进入关注状态',
  green: '处于健康状态',
  gray: '当前缺少可用数据',
}

const safeText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

const trimSentenceEnd = (value) => safeText(value).replace(/[。；;.\s]+$/u, '')

const issueSubject = (issue = {}) => {
  const systemName = safeText(issue.systemName) || '未命名系统'
  const serviceName = safeText(issue.serviceName) || '未命名服务'
  const metricLabel = safeText(issue.metricLabel) || safeText(issue.metricKey) || '指标'
  return `${systemName} 中 ${serviceName} 的${metricLabel}`
}

const issueValueText = (issue = {}) => {
  const value = safeText(issue.value)
  if (!value || value === '--') return ''
  return `当前值为 ${value}，`
}

export const describePendingIssue = (issue = {}) => {
  const statusText = STATUS_DESCRIPTIONS[issue.status] || '状态需要确认'
  const note = trimSentenceEnd(issue.note)
  return `${issueSubject(issue)}${issueValueText(issue)}${statusText}${note ? `；${note}` : ''}。`
}
