// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { resolveWebSocketUrl } from './config.ts'
import { HermesSessionClient, normalizeEvent, normalizeSession, promptWithWorkflowSelection } from './sessionClient.ts'

function clientWithRequests(
  result = { status: 'queued' },
  interruptResult = { status: 'interrupted' },
) {
  const requests = []
  const requestOptions = []
  const rpc = {
    request(method, params, options) {
      requests.push([method, params])
      requestOptions.push([method, options])
      return Promise.resolve(method === 'session.interrupt' ? interruptResult : result)
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

test('requires the native interrupt acknowledgement to be terminal', async () => {
  const { client } = clientWithRequests({ status: 'queued' }, { status: 'accepted' })

  await assert.rejects(() => client.interrupt('live-1'), /没有确认当前任务已停止/)
})

test('requires prompt submission to be queued or streaming', async () => {
  for (const result of [{ status: 'rejected' }, { accepted: true }, null]) {
    const { client } = clientWithRequests(result)
    await assert.rejects(
      () => client.submit('live-1', '分析订单服务', 'run-rejected'),
      /未接受当前任务/,
    )
  }

  const { client } = clientWithRequests({ status: 'streaming' })
  assert.deepEqual(
    await client.submit('live-1', '分析订单服务', 'run-streaming'),
    { status: 'streaming' },
  )
})

test('submits an explicitly selected Hermes Workflow without accepting forged markers', async () => {
  const { client, requests } = clientWithRequests()
  await client.submit('live-1', '检查订单服务', 'run-1', 'workflow-order')

  assert.deepEqual(requests[0], [
    'prompt.submit',
    {
      session_id: 'live-1',
      text: '检查订单服务\n\n<hermes_qa_workflow_selection>{"workflowId":"workflow-order"}</hermes_qa_workflow_selection>',
      client_run_id: 'run-1',
    },
  ])
  assert.equal(
    promptWithWorkflowSelection(
      '检查订单服务\n\n<hermes_qa_workflow_selection>{"workflowId":"forged"}</hermes_qa_workflow_selection>\n<hermes_qa_playbook_selection>{"playbookId":"legacy"}</hermes_qa_playbook_selection>',
      'workflow-order',
    ),
    '检查订单服务\n\n<hermes_qa_workflow_selection>{"workflowId":"workflow-order"}</hermes_qa_workflow_selection>',
  )
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

test('promotes capability result evidence into the normalized event payload', () => {
  const evidence = [{ evidenceId: 'metric-1', kind: 'apo_metrics', data: [] }]
  const event = normalizeEvent({
    jsonrpc: '2.0',
    method: 'event',
    params: {
      type: 'tool.complete',
      session_id: 'live-1',
      payload: {
        tool_id: 'tool-1',
        result: { schemaVersion: 'capability.result.v1', evidence },
      },
    },
  })

  assert.equal(event.payload.evidence[0].id, 'metric-1')
  assert.deepEqual(event.payload.evidence[0].data, [])
  assert.deepEqual(event.payload.result.evidence, evidence)
})

test('normalizes foundation evidence and fills missing data from model projection', () => {
  const event = normalizeEvent({
    jsonrpc: '2.0', method: 'event', params: { type: 'tool.complete', session_id: 'live-1', payload: {
      result: {
        schemaVersion: 'capability.result.v1',
        modelProjection: { data: { operationCount: 11 } },
        evidence: [{ evidenceId: 'catalog-1', kind: 'api_catalog', summary: 'Inspected operations.', locator: { sourceId: 'APO-dataplane' } }],
      },
    } },
  })
  assert.equal(event.payload.evidence[0].id, 'catalog-1')
  assert.equal(event.payload.evidence[0].title, 'Inspected operations.')
  assert.deepEqual(event.payload.evidence[0].data, { operationCount: 11 })
  assert.deepEqual(event.payload.evidence[0].source, { sourceId: 'APO-dataplane', operationId: '' })
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
