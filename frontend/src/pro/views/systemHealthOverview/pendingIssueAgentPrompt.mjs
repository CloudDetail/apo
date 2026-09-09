const STATUS_LABELS = {
  red: '异常',
  yellow: '关注',
}

const RELATIVE_TIME_LABELS = {
  '5m': '最近 5 分钟',
  '15m': '最近 15 分钟',
  '30m': '最近 30 分钟',
  '1h': '最近 1 小时',
  '6h': '最近 6 小时',
  '12h': '最近 12 小时',
  '24h': '最近 24 小时',
  '7d': '最近 7 天',
}

const safeText = (value, fallback = '未提供') => {
  const text = value === undefined || value === null ? '' : `${value}`.trim()
  return text || fallback
}

const scoreText = (value) => {
  const score = Number(value)
  return Number.isFinite(score) ? `${score}` : '未提供'
}

const formatTimestamp = (value) => {
  const timestamp = Number(value)
  if (!Number.isFinite(timestamp) || timestamp <= 0) return ''
  const date = new Date(timestamp > 1e14 ? timestamp / 1000 : timestamp)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'short',
    timeStyle: 'medium',
  }).format(date)
}

const formatTimeWindow = (params = {}) => {
  const startTime = formatTimestamp(params.startTime)
  const endTime = formatTimestamp(params.endTime)
  if (startTime && endTime) return `${startTime} 至 ${endTime}`

  const relativeTime = safeText(params.relativeTime, '')
  return RELATIVE_TIME_LABELS[relativeTime] || (relativeTime ? `最近 ${relativeTime}` : '当前健康度查询时间窗口')
}

const systemScopeText = (issue = {}) => {
  if (Array.isArray(issue.systems) && issue.systems.length) {
    return issue.systems
      .map((system) => {
        const name = safeText(system?.name || system?.systemName, '未命名系统')
        const groupId =
          system?.groupId === undefined || system?.groupId === null
            ? ''
            : `，数据组 ID：${system.groupId}`
        return `${name}${groupId}`
      })
      .join('；')
  }

  const groupId =
    issue.groupId === undefined || issue.groupId === null ? '' : `，数据组 ID：${issue.groupId}`
  return `${safeText(issue.systemName, '未命名系统')}${groupId}`
}

export const buildPendingIssueAgentPrompt = (issue = {}, params = {}) => {
  const status = safeText(issue.status, 'unknown')
  const statusLabel = STATUS_LABELS[status] || '待确认'
  const metricKey = safeText(issue.metricKey)
  const metricLabel = safeText(issue.metricLabel, metricKey)
  const issueCount = Number(issue.issueCount)
  const issueCountText =
    Number.isInteger(issueCount) && issueCount > 1 ? `\n同类异常涉及事项数：${issueCount}` : ''

  return [
    '请分析以下健康度页面发现的服务异常，并通过当前可访问的可观测数据验证原因。',
    '',
    '【健康度页面上下文】',
    `系统范围：${systemScopeText(issue)}`,
    `服务：${safeText(issue.serviceName, '未命名服务')}`,
    `异常指标：${metricLabel}（${metricKey}）`,
    `当前状态：${statusLabel}（${status}）`,
    `当前值：${safeText(issue.value)}`,
    `指标说明：${safeText(issue.note)}`,
    `服务健康分：${scoreText(issue.serviceScore)}`,
    `系统健康分：${scoreText(issue.systemScore)}`,
    `分析时间窗口：${formatTimeWindow(params)}${issueCountText}`,
    '',
    '【分析要求】',
    '1. 判断该异常的实际影响范围，以及是否影响请求成功率、响应时延、可用性或其他用户体验。',
    '2. 在上述时间窗口内查询并关联相关指标、日志、Trace、调用链、告警、实例和资源数据；优先使用可验证的观测证据。',
    '3. 给出有证据支撑的根因候选，区分已确认事实、推断结论和仍需验证的假设。',
    '4. 输出定位、排查、处置和恢复验证建议，明确下一步应执行的查询或操作。',
    '5. 不要只复述健康度页面字段；不要编造不存在的日志、Trace、告警、变更记录或根因。如果某类数据不可访问，请明确说明数据边界。',
  ].join('\n')
}
