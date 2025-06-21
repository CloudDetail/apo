/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { useLocation } from 'react-router-dom'

import routes from 'src/routes'

import { Breadcrumb } from 'antd'
import { useUserContext } from '../contexts/UserContext'

const AppBreadcrumb = () => {
  const currentLocation = useLocation().pathname
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)

  const { menuItems } = useUserContext()

  function getLabelPathFromMenuTree(tree, route) {
    const result = [];

    /**
     * @param {Array} nodes - Array of nodes at the current level
     * @param {Array} path - Array of title objects representing the current path
     */
    function dfs(nodes, path) {
      if (!nodes || nodes.length === 0) return;

      for (const node of nodes) {
        const newPath = [...path, { title: node.label }];

        if (node.router && node.router.to === route) {
          result.push(newPath);
          return;
        }

        // If the current node has children, continue recursive search
        if (node.children) {
          if (Array.isArray(node.children)) {
            dfs(node.children, newPath);
          }
          else if (typeof node.children === 'object' && node.children !== null) {
            dfs([node.children], newPath);
          }
        }

        if (result.length > 0) return;
      }
    }

    dfs(tree, []);

    return result.length > 0 ? result[0] : [];
  }

  const getRouteName = (pathname, routes) => {
    // 遍历所有路由，检查是否有符合的路径
    const currentRoute = routes.find((route) => {
      // 使用正则表达式替换动态参数（例如 :traceId）为通配符
      const routePattern = route.path.replace(/:\w+/g, '[^/]+') // 转换为 '/cause/report/[^/]+'
      const regex = new RegExp(`^${routePattern}$`) // 创建正则表达式
      return regex.test(pathname) // 使用正则测试
    })

    return currentRoute ? currentRoute.name : false
  }

  const getBreadcrumbs = (location) => {
    const breadcrumbs = []

    location.split('/').reduce((prev, curr, index, array) => {
      const currentPathname = `${prev}/${curr}`
      const routeName = getRouteName(currentPathname, routes)

      if (routeName) {
        breadcrumbs.push({
          href: index + 1 < array.length ? '/#' + currentPathname : null,
          title: routeName,
        })
      }

      return currentPathname
    })

    return breadcrumbs
  }

  const labelPath = getLabelPathFromMenuTree(menuItems, currentLocation);
  const routePath = getBreadcrumbs(currentLocation)
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
