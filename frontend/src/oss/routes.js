/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import TranslationCom from './components/TranslationCom.jsx'
import { Tooltip } from 'antd'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import AuthRouter from 'src/core/components/AuthRouter'

const BasicDashboard = AuthRouter(
  React.lazy(() => import('src/oss/views/dashboard/BasicDashboard')),
)
const SystemDashboard = AuthRouter(
  React.lazy(() => import('src/oss/views/dashboard/SystemDashboard')),
)
const ApplicationDashboard = AuthRouter(
  React.lazy(() => import('src/oss/views/dashboard/ApplicationDashboard')),
)
const MiddlewareDashboard = AuthRouter(
  React.lazy(() => import('src/oss/views/dashboard/MiddlewareDashboard')),
)
const Service = AuthRouter(React.lazy(() => import('src/oss/views/service/index.js')))
const ServiceInfo = AuthRouter(React.lazy(() => import('src/oss/views/serviceInfo/index.js')))
const FaultSiteLogsPage = AuthRouter(
  React.lazy(() => import('src/oss/views/logs/FaultSiteLogsPage')),
)
const FullLogsPage = AuthRouter(React.lazy(() => import('src/oss/views/logs/FullLogsPage')))
const FaultSiteTrace = AuthRouter(
  React.lazy(() => import('src/oss/views/trace/FaultSiteTracePage.tsx')),
)
const FullTrace = AuthRouter(React.lazy(() => import('src/oss/views/trace/FullTrace.jsx')))
const AlertsRule = AuthRouter(React.lazy(() => import('src/oss/views/alerts/AlertsRule')))
const AlertsNotify = AuthRouter(React.lazy(() => import('src/oss/views/alerts/AlertsNotify')))
const ConfigPage = AuthRouter(React.lazy(() => import('src/oss/views/config/index')))
const AlertEventsPage = AuthRouter(React.lazy(() => import('src/oss/views/alertEvents/index')))
const WorkflowsPage = AuthRouter(React.lazy(() => import('src/oss/views/workflows/index')))
const AlertEventDetailPage = AuthRouter(
  React.lazy(() => import('src/oss/views/alertEvents/detail/index')),
)

const namespace = 'core/routes'

const ossRoutes = [
  {
    path: '/service',
    exact: true,
    name: (
      <Tooltip title={<TranslationCom text="index.serviceTableToast" space={'oss/service'} />}>
        <div className="flex items-center">
          <TranslationCom text="servicesName" space={namespace} />
          <IoMdInformationCircleOutline size={20} color="#f7c01a" className="ml-2" />
        </div>
      </Tooltip>
    ),
    element: Service,
    showDataGroup: true,
  },
  {
    path: '/service/info',
    name: <TranslationCom text="serviceDetailName" space={namespace} />,
    element: ServiceInfo,
    showDataGroup: 'view',
  },
  {
    path: '/logs/fault-site',
    name: <TranslationCom text="faultLogsName" space={namespace} />,
    element: FaultSiteLogsPage,
    hideSystemTimeRangePicker: true,
    showDataGroup: true,
  },
  {
    path: '/logs/full',
    name: <TranslationCom text="allLogsName" space={namespace} />,
    element: FullLogsPage,
  },
  {
    path: '/trace/fault-site',
    name: <TranslationCom text="faultSiteTraces" space={namespace} />,
    element: FaultSiteTrace,
    hideSystemTimeRangePicker: true,
    showDataGroup: true,
  },
  {
    path: '/trace/full',
    name: <TranslationCom text="allTrace" space={namespace} />,
    element: FullTrace,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/system-dashboard',
    name: <TranslationCom text="overviewDashboardName" space={namespace} />,
    element: SystemDashboard,
  },
  {
    path: '/basic-dashboard',
    name: <TranslationCom text="infrastructureDashboardName" space={namespace} />,
    element: BasicDashboard,
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
  // { path: '/alerts', name: '告警规则', hideSystemTimeRangePicker: true },
  {
    path: '/alerts/rule',
    name: <TranslationCom text="alertRulesName" space={namespace} />,
    element: AlertsRule,
    hideSystemTimeRangePicker: true,
    showDataGroup: true,
  },
  {
    path: '/alerts/notify',
    name: <TranslationCom text="notificationChannelsName" space={namespace} />,
    element: AlertsNotify,
    hideSystemTimeRangePicker: true,
    showDataGroup: true,
  },
  {
    path: '/config',
    name: <TranslationCom text="configurationsName" space={namespace} />,
    element: ConfigPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/alerts/events',
    name: <TranslationCom text="alertEvents" space={namespace} />,
    element: AlertEventsPage,
    showDataGroup: true,
  },
  {
    path: '/workflows',
    name: <TranslationCom text="workflows" space={namespace} />,
    element: WorkflowsPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/alerts/events/detail/:alertId/:eventId',
    name: <TranslationCom text="eventDetail" space={namespace} />,
    element: AlertEventDetailPage,
    hideSystemTimeRangePicker: true,
  },
]
export default ossRoutes
