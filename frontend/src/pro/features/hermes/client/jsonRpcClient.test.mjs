// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { JsonRpcClient, JsonRpcError, JsonRpcTimeoutError } from './jsonRpcClient.ts'

class MockSocket {
  readyState = 0
  listeners = new Map()
  sent = []

  addEventListener(type, listener) {
    const listeners = this.listeners.get(type) || []
    listeners.push(listener)
    this.listeners.set(type, listeners)
  }

  send(value) {
    this.sent.push(value)
  }

  close() {
    this.readyState = 3
    this.emit('close', {})
  }

  emit(type, event) {
    if (type === 'open') this.readyState = 1
    for (const listener of this.listeners.get(type) || []) listener(event)
  }
}

test('JsonRpcClient sends requests and resolves matching responses', async () => {
  const socket = new MockSocket()
  const client = new JsonRpcClient('ws://localhost/hermes', () => socket)
  const responsePromise = client.request('session.create', { source: 'apo-front' })
  socket.emit('open', {})
  await Promise.resolve()

  const request = JSON.parse(socket.sent[0])
  assert.equal(request.method, 'session.create')
  assert.deepEqual(request.params, { source: 'apo-front' })

  socket.emit('message', {
    data: JSON.stringify({ jsonrpc: '2.0', id: request.id, result: { session_id: 'live-1' } }),
  })
  assert.deepEqual(await responsePromise, { session_id: 'live-1' })
  client.close()
})

test('JsonRpcClient forwards notifications and preserves JSON-RPC errors', async () => {
  const socket = new MockSocket()
  const client = new JsonRpcClient('ws://localhost/hermes', () => socket)
  const notifications = []
  client.subscribe((notification) => notifications.push(notification))
  const responsePromise = client.request('session.history')
  socket.emit('open', {})
  await Promise.resolve()
  const request = JSON.parse(socket.sent[0])

  socket.emit('message', {
    data: JSON.stringify({
      jsonrpc: '2.0',
      method: 'event',
      params: { type: 'message.delta', payload: { text: '处理中' } },
    }),
  })
  socket.emit('message', {
    data: JSON.stringify({
      jsonrpc: '2.0',
      id: request.id,
      error: { code: 4007, message: 'session missing' },
    }),
  })

  assert.equal(notifications[0].method, 'event')
  await assert.rejects(responsePromise, (reason) => {
    assert.ok(reason instanceof JsonRpcError)
    assert.equal(reason.payload.code, 4007)
    return true
  })
  client.close()
})

test('JsonRpcClient rejects timed out requests and ignores late responses', async () => {
  const socket = new MockSocket()
  const client = new JsonRpcClient('ws://localhost/hermes', () => socket)
  const responsePromise = client.request('session.resume', {}, { timeoutMs: 10 })
  socket.emit('open', {})
  await Promise.resolve()
  const request = JSON.parse(socket.sent[0])

  await assert.rejects(responsePromise, (reason) => {
    assert.ok(reason instanceof JsonRpcTimeoutError)
    assert.equal(reason.method, 'session.resume')
    assert.equal(reason.timeoutMs, 10)
    return true
  })

  socket.emit('message', {
    data: JSON.stringify({ jsonrpc: '2.0', id: request.id, result: { accepted: true } }),
  })
  client.close()
})
