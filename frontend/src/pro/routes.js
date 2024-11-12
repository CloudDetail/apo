import React from 'react'

// 商业版
const AlertAnalyzePage = React.lazy(() => import('src/pro/views/alertAnalyze/index'))
const ServiceAnalyzePage = React.lazy(() => import('src/pro/views/serviceAnalyze/index'))
const CauseReportPage = React.lazy(() => import('src/pro/views/causeReport/index'))

const ServiceHealthy = React.lazy(() => import('src/pro/views/serviceHealthy/index'))
const MysqlDashboard = React.lazy(() => import('src/pro/views/MysqlDashboard/index'))

let proRoutes = [
  {
    path: '/alert-analyze',
    exact: true,
    name: '告警分析',
    element: AlertAnalyzePage,
    // children: [{ path: 'info', name: '业务告警分析', element: ServiceAnalyzePage }],
  },
  { path: '/alert-analyze/info', name: '业务告警分析', element: ServiceAnalyzePage },
  {
    path: '/cause/report/:traceId/:spanId',
    name: '故障报告',
    element: CauseReportPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/cause/error-report/:traceId',
    name: '故障报告',
    element: CauseReportPage,
    hideSystemTimeRangePicker: true,
  },
  { path: '/mysql-dashboard', name: 'MySQL大盘', element: MysqlDashboard },
  {
    path: '/healthy-service',
    name: '服务健康状态',
    element: ServiceHealthy,
  },
]
export default proRoutes
