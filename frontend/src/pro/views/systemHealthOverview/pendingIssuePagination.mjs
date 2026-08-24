export const DEFAULT_PENDING_ISSUE_PAGE = 1
export const DEFAULT_PENDING_ISSUE_PAGE_SIZE = 20
export const MAX_PENDING_ISSUE_PAGE_SIZE = 100

const metricKeys = [
  'resource',
  'resourceShift',
  'errorRate',
  'latency',
  'errorLog',
  'badRequests',
  'lockCost',
]

const metricLabels = {
  resource: '资源水位',
  resourceShift: '资源波动',
  errorRate: '错误率',
  latency: '响应延时',
  errorLog: '错误日志',
  badRequests: '活跃告警',
  lockCost: '锁等待',
}

const severityValue = (status) =>
  ({
    red: 4,
    yellow: 3,
  })[status] || 0

const positiveInteger = (value) => {
  const number = Number(value)
  return Number.isInteger(number) && number > 0 ? number : 0
}

const safeText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

const numberValue = (value) => {
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

const metricOrder = (metricKey) => {
  const index = metricKeys.indexOf(metricKey)
  return index >= 0 ? index : metricKeys.length
}

const normalizedPageSize = (pageSize) => {
  const nextPageSize = positiveInteger(pageSize) || DEFAULT_PENDING_ISSUE_PAGE_SIZE
  return Math.min(nextPageSize, MAX_PENDING_ISSUE_PAGE_SIZE)
}

const comparePendingIssues = (left = {}, right = {}) => {
  const severityDelta = severityValue(right.status) - severityValue(left.status)
  if (severityDelta !== 0) return severityDelta

  const serviceScoreDelta = numberValue(left.serviceScore) - numberValue(right.serviceScore)
  if (serviceScoreDelta !== 0) return serviceScoreDelta

  const systemScoreDelta = numberValue(left.systemScore) - numberValue(right.systemScore)
  if (systemScoreDelta !== 0) return systemScoreDelta

  return (
    [
      safeText(left.systemName).localeCompare(safeText(right.systemName)),
      safeText(left.serviceName).localeCompare(safeText(right.serviceName)),
      metricOrder(left.metricKey) - metricOrder(right.metricKey),
    ].find((delta) => delta !== 0) || 0
  )
}

export const buildPendingIssuesFromSystems = (systems = []) =>
  systems
    .flatMap((system = {}) =>
      (Array.isArray(system.services) ? system.services : []).flatMap((service = {}) =>
        metricKeys
          .map((metricKey) => {
            const status = service.metrics?.[metricKey]
            if (status !== 'red' && status !== 'yellow') return null
            const detail = service.metricDetails?.[metricKey] || {}
            return {
              id: `${system.groupId || 0}:${safeText(service.name)}:${metricKey}`,
              groupId: system.groupId,
              systemName: safeText(system.name) || '未命名系统',
              serviceName: safeText(service.name) || '未命名服务',
              metricKey,
              metricLabel: metricLabels[metricKey] || metricKey,
              status,
              value: safeText(detail.value) || '--',
              note: safeText(detail.note) || '暂无指标说明',
              serviceScore: numberValue(service.score),
              systemScore: numberValue(system.score),
            }
          })
          .filter(Boolean),
      ),
    )
    .sort(comparePendingIssues)

export const paginatePendingIssues = (
  issues = [],
  page = DEFAULT_PENDING_ISSUE_PAGE,
  pageSize = DEFAULT_PENDING_ISSUE_PAGE_SIZE,
) => {
  const nextPage = positiveInteger(page) || DEFAULT_PENDING_ISSUE_PAGE
  const nextPageSize = normalizedPageSize(pageSize)
  const total = Array.isArray(issues) ? issues.length : 0
  const start = (nextPage - 1) * nextPageSize
  return {
    issues: Array.isArray(issues) ? issues.slice(start, start + nextPageSize) : [],
    total,
    page: nextPage,
    pageSize: nextPageSize,
  }
}
