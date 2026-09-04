import test from 'node:test'
import assert from 'node:assert/strict'

import {
  isLifecycleRefreshCurrent,
  isLifecycleStateReached,
  waitForLifecycleStates,
} from './lifecycleState.ts'

const target = { nodeId: 'node-43', probeType: 'alloy' }

test('does not treat a successful command as started while authoritative state is still fused', async () => {
  const fusedState = {
    ...target,
    status: 'FUSED',
    circuit: { state: 'FUSED' },
    runtime: { childRunning: false, circuit: { state: 'FUSED' } },
  }

  const result = await waitForLifecycleStates(async () => [fusedState], [target], true, {
    maxAttempts: 1,
  })

  assert.deepEqual(result.pending, [target])
  assert.equal(isLifecycleStateReached(fusedState, true), false)
})

test('waits until the authoritative state reaches the requested lifecycle state', async () => {
  let calls = 0
  const result = await waitForLifecycleStates(
    async () => {
      calls += 1
      return calls === 1
        ? [
            {
              ...target,
              status: 'FUSED',
              circuit: { state: 'FUSED' },
              runtime: { childRunning: false, circuit: { state: 'FUSED' } },
            },
          ]
        : [
            {
              ...target,
              status: 'RUNNING',
              circuit: { state: 'NORMAL' },
              runtime: { childRunning: true, circuit: { state: 'NORMAL' } },
            },
          ]
    },
    [target],
    true,
    { maxAttempts: 2, sleep: async () => undefined },
  )

  assert.equal(calls, 2)
  assert.deepEqual(result.pending, [])
})

test('requires the stopped child and fused circuit before confirming stop', () => {
  assert.equal(
    isLifecycleStateReached(
      {
        ...target,
        status: 'FUSED',
        circuit: { state: 'FUSED' },
        runtime: { childRunning: true, circuit: { state: 'FUSED' } },
      },
      false,
    ),
    false,
  )
})

test('rejects a periodic refresh started before a lifecycle state transition', () => {
  assert.equal(isLifecycleRefreshCurrent(4, 5), false)
  assert.equal(isLifecycleRefreshCurrent(5, 5), true)
})
