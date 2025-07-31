import React, { useLayoutEffect, useRef, useState, useEffect } from 'react'
import ToolRender from './AGUITool'
import styles from '../AGUIChat.module.scss'
import { useSelector } from 'react-redux'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import AgentToolDisplay from './AgentToolDisplay'
import { useTranslation } from 'react-i18next'
import { ExtendedAGUIMessage } from 'src/pro/contexts/AGUIContext'
import { PiShareFatBold } from 'react-icons/pi'
import { FaRegCopy } from 'react-icons/fa'

interface MessageListProps {
  messages: ExtendedAGUIMessage[]
  currentMessage: ExtendedAGUIMessage | null
  loading: boolean
  currentToolCall: any
  isRunning: boolean
}

const FinalMessageSuffix = () => {
  return (
    <div className="flex items-center justify-start ml-5 gap-3 text-[var(--ant-color-text-secondary)]">
      <FaRegCopy />
      <PiShareFatBold />
    </div>
  )
}

const MessageList: React.FC<MessageListProps> = ({ messages, currentMessage, isRunning }) => {
  const { theme } = useSelector((state: any) => state.settingReducer)
  const { t } = useTranslation('pro/agent')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messagesContainerRef = useRef<HTMLDivElement>(null)

  const [userScrolledUp, setUserScrolledUp] = useState(false)
  const [hasMounted, setHasMounted] = useState(false)

  // 初始化完成后再允许判断滚动
  useEffect(() => {
    const timer = setTimeout(() => setHasMounted(true), 300)
    return () => clearTimeout(timer)
  }, [])

  // 用户主动交互行为，标记为“用户滚动过了”
  useEffect(() => {
    const el = messagesContainerRef.current
    if (!el) return

    const onUserScroll = () => {
      if (!hasMounted) return
      const { scrollTop, scrollHeight, clientHeight } = el
      if (scrollHeight - scrollTop - clientHeight > 50) {
        setUserScrolledUp(true)
      } else {
        setUserScrolledUp(false)
      }
    }

    // const markUserInteracted = () => {
    //   setUserScrolledUp(true)
    // }

    el.addEventListener('scroll', onUserScroll)
    // el.addEventListener('wheel', markUserInteracted, { passive: true })
    // el.addEventListener('touchstart', markUserInteracted, { passive: true })
    // el.addEventListener('mousedown', markUserInteracted)

    return () => {
      el.removeEventListener('scroll', onUserScroll)
      //   el.removeEventListener('wheel', markUserInteracted)
      //   el.removeEventListener('touchstart', markUserInteracted)
      //   el.removeEventListener('mousedown', markUserInteracted)
    }
  }, [hasMounted])

  // 自动滚动到底部
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  // 监听内容变化后滚动
  useLayoutEffect(() => {
    if (!messagesContainerRef.current) return

    const el = messagesContainerRef.current

    const tryScroll = () => {
      if (!userScrolledUp) {
        scrollToBottom()
      }
    }

    const resizeObserver = new ResizeObserver(() => {
      tryScroll()
    })
    resizeObserver.observe(el)

    const mutationObserver = new MutationObserver(() => {
      tryScroll()
    })
    mutationObserver.observe(el, { childList: true, subtree: true })

    return () => {
      resizeObserver.disconnect()
      mutationObserver.disconnect()
    }
  }, [messages.length, userScrolledUp])

  // 每次新消息也尝试滚动
  useEffect(() => {
    if (!userScrolledUp) {
      scrollToBottom()
    }
  }, [messages, currentMessage])

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
          <div className="message-tool-calls ml-6 ">
            {message.toolCalls.map((toolCall: any) => (
              <ToolRender key={toolCall.id} toolCall={toolCall} />
            ))}
          </div>
        )}
        {message.agentToolDisplay?.length > 0 && (
          <div className="message-tool-calls ml-6">
            {message.agentToolDisplay.map((toolCall: any, index: number) => (
              <AgentToolDisplay
                key={toolCall.toolName + index}
                toolName={toolCall.toolName}
                title={toolCall.title}
                content={toolCall.content}
              />
            ))}
          </div>
        )}
        {message.isFinal && <FinalMessageSuffix />}
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
          <div className={styles.messageMeta}>{t('inputArea.typingIndicator')}</div>
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
