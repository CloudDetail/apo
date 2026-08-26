// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { createHermesConnection } from './useHermesConnection.ts'

test('does not create a WebSocket-backed session client without an APO user id', async () => {
  const connection = createHermesConnection(null)

  assert.equal(connection.rpc, null)
  assert.equal(connection.sessionClient, null)
  await assert.rejects(() => connection.sessionApi.list(), /APO 用户信息尚未加载/)
})
