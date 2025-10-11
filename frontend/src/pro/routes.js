import { t } from 'i18next'
import React from 'react'

// 商业版
const AlertAnalyzePage = React.lazy(() => import('src/pro/views/alertAnalyze/index'))
const ServiceAnalyzePage = React.lazy(() => import('src/pro/views/serviceAnalyze/index'))
const CauseReportPage = React.lazy(() => import('src/pro/views/causeReport/index'))

const ServiceHealthy = React.lazy(() => import('src/pro/views/serviceHealthy/index'))
const MysqlDashboard = React.lazy(() => import('src/pro/views/MysqlDashboard/index'))
const ChatAIPage = React.lazy(() => import('src/pro/views/chatAI/index'))
const SlowReportPage = React.lazy(() => import('src/pro/views/report/index'))
const PlatformManage = React.lazy(() => import('src/pro/views/platformManage/index'))

let proRoutes = [
  {
    path: '/alert-analyze',
    exact: true,
    name: t('pro/routes:alertAnalyze'),
    element: AlertAnalyzePage,
    showDataGroup: true,
  },
  {
    path: '/alert-analyze/info',
    name: t('pro/routes:businessAlertAnalyze'),
    element: ServiceAnalyzePage,
    showDataGroup: 'view',
  },
  {
    path: '/cause/report/:traceId/:spanId',
    name: t('pro/routes:causeReport'),
    element: CauseReportPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/cause/error-report/:traceId',
    name: t('pro/routes:causeReport'),
    element: CauseReportPage,
    hideSystemTimeRangePicker: true,
  },
  { path: '/mysql-dashboard', name: t('pro/routes:mysqlDashboard'), element: MysqlDashboard },
  {
    path: '/healthy-service',
    name: t('pro/routes:serviceHealth'),
    element: ServiceHealthy,
    showDataGroup: true,
  },
  {
    path: '/alert-analyze/chatAI/:service/:endpoint',
    name: t('pro/routes:aiRootCauseAnalysis'),
    element: ChatAIPage,
    showDataGroup: 'view',
  },
  {
    path: '/report',
    name: t('pro/routes:causeReport'),
    element: SlowReportPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/system/platform-manage',
    name: t('pro/routes:licenseManage'),
    element: PlatformManage
  },
  // {
  //   path: '/report/error',
  //   name: <TranslationCom text="errorReport" space={namespace} />,
  //   element: ErrorReportPage,
  //   hideSystemTimeRangePicker: true,
  // },
]
export default proRoutes
