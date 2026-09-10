// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { hermesMessageReducer, initialHermesMessageState } from './messageReducer.ts'

function event(type, payload = {}) {
  return { type: 'EVENT_RECEIVED', event: { type, sessionId: 'live-1', payload } }
}

function messages(state) {
  return state.timeline.filter((item) => item.kind === 'message').map((item) => item.message)
}

test('resolves only the matching automatic plan confirmation', () => {
  const waiting = hermesMessageReducer(initialHermesMessageState, event('clarify.request', {
    request_id: 'plan-auto', choices: ['确认并继续', '修改计划'],
  }))
  assert.equal(hermesMessageReducer(waiting, event('clarify.respond', {
    request_id: 'old-plan', source: 'plan_timeout', answer: '确认并继续',
  })), waiting)
  const state = hermesMessageReducer(waiting, event('clarify.respond', {
    request_id: 'plan-auto', source: 'plan_timeout', answer: '确认并继续',
  }))
  assert.equal(state.checkpoint, null)
  assert.equal(state.busy, true)
  assert.equal(state.timeline.find(item => item.kind === 'checkpoint').resolved, true)
})

function activities(state) {
  return state.timeline.filter((item) => item.kind === 'activity').map((item) => item.activity)
}

function plans(state) {
  return state.timeline.filter((item) => item.kind === 'plan').map((item) => item.plan)
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

test('streams an assistant message after the execution status and marks it complete', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '分析订单服务',
    runId: 'run-1',
  })
  state = hermesMessageReducer(state, event('message.delta', { text: '正在' }))
  state = hermesMessageReducer(state, event('message.delta', { text: '分析' }))
  state = hermesMessageReducer(state, event('message.complete', { text: '分析完成' }))

  const visibleMessages = messages(state)
  assert.equal(visibleMessages[0].content, '分析订单服务')
  assert.equal(visibleMessages[1].content, '分析完成')
  assert.equal(visibleMessages[1].complete, true)
  assert.deepEqual(
    state.timeline.map((item) => item.kind),
    ['message', 'run_status', 'message'],
  )
  assert.equal(state.timeline[1].status, 'completed')
  assert.equal(state.busy, false)
  assert.equal(state.statusText, '已完成')
})

test('keeps tool activity and clarification checkpoints in timeline order', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '检查错误',
    runId: 'run-2',
  })
  state = hermesMessageReducer(
    state,
    event('tool.start', { tool_id: 'tool-1', name: 'query_logs', args: { service: 'order' } }),
  )
  state = hermesMessageReducer(
    state,
    event('tool.completed', { tool_id: 'tool-1', name: 'query_logs', result: { total: 3 } }),
  )
  state = hermesMessageReducer(
    state,
    event('clarify.request', {
      request_id: 'clarify-1',
      question: '是否继续查询 Trace？',
      choices: ['确认并继续', '暂不查询'],
    }),
  )

  assert.equal(activities(state)[0].kind, 'tool')
  assert.equal(activities(state)[0].status, 'completed')
  assert.deepEqual(
    state.timeline.map((item) => item.kind),
    ['message', 'run_status', 'activity', 'checkpoint'],
  )
  assert.equal(state.checkpoint.requestId, 'clarify-1')
  assert.equal(state.checkpoint.choices[0].label, '确认并继续')
  assert.equal(state.busy, false)
})

test('attaches evidence to the producing tool activity', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED', text: '查询指标', runId: 'run-evidence',
  })
  state = hermesMessageReducer(state, event('tool.start', { tool_id: 'tool-evidence', name: 'query_metrics' }))
  state = hermesMessageReducer(state, event('tool.complete', {
    tool_id: 'tool-evidence',
    name: 'query_metrics',
    evidence: [{ id: 'metric-1', kind: 'apo_metrics', title: 'CPU', source: { sourceId: 'system-apo', operationId: 'queryMetrics' }, data: [] }],
  }))

  assert.equal(activities(state)[0].evidence[0].id, 'metric-1')
  assert.equal(state.timeline.find((item) => item.kind === 'activity').runId, 'run-evidence')
})

test('renders normalized capability result evidence on the tool activity', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED', text: '查询指标', runId: 'run-capability-evidence',
  })
  state = hermesMessageReducer(state, event('tool.complete', {
    tool_id: 'tool-capability',
    name: 'capability_invoke',
    result: { schemaVersion: 'capability.result.v1', evidence: [{ evidenceId: 'metric-2', kind: 'apo_metrics', data: [] }] },
    evidence: [{ evidenceId: 'metric-2', kind: 'apo_metrics', data: [] }],
  }))
  assert.equal(activities(state)[0].evidence[0].evidenceId, 'metric-2')
})

test('keeps every model-generated plan in realtime without version labels', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '分析服务延时',
    runId: 'run-plan',
  })
  state = hermesMessageReducer(
    state,
    event('tool.start', planToolPayload('plan-tool-1', ['查询 RED 指标', '查询 eBPF 指标'])),
  )
  state = hermesMessageReducer(
    state,
    event('tool.completed', planToolPayload('plan-tool-1', ['查询 RED 指标', '查询 eBPF 指标'])),
  )
  state = hermesMessageReducer(
    state,
    event('clarify.request', {
      request_id: 'plan-1',
      question: '模型方案已生成，请确认是否继续。',
      choices: ['确认并继续', '修改方案'],
    }),
  )
  state = hermesMessageReducer(state, { type: 'CHECKPOINT_RESPONDED', text: '修改方案' })
  state = hermesMessageReducer(
    state,
    event('tool.completed', planToolPayload('plan-tool-2', ['查询 RED 指标', '关联拓扑'])),
  )
  state = hermesMessageReducer(
    state,
    event('clarify.request', {
      request_id: 'plan-2',
      question: '模型方案已更新，请确认是否继续。',
      choices: ['确认并继续', '修改方案'],
    }),
  )

  assert.deepEqual(
    plans(state).map((plan) => plan.steps.map((step) => step.title)),
    [
      ['查询 RED 指标', '查询 eBPF 指标'],
      ['查询 RED 指标', '关联拓扑'],
    ],
  )
  assert.equal(
    activities(state).some((activity) => activity.name === 'task_flow_orchestration_nodes_upsert'),
    false,
  )
  assert.equal(state.checkpoint.requestId, 'plan-2')
})

test('appends a steering message without creating an empty assistant placeholder', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '分析全部服务',
    runId: 'run-3',
  })
  state = hermesMessageReducer(state, {
    type: 'USER_ADDITION',
    text: '只看订单服务',
    runId: 'run-3',
  })

  assert.deepEqual(
    messages(state).map((message) => [message.role, message.content]),
    [
      ['user', '分析全部服务'],
      ['user', '只看订单服务'],
    ],
  )
})

test('treats interrupted message completion as an interrupted run', () => {
  const started = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '停止测试',
    runId: 'run-interrupt',
  })
  const interrupted = hermesMessageReducer(started, {
    type: 'EVENT_RECEIVED',
    event: {
      type: 'message.complete',
      sessionId: 'session-1',
      payload: { text: null, status: 'interrupted' },
    },
  })

  assert.equal(interrupted.busy, false)
  assert.equal(interrupted.statusText, '处理已中断')
  assert.deepEqual(
    messages(interrupted).map((item) => item.role),
    ['user'],
  )
  assert.equal(interrupted.timeline[1].status, 'interrupted')
})

test('shows an immediate stopping state and enables retry after timeout', () => {
  const started = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '停止测试',
    runId: 'run-stop',
  })
  const stopping = hermesMessageReducer(started, { type: 'STOP_REQUESTED' })
  const acknowledged = hermesMessageReducer(stopping, { type: 'STOP_ACKNOWLEDGED' })
  const retryable = hermesMessageReducer(acknowledged, { type: 'STOP_RETRY_AVAILABLE' })

  assert.equal(stopping.stopping, true)
  assert.equal(stopping.statusText, '正在停止当前任务')
  assert.equal(stopping.timeline[1].status, 'stopping')
  assert.equal(retryable.stopRetryAllowed, true)
  assert.equal(retryable.statusText, '停止请求尚未确认，可重试')
})

test('finishes the stopping state when the interrupt acknowledgement is persisted', () => {
  const started = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '停止测试',
    runId: 'run-stop-ack',
  })
  const stopping = hermesMessageReducer(started, { type: 'STOP_REQUESTED' })
  const acknowledged = hermesMessageReducer(stopping, { type: 'STOP_ACKNOWLEDGED' })
  const interrupted = hermesMessageReducer(
    acknowledged,
    event('session.interrupted', { run_id: 'run-stop-ack', message: '用户已停止当前任务。' }),
  )

  assert.equal(interrupted.busy, false)
  assert.equal(interrupted.stopping, false)
  assert.equal(interrupted.stopRetryAllowed, false)
  assert.equal(interrupted.activeRunId, null)
  assert.equal(interrupted.statusText, '处理已中断')
  assert.equal(interrupted.timeline[1].status, 'interrupted')
  assert.equal(interrupted.timeline[1].label, '任务已停止')
})

for (const eventType of ['run.cancelled', 'run.canceled']) {
  test(`treats ${eventType} as an interrupted terminal event`, () => {
    const started = hermesMessageReducer(initialHermesMessageState, {
      type: 'SUBMIT_STARTED',
      text: '停止测试',
      runId: `run-${eventType}`,
    })
    const interrupted = hermesMessageReducer(
      started,
      event(eventType, { run_id: `run-${eventType}` }),
    )

    assert.equal(interrupted.busy, false)
    assert.equal(interrupted.activeRunId, null)
    assert.equal(interrupted.statusText, '处理已中断')
    assert.equal(interrupted.timeline[1].status, 'interrupted')
  })
}

test('ignores a delayed terminal event for an interrupted run after a new run starts', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '旧任务',
    runId: 'run-interrupted-old',
  })
  state = hermesMessageReducer(
    state,
    event('session.interrupted', { run_id: 'run-interrupted-old' }),
  )
  state = hermesMessageReducer(state, {
    type: 'SUBMIT_STARTED',
    text: '新任务',
    runId: 'run-active-new',
  })

  const lateEvents = [
    event('error', { run_id: 'run-interrupted-old', message: '旧任务迟到错误' }),
    event('session.interrupted', { run_id: 'run-interrupted-old' }),
    event('run.cancelled', { run_id: 'run-interrupted-old' }),
    event('message.complete', { run_id: 'run-interrupted-old', status: 'interrupted' }),
  ]
  for (const lateEvent of lateEvents) {
    const afterLateTerminal = hermesMessageReducer(state, lateEvent)
    assert.equal(afterLateTerminal, state)
    assert.equal(afterLateTerminal.busy, true)
    assert.equal(afterLateTerminal.activeRunId, 'run-active-new')
    assert.equal(afterLateTerminal.error, null)
  }
  assert.equal(
    state.timeline.find(
      (item) => item.kind === 'run_status' && item.runId === 'run-interrupted-old',
    ).status,
    'interrupted',
  )
  assert.equal(
    state.timeline.find(
      (item) => item.kind === 'run_status' && item.runId === 'run-active-new',
    ).status,
    'running',
  )
})

test('ignores unscoped activity that arrives after an interrupted session is reopened', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '旧任务',
    runId: 'run-reopened-stop',
  })
  state = hermesMessageReducer(
    state,
    event('session.interrupted', { run_id: 'run-reopened-stop' }),
  )

  for (const delayedEvent of [
    event('message.delta', { text: '旧任务迟到片段' }),
    event('tool.completed', { tool_id: 'old-tool', result: { stale: true } }),
  ]) {
    assert.equal(hermesMessageReducer(state, delayedEvent), state)
  }
})

test('ignores a delayed stop failure after the expected run already reached a terminal state', () => {
  const started = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '自然完成任务',
    runId: 'run-completed-before-stop-failure',
  })
  const completed = hermesMessageReducer(
    started,
    event('run.completed', { run_id: 'run-completed-before-stop-failure' }),
  )
  const afterLateFailure = hermesMessageReducer(completed, {
    type: 'STOP_FAILED',
    error: '迟到的停止失败',
    runId: 'run-completed-before-stop-failure',
  })

  assert.equal(afterLateFailure, completed)
  assert.equal(afterLateFailure.statusText, '已完成')
  assert.equal(afterLateFailure.error, null)
})

test('does not let an old stop failure overwrite a newer active run', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '旧任务',
    runId: 'run-old-stop-failure',
  })
  state = hermesMessageReducer(
    state,
    event('session.interrupted', { run_id: 'run-old-stop-failure' }),
  )
  state = hermesMessageReducer(state, {
    type: 'SUBMIT_STARTED',
    text: '新任务',
    runId: 'run-new-after-stop',
  })

  const afterLateFailure = hermesMessageReducer(state, {
    type: 'STOP_FAILED',
    error: '旧任务停止失败',
    runId: 'run-old-stop-failure',
  })

  assert.equal(afterLateFailure, state)
  assert.equal(afterLateFailure.busy, true)
  assert.equal(afterLateFailure.activeRunId, 'run-new-after-stop')
  assert.equal(afterLateFailure.error, null)
})

test('marks the active run failed when message delivery fails', () => {
  const started = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '分析服务异常',
    runId: 'run-failed',
  })
  const failed = hermesMessageReducer(started, {
    type: 'SEND_FAILED',
    error: '会话连接超时',
  })

  assert.equal(failed.busy, false)
  assert.equal(failed.activeRunId, null)
  assert.equal(failed.statusText, '发送失败')
  const runStatus = failed.timeline.find((item) => item.kind === 'run_status')
  assert.equal(runStatus.status, 'failed')
  assert.equal(runStatus.label, '消息发送失败')
})
