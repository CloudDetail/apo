import React from 'react'

const BasicDashboard = React.lazy(() => import('./views/dashboard/BasicDashboard'))
const SystemDashboard = React.lazy(() => import('./views/dashboard/SystemDashboard'))
const ApplicationDashboard = React.lazy(() => import('./views/dashboard/ApplicationDashboard'))
const MiddlewareDashboard = React.lazy(() => import('./views/dashboard/MiddlewareDashboard'))
const Service = React.lazy(() => import('./views/service/index.js'))
const ServiceInfo = React.lazy(() => import('./views/serviceInfo/index.js'))
const LogsPage = React.lazy(() => import('./views/logs/index.js'))
const TracePage = React.lazy(() => import('./views/trace/index.js'))
const Alerts = React.lazy(() => import('./views/alerts/index.js'))
const ConfigPage = React.lazy(() => import('./views/config/index'))

const routes = [
  { path: '/', exact: true, name: 'Home' },
  { path: '/service', exact: true, name: '服务概览', element: Service },
  { path: '/service/info', name: '服务详情', element: ServiceInfo },
  { path: '/logs', name: '日志检索', element: LogsPage, hideSystemTimeRangePicker: true },
  { path: '/trace', name: '链路追踪', element: TracePage, hideSystemTimeRangePicker: true },
  { path: '/basic-dashboard', name: '应用基础设施大盘', element: BasicDashboard },
  { path: '/system-dashboard', name: '全局资源大盘', element: SystemDashboard },
  { path: '/application-dashboard', name: '应用指标大盘', element: ApplicationDashboard },
  { path: '/middleware-dashboard', name: '中间件大盘', element: MiddlewareDashboard },
  { path: '/alerts', name: '告警规则', element: Alerts, hideSystemTimeRangePicker: true },
  { path: '/config', name: '配置中心', element: ConfigPage, hideSystemTimeRangePicker: true },
]

export default routes
