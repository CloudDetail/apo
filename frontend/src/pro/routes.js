import { t } from 'i18next'
import React from 'react'
import AuthRouter from 'src/core/components/AuthRouter'

// 商业版
const AlertAnalyzePage = AuthRouter(React.lazy(() => import('src/pro/views/alertAnalyze/index')))
const ServiceAnalyzePage = AuthRouter(
  React.lazy(() => import('src/pro/views/serviceAnalyze/index')),
)
const CauseReportPage = AuthRouter(React.lazy(() => import('src/pro/views/causeReport/index')))

const SystemHealthOverviewPage = React.lazy(
  () => import('src/pro/views/systemHealthOverview/index'),
)
const SystemHealthOverview = import.meta.env.DEV
  ? SystemHealthOverviewPage
  : AuthRouter(SystemHealthOverviewPage)
const HealthLabPage = AuthRouter(React.lazy(() => import('src/pro/views/healthLab/index')))
const MysqlDashboard = AuthRouter(React.lazy(() => import('src/pro/views/MysqlDashboard/index')))
const ChatAIPage = AuthRouter(React.lazy(() => import('src/pro/views/chatAI/index')))
const SlowReportPage = AuthRouter(React.lazy(() => import('src/pro/views/report/index')))
const AgentDataplanePage = AuthRouter(
  React.lazy(() => import('src/pro/views/agentDataplane/index')),
)
const AgentPage = AuthRouter(React.lazy(() => import('src/pro/views/agent/index')))
const KnowledgeBasesPage = AuthRouter(
  React.lazy(() => import('src/pro/views/knowledgeBases/index')),
)
const CustomDashboardList = AuthRouter(React.lazy(() => import('src/pro/views/CustomDashboard/List')))
const CustomDashboardViewer = AuthRouter(React.lazy(() => import('src/pro/views/CustomDashboard/Viewer')))
const CustomDashboardEditor = AuthRouter(React.lazy(() => import('src/pro/views/CustomDashboard/Editor')))
const ProbeManagementPage = AuthRouter(
  React.lazy(() => import('src/core/views/IntegrationCenter/DataIntegration')),
)

const IntegrationSettings = AuthRouter(
  React.lazy(() => import('src/core/views/IntegrationCenter/DataIntegration/Integration')),
)
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
    name: '健康度',
    element: SystemHealthOverview,
  },
  {
    path: '/health-lab',
    name: '健康度实验室',
    element: HealthLabPage,
    showDataGroup: true,
  },
  {
    path: '/custom-dashboard',
    name: t('pro/routes:customDashboard'),
    hideSystemTimeRangePicker: true,
    element: CustomDashboardList,
  },
  {
    path: '/custom-dashboard/:id',
    name: t('pro/routes:customDashboard'),
    showDataGroup: true,

    element: CustomDashboardViewer,
  },
  {
    path: '/custom-dashboard/:id/edit',
    name: t('pro/routes:customDashboardEditor'),
    showDataGroup: true,

    element: CustomDashboardEditor,
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
    path: '/dataplane',
    name: '外部数据源',
    element: AgentDataplanePage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/system/platform-manage',
    name: t('pro/routes:licenseManage'),
    element: PlatformManage,
  },
  // {
  //   path: '/report/error',
  //   name: <TranslationCom text="errorReport" space={namespace} />,
  //   element: ErrorReportPage,
  //   hideSystemTimeRangePicker: true,
  // },
  {
    path: '/agent',
    name: 'Agent',
    element: AgentPage,
    hideSystemTimeRangePicker: true,

  },
  {
    path: '/knowledge-bases',
    name: '知识库管理',
    element: KnowledgeBasesPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/probe-management',
    name: 'ProbeInstaller',
    element: ProbeManagementPage,
  },

  {
    path: '/probe-management/settings',
    element: IntegrationSettings,
  },
]
export default proRoutes
