export interface ConversationHistoryItem {
  threadId: string
  title: string
  timeValue: string
}

export interface ConversationHistoryListResponse {
  items: ConversationHistoryItem[]
  nextCursor?: string
  hasMore: boolean
}

export interface ConversationHistoryReconcileDecision {
  shouldClearOptimistic: boolean
  shouldClearPending: boolean
  nextAttemptCount?: number
}

export interface ConversationHistoryPendingRefreshState {
  threadId: string
  attemptCount: number
}

export type ConversationHistoryRefreshOutcome = 'applied' | 'error' | 'superseded'

export interface ConversationHistoryTimeMessages {
  minutesAgo: (value: number) => string
  hoursAgo: (value: number) => string
  yesterday: (value: string) => string
  daysAgo: (value: number) => string
}

export const DEFAULT_CONVERSATION_HISTORY_FALLBACK_TITLE = 'Untitled conversation'
export const MAX_CONVERSATION_HISTORY_RECONCILE_ATTEMPTS = 4

const DAY_IN_MS = 24 * 60 * 60 * 1000

const isRecord = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === 'object' && !Array.isArray(value)

const readString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

const readBoolean = (value: unknown): boolean | undefined => {
  if (typeof value === 'boolean') {
    return value
  }

  return undefined
}

const parseTimeValue = (timeValue: string): Date | null => {
  const parsed = new Date(timeValue)
  return Number.isNaN(parsed.getTime()) ? null : parsed
}

const formatTwoDigits = (value: number) => value.toString().padStart(2, '0')

const formatClockTime = (value: Date) =>
  `${formatTwoDigits(value.getHours())}:${formatTwoDigits(value.getMinutes())}`

const formatAbsoluteDateTime = (value: Date) =>
  `${value.getFullYear()}-${formatTwoDigits(value.getMonth() + 1)}-${formatTwoDigits(
    value.getDate(),
  )} ${formatTwoDigits(value.getHours())}:${formatTwoDigits(value.getMinutes())}`

const getCalendarDayDiff = (target: Date, now: Date) => {
  const targetDayStart = new Date(target.getFullYear(), target.getMonth(), target.getDate())
  const nowDayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  return Math.round((nowDayStart.getTime() - targetDayStart.getTime()) / DAY_IN_MS)
}

const normalizeConversationHistoryItem = (
  value: unknown,
  fallbackTitle: string,
): ConversationHistoryItem | null => {
  if (!isRecord(value)) {
    return null
  }

  const threadId = readString(value.threadId)
  const timeValue = readString(value.timeValue)
  if (!threadId || !timeValue || !parseTimeValue(timeValue)) {
    return null
  }

  return {
    threadId,
    title: readString(value.title) ?? fallbackTitle,
    timeValue,
  }
}

const extractConversationHistoryEnvelope = (value: unknown) => {
  if (Array.isArray(value)) {
    return {
      rawItems: value,
      hasMore: undefined as boolean | undefined,
      nextCursor: undefined as string | undefined,
    }
  }

  if (!isRecord(value)) {
    return {
      rawItems: [] as unknown[],
      hasMore: false,
      nextCursor: undefined,
    }
  }

  if (Array.isArray(value.items)) {
    return {
      rawItems: value.items,
      hasMore: readBoolean(value.hasMore),
      nextCursor: readString(value.nextCursor),
    }
  }

  if (isRecord(value.data) && Array.isArray(value.data.items)) {
    return {
      rawItems: value.data.items,
      hasMore: readBoolean(value.data.hasMore),
      nextCursor: readString(value.data.nextCursor),
    }
  }

  if (Array.isArray(value.conversations)) {
    return {
      rawItems: value.conversations,
      hasMore:
        readBoolean(value.hasMore) ??
        (isRecord(value.continuation) ? true : undefined),
      nextCursor: isRecord(value.continuation)
        ? readString(value.continuation.nextCursor)
        : readString(value.nextCursor),
    }
  }

  if (isRecord(value.data) && Array.isArray(value.data.conversations)) {
    return {
      rawItems: value.data.conversations,
      hasMore:
        readBoolean(value.data.hasMore) ??
        (isRecord(value.data.continuation) ? true : undefined),
      nextCursor: isRecord(value.data.continuation)
        ? readString(value.data.continuation.nextCursor)
        : readString(value.data.nextCursor),
    }
  }

  return {
    rawItems: [] as unknown[],
    hasMore: false,
    nextCursor: undefined,
  }
}

export const normalizeConversationHistoryListResponse = (
  value: unknown,
  fallbackTitle = DEFAULT_CONVERSATION_HISTORY_FALLBACK_TITLE,
): ConversationHistoryListResponse => {
  const envelope = extractConversationHistoryEnvelope(value)

  const items = envelope.rawItems
    .map((item) => normalizeConversationHistoryItem(item, fallbackTitle))
    .filter((item): item is ConversationHistoryItem => item !== null)
    .sort((left, right) => {
      const leftTime = parseTimeValue(left.timeValue)?.getTime() ?? 0
      const rightTime = parseTimeValue(right.timeValue)?.getTime() ?? 0
      return rightTime - leftTime
    })

  const normalized: ConversationHistoryListResponse = {
    items,
    hasMore: envelope.hasMore ?? Boolean(envelope.nextCursor),
  }

  if (envelope.nextCursor) {
    normalized.nextCursor = envelope.nextCursor
  }

  return normalized
}

export const upsertConversationHistoryItem = (
  items: ConversationHistoryItem[],
  nextItem: ConversationHistoryItem,
) =>
  [...items.filter((item) => item.threadId !== nextItem.threadId), nextItem].sort((left, right) => {
    const leftTime = parseTimeValue(left.timeValue)?.getTime() ?? 0
    const rightTime = parseTimeValue(right.timeValue)?.getTime() ?? 0
    return rightTime - leftTime
  })

export const hasConversationHistoryItem = (
  items: ConversationHistoryItem[],
  threadId?: string,
) => Boolean(threadId && items.some((item) => item.threadId === threadId))

export const mergeConversationHistoryItems = (
  items: ConversationHistoryItem[],
  optimisticItem?: ConversationHistoryItem,
  retainedItem?: ConversationHistoryItem,
) => {
  let nextItems = items

  for (const item of [optimisticItem, retainedItem]) {
    if (!item || hasConversationHistoryItem(nextItems, item.threadId)) {
      continue
    }

    nextItems = upsertConversationHistoryItem(nextItems, item)
  }

  return nextItems
}

export const appendConversationHistoryPageItems = (
  currentItems: ConversationHistoryItem[],
  nextPageItems: ConversationHistoryItem[],
  optimisticItem?: ConversationHistoryItem,
) => {
  const mergedByThreadId = new Map(currentItems.map((item) => [item.threadId, item]))

  nextPageItems.forEach((item) => {
    const currentItem = mergedByThreadId.get(item.threadId)
    const currentTime = parseTimeValue(currentItem?.timeValue ?? '')?.getTime() ?? 0
    const nextTime = parseTimeValue(item.timeValue)?.getTime() ?? 0

    if (!currentItem || nextTime > currentTime) {
      mergedByThreadId.set(item.threadId, item)
    }
  })

  return mergeConversationHistoryItems(
    [...mergedByThreadId.values()].sort((left, right) => {
      const leftTime = parseTimeValue(left.timeValue)?.getTime() ?? 0
      const rightTime = parseTimeValue(right.timeValue)?.getTime() ?? 0
      return rightTime - leftTime
    }),
    optimisticItem,
  )
}

export const removeConversationHistoryItem = (
  items: ConversationHistoryItem[],
  threadId: string,
) => items.filter((item) => item.threadId !== threadId)

export const shouldApplyConversationHistoryResponse = ({
  activeRequestRevision,
  responseRequestRevision,
}: {
  activeRequestRevision: number
  responseRequestRevision: number
}) => activeRequestRevision === responseRequestRevision

export const shouldApplyConversationHistoryLoadMoreResponse = ({
  activeRequestRevision,
  responseRequestRevision,
  activeCursor,
  responseCursor,
}: {
  activeRequestRevision: number
  responseRequestRevision: number
  activeCursor?: string
  responseCursor?: string
}) =>
  activeRequestRevision === responseRequestRevision &&
  (activeCursor ?? undefined) === (responseCursor ?? undefined)

export const shouldClearConversationHistoryLoadMoreState = ({
  activeRequestRevision,
  responseRequestRevision,
}: {
  activeRequestRevision: number
  responseRequestRevision: number
}) => activeRequestRevision === responseRequestRevision

export const getConversationHistoryReconcileDelayMs = (attemptCount: number) => {
  if (attemptCount <= 0) {
    return 0
  }

  return 2 ** (attemptCount - 1) * 1000
}

export const shouldScheduleConversationHistoryReconcile = ({
  hasPendingRefresh,
  isRunning,
  loading,
}: {
  hasPendingRefresh: boolean
  isRunning: boolean
  loading: boolean
}) => hasPendingRefresh && !isRunning && !loading

export const resolveConversationHistoryReconcileDecision = ({
  attemptCount,
  refreshOutcome,
  persistedPendingThread,
  maxAttempts = MAX_CONVERSATION_HISTORY_RECONCILE_ATTEMPTS,
}: {
  attemptCount: number
  refreshOutcome: ConversationHistoryRefreshOutcome
  persistedPendingThread: boolean
  maxAttempts?: number
}): ConversationHistoryReconcileDecision => {
  if (refreshOutcome === 'applied' && persistedPendingThread) {
    return {
      shouldClearOptimistic: false,
      shouldClearPending: true,
    }
  }

  if (refreshOutcome === 'superseded') {
    return {
      shouldClearOptimistic: false,
      shouldClearPending: false,
      nextAttemptCount: attemptCount,
    }
  }

  if (attemptCount + 1 >= maxAttempts) {
    return {
      shouldClearOptimistic: true,
      shouldClearPending: true,
    }
  }

  return {
    shouldClearOptimistic: false,
    shouldClearPending: false,
    nextAttemptCount: attemptCount + 1,
  }
}

export const getNextConversationHistoryPendingRefreshState = ({
  current,
  targetThreadId,
  nextAttemptCount,
}: {
  current?: ConversationHistoryPendingRefreshState
  targetThreadId: string
  nextAttemptCount?: number
}): ConversationHistoryPendingRefreshState | undefined => {
  if (!current || current.threadId !== targetThreadId || nextAttemptCount === undefined) {
    return current
  }

  if (current.attemptCount === nextAttemptCount) {
    return current
  }

  return {
    threadId: current.threadId,
    attemptCount: nextAttemptCount,
  }
}

export const formatConversationHistoryTime = (
  timeValue: string,
  messages: ConversationHistoryTimeMessages,
  now = new Date(),
) => {
  const parsedTime = parseTimeValue(timeValue)
  if (!parsedTime) {
    return ''
  }

  const diffMs = Math.max(now.getTime() - parsedTime.getTime(), 0)
  const diffMinutes = Math.floor(diffMs / (60 * 1000))
  if (diffMinutes < 60) {
    return messages.minutesAgo(Math.max(diffMinutes, 1))
  }

  const dayDiff = getCalendarDayDiff(parsedTime, now)
  if (dayDiff === 0) {
    const diffHours = Math.floor(diffMs / (60 * 60 * 1000))
    return messages.hoursAgo(Math.max(diffHours, 1))
  }

  if (dayDiff === 1) {
    return messages.yesterday(formatClockTime(parsedTime))
  }

  if (dayDiff < 7) {
    return messages.daysAgo(Math.max(dayDiff, 1))
  }

  return formatAbsoluteDateTime(parsedTime)
}
