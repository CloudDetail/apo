import React, { createContext, useContext, useState, useCallback, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { AGUIClient } from '../services/ag-ui-client'
import {
  AGUISession,
  AGUIClientConfig,
  AGUIMessage,
  Event,
  EventType,
  TextMessageStartEvent,
  TextMessageContentEvent,
  TextMessageEndEvent,
  ToolCallStartEvent,
  ToolCallArgsEvent,
  ToolCallEndEvent,
  ToolCallResultEvent,
  RunStartedEvent,
  RunFinishedEvent,
  RunErrorEvent,
  MessagesSnapshotEvent,
  StateSnapshotEvent,
  StateDeltaEvent,
  StepStartedEvent,
  StepFinishedEvent,
  RawEvent,
  CustomEvent,
  generateThreadId,
  generateRunId,
  generateMessageId,
} from '../types/ag-ui'
import dayjs from 'dayjs'

// 扩展AGUIMessage类型以支持工具调用
export interface ExtendedAGUIMessage extends AGUIMessage {
  toolCalls?: Array<{
    id: string
    name: string
    args: string
    result?: any
    status: 'pending' | 'render' | 'completed' | 'error'
  }>
  agentToolDisplay?: Array<{
    toolName: string
    title: string
    content: string
  }>
  isFinal?: boolean
  errorMessage?: string
}

interface AGUIContextValue {
  // 客户端状态
  client: AGUIClient
  isConnected: boolean
  error: string | null

  // 会话管理
  sessions: AGUISession[]
  currentSession: AGUISession | null

  // 聊天状态
  messages: ExtendedAGUIMessage[]
  loading: boolean
  state: Record<string, any>

  // 当前正在处理的消息和工具调用
  currentMessage?: {
    id: string
    role: 'assistant'
    content: string
  }
  currentToolCall?: {
    id: string
    name: string
    args: string
    result?: any
  }

  // 运行状态
  isRunning: boolean
  currentThreadId?: string
  currentRunId?: string
  currentStep?: string

  // 配置管理
  config: AGUIClientConfig
  updateConfig: (newConfig: Partial<AGUIClientConfig>) => void

  // 会话操作
  createSession: (title?: string) => Promise<AGUISession>
  loadSession: (sessionId: string) => Promise<void>
  deleteSession: (sessionId: string) => Promise<void>
  refreshSessions: () => Promise<void>

  // 聊天操作
  sendMessage: (content: string) => void
  cancelRun: () => void
  clearMessages: () => void

  // 错误处理
  clearError: () => void
}
const AGUIContext = createContext<AGUIContextValue | null>(null)

export interface AGUIProviderProps {
  children: React.ReactNode
  initialConfig?: Partial<AGUIClientConfig>
}

export const AGUIProvider: React.FC<AGUIProviderProps> = ({ children, initialConfig = {} }) => {
  const { t } = useTranslation('pro/agui')

  // 默认配置
  const defaultConfig: AGUIClientConfig = {
    baseUrl: (import.meta as any).env?.VITE_AGUI_BASE_URL || 'http://localhost:3001',
    apiKey: (import.meta as any).env?.VITE_AGUI_API_KEY || '',
    timeout: 30000,
    retryCount: 3,
    ...initialConfig,
  }

  // 基础状态
  const [config, setConfig] = useState<AGUIClientConfig>(defaultConfig)
  const [client, setClient] = useState<AGUIClient>(() => new AGUIClient(config))
  const [isConnected] = useState(true) // 简化实现，假设总是连接的
  const [error, setError] = useState<string | null>(null)
  const [sessions, setSessions] = useState<AGUISession[]>([])
  const [currentSession, setCurrentSession] = useState<AGUISession | null>(null)

  // 聊天状态
  const [messages, setMessages] = useState<ExtendedAGUIMessage[]>([])
  const [loading, setLoading] = useState(false)
  const [state, setState] = useState<Record<string, any>>({})
  const [currentMessage, setCurrentMessage] = useState<
    | {
        id: string
        role: 'assistant'
        content: string
      }
    | undefined
  >(undefined)
  const [currentToolCall, setCurrentToolCall] = useState<
    | {
        id: string
        name: string
        args: string
        result?: any
      }
    | undefined
  >(undefined)
  const [isRunning, setIsRunning] = useState(false)
  const [currentThreadId, setCurrentThreadId] = useState<string | undefined>(undefined)
  const [currentRunId, setCurrentRunId] = useState<string | undefined>(undefined)
  const [currentStep, setCurrentStep] = useState<string | undefined>(undefined)

  // 事件订阅引用
  const unsubscribeRef = useRef<(() => void) | null>(null)

  // 假设 messages: ExtendedAGUIMessage[] 是 React 状态
  // const [messages, setMessages] = useState<ExtendedAGUIMessage[]>([])

  const onTextMessageStart = useCallback((event: TextMessageStartEvent) => {
    const newMessage: ExtendedAGUIMessage = {
      id: event.messageId,
      role: event.role,
      content: '',
      timestamp: new Date(),
      toolCalls: [],
    }
    setMessages((prevMessages) => [...prevMessages, newMessage])
  }, [])

  const onTextMessageContent = useCallback((event: TextMessageContentEvent) => {
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      const index = updatedMessages.findIndex((m) => m.id === event.messageId)

      if (index !== -1) {
        updatedMessages[index] = {
          ...updatedMessages[index],
          content: updatedMessages[index].content + event.delta,
        }
      }
      return updatedMessages
    })
  }, [])

  const onTextMessageEnd = useCallback(
    (event: TextMessageEndEvent) => {
      setMessages((prevMessages) => {
        const updatedMessages = [...prevMessages]
        const index = updatedMessages.findIndex((m) => m.id === event.messageId)

        if (index !== -1) {
          const finalMessage = {
            ...updatedMessages[index],
            timestamp: new Date(),
          }

          updatedMessages[index] = finalMessage

          // 保存到当前会话
          if (currentSession) {
            client.addMessageToSession(currentSession.id, finalMessage)

            try {
              const stored = localStorage.getItem('agui_sessions') || '[]'
              const allSessions = JSON.parse(stored) as AGUISession[]
              const updatedSessions = allSessions.map((s) =>
                s.id === currentSession.id
                  ? {
                      ...s,
                      messages: [...s.messages, finalMessage],
                      updatedAt: new Date(),
                    }
                  : s,
              )
              localStorage.setItem('agui_sessions', JSON.stringify(updatedSessions))
              setSessions(updatedSessions)
            } catch (e) {
              console.warn(t('session.saveMessageFailed'), e)
            }
          }
        }

        return updatedMessages
      })
    },
    [currentSession, client],
  )

  const onToolCallStart = useCallback((event: ToolCallStartEvent) => {
    console.log('ToolCallStart event:', event)
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      const lastMessage = updatedMessages[updatedMessages.length - 1]

      if (lastMessage && lastMessage.role === 'assistant') {
        const toolCall = {
          id: event.toolCallId,
          name: event.toolCallName,
          args: '',
          status: 'pending' as const,
        }

        lastMessage.toolCalls = lastMessage.toolCalls || []
        lastMessage.toolCalls.push(toolCall)

        console.log(t('toolCall.addedToolCall'), toolCall)
      }

      return updatedMessages
    })
  }, [])

  const onToolCallArgs = useCallback((event: ToolCallArgsEvent) => {
    console.log('ToolCallArgs event:', event)

    setCurrentToolCall((prev) =>
      prev
        ? {
            ...prev,
            args: prev.args + event.delta,
          }
        : undefined,
    )

    // // 更新消息中的工具调用参数
    // setMessages((prevMessages) => {
    //   const updatedMessages = [...prevMessages]
    //   const lastMessage = updatedMessages[updatedMessages.length - 1]

    //   if (lastMessage && lastMessage.toolCalls) {
    //     const toolCall = lastMessage.toolCalls.find((tc) => tc.id === event.toolCallId)
    //     if (toolCall) {
    //       toolCall.args += event.delta
    //       console.log('Updated tool call args:', toolCall.args)
    //     }
    //   }

    //   return updatedMessages
    // })
  }, [])

  const onToolCallEnd = useCallback((event: ToolCallEndEvent) => {
    void event
    // 工具调用结束，等待结果
    // 更新工具调用状态为pending（等待结果）
    // setMessages((prevMessages) => {
    //   const updatedMessages = [...prevMessages]
    //   const lastMessage = updatedMessages[updatedMessages.length - 1]

    //   if (lastMessage && lastMessage.toolCalls) {
    //     const toolCall = lastMessage.toolCalls.find((tc) => tc.id === event.toolCallId)
    //     if (toolCall) {
    //       toolCall.status = 'render'
    //     }
    //   }

    //   return updatedMessages
    // })
  }, [])

  const onToolCallResult = useCallback((event: ToolCallResultEvent) => {
    setCurrentToolCall((prev) =>
      prev
        ? {
            ...prev,
            result: event.content,
          }
        : undefined,
    )

    // 更新消息中的工具调用结果
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      const lastMessage = updatedMessages[updatedMessages.length - 1]

      if (lastMessage && lastMessage.toolCalls) {
        const toolCall = lastMessage.toolCalls.find((tc) => tc.id === event.toolCallId)
        if (toolCall) {
          toolCall.args = event.content
          toolCall.status = 'completed'
        }
      }

      return updatedMessages
    })
  }, [])

  // 运行状态事件处理器
  const onRunStarted = useCallback((event: RunStartedEvent) => {
    console.log(t('run.runStarted'), event.runId)
    setIsRunning(true)
    setCurrentRunId(event.runId)
    setLoading(false)
    setError(null) // 清除之前的错误
  }, [])

  const onRunFinished = useCallback((event: RunFinishedEvent) => {
    void event
    setIsRunning(false)
    setCurrentRunId(undefined)
    setCurrentStep(undefined)
    setCurrentMessage(undefined)
    setCurrentToolCall(undefined)
    setLoading(false)

    // 确保所有pending状态的工具调用都标记为completed
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      updatedMessages.forEach((message) => {
        if (message.toolCalls) {
          message.toolCalls.forEach((toolCall) => {
            if (toolCall.status === 'pending') {
              toolCall.status = 'completed'
            }
          })
        }
      })
      const lastMessage = updatedMessages[updatedMessages.length - 1]
      if (lastMessage && lastMessage.role === 'assistant') {
        lastMessage.isFinal = true
      } else {
        updatedMessages.push({
          id: generateMessageId(),
          role: 'assistant',
          content: '',
          timestamp: new Date(),
          isFinal: true,
        })
      }
      return updatedMessages
    })
  }, [])

  const onRunError = useCallback((event: RunErrorEvent) => {
    setError(event.message || t('run.runtimeError'))
    setIsRunning(false)
    setCurrentRunId(undefined)
    setCurrentStep(undefined)
    setCurrentMessage(undefined)
    setCurrentToolCall(undefined)
    setLoading(false)

    // 将pending状态的工具调用标记为error
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      updatedMessages.forEach((message) => {
        if (message.toolCalls) {
          message.toolCalls.forEach((toolCall) => {
            if (toolCall.status !== 'completed') {
              toolCall.status = 'error'
            }
          })
        }
      })
      const lastMessage = updatedMessages[updatedMessages.length - 1]
      if (lastMessage && lastMessage.role === 'assistant') {
        lastMessage.isFinal = true
        lastMessage.errorMessage = event.message
      } else {
        updatedMessages.push({
          id: generateMessageId(),
          role: 'assistant',
          content: '',
          timestamp: new Date(),
          errorMessage: event.message,
          isFinal: true,
        })
      }

      return updatedMessages
    })
  }, [])

  // 步骤状态事件处理器
  const onStepStarted = useCallback((event: StepStartedEvent) => {
    setCurrentStep(event.stepName)
  }, [])

  const onStepFinished = useCallback((event: StepFinishedEvent) => {
    void event
    setCurrentStep(undefined)
  }, [])

  // 状态更新事件处理器
  const onStateSnapshot = useCallback((event: StateSnapshotEvent) => {
    setState(event.snapshot)
  }, [])

  const onStateDelta = useCallback((event: StateDeltaEvent) => {
    setState((prev) => ({ ...prev, ...event.delta }))
  }, [])

  // 消息更新事件处理器
  const onMessagesSnapshot = useCallback(
    (event: MessagesSnapshotEvent) => {
      const aguiMessages: ExtendedAGUIMessage[] = event.messages.map((msg) => ({
        id: msg.id,
        role: msg.role,
        content: typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content),
        timestamp: new Date(msg.timestamp),
        metadata: msg.metadata,
        toolCalls:
          (msg as any).toolCalls?.map((tc: any) => ({
            id: tc.id,
            name: tc.name,
            args: tc.args,
            result: tc.result,
            status: tc.status || 'completed',
          })) || [],
      }))
      setMessages(aguiMessages)

      // 同步到当前会话（如果存在）
      if (currentSession) {
        client.updateSessionState(currentSession.id, { messages: aguiMessages })
      }
    },
    [currentSession, client],
  )

  // Raw事件和自定义事件处理器
  const onRawEvent = useCallback((event: RawEvent) => {
    console.log(t('error.rawEvent'), event.event)
  }, [])

  const onCustomEvent = useCallback((event: CustomEvent) => {
    if (event.name === 'AgentToolDisplay') {
      const toolCallName = event.value.toolCallName
      const content = event.value.content
      const title = event.value.title
      setMessages((prevMessages) => {
        const updatedMessages = [...prevMessages]
        const lastMessage = updatedMessages[updatedMessages.length - 1]

        if (lastMessage && lastMessage.role === 'assistant') {
          const toolCall = {
            toolName: toolCallName,
            title: title,
            content: content,
          }

          lastMessage.agentToolDisplay = lastMessage.agentToolDisplay || []
          lastMessage.agentToolDisplay.push(toolCall)

          console.log(t('toolCall.addedToolCall'), toolCall)
        }

        return updatedMessages
      })
    }
  }, [])

  // 主事件处理器
  const handleEvent = useCallback(
    (event: Event) => {
      console.log('AG-UI Event:', event.type, event)

      switch (event.type) {
        case EventType.TEXT_MESSAGE_START:
          onTextMessageStart(event as TextMessageStartEvent)
          break
        case EventType.TEXT_MESSAGE_CONTENT:
          onTextMessageContent(event as TextMessageContentEvent)
          break
        case EventType.TEXT_MESSAGE_END:
          onTextMessageEnd(event as TextMessageEndEvent)
          break
        case EventType.TOOL_CALL_START:
          onToolCallStart(event as ToolCallStartEvent)
          break
        case EventType.TOOL_CALL_ARGS:
          onToolCallArgs(event as ToolCallArgsEvent)
          break
        case EventType.TOOL_CALL_END:
          onToolCallEnd(event as ToolCallEndEvent)
          break
        case EventType.TOOL_CALL_RESULT:
          onToolCallResult(event as ToolCallResultEvent)
          break
        case EventType.RUN_STARTED:
          onRunStarted(event as RunStartedEvent)
          break
        case EventType.RUN_FINISHED:
          onRunFinished(event as RunFinishedEvent)
          break
        case EventType.RUN_ERROR:
          onRunError(event as RunErrorEvent)
          break
        case EventType.STEP_STARTED:
          onStepStarted(event as StepStartedEvent)
          break
        case EventType.STEP_FINISHED:
          onStepFinished(event as StepFinishedEvent)
          break
        case EventType.STATE_SNAPSHOT:
          onStateSnapshot(event as StateSnapshotEvent)
          break
        case EventType.STATE_DELTA:
          onStateDelta(event as StateDeltaEvent)
          break
        case EventType.MESSAGES_SNAPSHOT:
          onMessagesSnapshot(event as MessagesSnapshotEvent)
          break
        case EventType.RAW:
          onRawEvent(event as RawEvent)
          break
        case EventType.CUSTOM:
          onCustomEvent(event as CustomEvent)
          break
        default:
          console.log(t('error.unknownEventType'), (event as any).type)
      }
    },
    [
      onTextMessageStart,
      onTextMessageContent,
      onTextMessageEnd,
      onToolCallStart,
      onToolCallArgs,
      onToolCallEnd,
      onToolCallResult,
      onRunStarted,
      onRunFinished,
      onRunError,
      onStepStarted,
      onStepFinished,
      onStateSnapshot,
      onStateDelta,
      onMessagesSnapshot,
      onRawEvent,
      onCustomEvent,
    ],
  )

  // 清理订阅
  useEffect(() => {
    return () => {
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
      }
    }
  }, [])

  // 配置更新
  const updateConfig = useCallback(
    (newConfig: Partial<AGUIClientConfig>) => {
      const updatedConfig = { ...config, ...newConfig }
      setConfig(updatedConfig)
      setClient(new AGUIClient(updatedConfig))
    },
    [config],
  )

  // 会话操作
  const createSession = useCallback(async (title?: string): Promise<AGUISession> => {
    const sessionId = `session_${Date.now()}`
    const newSession: AGUISession = {
      id: sessionId,
      title: title || `${t('session.newConversation')} ${new Date().toLocaleString()}`,
      messages: [],
      state: {},
      createdAt: new Date(),
      updatedAt: new Date(),
    }

    setSessions((prev) => [newSession, ...prev])
    setCurrentSession(newSession)
    setMessages([])
    setState({})

    // 为新会话生成threadId
    const threadId = generateThreadId()
    setCurrentThreadId(threadId)

    // 保存到本地存储 (简化实现)
    try {
      const stored = localStorage.getItem('agui_sessions') || '[]'
      const allSessions = JSON.parse(stored)
      allSessions.unshift(newSession)
      localStorage.setItem('agui_sessions', JSON.stringify(allSessions))
    } catch (e) {
      console.warn('保存会话失败:', e)
    }

    return newSession
  }, [])

  const loadSession = useCallback(
    async (sessionId: string) => {
      try {
        // 简化实现：从sessions状态中查找
        const session = sessions.find((s) => s.id === sessionId)
        if (!session) {
          throw new Error(t('session.sessionNotFound'))
        }

        setCurrentSession(session)

        // 确保消息的 timestamp 是 Date 对象，并处理工具调用
        const messagesWithDates = (session.messages || []).map((msg) => ({
          ...msg,
          timestamp: new Date(msg.timestamp),
          toolCalls: (msg as any).toolCalls || [],
        })) as ExtendedAGUIMessage[]
        setMessages(messagesWithDates)
        setState(session.state || {})

        // 为加载的会话生成新的threadId (每次载入会话都是新的对话线程)
        const threadId = generateThreadId()
        setCurrentThreadId(threadId)

        setError(null)
      } catch (error) {
        console.error(t('session.loadSessionFailed'), error)
        setError(t('session.loadSessionFailed'))
      }
    },
    [sessions],
  )

  const deleteSession = useCallback(
    async (sessionId: string) => {
      try {
        setSessions((prev) => prev.filter((s) => s.id !== sessionId))

        if (currentSession?.id === sessionId) {
          setCurrentSession(null)
          setMessages([])
          setState({})
          setCurrentThreadId(undefined)
        }

        // 从本地存储删除
        try {
          const stored = localStorage.getItem('agui_sessions') || '[]'
          const allSessions = JSON.parse(stored)
          const filtered = allSessions.filter((s: AGUISession) => s.id !== sessionId)
          localStorage.setItem('agui_sessions', JSON.stringify(filtered))
        } catch (e) {
          console.warn(t('session.deleteSessionFailed'), e)
        }
      } catch (error) {
        console.error(t('session.deleteSessionFailed'), error)
        setError(t('session.deleteSessionFailed'))
      }
    },
    [currentSession],
  )

  const refreshSessions = useCallback(async () => {
    try {
      // 简化实现：从本地存储加载
      const stored = localStorage.getItem('agui_sessions') || '[]'
      const allSessions = JSON.parse(stored) as AGUISession[]

      // 转换时间字段为 Date 对象，并处理工具调用
      const sessionsWithDates = allSessions.map((session) => ({
        ...session,
        createdAt: new Date(session.createdAt),
        updatedAt: new Date(session.updatedAt),
        messages:
          session.messages?.map((msg) => ({
            ...msg,
            timestamp: new Date(msg.timestamp),
            toolCalls: (msg as any).toolCalls || [],
          })) || [],
      }))

      setSessions(sessionsWithDates)
    } catch (error) {
      console.error(t('session.refreshSessionsFailed'), error)
      setError(t('session.refreshSessionsFailed'))
    }
  }, [])

  // 发送消息
  const sendMessage = useCallback(
    (content: string) => {
      console.log('content', content)
      // if (!currentSession) return

      // 先清理之前的状态，确保没有残留的运行状态
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
      }

      // 重置运行相关状态
      setIsRunning(false)
      setCurrentRunId(undefined)
      setCurrentStep(undefined)
      setCurrentMessage(undefined)
      setCurrentToolCall(undefined)
      setError(null)

      const userMessage: ExtendedAGUIMessage = {
        id: generateMessageId(),
        role: 'user',
        content,
        timestamp: new Date(),
        toolCalls: [], // 用户消息不包含工具调用
      }

      setMessages((prev) => [...prev, userMessage])
      setLoading(true)

      // 保存用户消息到会话 (简化实现)
      if (currentSession) {
        try {
          const stored = localStorage.getItem('agui_sessions') || '[]'
          const allSessions = JSON.parse(stored) as AGUISession[]
          const updatedSessions = allSessions.map((s) =>
            s.id === currentSession.id
              ? {
                  ...s,
                  messages: [...s.messages, userMessage],
                  updatedAt: new Date(),
                }
              : s,
          )
          localStorage.setItem('agui_sessions', JSON.stringify(updatedSessions))
          setSessions(updatedSessions)
        } catch (e) {
          console.warn(t('session.saveMessageFailed'), e)
        }
      }

      // 如果没有threadId，生成一个新的
      let threadId = currentThreadId
      if (!threadId) {
        threadId = generateThreadId()
        setCurrentThreadId(threadId)
      }

      // 生成新的runId
      const runId = generateRunId()
      setCurrentRunId(runId)

      // 订阅新的运行事件
      unsubscribeRef.current = client.subscribeToRun(
        {
          threadId,
          runId,
          state: state,
          messages: [...messages, userMessage],
          tools: [],
          context: [],
          forwardedProps: {
            clientDate: dayjs().format('YYYY-MM-DDTHH:mm:ssZ'),
          },
        },
        handleEvent,
      )
    },
    [currentSession, currentThreadId, messages, state, client, handleEvent],
  )

  // 取消运行
  const cancelRun = useCallback(() => {
    if (currentRunId) {
      // 先重置状态，避免UI卡住
      setIsRunning(false)
      setCurrentRunId(undefined)
      setCurrentStep(undefined)
      setCurrentMessage(undefined)
      setCurrentToolCall(undefined)
      setLoading(false)
      // 将pending状态的工具调用标记为error
      setMessages((prevMessages) => {
        const updatedMessages = [...prevMessages]
        updatedMessages.forEach((message) => {
          if (message.toolCalls) {
            message.toolCalls.forEach((toolCall) => {
              if (toolCall.status !== 'completed') {
                toolCall.status = 'error'
              }
            })
          }
        })
        const lastMessage = updatedMessages[updatedMessages.length - 1]
        if (lastMessage && lastMessage.role === 'assistant') {
          lastMessage.isFinal = true
        } else {
          updatedMessages.push({
            id: generateMessageId(),
            role: 'assistant',
            content: '',
            timestamp: new Date(),
            isFinal: true,
          })
        }
        return updatedMessages
      })
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
      }

      // 然后调用客户端的取消方法
      client
        .cancelRun(currentRunId)
        .then(() => {
          console.log(t('run.runCancelled'))
        })
        .catch((error) => {
          console.error(t('run.cancelRunFailed'), error)
          // 不设置错误状态，因为取消操作本身不应该被视为错误
          // setError(t('run.cancelRunFailed'))
        })
    }
  }, [client, currentRunId])

  // 错误处理
  const clearError = useCallback(() => {
    setError(null)
  }, [])

  // 清空消息
  const clearMessages = useCallback(() => {
    setMessages([])
    setCurrentMessage(undefined)
    setCurrentToolCall(undefined)
    setState({})
  }, [])

  // 初始化
  useEffect(() => {
    refreshSessions()
  }, [refreshSessions])

  const contextValue: AGUIContextValue = {
    client,
    isConnected,
    error,
    sessions,
    currentSession,
    messages,
    loading,
    state,
    currentMessage,
    currentToolCall,
    isRunning,
    currentThreadId,
    currentRunId,
    currentStep,
    config,
    updateConfig,
    createSession,
    loadSession,
    deleteSession,
    refreshSessions,
    sendMessage,
    cancelRun,
    clearMessages,
    clearError,
  }

  return <AGUIContext.Provider value={contextValue}>{children}</AGUIContext.Provider>
}

// Hook for using AGUI chat functionality
export const useAGUIChat = () => {
  const context = useContext(AGUIContext)
  if (!context) {
    throw new Error('useAGUIChat must be used within an AGUIProvider')
  }

  return {
    messages: context.messages,
    loading: context.loading,
    currentMessage: context.currentMessage,
    currentToolCall: context.currentToolCall,
    isRunning: context.isRunning,
    runId: context.currentRunId,
    currentStep: context.currentStep,
    sendMessage: context.sendMessage,
    cancelRun: context.cancelRun,
    clearMessages: context.clearMessages,
  }
}

// Hook for session management
export const useAGUISession = () => {
  const context = useContext(AGUIContext)
  if (!context) {
    throw new Error('useAGUISession must be used within an AGUIProvider')
  }

  return {
    sessions: context.sessions,
    currentSession: context.currentSession,
    createSession: context.createSession,
    loadSession: context.loadSession,
    deleteSession: context.deleteSession,
    refreshSessions: context.refreshSessions,
  }
}

// Hook for accessing state and configuration
export const useAGUIState = () => {
  const context = useContext(AGUIContext)
  if (!context) {
    throw new Error('useAGUIState must be used within an AGUIProvider')
  }

  return {
    client: context.client,
    isConnected: context.isConnected,
    error: context.error,
    state: context.state,
    config: context.config,
    currentThreadId: context.currentThreadId,
    updateConfig: context.updateConfig,
    clearError: context.clearError,
  }
}

export default AGUIContext
