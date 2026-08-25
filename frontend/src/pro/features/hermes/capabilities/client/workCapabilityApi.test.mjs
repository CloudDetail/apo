// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import {
  diffWorkflowUpdate,
  HermesWorkflowUnavailableError,
  normalizeWorkflowUpdate,
  toCapabilityProjection,
  toWorkflow,
  toWorkflowEditValues,
  WorkCapabilityApi,
} from './workCapabilityApi.ts'

function response(value, status = 200) {
  return new globalThis.Response(JSON.stringify(value), {
    status,
    headers: { 'content-type': 'application/json' },
  })
}

const readUserId = () => 101

const stepMetric = {
  id: 'query_metrics',
  purpose: '查询订单错误率',
  providerRef: 'openapi:apo',
  capabilityRef: 'api.operation.apo.metric-query@1',
  inputMap: { service: '$inputs.service' },
  expectedOutput: '订单服务错误率时序',
  effects: ['read'],
}

const stepLog = {
  id: 'query_logs',
  purpose: '核对错误日志',
  providerRef: 'openapi:apo',
  capabilityRef: 'api.operation.apo.log-query@1',
  inputMap: { service: '$inputs.service' },
  expectedOutput: '关联错误日志',
  effects: ['read'],
}

const detailFixture = {
  schemaVersion: 'hermes-qa.workflow.v1',
  id: 'workflow-order',
  name: '订单错误排查',
  goal: '定位订单服务错误率升高原因',
  examples: ['检查订单服务错误率', '订单接口为什么报错'],
  inputs: [
    { name: '服务', required: true, source: 'user' },
    { name: '时间范围', required: false, source: 'conversation_or_prompt' },
  ],
  steps: [stepMetric, stepLog],
  notes: ['确认错误集中时间段'],
  effects: ['read'],
  expectedResult: 'incident_summary',
  status: 'active',
  pinned: true,
  useCount: 4,
  successCount: 3,
  recentFailures: 1,
  lastUsedAt: '2026-08-19T10:00:00Z',
  version: 3,
  sourceRunId: 'run-source',
  lastRunStatus: 'succeeded',
  recentRuns: [
    { runId: 'run-2', sessionId: 'session-2', relationship: 'reused', status: 'succeeded', createdAt: '2026-08-19T10:00:00Z' },
    { runId: 'run-source', sessionId: 'session-1', relationship: 'created', status: 'succeeded', createdAt: '2026-08-18T10:00:00Z' },
  ],
  versions: [
    { version: 3, digest: 'digest-3', created_by: 'user', source_run_id: null, created_at: '2026-08-19T10:00:00Z' },
    { version: 2, digest: 'digest-2', created_by: 'learning', source_run_id: 'run-source', created_at: '2026-08-18T10:00:00Z' },
  ],
  createdAt: '2026-08-18T09:00:00Z',
  updatedAt: '2026-08-19T10:00:00Z',
}

test('normalizes the current Hermes Workflow read model without inventing totals', () => {
  const workflow = toWorkflow(detailFixture)

  assert.equal(workflow.id, 'workflow-order')
  assert.equal(workflow.currentVersion, 3)
  assert.equal(workflow.inputs[0].source, 'user')
  assert.equal(workflow.steps[0].purpose, '查询订单错误率')
  assert.equal(workflow.steps[0].capabilityRef, 'api.operation.apo.metric-query@1')
  assert.deepEqual(workflow.steps[0].inputMap, { service: '$inputs.service' })
  assert.equal(workflow.reuseCount, 4)
  assert.equal(workflow.successCount, 3)
  assert.equal(workflow.recentFailures, 1)
  assert.equal(workflow.lastRunStatus, 'succeeded')
  assert.equal(workflow.successRate, null)
  assert.equal(workflow.runs[0].runId, 'run-2')
  assert.equal(workflow.versions[0].createdBy, 'user')
  assert.deepEqual(workflow.allowedActions, ['use', 'edit', 'disable', 'archive', 'unpin', 'delete'])
})

test('disabled and archived Workflows cannot be used', () => {
  const disabled = toWorkflow({ ...detailFixture, status: 'disabled', pinned: false })
  const archived = toWorkflow({ ...detailFixture, status: 'archived' })

  assert.equal(disabled.allowedActions.includes('use'), false)
  assert.equal(disabled.allowedActions.includes('enable'), true)
  assert.deepEqual(archived.allowedActions, ['restore', 'delete'])
})

test('unhealthy active and unknown Workflows fail closed', () => {
  const repeatedFailures = toWorkflow({ ...detailFixture, recentFailures: 3 })
  const blockedLastRun = toWorkflow({ ...detailFixture, recentFailures: 0, lastRunStatus: 'blocked' })
  const unknown = toWorkflow({ ...detailFixture, status: 'unexpected' })

  assert.equal(repeatedFailures.allowedActions.includes('use'), false)
  assert.equal(repeatedFailures.actionBlockers[0]?.code, 'workflow_unhealthy')
  assert.equal(blockedLastRun.allowedActions.includes('use'), false)
  assert.deepEqual(unknown.allowedActions, [])
})

test('list requests archived Workflows and filters the Workflow envelope locally', async () => {
  const calls = []
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async (input) => {
      calls.push(String(input))
      return response({
        workflows: [
          detailFixture,
          { ...detailFixture, id: 'workflow-disabled', name: '库存检查', status: 'disabled' },
        ],
      })
    },
  })

  const result = await api.list({ search: '订单', status: 'active', page: 1, pageSize: 10 })
  assert.deepEqual(calls, ['/hermes-api/workflows?includeArchived=true'])
  assert.equal(result.total, 1)
  assert.equal(result.items[0].id, 'workflow-order')
})

test('paginates locally without sending unsupported query parameters', async () => {
  const calls = []
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async (input) => {
      calls.push(String(input))
      return response({ workflows: [
        detailFixture,
        { ...detailFixture, id: 'workflow-2' },
        { ...detailFixture, id: 'workflow-3' },
      ] })
    },
  })

  const result = await api.list({ page: 2, pageSize: 2 })
  assert.deepEqual(calls, ['/hermes-api/workflows?includeArchived=true'])
  assert.equal(result.total, 3)
  assert.deepEqual(result.items.map((item) => item.id), ['workflow-3'])
})

test('reads versions and recent runs from Workflow detail', async () => {
  const calls = []
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async (input) => {
      calls.push(String(input))
      return response(detailFixture)
    },
  })

  const versions = await api.versions('workflow-order')
  const runs = await api.runs('workflow-order', { page: 1, pageSize: 1 })

  assert.deepEqual(calls, [
    '/hermes-api/workflows/workflow-order',
    '/hermes-api/workflows/workflow-order',
  ])
  assert.deepEqual(versions.map((item) => item.version), [3, 2])
  assert.equal(runs.total, 2)
  assert.deepEqual(runs.items.map((item) => item.runId), ['run-2'])
})

test('uses the migrated Workflow action and projection routes', async () => {
  const calls = []
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async (input, init = {}) => {
      const url = String(input)
      calls.push([url, init.method || 'GET', init.body])
      if (url.includes('/workflow-runs/')) {
        return response({
          runId: 'run/2',
          nodes: [],
          evidence: [],
          result: { status: 'succeeded', summary: '已完成', confidence: 'high', nextAction: null, evidenceRefs: [] },
          hasPlan: false,
          eventCount: 0,
        })
      }
      if (url.endsWith(':archive')) return response({ ...detailFixture, status: 'archived' })
      return response({ ...detailFixture, status: init.body && JSON.parse(init.body).enabled === false ? 'disabled' : 'active' })
    },
  })

  await api.setEnabled('workflow-order', false)
  await api.setPinned('workflow-order', false)
  await api.setArchived('workflow-order', true)
  await api.setArchived('workflow-order', false)
  await api.restore('workflow-order', 2)
  await api.projection('run/2')

  assert.deepEqual(calls, [
    ['/hermes-api/workflows/workflow-order', 'PUT', '{"enabled":false}'],
    ['/hermes-api/workflows/workflow-order:unpin', 'POST', undefined],
    ['/hermes-api/workflows/workflow-order:archive', 'POST', undefined],
    ['/hermes-api/workflows/workflow-order:restore', 'POST', undefined],
    ['/hermes-api/workflows/workflow-order/versions/2:restore', 'POST', undefined],
    ['/hermes-api/workflow-runs/run%2F2', 'GET', undefined],
  ])
})

test('updates camelCase Workflow fields and preserves complete step bindings', async () => {
  const calls = []
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async (input, init = {}) => {
      const body = JSON.parse(String(init.body))
      calls.push([String(input), init.method, body])
      return response({ ...detailFixture, ...body })
    },
  })
  const inputs = [{ name: '服务名称', required: true, source: 'user' }]
  const steps = [{ ...stepLog, purpose: '先核对日志' }, stepMetric]

  const updated = await api.update('workflow/order', {
    name: ' 订单链路排查 ',
    goal: '定位订单链路错误',
    examples: ['检查订单链路', '检查订单链路'],
    inputs,
    steps,
  })

  assert.deepEqual(calls, [[
    '/hermes-api/workflows/workflow%2Forder',
    'PUT',
    {
      name: '订单链路排查',
      goal: '定位订单链路错误',
      examples: ['检查订单链路'],
      inputs,
      steps,
    },
  ]])
  assert.equal(updated.name, '订单链路排查')
  assert.equal(updated.steps[0].id, 'query_logs')
  assert.equal(updated.steps[0].providerRef, 'openapi:apo')
})

test('only submits changed fields when other content is redacted for display', async () => {
  const calls = []
  const sensitiveFixture = {
    ...detailFixture,
    goal: '定位故障，token: abcdefghijklmnop',
    steps: [
      { ...stepMetric, purpose: '使用 password: p@ssword-value 查询指标' },
      stepLog,
    ],
  }
  const capability = toWorkflow(sensitiveFixture)
  const initial = toWorkflowEditValues(capability)
  const update = diffWorkflowUpdate(initial, { ...initial, name: '订单链路排查' })
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async (input, init = {}) => {
      calls.push([String(input), init.method, JSON.parse(String(init.body))])
      return response({ ...sensitiveFixture, name: '订单链路排查' })
    },
  })

  assert.equal(initial.goal, '定位故障，token: [已隐藏]')
  assert.match(initial.steps[0].purpose, /password: \[已隐藏\]/)
  assert.deepEqual(update, { name: '订单链路排查' })
  await api.update(capability.id, update)
  assert.deepEqual(calls, [[
    '/hermes-api/workflows/workflow-order',
    'PUT',
    { name: '订单链路排查' },
  ]])
})

test('step reordering restores unchanged raw values and keeps bindings attached', () => {
  const capability = toWorkflow({
    ...detailFixture,
    steps: [
      { ...stepMetric, purpose: '使用 token: abcdefghijklmnop 查询指标' },
      stepLog,
    ],
  })
  const initial = toWorkflowEditValues(capability)
  const update = diffWorkflowUpdate(initial, {
    ...initial,
    steps: [initial.steps[1], initial.steps[0]],
  })

  assert.equal(update.steps[1].purpose, '使用 token: abcdefghijklmnop 查询指标')
  assert.equal(update.steps[1].id, 'query_metrics')
  assert.equal(update.steps[1].capabilityRef, 'api.operation.apo.metric-query@1')
  assert.deepEqual(update.steps[1].inputMap, { service: '$inputs.service' })
})

test('rejects incomplete Workflow steps before sending them to Hermes', () => {
  assert.throws(
    () => normalizeWorkflowUpdate({ steps: [{ ...stepMetric }] }),
    (error) => error?.code === 'invalid_steps',
  )
  assert.throws(
    () => normalizeWorkflowUpdate({ steps: [{ ...stepMetric, capabilityRef: null }, stepLog] }),
    (error) => error?.code === 'invalid_step_binding',
  )
  assert.throws(
    () => normalizeWorkflowUpdate({ steps: [{ ...stepMetric, expectedOutput: '' }, stepLog] }),
    (error) => error?.code === 'invalid_update',
  )
})

test('never writes a changed redaction placeholder back to Hermes', () => {
  assert.throws(
    () => normalizeWorkflowUpdate({ goal: '定位故障，token: [已隐藏]' }),
    (error) => error?.code === 'redacted_value',
  )
})

test('uses single and atomic bulk delete contracts', async () => {
  const calls = []
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async (input, init = {}) => {
      calls.push([String(input), init.method, init.body ? JSON.parse(String(init.body)) : undefined])
      const ids = init.body ? JSON.parse(String(init.body)).workflowIds : ['workflow/1']
      return response({ deletedIds: ids, deletedCount: ids.length })
    },
  })

  await api.delete('workflow/1')
  await api.deleteMany(['workflow-1', 'workflow-2', 'workflow-1'])

  assert.deepEqual(calls, [
    ['/hermes-api/workflows/workflow%2F1', 'DELETE', undefined],
    ['/hermes-api/workflows:delete', 'POST', { workflowIds: ['workflow-1', 'workflow-2'] }],
  ])
})

test('surfaces a running Workflow delete conflict', async () => {
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async () => response({ detail: '运行中的任务方案不能删除，请等待运行结束。' }, 409),
  })

  await assert.rejects(
    () => api.delete('workflow-order'),
    (error) => error?.code === 'conflict' && error?.status === 409 && error.message.includes('运行中'),
  )
})

test('reports Hermes connectivity failures without mentioning an APO Adapter', async () => {
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async () => { throw new Error('connection refused') },
  })

  await assert.rejects(
    () => api.list(),
    (error) => error instanceof HermesWorkflowUnavailableError
      && error.code === 'hermes_workflow_unavailable'
      && !error.message.includes('Adapter'),
  )
})

test('keeps a missing Workflow distinct from a connection failure', async () => {
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async () => response({ detail: '任务方案不存在。' }, 404),
  })

  await assert.rejects(
    () => api.get('missing'),
    (error) => error?.code === 'not_found' && error?.message === '任务方案不存在。',
  )
})

test('adds the current APO user id to every Workflow request', async () => {
  const calls = []
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId,
    request: async (_input, init) => {
      calls.push('request')
      assert.equal(init.credentials, 'include')
      assert.equal(new Headers(init.headers).has('Authorization'), false)
      assert.equal(new Headers(init.headers).get('X-Hermes-User-Id'), '101')
      return response({ workflows: [] })
    },
  })

  await api.list()
  await api.list()
  assert.deepEqual(calls, ['request', 'request'])
})

test('does not send a Workflow request before the APO user id is available', async () => {
  let requested = false
  const api = new WorkCapabilityApi({
    baseUrl: '/hermes-api',
    readUserId: () => undefined,
    request: async () => {
      requested = true
      return response({ workflows: [] })
    },
  })

  await assert.rejects(() => api.list(), /APO 用户信息尚未加载/)
  assert.equal(requested, false)
})

test('normalizes the current Workflow projection result, nodes and evidence', () => {
  const result = toCapabilityProjection({
    runId: 'run-1',
    profile: 'hermes_qa',
    displayType: 'workflow_graph',
    goal: '检查订单服务',
    revision: 2,
    hasPlan: true,
    eventCount: 8,
    nodes: [{
      id: 'stage-query',
      title: '查询指标',
      status: 'failed',
      current: false,
      requiresUserAction: false,
      dependencies: [],
      facts: ['password: p@ssword-value'],
      gaps: ['缺少 token: abcdefghijklmnop'],
      failure: 'credential=abcdefghijklmno',
    }],
    evidence: [{
      evidenceId: 'evidence-metric',
      stepId: 'stage-query',
      kind: 'apo_metrics',
      summary: '错误率已恢复',
      trust: 'verified_read',
      receiptId: 'receipt-1',
    }],
    result: {
      status: 'partial',
      summary: '部分指标已返回',
      confidence: 'medium',
      nextAction: '补充实例范围',
      evidenceRefs: ['evidence-metric'],
    },
    workflow: {
      id: 'workflow-order',
      name: '订单错误排查',
      version: 3,
      relationship: 'reused',
      executionMode: 'workflow_reuse',
    },
    workflowRun: {
      runId: 'run-1',
      sessionId: 'session-1',
      status: 'partial',
      stateRevision: 4,
    },
  })

  assert.equal(result.status, 'partial')
  assert.equal(result.hasPlan, true)
  assert.equal(result.revision, 2)
  assert.equal(result.nodes?.[0].title, '查询指标')
  assert.equal(result.nodes?.[0].facts[0], 'password: [已隐藏]')
  assert.equal(result.evidence?.[0].id, 'evidence-metric')
  assert.equal(result.evidence?.[0].trust, 'verified_read')
  assert.equal(result.evidence?.[0].receiptId, 'receipt-1')
  assert.equal(result.result?.confidence, 'medium')
  assert.equal(result.workflow?.version, 3)
  assert.match(String(result.gaps?.[0]), /缺少 token: \[已隐藏\]/)
  assert.match(String(result.failures?.[0]), /credential=\[已隐藏\]/)
})

test('uses Workflow Run status while a projection result is still pending', () => {
  const result = toCapabilityProjection({
    runId: 'run-running',
    hasPlan: false,
    nodes: [],
    evidence: [{
      evidenceId: 'evidence-auth',
      stepId: 'authorize',
      kind: 'authorization',
      summary: 'password: p@ssword-value',
      trust: 'declared',
    }],
    result: null,
    workflowRun: {
      runId: 'run-running',
      status: 'running',
      outcome: null,
    },
  })

  assert.equal(result.status, 'running')
  assert.equal(result.hasPlan, false)
  assert.equal(result.evidence?.[0].summary, 'password: [已隐藏]')
  assert.equal(result.evidence?.[0].receiptId, undefined)
})
