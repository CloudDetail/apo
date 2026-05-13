import test from 'node:test'
import assert from 'node:assert/strict'

import { consumeFindingsStreamResponse } from './findingsStream'
import type { FindingsStreamEvent } from './findingsStream'

const responseFromChunks = (chunks: string[], init: ResponseInit = {}) => {
  const encoder = new TextEncoder()
  return new Response(
    new ReadableStream({
      start(controller) {
        chunks.forEach((chunk) => controller.enqueue(encoder.encode(chunk)))
        controller.close()
      },
    }),
    {
      status: 200,
      headers: {
        'content-type': 'application/x-ndjson; charset=utf-8',
      },
      ...init,
    },
  )
}

test('consumeFindingsStreamResponse parses multiple NDJSON events from one chunk', async () => {
  const events: FindingsStreamEvent[] = []
  const response = responseFromChunks([
    '{"type":"tree_shell","hydrationId":"h1","nodes":[]}\n{"type":"tree_done","hydrationId":"h1"}\n',
  ])

  await consumeFindingsStreamResponse(response, {
    onEvent: (event) => events.push(event),
  })

  assert.deepEqual(
    events.map((event) => event.type),
    ['tree_shell', 'tree_done'],
  )
})

test('consumeFindingsStreamResponse preserves partial NDJSON lines across chunks', async () => {
  const events: FindingsStreamEvent[] = []
  const response = responseFromChunks([
    '{"type":"node_failed","hydrationId":"h1",',
    '"nodeId":"cpu_check","message":"timeout"}\n',
  ])

  await consumeFindingsStreamResponse(response, {
    onEvent: (event) => events.push(event),
  })

  assert.equal(events.length, 1)
  const [event] = events
  assert.equal(event.type, 'node_failed')
  if (event.type === 'node_failed') {
    assert.equal(event.nodeId, 'cpu_check')
  }
})

test('consumeFindingsStreamResponse rejects non-2xx responses before parsing events', async () => {
  const response = responseFromChunks(['{"type":"tree_shell","hydrationId":"h1","nodes":[]}\n'], {
    status: 502,
  })

  await assert.rejects(
    () =>
      consumeFindingsStreamResponse(response, {
        onEvent: () => {
          throw new Error('should not parse events')
        },
      }),
    /Findings stream request failed/,
  )
})

test('consumeFindingsStreamResponse rejects invalid NDJSON', async () => {
  const response = responseFromChunks(['{"type":"tree_shell"\n'])

  await assert.rejects(
    () => consumeFindingsStreamResponse(response, {}),
    /Invalid findings stream event/,
  )
})
