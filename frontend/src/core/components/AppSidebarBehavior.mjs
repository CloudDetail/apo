/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const parentMenuNavigationTarget = (menu = {}) => (
  menu.children?.length > 0 && menu.router?.to ? menu.router.to : ''
)

export const parentMenuTitleClickProps = (menu = {}, navigate = () => {}) => {
  const target = parentMenuNavigationTarget(menu)
  if (!target) return {}

  return {
    onTitleClick: () => navigate(target),
  }
}

export const BUILT_IN_DASHBOARD_ROUTES = Object.freeze({
  system: '/custom-dashboard/5242f1c8-14b7-5a76-a037-ba8a46039463',
  basic: '/custom-dashboard/ee0dea3e-9030-5bfb-8d3d-e0d7f3d063cb',
  hostResource: '/custom-dashboard/7ae6c1f4-ea59-57fc-b675-1398efa0bd87',
  application: '/custom-dashboard/2878dcff-dc5b-55ae-9bef-37152af692c3',
})

export const MIDDLEWARE_DASHBOARD_MENUS = Object.freeze([
  {
    key: 'middleware-service',
    label: '服务调用中间件',
    abbreviation: '服务调用',
    to: '/custom-dashboard/13f7b0b8-6e9c-5999-8a61-ab4003a1ab95',
  },
  {
    key: 'middleware-elasticsearch',
    label: 'Elasticsearch 指标',
    abbreviation: 'Elasticsearch',
    to: '/custom-dashboard/7c2204c5-0aa7-5f5b-8244-7888978cd0bc',
  },
  {
    key: 'middleware-kafka',
    label: 'Kafka 指标',
    abbreviation: 'Kafka',
    to: '/custom-dashboard/aa78fb07-aad0-5d46-aca2-89b9835bd067',
  },
  {
    key: 'middleware-mysql',
    label: 'MySQL 指标',
    abbreviation: 'MySQL',
    to: '/custom-dashboard/b8b226ed-af19-5199-8b70-50e823bcde52',
  },
  {
    key: 'middleware-rabbitmq-exporter',
    label: 'RabbitMQ Exporter 指标',
    abbreviation: 'RabbitMQ Exporter',
    to: '/custom-dashboard/66a35e74-bc94-5a10-aa82-4931c86ddd7e',
  },
  {
    key: 'middleware-rabbitmq-official',
    label: 'RabbitMQ 官方指标',
    abbreviation: 'RabbitMQ 官方',
    to: '/custom-dashboard/045d75e9-7c8a-5e13-b3f5-8877916a4dfe',
  },
  {
    key: 'middleware-redis',
    label: 'Redis 指标',
    abbreviation: 'Redis',
    to: '/custom-dashboard/8611296f-2b15-5556-ad8a-23fe6478377e',
  },
])

const collectMenuItemsByRoute = (menuItems = [], result = new Map()) => {
  menuItems.forEach((item) => {
    if (item?.router?.to) result.set(item.router.to, item)
    if (item?.children?.length) collectMenuItemsByRoute(item.children, result)
  })
  return result
}

const withDashboardRoute = (menu, to) => ({
  ...menu,
  router: {
    ...(menu.router || {}),
    to,
    hideTimeSelector: false,
  },
})

export const replaceBuiltInDashboardMenuRoutes = (menuItems = []) => {
  const itemsByRoute = collectMenuItemsByRoute(menuItems)
  const middlewareRoutes = new Set(MIDDLEWARE_DASHBOARD_MENUS.map((item) => item.to))
  const reservedRoutes = new Set([
    ...Object.values(BUILT_IN_DASHBOARD_ROUTES),
    ...middlewareRoutes,
  ])
  const hasMiddlewareMenu = menuItems.some((item) => item.key === 'middleware')

  const routedItems = menuItems
    .filter((item) => !(hasMiddlewareMenu && item.key === 'mysql'))
    .map((item) => {
      const directRoute = BUILT_IN_DASHBOARD_ROUTES[item.key]
      if (directRoute) return withDashboardRoute(item, directRoute)

      if (item.key === 'middleware') {
        return {
          ...item,
          router: undefined,
          children: MIDDLEWARE_DASHBOARD_MENUS.map((definition) => {
            const source = itemsByRoute.get(definition.to) || {}
            return withDashboardRoute(
              {
                ...source,
                key: definition.key,
                label: source.label || definition.label,
                abbreviation: source.abbreviation || definition.abbreviation,
                children: undefined,
              },
              definition.to,
            )
          }),
        }
      }

      if (item.key === 'customDashboard') {
        return {
          ...item,
          children: (item.children || []).filter(
            (child) => !reservedRoutes.has(child?.router?.to),
          ),
        }
      }

      return item
    })

  const hostRoute = BUILT_IN_DASHBOARD_ROUTES.hostResource
  const hostSource = itemsByRoute.get(hostRoute)
  if (!hostSource || routedItems.some((item) => item.key === 'hostResource')) {
    return routedItems
  }

  const hostMenu = withDashboardRoute(
    {
      ...hostSource,
      key: 'hostResource',
      label: '主机资源',
      abbreviation: '主机资源',
      children: undefined,
    },
    hostRoute,
  )
  const basicIndex = routedItems.findIndex((item) => item.key === 'basic')
  const insertAt = basicIndex >= 0 ? basicIndex + 1 : routedItems.length
  return [...routedItems.slice(0, insertAt), hostMenu, ...routedItems.slice(insertAt)]
}
