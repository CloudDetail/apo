import {
  getServiceErrorInstancesLogsApi,
  getServiceLogLogsApi,
  getServiceTraceLogsApi,
} from './api/serviceInfo'

export const DelaySourceTimeUnit = {
  self: '自身',
  dependency: '依赖',
}

export const MetricsLineChartColor = {
  latency: 'rgba(212, 164, 235, 1)',
  successRate: 'rgba(144, 202, 140, 1)',
  errorRate: 'rgba(255, 99, 132, 1)',
  tps: 'rgba(55, 162,235, 1)',
  logs: 'rgba(255, 158, 64, 1)',
}

export const StatusColorMap = {
  normal: '#24d160',
  warning: '#f9bb07',
  critical: '#ff3366',
  success: '#24d160',
  error: '#ff3366',
  unknown: '',
}

export const TimeLineTypeApiMap = {
  errorLogs: getServiceErrorInstancesLogsApi,
  logsInfo: getServiceLogLogsApi,
  traceLogs: getServiceTraceLogsApi,
}
export const TimeLineTypeTitleMap = {
  errorLogs: '错误日志',
  logsInfo: '故障现场日志',
  traceLogs: '故障现场Trace',
}

export const DelayLineChartTitleMap = {
  latency: '平均响应时间',
  errorRate: '错误率',
  logs: '日志错误数量',
}

export const YValueMinInterval = {
  latency: 0.01,
  errorRate: 0.01,
  logs: 1,
}
