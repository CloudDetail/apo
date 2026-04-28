import {
  formatConversationHistoryTime,
  type ConversationHistoryItem,
  type ConversationHistoryTimeMessages,
} from './conversationHistory'

export interface ConversationHistoryItemViewModel {
  threadId: string
  title: string
  timeLabel: string
  isActive: boolean
}

export type ConversationHistoryPanelState = 'loading' | 'error' | 'empty' | 'list'

export type ChatMainConversationSurface = 'loading' | 'empty' | 'messages'

export interface ConversationHistoryPanelLayout {
  showHeader: boolean
  showHistoryToggle: boolean
  drawerOpen: boolean
  siderCollapsed: boolean
}

export interface ConversationHistoryPanelMotionState {
  keepFixedWidth: boolean
  hideContentDuringCollapse: boolean
}

export interface ChatMainInlineAlertState {
  showError: boolean
  message?: string
}

export interface ChatHistoryMessageLike {
  role?: string
  hidden?: boolean
  content?: unknown
}

export interface FinalMessageSuffixMessageLike {
  role?: string
  isFinal?: boolean
}

export const resolveChatMainConversationSurface = ({
  messageCount,
  loading,
}: {
  messageCount: number
  loading: boolean
}): ChatMainConversationSurface => {
  if (messageCount > 0) {
    return 'messages'
  }

  if (loading) {
    return 'loading'
  }

  return 'empty'
}

export const buildConversationHistoryItemViewModels = (
  items: ConversationHistoryItem[],
  {
    activeThreadId,
    messages,
    now,
  }: {
    activeThreadId?: string
    messages: ConversationHistoryTimeMessages
    now?: Date
  },
): ConversationHistoryItemViewModel[] =>
  items.map((item) => ({
    threadId: item.threadId,
    title: item.title,
    timeLabel: formatConversationHistoryTime(item.timeValue, messages, now),
    isActive: item.threadId === activeThreadId,
  }))

export const resolveConversationHistoryPanelState = ({
  itemCount,
  loading,
  error,
}: {
  itemCount: number
  loading: boolean
  error: string | null
}): ConversationHistoryPanelState => {
  if (itemCount > 0) {
    return 'list'
  }

  if (loading) {
    return 'loading'
  }

  if (error) {
    return 'error'
  }

  return 'empty'
}

export const resolveConversationHistoryPanelLayout = ({
  hasMessages,
  useDrawerMode,
  isCollapsed,
}: {
  hasMessages: boolean
  useDrawerMode: boolean
  isCollapsed: boolean
}): ConversationHistoryPanelLayout => ({
  showHeader: true,
  showHistoryToggle: true,
  drawerOpen: useDrawerMode && !isCollapsed,
  siderCollapsed: isCollapsed,
})

export const resolveConversationHistoryPanelMotionState = ({
  useDrawerMode,
  isCollapsed,
}: {
  useDrawerMode: boolean
  isCollapsed: boolean
}): ConversationHistoryPanelMotionState => ({
  keepFixedWidth: !useDrawerMode,
  hideContentDuringCollapse: !useDrawerMode && isCollapsed,
})

export const resolveChatMainInlineAlert = ({
  error,
  hideError = false,
}: {
  error: string | null
  hideError?: boolean
}): ChatMainInlineAlertState => {
  if (hideError) {
    return {
      showError: false,
    }
  }

  const message = typeof error === 'string' ? error.trim() : ''

  if (!message) {
    return {
      showError: false,
    }
  }

  return {
    showError: true,
    message,
  }
}

export const shouldShowMissingAssistantResponseNotice = ({
  messages,
  isHistoryThread,
  loading,
  isRunning,
}: {
  messages: ChatHistoryMessageLike[]
  isHistoryThread: boolean
  loading: boolean
  isRunning: boolean
}) => {
  if (!isHistoryThread || loading || isRunning) {
    return false
  }

  const visibleMessages = messages.filter((message) => !message.hidden)
  const lastMessage = visibleMessages[visibleMessages.length - 1]

  return lastMessage?.role === 'user'
}

export const shouldRenderFinalMessageSuffix = (message: FinalMessageSuffixMessageLike) =>
  Boolean(message.isFinal && message.role === 'assistant')

export const buildThreadSearchParams = (search: string, threadId?: string) => {
  const params = new URLSearchParams(search)

  if (threadId) {
    params.set('threadId', threadId)
  } else {
    params.delete('threadId')
  }

  return params.toString()
}

export const shouldWriteActiveThreadToRoute = ({
  routeThreadId,
  activeThreadId,
  hasPendingNewThread,
}: {
  routeThreadId?: string
  activeThreadId?: string
  hasPendingNewThread: boolean
}) => Boolean(activeThreadId && hasPendingNewThread && routeThreadId !== activeThreadId)

export const shouldMaterializePendingNewThreadHistory = ({
  hasPendingNewThread,
  activeThreadId,
}: {
  hasPendingNewThread: boolean
  activeThreadId?: string
}) => Boolean(hasPendingNewThread && activeThreadId)

export const shouldAllowAgentHistoryIdentityRequest = (userId?: string | number) => {
  if (typeof userId !== 'string' && typeof userId !== 'number') {
    return false
  }

  return String(userId).trim().length > 0
}

export const resolveRecoverableRouteThreadId = ({
  routeThreadId,
  activeThreadId,
  loading,
  error,
  threadLoadErrorKind,
  recoverableRouteThreadId,
}: {
  routeThreadId?: string
  activeThreadId?: string
  loading: boolean
  error: string | null
  threadLoadErrorKind?: string | null
  recoverableRouteThreadId?: string
}) => {
  if (!routeThreadId || activeThreadId) {
    return undefined
  }

  if (recoverableRouteThreadId && recoverableRouteThreadId !== routeThreadId) {
    return undefined
  }

  if (!loading && error && threadLoadErrorKind === 'not_found') {
    return routeThreadId
  }

  if (recoverableRouteThreadId === routeThreadId) {
    return recoverableRouteThreadId
  }

  return undefined
}

export const shouldShowRecoverableRouteThreadErrorToast = ({
  routeThreadId,
  recoverableRouteThreadId,
  error,
}: {
  routeThreadId?: string
  recoverableRouteThreadId?: string
  error: string | null
}) => {
  const message = typeof error === 'string' ? error.trim() : ''

  return Boolean(
    routeThreadId &&
      recoverableRouteThreadId &&
      routeThreadId === recoverableRouteThreadId &&
      message,
  )
}

export const shouldRecoverFailedRouteThreadOnSend = ({
  routeThreadId,
  activeThreadId,
  loading,
  error,
  threadLoadErrorKind,
  recoverableRouteThreadId,
}: {
  routeThreadId?: string
  activeThreadId?: string
  loading: boolean
  error: string | null
  threadLoadErrorKind?: string | null
  recoverableRouteThreadId?: string
}) =>
  Boolean(
    routeThreadId &&
      !activeThreadId &&
      !loading &&
      (recoverableRouteThreadId === routeThreadId ||
        (threadLoadErrorKind === 'not_found' && error)),
  )

export const resolveConversationHistorySelectionThreadId = ({
  routeThreadId,
  currentThreadId,
}: {
  routeThreadId?: string
  currentThreadId?: string
}) => routeThreadId ?? currentThreadId
