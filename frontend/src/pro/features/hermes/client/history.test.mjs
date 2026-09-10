// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { historyFromGateway, historyFromPersistedEvents, visibleMessageText } from './history.ts'

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

test('restores automatic confirmation without an unanswered checkpoint', () => {
  const snapshot = historyFromPersistedEvents([
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '检查服务' }),
    record(2, 'gateway_to_browser', 'clarify.request', {
      request_id: 'auto-plan', choices: ['确认并继续', '修改计划'],
    }),
    record(3, 'gateway_to_browser', 'clarify.respond', {
      request_id: 'auto-plan', answer: '确认并继续', source: 'plan_timeout',
    }),
  ])
  assert.equal(snapshot.checkpoint, null)
  assert.equal(snapshot.timeline.find(item => item.kind === 'checkpoint').resolved, true)
})

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

test('restores a persisted user interruption as a terminal run', () => {
  const snapshot = historyFromPersistedEvents([
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '分析主机资源' }, 'run-stop'),
    record(
      2,
      'gateway_to_browser',
      'session.interrupted',
      { message: '用户已停止当前任务。' },
      'run-stop',
    ),
  ])

  const runStatus = snapshot.timeline.find((item) => item.kind === 'run_status')
  assert.equal(runStatus.status, 'interrupted')
  assert.equal(runStatus.label, '任务已停止')
  assert.equal(snapshot.activeRunId, null)
  assert.equal(snapshot.checkpoint, null)
})

for (const eventType of ['run.cancelled', 'run.canceled']) {
  test(`restores ${eventType} as an interrupted terminal run`, () => {
    const runId = `run-${eventType}`
    const snapshot = historyFromPersistedEvents([
      record(1, 'browser_to_gateway', 'prompt.submit', { text: '停止任务' }, runId),
      record(2, 'gateway_to_browser', eventType, {}, runId),
    ])

    const runStatus = snapshot.timeline.find(
      (item) => item.kind === 'run_status' && item.runId === runId,
    )
    assert.equal(runStatus.status, 'interrupted')
    assert.equal(runStatus.label, '任务已停止')
    assert.equal(snapshot.activeRunId, null)
  })
}

test('keeps a newer run active when an interrupted run receives a delayed terminal event', () => {
  const snapshot = historyFromPersistedEvents([
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '旧任务' }, 'run-old'),
    record(2, 'gateway_to_browser', 'session.interrupted', {}, 'run-old'),
    record(3, 'browser_to_gateway', 'prompt.submit', { text: '新任务' }, 'run-new'),
    record(4, 'gateway_to_browser', 'error', { message: '旧任务迟到错误' }, 'run-old'),
    record(5, 'gateway_to_browser', 'message.complete', { text: '旧任务迟到回答' }, 'run-old'),
    record(6, 'gateway_to_browser', 'session.interrupted', {}, 'run-old'),
    record(7, 'gateway_to_browser', 'run.cancelled', {}, 'run-old'),
    record(
      8,
      'gateway_to_browser',
      'message.complete',
      { text: '旧任务迟到的中断回答', status: 'interrupted' },
      'run-old',
    ),
  ])

  const runStatuses = snapshot.timeline.filter((item) => item.kind === 'run_status')
  assert.equal(runStatuses.find((item) => item.runId === 'run-old').status, 'interrupted')
  assert.equal(runStatuses.find((item) => item.runId === 'run-new').status, 'running')
  assert.equal(snapshot.activeRunId, 'run-new')
  assert.equal(
    snapshot.timeline.some(
      (item) =>
        item.kind === 'message' &&
        ['旧任务迟到回答', '旧任务迟到的中断回答'].includes(item.message.content),
    ),
    false,
  )
})

test('ignores unscoped activity persisted after an interrupted run', () => {
  const snapshot = historyFromPersistedEvents([
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '旧任务' }, 'run-old'),
    record(2, 'gateway_to_browser', 'session.interrupted', {}, 'run-old'),
    record(
      3,
      'gateway_to_browser',
      'tool.completed',
      { tool_id: 'old-tool', name: 'query_old_data', result: { stale: true } },
      'run-old',
    ),
  ])

  assert.equal(snapshot.timeline.some((item) => item.kind === 'activity'), false)
  assert.equal(snapshot.timeline.find((item) => item.kind === 'run_status').status, 'interrupted')
  assert.equal(snapshot.activeRunId, null)
})

test('ignores an unscoped old completion persisted against a queued run before message.start', () => {
  const snapshot = historyFromPersistedEvents([
    record(1, 'browser_to_gateway', 'prompt.submit', { text: '旧任务' }, 'run-old'),
    record(
      2,
      'gateway_to_browser',
      'session.interrupted',
      { run_id: 'run-old' },
      'run-old',
    ),
    record(3, 'browser_to_gateway', 'prompt.submit', { text: '新任务' }, 'run-new'),
    record(
      4,
      'gateway_to_browser',
      'message.complete',
      { text: '旧任务迟到回答' },
      'run-new',
    ),
    record(5, 'gateway_to_browser', 'message.start', {}, 'run-new'),
    record(
      6,
      'gateway_to_browser',
      'message.complete',
      { text: '新任务回答' },
      'run-new',
    ),
  ])

  const runStatuses = snapshot.timeline.filter((item) => item.kind === 'run_status')
  assert.equal(runStatuses.find((item) => item.runId === 'run-old').status, 'interrupted')
  assert.equal(runStatuses.find((item) => item.runId === 'run-new').status, 'completed')
  assert.equal(snapshot.activeRunId, null)
  assert.equal(
    snapshot.timeline.some(
      (item) => item.kind === 'message' && item.message.content === '旧任务迟到回答',
    ),
    false,
  )
  assert.equal(
    snapshot.timeline.some(
      (item) => item.kind === 'message' && item.message.content === '新任务回答',
    ),
    true,
  )
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

test('hides current Workflow and legacy Playbook selection markers from history', () => {
  assert.equal(
    visibleMessageText('检查订单服务\n\n<hermes_qa_workflow_selection>{"workflowId":"workflow-order"}</hermes_qa_workflow_selection>'),
    '检查订单服务',
  )
  assert.equal(
    visibleMessageText('检查订单服务\n\n<hermes_qa_playbook_selection>{"playbookId":"legacy"}</hermes_qa_playbook_selection>'),
    '检查订单服务',
  )
})
