/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import ossRoutes from './oss/routes'
import TranslationCom from './oss/components/TranslationCom'
import AuthRouter from 'src/core/components/AuthRouter'
const UserPage = AuthRouter(React.lazy(() => import('src/core/views/UserPage/index.jsx')))
const UserManage = AuthRouter(React.lazy(() => import('src/core/views/UserManage/index')))
const MenuManage = AuthRouter(React.lazy(() => import('src/core/views/MenuManage/index')))
const RoleManage = AuthRouter(React.lazy(() => import('src/core/views/RoleManage/index')))
const AlertsIntegrationPage = AuthRouter(React.lazy(
  () => import('src/core/views/IntegrationCenter/AlertsIntegration'),
))
const DataGroupPage = AuthRouter(React.lazy(() => import('src/core/views/DataGroup/index')))
const TeamPage = AuthRouter(React.lazy(() => import('src/core/views/Team/index')))
const DataIntegrationPage = AuthRouter(React.lazy(
  () => import('src/core/views/IntegrationCenter/DataIntegration'),
))
const IntegrationSettings = AuthRouter(React.lazy(
  () => import('src/core/views/IntegrationCenter/DataIntegration/Integration'),
))
const namespace = 'core/routes'

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
    name: <TranslationCom text="menuManageName" space={namespace} />,
    element: MenuManage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/system/role-manage',
    name: <TranslationCom text="roleManageName" space={namespace} />,
    element: RoleManage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/integration/alerts',
    name: <TranslationCom text="alertsIntegration" space={namespace} />,
    element: AlertsIntegrationPage,
  },
  {
    path: '/integration/data',
    name: <TranslationCom text="dataIntegration" space={namespace} />,
    element: DataIntegrationPage,
  },
  {
    path: '/integration/data/settings',
    name: <TranslationCom text="dataIntegrationSettings" space={namespace} />,
    element: IntegrationSettings,
  },
  {
    path: '/system/data-group',
    name: <TranslationCom text="dataGroup" space={namespace} />,
    element: DataGroupPage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/system/team',
    name: <TranslationCom text="team" space={namespace} />,
    element: TeamPage,
    hideSystemTimeRangePicker: true,
  },
]
const routes = [...baseRoutes, ...ossRoutes]
export default routes
