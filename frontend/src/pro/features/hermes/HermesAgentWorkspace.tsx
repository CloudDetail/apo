import { Alert, Button, Spin, Tooltip } from 'antd'
import { MenuUnfoldOutlined, ReloadOutlined } from '@ant-design/icons'
import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import ChatInput from './components/ChatInput'
import ChatMessages from './components/ChatMessages'
import ConnectionStatus from './components/ConnectionStatus'
import SessionSidebar from './components/SessionSidebar'
import { useHermesSession } from './state/useHermesSession.ts'
import styles from './HermesAgentWorkspace.module.scss'

export default function HermesAgentWorkspace() {
  const [searchParams, setSearchParams] = useSearchParams()
  const requestedSessionId = searchParams.get('sessionId') || undefined
  const [sidebarOpen, setSidebarOpen] = useState(
    () => typeof window !== 'undefined' && window.innerWidth > 1366,
  )
  const workspace = useHermesSession(requestedSessionId)

  const setRouteSession = (storedSessionId?: string) => {
    const next = new URLSearchParams(searchParams)
    if (storedSessionId) next.set('sessionId', storedSessionId)
    else next.delete('sessionId')
    next.delete('threadId')
    setSearchParams(next, { replace: true })
  }

  const openSession = async (storedSessionId: string) => {
    setRouteSession(storedSessionId)
    await workspace.openSession(storedSessionId)
  }

  const newConversation = () => {
    setRouteSession()
    workspace.newConversation()
  }

  const send = async (text: string) => {
    const storedSessionId = await workspace.sendMessage(text)
    if (storedSessionId && storedSessionId !== requestedSessionId) setRouteSession(storedSessionId)
  }

  const removeSession = async (storedSessionId: string) => {
    await workspace.deleteSession(storedSessionId)
    if (storedSessionId === workspace.selectedStoredSessionId) setRouteSession()
  }

  const title =
    workspace.sessions.find((item) => item.storedSessionId === workspace.selectedStoredSessionId)
      ?.title || '新的运维分析'

  return (
    <div className={styles.workspace}>
      {sidebarOpen && (
        <SessionSidebar
          sessions={workspace.sessions}
          activeSessionId={workspace.selectedStoredSessionId || undefined}
          loading={workspace.loadingSessions}
          onSelect={(id) => void openSession(id)}
          onNew={newConversation}
          onRename={workspace.renameSession}
          onDelete={removeSession}
          onClose={() => setSidebarOpen(false)}
        />
      )}

      <main className={styles.chatWorkspace}>
        <header className={styles.workspaceHeader}>
          <div className={styles.headerTitle}>
            {!sidebarOpen && (
              <Tooltip title="展开会话列表">
                <Button
                  type="text"
                  icon={<MenuUnfoldOutlined />}
                  onClick={() => setSidebarOpen(true)}
                />
              </Tooltip>
            )}
            <div>
              <span className={styles.eyebrow}>Hermes 运维助手</span>
              <strong title={title}>{title}</strong>
            </div>
          </div>
          <div className={styles.headerStatus}>
            <span
              className={`${styles.taskStatus} ${
                workspace.state.busy ? styles.taskStatusActive : ''
              }`}
            >
              {workspace.state.statusText}
            </span>
            <ConnectionStatus status={workspace.connectionStatus} />
            <Tooltip title="刷新历史会话">
              <Button
                type="text"
                icon={<ReloadOutlined />}
                onClick={() => void workspace.refreshSessions()}
              />
            </Tooltip>
          </div>
        </header>

        {workspace.error && (
          <Alert
            type="error"
            showIcon
            closable
            className={styles.workspaceAlert}
            message={workspace.error}
            onClose={workspace.clearError}
          />
        )}

        <div className={styles.conversationArea}>
          {workspace.openingSession ? (
            <div className={styles.workspaceLoading}>
              <Spin size="large" />
              <span>正在加载历史记录…</span>
            </div>
          ) : (
            <ChatMessages
              timeline={workspace.state.timeline}
              reasoning={workspace.state.reasoning}
              busy={workspace.state.busy}
              activeCheckpoint={workspace.state.checkpoint}
              hydrationVersion={workspace.state.hydrationVersion}
              onCheckpointRespond={workspace.respondCheckpoint}
              onExample={(text) => void send(text)}
            />
          )}
        </div>

        <ChatInput
          busy={workspace.state.busy}
          stopping={workspace.state.stopping}
          stopRetryAllowed={workspace.state.stopRetryAllowed}
          disabled={
            workspace.openingSession ||
            Boolean(workspace.state.checkpoint) ||
            workspace.state.stopping
          }
          onSend={send}
          onStop={workspace.interrupt}
        />
      </main>
    </div>
  )
}
