/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import ossRoutes from './oss/routes'
const UserPage = React.lazy(() => import('src/core/views/UserPage/index.jsx'))
const UserManage = React.lazy(() => import('src/core/views/UserManage/index.jsx'))
const MenuManage = React.lazy(() => import('src/core/views/MenuManage/index.jsx'))

const baseRoutes = [
  { path: '/', exact: true, name: 'Home' },
  { path: '/user', name: '个人中心', element: UserPage, hideSystemTimeRangePicker: true },
  {
    path: '/system/user-manage',
    name: '用户管理',
    element: UserManage,
    hideSystemTimeRangePicker: true,
  },
  {
    path: '/system/menu-manage',
    name: '菜单管理',
    element: MenuManage,
    hideSystemTimeRangePicker: true,
  },
]
const routes = [...baseRoutes, ...ossRoutes]
export default routes
