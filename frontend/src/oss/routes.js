/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

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
const FaultSiteTrace = React.lazy(() => import('src/oss/views/trace/FaultSiteTrace.jsx'))
const FullTrace = React.lazy(() => import('src/oss/views/trace/FullTrace.jsx'))
const AlertsRule = React.lazy(() => import('src/oss/views/alerts/AlertsRule'))
const AlertsNotify = React.lazy(() => import('src/oss/views/alerts/AlertsNotify'))
const ConfigPage = React.lazy(() => import('src/oss/views/config/index'))

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
    path: '/trace/fault-site',
    name: <TranslationCom text="faultSiteTraces" space={namespace} />,
    element: FaultSiteTrace,
    hideSystemTimeRangePicker: true,
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
  },
  {
    path: '/alerts/notify',
    name: <TranslationCom text="notificationChannelsName" space={namespace} />,
    element: AlertsNotify,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/config',
    name: <TranslationCom text="configurationsName" space={namespace} />,
    element: ConfigPage,
    hideSystemTimeRangePicker: true,
  },
]
export default ossRoutes
