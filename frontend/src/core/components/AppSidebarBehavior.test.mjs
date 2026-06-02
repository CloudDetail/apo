/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import test from 'node:test'
import assert from 'node:assert/strict'
import {
  parentMenuNavigationTarget,
  parentMenuTitleClickProps,
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
