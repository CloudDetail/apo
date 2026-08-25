// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { hermesWebSocketUrl } from './config.ts'
import {
  canonicalHermesUserId,
  HermesUserScopeGeneration,
  requireHermesUserId,
} from './userScope.ts'

const apoLocation = {
  protocol: 'https:',
  host: 'apo.example',
}

test('preserves the trimmed APO user id without rewriting it', () => {
  assert.equal(canonicalHermesUserId('  user-001  '), 'user-001')
  assert.equal(canonicalHermesUserId('001'), '001')
  assert.equal(canonicalHermesUserId(101), '101')
  assert.equal(canonicalHermesUserId('a.b_c-9'), 'a.b_c-9')
})

test('enforces the Hermes URL-safe owner contract and reserved local owner', () => {
  assert.equal(canonicalHermesUserId(undefined), null)
  assert.equal(canonicalHermesUserId(''), null)
  assert.equal(canonicalHermesUserId('   '), null)
  assert.equal(canonicalHermesUserId('local-host'), null)
  assert.equal(canonicalHermesUserId('_user'), null)
  assert.equal(canonicalHermesUserId('user/name'), null)
  assert.equal(canonicalHermesUserId('a'.repeat(65)), null)
  assert.equal(canonicalHermesUserId(Number.NaN), null)
  assert.throws(() => requireHermesUserId('local-host'), /APO 用户信息尚未加载/)
})

test('requires a user id before building a Hermes WebSocket URL', () => {
  assert.throws(() => hermesWebSocketUrl(undefined, apoLocation), /APO 用户信息尚未加载/)
  assert.throws(() => hermesWebSocketUrl(' ', apoLocation), /APO 用户信息尚未加载/)

  const url = new URL(hermesWebSocketUrl(' user-001 ', apoLocation))
  assert.equal(url.href, 'wss://apo.example/hermes-ws?user_id=user-001')
  assert.equal(url.searchParams.get('user_id'), 'user-001')
})

test('invalidates delayed work when either user scope or API identity changes', () => {
  const guard = new HermesUserScopeGeneration()
  const apiA = { owner: 'A' }
  const replacementApiA = { owner: 'A' }
  const apiB = { owner: 'B' }

  guard.update('user-A')
  const requestA = guard.capture(apiA)
  assert.equal(guard.isCurrent(requestA, apiA), true)
  assert.equal(guard.isCurrent(requestA, replacementApiA), false)

  guard.update('user-B')
  assert.equal(guard.isCurrent(requestA, apiA), false)
  const requestB = guard.capture(apiB)
  assert.equal(guard.isCurrent(requestB, apiB), true)

  guard.update(null)
  assert.equal(guard.isCurrent(requestB, apiB), false)
  assert.throws(() => guard.capture(apiB), /APO 用户信息尚未加载/)
})
