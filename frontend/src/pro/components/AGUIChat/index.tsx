import React, { useState, useEffect } from 'react'
import { Layout } from 'antd'
import { useAGUIChat, useAGUISession, useAGUIState } from '../../contexts/AGUIContext'
import { ChatHeader, ChatMain, SettingsModal } from './components'
import { ClientConfig } from './types/index'
import styles from './AGUIChat.module.scss'
import { useSearchParams } from 'react-router-dom'

// 控制会话功能的开关
const ENABLE_SESSION_MANAGEMENT = false

interface AGUIChatProps {
  initialSessionId?: string
  threadId?: string
}

const AGUIChat: React.FC<AGUIChatProps> = ({ initialSessionId, threadId }) => {
  // 使用AI权限检查hook - 目前暂时未使用，但保留以备将来使用
  // const aiPermission = useAGUIPermission()

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
    clearMessages,
    loadHistoryMessages,
  } = useAGUIChat()

  // 会话管理 - 仅在启用时使用
  const sessionContext = useAGUISession()
  const currentSession = ENABLE_SESSION_MANAGEMENT ? sessionContext.currentSession : null
  const { createSession, loadSession, refreshSessions } = sessionContext

  const { state, error, isConnected, clearError } = useAGUIState()

  // 本地UI状态
  const [inputText, setInputText] = useState('')
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [siderCollapsed, setSiderCollapsed] = useState(false)
  const [clientConfig, setClientConfig] = useState<ClientConfig>({
    baseUrl: (import.meta as any).env?.VITE_AGUI_BASE_URL || 'http://localhost:3001',
    apiKey: (import.meta as any).env?.VITE_AGUI_API_KEY || '',
  })

  // 初始化 - 仅在启用会话管理时执行
  useEffect(() => {
    if (ENABLE_SESSION_MANAGEMENT) {
      refreshSessions()
    }
  }, [refreshSessions])

  // 加载初始会话 - 仅在启用会话管理时执行
  useEffect(() => {
    if (ENABLE_SESSION_MANAGEMENT && initialSessionId && !currentSession) {
      loadSession(initialSessionId)
    }
  }, [initialSessionId, currentSession, loadSession])

  // 加载历史消息 - 当提供 threadId 时
  useEffect(() => {
    if (threadId) {
      loadHistoryMessages(threadId)
    }
  }, [threadId, loadHistoryMessages])

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
      // 中止当前运行的消息流
      if (isRunning) {
        cancelRun()
      }
      // 清除错误状态
      clearError()
      // 清空消息列表
      clearMessages()

      // 如果当前 URL 中有 threadId 参数，清空它
      if (searchParams.has('threadId')) {
        const newSearchParams = new URLSearchParams(searchParams)
        newSearchParams.delete('threadId')
        setSearchParams(newSearchParams, { replace: true })
      }
    }
  }

  // 消息处理函数
  const handleSendMessage = async (props?: { question?: string }) => {
    const { question } = props || {}
    console.log('handleSendMessage', question)
    if (!inputText.trim() && !question) return

    // 如果启用了会话管理，需要检查当前会话
    if (ENABLE_SESSION_MANAGEMENT && !currentSession) return

    try {
      console.log('sendMessage', question || inputText)
      sendMessage(question || inputText)
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
      {messages.length > 0 && (
        <ChatHeader
          siderCollapsed={siderCollapsed}
          onToggleSider={() => setSiderCollapsed(!siderCollapsed)}
          onOpenSettings={() => setSettingsOpen(true)}
          onRefreshSessions={handleRefresh}
        />
      )}

      <Layout className={styles.content}>
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
          isConnected={isConnected}
          state={state}
          inputText={inputText}
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
