import { useCallback, useEffect, useRef, useState } from 'react'
import { checkpointFromPayload, historyFromPersistedEvents } from '../client/history.ts'
import { JsonRpcTimeoutError } from '../client/jsonRpcClient.ts'
import { createSerializedSessionWriter } from '../client/sessionApi.ts'
import type {
  HermesEvent,
  HermesHistorySnapshot,
  HermesSession,
  HermesSessionBookmark,
  HermesSessionStatus,
} from '../client/types.ts'
import { useHermesConnection } from './useHermesConnection.ts'
import { useHermesMessages } from './useHermesMessages.ts'
import { shouldRestoreInBackground } from './sessionPolicy.ts'

export function useHermesSession(requestedSessionId?: string) {
  const { sessionClient, sessionApi, connectionStatus, connectionError } = useHermesConnection()
  const {
    state: messageState,
    reset: resetMessages,
    hydrate: hydrateMessages,
    submitStarted,
    addUserMessage,
    checkpointResponded,
    applyEvent,
    sendFailed,
    stopRequested,
    stopAcknowledged,
    stopRetryAvailable,
    stopFailed,
    setStatus,
    clearError: clearMessageError,
  } = useHermesMessages()
  const [sessions, setSessions] = useState<HermesSessionBookmark[]>([])
  const [activeSession, setActiveSession] = useState<HermesSession | null>(null)
  const [selectedStoredSessionId, setSelectedStoredSessionId] = useState<string | null>(null)
  const [restoringSessionId, setRestoringSessionId] = useState<string | null>(null)
  const [loadingSessions, setLoadingSessions] = useState(true)
  const [openingSession, setOpeningSession] = useState(false)
  const [sessionError, setSessionError] = useState<string | null>(null)
  const activeSessionRef = useRef<HermesSession | null>(null)
  const selectedStoredSessionIdRef = useRef<string | null>(null)
  const selectedBookmarkRef = useRef<HermesSessionBookmark | null>(null)
  const restorePromiseRef = useRef<{
    storedSessionId: string
    promise: Promise<HermesSession>
  } | null>(null)
  const messageStateRef = useRef(messageState)
  const sessionsRef = useRef<HermesSessionBookmark[]>([])
  const writeBookmarkRef = useRef<ReturnType<typeof createSerializedSessionWriter> | null>(null)
  const openGeneration = useRef(0)
  const previousConnectionStatus = useRef(connectionStatus)
  const stopFeedbackTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  messageStateRef.current = messageState
  if (!writeBookmarkRef.current) {
    writeBookmarkRef.current = createSerializedSessionWriter(sessionApi)
  }

  const replaceSessions = useCallback((next: HermesSessionBookmark[]) => {
    const sorted = [...next].sort((left, right) => right.updatedAt.localeCompare(left.updatedAt))
    sessionsRef.current = sorted
    setSessions(sorted)
  }, [])

  const mergeSession = useCallback(
    (session: HermesSessionBookmark) => {
      replaceSessions([
        session,
        ...sessionsRef.current.filter((item) => item.storedSessionId !== session.storedSessionId),
      ])
    },
    [replaceSessions],
  )

  const refreshSessions = useCallback(async () => {
    setLoadingSessions(true)
    try {
      const next = await sessionApi.list()
      replaceSessions(next)
      setSessionError(null)
    } catch (reason) {
      setSessionError(errorMessage(reason, '无法加载 Hermes 历史会话。'))
    } finally {
      setLoadingSessions(false)
    }
  }, [replaceSessions, sessionApi])

  const persistBookmark = useCallback(
    async (bookmark: HermesSessionBookmark, required = false) => {
      if (selectedStoredSessionIdRef.current === bookmark.storedSessionId) {
        selectedBookmarkRef.current = bookmark
      }
      mergeSession(bookmark)
      try {
        const saved = await writeBookmarkRef.current!(bookmark)
        if (selectedStoredSessionIdRef.current === saved.storedSessionId) {
          selectedBookmarkRef.current = saved
        }
        mergeSession(saved)
        return saved
      } catch (reason) {
        const message = errorMessage(
          reason,
          required ? '无法保存新会话，消息尚未提交。' : '会话状态暂时无法同步。',
        )
        setSessionError(message)
        if (required) throw new Error(message)
        return bookmark
      }
    },
    [mergeSession],
  )

  const restoreSelectedSession = useCallback(
    async (
      bookmark: HermesSessionBookmark | null = selectedBookmarkRef.current,
      force = false,
    ): Promise<HermesSession> => {
      if (!bookmark) throw new Error('请先选择需要继续的历史会话。')
      if (!force && activeSessionRef.current?.storedSessionId === bookmark.storedSessionId) {
        return activeSessionRef.current
      }
      if (restorePromiseRef.current?.storedSessionId === bookmark.storedSessionId) {
        return restorePromiseRef.current.promise
      }

      setRestoringSessionId(bookmark.storedSessionId)
      const promise = sessionClient.restore(bookmark.storedSessionId, bookmark.liveSessionId)
      restorePromiseRef.current = { storedSessionId: bookmark.storedSessionId, promise }
      try {
        const restored = await promise
        if (selectedStoredSessionIdRef.current !== bookmark.storedSessionId) {
          throw new Error('会话已切换，本次连接已取消。')
        }
        activeSessionRef.current = restored
        setActiveSession(restored)
        const updatedBookmark = {
          ...bookmark,
          liveSessionId: restored.liveSessionId,
          title: restored.title || bookmark.title,
        }
        selectedBookmarkRef.current = updatedBookmark
        void persistBookmark(updatedBookmark)
        return restored
      } finally {
        if (restorePromiseRef.current?.promise === promise) restorePromiseRef.current = null
        if (selectedStoredSessionIdRef.current === bookmark.storedSessionId) {
          setRestoringSessionId(null)
        }
      }
    },
    [persistBookmark, sessionClient],
  )

  const openSession = useCallback(
    async (storedSessionId: string, force = false) => {
      if (!storedSessionId) return
      if (!force && selectedStoredSessionIdRef.current === storedSessionId) return
      const generation = ++openGeneration.current
      selectedStoredSessionIdRef.current = storedSessionId
      selectedBookmarkRef.current = null
      setSelectedStoredSessionId(storedSessionId)
      activeSessionRef.current = null
      setActiveSession(null)
      restorePromiseRef.current = null
      setRestoringSessionId(null)
      setOpeningSession(true)
      setSessionError(null)
      resetMessages()
      try {
        const knownBookmark = sessionsRef.current.find(
          (item) => item.storedSessionId === storedSessionId,
        )
        const [bookmark, persistedEvents] = await Promise.all([
          knownBookmark ? Promise.resolve(knownBookmark) : sessionApi.get(storedSessionId),
          sessionApi.history(storedSessionId).catch(() => []),
        ])
        if (generation !== openGeneration.current) return
        selectedBookmarkRef.current = bookmark

        let snapshot: HermesHistorySnapshot | null = null
        if (persistedEvents.length) snapshot = historyFromPersistedEvents(persistedEvents)
        if (!snapshot) {
          const restored = await restoreSelectedSession(bookmark)
          snapshot = await sessionClient.historySnapshot(restored.liveSessionId)
        }
        if (generation !== openGeneration.current) return

        const hydratedSnapshot = {
          ...snapshot,
          checkpoint: snapshot.checkpoint || bookmark.checkpoint,
          activeRunId: snapshot.activeRunId || bookmark.activeRunId,
        }
        const running = bookmark.status === 'running'
        hydrateMessages(
          hydratedSnapshot,
          running,
          running
            ? '历史记录已加载，正在连接运行会话…'
            : bookmark.status === 'waiting_confirmation'
              ? '等待用户回复'
              : '历史记录已加载',
        )
        setOpeningSession(false)

        if (persistedEvents.length && shouldRestoreInBackground(bookmark.status)) {
          void restoreSelectedSession(bookmark)
            .then(() => {
              if (selectedStoredSessionIdRef.current !== storedSessionId) return
              setStatus(
                bookmark.status === 'waiting_confirmation' ? '等待用户回复' : '任务仍在运行',
                bookmark.status === 'running',
              )
            })
            .catch((reason) => {
              if (selectedStoredSessionIdRef.current !== storedSessionId) return
              setSessionError(errorMessage(reason, '历史记录已显示，但实时会话暂时无法连接。'))
              setStatus('历史记录已加载，发送消息时将重试连接', running)
            })
        }
      } catch (reason) {
        if (generation !== openGeneration.current) return
        resetMessages()
        setSessionError(errorMessage(reason, '无法打开该会话，请稍后重试。'))
      } finally {
        if (generation === openGeneration.current) setOpeningSession(false)
      }
    },
    [hydrateMessages, resetMessages, restoreSelectedSession, sessionApi, sessionClient, setStatus],
  )

  const newConversation = useCallback(() => {
    openGeneration.current += 1
    selectedStoredSessionIdRef.current = null
    selectedBookmarkRef.current = null
    setSelectedStoredSessionId(null)
    activeSessionRef.current = null
    setActiveSession(null)
    restorePromiseRef.current = null
    setRestoringSessionId(null)
    setOpeningSession(false)
    setSessionError(null)
    resetMessages()
  }, [resetMessages])

  const ensureSession = useCallback(
    async (firstPrompt: string): Promise<HermesSession> => {
      const selectedBookmark = selectedBookmarkRef.current
      if (
        activeSessionRef.current &&
        (!selectedBookmark ||
          activeSessionRef.current.storedSessionId === selectedBookmark.storedSessionId)
      ) {
        return activeSessionRef.current
      }
      if (selectedBookmark) {
        setStatus('正在连接会话', messageStateRef.current.busy)
        return restoreSelectedSession(selectedBookmark)
      }
      const created = await sessionClient.create()
      activeSessionRef.current = created
      setActiveSession(created)
      const now = new Date().toISOString()
      const bookmark: HermesSessionBookmark = {
        storedSessionId: created.storedSessionId,
        liveSessionId: created.liveSessionId,
        title: created.title || promptTitle(firstPrompt),
        createdAt: now,
        updatedAt: now,
        status: 'ready',
        activeRunId: null,
        checkpoint: null,
      }
      selectedStoredSessionIdRef.current = created.storedSessionId
      selectedBookmarkRef.current = bookmark
      setSelectedStoredSessionId(created.storedSessionId)
      try {
        await persistBookmark(bookmark, true)
      } catch (reason) {
        if (activeSessionRef.current?.storedSessionId === created.storedSessionId) {
          activeSessionRef.current = null
          setActiveSession(null)
        }
        if (selectedStoredSessionIdRef.current === created.storedSessionId) {
          selectedStoredSessionIdRef.current = null
          selectedBookmarkRef.current = null
          setSelectedStoredSessionId(null)
        }
        replaceSessions(
          sessionsRef.current.filter((item) => item.storedSessionId !== created.storedSessionId),
        )
        throw reason
      }
      return created
    },
    [persistBookmark, replaceSessions, restoreSelectedSession, sessionClient, setStatus],
  )

  const updateActiveBookmark = useCallback(
    async (
      patch: Partial<HermesSessionBookmark>,
      session: HermesSession | null = activeSessionRef.current,
    ) => {
      if (!session) return null
      const current = sessionsRef.current.find(
        (item) => item.storedSessionId === session.storedSessionId,
      )
      const now = new Date().toISOString()
      const bookmark: HermesSessionBookmark = {
        storedSessionId: session.storedSessionId,
        liveSessionId: session.liveSessionId,
        title: session.title || current?.title || '新对话',
        createdAt: current?.createdAt || now,
        updatedAt: now,
        status: current?.status || 'ready',
        activeRunId: current?.activeRunId || null,
        checkpoint: current?.checkpoint || null,
        ...patch,
      }
      return persistBookmark(bookmark)
    },
    [persistBookmark],
  )

  const sendMessage = useCallback(
    async (rawText: string, workflowId?: string): Promise<string | undefined> => {
      const text = rawText.trim()
      if (!text) return selectedStoredSessionIdRef.current || undefined
      if (messageState.busy) {
        addUserMessage(text, messageState.activeRunId)
        try {
          const session = await ensureSession(text)
          await sessionClient.steer(session.liveSessionId, text)
          return session.storedSessionId
        } catch (reason) {
          setSessionError(errorMessage(reason, '无法补充当前任务要求。'))
          return undefined
        }
      }

      let session: HermesSession | null = null
      const runId = sessionClient.createRunId()
      submitStarted(text, runId)
      try {
        session = await ensureSession(text)
        void updateActiveBookmark(
          {
            status: 'running',
            activeRunId: runId,
            checkpoint: null,
          },
          session,
        )
        try {
          await sessionClient.submit(session.liveSessionId, text, runId, workflowId)
        } catch (reason) {
          const currentState = messageStateRef.current
          if (currentState.activeRunId !== runId || !currentState.busy) {
            return session.storedSessionId
          }
          if (reason instanceof JsonRpcTimeoutError) {
            setSessionError('任务已提交，但服务未及时确认，正在继续等待运行结果。')
            setStatus('任务已提交，正在等待运行事件', true)
            return session.storedSessionId
          }
          throw reason
        }
        return session.storedSessionId
      } catch (reason) {
        sendFailed(errorMessage(reason, '消息发送失败。'))
        if (session) {
          void updateActiveBookmark(
            { status: 'failed', activeRunId: null, checkpoint: null },
            session,
          )
        }
        return undefined
      }
    },
    [
      addUserMessage,
      ensureSession,
      messageState.activeRunId,
      messageState.busy,
      sendFailed,
      sessionClient,
      setStatus,
      submitStarted,
      updateActiveBookmark,
    ],
  )

  const interrupt = useCallback(async () => {
    if (!messageState.busy) return
    if (messageState.stopping && !messageState.stopRetryAllowed) return
    if (stopFeedbackTimer.current) clearTimeout(stopFeedbackTimer.current)
    stopRequested()
    stopFeedbackTimer.current = setTimeout(stopRetryAvailable, 3_000)
    try {
      const session = activeSessionRef.current || (await ensureSession(''))
      await sessionClient.interrupt(session.liveSessionId)
      stopAcknowledged()
    } catch (reason) {
      if (stopFeedbackTimer.current) {
        clearTimeout(stopFeedbackTimer.current)
        stopFeedbackTimer.current = null
      }
      stopFailed(errorMessage(reason, '无法停止当前任务。'))
    }
  }, [
    ensureSession,
    messageState.busy,
    messageState.stopRetryAllowed,
    messageState.stopping,
    sessionClient,
    stopAcknowledged,
    stopFailed,
    stopRequested,
    stopRetryAvailable,
  ])

  const respondCheckpoint = useCallback(
    async (answer: string) => {
      const checkpoint = messageState.checkpoint
      if (!checkpoint || !answer.trim()) return
      const visibleAnswer =
        checkpoint.choices.find((choice) => choice.value === answer)?.label || answer
      try {
        const session = await ensureSession(visibleAnswer)
        if (checkpoint.type === 'approval') {
          await sessionClient.respondApproval(session.liveSessionId, checkpoint.requestId, answer)
        } else {
          await sessionClient.respondClarification(checkpoint.requestId, answer)
        }
        checkpointResponded(visibleAnswer)
        await updateActiveBookmark({
          status: 'running',
          checkpoint: null,
          activeRunId: messageState.activeRunId,
        })
      } catch (reason) {
        sendFailed(errorMessage(reason, '无法提交确认结果。'))
      }
    },
    [
      checkpointResponded,
      ensureSession,
      messageState.activeRunId,
      messageState.checkpoint,
      sendFailed,
      sessionClient,
      updateActiveBookmark,
    ],
  )

  const renameSession = useCallback(
    async (storedSessionId: string, title: string) => {
      const current = sessionsRef.current.find((item) => item.storedSessionId === storedSessionId)
      if (!current || !title.trim()) return
      await persistBookmark({
        ...current,
        title: title.trim(),
        updatedAt: new Date().toISOString(),
      })
    },
    [persistBookmark],
  )

  const deleteSession = useCallback(
    async (storedSessionId: string) => {
      await sessionApi.remove(storedSessionId)
      replaceSessions(
        sessionsRef.current.filter((item) => item.storedSessionId !== storedSessionId),
      )
      if (selectedStoredSessionIdRef.current === storedSessionId) newConversation()
    },
    [newConversation, replaceSessions, sessionApi],
  )

  useEffect(() => {
    void refreshSessions()
  }, [refreshSessions])

  useEffect(() => {
    if (
      requestedSessionId &&
      requestedSessionId !== selectedStoredSessionIdRef.current &&
      !openingSession
    ) {
      void openSession(requestedSessionId)
    }
  }, [openSession, openingSession, requestedSessionId])

  useEffect(() => {
    const unsubscribe = sessionClient.subscribe((event) => {
      const session = activeSessionRef.current
      if (!session || (event.sessionId && event.sessionId !== session.liveSessionId)) return
      applyEvent(event)
      void persistEventState(event, session, updateActiveBookmark)
      if (terminalStatus(event) && stopFeedbackTimer.current) {
        clearTimeout(stopFeedbackTimer.current)
        stopFeedbackTimer.current = null
      }
      if (event.type === 'session.title') {
        const title = String(event.payload.title || '').trim()
        if (title) {
          const updated = { ...session, title }
          activeSessionRef.current = updated
          setActiveSession(updated)
          void updateActiveBookmark({ title })
        }
      }
    })
    return unsubscribe
  }, [applyEvent, sessionClient, updateActiveBookmark])

  useEffect(() => {
    const previous = previousConnectionStatus.current
    previousConnectionStatus.current = connectionStatus
    if (
      connectionStatus === 'connected' &&
      ['disconnected', 'error', 'reconnecting'].includes(previous) &&
      selectedBookmarkRef.current &&
      (activeSessionRef.current || shouldRestoreInBackground(selectedBookmarkRef.current.status))
    ) {
      const bookmark = selectedBookmarkRef.current
      void restoreSelectedSession(bookmark, true).catch((reason) => {
        if (selectedStoredSessionIdRef.current !== bookmark.storedSessionId) return
        setSessionError(errorMessage(reason, '实时会话重新连接失败。'))
      })
    }
  }, [connectionStatus, restoreSelectedSession])

  useEffect(
    () => () => {
      if (stopFeedbackTimer.current) clearTimeout(stopFeedbackTimer.current)
    },
    [],
  )

  return {
    sessions,
    activeSession,
    selectedStoredSessionId,
    restoringSession: Boolean(restoringSessionId && restoringSessionId === selectedStoredSessionId),
    loadingSessions,
    openingSession,
    connectionStatus,
    state: messageState,
    error: messageState.error || sessionError || connectionError,
    refreshSessions,
    openSession,
    newConversation,
    sendMessage,
    interrupt,
    respondCheckpoint,
    renameSession,
    deleteSession,
    clearError: () => {
      setSessionError(null)
      clearMessageError()
    },
  }
}

async function persistEventState(
  event: HermesEvent,
  session: HermesSession,
  update: (
    patch: Partial<HermesSessionBookmark>,
    session?: HermesSession | null,
  ) => Promise<HermesSessionBookmark | null>,
) {
  if (event.type === 'approval.request' || event.type === 'clarify.request') {
    await update(
      {
        status: 'waiting_confirmation',
        checkpoint: checkpointFromPayload(event.type, event.payload),
      },
      session,
    )
    return
  }
  const status = terminalStatus(event)
  if (status) await update({ status, activeRunId: null, checkpoint: null }, session)
}

function terminalStatus(event: HermesEvent): HermesSessionStatus | null {
  if (event.type === 'run.failed' || event.type === 'error') return 'failed'
  if (
    event.type === 'session.interrupted' ||
    (event.type === 'message.complete' && event.payload.status === 'interrupted')
  ) {
    return 'interrupted'
  }
  if (event.type === 'message.complete' || event.type === 'run.completed') return 'ready'
  return null
}

function promptTitle(prompt: string): string {
  const normalized = prompt.replace(/\s+/g, ' ').trim()
  return normalized.length > 24 ? `${normalized.slice(0, 24)}…` : normalized || '新对话'
}

function errorMessage(reason: unknown, fallback: string): string {
  return reason instanceof Error && reason.message ? reason.message : fallback
}
