import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildConversationHistoryItemViewModels,
  buildFinalMessageShareUrl,
  buildThreadSearchParams,
  getFinalMessageCopySuccessKey,
  getFinalMessageCopyText,
  isFinalMessageFeedbackSelected,
  resolveChatMainConversationSurface,
  resolveChatMainInlineAlert,
  shouldRenderFinalMessageSuffix,
  shouldShowMissingAssistantResponseNotice,
  resolveConversationHistoryPanelMotionState,
  resolveConversationHistoryPanelLayout,
  resolveConversationHistoryPanelState,
  resolveConversationHistorySelectionThreadId,
  resolveRecoverableRouteThreadId,
  shouldAllowAgentHistoryIdentityRequest,
  shouldMaterializePendingNewThreadHistory,
  shouldShowRecoverableRouteThreadErrorToast,
  shouldRecoverFailedRouteThreadOnSend,
  shouldWriteActiveThreadToRoute,
} from './conversationHistoryView'

const messages = {
  minutesAgo: (value: number) => `${value} 分钟前`,
  hoursAgo: (value: number) => `${value} 小时前`,
  yesterday: (value: string) => `昨天 ${value}`,
  daysAgo: (value: number) => `${value} 天前`,
}

test('builds history list item view models with formatted time and active state', () => {
  const items = buildConversationHistoryItemViewModels(
    [
      {
        threadId: 'thread-active',
        title: 'Active thread',
        timeValue: '2026-04-23T10:00:00',
      },
      {
        threadId: 'thread-idle',
        title: 'Idle thread',
        timeValue: '2026-04-22T09:15:00',
      },
    ],
    {
      activeThreadId: 'thread-active',
      messages,
      now: new Date('2026-04-23T10:30:00'),
    },
  )

  assert.deepEqual(items, [
    {
      threadId: 'thread-active',
      title: 'Active thread',
      timeLabel: '30 分钟前',
      isActive: true,
    },
    {
      threadId: 'thread-idle',
      title: 'Idle thread',
      timeLabel: '昨天 09:15',
      isActive: false,
    },
  ])
})

test('builds next search params for selecting and clearing a thread without dropping unrelated params', () => {
  assert.equal(buildThreadSearchParams('foo=bar', 'thread-123'), 'foo=bar&threadId=thread-123')
  assert.equal(
    buildThreadSearchParams('foo=bar&threadId=thread-old', 'thread-new'),
    'foo=bar&threadId=thread-new',
  )
  assert.equal(buildThreadSearchParams('foo=bar&threadId=thread-123', undefined), 'foo=bar')
  assert.equal(buildThreadSearchParams('', undefined), '')
})

test('requires a stable current user id before creator-scoped history requests or first sends', () => {
  assert.equal(shouldAllowAgentHistoryIdentityRequest(undefined), false)
  assert.equal(shouldAllowAgentHistoryIdentityRequest('   '), false)
  assert.equal(shouldAllowAgentHistoryIdentityRequest('user-123'), true)
  assert.equal(shouldAllowAgentHistoryIdentityRequest(42), true)
})

test('writes a new threadId back to the route only while a pending new conversation is being materialized', () => {
  assert.equal(
    shouldWriteActiveThreadToRoute({
      routeThreadId: undefined,
      activeThreadId: 'thread-new',
      hasPendingNewThread: true,
    }),
    true,
  )

  assert.equal(
    shouldWriteActiveThreadToRoute({
      routeThreadId: undefined,
      activeThreadId: 'thread-existing',
      hasPendingNewThread: false,
    }),
    false,
  )

  assert.equal(
    shouldWriteActiveThreadToRoute({
      routeThreadId: 'thread-existing',
      activeThreadId: 'thread-existing',
      hasPendingNewThread: true,
    }),
    false,
  )

  assert.equal(
    shouldWriteActiveThreadToRoute({
      routeThreadId: 'thread-missing',
      activeThreadId: 'thread-new',
      hasPendingNewThread: true,
    }),
    true,
  )
})

test('materializes a pending new thread history item even when recovering from a stale route thread', () => {
  assert.equal(
    shouldMaterializePendingNewThreadHistory({
      hasPendingNewThread: true,
      activeThreadId: 'thread-new',
    }),
    true,
  )

  assert.equal(
    shouldMaterializePendingNewThreadHistory({
      hasPendingNewThread: true,
      activeThreadId: undefined,
    }),
    false,
  )

  assert.equal(
    shouldMaterializePendingNewThreadHistory({
      hasPendingNewThread: false,
      activeThreadId: 'thread-new',
    }),
    false,
  )
})

test('prefers the route threadId for history selection and falls back to the in-memory thread while route writeback catches up', () => {
  assert.equal(
    resolveConversationHistorySelectionThreadId({
      routeThreadId: 'thread-route',
      currentThreadId: 'thread-memory',
    }),
    'thread-route',
  )

  assert.equal(
    resolveConversationHistorySelectionThreadId({
      routeThreadId: undefined,
      currentThreadId: 'thread-memory',
    }),
    'thread-memory',
  )

  assert.equal(
    resolveConversationHistorySelectionThreadId({
      routeThreadId: undefined,
      currentThreadId: undefined,
    }),
    undefined,
  )
})

test('uses blocking loading only when the history list is still empty', () => {
  assert.equal(
    resolveConversationHistoryPanelState({
      itemCount: 0,
      loading: true,
      error: null,
    }),
    'loading',
  )

  assert.equal(
    resolveConversationHistoryPanelState({
      itemCount: 2,
      loading: true,
      error: null,
    }),
    'list',
  )
})

test('keeps rendering the list when items already exist even if the latest refresh errored', () => {
  assert.equal(
    resolveConversationHistoryPanelState({
      itemCount: 1,
      loading: false,
      error: '刷新失败',
    }),
    'list',
  )

  assert.equal(
    resolveConversationHistoryPanelState({
      itemCount: 0,
      loading: false,
      error: '刷新失败',
    }),
    'error',
  )
})

test('keeps the chat surface in a loading state while an empty history thread is loading', () => {
  assert.equal(
    resolveChatMainConversationSurface({
      messageCount: 0,
      loading: true,
    }),
    'loading',
  )

  assert.equal(
    resolveChatMainConversationSurface({
      messageCount: 0,
      loading: false,
    }),
    'empty',
  )

  assert.equal(
    resolveChatMainConversationSurface({
      messageCount: 1,
      loading: true,
    }),
    'messages',
  )
})

test('shows a missing assistant response notice for idle history threads ending with a user turn', () => {
  assert.equal(
    shouldShowMissingAssistantResponseNotice({
      messages: [
        {
          role: 'user',
          content: 'Why is checkout failing?',
        },
      ],
      isHistoryThread: true,
      loading: false,
      isRunning: false,
    }),
    true,
  )

  assert.equal(
    shouldShowMissingAssistantResponseNotice({
      messages: [
        {
          role: 'user',
          content: 'Why is checkout failing?',
        },
      ],
      isHistoryThread: true,
      loading: true,
      isRunning: false,
    }),
    false,
  )

  assert.equal(
    shouldShowMissingAssistantResponseNotice({
      messages: [
        {
          role: 'user',
          content: 'Why is checkout failing?',
        },
        {
          role: 'assistant',
          content: 'The checkout service is timing out.',
        },
      ],
      isHistoryThread: true,
      loading: false,
      isRunning: false,
    }),
    false,
  )

  assert.equal(
    shouldShowMissingAssistantResponseNotice({
      messages: [
        {
          role: 'user',
          content: 'Why is checkout failing?',
        },
      ],
      isHistoryThread: false,
      loading: false,
      isRunning: false,
    }),
    false,
  )
})

test('renders final message suffix only after assistant responses', () => {
  assert.equal(
    shouldRenderFinalMessageSuffix({
      role: 'assistant',
      isFinal: true,
    }),
    true,
  )

  assert.equal(
    shouldRenderFinalMessageSuffix({
      role: 'user',
      isFinal: true,
    }),
    false,
  )

  assert.equal(
    shouldRenderFinalMessageSuffix({
      role: 'assistant',
      isFinal: false,
    }),
    false,
  )
})

test('uses final assistant message content as copy text', () => {
  assert.equal(
    getFinalMessageCopyText({
      role: 'assistant',
      isFinal: true,
      content: 'Root cause summary\n\nRecommended action',
    }),
    'Root cause summary\n\nRecommended action',
  )

  assert.equal(
    getFinalMessageCopyText({
      role: 'user',
      isFinal: true,
      content: 'Do not copy user prompt',
    }),
    '',
  )
})

test('includes final assistant result component text in copy text', () => {
  assert.equal(
    getFinalMessageCopyText({
      role: 'assistant',
      isFinal: true,
      content: 'Analysis complete.',
      agentToolDisplay: [
        {
          toolName: 'result',
          title: 'Result',
          content: {
            fields: [
              {
                key: 'summary',
                label: 'Summary',
                children: 'Root cause: database latency.',
              },
              {
                key: 'recommendation',
                label: 'Recommendation',
                children: 'Scale the database connection pool.',
              },
              {
                key: 'evidence_json',
                label: 'Evidence',
                children: { nodes: [] },
              },
            ],
          },
        },
      ],
    }),
    [
      'Analysis complete.',
      'Summary',
      'Root cause: database latency.',
      'Recommendation',
      'Scale the database connection pool.',
    ].join('\n\n'),
  )
})

test('includes completed findings result text in copy text', () => {
  assert.equal(
    getFinalMessageCopyText({
      role: 'assistant',
      isFinal: true,
      content: 'Findings complete.',
      findingsData: {
        result: {
          title: 'Final Result',
          content: 'The checkout service is impacted by Redis timeouts.',
        },
      },
    }),
    [
      'Findings complete.',
      'Final Result',
      'The checkout service is impacted by Redis timeouts.',
    ].join('\n\n'),
  )
})

test('builds a share URL that preserves existing route params and selects the thread', () => {
  assert.equal(
    buildFinalMessageShareUrl({
      href: 'https://apo.example.com/agent?foo=bar',
      threadId: 'thread-123',
    }),
    'https://apo.example.com/agent?foo=bar&threadId=thread-123',
  )

  assert.equal(
    buildFinalMessageShareUrl({
      href: 'https://apo.example.com/agent?foo=bar&threadId=old',
      threadId: 'thread-123',
    }),
    'https://apo.example.com/agent?foo=bar&threadId=thread-123',
  )
})

test('uses distinct success feedback keys for response and link copy actions', () => {
  assert.equal(getFinalMessageCopySuccessKey('response'), 'messageActions.copyResponseSuccess')
  assert.equal(getFinalMessageCopySuccessKey('share'), 'messageActions.copyLinkSuccess')
})

test('marks only the clicked final-message feedback action as selected', () => {
  assert.equal(isFinalMessageFeedbackSelected('positive', 'positive'), true)
  assert.equal(isFinalMessageFeedbackSelected('positive', 'negative'), false)
  assert.equal(isFinalMessageFeedbackSelected(undefined, 'positive'), false)
})

test('shows an inline chat error alert when a thread-load error is present', () => {
  assert.equal(typeof resolveChatMainInlineAlert, 'function')
  assert.deepEqual(
    resolveChatMainInlineAlert({
      error: 'Failed to load conversation history: thread not found',
    }),
    {
      showError: true,
      message: 'Failed to load conversation history: thread not found',
    },
  )
})

test('suppresses the inline chat error alert when a recoverable route failure should use transient toast feedback', () => {
  assert.deepEqual(
    resolveChatMainInlineAlert({
      error: 'Conversation was not found.',
      hideError: true,
    }),
    {
      showError: false,
    },
  )
})

test('hides the inline chat error alert when there is no usable error message', () => {
  assert.deepEqual(
    resolveChatMainInlineAlert({
      error: '   ',
    }),
    {
      showError: false,
    },
  )
})

test('marks recoverable route-thread errors for a one-shot top-right notification', () => {
  assert.equal(
    shouldShowRecoverableRouteThreadErrorToast({
      routeThreadId: 'thread-missing',
      recoverableRouteThreadId: 'thread-missing',
      error: 'Conversation was not found.',
    }),
    true,
  )

  assert.equal(
    shouldShowRecoverableRouteThreadErrorToast({
      routeThreadId: 'thread-missing',
      recoverableRouteThreadId: undefined,
      error: 'Conversation was not found.',
    }),
    false,
  )
})

test('uses drawer mode on narrow screens and keeps the header toggle available before the first message', () => {
  assert.deepEqual(
    resolveConversationHistoryPanelLayout({
      hasMessages: false,
      useDrawerMode: true,
      isCollapsed: true,
    }),
    {
      showHeader: true,
      showHistoryToggle: true,
      drawerOpen: false,
      siderCollapsed: true,
    },
  )

  assert.deepEqual(
    resolveConversationHistoryPanelLayout({
      hasMessages: false,
      useDrawerMode: true,
      isCollapsed: false,
    }),
    {
      showHeader: true,
      showHistoryToggle: true,
      drawerOpen: true,
      siderCollapsed: false,
    },
  )
})

test('keeps desktop sider behavior while exposing a persistent header toggle', () => {
  assert.deepEqual(
    resolveConversationHistoryPanelLayout({
      hasMessages: false,
      useDrawerMode: false,
      isCollapsed: true,
    }),
    {
      showHeader: true,
      showHistoryToggle: true,
      drawerOpen: false,
      siderCollapsed: true,
    },
  )

  assert.deepEqual(
    resolveConversationHistoryPanelLayout({
      hasMessages: true,
      useDrawerMode: false,
      isCollapsed: false,
    }),
    {
      showHeader: true,
      showHistoryToggle: true,
      drawerOpen: false,
      siderCollapsed: false,
    },
  )
})

test('pins desktop history panel content width and hides content during collapse to avoid text squeezing', () => {
  assert.deepEqual(
    resolveConversationHistoryPanelMotionState({
      useDrawerMode: false,
      isCollapsed: false,
    }),
    {
      keepFixedWidth: true,
      hideContentDuringCollapse: false,
    },
  )

  assert.deepEqual(
    resolveConversationHistoryPanelMotionState({
      useDrawerMode: false,
      isCollapsed: true,
    }),
    {
      keepFixedWidth: true,
      hideContentDuringCollapse: true,
    },
  )

  assert.deepEqual(
    resolveConversationHistoryPanelMotionState({
      useDrawerMode: true,
      isCollapsed: true,
    }),
    {
      keepFixedWidth: false,
      hideContentDuringCollapse: false,
    },
  )
})

test('recovers a failed route thread by treating the next send as a new conversation', () => {
  assert.equal(
    shouldRecoverFailedRouteThreadOnSend({
      routeThreadId: 'thread-missing',
      activeThreadId: undefined,
      loading: false,
      error: 'Failed to load conversation history: thread not found',
      threadLoadErrorKind: 'not_found',
    }),
    true,
  )

  assert.equal(
    shouldRecoverFailedRouteThreadOnSend({
      routeThreadId: 'thread-loading',
      activeThreadId: undefined,
      loading: true,
      error: 'Failed to load conversation history: network timeout',
      threadLoadErrorKind: 'transient',
    }),
    false,
  )

  assert.equal(
    shouldRecoverFailedRouteThreadOnSend({
      routeThreadId: 'thread-loaded',
      activeThreadId: 'thread-loaded',
      loading: false,
      error: 'Previous dismissible error',
      threadLoadErrorKind: 'unknown',
    }),
    false,
  )
})

test('does not recover a failed route thread after a transient load error', () => {
  assert.equal(
    shouldRecoverFailedRouteThreadOnSend({
      routeThreadId: 'thread-transient',
      activeThreadId: undefined,
      loading: false,
      error: 'Failed to load conversation history: network timeout',
      threadLoadErrorKind: 'transient',
      recoverableRouteThreadId: undefined,
    }),
    false,
  )
})

test('retains a recoverable route thread after dismissing the inline error copy', () => {
  assert.equal(
    resolveRecoverableRouteThreadId({
      routeThreadId: 'thread-missing',
      activeThreadId: undefined,
      loading: false,
      error: 'Failed to load conversation history: thread not found',
      threadLoadErrorKind: 'not_found',
      recoverableRouteThreadId: undefined,
    }),
    'thread-missing',
  )

  assert.equal(
    resolveRecoverableRouteThreadId({
      routeThreadId: 'thread-missing',
      activeThreadId: undefined,
      loading: false,
      error: null,
      threadLoadErrorKind: null,
      recoverableRouteThreadId: 'thread-missing',
    }),
    'thread-missing',
  )
})

test('clears recoverable route state once the route changes or a thread becomes active', () => {
  assert.equal(
    resolveRecoverableRouteThreadId({
      routeThreadId: 'thread-other',
      activeThreadId: undefined,
      loading: false,
      error: null,
      threadLoadErrorKind: null,
      recoverableRouteThreadId: 'thread-missing',
    }),
    undefined,
  )

  assert.equal(
    resolveRecoverableRouteThreadId({
      routeThreadId: 'thread-missing',
      activeThreadId: 'thread-new',
      loading: false,
      error: null,
      threadLoadErrorKind: null,
      recoverableRouteThreadId: 'thread-missing',
    }),
    undefined,
  )
})

test('keeps failed route recovery active after the inline error copy is dismissed', () => {
  const dismissedInlineAlertState = {
    routeThreadId: 'thread-missing',
    activeThreadId: undefined,
    loading: false,
    error: null,
    threadLoadErrorKind: null,
    recoverableRouteThreadId: 'thread-missing',
  }

  assert.equal(shouldRecoverFailedRouteThreadOnSend(dismissedInlineAlertState), true)
})
