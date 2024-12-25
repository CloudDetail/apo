import { element } from 'prop-types'
import React from 'react'
import TranslationCom from './components/TranslationCom.jsx'

const BasicDashboard = React.lazy(() => import('src/oss/views/dashboard/BasicDashboard'))
const SystemDashboard = React.lazy(() => import('src/oss/views/dashboard/SystemDashboard'))
const ApplicationDashboard = React.lazy(
  () => import('src/oss/views/dashboard/ApplicationDashboard'),
)
const MiddlewareDashboard = React.lazy(() => import('src/oss/views/dashboard/MiddlewareDashboard'))
const Service = React.lazy(() => import('src/oss/views/service/index.js'))
const ServiceInfo = React.lazy(() => import('src/oss/views/serviceInfo/index.js'))
const FaultSiteLogsPage = React.lazy(() => import('src/oss/views/logs/FaultSiteLogs'))
const FullLogsPage = React.lazy(() => import('src/oss/views/logs/FullLogsPage'))
const TracePage = React.lazy(() => import('src/oss/views/trace/index.js'))
const Alerts = React.lazy(() => import('src/oss/views/alerts/index.js'))
const ConfigPage = React.lazy(() => import('src/oss/views/config/index'))
const UserPage = React.lazy(() => import('./views/user/UserPage'))
const UserManage = React.lazy(() => import('./views/user/UserManage'))

const translationUrl = 'oss/routes'

const ossRoutes = [
  {
    path: '/service',
    exact: true,
    name: <TranslationCom text="servicesName" url={translationUrl} />,
    element: Service,
  },
  {
    path: '/service/info',
    name: <TranslationCom text="serviceDetailName" url={translationUrl} />,
    element: ServiceInfo,
  },
  {
    path: '/logs/fault-site',
    name: <TranslationCom text="faultLogsName" url={translationUrl} />,
    element: FaultSiteLogsPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/logs/full',
    name: <TranslationCom text="allLogsName" url={translationUrl} />,
    element: FullLogsPage,
  },
  {
    path: '/trace',
    name: <TranslationCom text="tracesName" url={translationUrl} />,
    element: TracePage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/basic-dashboard',
    name: <TranslationCom text="infrastructureDashboardName" url={translationUrl} />,
    element: BasicDashboard,
  },
  {
    path: '/system-dashboard',
    name: <TranslationCom text="overviewDashboardName" url={translationUrl} />,
    element: SystemDashboard,
  },
  {
    path: '/application-dashboard',
    name: <TranslationCom text="applicationDashboardName" url={translationUrl} />,
    element: ApplicationDashboard,
  },
  {
    path: '/middleware-dashboard',
    name: <TranslationCom text="middlewareDashboardName" url={translationUrl} />,
    element: MiddlewareDashboard,
  },
  {
    path: '/alerts',
    name: <TranslationCom text="alertsName" url={translationUrl} />,
    element: Alerts,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/config',
    name: <TranslationCom text="configurationsName" url={translationUrl} />,
    element: ConfigPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/user',
    name: <TranslationCom text="userCenterName" url={translationUrl} />,
    element: UserPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/system/user-manage',
    name: <TranslationCom text="userManageName" url={translationUrl} />,
    element: UserManage,
    hideSystemTimeRangePicker: true,
  },
]
export default ossRoutes
