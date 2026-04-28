import React, { useState, useEffect, useRef, useCallback } from 'react'
import { Drawer, Grid, Layout } from 'antd'
import { useSelector } from 'react-redux'
import { useAGUIChat, useAGUISession, useAGUIState } from '../../contexts/AGUIContext'
import { ChatHeader, ChatMain, ConversationHistoryPanel, SettingsModal } from './components'
import { ClientConfig } from './types/index'
import styles from './AGUIChat.module.scss'
import { useSearchParams } from 'react-router-dom'
import { shouldApplyRouteThreadSelection } from '../../contexts/threadTransition'
import { useTranslation } from 'react-i18next'
import { getConversationHistoryListApi, type ConversationHistoryItem } from '../../api/conversation'
import {
  appendConversationHistoryPageItems,
  getNextConversationHistoryPendingRefreshState,
  getConversationHistoryReconcileDelayMs,
  hasConversationHistoryItem,
  mergeConversationHistoryItems,
  removeConversationHistoryItem,
  resolveConversationHistoryReconcileDecision,
  shouldClearConversationHistoryLoadMoreState,
  shouldApplyConversationHistoryLoadMoreResponse,
  shouldScheduleConversationHistoryReconcile,
  shouldApplyConversationHistoryResponse,
  upsertConversationHistoryItem,
  type ConversationHistoryPendingRefreshState,
} from '../../api/conversationHistory'
import {
  buildThreadSearchParams,
  resolveRecoverableRouteThreadId,
  resolveConversationHistoryPanelLayout,
  resolveConversationHistorySelectionThreadId,
  shouldAllowAgentHistoryIdentityRequest,
  shouldMaterializePendingNewThreadHistory,
  shouldShowRecoverableRouteThreadErrorToast,
  shouldRecoverFailedRouteThreadOnSend,
  shouldWriteActiveThreadToRoute,
} from '../../api/conversationHistoryView'
import { notify } from 'src/core/utils/notify'

// 控制会话功能的开关
const ENABLE_SESSION_MANAGEMENT = false
const CONVERSATION_HISTORY_PAGE_SIZE = 30

interface AGUIChatProps {
  initialSessionId?: string
  threadId?: string
}

const AGUIChat: React.FC<AGUIChatProps> = ({ initialSessionId, threadId }) => {
  // 使用AI权限检查hook - 目前暂时未使用，但保留以备将来使用
  // const aiPermission = useAGUIPermission()
  const { t } = useTranslation('pro/agent')
  const screens = Grid.useBreakpoint()
  const useHistoryDrawer = screens.lg === false

  // 路由相关 hooks
  const [searchParams, setSearchParams] = useSearchParams()

  // 使用Context hooks
  const {
    messages,
    loading,
    currentMessage,
    currentToolCall,
    isRunning,
    runId,
    currentStep,
    sendMessage,
    cancelRun,
    openThread,
    resetToNewConversation,
  } = useAGUIChat()

  // 会话管理 - 仅在启用时使用
  const sessionContext = useAGUISession()
  const currentSession = ENABLE_SESSION_MANAGEMENT ? sessionContext.currentSession : null
  const { createSession, loadSession, refreshSessions } = sessionContext

  const { state, error, threadLoadErrorKind, isConnected, clearError, currentThreadId } =
    useAGUIState()
  const currentUserId = useSelector((store: any) => store.userReducer?.user?.userId)
  const hasResolvedAgentUserId = shouldAllowAgentHistoryIdentityRequest(currentUserId)

  // 本地UI状态
  const [inputText, setInputText] = useState('')
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [siderCollapsed, setSiderCollapsed] = useState(true)
  const syncedThreadIdRef = useRef<string | undefined>(undefined)
  const [historyItems, setHistoryItems] = useState<ConversationHistoryItem[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)
  const [historyError, setHistoryError] = useState<string | null>(null)
  const [historyHasMore, setHistoryHasMore] = useState(false)
  const [historyNextCursor, setHistoryNextCursor] = useState<string | undefined>(undefined)
  const [historyLoadingMore, setHistoryLoadingMore] = useState(false)
  const [historyLoadMoreError, setHistoryLoadMoreError] = useState<string | null>(null)
  const [recoverableRouteThreadId, setRecoverableRouteThreadId] = useState<string | undefined>(
    undefined,
  )
  const [pendingNewThreadHistorySeed, setPendingNewThreadHistorySeed] = useState<{
    title: string
    timeValue: string
  } | null>(null)
  const [pendingHistoryRefresh, setPendingHistoryRefresh] = useState<
    ConversationHistoryPendingRefreshState | undefined
  >(undefined)
  const [clientConfig, setClientConfig] = useState<ClientConfig>({
    baseUrl: (import.meta as any).env?.VITE_AGUI_BASE_URL || 'http://localhost:3001',
    apiKey: (import.meta as any).env?.VITE_AGUI_API_KEY || '',
  })
  const currentSearch = searchParams.toString()
  const optimisticHistoryItemRef = useRef<ConversationHistoryItem | null>(null)
  const historyRequestRevisionRef = useRef(0)
  const historyLoadMoreRequestRevisionRef = useRef(0)
  const historyNextCursorRef = useRef<string | undefined>(undefined)
  const activeHistoryThreadIdRef = useRef<string | undefined>(undefined)
  const notifiedRecoverableRouteThreadIdRef = useRef<string | undefined>(undefined)
  const historyPanelLayout = resolveConversationHistoryPanelLayout({
    hasMessages: messages.length > 0,
    useDrawerMode: useHistoryDrawer,
    isCollapsed: siderCollapsed,
  })
  const activeHistoryThreadId = resolveConversationHistorySelectionThreadId({
    routeThreadId: threadId,
    currentThreadId,
  })
  activeHistoryThreadIdRef.current = activeHistoryThreadId

  const clearOptimisticHistoryItem = useCallback((threadId: string) => {
    if (optimisticHistoryItemRef.current?.threadId === threadId) {
      optimisticHistoryItemRef.current = null
    }

    setHistoryItems((prev) => removeConversationHistoryItem(prev, threadId))
  }, [])

  const clearPendingHistoryRefresh = useCallback((threadId: string) => {
    setPendingHistoryRefresh((current) => {
      if (!current || current.threadId !== threadId) {
        return current
      }

      return undefined
    })
  }, [])

  const refreshHistoryList = useCallback(
    async (options?: { expectedPersistedThreadId?: string }) => {
      const responseRequestRevision = historyRequestRevisionRef.current + 1
      historyRequestRevisionRef.current = responseRequestRevision
      historyLoadMoreRequestRevisionRef.current += 1

      setHistoryLoading(true)
      setHistoryLoadingMore(false)
      setHistoryError(null)
      setHistoryLoadMoreError(null)

      try {
        if (!hasResolvedAgentUserId) {
          setHistoryError(t('history.errorDescription'))
          return { outcome: 'error' as const, persistedPendingThread: false }
        }

        const response = await getConversationHistoryListApi({
          limit: CONVERSATION_HISTORY_PAGE_SIZE,
          fallbackTitle: t('history.fallbackTitle'),
          userId: String(currentUserId).trim(),
        })

        if (
          !shouldApplyConversationHistoryResponse({
            activeRequestRevision: historyRequestRevisionRef.current,
            responseRequestRevision,
          })
        ) {
          return { outcome: 'superseded' as const, persistedPendingThread: false }
        }

        const latestOptimisticHistoryItem = optimisticHistoryItemRef.current ?? undefined
        const expectedPersistedThreadId =
          options?.expectedPersistedThreadId ?? latestOptimisticHistoryItem?.threadId
        const persistedPendingThread = hasConversationHistoryItem(
          response.items,
          expectedPersistedThreadId,
        )

        setHistoryItems((prev) => {
          const retainedActiveHistoryItem = prev.find(
            (item) => item.threadId === activeHistoryThreadIdRef.current,
          )

          return mergeConversationHistoryItems(
            response.items,
            latestOptimisticHistoryItem,
            retainedActiveHistoryItem,
          )
        })
        setHistoryNextCursor(response.nextCursor)
        historyNextCursorRef.current = response.nextCursor
        setHistoryHasMore(response.hasMore)

        if (persistedPendingThread) {
          optimisticHistoryItemRef.current = null
        }

        return {
          outcome: 'applied' as const,
          persistedPendingThread,
        }
      } catch (refreshError) {
        if (
          !shouldApplyConversationHistoryResponse({
            activeRequestRevision: historyRequestRevisionRef.current,
            responseRequestRevision,
          })
        ) {
          return { outcome: 'superseded' as const, persistedPendingThread: false }
        }

        console.error('加载历史对话列表失败:', refreshError)
        setHistoryError(t('history.errorDescription'))
        return { outcome: 'error' as const, persistedPendingThread: false }
      } finally {
        if (
          shouldApplyConversationHistoryResponse({
            activeRequestRevision: historyRequestRevisionRef.current,
            responseRequestRevision,
          })
        ) {
          setHistoryLoading(false)
        }
      }
    },
    [currentUserId, hasResolvedAgentUserId, t],
  )

  const loadMoreHistoryList = useCallback(async () => {
    const responseCursor = historyNextCursorRef.current
    if (!responseCursor || historyLoading || historyLoadingMore) {
      return
    }

    const responseRequestRevision = historyLoadMoreRequestRevisionRef.current + 1
    historyLoadMoreRequestRevisionRef.current = responseRequestRevision

    setHistoryLoadingMore(true)
    setHistoryLoadMoreError(null)

    try {
      if (!hasResolvedAgentUserId) {
        return
      }

      const response = await getConversationHistoryListApi({
        cursor: responseCursor,
        limit: CONVERSATION_HISTORY_PAGE_SIZE,
        fallbackTitle: t('history.fallbackTitle'),
        userId: String(currentUserId).trim(),
      })

      if (
        !shouldApplyConversationHistoryLoadMoreResponse({
          activeRequestRevision: historyLoadMoreRequestRevisionRef.current,
          responseRequestRevision,
          activeCursor: historyNextCursorRef.current,
          responseCursor,
        })
      ) {
        return
      }

      const latestOptimisticHistoryItem = optimisticHistoryItemRef.current ?? undefined
      setHistoryItems((prev) =>
        appendConversationHistoryPageItems(prev, response.items, latestOptimisticHistoryItem),
      )
      setHistoryNextCursor(response.nextCursor)
      historyNextCursorRef.current = response.nextCursor
      setHistoryHasMore(response.hasMore)
    } catch (loadMoreError) {
      if (
        !shouldApplyConversationHistoryLoadMoreResponse({
          activeRequestRevision: historyLoadMoreRequestRevisionRef.current,
          responseRequestRevision,
          activeCursor: historyNextCursorRef.current,
          responseCursor,
        })
      ) {
        return
      }

      console.error('加载更多历史对话失败:', loadMoreError)
      setHistoryLoadMoreError(t('history.loadMoreError'))
    } finally {
      if (
        shouldClearConversationHistoryLoadMoreState({
          activeRequestRevision: historyLoadMoreRequestRevisionRef.current,
          responseRequestRevision,
        })
      ) {
        setHistoryLoadingMore(false)
      }
    }
  }, [currentUserId, hasResolvedAgentUserId, historyLoading, historyLoadingMore, t])

  // 初始化 - 仅在启用会话管理时执行
  useEffect(() => {
    if (ENABLE_SESSION_MANAGEMENT) {
      refreshSessions()
    }
  }, [refreshSessions])

  useEffect(() => {
    setSiderCollapsed(true)
  }, [useHistoryDrawer])

  useEffect(() => {
    if (!hasResolvedAgentUserId) {
      return
    }

    void refreshHistoryList()
  }, [hasResolvedAgentUserId, refreshHistoryList])

  // 加载初始会话 - 仅在启用会话管理时执行
  useEffect(() => {
    if (ENABLE_SESSION_MANAGEMENT && initialSessionId && !currentSession) {
      loadSession(initialSessionId)
    }
  }, [initialSessionId, currentSession, loadSession])

  // 加载历史消息 - 当提供 threadId 时
  useEffect(() => {
    if (!hasResolvedAgentUserId) {
      return
    }

    if (
      !shouldApplyRouteThreadSelection({
        previousRouteThreadId: syncedThreadIdRef.current,
        nextRouteThreadId: threadId,
        activeThreadId: currentThreadId,
      })
    ) {
      syncedThreadIdRef.current = threadId
      return
    }

    syncedThreadIdRef.current = threadId

    if (threadId) {
      void openThread(threadId)
      return
    }

    void resetToNewConversation()
  }, [hasResolvedAgentUserId, threadId, currentThreadId, openThread, resetToNewConversation])

  useEffect(() => {
    if (
      !shouldWriteActiveThreadToRoute({
        routeThreadId: threadId,
        activeThreadId: currentThreadId,
        hasPendingNewThread: pendingNewThreadHistorySeed !== null,
      })
    ) {
      return
    }

    setSearchParams(buildThreadSearchParams(currentSearch, currentThreadId), {
      replace: true,
    })
  }, [threadId, currentThreadId, pendingNewThreadHistorySeed, currentSearch, setSearchParams])

  useEffect(() => {
    setRecoverableRouteThreadId((current) => {
      const next = resolveRecoverableRouteThreadId({
        routeThreadId: threadId,
        activeThreadId: currentThreadId,
        loading,
        error,
        threadLoadErrorKind,
        recoverableRouteThreadId: current,
      })

      return current === next ? current : next
    })
  }, [threadId, currentThreadId, loading, error, threadLoadErrorKind])

  useEffect(() => {
    if (
      !shouldShowRecoverableRouteThreadErrorToast({
        routeThreadId: threadId,
        recoverableRouteThreadId,
        error,
      })
    ) {
      if (!recoverableRouteThreadId || recoverableRouteThreadId !== threadId) {
        notifiedRecoverableRouteThreadIdRef.current = undefined
      }
      return
    }

    if (notifiedRecoverableRouteThreadIdRef.current === recoverableRouteThreadId) {
      return
    }

    notifiedRecoverableRouteThreadIdRef.current = recoverableRouteThreadId
    notify({
      type: 'error',
      message: error!,
    })
  }, [threadId, recoverableRouteThreadId, error])

  useEffect(() => {
    if (
      !shouldMaterializePendingNewThreadHistory({
        hasPendingNewThread: pendingNewThreadHistorySeed !== null,
        activeThreadId: currentThreadId,
      }) ||
      !pendingNewThreadHistorySeed ||
      !currentThreadId
    ) {
      return
    }

    const nextOptimisticItem = {
      threadId: currentThreadId,
      title: pendingNewThreadHistorySeed.title,
      timeValue: pendingNewThreadHistorySeed.timeValue,
    }
    setHistoryItems((prev) => upsertConversationHistoryItem(prev, nextOptimisticItem))
    optimisticHistoryItemRef.current = nextOptimisticItem
    setPendingHistoryRefresh({
      threadId: currentThreadId,
      attemptCount: 0,
    })
    setPendingNewThreadHistorySeed(null)
  }, [pendingNewThreadHistorySeed, currentThreadId])

  useEffect(() => {
    if (
      !shouldScheduleConversationHistoryReconcile({
        hasPendingRefresh: pendingHistoryRefresh !== undefined,
        isRunning,
        loading,
      }) ||
      !pendingHistoryRefresh
    ) {
      return
    }

    let active = true
    const retryDelayMs = getConversationHistoryReconcileDelayMs(pendingHistoryRefresh.attemptCount)

    const timerId = window.setTimeout(() => {
      void refreshHistoryList({
        expectedPersistedThreadId: pendingHistoryRefresh.threadId,
      }).then((result) => {
        if (!active) {
          return
        }

        const decision = resolveConversationHistoryReconcileDecision({
          attemptCount: pendingHistoryRefresh.attemptCount,
          refreshOutcome: result.outcome,
          persistedPendingThread: result.persistedPendingThread,
        })

        if (decision.shouldClearOptimistic) {
          clearOptimisticHistoryItem(pendingHistoryRefresh.threadId)
        }

        if (decision.shouldClearPending) {
          clearPendingHistoryRefresh(pendingHistoryRefresh.threadId)
          return
        }

        setPendingHistoryRefresh((current) => {
          return getNextConversationHistoryPendingRefreshState({
            current,
            targetThreadId: pendingHistoryRefresh.threadId,
            nextAttemptCount: decision.nextAttemptCount,
          })
        })
      })
    }, retryDelayMs)

    return () => {
      active = false
      window.clearTimeout(timerId)
    }
  }, [
    pendingHistoryRefresh,
    isRunning,
    loading,
    refreshHistoryList,
    clearOptimisticHistoryItem,
    clearPendingHistoryRefresh,
  ])

  // 会话管理处理函数 - 仅在启用会话管理时使用
  const handleCreateSession = async () => {
    if (!ENABLE_SESSION_MANAGEMENT) return

    try {
      await createSession()
    } catch (err) {
      console.error('创建会话失败:', err)
    }
  }

  // 刷新处理函数 - 根据是否启用会话管理执行不同的刷新操作
  const handleRefresh = async () => {
    if (ENABLE_SESSION_MANAGEMENT) {
      // 会话管理模式下刷新会话列表
      await refreshSessions()
    } else {
      // 非会话管理模式下执行其他刷新操作
      // 清除错误状态
      clearError()

      if (threadId) {
        setSearchParams(buildThreadSearchParams(searchParams.toString(), undefined), {
          replace: true,
        })
      } else {
        await resetToNewConversation()
      }

      const pendingThreadId = pendingHistoryRefresh?.threadId
      const refreshResult = await refreshHistoryList({
        expectedPersistedThreadId: pendingThreadId,
      })
      if (
        pendingThreadId &&
        refreshResult.outcome === 'applied' &&
        refreshResult.persistedPendingThread
      ) {
        clearPendingHistoryRefresh(pendingThreadId)
      }
    }
  }

  const handleSelectHistoryThread = (selectedThreadId: string) => {
    setSearchParams(buildThreadSearchParams(searchParams.toString(), selectedThreadId))
    if (useHistoryDrawer) {
      setSiderCollapsed(true)
    }
  }

  const handleStartNewConversation = async () => {
    if (threadId) {
      setSearchParams(buildThreadSearchParams(searchParams.toString(), undefined))
      if (useHistoryDrawer) {
        setSiderCollapsed(true)
      }
      return
    }

    await resetToNewConversation()
    if (useHistoryDrawer) {
      setSiderCollapsed(true)
    }
  }

  const handleRetryHistoryList = useCallback(async () => {
    const pendingThreadId = pendingHistoryRefresh?.threadId
    const refreshResult = await refreshHistoryList({
      expectedPersistedThreadId: pendingThreadId,
    })

    if (
      pendingThreadId &&
      refreshResult.outcome === 'applied' &&
      refreshResult.persistedPendingThread
    ) {
      clearPendingHistoryRefresh(pendingThreadId)
    }
  }, [pendingHistoryRefresh, refreshHistoryList, clearPendingHistoryRefresh])

  // 消息处理函数
  const handleSendMessage = async (props?: { question?: string }) => {
    const nextQuestion = (props?.question || inputText).trim()
    if (!nextQuestion) return

    // 如果启用了会话管理，需要检查当前会话
    if (ENABLE_SESSION_MANAGEMENT && !currentSession) return
    if (!hasResolvedAgentUserId) return

    try {
      if (
        (!threadId && !currentThreadId) ||
        shouldRecoverFailedRouteThreadOnSend({
          routeThreadId: threadId,
          activeThreadId: currentThreadId,
          loading,
          error,
          threadLoadErrorKind,
          recoverableRouteThreadId,
        })
      ) {
        setPendingNewThreadHistorySeed({
          title: nextQuestion,
          timeValue: new Date().toISOString(),
        })
      }

      sendMessage(nextQuestion)
      setInputText('')
    } catch (err) {
      console.error('发送消息失败:', err)
    }
  }

  const handleCancelRun = () => {
    cancelRun()
  }

  // 设置处理函数
  const handleUpdateClientConfig = () => {
    setSettingsOpen(false)
  }

  const handleConfigChange = (config: ClientConfig) => {
    setClientConfig(config)
  }

  // 如果有权限，显示正常的AGUIChat界面
  return (
    <Layout className={styles.container}>
      {/* 头部 */}
      {historyPanelLayout.showHeader && (
        <ChatHeader
          siderCollapsed={siderCollapsed}
          showHistoryToggle={historyPanelLayout.showHistoryToggle}
          onToggleSider={() => setSiderCollapsed(!siderCollapsed)}
          onOpenSettings={() => setSettingsOpen(true)}
          onRefreshSessions={handleRefresh}
        />
      )}

      <Layout className={styles.content}>
        {useHistoryDrawer ? (
          <Drawer
            placement="left"
            width={300}
            open={historyPanelLayout.drawerOpen}
            onClose={() => setSiderCollapsed(true)}
            closable={false}
            maskClosable
            styles={{
              body: {
                padding: 0,
                height: '100%',
              },
            }}
          >
            <ConversationHistoryPanel
              mode="drawer"
              items={historyItems}
              loading={historyLoading}
              error={historyError}
              hasMore={historyHasMore && Boolean(historyNextCursor)}
              loadingMore={historyLoadingMore}
              loadMoreError={historyLoadMoreError}
              activeThreadId={activeHistoryThreadId}
              onSelectThread={handleSelectHistoryThread}
              onStartNewConversation={handleStartNewConversation}
              onLoadMore={() => {
                void loadMoreHistoryList()
              }}
              onRetry={() => {
                void handleRetryHistoryList()
              }}
            />
          </Drawer>
        ) : (
          <ConversationHistoryPanel
            items={historyItems}
            loading={historyLoading}
            error={historyError}
            hasMore={historyHasMore && Boolean(historyNextCursor)}
            loadingMore={historyLoadingMore}
            loadMoreError={historyLoadMoreError}
            activeThreadId={activeHistoryThreadId}
            collapsed={historyPanelLayout.siderCollapsed}
            onSelectThread={handleSelectHistoryThread}
            onStartNewConversation={handleStartNewConversation}
            onLoadMore={() => {
              void loadMoreHistoryList()
            }}
            onRetry={() => {
              void handleRetryHistoryList()
            }}
          />
        )}
        {/* 聊天主体 */}
        <ChatMain
          currentSession={currentSession}
          messages={messages}
          currentMessage={currentMessage}
          loading={loading}
          currentToolCall={currentToolCall}
          isRunning={isRunning}
          runId={runId}
          currentStep={currentStep}
          error={error}
          hideInlineError={Boolean(
            shouldShowRecoverableRouteThreadErrorToast({
              routeThreadId: threadId,
              recoverableRouteThreadId,
              error,
            }),
          )}
          isConnected={isConnected}
          state={state}
          inputText={inputText}
          inputDisabled={!hasResolvedAgentUserId}
          isHistoryThread={Boolean(threadId)}
          onInputChange={setInputText}
          onSendMessage={handleSendMessage}
          onCancelRun={handleCancelRun}
          onClearError={clearError}
          onCreateSession={ENABLE_SESSION_MANAGEMENT ? handleCreateSession : undefined}
        />
      </Layout>

      {/* 设置对话框 */}

      <SettingsModal
        open={settingsOpen}
        clientConfig={clientConfig}
        onClose={() => setSettingsOpen(false)}
        onConfigChange={handleConfigChange}
        onSave={handleUpdateClientConfig}
      />
    </Layout>
  )
}

export default AGUIChat
