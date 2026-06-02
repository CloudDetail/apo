import test from 'node:test'
import assert from 'node:assert/strict'

import {
  applyFindingsStreamEvent,
  abortStaleFindingsStreams,
  buildFindingsCacheKey,
  buildFindingsNodeRetryParams,
  createPendingFindingsShell,
  getFindingsStreamOverflowKey,
  getStaleFindingsStreamKeys,
  isTerminalFindingsPhase,
  isWholeTreeDoneEvent,
  runFindingsStreamRequest,
  shouldStartFindingsStream,
  shouldSyncSelectedFindings,
} from './findingsHydration'
import type { FindingsNodePayload } from '../../../../api/conversation'

const node = (
  nodeId: string,
  parentId: string | null = null,
  title = nodeId,
): FindingsNodePayload => ({
  nodeId,
  title,
  nodeKind: parentId === null ? 'summary' : 'evidence',
  status: 'normal',
  statusColor: 'normal',
  parentId,
  summaryConclusion: `${title} summary`,
})

test('builds findings cache key from thread and service', () => {
  assert.equal(buildFindingsCacheKey('thread-1', 'svc-a'), 'thread-1:svc-a')
})

test('done and failed findings cache entries are terminal unless refreshed', () => {
  assert.equal(isTerminalFindingsPhase('done'), true)
  assert.equal(isTerminalFindingsPhase('failed'), true)
  assert.equal(isTerminalFindingsPhase('streaming'), false)

  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'done',
      isLoading: false,
      refresh: false,
    }),
    false,
  )
  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'failed',
      isLoading: false,
      refresh: false,
    }),
    false,
  )
  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'done',
      isLoading: false,
      refresh: true,
    }),
    true,
  )
})

test('terminal findings cache with an old contract version is refreshed', () => {
  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'done',
      isLoading: false,
      isCacheCurrent: false,
    }),
    true,
  )
})

test('node retry tree_done does not complete the whole tree phase', () => {
  assert.equal(
    isWholeTreeDoneEvent({
      type: 'tree_done',
      hydrationId: 'hydration-1',
      scope: 'node_retry',
      nodeId: 'code_cpu_check',
    }),
    false,
  )
  assert.equal(
    isWholeTreeDoneEvent({
      type: 'tree_done',
      hydrationId: 'hydration-1',
    }),
    true,
  )
})

test('in-flight findings streams are not duplicated', () => {
  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'streaming',
      isLoading: true,
      refresh: false,
    }),
    false,
  )
})

test('stream overflow aborts the oldest non-active key only when at capacity', () => {
  assert.equal(getFindingsStreamOverflowKey(['thread:svc-a'], 'thread:svc-b', 2), null)
  assert.equal(
    getFindingsStreamOverflowKey(['thread:svc-a', 'thread:svc-b'], 'thread:svc-b', 2),
    'thread:svc-a',
  )
  assert.equal(
    getFindingsStreamOverflowKey(['thread:svc-a', 'thread:svc-b'], 'thread:svc-a', 2),
    'thread:svc-b',
  )
  assert.equal(getFindingsStreamOverflowKey(['thread:svc-a'], 'thread:svc-a', 1), null)
})

test('service switch treats all other in-flight streams as stale', () => {
  assert.deepEqual(
    getStaleFindingsStreamKeys(['thread:svc-a', 'thread:svc-b', 'thread:svc-c'], 'thread:svc-b'),
    ['thread:svc-a', 'thread:svc-c'],
  )
})

test('service switch aborts stale in-flight stream keys through lifecycle adapter', () => {
  const aborted: string[] = []

  const staleKeys = abortStaleFindingsStreams({
    loadingKeys: ['thread:svc-a', 'thread:svc-b', 'thread:svc-c'],
    activeKey: 'thread:svc-b',
    abort: (cacheKey) => aborted.push(cacheKey),
  })

  assert.deepEqual(staleKeys, ['thread:svc-a', 'thread:svc-c'])
  assert.deepEqual(aborted, ['thread:svc-a', 'thread:svc-c'])
})

test('selected findings sync is ignored after active service changes', () => {
  assert.equal(
    shouldSyncSelectedFindings({
      cacheKey: 'thread-1:svc-a',
      activeKey: 'thread-1:svc-b',
      currentThreadId: 'thread-1',
      selectedServiceName: 'svc-b',
    }),
    false,
  )
})

test('stream lifecycle adapter ignores late events after active service changes', async () => {
  const appliedEvents: string[] = []
  let activeKey: string | null = 'thread:svc-a'

  await runFindingsStreamRequest({
    streamFindingsData: async (_params, handlers) => {
      handlers.onEvent?.({
        type: 'node_started',
        hydrationId: 'h1',
        nodeId: 'root',
      })
      activeKey = 'thread:svc-b'
      handlers.onEvent?.({
        type: 'node_patch',
        hydrationId: 'h1',
        nodeId: 'root',
        node: node('root'),
      })
    },
    params: {
      threadId: 'thread',
      serviceName: 'svc-a',
    },
    cacheKey: 'thread:svc-a',
    getActiveKey: () => activeKey,
    onActiveEvent: (event) => appliedEvents.push(event.type),
  })

  assert.deepEqual(appliedEvents, ['node_started'])
})

test('node retry params always target one node with refresh enabled', () => {
  const signal = new AbortController().signal

  assert.deepEqual(buildFindingsNodeRetryParams('thread', 'svc-a', 'code_lock_check', signal), {
    threadId: 'thread',
    serviceName: 'svc-a',
    nodeId: 'code_lock_check',
    refresh: true,
    signal,
  })
})

test('selected findings sync requires selected service to match cache key', () => {
  assert.equal(
    shouldSyncSelectedFindings({
      cacheKey: 'thread-1:svc-a',
      activeKey: 'thread-1:svc-a',
      currentThreadId: 'thread-1',
      selectedServiceName: 'svc-b',
    }),
    false,
  )
  assert.equal(
    shouldSyncSelectedFindings({
      cacheKey: 'thread-1:svc-a',
      activeKey: 'thread-1:svc-a',
      currentThreadId: 'thread-1',
      selectedServiceName: 'svc-a',
    }),
    true,
  )
})

test('tree_shell replaces nodes and marks shell nodes pending', () => {
  const result = applyFindingsStreamEvent([], {
    type: 'tree_shell',
    hydrationId: 'hydration-1',
    threadId: 'thread-1',
    serviceName: 'svc-a',
    nodes: [node('root'), node('cpu_check', 'root')],
  })

  assert.deepEqual(
    result.map((item) => [item.nodeId, item.hydrationStatus]),
    [
      ['root', 'pending'],
      ['cpu_check', 'pending'],
    ],
  )
})

test('creates a pending shell from a previous backend shell for service switches', () => {
  const result = createPendingFindingsShell([
    {
      ...node('root'),
      hydrationStatus: 'success',
      content: 'old content',
      metrics: { normalMetrics: [], abnormalMetrics: [] },
    },
    {
      ...node('cpu_check', 'root', 'CPU检查'),
      hydrationStatus: 'failed',
      errorMessage: 'old error',
    },
  ])

  assert.deepEqual(
    result.map((item) => [item.nodeId, item.hydrationStatus, item.content, item.errorMessage]),
    [
      ['root', 'pending', undefined, undefined],
      ['cpu_check', 'pending', undefined, undefined],
    ],
  )
})

test('node_started only marks the target node running', () => {
  const result = applyFindingsStreamEvent([node('root'), node('cpu_check', 'root')], {
    type: 'node_started',
    hydrationId: 'hydration-1',
    nodeId: 'cpu_check',
    attempt: 1,
  })

  assert.equal(result.find((item) => item.nodeId === 'root')?.hydrationStatus, undefined)
  assert.equal(result.find((item) => item.nodeId === 'cpu_check')?.hydrationStatus, 'running')
  assert.equal(result.find((item) => item.nodeId === 'cpu_check')?.attempt, 1)
})

test('node_patch replaces target node and marks hydration success', () => {
  const result = applyFindingsStreamEvent([node('root'), node('cpu_check', 'root')], {
    type: 'node_patch',
    hydrationId: 'hydration-1',
    nodeId: 'cpu_check',
    attempt: 1,
    node: {
      ...node('cpu_check', 'root', 'CPU检查'),
      status: 'abnormal',
      statusColor: 'abnormal',
      summaryConclusion: 'CPU异常',
    },
  })

  const patched = result.find((item) => item.nodeId === 'cpu_check')
  assert.equal(patched?.title, 'CPU检查')
  assert.equal(patched?.status, 'abnormal')
  assert.equal(patched?.hydrationStatus, 'success')
  assert.equal(patched?.attempt, 1)
})

test('node_patch normalizes raw traces before merging stream data', () => {
  const result = applyFindingsStreamEvent([node('root'), node('db_traces_check', 'root')], {
    type: 'node_patch',
    hydrationId: 'hydration-1',
    nodeId: 'db_traces_check',
    attempt: 1,
    node: {
      ...node('db_traces_check', 'root', 'DB traces'),
      traces: [
        {
          serviceName: 'ts-order-service',
          operationName: 'SELECT Order',
          duration: 100,
          isError: false,
        },
      ],
    } as any,
  })

  const patched = result.find((item) => item.nodeId === 'db_traces_check')
  assert.deepEqual(patched?.traces, [
    {
      service: 'ts-order-service',
      calls: [
        {
          target: 'SELECT Order',
          type: 'database',
          count: 1,
          avgDuration: 100,
          maxDuration: 100,
          errorCount: 0,
          operations: [
            {
              name: 'SELECT Order',
              count: 1,
              avgDuration: 100,
            },
          ],
        },
      ],
    },
  ])
  assert.equal(patched?.hydrationStatus, 'success')
})

test('node_failed preserves target node and attaches failure metadata', () => {
  const result = applyFindingsStreamEvent([node('root'), node('code_lock_check', 'root')], {
    type: 'node_failed',
    hydrationId: 'hydration-1',
    nodeId: 'code_lock_check',
    attempt: 1,
    retryable: true,
    errorType: 'timeout',
    message: 'query timed out',
  })

  const failed = result.find((item) => item.nodeId === 'code_lock_check')
  assert.equal(failed?.hydrationStatus, 'failed')
  assert.equal(failed?.retryable, true)
  assert.equal(failed?.errorType, 'timeout')
  assert.equal(failed?.errorMessage, 'query timed out')
})

test('node_skipped removes target node and descendants', () => {
  const result = applyFindingsStreamEvent(
    [
      node('root'),
      node('dependency_check', 'root'),
      node('dependency_red_check', 'dependency_check'),
      node('dependency_red_child', 'dependency_red_check'),
      node('code_check', 'root'),
    ],
    {
      type: 'node_skipped',
      hydrationId: 'hydration-1',
      nodeId: 'dependency_check',
      reason: 'no-downstream-services',
    },
  )

  assert.deepEqual(
    result.map((item) => item.nodeId),
    ['root', 'code_check'],
  )
})
