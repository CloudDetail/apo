import test from 'node:test'
import assert from 'node:assert/strict'

import {
  appendConversationHistoryPageItems,
  formatConversationHistoryTime,
  getNextConversationHistoryPendingRefreshState,
  getConversationHistoryReconcileDelayMs,
  hasConversationHistoryItem,
  mergeConversationHistoryItems,
  normalizeConversationHistoryListResponse,
  removeConversationHistoryItem,
  resolveConversationHistoryReconcileDecision,
  shouldClearConversationHistoryLoadMoreState,
  shouldApplyConversationHistoryLoadMoreResponse,
  shouldScheduleConversationHistoryReconcile,
  shouldApplyConversationHistoryResponse,
  upsertConversationHistoryItem,
} from './conversationHistory'
import {
  buildConversationHistoryListRequestParams,
  buildConversationMessagesRequestParams,
} from './conversationRequestParams'

const messages = {
  minutesAgo: (value: number) => `${value} minutes ago`,
  hoursAgo: (value: number) => `${value} hours ago`,
  yesterday: (value: string) => `Yesterday ${value}`,
  daysAgo: (value: number) => `${value} days ago`,
}

test('normalizes list payloads, inserts fallback titles, and sorts by timeValue descending', () => {
  const response = {
    data: {
      items: [
        {
          threadId: 'thread-older',
          title: 'Older conversation',
          timeValue: '2026-04-22T08:00:00',
        },
        {
          threadId: 'thread-newer',
          title: '   ',
          timeValue: '2026-04-23T08:00:00',
        },
      ],
      nextCursor: 'cursor-2',
      hasMore: true,
    },
  }

  const normalized = normalizeConversationHistoryListResponse(response, 'Untitled thread')

  assert.deepEqual(normalized, {
    items: [
      {
        threadId: 'thread-newer',
        title: 'Untitled thread',
        timeValue: '2026-04-23T08:00:00',
      },
      {
        threadId: 'thread-older',
        title: 'Older conversation',
        timeValue: '2026-04-22T08:00:00',
      },
    ],
    nextCursor: 'cursor-2',
    hasMore: true,
  })
})

test('accepts canonical owner-side conversations and continuation envelopes', () => {
  const normalized = normalizeConversationHistoryListResponse(
    {
      status: 'ok',
      conversations: [
        {
          threadId: 'thread-newer',
          title: 'Newest conversation',
          timeValue: '2026-04-23T08:00:00Z',
        },
        {
          threadId: 'thread-older',
          title: 'Older conversation',
          timeValue: '2026-04-22T08:00:00Z',
        },
      ],
      continuation: {
        nextCursor: 'cursor-canonical',
      },
    },
    'Untitled thread',
  )

  assert.deepEqual(normalized, {
    items: [
      {
        threadId: 'thread-newer',
        title: 'Newest conversation',
        timeValue: '2026-04-23T08:00:00Z',
      },
      {
        threadId: 'thread-older',
        title: 'Older conversation',
        timeValue: '2026-04-22T08:00:00Z',
      },
    ],
    nextCursor: 'cursor-canonical',
    hasMore: true,
  })
})

test('accepts raw array payloads and drops invalid items', () => {
  const normalized = normalizeConversationHistoryListResponse(
    [
      {
        threadId: 'thread-valid',
        title: 'Valid conversation',
        timeValue: '2026-04-23T08:00:00',
      },
      {
        threadId: '',
        title: 'Missing id',
        timeValue: '2026-04-23T08:30:00',
      },
      {
        threadId: 'thread-missing-time',
        title: 'Missing time',
      },
    ],
    'Untitled thread',
  )

  assert.deepEqual(normalized, {
    items: [
      {
        threadId: 'thread-valid',
        title: 'Valid conversation',
        timeValue: '2026-04-23T08:00:00',
      },
    ],
    hasMore: false,
  })
})

test('uses a deterministic fallback title even when caller does not pass one', () => {
  const normalized = normalizeConversationHistoryListResponse([
    {
      threadId: 'thread-without-title',
      timeValue: '2026-04-23T08:00:00',
    },
  ])

  assert.deepEqual(normalized, {
    items: [
      {
        threadId: 'thread-without-title',
        title: 'Untitled conversation',
        timeValue: '2026-04-23T08:00:00',
      },
    ],
    hasMore: false,
  })
})

test('upserts a new history thread and keeps the list sorted by timeValue descending', () => {
  const items = upsertConversationHistoryItem(
    [
      {
        threadId: 'thread-older',
        title: 'Older conversation',
        timeValue: '2026-04-22T08:00:00',
      },
    ],
    {
      threadId: 'thread-new',
      title: 'Fresh conversation',
      timeValue: '2026-04-23T08:00:00',
    },
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-new',
      title: 'Fresh conversation',
      timeValue: '2026-04-23T08:00:00',
    },
    {
      threadId: 'thread-older',
      title: 'Older conversation',
      timeValue: '2026-04-22T08:00:00',
    },
  ])
})

test('upserts an existing history thread by replacing the old entry instead of duplicating it', () => {
  const items = upsertConversationHistoryItem(
    [
      {
        threadId: 'thread-existing',
        title: 'Old title',
        timeValue: '2026-04-22T08:00:00',
      },
      {
        threadId: 'thread-other',
        title: 'Other conversation',
        timeValue: '2026-04-21T08:00:00',
      },
    ],
    {
      threadId: 'thread-existing',
      title: 'Updated title',
      timeValue: '2026-04-23T08:00:00',
    },
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-existing',
      title: 'Updated title',
      timeValue: '2026-04-23T08:00:00',
    },
    {
      threadId: 'thread-other',
      title: 'Other conversation',
      timeValue: '2026-04-21T08:00:00',
    },
  ])
})

test('merges an optimistic history thread into fetched items so late responses do not hide it', () => {
  const items = mergeConversationHistoryItems(
    [
      {
        threadId: 'thread-existing',
        title: 'Existing conversation',
        timeValue: '2026-04-22T08:00:00',
      },
    ],
    {
      threadId: 'thread-optimistic',
      title: 'Fresh conversation',
      timeValue: '2026-04-23T08:00:00',
    },
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-optimistic',
      title: 'Fresh conversation',
      timeValue: '2026-04-23T08:00:00',
    },
    {
      threadId: 'thread-existing',
      title: 'Existing conversation',
      timeValue: '2026-04-22T08:00:00',
    },
  ])
})

test('keeps the backend item when the fetched list already contains the optimistic thread', () => {
  const items = mergeConversationHistoryItems(
    [
      {
        threadId: 'thread-optimistic',
        title: 'Persisted conversation',
        timeValue: '2026-04-23T08:30:00',
      },
      {
        threadId: 'thread-existing',
        title: 'Existing conversation',
        timeValue: '2026-04-22T08:00:00',
      },
    ],
    {
      threadId: 'thread-optimistic',
      title: 'Fresh conversation',
      timeValue: '2026-04-23T08:00:00',
    },
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-optimistic',
      title: 'Persisted conversation',
      timeValue: '2026-04-23T08:30:00',
    },
    {
      threadId: 'thread-existing',
      title: 'Existing conversation',
      timeValue: '2026-04-22T08:00:00',
    },
  ])
})

test('keeps a previously loaded selected older thread visible when a first-page refresh does not include it', () => {
  const items = mergeConversationHistoryItems(
    [
      {
        threadId: 'thread-newest',
        title: 'Newest conversation',
        timeValue: '2026-04-23T10:00:00',
      },
      {
        threadId: 'thread-recent',
        title: 'Recent conversation',
        timeValue: '2026-04-23T09:00:00',
      },
    ],
    undefined,
    {
      threadId: 'thread-selected-older',
      title: 'Selected older conversation',
      timeValue: '2026-04-22T08:00:00',
    },
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-newest',
      title: 'Newest conversation',
      timeValue: '2026-04-23T10:00:00',
    },
    {
      threadId: 'thread-recent',
      title: 'Recent conversation',
      timeValue: '2026-04-23T09:00:00',
    },
    {
      threadId: 'thread-selected-older',
      title: 'Selected older conversation',
      timeValue: '2026-04-22T08:00:00',
    },
  ])
})

test('appends a paginated history page with stable descending order and without duplicates', () => {
  const items = appendConversationHistoryPageItems(
    [
      {
        threadId: 'thread-newest',
        title: 'Newest conversation',
        timeValue: '2026-04-23T10:00:00',
      },
      {
        threadId: 'thread-overlap',
        title: 'Already loaded conversation',
        timeValue: '2026-04-23T09:00:00',
      },
    ],
    [
      {
        threadId: 'thread-overlap',
        title: 'Older duplicate from next page',
        timeValue: '2026-04-23T08:30:00',
      },
      {
        threadId: 'thread-older',
        title: 'Older conversation',
        timeValue: '2026-04-22T12:00:00',
      },
    ],
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-newest',
      title: 'Newest conversation',
      timeValue: '2026-04-23T10:00:00',
    },
    {
      threadId: 'thread-overlap',
      title: 'Already loaded conversation',
      timeValue: '2026-04-23T09:00:00',
    },
    {
      threadId: 'thread-older',
      title: 'Older conversation',
      timeValue: '2026-04-22T12:00:00',
    },
  ])
})

test('keeps the optimistic thread visible while appending older history pages', () => {
  const items = appendConversationHistoryPageItems(
    [
      {
        threadId: 'thread-existing',
        title: 'Existing conversation',
        timeValue: '2026-04-23T09:00:00',
      },
    ],
    [
      {
        threadId: 'thread-older',
        title: 'Older conversation',
        timeValue: '2026-04-22T12:00:00',
      },
    ],
    {
      threadId: 'thread-optimistic',
      title: 'Fresh optimistic conversation',
      timeValue: '2026-04-23T10:00:00',
    },
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-optimistic',
      title: 'Fresh optimistic conversation',
      timeValue: '2026-04-23T10:00:00',
    },
    {
      threadId: 'thread-existing',
      title: 'Existing conversation',
      timeValue: '2026-04-23T09:00:00',
    },
    {
      threadId: 'thread-older',
      title: 'Older conversation',
      timeValue: '2026-04-22T12:00:00',
    },
  ])
})

test('removes a dropped optimistic history thread from the rendered list', () => {
  const items = removeConversationHistoryItem(
    [
      {
        threadId: 'thread-optimistic',
        title: 'Fresh conversation',
        timeValue: '2026-04-23T08:00:00',
      },
      {
        threadId: 'thread-existing',
        title: 'Existing conversation',
        timeValue: '2026-04-22T08:00:00',
      },
    ],
    'thread-optimistic',
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-existing',
      title: 'Existing conversation',
      timeValue: '2026-04-22T08:00:00',
    },
  ])
})

test('detects whether a fetched history payload already contains a target thread', () => {
  const items = [
    {
      threadId: 'thread-123',
      title: 'Conversation',
      timeValue: '2026-04-23T08:00:00',
    },
  ]

  assert.equal(hasConversationHistoryItem(items, 'thread-123'), true)
  assert.equal(hasConversationHistoryItem(items, 'thread-456'), false)
  assert.equal(hasConversationHistoryItem(items, undefined), false)
})

test('only applies the newest in-flight history response revision', () => {
  assert.equal(
    shouldApplyConversationHistoryResponse({
      activeRequestRevision: 3,
      responseRequestRevision: 3,
    }),
    true,
  )
  assert.equal(
    shouldApplyConversationHistoryResponse({
      activeRequestRevision: 4,
      responseRequestRevision: 3,
    }),
    false,
  )
})

test('only applies a load-more response when both revision and cursor still match', () => {
  assert.equal(
    shouldApplyConversationHistoryLoadMoreResponse({
      activeRequestRevision: 2,
      responseRequestRevision: 2,
      activeCursor: 'cursor-2',
      responseCursor: 'cursor-2',
    }),
    true,
  )
  assert.equal(
    shouldApplyConversationHistoryLoadMoreResponse({
      activeRequestRevision: 3,
      responseRequestRevision: 2,
      activeCursor: 'cursor-2',
      responseCursor: 'cursor-2',
    }),
    false,
  )
  assert.equal(
    shouldApplyConversationHistoryLoadMoreResponse({
      activeRequestRevision: 2,
      responseRequestRevision: 2,
      activeCursor: 'cursor-3',
      responseCursor: 'cursor-2',
    }),
    false,
  )
})

test('still clears the active load-more spinner when a full refresh changed the cursor mid-request', () => {
  assert.equal(
    shouldApplyConversationHistoryLoadMoreResponse({
      activeRequestRevision: 2,
      responseRequestRevision: 2,
      activeCursor: 'cursor-3',
      responseCursor: 'cursor-2',
    }),
    false,
  )
  assert.equal(
    shouldClearConversationHistoryLoadMoreState({
      activeRequestRevision: 2,
      responseRequestRevision: 2,
    }),
    true,
  )
  assert.equal(
    shouldClearConversationHistoryLoadMoreState({
      activeRequestRevision: 3,
      responseRequestRevision: 2,
    }),
    false,
  )
})

test('keeps retrying reconcile until the backend payload confirms the pending thread', () => {
  const decision = resolveConversationHistoryReconcileDecision({
    attemptCount: 1,
    refreshOutcome: 'applied',
    persistedPendingThread: false,
  })

  assert.deepEqual(decision, {
    shouldClearOptimistic: false,
    shouldClearPending: false,
    nextAttemptCount: 2,
  })
})

test('retries reconcile after a failed refresh instead of leaving pending stuck', () => {
  const decision = resolveConversationHistoryReconcileDecision({
    attemptCount: 0,
    refreshOutcome: 'error',
    persistedPendingThread: false,
  })

  assert.deepEqual(decision, {
    shouldClearOptimistic: false,
    shouldClearPending: false,
    nextAttemptCount: 1,
  })
})

test('does not consume reconcile retry budget when a refresh was superseded by a newer request', () => {
  const decision = resolveConversationHistoryReconcileDecision({
    attemptCount: 2,
    refreshOutcome: 'superseded',
    persistedPendingThread: false,
  })

  assert.deepEqual(decision, {
    shouldClearOptimistic: false,
    shouldClearPending: false,
    nextAttemptCount: 2,
  })
})

test('reuses the same pending refresh object when superseded reconcile keeps the same attempt count', () => {
  const current = {
    threadId: 'thread-pending',
    attemptCount: 2,
  }

  const next = getNextConversationHistoryPendingRefreshState({
    current,
    targetThreadId: 'thread-pending',
    nextAttemptCount: 2,
  })

  assert.equal(next, current)
})

test('drops the optimistic thread once reconcile exhausts its retry budget without backend confirmation', () => {
  const decision = resolveConversationHistoryReconcileDecision({
    attemptCount: 3,
    refreshOutcome: 'applied',
    persistedPendingThread: false,
  })

  assert.deepEqual(decision, {
    shouldClearOptimistic: true,
    shouldClearPending: true,
  })
})

test('uses an immediate first reconcile attempt and delayed follow-up retries', () => {
  assert.equal(getConversationHistoryReconcileDelayMs(0), 0)
  assert.equal(getConversationHistoryReconcileDelayMs(1), 1000)
  assert.equal(getConversationHistoryReconcileDelayMs(2), 2000)
  assert.equal(getConversationHistoryReconcileDelayMs(3), 4000)
})

test('schedules reconcile whenever pending exists and the chat run itself is idle', () => {
  assert.equal(
    shouldScheduleConversationHistoryReconcile({
      hasPendingRefresh: true,
      isRunning: false,
      loading: false,
    }),
    true,
  )
  assert.equal(
    shouldScheduleConversationHistoryReconcile({
      hasPendingRefresh: true,
      isRunning: true,
      loading: false,
    }),
    false,
  )
  assert.equal(
    shouldScheduleConversationHistoryReconcile({
      hasPendingRefresh: true,
      isRunning: false,
      loading: true,
    }),
    false,
  )
})

test('formats minutes and hours using relative thresholds inside the same day', () => {
  const now = new Date('2026-04-23T10:30:00')

  assert.equal(
    formatConversationHistoryTime('2026-04-23T10:00:00', messages, now),
    '30 minutes ago',
  )
  assert.equal(formatConversationHistoryTime('2026-04-23T08:00:00', messages, now), '2 hours ago')
})

test('formats yesterday, days, and absolute dates using documented thresholds', () => {
  const now = new Date('2026-04-23T10:30:00')

  assert.equal(
    formatConversationHistoryTime('2026-04-22T09:15:00', messages, now),
    'Yesterday 09:15',
  )
  assert.equal(formatConversationHistoryTime('2026-04-20T08:00:00', messages, now), '3 days ago')
  assert.equal(
    formatConversationHistoryTime('2026-04-10T08:05:00', messages, now),
    '2026-04-10 08:05',
  )
})

test('builds replay request params without creator-scoped userId filtering', () => {
  assert.deepEqual(buildConversationMessagesRequestParams('thread-shared'), {
    threadId: 'thread-shared',
  })
})

test('builds creator-scoped history list params with explicit userId when provided', () => {
  assert.deepEqual(
    buildConversationHistoryListRequestParams(
      {
        cursor: 'cursor-2',
        limit: 30,
      },
      'user-123',
    ),
    {
      cursor: 'cursor-2',
      limit: 30,
      userId: 'user-123',
    },
  )
})
