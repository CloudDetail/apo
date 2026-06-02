/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { useLocation } from 'react-router-dom'

import routes from 'src/routes'

import { Breadcrumb } from 'antd'
import { useUserContext } from '../contexts/UserContext'
import { getMenuBreadcrumbs, getRouteBreadcrumbs } from './AppBreadcrumbBehavior.mjs'

const AppBreadcrumb = () => {
  const currentLocation = useLocation().pathname
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)

  const { menuItems } = useUserContext()

  const labelPath = getMenuBreadcrumbs(menuItems, currentLocation);
  const routePath = getRouteBreadcrumbs(currentLocation, routes)
  const breadcrumbPath = labelPath.length > 0 ? labelPath : routePath
  return (
    <Breadcrumb items={breadcrumbPath} className="text-base" />
    // <CBreadcrumb className="my-0">
    //   {/* <CBreadcrumbItem href="/">Home</CBreadcrumbItem> */}
    //   {breadcrumbs.map((breadcrumb, index) => {
    //     return (
    //       <CBreadcrumbItem
    //         {...(breadcrumb.active ? { active: true } : { href: breadcrumb.pathname })}
    //         key={index}
    //       >
    //         {breadcrumb.active && breadcrumbName ? breadcrumbName : breadcrumb.name}
    //       </CBreadcrumbItem>
    //     )
    //   })}
    // </CBreadcrumb>
  )
}

export default React.memo(AppBreadcrumb)
