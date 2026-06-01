import React, { createContext, useContext, useState, useCallback, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
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
  AnalysisMessageEndEvent,
  AnalysisMessageContentEvent,
  AnalysisMessageStartEvent,
  AnalysisResultEvent,
  FindingsToolStartEvent,
  FindingsStartEvent,
  WarningEvent,
  FindingsAddEvent,
  FindingsEndEvent,
  FindingsResultEvent,
  FindingsToolArgsEvent,
  FindingsToolDisplayEvent,
  FindingsToolEndEvent,
  FindingsToolResultEvent,
} from '../types/ag-ui'
import dayjs from 'dayjs'
import { checkDataplaneDataApi } from '../api/dataplane'
import { getModelListApi, setDefaultModelApi } from '../api/modelSetting'
import { getCurrentAgentUserId } from '../api/agentUser'
import { getConversationMessagesApi } from '../api/conversation'
import { buildThreadTransitionReset, shouldApplyThreadLoadResult } from './threadTransition'
import {
  classifyThreadLoadError,
  formatThreadLoadErrorMessage,
  type ThreadLoadErrorKind,
} from './threadLoadError'

// 扩展AGUIMessage类型以支持工具调用
export interface ToolCallItem {
  id: string
  name: string
  args: string
  result?: any
  status: 'pending' | 'completed' | 'error'
}
export interface AgentToolDisplayItem {
  toolName: string
  title: string
  content: string
}

// Findings 相关的独立数据结构
export interface FindingsNode {
  id: string
  content: string
  title?: string
  parentId?: string
}

export interface FindingsEdge {
  source: string
  target: string
}

export interface FindingsResult {
  title: string
  content: string
}

// 简化后的 ExtendedAGUIMessage
export interface ExtendedAGUIMessage extends AGUIMessage {
  toolCalls?: ToolCallItem[]
  agentToolDisplay?: AgentToolDisplayItem[]
  isFinal?: boolean
  errorMessage?: string
  hidden?: boolean
  isEnd?: boolean
  type?: 'analysisResult' | 'findings'
  status?: 'pending' | 'completed' | 'error'
  parentId?: string // 添加 parentId 属性
  // 简化的 findingsData，只保留基本结构信息
  findingsData?: {
    nodes: FindingsNode[]
    edges: FindingsEdge[]
    result?: FindingsResult
    tools: Map<string, any[]> // 工具调用数据，按节点ID索引
    displays: Map<string, any[]> // 代理工具显示数据，按节点ID索引
  }
  result?: {
    title: string
    content: string
  }
}

// AI权限相关的类型定义
export interface Model {
  name: string
}

export interface Config {
  name: string
  label: string
  placeholder: string
  required?: boolean
  value?: string
}

export interface Provider {
  name: string
  description: string
  models: Model[]
  config: Config[]
  setUp: boolean
}

export interface ModelOption {
  provider: string
  model: string
  label?: string
  value?: string
}

interface AGUIContextValue {
  // 客户端状态
  client: AGUIClient
  isConnected: boolean
  error: string | null
  threadLoadErrorKind: ThreadLoadErrorKind | null

  // 会话管理
  sessions: AGUISession[]
  currentSession: AGUISession | null

  // 聊天状态
  messages: ExtendedAGUIMessage[]
  loading: boolean
  state: Record<string, any>

  // Findings 相关的独立存储 - 优化版本
  findingsToolCalls: Map<string, Map<string, ToolCallItem>> // 外层 key: `${findingsId}-${nodeId}`, 内层 key: toolCallId
  findingsAgentToolDisplay: Map<string, AgentToolDisplayItem[]> // key: `${findingsId}-${nodeId}`

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

  // AI权限相关状态
  hasPermission: boolean | null
  permissionLoading: boolean
  agentReady: boolean | null
  dataPlaneExist: boolean | null
  allModels: ModelOption[]
  originalProviders: Provider[]
  defaultModel: ModelOption | null

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
  openThread: (threadId: string) => Promise<void>
  resetToNewConversation: () => Promise<void>
  loadHistoryMessages: (threadId: string) => Promise<void>

  // Findings 数据管理 - 优化版本
  getFindingsToolCalls: (findingsId: string, nodeId: string) => ToolCallItem[]
  getFindingsAgentToolDisplay: (findingsId: string, nodeId: string) => AgentToolDisplayItem[]
  setFindingsToolCalls: (findingsId: string, nodeId: string, toolCalls: ToolCallItem[]) => void
  setFindingsAgentToolDisplay: (
    findingsId: string,
    nodeId: string,
    agentToolDisplay: AgentToolDisplayItem[],
  ) => void
  addFindingsToolCall: (findingsId: string, nodeId: string, toolCall: ToolCallItem) => void
  updateFindingsToolCall: (
    findingsId: string,
    nodeId: string,
    toolCallId: string,
    updates: Partial<ToolCallItem>,
  ) => void
  addFindingsAgentToolDisplay: (
    findingsId: string,
    nodeId: string,
    agentToolDisplay: AgentToolDisplayItem,
  ) => void
  clearFindingsData: () => void

  // AI权限操作
  checkPermission: () => Promise<void>
  refreshPermission: () => Promise<void>
  refreshModelList: () => Promise<void>
  changeDefaultModel: ({ provider, model }: { provider: string; model: string }) => void

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
  const [threadLoadErrorKind, setThreadLoadErrorKind] = useState<ThreadLoadErrorKind | null>(null)
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

  // AI权限相关状态
  const [hasPermission, setHasPermission] = useState<boolean | null>(null)
  const [permissionLoading, setPermissionLoading] = useState(true)
  const [agentReady, setAgentReady] = useState<boolean | null>(null)
  const [dataPlaneExist, setDataPlaneExist] = useState<boolean | null>(null)
  const [allModels, setAllModels] = useState<ModelOption[]>([])
  const [originalProviders, setOriginalProviders] = useState<Provider[]>([])
  const [defaultModel, setDefaultModel] = useState<ModelOption | null>(null)
  const { dataGroupId } = useSelector((state: any) => state.dataGroupReducer)

  // Findings 相关的独立存储状态 - 使用嵌套 Map 提高性能
  // 外层 key: `${findingsId}-${nodeId}`, 内层 key: toolCallId
  const [findingsToolCalls, setFindingsToolCalls] = useState<
    Map<string, Map<string, ToolCallItem>>
  >(new Map())
  const [findingsAgentToolDisplay, setFindingsAgentToolDisplay] = useState<
    Map<string, AgentToolDisplayItem[]>
  >(new Map())

  // 事件订阅引用
  const unsubscribeRef = useRef<(() => void) | null>(null)
  const currentRunIdRef = useRef<string | undefined>(undefined)
  const clientRef = useRef(client)
  const threadLoadRevisionRef = useRef(0)

  useEffect(() => {
    currentRunIdRef.current = currentRunId
  }, [currentRunId])

  useEffect(() => {
    clientRef.current = client
  }, [client])

  // Findings 数据管理辅助函数 - 优化版本
  const getFindingsToolCalls = useCallback(
    (findingsId: string, nodeId: string): ToolCallItem[] => {
      const key = `${findingsId}-${nodeId}`
      const toolCallsMap = findingsToolCalls.get(key)
      return toolCallsMap ? Array.from(toolCallsMap.values()) : []
    },
    [findingsToolCalls],
  )

  const getFindingsAgentToolDisplay = useCallback(
    (findingsId: string, nodeId: string): AgentToolDisplayItem[] => {
      const key = `${findingsId}-${nodeId}`
      return findingsAgentToolDisplay.get(key) || []
    },
    [findingsAgentToolDisplay],
  )

  const setFindingsToolCallsData = useCallback(
    (findingsId: string, nodeId: string, toolCalls: ToolCallItem[]) => {
      const key = `${findingsId}-${nodeId}`
      setFindingsToolCalls((prev) => {
        const newMap = new Map(prev)
        const toolCallsMap = new Map<string, ToolCallItem>()
        toolCalls.forEach((toolCall) => {
          toolCallsMap.set(toolCall.id, toolCall)
        })
        newMap.set(key, toolCallsMap)
        return newMap
      })
    },
    [],
  )

  // 添加单个工具调用的高效方法
  const addFindingsToolCall = useCallback(
    (findingsId: string, nodeId: string, toolCall: ToolCallItem) => {
      const key = `${findingsId}-${nodeId}`
      setFindingsToolCalls((prev) => {
        const newMap = new Map(prev)
        const existingToolCalls = newMap.get(key) || new Map<string, ToolCallItem>()
        const updatedToolCalls = new Map(existingToolCalls)
        updatedToolCalls.set(toolCall.id, toolCall)
        newMap.set(key, updatedToolCalls)
        console.log('addFindingsToolCall', newMap)
        return newMap
      })
    },
    [],
  )

  // 更新单个工具调用的高效方法
  const updateFindingsToolCall = useCallback(
    (findingsId: string, nodeId: string, toolCallId: string, updates: Partial<ToolCallItem>) => {
      const key = `${findingsId}-${nodeId}`
      setFindingsToolCalls((prev) => {
        const newMap = new Map(prev)
        const existingToolCalls = newMap.get(key)
        if (existingToolCalls && existingToolCalls.has(toolCallId)) {
          const updatedToolCalls = new Map(existingToolCalls)
          const existingToolCall = updatedToolCalls.get(toolCallId)!

          // 对于 args 字段，进行叠加而不是覆盖
          const mergedUpdates = { ...updates }
          if (updates.args !== undefined) {
            mergedUpdates.args = (existingToolCall.args || '') + updates.args
          }

          updatedToolCalls.set(toolCallId, { ...existingToolCall, ...mergedUpdates })
          newMap.set(key, updatedToolCalls)
          console.log('updateFindingsToolCall', newMap)
          return newMap
        }
        return prev // 如果没找到，返回原状态
      })
    },
    [],
  )

  const setFindingsAgentToolDisplayData = useCallback(
    (findingsId: string, nodeId: string, agentToolDisplay: AgentToolDisplayItem[]) => {
      const key = `${findingsId}-${nodeId}`
      setFindingsAgentToolDisplay((prev) => {
        const newMap = new Map(prev)
        newMap.set(key, agentToolDisplay)
        return newMap
      })
    },
    [],
  )

  // 添加单个 AgentToolDisplay 的高效方法
  const addFindingsAgentToolDisplay = useCallback(
    (findingsId: string, nodeId: string, agentToolDisplay: AgentToolDisplayItem) => {
      const key = `${findingsId}-${nodeId}`
      setFindingsAgentToolDisplay((prev) => {
        const newMap = new Map(prev)
        const existingItems = newMap.get(key) || []
        // 直接 push 到现有数组，避免展开操作
        const updatedItems = [...existingItems, agentToolDisplay]
        newMap.set(key, updatedItems)
        console.log('addFindingsAgentToolDisplay', newMap)
        return newMap
      })
    },
    [],
  )

  const clearFindingsData = useCallback(() => {
    setFindingsToolCalls(new Map())
    setFindingsAgentToolDisplay(new Map())
  }, [])

  const onTextMessageStart = useCallback((event: TextMessageStartEvent) => {
    const newMessage: ExtendedAGUIMessage = {
      id: event.messageId,
      role: event.role,
      content: '',
      timestamp: new Date(),
      toolCalls: [],
    }
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      if (
        updatedMessages?.length > 0 &&
        updatedMessages[updatedMessages.length - 1]?.status === 'pending'
      ) {
        updatedMessages[updatedMessages.length - 1].status = 'completed'
      }
      updatedMessages.push(newMessage)
      return updatedMessages
    })
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
      }

      return updatedMessages
    })
  }, [])

  const onToolCallArgs = useCallback((event: ToolCallArgsEvent) => {
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
    setIsRunning(true)
    setCurrentRunId(event.runId)
    setLoading(false)
    setError(null) // 清除之前的错误
    setThreadLoadErrorKind(null)
  }, [])

  const onRunFinished = useCallback(
    (event: RunFinishedEvent) => {
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
          if (message?.status === 'pending') {
            message.status = 'completed'
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
        console.log('updatedMessages', updatedMessages)
        console.log('findingsToolCalls', findingsToolCalls)
        console.log('findingsAgentToolDisplay', findingsAgentToolDisplay)

        return updatedMessages
      })
    },
    [findingsToolCalls, findingsAgentToolDisplay],
  )

  const onRunError = useCallback((event: RunErrorEvent) => {
    setError(event.message || t('run.runtimeError'))
    setThreadLoadErrorKind(null)
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
        if (message?.status === 'pending') {
          message.status = 'completed'
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

  const onAnalysisMessageStart = useCallback((event: AnalysisMessageStartEvent) => {
    const newMessage: ExtendedAGUIMessage = {
      id: event.value.id,
      role: 'assistant',
      content: '',
      timestamp: new Date(),
      toolCalls: [],
      agentToolDisplay: [],
      hidden: true,
      status: 'pending',
    }
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      if (
        updatedMessages?.length > 0 &&
        updatedMessages[updatedMessages.length - 1]?.status === 'pending'
      ) {
        updatedMessages[updatedMessages.length - 1].status = 'completed'
      }
      updatedMessages.push(newMessage)
      return updatedMessages
    })
  }, [])

  const onAnalysisMessageContent = useCallback((event: AnalysisMessageContentEvent) => {
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      const index = updatedMessages.findIndex((m) => m.id === event.value.id)

      if (index !== -1) {
        updatedMessages[index] = {
          ...updatedMessages[index],
          content: updatedMessages[index].content + event.value.delta,
        }
      }
      return updatedMessages
    })
  }, [])

  const onAnalysisMessageEnd = useCallback((event: AnalysisMessageEndEvent) => {
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      const index = updatedMessages.findIndex((m) => m.id === event.value.id)

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
  }, [])
  const onAnalysisResult = useCallback((event: AnalysisResultEvent) => {
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      const index = updatedMessages.findIndex((m) => m.id === event.value.parentId)
      const newMessage: ExtendedAGUIMessage = {
        id: event.value.id,
        role: 'assistant',
        content: event.value.content,
        timestamp: new Date(),
        toolCalls: [],
        agentToolDisplay: [],
        type: 'analysisResult',
      }
      if (index !== -1) {
        updatedMessages.splice(index + 1, 0, newMessage)
      } else {
        updatedMessages.push(newMessage)
      }
      return updatedMessages
    })
  }, [])

  const onFindingsRootStart = useCallback((event: FindingsStartEvent) => {
    const value = event.value
    const newMessage: ExtendedAGUIMessage = {
      id: value.findingsId,
      role: 'assistant',
      content: value.content,
      timestamp: new Date(),
      toolCalls: [],
      agentToolDisplay: [],
      hidden: false,
      status: 'pending',
      type: 'findings',
      findingsData: {
        nodes: [
          {
            id: value.nodeId,
            content: value.content,
          },
        ],
        edges: [],
        tools: new Map(),
        displays: new Map(),
      },
    }

    // 不再在节点创建时初始化工具调用和展示数据
    // 这些数据将在收到 FINDINGS_TOOL_START 和 FINDINGS_TOOL_DISPLAY_EVENT 时创建

    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      if (
        updatedMessages?.length > 0 &&
        updatedMessages[updatedMessages.length - 1]?.status === 'pending'
      ) {
        updatedMessages[updatedMessages.length - 1].status = 'completed'
      }
      updatedMessages.push(newMessage)
      return updatedMessages
    })
  }, [])

  const onFindingsNodeAdd = useCallback((event: FindingsAddEvent) => {
    // 不再在节点创建时初始化工具调用和展示数据
    // 这些数据将在收到 FINDINGS_TOOL_START 和 FINDINGS_TOOL_DISPLAY_EVENT 时创建

    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      const index = updatedMessages.findIndex((m) => m.id === event.value.findingsId)
      if (index !== -1) {
        updatedMessages[index].findingsData?.nodes.push({
          id: event.value.nodeId,
          title: event.value.title,
          content: event.value.content,
          parentId: event.value.parentId,
        })
        updatedMessages[index].findingsData?.edges.push({
          source: event.value.parentId,
          target: event.value.nodeId,
        })
      }
      return updatedMessages
    })
  }, [])

  const onFindingsEnd = useCallback((event: FindingsEndEvent) => {
    // setMessages((prevMessages) => {
    //   const updatedMessages = [...prevMessages]
    //   const index = updatedMessages.findIndex((m) => m.id === event.value.findingsId)
    //   if (index !== -1) {
    //     updatedMessages[index].status = 'completed'
    //   }
    //   return updatedMessages
    // })
  }, [])

  const onFindingsRootResult = useCallback((event: FindingsResultEvent) => {
    setMessages((prevMessages) => {
      const updatedMessages = [...prevMessages]
      const index = updatedMessages.findIndex((m) => m.id === event.value.findingsId)
      if (index !== -1 && updatedMessages[index].findingsData) {
        console.log('updatedMessages[index].findingsData', updatedMessages[index].findingsData)
        console.log('event.value', event.value)
        if (!updatedMessages[index].findingsData!.result) {
          updatedMessages[index].findingsData!.result = { title: '', content: '' }
        }
        updatedMessages[index].findingsData!.result!.title = event.value.title
        updatedMessages[index].findingsData!.result!.content = event.value.content
        updatedMessages[index].status = 'completed'
      }
      return updatedMessages
    })
  }, [])

  const onFindingsToolStart = useCallback(
    (event: FindingsToolStartEvent) => {
      const { findingsId, nodeId, toolCallId, toolCallName } = event.value

      // 使用高效的添加方法，直接添加单个工具调用
      const newToolCall: ToolCallItem = {
        id: toolCallId,
        name: toolCallName,
        args: '',
        status: 'pending',
      }

      addFindingsToolCall(findingsId, nodeId, newToolCall)
    },
    [addFindingsToolCall],
  )

  const onFindingsToolArgs = useCallback(
    (event: FindingsToolArgsEvent) => {
      const { findingsId, nodeId, toolCallId, delta } = event.value

      // 使用高效的更新方法，只更新特定的工具调用
      // 先获取当前的 args，然后追加 delta
      const key = `${findingsId}-${nodeId}`
      const existingToolCalls = findingsToolCalls.get(key)
      const existingToolCall = existingToolCalls?.get(toolCallId)

      if (existingToolCall) {
        updateFindingsToolCall(findingsId, nodeId, toolCallId, {
          args: existingToolCall.args + delta,
        })
      }
    },
    [findingsToolCalls, updateFindingsToolCall],
  )

  const onFindingsToolEnd = useCallback((event: FindingsToolEndEvent) => {
    void event
  }, [])

  const onFindingsToolResult = useCallback((event: FindingsToolResultEvent) => {
    // 根据实际事件结构调整
    console.log('FindingsToolResult event:', event)
    // 暂时注释掉具体实现，等确认事件结构
    const { findingsId, nodeId, toolCallId, content } = event.value
    updateFindingsToolCall(findingsId, nodeId, toolCallId, { status: 'completed', args: content })
  }, [])

  const onFindingsToolDisplayEvent = useCallback(
    (event: FindingsToolDisplayEvent) => {
      // 检查事件值的类型，确保正确处理
      if (event.value && typeof event.value === 'object') {
        const eventValue = event.value as any
        const { findingsId, nodeId, toolCallName, title, content } = eventValue
        console.log('onFindingsToolDisplayEvent', eventValue)
        if (findingsId && nodeId && toolCallName && title && content) {
          // 使用高效的添加方法，直接添加单个展示项
          const newAgentToolDisplay: AgentToolDisplayItem = {
            toolName: toolCallName,
            title,
            content,
          }

          addFindingsAgentToolDisplay(findingsId, nodeId, newAgentToolDisplay)
        }
      }
    },
    [addFindingsAgentToolDisplay],
  )

  const onWarning = useCallback((event: WarningEvent) => {
    void event
  }, [])

  const onCustomEvent = useCallback((event: CustomEvent) => {
    switch (event.name) {
      case 'AgentToolDisplay':
        if (event.value.toolCallName !== 'PING') {
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
            }

            return updatedMessages
          })
        }
        break

      case 'ANALYSIS_MESSAGE_START':
        onAnalysisMessageStart(event as AnalysisMessageStartEvent)
        break

      case 'ANALYSIS_MESSAGE_CONTENT':
        onAnalysisMessageContent(event as AnalysisMessageContentEvent)
        break

      case 'ANALYSIS_MESSAGE_END':
        onAnalysisMessageEnd(event as AnalysisMessageEndEvent)
        break

      case 'ANALYSIS_RESULT':
        onAnalysisResult(event as AnalysisResultEvent)
        break

      case 'FINDINGS_ROOT_START':
        onFindingsRootStart(event as FindingsStartEvent)
        break

      case 'FINDINGS_NODE_ADD':
        onFindingsNodeAdd(event as FindingsAddEvent)
        break

      case 'FINDINGS_TREE_END':
        onFindingsEnd(event as FindingsEndEvent)
        break

      case 'FINDINGS_TREE_RESULT':
        onFindingsRootResult(event as FindingsResultEvent)
        break

      case 'FINDINGS_TOOL_START':
        onFindingsToolStart(event as FindingsToolStartEvent)
        break

      case 'FINDINGS_TOOL_ARGS':
        onFindingsToolArgs(event as FindingsToolArgsEvent)
        break

      case 'FINDINGS_TOOL_END':
        onFindingsToolEnd(event as FindingsToolEndEvent)
        break

      case 'FINDINGS_TOOL_RESULT':
        onFindingsToolResult(event as FindingsToolResultEvent)
        break

      case 'FINDINGS_TOOL_DISPLAY':
        onFindingsToolDisplayEvent(event as FindingsToolDisplayEvent)
        break

      case 'WARNING':
        onWarning(event as WarningEvent)
        break

      default:
        break
    }
  }, [])

  // 主事件处理器
  const handleEvent = useCallback(
    (event: Event) => {
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

  const applyThreadTransitionReset = useCallback(
    (activeRunId: string | undefined, targetThreadId?: string) => {
      const reset = buildThreadTransitionReset<
        ExtendedAGUIMessage,
        ToolCallItem,
        AgentToolDisplayItem
      >({
        currentRunId: activeRunId,
        loadRevision: threadLoadRevisionRef.current,
        targetThreadId,
      })

      threadLoadRevisionRef.current = reset.nextLoadRevision

      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
      }

      setIsRunning(reset.runtime.isRunning)
      setCurrentRunId(reset.runtime.currentRunId)
      setCurrentStep(reset.runtime.currentStep)
      setCurrentMessage(reset.runtime.currentMessage)
      setCurrentToolCall(reset.runtime.currentToolCall)
      setLoading(reset.runtime.loading)
      setError(reset.runtime.error)
      setThreadLoadErrorKind(null)
      setMessages(reset.scoped.messages)
      setState(reset.scoped.state)
      setCurrentThreadId(undefined)
      setFindingsToolCalls(reset.scoped.findingsToolCalls)
      setFindingsAgentToolDisplay(reset.scoped.findingsAgentToolDisplay)

      return reset
    },
    [],
  )

  const transitionToThread = useCallback(
    async (targetThreadId?: string) => {
      const activeRunId = currentRunIdRef.current
      const reset = applyThreadTransitionReset(activeRunId, targetThreadId)

      if (reset.shouldCancelRun && activeRunId) {
        clientRef.current.cancelRun(activeRunId).catch((transitionError) => {
          console.error(t('run.cancelRunFailed'), transitionError)
        })
      }

      if (!reset.shouldLoadHistory || !reset.targetThreadId) {
        return
      }

      const responseLoadRevision = reset.nextLoadRevision

      try {
        const response = await getConversationMessagesApi(reset.targetThreadId)

        if (
          !shouldApplyThreadLoadResult({
            activeLoadRevision: threadLoadRevisionRef.current,
            responseLoadRevision,
          })
        ) {
          return
        }

        const rawMessages: ExtendedAGUIMessage[] = Array.isArray(response?.messages)
          ? response.messages.map((msg: any) => ({
              id: msg.id,
              role: msg.role,
              content: msg.content,
              timestamp: new Date(msg.timestamp),
              toolCalls:
                msg.toolCalls?.map((tc: any) => ({
                  id: tc.id || generateMessageId(),
                  name: tc.name,
                  args: tc.args,
                  result: tc.result,
                  status: 'completed' as const,
                })) || [],
              agentToolDisplay: msg.agentToolDisplay || [],
              hidden: msg.hidden,
              parentId: msg.parentId,
              findingsData: msg.findingsData,
              type: msg.type,
              status: 'completed',
            }))
          : []

        const historyMessages: ExtendedAGUIMessage[] = []
        const messageMap = new Map<string, ExtendedAGUIMessage>()

        rawMessages.forEach((message) => {
          messageMap.set(message.id, message)
        })

        rawMessages.forEach((message) => {
          if (message.parentId && !message.hidden) {
            const parentIndex = historyMessages.findIndex((historyMessage) => {
              const parent = messageMap.get(message.parentId!)
              return historyMessage.id === (parent?.id || message.parentId)
            })

            if (parentIndex !== -1) {
              historyMessages.splice(parentIndex + 1, 0, message)
              return
            }
          }

          historyMessages.push(message)
        })

        if (historyMessages.length > 0) {
          historyMessages[historyMessages.length - 1].isFinal = true
        }

        setMessages(historyMessages)
        setCurrentThreadId(reset.targetThreadId)
        setThreadLoadErrorKind(null)
      } catch (error) {
        if (
          !shouldApplyThreadLoadResult({
            activeLoadRevision: threadLoadRevisionRef.current,
            responseLoadRevision,
          })
        ) {
          return
        }

        console.error('加载历史消息失败:', error)
        setThreadLoadErrorKind(classifyThreadLoadError(error))
        setError(
          formatThreadLoadErrorMessage({
            error,
            messages: {
              loadFailed: t('thread.loadHistoryFailed'),
              unknownError: t('thread.unknownHistoryError'),
            },
          }),
        )
      } finally {
        if (
          shouldApplyThreadLoadResult({
            activeLoadRevision: threadLoadRevisionRef.current,
            responseLoadRevision,
          })
        ) {
          setLoading(false)
        }
      }
    },
    [applyThreadTransitionReset, t],
  )

  // 发送消息
  const sendMessage = useCallback(
    (content: string) => {
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
      setThreadLoadErrorKind(null)

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
      const userId = getCurrentAgentUserId()

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
            groupId: dataGroupId ?? 0,
            ...(userId ? { userId } : {}),
          },
        },
        handleEvent,
      )
    },
    [currentSession, currentThreadId, messages, state, client, handleEvent, dataGroupId],
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
    setThreadLoadErrorKind(null)
  }, [])

  const openThread = useCallback(
    async (threadId: string) => {
      await transitionToThread(threadId)
    },
    [transitionToThread],
  )

  const resetToNewConversation = useCallback(async () => {
    await transitionToThread(undefined)
  }, [transitionToThread])

  // 清空消息
  const clearMessages = useCallback(() => {
    void resetToNewConversation()
  }, [resetToNewConversation])

  // 加载历史消息
  const loadHistoryMessages = useCallback(
    async (threadId: string) => {
      await openThread(threadId)
    },
    [openThread],
  )

  // AI权限相关方法
  const fetchModelList = useCallback(async () => {
    try {
      const res = await getModelListApi()
      const providers: Provider[] = res?.data?.providers || []
      setOriginalProviders(providers)

      // 可能后端的 default 不是完整的 ModelOption，这里做一次兜底组装
      const defaultFromApi = res?.data?.default || null
      setDefaultModel(defaultFromApi)

      const models: ModelOption[] = []
      providers.forEach((provider: Provider) => {
        if (provider.name === 'openai-api-compatible') {
          provider.models.forEach((model: Model) => {
            models.push({
              provider: provider.name,
              model: model.name,
              label: `${model.name}`,
              value: `${provider.name}-${model.name}`,
            })
          })
        } else {
          if (provider.setUp) {
            provider.models.forEach((model: Model) => {
              models.push({
                provider: provider.name,
                model: model.name,
                label: `${model.name}`,
                value: `${provider.name}-${model.name}`,
              })
            })
          }
        }
      })
      setAllModels(models)
    } catch (error) {
      console.error('获取模型列表失败:', error)
    }
  }, [])

  const checkPermission = useCallback(async () => {
    setPermissionLoading(true)
    try {
      // 检查数据平面状态
      const dataplaneData = await checkDataplaneDataApi()
      const { agentReady: ready, dataPlaneExist: exist } = dataplaneData.data || dataplaneData
      setAgentReady(ready)
      setDataPlaneExist(exist)

      // 根据数据平面状态决定权限
      if (!ready) {
        // agentReady为false，提示未安装minimal
        setError(t('agentNotReady', 'Agent未就绪，请先安装minimal'))
        setHasPermission(false)
      } else if (!exist) {
        // 只有dataPlaneExist为false，可以正常使用agent，仅提示
        setHasPermission(true)
        setError(t('dataPlaneNotExist', '数据平面不存在，但可以正常使用Agent'))
      } else {
        // 都正常
        setHasPermission(true)
      }
    } catch (err) {
      console.error('数据平面状态检查失败:', err)
      setError(t('dataplaneCheckError', '数据平面状态检查失败，请稍后重试'))
      setHasPermission(false)
    } finally {
      setPermissionLoading(false)
    }
  }, [t])

  const refreshPermission = useCallback(async () => {
    await checkPermission()
  }, [checkPermission])

  const refreshModelList = useCallback(async () => {
    await fetchModelList()
  }, [fetchModelList])

  const changeDefaultModel = useCallback(
    ({ provider, model }: { provider: string; model: string }) => {
      setDefaultModelApi({ provider, model }).then(() => {
        fetchModelList()
      })
    },
    [fetchModelList],
  )

  // 初始化
  useEffect(() => {
    refreshSessions()
    checkPermission()
    fetchModelList()
  }, [refreshSessions, checkPermission, fetchModelList])

  const contextValue: AGUIContextValue = {
    client,
    isConnected,
    error,
    threadLoadErrorKind,
    sessions,
    currentSession,
    messages,
    loading,
    state,
    findingsToolCalls,
    findingsAgentToolDisplay,
    currentMessage,
    currentToolCall,
    isRunning,
    currentThreadId,
    currentRunId,
    currentStep,
    hasPermission,
    permissionLoading,
    agentReady,
    dataPlaneExist,
    allModels,
    originalProviders,
    defaultModel,
    config,
    updateConfig,
    createSession,
    loadSession,
    deleteSession,
    refreshSessions,
    sendMessage,
    cancelRun,
    clearMessages,
    openThread,
    resetToNewConversation,
    loadHistoryMessages,
    getFindingsToolCalls,
    getFindingsAgentToolDisplay,
    setFindingsToolCalls: setFindingsToolCallsData,
    setFindingsAgentToolDisplay: setFindingsAgentToolDisplayData,
    addFindingsToolCall,
    updateFindingsToolCall,
    addFindingsAgentToolDisplay,
    clearFindingsData,
    checkPermission,
    refreshPermission,
    refreshModelList,
    changeDefaultModel,
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
    openThread: context.openThread,
    resetToNewConversation: context.resetToNewConversation,
    loadHistoryMessages: context.loadHistoryMessages,
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
    threadLoadErrorKind: context.threadLoadErrorKind,
    state: context.state,
    config: context.config,
    currentThreadId: context.currentThreadId,
    updateConfig: context.updateConfig,
    clearError: context.clearError,
  }
}

// Hook for AI permission management
export const useAGUIPermission = () => {
  const context = useContext(AGUIContext)
  if (!context) {
    throw new Error('useAGUIPermission must be used within an AGUIProvider')
  }

  return {
    hasPermission: context.hasPermission,
    permissionLoading: context.permissionLoading,
    agentReady: context.agentReady,
    dataPlaneExist: context.dataPlaneExist,
    allModels: context.allModels,
    originalProviders: context.originalProviders,
    defaultModel: context.defaultModel,
    checkPermission: context.checkPermission,
    refreshPermission: context.refreshPermission,
    refreshModelList: context.refreshModelList,
    changeDefaultModel: context.changeDefaultModel,
  }
}

// Hook for findings data management
export const useAGUIFindings = () => {
  const context = useContext(AGUIContext)
  if (!context) {
    throw new Error('useAGUIFindings must be used within an AGUIProvider')
  }

  return {
    getFindingsToolCalls: context.getFindingsToolCalls,
    getFindingsAgentToolDisplay: context.getFindingsAgentToolDisplay,
    setFindingsToolCalls: context.setFindingsToolCalls,
    setFindingsAgentToolDisplay: context.setFindingsAgentToolDisplay,
    addFindingsToolCall: context.addFindingsToolCall,
    updateFindingsToolCall: context.updateFindingsToolCall,
    addFindingsAgentToolDisplay: context.addFindingsAgentToolDisplay,
    clearFindingsData: context.clearFindingsData,
  }
}

export default AGUIContext
