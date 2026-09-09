import { Alert, Button, Spin, Tooltip } from 'antd'
import { MenuUnfoldOutlined, ReloadOutlined } from '@ant-design/icons'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import ChatInput from './components/ChatInput'
import ChatMessages from './components/ChatMessages'
import ConnectionStatus from './components/ConnectionStatus'
import SessionSidebar from './components/SessionSidebar'
import { useHermesSession } from './state/useHermesSession.ts'
import type { AgentLaunchContext, WorkflowLaunchContext } from './capabilities/client/types.ts'
import { canonicalHermesUserId } from './client/userScope.ts'
import styles from './HermesAgentWorkspace.module.scss'

export default function HermesAgentWorkspace() {
  const rawUserId = useSelector((state: any) => state.userReducer?.user?.userId)
  const userId = canonicalHermesUserId(rawUserId)
  return <HermesAgentWorkspaceContent key={userId ? `user:${userId}` : 'pending-user'} />
}

function HermesAgentWorkspaceContent() {
  const [searchParams, setSearchParams] = useSearchParams()
  const location = useLocation()
  const navigate = useNavigate()
  const requestedSessionId = searchParams.get('sessionId') || undefined
  const [sidebarOpen, setSidebarOpen] = useState(
    () => typeof window !== 'undefined' && window.innerWidth > 1366,
  )
  const workspace = useHermesSession(requestedSessionId)
  const [workflowLaunch, setWorkflowLaunch] = useState<WorkflowLaunchContext | null>(null)
  const [agentLaunch, setAgentLaunch] = useState<AgentLaunchContext | null>(null)
  const [draft, setDraft] = useState('')
  const autoSubmittedLaunchRef = useRef('')

  const setRouteSession = useCallback((storedSessionId?: string) => {
    const next = new URLSearchParams(searchParams)
    if (storedSessionId) next.set('sessionId', storedSessionId)
    else next.delete('sessionId')
    next.delete('threadId')
    setSearchParams(next, { replace: true })
  }, [searchParams, setSearchParams])

  useEffect(() => {
    const state = location.state as {
      agentLaunch?: AgentLaunchContext
      workflowLaunch?: WorkflowLaunchContext
    } | null
    const launch = state?.workflowLaunch
    const question = state?.agentLaunch
    if (!launch?.goal && !question?.goal) return
    workspace.newConversation()
    autoSubmittedLaunchRef.current = ''
    setWorkflowLaunch(launch || null)
    setAgentLaunch(question || null)
    setDraft(launch?.goal || question?.goal || '')
    navigate(`${location.pathname}${location.search}${location.hash}`, {
      replace: true,
      state: null,
    })
  }, [location.hash, location.key, location.pathname, location.search, location.state, navigate, workspace.newConversation])

  useEffect(() => {
    const launch = agentLaunch
    const text = draft.trim()
    if (
      !launch?.autoSend ||
      !text ||
      !workspace.userScopeReady ||
      workspace.openingSession ||
      workspace.state.busy ||
      workspace.state.checkpoint ||
      workspace.state.stopping
    ) return

    const launchKey = `${launch.source}:${launch.goal}`
    if (autoSubmittedLaunchRef.current === launchKey) return
    autoSubmittedLaunchRef.current = launchKey
    setDraft('')

    void workspace.sendMessage(text)
      .then((storedSessionId) => {
        if (!storedSessionId) {
          setDraft(text)
          return
        }
        setAgentLaunch(null)
        if (storedSessionId !== requestedSessionId) setRouteSession(storedSessionId)
      })
      .catch(() => {
        setDraft(text)
      })
  }, [
    agentLaunch,
    draft,
    requestedSessionId,
    setRouteSession,
    workspace.openingSession,
    workspace.sendMessage,
    workspace.state.busy,
    workspace.state.checkpoint,
    workspace.state.stopping,
    workspace.userScopeReady,
  ])

  const openSession = async (storedSessionId: string) => {
    setWorkflowLaunch(null)
    setAgentLaunch(null)
    setDraft('')
    setRouteSession(storedSessionId)
    await workspace.openSession(storedSessionId)
  }

  const newConversation = () => {
    setRouteSession()
    workspace.newConversation()
    setWorkflowLaunch(null)
    setAgentLaunch(null)
    setDraft('')
  }

  const useExample = (text: string) => {
    setWorkflowLaunch(null)
    setAgentLaunch(null)
    setDraft(text)
  }

  const send = async (text: string) => {
    const storedSessionId = await workspace.sendMessage(text, workflowLaunch?.workflowId)
    if (!storedSessionId) throw new Error('消息发送失败。')
    setWorkflowLaunch(null)
    setAgentLaunch(null)
    if (storedSessionId && storedSessionId !== requestedSessionId) setRouteSession(storedSessionId)
  }

  const removeSession = async (storedSessionId: string) => {
    const deleted = await workspace.deleteSession(storedSessionId)
    if (deleted && storedSessionId === workspace.selectedStoredSessionId) setRouteSession()
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
          onOpenCapabilities={() => navigate('/agent/capabilities')}
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
                disabled={!workspace.userScopeReady}
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

        {workflowLaunch && (
          <Alert
            className={styles.workspaceAlert}
            type="info"
            showIcon
            closable
            message={`已带入任务：${workflowLaunch.workflowName || '运维任务'} · 查看时版本 v${workflowLaunch.version}`}
            description="发送后将明确选择该 Hermes Workflow，并由 Hermes 使用发送时的当前可用版本创建新运行；不会修改任务库。"
            onClose={() => setWorkflowLaunch(null)}
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
              onExample={useExample}
            />
          )}
        </div>

        <ChatInput
          value={draft}
          busy={workspace.state.busy}
          stopping={workspace.state.stopping}
          stopRetryAllowed={workspace.state.stopRetryAllowed}
          disabled={
            !workspace.userScopeReady ||
            workspace.openingSession ||
            Boolean(workspace.state.checkpoint) ||
            workspace.state.stopping
          }
          onChange={setDraft}
          onSend={send}
          onStop={workspace.interrupt}
        />
      </main>
    </div>
  )
}
