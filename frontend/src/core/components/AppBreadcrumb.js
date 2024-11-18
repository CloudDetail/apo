import React from 'react'
import { useLocation } from 'react-router-dom'

import routes from 'src/routes'

import { Breadcrumb } from 'antd'

const AppBreadcrumb = () => {
  const currentLocation = useLocation().pathname
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)

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

  const breadcrumbs = getBreadcrumbs(currentLocation)
  return (
    <Breadcrumb items={getBreadcrumbs(currentLocation)} className="text-base" />
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
