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
    name: '故障现场链路',
    element: FaultSiteTrace,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/trace/full',
    name: '全量链路',
    element: FullTrace,
    hideSystemTimeRangePicker: true,
  },
  { path: '/basic-dashboard', name: '应用基础设施大盘', element: BasicDashboard },
  { path: '/system-dashboard', name: '全局资源大盘', element: SystemDashboard },
  { path: '/application-dashboard', name: '应用指标大盘', element: ApplicationDashboard },
  { path: '/middleware-dashboard', name: '中间件大盘', element: MiddlewareDashboard },
  // { path: '/alerts', name: '告警规则', hideSystemTimeRangePicker: true },
  { path: '/alerts/rule', name: '告警规则', element: AlertsRule, hideSystemTimeRangePicker: true },
  {
    path: '/alerts/notify',
    name: '告警通知',
    element: AlertsNotify,
    hideSystemTimeRangePicker: true,
  },
  { path: '/config', name: '配置中心', element: ConfigPage, hideSystemTimeRangePicker: true },
]
export default ossRoutes
