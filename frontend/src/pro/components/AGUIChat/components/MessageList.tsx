import React, { useLayoutEffect, useRef } from 'react'
import ToolRender from '../ToolRender'
import styles from '../AGUIChat.module.scss'
import { useSelector } from 'react-redux'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'

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

interface MessageListProps {
  messages: Message[]
  currentMessage: Message | null
  loading: boolean
  currentToolCall: any
  isRunning: boolean
}

const MessageList: React.FC<MessageListProps> = ({
  messages,
  currentMessage,
  loading,
  currentToolCall,
  isRunning,
}) => {
  const { theme } = useSelector((state: any) => state.settingReducer)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messagesContainerRef = useRef<HTMLDivElement>(null)

  // 检查用户是否在底部（增加容差）
  const isUserAtBottom = () => {
    if (!messagesContainerRef.current) return true
    const { scrollTop, scrollHeight, clientHeight } = messagesContainerRef.current

    return scrollTop + clientHeight >= scrollHeight - 500 // 容差加大
  }

  // 自动滚动到底部
  const scrollToBottom = () => {
    if (messagesContainerRef.current) {
      messagesContainerRef.current.scrollTo({
        top: messagesContainerRef.current.scrollHeight,
        behavior: 'smooth',
      })
    }
  }

  useLayoutEffect(() => {
    const container = messagesContainerRef.current
    if (!container) return

    const handleChange = () => {
      if (isUserAtBottom()) {
        scrollToBottom()
      }
    }

    // ResizeObserver 监听容器内容尺寸变化（比如 Markdown 高亮加载）
    const resizeObserver = new ResizeObserver(handleChange)
    resizeObserver.observe(container)

    // MutationObserver 监听子元素变化（比如消息 DOM 加载完）
    const mutationObserver = new MutationObserver(handleChange)
    mutationObserver.observe(container, { childList: true, subtree: true })

    return () => {
      resizeObserver.disconnect()
      mutationObserver.disconnect()
    }
  }, [messages.length])

  const renderMessage = (message: Message) => {
    const messageClass = `${styles.message} ${message.role === 'user' ? styles.user : styles.assistant}`
    return (
      <div key={message.id} className={messageClass}>
        <div className={styles.messageContent}>
          <div className={styles.messageText}>
            <MarkdownRender content={message.content} theme={theme} />
          </div>
          <div className={styles.messageMeta}>
            {message.metadata?.model && `· ${message.metadata.model}`}
          </div>
        </div>
        {message.toolCalls?.length > 0 && (
          <div className="message-tool-calls">
            {message.toolCalls.map((toolCall: any) => (
              <ToolRender key={toolCall.id} toolCall={toolCall} />
            ))}
          </div>
        )}
      </div>
    )
  }

  const renderCurrentMessage = () => {
    if (!currentMessage) return null

    return (
      <div className={`${styles.message} ${styles.assistant} ${styles.currentMessage}`}>
        <div className={styles.messageContent}>
          <div className={styles.messageText}>
            <MarkdownRender content={currentMessage.content} theme={theme} />
            <span className={styles.typingIndicator}>▎</span>
          </div>
          <div className={styles.messageMeta}>正在输入...</div>
        </div>
      </div>
    )
  }

  const renderLoadingIndicator = () => {
    if (!isRunning) return null

    return (
      <div className="h-6">
        <span className={styles.loadingDots}>
          <span>.</span>
          <span>.</span>
          <span>.</span>
        </span>
      </div>
    )
  }

  return (
    <div ref={messagesContainerRef} className={styles.messagesArea}>
      {messages.map(renderMessage)}
      {renderCurrentMessage()}
      {renderLoadingIndicator()}
      <div ref={messagesEndRef} />
    </div>
  )
}

export default MessageList
