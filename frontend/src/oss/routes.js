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
const UserPage = React.lazy(() => import('../core/views/user/UserPage'))
const UserManage = React.lazy(() => import('../core/views/user/UserManage'))

const namespace = 'oss/routes'

const ossRoutes = [
  {
    path: '/service',
    exact: true,
    name: <TranslationCom text="servicesName" space={namespace} />,
    element: Service,
  },
  {
    path: '/service/info',
    name: <TranslationCom text="serviceDetailName" space={namespace} />,
    element: ServiceInfo,
  },
  {
    path: '/logs/fault-site',
    name: <TranslationCom text="faultLogsName" space={namespace} />,
    element: FaultSiteLogsPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/logs/full',
    name: <TranslationCom text="allLogsName" space={namespace} />,
    element: FullLogsPage,
  },
  {
    path: '/trace',
    name: <TranslationCom text="tracesName" space={namespace} />,
    element: TracePage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/basic-dashboard',
    name: <TranslationCom text="infrastructureDashboardName" space={namespace} />,
    element: BasicDashboard,
  },
  {
    path: '/system-dashboard',
    name: <TranslationCom text="overviewDashboardName" space={namespace} />,
    element: SystemDashboard,
  },
  {
    path: '/application-dashboard',
    name: <TranslationCom text="applicationDashboardName" space={namespace} />,
    element: ApplicationDashboard,
  },
  {
    path: '/middleware-dashboard',
    name: <TranslationCom text="middlewareDashboardName" space={namespace} />,
    element: MiddlewareDashboard,
  },
  {
    path: '/alerts',
    name: <TranslationCom text="alertsName" space={namespace} />,
    element: Alerts,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/config',
    name: <TranslationCom text="configurationsName" space={namespace} />,
    element: ConfigPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/user',
    name: <TranslationCom text="userCenterName" space={namespace} />,
    element: UserPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/system/user-manage',
    name: <TranslationCom text="userManageName" space={namespace} />,
    element: UserManage,
    hideSystemTimeRangePicker: true,
  },
]
export default ossRoutes
