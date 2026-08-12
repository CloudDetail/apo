// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { createSerializedSessionWriter, HermesSessionApi } from './sessionApi.ts'

test('normalizes server session checkpoints with string choices', async () => {
  const api = new HermesSessionApi(
    '/hermes-api',
    async () =>
      new Response(
        JSON.stringify({
          sessions: [
            {
              stored_session_id: 'stored-1',
              live_session_id: 'live-1',
              title: '待确认方案',
              created_at: '2026-08-06T00:00:00.000Z',
              updated_at: '2026-08-06T00:01:00.000Z',
              status: 'waiting_confirmation',
              active_run_id: 'run-1',
              checkpoint: {
                type: 'clarification',
                title: '需要确认后继续',
                description: '方案已生成。是否继续？',
                choices: ['确认并继续', '修改计划'],
                purpose: 'plan_confirmation',
              },
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
  )

  const sessions = await api.list()
  assert.equal(sessions[0].status, 'waiting_confirmation')
  assert.deepEqual(sessions[0].checkpoint.choices, [
    { value: '确认并继续', label: '确认并继续' },
    { value: '修改计划', label: '修改计划' },
  ])
  assert.equal(sessions[0].checkpoint.purpose, 'plan_confirmation')
})

test('serializes writes for the same newly created session', async () => {
  const calls = []
  let releaseFirst
  const firstPending = new Promise((resolve) => {
    releaseFirst = resolve
  })
  const writer = createSerializedSessionWriter({
    async upsert(session) {
      calls.push(`start:${session.status}`)
      if (session.status === 'ready') await firstPending
      calls.push(`end:${session.status}`)
      return session
    },
  })
  const base = {
    storedSessionId: 'stored-new',
    liveSessionId: 'live-new',
    title: '新会话',
    createdAt: '2026-08-11T00:00:00.000Z',
    updatedAt: '2026-08-11T00:00:00.000Z',
    activeRunId: null,
    checkpoint: null,
  }

  const first = writer({ ...base, status: 'ready' })
  const second = writer({
    ...base,
    status: 'running',
    activeRunId: 'run-1',
    updatedAt: '2026-08-11T00:00:01.000Z',
  })

  await new Promise((resolve) => setImmediate(resolve))
  assert.deepEqual(calls, ['start:ready'])
  releaseFirst()
  await Promise.all([first, second])
  assert.deepEqual(calls, ['start:ready', 'end:ready', 'start:running', 'end:running'])
})
