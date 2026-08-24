import test from 'node:test'
import assert from 'node:assert/strict'

import { canManageSystemHealthConfig } from './systemHealthPermissions.mjs'

test('system health config is restricted to admin users', () => {
  assert.equal(canManageSystemHealthConfig(null), false)
  assert.equal(canManageSystemHealthConfig({ roleList: [{ roleName: 'viewer' }] }), false)
  assert.equal(canManageSystemHealthConfig({ roleList: [{ roleName: 'admin' }] }), true)
})

test('system health config supports the local admin preview shape', () => {
  assert.equal(canManageSystemHealthConfig({ role: 'admin' }), true)
  assert.equal(canManageSystemHealthConfig({ roleList: 'admin' }), true)
})
