/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import test from 'node:test'
import assert from 'node:assert/strict'
import * as breadcrumbBehavior from './AppBreadcrumbBehavior.mjs'

const { getRouteBreadcrumbs } = breadcrumbBehavior

test('custom dashboard edit breadcrumbs skip duplicate parent and detail labels', () => {
  const routes = [
    { path: '/custom-dashboard', name: '我的大盘' },
    { path: '/custom-dashboard/:id', name: '我的大盘' },
    { path: '/custom-dashboard/:id/edit', name: '大盘编辑' },
  ]

  assert.deepEqual(
    getRouteBreadcrumbs('/custom-dashboard/107bd858-663e-4d75-85d2-33ed168f9de4/edit', routes),
    [
      { href: '/#/custom-dashboard', title: '我的大盘' },
      { href: null, title: '大盘编辑' },
    ],
  )
})

test('menu breadcrumbs link parent routes for exact child dashboard pages', () => {
  const menuItems = [
    {
      label: '我的大盘',
      router: { to: '/custom-dashboard' },
      children: [
        {
          label: '集群总览',
          router: { to: '/custom-dashboard/dash-1' },
        },
      ],
    },
  ]
  const getMenuBreadcrumbs = breadcrumbBehavior.getMenuBreadcrumbs || (() => [])

  assert.deepEqual(
    getMenuBreadcrumbs(menuItems, '/custom-dashboard/dash-1'),
    [
      { href: '/#/custom-dashboard', title: '我的大盘' },
      { href: null, title: '集群总览' },
    ],
  )
})
