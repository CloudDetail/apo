import React from 'react'
import { CloseOutlined } from '@ant-design/icons'
import { Button, Layout, Spin } from 'antd'
import MessageList from './MeesageRender/MessageList'
import InputArea from './InputArea'
import QuickQuestions from './QuickQuestions'
import styles from '../AGUIChat.module.scss'
import { useAGUIPermission } from 'src/pro/contexts/AGUIContext'
import {
  resolveChatMainConversationSurface,
  resolveChatMainInlineAlert,
  shouldShowMissingAssistantResponseNotice,
} from 'src/pro/api/conversationHistoryView'

const { Content } = Layout

interface Session {
  id: string
  title: string
  updatedAt: Date | string
}

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp?: Date | string
  metadata?: {
    model?: string
  }
  toolCalls?: any[]
}

interface ChatMainProps {
  currentSession: Session | null
  messages: Message[]
  currentMessage: Message | null
  loading: boolean
  currentToolCall: any
  isRunning: boolean
  runId: string | null
  currentStep: string | null
  error: string | null
  hideInlineError?: boolean
  isConnected: boolean
  state: Record<string, any> | null
  inputText: string
  inputDisabled?: boolean
  isHistoryThread?: boolean
  onInputChange: (text: string) => void
  onSendMessage: (props?: { question?: string }) => void
  onCancelRun: () => void
  onClearError: () => void
  onCreateSession?: () => void
}

// 控制会话功能的开关
const ENABLE_SESSION_MANAGEMENT = false

const ChatMain: React.FC<ChatMainProps> = ({
  messages,
  currentMessage,
  loading,
  currentToolCall,
  isRunning,
  runId,
  error,
  hideInlineError = false,
  inputText,
  inputDisabled = false,
  isHistoryThread = false,
  onInputChange,
  onSendMessage,
  onCancelRun,
  onClearError,
  onCreateSession,
}) => {
  // 快捷问答处理函数
  const { hasPermission, permissionLoading, dataPlaneExist, defaultModel } = useAGUIPermission()
  const inlineAlert = resolveChatMainInlineAlert({ error, hideError: hideInlineError })
  const conversationSurface = resolveChatMainConversationSurface({
    messageCount: messages.length,
    loading,
  })
  const showMissingAssistantResponseNotice = shouldShowMissingAssistantResponseNotice({
    messages,
    isHistoryThread,
    loading,
    isRunning,
  })

  const handleQuickQuestion = (question?: string) => {
    if (ENABLE_SESSION_MANAGEMENT && onCreateSession) {
      onCreateSession()
      // 延迟设置输入文本并发送，确保会话已创建
      setTimeout(() => {
        onInputChange(question || '')
        onSendMessage({ question })
      }, 100)
    } else {
      // 直接设置输入文本并发送
      onSendMessage({ question })
    }
  }

  return (
    <Content className={styles.main}>
      {inlineAlert.showError && inlineAlert.message && (
        <div className={styles.errorAlert}>
          <span>{inlineAlert.message}</span>
          <Button
            type="text"
            icon={<CloseOutlined />}
            onClick={onClearError}
            className={styles.errorClose}
          />
        </div>
      )}

      {/* 消息区域或快捷问答 */}
      {conversationSurface === 'loading' ? (
        <div className={styles.chatLoadingState}>
          <Spin size="large" />
        </div>
      ) : conversationSurface === 'empty' ? (
        <QuickQuestions onQuestionClick={handleQuickQuestion} />
      ) : (
        <MessageList
          messages={messages}
          currentMessage={currentMessage}
          loading={loading}
          currentToolCall={currentToolCall}
          isRunning={isRunning}
          showMissingAssistantResponseNotice={showMissingAssistantResponseNotice}
        />
      )}

      {/* 输入区域 */}
      {!permissionLoading && hasPermission && dataPlaneExist && defaultModel && (
        <InputArea
          inputText={inputText}
          disabled={inputDisabled}
          loading={loading}
          isRunning={isRunning}
          runId={runId}
          onInputChange={onInputChange}
          onSendMessage={onSendMessage}
          onCancelRun={onCancelRun}
        />
      )}
    </Content>
  )
}

export default ChatMain
