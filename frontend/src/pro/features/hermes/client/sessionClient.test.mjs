// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { resolveWebSocketUrl } from './config.ts'
import { HermesSessionClient, normalizeEvent, normalizeSession } from './sessionClient.ts'

function clientWithRequests(result = { accepted: true }) {
  const requests = []
  const requestOptions = []
  const rpc = {
    request(method, params, options) {
      requests.push([method, params])
      requestOptions.push([method, options])
      return Promise.resolve(result)
    },
    subscribe() {
      return () => {}
    },
  }
  return { client: new HermesSessionClient(rpc), requests, requestOptions }
}

test('HermesSessionClient uses the native session JSON-RPC methods', async () => {
  const { client, requests, requestOptions } = clientWithRequests()
  await client.submit('live-1', '分析订单服务', 'run-1')
  await client.steer('live-1', '只看最近一小时')
  await client.interrupt('live-1')
  await client.respondApproval('live-1', 'approval-1', 'once')
  await client.respondClarification('clarify-1', '确认并继续')

  assert.deepEqual(requests, [
    ['prompt.submit', { session_id: 'live-1', text: '分析订单服务', client_run_id: 'run-1' }],
    ['session.steer', { session_id: 'live-1', text: '只看最近一小时' }],
    ['session.interrupt', { session_id: 'live-1' }],
    ['approval.respond', { session_id: 'live-1', request_id: 'approval-1', choice: 'once' }],
    ['clarify.respond', { request_id: 'clarify-1', answer: '确认并继续' }],
  ])
  assert.deepEqual(requestOptions, [
    ['prompt.submit', { timeoutMs: 10_000 }],
    ['session.steer', { timeoutMs: 5_000 }],
    ['session.interrupt', { timeoutMs: 5_000 }],
    ['approval.respond', { timeoutMs: 5_000 }],
    ['clarify.respond', { timeoutMs: 5_000 }],
  ])
})

test('normalizes native sessions and event notifications', () => {
  assert.deepEqual(
    normalizeSession({
      session_id: 'live-1',
      stored_session_id: 'stored-1',
      info: { title: '订单分析' },
      running: true,
    }),
    {
      liveSessionId: 'live-1',
      storedSessionId: 'stored-1',
      title: '订单分析',
      running: true,
    },
  )
  assert.deepEqual(
    normalizeEvent({
      jsonrpc: '2.0',
      method: 'event',
      params: {
        type: 'clarify.request',
        session_id: 'live-1',
        payload: { request_id: 'request-1', question: '确认计划？' },
      },
    }),
    {
      type: 'clarify.request',
      sessionId: 'live-1',
      payload: { request_id: 'request-1', question: '确认计划？' },
    },
  )
})

test('resolves relative websocket proxy paths against the current APO origin', () => {
  assert.equal(
    resolveWebSocketUrl('/hermes-ws', {
      protocol: 'http:',
      host: '127.0.0.1:3000',
    }),
    'ws://127.0.0.1:3000/hermes-ws',
  )
})
