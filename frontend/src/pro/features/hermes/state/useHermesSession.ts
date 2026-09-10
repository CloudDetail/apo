import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import {
  isInterruptedTerminalEvent,
  isRunScopedEventType,
  runIdFromEventPayload,
} from '../client/eventPayload.ts'
import { checkpointFromPayload, historyFromPersistedEvents } from '../client/history.ts'
import { JsonRpcTimeoutError } from '../client/jsonRpcClient.ts'
import { createSerializedSessionWriter } from '../client/sessionApi.ts'
import type { HermesSessionClient } from '../client/sessionClient.ts'
import {
  HermesUserScopeError,
  HermesUserScopeGeneration,
  type HermesUserScopeSnapshot,
} from '../client/userScope.ts'
import type {
  HermesEvent,
  HermesHistorySnapshot,
  HermesSession,
  HermesSessionBookmark,
  HermesSessionStatus,
} from '../client/types.ts'
import { useHermesConnection } from './useHermesConnection.ts'
import { useHermesMessages } from './useHermesMessages.ts'
import { shouldIgnoreEventAfterInterrupt } from './messageReducer.ts'
import {
  canBindNewSessionOperation,
  canApplyInterruptReconciliation,
  currentSessionCreation,
  hydrationRuntimeFromHistory,
  isCurrentSessionOperation,
  reconcileInterruptedBookmark,
  runStartBarrierDecision,
  shouldIgnoreKnownInterruptedRunEvent,
  shouldRestoreInBackground,
} from './sessionPolicy.ts'

export function useHermesSession(requestedSessionId?: string) {
  const {
    sessionClient,
    sessionApi,
    userScopeKey,
    userScopeReady,
    connectionStatus,
    connectionError,
  } = useHermesConnection()
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
    scope: HermesUserScopeSnapshot<HermesSessionClient>
  } | null>(null)
  const createPromiseRef = useRef<{
    generation: number
    promise: Promise<HermesSession>
    scope: HermesUserScopeSnapshot<HermesSessionClient>
  } | null>(null)
  const messageStateRef = useRef(messageState)
  const sessionsRef = useRef<HermesSessionBookmark[]>([])
  const interruptedRunIdsRef = useRef<Set<string>>(new Set())
  const interruptedSessionAwaitingNextRunRef = useRef<string | null>(null)
  const runStartBarrierRef = useRef<{ storedSessionId: string; runId: string } | null>(null)
  const scopeGuardRef = useRef<HermesUserScopeGeneration | null>(null)
  const previousUserScopeKeyRef = useRef(userScopeKey)
  const sessionApiIdentityRef = useRef(sessionApi)
  const sessionClientIdentityRef = useRef(sessionClient)
  const writeBookmarkRef = useRef<{
    api: typeof sessionApi
    write: ReturnType<typeof createSerializedSessionWriter>
  } | null>(null)
  const openGeneration = useRef(0)
  const previousConnectionStatus = useRef(connectionStatus)
  const stopFeedbackTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  if (previousUserScopeKeyRef.current !== userScopeKey) {
    previousUserScopeKeyRef.current = userScopeKey
    activeSessionRef.current = null
    selectedStoredSessionIdRef.current = null
    selectedBookmarkRef.current = null
    restorePromiseRef.current = null
    createPromiseRef.current = null
    sessionsRef.current = []
    interruptedRunIdsRef.current.clear()
    interruptedSessionAwaitingNextRunRef.current = null
    runStartBarrierRef.current = null
    openGeneration.current += 1
  }
  if (!scopeGuardRef.current) scopeGuardRef.current = new HermesUserScopeGeneration()
  const scopeGuard = scopeGuardRef.current
  scopeGuard.update(userScopeKey)
  sessionApiIdentityRef.current = sessionApi
  sessionClientIdentityRef.current = sessionClient
  messageStateRef.current = messageState
  if (writeBookmarkRef.current?.api !== sessionApi) {
    writeBookmarkRef.current = {
      api: sessionApi,
      write: createSerializedSessionWriter(sessionApi),
    }
  }
  const bookmarkWriter = writeBookmarkRef.current

  useLayoutEffect(() => {
    scopeGuard.update(userScopeKey)
    return () => {
      scopeGuard.update(null)
      if (stopFeedbackTimer.current) {
        clearTimeout(stopFeedbackTimer.current)
        stopFeedbackTimer.current = null
      }
    }
  }, [scopeGuard, userScopeKey])

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
    let scope: HermesUserScopeSnapshot<typeof sessionApi>
    try {
      scope = captureCurrentScope(scopeGuard, sessionApi, sessionApiIdentityRef.current)
    } catch (reason) {
      if (!(reason instanceof HermesUserScopeChangedError)) {
        setSessionError(errorMessage(reason, '无法加载 Agent 历史会话。'))
        setLoadingSessions(false)
      }
      return
    }
    setLoadingSessions(true)
    try {
      const next = await sessionApi.list()
      if (!scopeGuard.isCurrent(scope, sessionApiIdentityRef.current)) return
      replaceSessions(next)
      setSessionError(null)
    } catch (reason) {
      if (!scopeGuard.isCurrent(scope, sessionApiIdentityRef.current)) return
      setSessionError(errorMessage(reason, '无法加载 Agent 历史会话。'))
    } finally {
      if (scopeGuard.isCurrent(scope, sessionApiIdentityRef.current)) {
        setLoadingSessions(false)
      }
    }
  }, [replaceSessions, scopeGuard, sessionApi])

  const persistBookmark = useCallback(
    async (bookmark: HermesSessionBookmark, required = false) => {
      let scope: HermesUserScopeSnapshot<typeof sessionApi>
      try {
        scope = captureCurrentScope(
          scopeGuard,
          bookmarkWriter.api,
          sessionApiIdentityRef.current,
        )
      } catch (reason) {
        if (required) throw reason
        return bookmark
      }
      if (selectedStoredSessionIdRef.current === bookmark.storedSessionId) {
        selectedBookmarkRef.current = bookmark
      }
      mergeSession(bookmark)
      try {
        const saved = await bookmarkWriter.write(bookmark)
        if (!scopeGuard.isCurrent(scope, sessionApiIdentityRef.current)) {
          if (required) throw new HermesUserScopeChangedError()
          return bookmark
        }
        if (selectedStoredSessionIdRef.current === saved.storedSessionId) {
          selectedBookmarkRef.current = saved
        }
        mergeSession(saved)
        return saved
      } catch (reason) {
        if (!scopeGuard.isCurrent(scope, sessionApiIdentityRef.current)) {
          if (required) throw new HermesUserScopeChangedError()
          return bookmark
        }
        const message = errorMessage(
          reason,
          required ? '无法保存新会话，消息尚未提交。' : '会话状态暂时无法同步。',
        )
        setSessionError(message)
        if (required) throw new Error(message)
        return bookmark
      }
    },
    [bookmarkWriter, mergeSession, scopeGuard],
  )

  const restoreSelectedSession = useCallback(
    async (
      bookmark: HermesSessionBookmark | null = selectedBookmarkRef.current,
      force = false,
    ): Promise<HermesSession> => {
      if (!sessionClient) throw new HermesUserScopeError()
      const scope = captureCurrentScope(
        scopeGuard,
        sessionClient,
        sessionClientIdentityRef.current,
      )
      if (!bookmark) throw new Error('请先选择需要继续的历史会话。')
      if (!force && activeSessionRef.current?.storedSessionId === bookmark.storedSessionId) {
        return activeSessionRef.current
      }
      if (
        restorePromiseRef.current?.storedSessionId === bookmark.storedSessionId &&
        scopeGuard.isCurrent(
          restorePromiseRef.current.scope,
          sessionClientIdentityRef.current,
        )
      ) {
        return restorePromiseRef.current.promise
      }

      setRestoringSessionId(bookmark.storedSessionId)
      const promise = sessionClient.restore(bookmark.storedSessionId, bookmark.liveSessionId)
      restorePromiseRef.current = { storedSessionId: bookmark.storedSessionId, promise, scope }
      try {
        const restored = await promise
        if (
          !scopeGuard.isCurrent(scope, sessionClientIdentityRef.current) ||
          selectedStoredSessionIdRef.current !== bookmark.storedSessionId
        ) {
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
        if (
          scopeGuard.isCurrent(scope, sessionClientIdentityRef.current) &&
          selectedStoredSessionIdRef.current === bookmark.storedSessionId
        ) {
          setRestoringSessionId(null)
        }
      }
    },
    [persistBookmark, scopeGuard, sessionClient],
  )

  const openSession = useCallback(
    async (storedSessionId: string, force = false) => {
      if (!storedSessionId) return
      if (!sessionClient) {
        setSessionError(new HermesUserScopeError().message)
        return
      }
      let scope: HermesUserScopeSnapshot<typeof sessionApi>
      try {
        scope = captureCurrentScope(scopeGuard, sessionApi, sessionApiIdentityRef.current)
        if (sessionClientIdentityRef.current !== sessionClient) {
          throw new HermesUserScopeChangedError()
        }
      } catch (reason) {
        if (!(reason instanceof HermesUserScopeChangedError)) {
          setSessionError(errorMessage(reason, '无法打开该会话，请稍后重试。'))
        }
        return
      }
      if (!force && selectedStoredSessionIdRef.current === storedSessionId) return
      const generation = ++openGeneration.current
      createPromiseRef.current = null
      interruptedRunIdsRef.current.clear()
      interruptedSessionAwaitingNextRunRef.current = null
      runStartBarrierRef.current = null
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
        if (
          generation !== openGeneration.current ||
          !scopeGuard.isCurrent(scope, sessionApiIdentityRef.current) ||
          sessionClientIdentityRef.current !== sessionClient
        ) return
        selectedBookmarkRef.current = bookmark

        let snapshot: HermesHistorySnapshot | null = null
        if (persistedEvents.length) snapshot = historyFromPersistedEvents(persistedEvents)
        if (!snapshot) {
          const restored = await restoreSelectedSession(bookmark)
          snapshot = await sessionClient.historySnapshot(restored.liveSessionId)
        }
        if (
          generation !== openGeneration.current ||
          !scopeGuard.isCurrent(scope, sessionApiIdentityRef.current) ||
          sessionClientIdentityRef.current !== sessionClient
        ) return

        const runtime = hydrationRuntimeFromHistory(bookmark, snapshot)
        const hydratedBookmark: HermesSessionBookmark = {
          ...bookmark,
          status: runtime.status,
          activeRunId: runtime.activeRunId,
          checkpoint: runtime.checkpoint,
        }
        selectedBookmarkRef.current = hydratedBookmark
        mergeSession(hydratedBookmark)
        const hydratedSnapshot = {
          ...snapshot,
          checkpoint: runtime.checkpoint,
          activeRunId: runtime.activeRunId,
        }
        for (const item of hydratedSnapshot.timeline) {
          if (item.kind === 'run_status' && item.status === 'interrupted' && item.runId) {
            interruptedRunIdsRef.current.add(item.runId)
          }
        }
        const latestRunStatus = [...hydratedSnapshot.timeline]
          .reverse()
          .find((item) => item.kind === 'run_status')
        interruptedSessionAwaitingNextRunRef.current =
          latestRunStatus?.status === 'interrupted' ? bookmark.storedSessionId : null
        const running = runtime.status === 'running'
        hydrateMessages(
          hydratedSnapshot,
          running,
          running
            ? '历史记录已加载，正在连接运行会话…'
            : runtime.status === 'waiting_confirmation'
              ? '等待用户回复'
              : '历史记录已加载',
        )
        setOpeningSession(false)

        if (persistedEvents.length && runtime.restoreInBackground) {
          void restoreSelectedSession(hydratedBookmark)
            .then(() => {
              if (
                !scopeGuard.isCurrent(scope, sessionApiIdentityRef.current) ||
                sessionClientIdentityRef.current !== sessionClient ||
                selectedStoredSessionIdRef.current !== storedSessionId
              ) return
              setStatus(
                runtime.status === 'waiting_confirmation' ? '等待用户回复' : '任务仍在运行',
                runtime.status === 'running',
              )
            })
            .catch((reason) => {
              if (
                !scopeGuard.isCurrent(scope, sessionApiIdentityRef.current) ||
                sessionClientIdentityRef.current !== sessionClient ||
                selectedStoredSessionIdRef.current !== storedSessionId
              ) return
              setSessionError(errorMessage(reason, '历史记录已显示，但实时会话暂时无法连接。'))
              setStatus('历史记录已加载，发送消息时将重试连接', running)
            })
        }
      } catch (reason) {
        if (
          generation !== openGeneration.current ||
          !scopeGuard.isCurrent(scope, sessionApiIdentityRef.current) ||
          sessionClientIdentityRef.current !== sessionClient
        ) return
        resetMessages()
        setSessionError(errorMessage(reason, '无法打开该会话，请稍后重试。'))
      } finally {
        if (
          generation === openGeneration.current &&
          scopeGuard.isCurrent(scope, sessionApiIdentityRef.current) &&
          sessionClientIdentityRef.current === sessionClient
        ) setOpeningSession(false)
      }
    },
    [
      hydrateMessages,
      mergeSession,
      resetMessages,
      restoreSelectedSession,
      scopeGuard,
      sessionApi,
      sessionClient,
      setStatus,
    ],
  )

  const newConversation = useCallback(() => {
    openGeneration.current += 1
    createPromiseRef.current = null
    interruptedRunIdsRef.current.clear()
    interruptedSessionAwaitingNextRunRef.current = null
    runStartBarrierRef.current = null
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
      if (!sessionClient) throw new HermesUserScopeError()
      const scope = captureCurrentScope(
        scopeGuard,
        sessionClient,
        sessionClientIdentityRef.current,
      )
      const generation = openGeneration.current
      const sharedCreation = currentSessionCreation(
        createPromiseRef.current,
        generation,
        (pendingScope) => scopeGuard.isCurrent(pendingScope, sessionClient),
      )
      if (sharedCreation) return sharedCreation
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
      if (
        !canBindNewSessionOperation(
          generation,
          openGeneration.current,
          selectedStoredSessionIdRef.current,
        )
      ) {
        throw new HermesSessionOperationChangedError()
      }

      const promise = (async () => {
        const created = await sessionClient.create()
        if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) {
          throw new HermesUserScopeChangedError()
        }
        if (
          !canBindNewSessionOperation(
            generation,
            openGeneration.current,
            selectedStoredSessionIdRef.current,
          )
        ) {
          throw new HermesSessionOperationChangedError()
        }
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
          if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) {
            throw new HermesUserScopeChangedError()
          }
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
          if (generation !== openGeneration.current) {
            throw new HermesSessionOperationChangedError()
          }
          throw reason
        }
        if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) {
          throw new HermesUserScopeChangedError()
        }
        if (
          generation !== openGeneration.current ||
          selectedStoredSessionIdRef.current !== created.storedSessionId
        ) {
          throw new HermesSessionOperationChangedError()
        }
        return created
      })()
      createPromiseRef.current = { generation, promise, scope }
      try {
        return await promise
      } finally {
        if (createPromiseRef.current?.promise === promise) createPromiseRef.current = null
      }
    },
    [
      persistBookmark,
      replaceSessions,
      restoreSelectedSession,
      scopeGuard,
      sessionClient,
      setStatus,
    ],
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
      if (!userScopeReady || !sessionClient) {
        setSessionError(new HermesUserScopeError().message)
        return undefined
      }
      let scope: HermesUserScopeSnapshot<HermesSessionClient>
      try {
        scope = captureCurrentScope(
          scopeGuard,
          sessionClient,
          sessionClientIdentityRef.current,
        )
      } catch (reason) {
        if (!(reason instanceof HermesUserScopeChangedError)) {
          setSessionError(errorMessage(reason, '消息发送失败。'))
        }
        return undefined
      }
      if (messageState.busy) {
        addUserMessage(text, messageState.activeRunId)
        try {
          const session = await ensureSession(text)
          if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return undefined
          await sessionClient.steer(session.liveSessionId, text)
          if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return undefined
          return session.storedSessionId
        } catch (reason) {
          if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return undefined
          if (reason instanceof HermesSessionOperationChangedError) return undefined
          setSessionError(errorMessage(reason, '无法补充当前任务要求。'))
          return undefined
        }
      }

      let session: HermesSession | null = null
      const runId = sessionClient.createRunId()
      const selectedSessionId = selectedStoredSessionIdRef.current
      if (
        selectedSessionId &&
        interruptedSessionAwaitingNextRunRef.current === selectedSessionId
      ) {
        runStartBarrierRef.current = { storedSessionId: selectedSessionId, runId }
        interruptedSessionAwaitingNextRunRef.current = null
      }
      submitStarted(text, runId)
      try {
        session = await ensureSession(text)
        if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return undefined
        if (interruptedSessionAwaitingNextRunRef.current === session.storedSessionId) {
          runStartBarrierRef.current = { storedSessionId: session.storedSessionId, runId }
          interruptedSessionAwaitingNextRunRef.current = null
        }
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
          if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return undefined
        } catch (reason) {
          if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return undefined
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
        if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return undefined
        if (runStartBarrierRef.current?.runId === runId) {
          interruptedSessionAwaitingNextRunRef.current =
            session?.storedSessionId || runStartBarrierRef.current.storedSessionId
          runStartBarrierRef.current = null
        }
        if (reason instanceof HermesSessionOperationChangedError) return undefined
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
      scopeGuard,
      sendFailed,
      sessionClient,
      setStatus,
      submitStarted,
      updateActiveBookmark,
      userScopeReady,
    ],
  )

  const interrupt = useCallback(async () => {
    if (!messageState.busy) return
    if (messageState.stopping && !messageState.stopRetryAllowed) return
    if (!sessionClient) {
      setSessionError(new HermesUserScopeError().message)
      return
    }
    let scope: HermesUserScopeSnapshot<HermesSessionClient>
    try {
      scope = captureCurrentScope(
        scopeGuard,
        sessionClient,
        sessionClientIdentityRef.current,
      )
    } catch (reason) {
      if (!(reason instanceof HermesUserScopeChangedError)) {
        setSessionError(errorMessage(reason, '无法停止当前任务。'))
      }
      return
    }
    const runId = messageState.activeRunId || selectedBookmarkRef.current?.activeRunId || null
    const operation = {
      generation: openGeneration.current,
      storedSessionId: selectedStoredSessionIdRef.current,
    }
    const operationIsCurrent = (storedSessionId = operation.storedSessionId) =>
      scopeGuard.isCurrent(scope, sessionClientIdentityRef.current) &&
      isCurrentSessionOperation(
        { ...operation, storedSessionId },
        openGeneration.current,
        selectedStoredSessionIdRef.current,
      )
    if (stopFeedbackTimer.current) {
      clearTimeout(stopFeedbackTimer.current)
      stopFeedbackTimer.current = null
    }
    stopRequested()
    const feedbackTimer = setTimeout(() => {
      if (stopFeedbackTimer.current !== feedbackTimer) return
      stopFeedbackTimer.current = null
      if (operationIsCurrent()) stopRetryAvailable()
    }, 3_000)
    stopFeedbackTimer.current = feedbackTimer
    const clearFeedbackTimer = () => {
      if (stopFeedbackTimer.current !== feedbackTimer) return
      clearTimeout(feedbackTimer)
      stopFeedbackTimer.current = null
    }
    let session: HermesSession | null = null
    try {
      session = await ensureSession('')
      if (!operationIsCurrent(session.storedSessionId)) {
        clearFeedbackTimer()
        return
      }
      await sessionClient.interrupt(session.liveSessionId)
      if (!operationIsCurrent(session.storedSessionId)) {
        clearFeedbackTimer()
        return
      }
      const currentState = messageStateRef.current
      if (!currentState.busy || (runId && currentState.activeRunId !== runId)) {
        clearFeedbackTimer()
        return
      }
      stopAcknowledged()
      const interruptedRunId = runId || currentState.activeRunId
      if (interruptedRunId) interruptedRunIdsRef.current.add(interruptedRunId)
      interruptedSessionAwaitingNextRunRef.current = session.storedSessionId
      const currentBookmark =
        sessionsRef.current.find((item) => item.storedSessionId === session.storedSessionId) ||
        (selectedBookmarkRef.current?.storedSessionId === session.storedSessionId
          ? selectedBookmarkRef.current
          : null)
      const now = new Date().toISOString()
      const optimisticBookmark: HermesSessionBookmark = {
        storedSessionId: session.storedSessionId,
        liveSessionId: session.liveSessionId,
        title: session.title || currentBookmark?.title || '新对话',
        createdAt: currentBookmark?.createdAt || now,
        updatedAt: now,
        status: 'interrupted',
        activeRunId: interruptedRunId,
        checkpoint: null,
      }
      selectedBookmarkRef.current = optimisticBookmark
      mergeSession(optimisticBookmark)
      clearFeedbackTimer()
      applyEvent(interruptedEvent(session.liveSessionId, interruptedRunId))

      const reconciledBookmark = await reconcileInterruptedBookmark(() =>
        sessionApi.interrupt(session.storedSessionId, interruptedRunId),
      )
      if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return
      const latestState = messageStateRef.current
      const latestBookmark = selectedBookmarkRef.current
      const selectedSessionIsCurrent =
        selectedStoredSessionIdRef.current === session.storedSessionId
      const reconciliationIsCurrent =
        operationIsCurrent(session.storedSessionId) &&
        canApplyInterruptReconciliation(
          interruptedRunId,
          latestState.busy,
          latestState.activeRunId,
          latestBookmark?.status || null,
          latestBookmark?.activeRunId || null,
        )
      if (reconciledBookmark) {
        if (selectedSessionIsCurrent && !reconciliationIsCurrent) return
        if (selectedSessionIsCurrent) {
          selectedBookmarkRef.current = reconciledBookmark
        }
        mergeSession(reconciledBookmark)
      } else if (reconciliationIsCurrent) {
        setSessionError('任务已停止，但会话列表状态暂时未同步，请稍后刷新。')
      }
    } catch (reason) {
      if (!operationIsCurrent(session ? session.storedSessionId : operation.storedSessionId)) {
        clearFeedbackTimer()
        return
      }
      const latestState = messageStateRef.current
      if (!latestState.busy || (runId && latestState.activeRunId !== runId)) {
        clearFeedbackTimer()
        return
      }
      clearFeedbackTimer()
      stopFailed(errorMessage(reason, '无法停止当前任务。'), runId)
    }
  }, [
    applyEvent,
    ensureSession,
    mergeSession,
    messageState.activeRunId,
    messageState.busy,
    messageState.stopRetryAllowed,
    messageState.stopping,
    scopeGuard,
    sessionApi,
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
      if (!sessionClient) {
        setSessionError(new HermesUserScopeError().message)
        return
      }
      let scope: HermesUserScopeSnapshot<HermesSessionClient>
      try {
        scope = captureCurrentScope(
          scopeGuard,
          sessionClient,
          sessionClientIdentityRef.current,
        )
      } catch (reason) {
        if (!(reason instanceof HermesUserScopeChangedError)) {
          setSessionError(errorMessage(reason, '无法提交确认结果。'))
        }
        return
      }
      const visibleAnswer =
        checkpoint.choices.find((choice) => choice.value === answer)?.label || answer
      try {
        const session = await ensureSession(visibleAnswer)
        if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return
        if (checkpoint.type === 'approval') {
          await sessionClient.respondApproval(session.liveSessionId, checkpoint.requestId, answer)
        } else {
          await sessionClient.respondClarification(checkpoint.requestId, answer, checkpoint.runId)
        }
        if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return
        checkpointResponded(visibleAnswer)
        await updateActiveBookmark({
          status: 'running',
          checkpoint: null,
          activeRunId: messageState.activeRunId,
        })
      } catch (reason) {
        if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return
        if (reason instanceof HermesSessionOperationChangedError) return
        sendFailed(errorMessage(reason, '无法提交确认结果。'))
      }
    },
    [
      checkpointResponded,
      ensureSession,
      messageState.activeRunId,
      messageState.checkpoint,
      scopeGuard,
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
      let scope: HermesUserScopeSnapshot<typeof sessionApi>
      try {
        scope = captureCurrentScope(scopeGuard, sessionApi, sessionApiIdentityRef.current)
      } catch (reason) {
        if (!(reason instanceof HermesUserScopeChangedError)) {
          setSessionError(errorMessage(reason, '无法删除该会话，请稍后重试。'))
        }
        return false
      }
      await sessionApi.remove(storedSessionId)
      if (!scopeGuard.isCurrent(scope, sessionApiIdentityRef.current)) return false
      replaceSessions(
        sessionsRef.current.filter((item) => item.storedSessionId !== storedSessionId),
      )
      if (selectedStoredSessionIdRef.current === storedSessionId) newConversation()
      return true
    },
    [newConversation, replaceSessions, scopeGuard, sessionApi],
  )

  useEffect(() => {
    if (stopFeedbackTimer.current) {
      clearTimeout(stopFeedbackTimer.current)
      stopFeedbackTimer.current = null
    }
    previousConnectionStatus.current = 'idle'
    setLoadingSessions(true)
    replaceSessions([])
    newConversation()
    if (userScopeReady) void refreshSessions()
    else setLoadingSessions(false)
  }, [newConversation, refreshSessions, replaceSessions, userScopeKey, userScopeReady])

  useEffect(() => {
    if (
      userScopeReady &&
      requestedSessionId &&
      requestedSessionId !== selectedStoredSessionIdRef.current &&
      !openingSession
    ) {
      void openSession(requestedSessionId)
    }
  }, [openSession, openingSession, requestedSessionId, userScopeReady])

  useEffect(() => {
    if (!sessionClient) return undefined
    let scope: HermesUserScopeSnapshot<HermesSessionClient>
    try {
      scope = captureCurrentScope(
        scopeGuard,
        sessionClient,
        sessionClientIdentityRef.current,
      )
    } catch {
      return undefined
    }
    const unsubscribe = sessionClient.subscribe((event) => {
      if (!scopeGuard.isCurrent(scope, sessionClientIdentityRef.current)) return
      const session = activeSessionRef.current
      if (!session || (event.sessionId && event.sessionId !== session.liveSessionId)) return
      const currentState = messageStateRef.current
      const barrierDecision = runStartBarrierDecision(
        runStartBarrierRef.current,
        session.storedSessionId,
        event.type,
        runIdFromEventPayload(event.payload),
      )
      if (barrierDecision === 'ignore') return
      if (barrierDecision === 'clear') runStartBarrierRef.current = null
      if (
        shouldIgnoreEventAfterInterrupt(currentState, event) ||
        shouldIgnoreKnownInterruptedEvent(
          event,
          currentState.activeRunId,
          interruptedRunIdsRef.current,
        )
      ) return
      if (isInterruptedTerminalEvent(event.type, event.payload)) {
        const runId = runIdFromEventPayload(event.payload) || currentState.activeRunId
        if (runId) interruptedRunIdsRef.current.add(runId)
        interruptedSessionAwaitingNextRunRef.current = session.storedSessionId
      }
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
  }, [applyEvent, scopeGuard, sessionClient, updateActiveBookmark])

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
      const client = sessionClientIdentityRef.current
      let scope: HermesUserScopeSnapshot<typeof client>
      try {
        scope = captureCurrentScope(scopeGuard, client, sessionClientIdentityRef.current)
      } catch {
        return
      }
      void restoreSelectedSession(bookmark, true).catch((reason) => {
        if (
          !scopeGuard.isCurrent(scope, sessionClientIdentityRef.current) ||
          selectedStoredSessionIdRef.current !== bookmark.storedSessionId
        ) return
        setSessionError(errorMessage(reason, '实时会话重新连接失败。'))
      })
    }
  }, [connectionStatus, restoreSelectedSession, scopeGuard])

  return {
    sessions,
    activeSession,
    userScopeReady,
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

function interruptedEvent(liveSessionId: string, runId: string | null): HermesEvent {
  return {
    type: 'session.interrupted',
    sessionId: liveSessionId,
    payload: {
      ...(runId ? { run_id: runId } : {}),
      message: '用户已停止当前任务。',
    },
  }
}

function shouldIgnoreKnownInterruptedEvent(
  event: HermesEvent,
  activeRunId: string | null,
  interruptedRunIds: ReadonlySet<string>,
): boolean {
  const interruptedTerminal = isInterruptedTerminalEvent(event.type, event.payload)
  if (!interruptedTerminal && !isRunScopedEventType(event.type)) return false
  return shouldIgnoreKnownInterruptedRunEvent(
    interruptedRunIds,
    runIdFromEventPayload(event.payload),
    activeRunId,
    interruptedTerminal,
  )
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
        checkpoint: checkpointFromPayload(
          event.type,
          event.payload,
          runIdFromEventPayload(event.payload),
        ),
      },
      session,
    )
    return
  }
  const status = terminalStatus(event)
  if (event.type === 'clarify.respond' && event.payload.source === 'plan_timeout') {
    await update({ status: 'running', checkpoint: null }, session)
    return
  }
  if (status) await update({ status, activeRunId: null, checkpoint: null }, session)
}

function terminalStatus(event: HermesEvent): HermesSessionStatus | null {
  if (event.type === 'run.failed' || event.type === 'error') return 'failed'
  if (isInterruptedTerminalEvent(event.type, event.payload)) return 'interrupted'
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

class HermesUserScopeChangedError extends Error {
  constructor() {
    super('APO 用户已切换，本次 Agent 操作已取消。')
    this.name = 'HermesUserScopeChangedError'
  }
}

class HermesSessionOperationChangedError extends Error {
  constructor() {
    super('会话已切换，本次 Agent 操作已取消。')
    this.name = 'HermesSessionOperationChangedError'
  }
}

function captureCurrentScope<T>(
  guard: HermesUserScopeGeneration,
  identity: T,
  currentIdentity: T,
): HermesUserScopeSnapshot<T> {
  const snapshot = guard.capture(identity)
  if (!guard.isCurrent(snapshot, currentIdentity)) {
    throw new HermesUserScopeChangedError()
  }
  return snapshot
}
