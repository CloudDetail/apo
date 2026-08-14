// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { historyFromGateway, historyFromPersistedEvents } from './history.ts'

function record(sequenceNo, direction, eventType, payload, runId = 'run-1') {
  return {
    eventId: `event-${sequenceNo}`,
    sequenceNo,
    direction,
    eventType,
    runId,
    payload,
    occurredAt: new Date(Date.UTC(2026, 7, 11, 8, 0, sequenceNo)).toISOString(),
  }
}

function planToolPayload(id, titles) {
  return {
    tool_id: id,
    name: 'task_flow_orchestration_nodes_upsert',
    args: {
      data: {
        core: {
          nodes: titles.map((title, index) => ({ id: `node-${index + 1}`, title })),
        },
      },
    },
  }
}

test('restores messages and execution activities in persisted event order', () => {
  const snapshot = historyFromPersistedEvents([
    record(5, 'gateway_to_browser', 'message.complete', { text: '最终结论' }),
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '分析订单服务' }),
    record(4, 'gateway_to_browser', 'tool.completed', {
      tool_id: 'tool-1',
      name: 'query_logs',
      result: { total: 3 },
    }),
    record(2, 'gateway_to_browser', 'reasoning.delta', { text: '先确认影响范围' }),
    record(3, 'gateway_to_browser', 'tool.start', {
      tool_id: 'tool-1',
      name: 'query_logs',
      args: { service: 'order' },
    }),
  ])

  assert.deepEqual(
    snapshot.timeline.map((item) => item.kind),
    ['message', 'run_status', 'activity', 'activity', 'message'],
  )
  assert.equal(snapshot.timeline[0].message.content, '分析订单服务')
  assert.equal(snapshot.timeline[2].activity.kind, 'commentary')
  assert.equal(snapshot.timeline[3].activity.status, 'completed')
  assert.equal(snapshot.timeline[4].message.content, '最终结论')
  assert.equal(snapshot.timeline[1].status, 'completed')
})

test('restores persisted evidence on its original tool activity', () => {
  const snapshot = historyFromPersistedEvents([
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '查询日志' }, 'run-evidence'),
    record(2, 'gateway_to_browser', 'tool.start', { tool_id: 'tool-evidence', name: 'query_logs' }, 'run-evidence'),
    record(3, 'gateway_to_browser', 'tool.completed', {
      tool_id: 'tool-evidence',
      name: 'query_logs',
      evidence: [{ id: 'logs-1', kind: 'apo_logs', title: '错误日志', source: { sourceId: 'system-apo', operationId: 'queryLogs' }, data: [] }],
    }, 'run-evidence'),
  ])
  const activity = snapshot.timeline.find((item) => item.kind === 'activity')
  assert.equal(activity.activity.evidence[0].id, 'logs-1')
  assert.equal(activity.runId, 'run-evidence')
})

test('keeps a resolved confirmation card at its original timeline position', () => {
  const snapshot = historyFromPersistedEvents([
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '检查异常' }),
    record(2, 'gateway_to_browser', 'clarify.request', {
      request_id: 'clarify-1',
      question: '只检查订单服务吗？',
      choices: ['确认并继续', '修改范围'],
    }),
    record(3, 'browser_to_gateway', 'clarify.respond', { label: '确认并继续' }),
    record(4, 'gateway_to_browser', 'message.complete', { text: '检查完成' }),
  ])

  assert.deepEqual(
    snapshot.timeline.map((item) => item.kind),
    ['message', 'run_status', 'checkpoint', 'message', 'message'],
  )
  assert.equal(snapshot.timeline[2].resolved, true)
  assert.equal(snapshot.timeline[2].responseText, '确认并继续')
  assert.equal(snapshot.timeline[3].message.role, 'user')
  assert.equal(snapshot.timeline[4].message.role, 'assistant')
})

test('restores every model-generated plan without labeling plan versions', () => {
  const snapshot = historyFromPersistedEvents([
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '分析服务延时' }),
    record(
      2,
      'gateway_to_browser',
      'tool.start',
      planToolPayload('plan-tool-1', ['查询服务 RED 指标', '查询 eBPF 指标']),
    ),
    record(
      3,
      'gateway_to_browser',
      'tool.completed',
      planToolPayload('plan-tool-1', ['查询服务 RED 指标', '查询 eBPF 指标']),
    ),
    record(4, 'gateway_to_browser', 'clarify.request', {
      request_id: 'plan-1',
      question: '模型已完成规划，请确认是否继续。',
      choices: ['确认并继续', '修改方案'],
    }),
    record(5, 'browser_to_gateway', 'clarify.respond', { answer: '修改方案' }),
    record(6, 'gateway_to_browser', 'clarify.request', {
      request_id: 'modify-input',
      question: '请说明希望如何修改当前任务方案。',
      choices: null,
    }),
    record(7, 'browser_to_gateway', 'clarify.respond', {
      answer: '不使用 eBPF，只查询 RED 指标',
    }),
    record(
      8,
      'gateway_to_browser',
      'tool.completed',
      planToolPayload('plan-tool-2', ['查询服务 RED 指标', '关联拓扑判断异常来源']),
    ),
    record(9, 'gateway_to_browser', 'clarify.request', {
      request_id: 'plan-2',
      question: '模型已更新规划，请确认是否继续。',
      choices: ['确认并继续', '修改方案'],
    }),
    record(10, 'browser_to_gateway', 'clarify.respond', { answer: '确认并继续' }),
  ])

  const plans = snapshot.timeline.filter((item) => item.kind === 'plan')
  assert.equal(plans.length, 2)
  assert.deepEqual(
    plans.map((item) => item.plan.steps.map((step) => step.title)),
    [
      ['查询服务 RED 指标', '查询 eBPF 指标'],
      ['查询服务 RED 指标', '关联拓扑判断异常来源'],
    ],
  )
  assert.equal(
    snapshot.timeline.some(
      (item) =>
        item.kind === 'activity' &&
        item.activity.kind === 'tool' &&
        item.activity.name === 'task_flow_orchestration_nodes_upsert',
    ),
    false,
  )
  const ordinaryClarification = snapshot.timeline.find(
    (item) => item.kind === 'checkpoint' && item.checkpoint.requestId === 'modify-input',
  )
  assert.equal(ordinaryClarification.checkpoint.purpose, undefined)
})

test('restores a model-generated plan from gateway message history', () => {
  const snapshot = historyFromGateway({
    messages: [
      { role: 'user', content: '分析订单服务' },
      { role: 'tool', ...planToolPayload('plan-tool-1', ['确认影响范围', '查询异常指标']) },
      { role: 'assistant', content: '分析完成' },
    ],
  })

  assert.deepEqual(
    snapshot.timeline.map((item) => item.kind),
    ['message', 'plan', 'message'],
  )
  const plan = snapshot.timeline.find((item) => item.kind === 'plan')
  assert.deepEqual(
    plan.plan.steps.map((step) => step.title),
    ['确认影响范围', '查询异常指标'],
  )
})
