// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { createSerializedSessionWriter, HermesSessionApi } from './sessionApi.ts'

const readUserId = () => 101

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
    readUserId,
  )

  const sessions = await api.list()
  assert.equal(sessions[0].status, 'waiting_confirmation')
  assert.deepEqual(sessions[0].checkpoint.choices, [
    { value: '确认并继续', label: '确认并继续' },
    { value: '修改计划', label: '修改计划' },
  ])
  assert.equal(sessions[0].checkpoint.purpose, 'plan_confirmation')
})

test('persists a user-requested interruption for the active run', async () => {
  const requests = []
  const api = new HermesSessionApi(
    '/hermes-api',
    async (input, init) => {
      requests.push({ input, init })
      return new Response(
        JSON.stringify({
          stored_session_id: 'stored/stop',
          live_session_id: 'live-stop',
          title: '停止测试',
          created_at: '2026-08-25T03:15:30.000Z',
          updated_at: '2026-08-25T03:16:00.000Z',
          status: 'interrupted',
          active_run_id: 'run-stop',
          checkpoint: null,
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      )
    },
    readUserId,
  )

  const interrupted = await api.interrupt('stored/stop', 'run-stop')

  assert.equal(interrupted.status, 'interrupted')
  assert.equal(interrupted.activeRunId, 'run-stop')
  assert.equal(requests[0].input, '/hermes-api/sessions/stored%2Fstop/interrupt')
  assert.equal(requests[0].init.method, 'POST')
  assert.deepEqual(JSON.parse(requests[0].init.body), {
    active_run_id: 'run-stop',
    cause: 'user_requested',
  })
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

test('promotes persisted capability result evidence into event payload', async () => {
  const evidence = [{ evidenceId: 'logs-1', kind: 'apo_logs', data: [] }]
  const api = new HermesSessionApi('/hermes-api', async () => new Response(JSON.stringify({
    events: [{
      event_id: 'event-1',
      sequence_no: 1,
      direction: 'gateway_to_browser',
      event_type: 'tool.complete',
      run_id: 'run-1',
      occurred_at: '2026-08-12T00:00:00.000Z',
      payload: { tool_id: 'tool-1', result: { schemaVersion: 'capability.result.v1', evidence } },
    }],
  }), { status: 200, headers: { 'Content-Type': 'application/json' } }), readUserId)

  const events = await api.history('stored-1')
  assert.equal(events[0].payload.evidence[0].id, 'logs-1')
  assert.deepEqual(events[0].payload.evidence[0].data, [])
})

test('adds the current APO user id to every session request', async () => {
  const calls = []
  const api = new HermesSessionApi(
    '/hermes-api',
    async (_input, init) => {
      calls.push('request')
      assert.equal(init.credentials, 'include')
      assert.equal(new Headers(init.headers).has('Authorization'), false)
      assert.equal(new Headers(init.headers).get('X-Hermes-User-Id'), '101')
      return new Response(JSON.stringify({ sessions: [] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    },
    readUserId,
  )

  await api.list()
  await api.list()
  assert.deepEqual(calls, ['request', 'request'])
})

test('does not send a session request before the APO user id is available', async () => {
  let requested = false
  const api = new HermesSessionApi(
    '/hermes-api',
    async () => {
      requested = true
      return new Response(null, { status: 204 })
    },
    () => undefined,
  )

  await assert.rejects(() => api.remove('stored-1'), /APO 用户信息尚未加载/)
  assert.equal(requested, false)
})
