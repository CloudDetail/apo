/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const routeNameForPath = (pathname, routes = []) => {
  const currentRoute = routes.find((route) => {
    const routePattern = route.path.replace(/:\w+/g, '[^/]+')
    const regex = new RegExp(`^${routePattern}$`)
    return regex.test(pathname)
  })

  return currentRoute ? currentRoute.name : false
}

export const getRouteBreadcrumbs = (location, routes = []) => {
  const breadcrumbs = []

  location.split('/').reduce((prev, curr, index, array) => {
    const currentPathname = `${prev}/${curr}`
    const routeName = routeNameForPath(currentPathname, routes)

    if (routeName && breadcrumbs[breadcrumbs.length - 1]?.title !== routeName) {
      breadcrumbs.push({
        href: index + 1 < array.length ? '/#' + currentPathname : null,
        title: routeName,
      })
    }

    return currentPathname
  })

  return breadcrumbs
}

export const getMenuBreadcrumbs = (tree = [], route = '') => {
  let result = []

  const dfs = (nodes, path) => {
    if (!nodes || nodes.length === 0 || result.length > 0) return

    for (const node of nodes) {
      const nodeRoute = node.router?.to || ''
      const nextPath = [
        ...path,
        {
          href: nodeRoute && nodeRoute !== route ? '/#' + nodeRoute : null,
          title: node.label,
        },
      ]

      if (nodeRoute === route) {
        result = nextPath
        return
      }

      if (Array.isArray(node.children)) {
        dfs(node.children, nextPath)
      } else if (node.children && typeof node.children === 'object') {
        dfs([node.children], nextPath)
      }

      if (result.length > 0) return
    }
  }

  dfs(tree, [])

  return result
}
