/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import ossRoutes from './oss/routes'
import TranslationCom from './oss/components/TranslationCom'
const UserPage = React.lazy(() => import('src/core/views/UserPage/index.jsx'))
const UserManage = React.lazy(() => import('src/core/views/UserManage/index.jsx'))
const MenuManage = React.lazy(() => import('src/core/views/MenuManage/index.jsx'))
const AlertsIntegrationPage = React.lazy(
  () => import('src/core/views/IntegrationCenter/AlertsIntegration'),
)
const SystemConfiguration = React.lazy(() => import('src/core/views/SystemConfiguration/index.jsx'))
const namespace = 'oss/routes'

const baseRoutes = [
  { path: '/', exact: true, name: 'Home' },
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
  {
    path: '/system/menu-manage',
    name: <TranslationCom text="memuManageName" space={namespace} />,
    element: MenuManage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/integration/alerts',
    name: <TranslationCom text="alertsIntegration" space={namespace} />,
    element: AlertsIntegrationPage,
  },
  {
    path: '/system/config',
    name: <TranslationCom text="systemConfigName" space={namespace} />,
    element: SystemConfiguration,
    hideSystemTimeRangePicker: true,
  },
]
const routes = [...baseRoutes, ...ossRoutes]
export default routes
