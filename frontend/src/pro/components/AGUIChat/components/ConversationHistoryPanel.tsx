import React from 'react'
import { Button, Empty, Layout, Spin } from 'antd'
import { useTranslation } from 'react-i18next'

import type { ConversationHistoryItem } from '../../../api/conversation'
import {
  buildConversationHistoryItemViewModels,
  resolveConversationHistoryPanelMotionState,
  resolveConversationHistoryPanelState,
} from '../../../api/conversationHistoryView'
import styles from '../AGUIChat.module.scss'

const { Sider } = Layout

interface ConversationHistoryPanelProps {
  items: ConversationHistoryItem[]
  loading: boolean
  error: string | null
  hasMore: boolean
  loadingMore: boolean
  loadMoreError: string | null
  activeThreadId?: string
  onSelectThread: (threadId: string) => void
  onStartNewConversation: () => void
  onLoadMore: () => void
  onRetry: () => void
  collapsed?: boolean
  mode?: 'sider' | 'drawer'
}

const ConversationHistoryPanel: React.FC<ConversationHistoryPanelProps> = ({
  items,
  loading,
  error,
  hasMore,
  loadingMore,
  loadMoreError,
  activeThreadId,
  onSelectThread,
  onStartNewConversation,
  onLoadMore,
  onRetry,
  collapsed = false,
  mode = 'sider',
}) => {
  const { t } = useTranslation('pro/agent')

  const itemViewModels = buildConversationHistoryItemViewModels(items, {
    activeThreadId,
    messages: {
      minutesAgo: (value) => t('history.time.minutesAgo', { count: value }),
      hoursAgo: (value) => t('history.time.hoursAgo', { count: value }),
      yesterday: (value) => t('history.time.yesterday', { value }),
      daysAgo: (value) => t('history.time.daysAgo', { count: value }),
    },
  })
  const panelState = resolveConversationHistoryPanelState({
    itemCount: itemViewModels.length,
    loading,
    error,
  })
  const motionState = resolveConversationHistoryPanelMotionState({
    useDrawerMode: mode === 'drawer',
    isCollapsed: collapsed,
  })

  let content: React.ReactNode = null

  if (panelState === 'loading') {
    content = (
      <div className={styles.historyPanelState}>
        <Spin size="small" />
        <div className={styles.historyPanelStateTitle}>{t('history.loading')}</div>
      </div>
    )
  } else if (panelState === 'error') {
    content = (
      <div className={styles.historyPanelState}>
        <div className={styles.historyPanelStateTitle}>{t('history.errorTitle')}</div>
        <div className={styles.historyPanelStateDescription}>{error}</div>
        <Button onClick={onRetry}>{t('history.retry')}</Button>
      </div>
    )
  } else if (panelState === 'empty') {
    content = (
      <div className={styles.historyPanelState}>
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={
            <div className={styles.historyPanelStateDescription}>
              {t('history.emptyDescription')}
            </div>
          }
        />
      </div>
    )
  } else {
    content = (
      <div className={styles.historyPanelList}>
        {itemViewModels.map((item) => (
          <button
            key={item.threadId}
            type="button"
            className={`${styles.historyPanelItem} ${item.isActive ? styles.historyPanelItemActive : ''}`}
            onClick={() => onSelectThread(item.threadId)}
          >
            <div className={styles.historyPanelItemTitle}>{item.title}</div>
            <div className={styles.historyPanelItemTime}>{item.timeLabel}</div>
          </button>
        ))}
      </div>
    )
  }

  const panelContent = (
    <div
      className={[
        styles.historyPanel,
        motionState.keepFixedWidth ? styles.historyPanelDesktop : '',
        mode === 'drawer' ? styles.historyPanelDrawer : '',
        motionState.hideContentDuringCollapse ? styles.historyPanelCollapsed : '',
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <div className={styles.historyPanelHeader}>
        <div className={styles.historyPanelTitle}>{t('history.title')}</div>
        <Button type="primary" onClick={onStartNewConversation}>
          {t('history.newConversation')}
        </Button>
      </div>
      {content}
      {panelState === 'list' && (
        <div className={styles.historyPanelFooter}>
          {error && (
            <div className={styles.historyPanelInlineError}>
              <span className={styles.historyPanelInlineErrorText}>{error}</span>
              <Button type="link" size="small" onClick={onRetry}>
                {t('history.retry')}
              </Button>
            </div>
          )}
          {loadMoreError && (
            <div className={styles.historyPanelInlineError}>
              <span className={styles.historyPanelInlineErrorText}>{loadMoreError}</span>
            </div>
          )}
          {(hasMore || loadingMore) && (
            <Button block onClick={onLoadMore} loading={loadingMore} disabled={loading || loadingMore}>
              {loadingMore ? t('history.loadingMore') : t('history.loadMore')}
            </Button>
          )}
        </div>
      )}
    </div>
  )

  if (mode === 'drawer') {
    return panelContent
  }

  return (
    <Sider
      width={300}
      collapsed={collapsed}
      collapsedWidth={0}
      trigger={null}
      collapsible
      className={styles.historyPanelSider}
    >
      {panelContent}
    </Sider>
  )
}

export default ConversationHistoryPanel
