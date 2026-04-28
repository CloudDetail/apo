import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildThreadTransitionReset,
  hasThreadSelectionChanged,
  shouldApplyRouteThreadSelection,
  shouldApplyThreadLoadResult,
} from './threadTransition'

test('switching to a history thread cancels the active run and clears thread-scoped state', () => {
  const reset = buildThreadTransitionReset({
    currentRunId: 'run-123',
    loadRevision: 4,
    targetThreadId: 'thread-456',
  })

  assert.equal(reset.shouldCancelRun, true)
  assert.equal(reset.shouldLoadHistory, true)
  assert.equal(reset.targetThreadId, 'thread-456')
  assert.equal(reset.nextLoadRevision, 5)
  assert.deepEqual(reset.runtime, {
    isRunning: false,
    currentRunId: undefined,
    currentStep: undefined,
    currentMessage: undefined,
    currentToolCall: undefined,
    loading: true,
    error: null,
  })
  assert.deepEqual(reset.scoped, {
    messages: [],
    state: {},
    findingsToolCalls: new Map(),
    findingsAgentToolDisplay: new Map(),
  })
})

test('returning to a new conversation clears the active thread without scheduling a history load', () => {
  const reset = buildThreadTransitionReset({
    currentRunId: undefined,
    loadRevision: 0,
  })

  assert.equal(reset.shouldCancelRun, false)
  assert.equal(reset.shouldLoadHistory, false)
  assert.equal(reset.targetThreadId, undefined)
  assert.equal(reset.nextLoadRevision, 1)
  assert.equal(reset.runtime.loading, false)
})

test('returning to a new conversation during a run cancels the run and clears thread-scoped state', () => {
  const reset = buildThreadTransitionReset({
    currentRunId: 'run-123',
    loadRevision: 4,
  })

  assert.equal(reset.shouldCancelRun, true)
  assert.equal(reset.shouldLoadHistory, false)
  assert.equal(reset.targetThreadId, undefined)
  assert.equal(reset.nextLoadRevision, 5)
  assert.deepEqual(reset.runtime, {
    isRunning: false,
    currentRunId: undefined,
    currentStep: undefined,
    currentMessage: undefined,
    currentToolCall: undefined,
    loading: false,
    error: null,
  })
  assert.deepEqual(reset.scoped, {
    messages: [],
    state: {},
    findingsToolCalls: new Map(),
    findingsAgentToolDisplay: new Map(),
  })
})

test('only the newest history load result is allowed to update state', () => {
  assert.equal(shouldApplyThreadLoadResult({ activeLoadRevision: 3, responseLoadRevision: 3 }), true)
  assert.equal(
    shouldApplyThreadLoadResult({ activeLoadRevision: 4, responseLoadRevision: 3 }),
    false,
  )
})

test('does not request another thread transition when the selected threadId did not change', () => {
  assert.equal(hasThreadSelectionChanged(undefined, undefined), false)
  assert.equal(hasThreadSelectionChanged('thread-123', 'thread-123'), false)
  assert.equal(hasThreadSelectionChanged(undefined, 'thread-123'), true)
  assert.equal(hasThreadSelectionChanged('thread-123', undefined), true)
})

test('does not reload history when the route threadId only catches up with the already active thread', () => {
  assert.equal(
    shouldApplyRouteThreadSelection({
      previousRouteThreadId: undefined,
      nextRouteThreadId: 'thread-123',
      activeThreadId: 'thread-123',
    }),
    false,
  )
})

test('applies the initial route thread selection for refresh, manual URL entry, and share-style deep links', () => {
  assert.equal(
    shouldApplyRouteThreadSelection({
      previousRouteThreadId: undefined,
      nextRouteThreadId: 'thread-deep-link',
      activeThreadId: undefined,
    }),
    true,
  )
})

test('still applies route-driven thread transitions for true thread switches and reset actions', () => {
  assert.equal(
    shouldApplyRouteThreadSelection({
      previousRouteThreadId: 'thread-123',
      nextRouteThreadId: 'thread-456',
      activeThreadId: 'thread-123',
    }),
    true,
  )
  assert.equal(
    shouldApplyRouteThreadSelection({
      previousRouteThreadId: 'thread-123',
      nextRouteThreadId: undefined,
      activeThreadId: 'thread-123',
    }),
    true,
  )
})
