import { get } from 'src/core/utils/request'
import {
  buildConversationHistoryListRequestParams,
  buildConversationMessagesRequestParams,
} from './conversationRequestParams'
import streamFetch from '../util/streamFetch'
import {
  consumeFindingsStreamResponse,
  type FindingsStreamEvent,
  type StreamFindingsDataHandlers,
} from './findingsStream'
import {
  DEFAULT_CONVERSATION_HISTORY_FALLBACK_TITLE,
  normalizeConversationHistoryListResponse,
  type ConversationHistoryListResponse,
} from './conversationHistory'
import { normalizeFindingsDataResponse } from './findingsDataNormalizer'
import type { FindingsReportLink } from './findingsReportLinks'

export { normalizeFindingsStatus } from './findingsPresentation'
export {
  normalizeFindingsDataResponse,
  normalizeFindingsNode,
  normalizeFindingsNodeKind,
} from './findingsDataNormalizer'
export type {
  ConversationHistoryItem,
  ConversationHistoryListResponse,
} from './conversationHistory'
export type { FindingsReportLink } from './findingsReportLinks'

export type FindingsStatus = 'normal' | 'warning' | 'abnormal'

export type FindingsNodeKind = 'summary' | 'evidence' | 'unsupported'

export type FindingsHydrationStatus =
  | 'pending'
  | 'running'
  | 'success'
  | 'failed'
  | 'unsupported'
  | 'skipped'

export type { FindingsStreamEvent, StreamFindingsDataHandlers }

export interface FindingsMetricItem {
  title: string
  unit: string
  timeseries: Array<{
    labels: Record<string, unknown>
    chart: {
      chartData: Record<string, number>
    }
  }>
  detection?: string
}

export interface FindingsMetrics {
  abnormalMetrics: FindingsMetricItem[]
  normalMetrics: FindingsMetricItem[]
}

export interface FindingsLogItem {
  timestamp: string
  level?: string
  message: string
}

export interface FindingsTraceOperation {
  name: string
  count: number
  avgDuration: number
}

export interface FindingsTraceCall {
  target: string
  type: string
  count: number
  avgDuration: number
  maxDuration: number
  errorCount: number
  operations?: FindingsTraceOperation[]
}

export interface FindingsTraceData {
  service: string
  calls: FindingsTraceCall[]
}

export interface FindingsTableHeader {
  name: string
  key: string
  width?: number
}

export interface FindingsTable {
  title?: string
  headers?: FindingsTableHeader[]
  data?: Array<Record<string, unknown>>
  emptyText?: string
}

export interface FindingsNodePayload {
  nodeId: string
  title: string
  nodeKind: FindingsNodeKind
  status: FindingsStatus
  statusColor?: FindingsStatus
  parentId: string | null
  subtitle?: string
  summaryConclusion?: string
  content?: string
  metrics?: FindingsMetrics
  table?: FindingsTable
  logs?: FindingsLogItem[]
  traces?: FindingsTraceData[]
  reportLinks?: FindingsReportLink[]
  unsupportedReason?: string
  hydrationStatus?: FindingsHydrationStatus
  attempt?: number
  retryable?: boolean
  errorType?: string
  errorMessage?: string
}

export interface FindingsDataResponse {
  serviceName?: string
  analysisPeriod?: string
  theme?: string
  timestamp?: string
  nodes: FindingsNodePayload[]
}
// 获取对话历史消息
export function getConversationMessagesApi(threadId: string) {
  return get('/ai/api/v1/conversations/messages', buildConversationMessagesRequestParams(threadId))
}

export async function getConversationHistoryListApi(params?: {
  cursor?: string
  limit?: number
  fallbackTitle?: string
  userId?: string
}): Promise<ConversationHistoryListResponse> {
  const {
    fallbackTitle = DEFAULT_CONVERSATION_HISTORY_FALLBACK_TITLE,
    userId,
    ...queryParams
  } = params || {}
  const response = await get(
    '/ai/api/v1/conversations/list',
    buildConversationHistoryListRequestParams(queryParams, userId),
    {
      timeout: 0,
    },
  )

  return normalizeConversationHistoryListResponse(response, fallbackTitle)
}

// 获取选定服务的 findings 数据
export async function getFindingsDataApi(params: {
  threadId: string
  serviceName?: string
  [key: string]: string | undefined
}): Promise<FindingsDataResponse | null> {
  const queryParams = new URLSearchParams(
    Object.entries(params).reduce<Record<string, string>>((acc, [key, value]) => {
      if (typeof value === 'string') {
        acc[key] = value
      }
      return acc
    }, {}),
  ).toString()
  const response = await get(
    `/ai/api/v1/conversations/findings-data?${queryParams}`,
    {},
    {
      timeout: 0,
    },
  )
  return normalizeFindingsDataResponse(response)
}

export interface StreamFindingsDataParams {
  threadId: string
  serviceName: string
  nodeId?: string
  refresh?: boolean
  signal?: AbortSignal
}

export async function streamFindingsDataApi(
  params: StreamFindingsDataParams,
  handlers: StreamFindingsDataHandlers,
) {
  const { signal, ...query } = params
  const queryParams = new URLSearchParams(
    Object.entries(query).reduce<Record<string, string>>((acc, [key, value]) => {
      if (typeof value === 'string') {
        acc[key] = value
      } else if (typeof value === 'boolean') {
        acc[key] = String(value)
      }
      return acc
    }, {}),
  ).toString()

  const response = await streamFetch(
    `/ai/api/v1/conversations/findings-data/stream?${queryParams}`,
    {
      method: 'GET',
      signal,
      headers: {
        Accept: 'application/x-ndjson',
      },
    },
  )

  await consumeFindingsStreamResponse(response, handlers)
}
