import React, { useLayoutEffect, useRef, useState, useEffect } from 'react'
import styles from '../../AGUIChat.module.scss'
import { useSelector } from 'react-redux'
import { useTranslation } from 'react-i18next'
import { ExtendedAGUIMessage } from 'src/pro/contexts/AGUIContext'
import MessageItem from './MessageItem'

interface MessageListProps {
  messages: ExtendedAGUIMessage[]
  currentMessage: ExtendedAGUIMessage | null
  loading: boolean
  currentToolCall: any
  isRunning: boolean
  activeThreadId?: string
  showMissingAssistantResponseNotice?: boolean
}

const MessageList: React.FC<MessageListProps> = ({
  messages,
  currentMessage,
  isRunning,
  activeThreadId,
  showMissingAssistantResponseNotice = false,
}) => {
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
  }, [messages])

  const renderLoadingIndicator = () => {
    if (!isRunning) return null
    return (
      <div className="h-0">
        <span className={styles.loadingDots}>
          <span>.</span>
          <span>.</span>
          <span>.</span>
        </span>
      </div>
    )
  }

  // 过滤消息：如果有多个 findings 类型的消息，只保留第一个
  const filteredMessages = React.useMemo(() => {
    let findingsFound = false
    return messages.filter((msg) => {
      if (msg.type === 'findings') {
        if (findingsFound) {
          return false // 已经有一个 findings 了，跳过这个
        }
        findingsFound = true
        return true // 这是第一个 findings，保留
      }
      return true // 非 findings 类型的消息全部保留
    })
  }, [messages])

  return (
    <div ref={messagesContainerRef} className={styles.messagesArea}>
      {filteredMessages.map((message) => (
        <MessageItem key={message.id} message={message} activeThreadId={activeThreadId} />
      ))}
      {showMissingAssistantResponseNotice && (
        <div className={`${styles.message} ${styles.assistant}`}>
          <div className={`${styles.messageContent} ${styles.historyMissingAssistantNotice}`}>
            <div className={styles.messageText}>{t('history.missingAssistantResponse')}</div>
          </div>
        </div>
      )}
      {/* {renderCurrentMessage()} */}
      {renderLoadingIndicator()}
      <div ref={messagesEndRef} />
    </div>
  )
}

export default MessageList
