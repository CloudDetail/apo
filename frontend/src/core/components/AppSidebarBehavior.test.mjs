/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import test from 'node:test'
import assert from 'node:assert/strict'
import {
  BUILT_IN_DASHBOARD_ROUTES,
  MIDDLEWARE_DASHBOARD_MENUS,
  parentMenuNavigationTarget,
  parentMenuTitleClickProps,
  replaceBuiltInDashboardMenuRoutes,
} from './AppSidebarBehavior.mjs'

test('parent menu with child dashboards can navigate to its own list route', () => {
  const menu = {
    key: 'customDashboard',
    router: { to: '/custom-dashboard' },
    children: [{ key: 'cd-one' }],
  }
  let navigatedTo = ''

  const props = parentMenuTitleClickProps(menu, (target) => {
    navigatedTo = target
  })

  assert.equal(parentMenuNavigationTarget(menu), '/custom-dashboard')
  assert.equal(typeof props.onTitleClick, 'function')
  props.onTitleClick()
  assert.equal(navigatedTo, '/custom-dashboard')
})

test('menus without both children and own route do not get title navigation', () => {
  assert.deepEqual(parentMenuTitleClickProps({ router: { to: '/service' } }, () => {}), {})
  assert.deepEqual(parentMenuTitleClickProps({ children: [{ key: 'child' }] }, () => {}), {})
})

test('built-in dashboard entries use native custom dashboard routes and middleware submenu', () => {
  const customDashboardChildren = [
    ...Object.entries(BUILT_IN_DASHBOARD_ROUTES).map(([key, to]) => ({
      key: `custom-${key}`,
      label: `custom ${key}`,
      router: { to },
    })),
    ...MIDDLEWARE_DASHBOARD_MENUS.map((item) => ({
      key: `custom-${item.key}`,
      label: `后端名称 ${item.key}`,
      router: { to: item.to },
    })),
    {
      key: 'custom-unrelated',
      label: '其他大盘',
      router: { to: '/custom-dashboard/unrelated' },
    },
  ]
  const originalMenus = [
    { key: 'system', router: { to: '/system-dashboard', page: { type: 'grafana' } } },
    { key: 'basic', router: { to: '/basic-dashboard' } },
    { key: 'application', router: { to: '/application-dashboard' } },
    { key: 'middleware', router: { to: '/middleware-dashboard' } },
    { key: 'mysql', router: { to: '/mysql-dashboard' } },
    { key: 'customDashboard', children: customDashboardChildren },
  ]

  const result = replaceBuiltInDashboardMenuRoutes(originalMenus)

  for (const [key, to] of Object.entries(BUILT_IN_DASHBOARD_ROUTES)) {
    assert.equal(result.find((item) => item.key === key)?.router?.to, to)
  }
  const hostIndex = result.findIndex((item) => item.key === 'hostResource')
  assert.equal(hostIndex, result.findIndex((item) => item.key === 'basic') + 1)
  assert.equal(result[hostIndex].label, '主机资源')
  assert.equal(result[hostIndex].abbreviation, '主机资源')
  assert.equal(result.some((item) => item.key === 'mysql'), false)

  const middleware = result.find((item) => item.key === 'middleware')
  assert.equal(middleware.router, undefined)
  assert.deepEqual(
    middleware.children.map((item) => item.router.to),
    MIDDLEWARE_DASHBOARD_MENUS.map((item) => item.to),
  )
  assert.deepEqual(
    middleware.children.map((item) => item.label),
    MIDDLEWARE_DASHBOARD_MENUS.map((item) => `后端名称 ${item.key}`),
  )
  assert.equal(new Set(middleware.children.map((item) => item.key)).size, middleware.children.length)

  const customDashboard = result.find((item) => item.key === 'customDashboard')
  assert.deepEqual(customDashboard.children.map((item) => item.key), ['custom-unrelated'])
  assert.equal(originalMenus[0].router.to, '/system-dashboard')
  assert.equal(originalMenus.find((item) => item.key === 'middleware').children, undefined)
})

test('standalone MySQL menu remains when middleware menu is unavailable', () => {
  const menus = [{ key: 'mysql', router: { to: '/mysql-dashboard' } }]
  assert.deepEqual(replaceBuiltInDashboardMenuRoutes(menus), menus)
})

test('host resource menu is hidden until its built-in dashboard is available', () => {
  const menus = [
    { key: 'basic', router: { to: '/basic-dashboard' } },
    { key: 'customDashboard', children: [] },
  ]
  assert.equal(
    replaceBuiltInDashboardMenuRoutes(menus).some((item) => item.key === 'hostResource'),
    false,
  )
})
