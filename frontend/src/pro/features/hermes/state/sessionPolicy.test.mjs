// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { isRunScopedEventType, runIdFromEventPayload } from '../client/eventPayload.ts'
import { hermesMessageReducer, initialHermesMessageState } from './messageReducer.ts'
import {
  canBindNewSessionOperation,
  canApplyInterruptReconciliation,
  currentSessionCreation,
  hydrationRuntimeFromHistory,
  isCurrentSessionOperation,
  isKnownInterruptedRun,
  reconcileInterruptedBookmark,
  runStartBarrierDecision,
  shouldIgnoreKnownInterruptedRunEvent,
  shouldRestoreInBackground,
} from './sessionPolicy.ts'

test('only live workflow states restore in the background', () => {
  assert.equal(shouldRestoreInBackground('running'), true)
  assert.equal(shouldRestoreInBackground('waiting_confirmation'), true)
  assert.equal(shouldRestoreInBackground('ready'), false)
  assert.equal(shouldRestoreInBackground('failed'), false)
  assert.equal(shouldRestoreInBackground('interrupted'), false)
})

test('invalidates delayed session work after the selected conversation changes', () => {
  const operation = { generation: 3, storedSessionId: 'stored-a' }

  assert.equal(isCurrentSessionOperation(operation, 3, 'stored-a'), true)
  assert.equal(isCurrentSessionOperation(operation, 3, 'stored-b'), false)
  assert.equal(isCurrentSessionOperation(operation, 4, 'stored-a'), false)
})

test('allows a new conversation operation to bind its session within the same generation', () => {
  const operation = { generation: 7, storedSessionId: null }

  assert.equal(isCurrentSessionOperation(operation, 7, null), true)
  assert.equal(isCurrentSessionOperation(operation, 7, 'stored-created'), true)
  assert.equal(isCurrentSessionOperation(operation, 8, 'stored-created'), false)
})

test('binds a newly created session only while its empty conversation generation is current', () => {
  assert.equal(canBindNewSessionOperation(7, 7, null), true)
  assert.equal(canBindNewSessionOperation(7, 8, null), false)
  assert.equal(canBindNewSessionOperation(7, 7, 'stored-selected'), false)
})

test('awaits the current deferred session creation before using an active-session fast path', async () => {
  let releaseCreation
  const promise = new Promise((resolve) => {
    releaseCreation = resolve
  })
  const pending = { generation: 7, promise, scope: 'user-a' }

  const shared = currentSessionCreation(pending, 7, (scope) => scope === 'user-a')
  assert.equal(shared, promise)

  let settled = false
  void shared.then(() => {
    settled = true
  })
  await Promise.resolve()
  assert.equal(settled, false)

  releaseCreation({ storedSessionId: 'stored-created' })
  assert.deepEqual(await shared, { storedSessionId: 'stored-created' })
  assert.equal(currentSessionCreation(pending, 8, () => true), null)
  assert.equal(currentSessionCreation(pending, 7, () => false), null)
})

test('lets the latest persisted terminal run override a stale running bookmark', () => {
  const bookmark = {
    storedSessionId: 'stored-stop',
    liveSessionId: 'live-stop',
    title: '停止测试',
    createdAt: '2026-08-25T03:15:30.000Z',
    updatedAt: '2026-08-25T03:16:00.000Z',
    status: 'running',
    activeRunId: 'run-stop',
    checkpoint: { type: 'approval', requestId: 'stale-checkpoint' },
  }
  const snapshot = {
    timeline: [
      {
        id: 'run-status-run-stop',
        kind: 'run_status',
        sequenceNo: 2,
        occurredAt: '2026-08-25T03:16:01.000Z',
        runId: 'run-stop',
        status: 'interrupted',
        label: '任务已停止',
      },
    ],
    checkpoint: null,
    activeRunId: null,
  }

  assert.deepEqual(hydrationRuntimeFromHistory(bookmark, snapshot), {
    status: 'interrupted',
    activeRunId: null,
    checkpoint: null,
    restoreInBackground: false,
  })
})

test('lets a newer persisted active run override the interrupted bookmark for the old run', () => {
  const bookmark = {
    storedSessionId: 'stored-stop',
    liveSessionId: 'live-stop',
    title: '停止后继续',
    createdAt: '2026-08-25T03:15:30.000Z',
    updatedAt: '2026-08-25T03:16:00.000Z',
    status: 'interrupted',
    activeRunId: 'run-old',
    checkpoint: null,
  }
  const snapshot = {
    timeline: [
      {
        id: 'run-status-run-old',
        kind: 'run_status',
        sequenceNo: 2,
        occurredAt: '2026-08-25T03:16:01.000Z',
        runId: 'run-old',
        status: 'interrupted',
        label: '任务已停止',
      },
      {
        id: 'run-status-run-new',
        kind: 'run_status',
        sequenceNo: 4,
        occurredAt: '2026-08-25T03:16:02.000Z',
        runId: 'run-new',
        status: 'running',
        label: '任务已开始',
      },
    ],
    checkpoint: null,
    activeRunId: 'run-new',
  }

  assert.deepEqual(hydrationRuntimeFromHistory(bookmark, snapshot), {
    status: 'running',
    activeRunId: 'run-new',
    checkpoint: null,
    restoreInBackground: true,
  })
})

test('retains bookmark runtime fallback when persisted history has no terminal run', () => {
  const checkpoint = { type: 'approval', requestId: 'checkpoint-1' }
  const bookmark = {
    status: 'running',
    activeRunId: 'run-bookmark',
    checkpoint,
  }

  assert.deepEqual(
    hydrationRuntimeFromHistory(bookmark, {
      timeline: [],
      checkpoint: null,
      activeRunId: null,
    }),
    {
      status: 'running',
      activeRunId: 'run-bookmark',
      checkpoint,
      restoreInBackground: true,
    },
  )
})

test('recognizes delayed events from an interrupted run during and after a new run starts', () => {
  const interruptedRunIds = new Set(['run-stopped'])

  assert.equal(isKnownInterruptedRun(interruptedRunIds, null, 'run-stopped'), true)
  assert.equal(isKnownInterruptedRun(interruptedRunIds, 'run-stopped', 'run-new'), true)
  assert.equal(isKnownInterruptedRun(interruptedRunIds, null, 'run-new'), false)
  assert.equal(isKnownInterruptedRun(interruptedRunIds, 'run-new', 'run-new'), false)
})

test('ignores every explicit old-run event but allows the current interruption terminal', () => {
  const interruptedRunIds = new Set(['run-stopped'])

  assert.equal(
    shouldIgnoreKnownInterruptedRunEvent(interruptedRunIds, 'run-stopped', 'run-stopped', false),
    true,
  )
  assert.equal(
    shouldIgnoreKnownInterruptedRunEvent(interruptedRunIds, 'run-stopped', 'run-stopped', true),
    false,
  )
  assert.equal(
    shouldIgnoreKnownInterruptedRunEvent(interruptedRunIds, 'run-stopped', 'run-new', true),
    true,
  )
  assert.equal(
    shouldIgnoreKnownInterruptedRunEvent(interruptedRunIds, null, 'run-new', true),
    false,
  )
})

test('keeps an acknowledged stop terminal when REST reconciliation is stale or unavailable', async () => {
  const interrupted = { status: 'interrupted', storedSessionId: 'stored-stop' }

  assert.equal(await reconcileInterruptedBookmark(async () => interrupted), interrupted)
  assert.equal(await reconcileInterruptedBookmark(async () => ({ status: 'ready' })), null)
  assert.equal(
    await reconcileInterruptedBookmark(async () => {
      throw new Error('REST unavailable')
    }),
    null,
  )
})

test('does not let delayed interrupt reconciliation overwrite a newer run in the same session', () => {
  assert.equal(
    canApplyInterruptReconciliation(
      'run-stopped',
      true,
      'run-stopped',
      'interrupted',
      'run-stopped',
    ),
    true,
  )
  assert.equal(
    canApplyInterruptReconciliation('run-stopped', true, 'run-new', 'running', 'run-new'),
    false,
  )
  assert.equal(
    canApplyInterruptReconciliation('run-stopped', false, null, 'running', 'run-new'),
    false,
  )
})

test('blocks unscoped old-turn events until the queued run emits message.start', () => {
  const barrier = { storedSessionId: 'stored-stop', runId: 'run-new' }

  assert.equal(
    runStartBarrierDecision(barrier, 'stored-stop', 'message.complete', null),
    'ignore',
  )
  assert.equal(
    runStartBarrierDecision(barrier, 'stored-stop', 'tool.completed', null),
    'ignore',
  )
  assert.equal(
    runStartBarrierDecision(barrier, 'stored-stop', 'session.interrupted', null),
    'ignore',
  )
  assert.equal(runStartBarrierDecision(barrier, 'stored-stop', 'message.start', null), 'clear')
  assert.equal(
    runStartBarrierDecision(barrier, 'stored-stop', 'message.start', 'run-new'),
    'clear',
  )
})

test('reads every supported run id alias from raw event payloads', () => {
  assert.equal(runIdFromEventPayload({ run_id: 'run-snake' }), 'run-snake')
  assert.equal(runIdFromEventPayload({ runId: 'run-camel' }), 'run-camel')
  assert.equal(runIdFromEventPayload({ client_run_id: 'client-snake' }), 'client-snake')
  assert.equal(runIdFromEventPayload({ clientRunId: 'client-camel' }), 'client-camel')
  assert.equal(runIdFromEventPayload({}), null)
})

test('classifies delayed message and tool activity as run-scoped events', () => {
  assert.equal(isRunScopedEventType('message.delta'), true)
  assert.equal(isRunScopedEventType('tool.completed'), true)
  assert.equal(isRunScopedEventType('session.title'), false)
})

test('leaves explicit run events and unrelated sessions to their run-aware guards', () => {
  const barrier = { storedSessionId: 'stored-stop', runId: 'run-new' }

  assert.equal(
    runStartBarrierDecision(barrier, 'stored-stop', 'message.complete', 'run-stopped'),
    'pass',
  )
  assert.equal(
    runStartBarrierDecision(barrier, 'stored-stop', 'message.start', 'run-stopped'),
    'pass',
  )
  assert.equal(
    runStartBarrierDecision(barrier, 'stored-other', 'message.complete', null),
    'pass',
  )
})

test('keeps the queued run active when the old turn completes without a run id', () => {
  let state = hermesMessageReducer(initialHermesMessageState, {
    type: 'SUBMIT_STARTED',
    text: '旧任务',
    runId: 'run-stopped',
  })
  state = hermesMessageReducer(state, {
    type: 'EVENT_RECEIVED',
    event: {
      type: 'session.interrupted',
      sessionId: 'live-stop',
      payload: { run_id: 'run-stopped' },
    },
  })
  state = hermesMessageReducer(state, {
    type: 'SUBMIT_STARTED',
    text: '新任务',
    runId: 'run-new',
  })
  let barrier = { storedSessionId: 'stored-stop', runId: 'run-new' }

  const delayedCompletion = {
    type: 'message.complete',
    sessionId: 'live-stop',
    payload: { text: '旧任务迟到回答' },
  }
  if (
    runStartBarrierDecision(
      barrier,
      'stored-stop',
      delayedCompletion.type,
      null,
    ) !== 'ignore'
  ) {
    state = hermesMessageReducer(state, { type: 'EVENT_RECEIVED', event: delayedCompletion })
  }

  assert.equal(state.busy, true)
  assert.equal(state.activeRunId, 'run-new')
  assert.equal(
    state.timeline.some(
      (item) => item.kind === 'message' && item.message.content === '旧任务迟到回答',
    ),
    false,
  )

  const startDecision = runStartBarrierDecision(barrier, 'stored-stop', 'message.start', null)
  if (startDecision === 'clear') barrier = null
  assert.equal(barrier, null)
})
