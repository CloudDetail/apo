import { get } from 'src/core/utils/request'
import { stabilizeFindingsNodeStatus } from './findingsPresentation'

export { normalizeFindingsStatus } from './findingsPresentation'

export type FindingsStatus = 'normal' | 'warning' | 'abnormal'

export type FindingsNodeKind = 'summary' | 'evidence' | 'unsupported'

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
  unsupportedReason?: string
}

export interface FindingsDataResponse {
  serviceName?: string
  analysisPeriod?: string
  theme?: string
  timestamp?: string
  nodes: FindingsNodePayload[]
}

const isObject = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === 'object' && !Array.isArray(value)

const readString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

const hasMetrics = (metrics?: FindingsMetrics) =>
  Boolean(
    metrics &&
      ((metrics.abnormalMetrics?.length ?? 0) > 0 || (metrics.normalMetrics?.length ?? 0) > 0),
  )

const hasTable = (table?: FindingsTable) =>
  Boolean(table && ((table.headers?.length ?? 0) > 0 || (table.data?.length ?? 0) > 0))

const hasLogs = (logs?: FindingsLogItem[]) => (logs?.length ?? 0) > 0

const hasTraces = (traces?: FindingsTraceData[]) => (traces?.length ?? 0) > 0

const normalizeMetrics = (value: unknown): FindingsMetrics | undefined => {
  if (!isObject(value)) {
    return undefined
  }

  const abnormalMetrics = Array.isArray(value.abnormalMetrics)
    ? (value.abnormalMetrics as FindingsMetricItem[])
    : []
  const normalMetrics = Array.isArray(value.normalMetrics)
    ? (value.normalMetrics as FindingsMetricItem[])
    : []

  if (abnormalMetrics.length === 0 && normalMetrics.length === 0) {
    return undefined
  }

  return {
    abnormalMetrics,
    normalMetrics,
  }
}

const normalizeTable = (value: unknown): FindingsTable | undefined => {
  if (!isObject(value)) {
    return undefined
  }

  const headers = Array.isArray(value.headers) ? (value.headers as FindingsTableHeader[]) : []
  const data = Array.isArray(value.data) ? (value.data as Array<Record<string, unknown>>) : []

  if (headers.length === 0 && data.length === 0) {
    return undefined
  }

  return {
    title: readString(value.title),
    headers,
    data,
    emptyText: readString(value.emptyText),
  }
}

const normalizeLogs = (value: unknown): FindingsLogItem[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined
  }

  const logs = value
    .filter(isObject)
    .map((item) => {
      const message = readString(item.message)
      const timestamp = readString(item.timestamp)

      if (!message || !timestamp) {
        return null
      }

      return {
        timestamp,
        level: readString(item.level),
        message,
      }
    })
    .filter((item) => item !== null) as FindingsLogItem[]

  return logs.length > 0 ? logs : undefined
}

const normalizeTraces = (value: unknown): FindingsTraceData[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined
  }

  const traces = value
    .filter(isObject)
    .map((trace) => {
      const service = readString(trace.service)
      const calls = Array.isArray(trace.calls)
        ? trace.calls.filter(isObject).map((call) => ({
            target: readString(call.target) ?? '',
            type: readString(call.type) ?? '',
            count: typeof call.count === 'number' ? call.count : 0,
            avgDuration: typeof call.avgDuration === 'number' ? call.avgDuration : 0,
            maxDuration: typeof call.maxDuration === 'number' ? call.maxDuration : 0,
            errorCount: typeof call.errorCount === 'number' ? call.errorCount : 0,
            operations: Array.isArray(call.operations)
              ? call.operations.filter(isObject).map((operation) => ({
                  name: readString(operation.name) ?? '',
                  count: typeof operation.count === 'number' ? operation.count : 0,
                  avgDuration:
                    typeof operation.avgDuration === 'number' ? operation.avgDuration : 0,
                }))
              : undefined,
          }))
        : []

      if (!service || calls.length === 0) {
        return null
      }

      return {
        service,
        calls,
      }
    })
    .filter((trace) => trace !== null) as FindingsTraceData[]

  return traces.length > 0 ? traces : undefined
}

export const normalizeFindingsNodeKind = (
  value: unknown,
  fallback: FindingsNodeKind = 'evidence',
): FindingsNodeKind => {
  const normalized = readString(value)?.toLowerCase()
  if (normalized === 'summary' || normalized === 'evidence' || normalized === 'unsupported') {
    return normalized
  }

  return fallback
}

export const normalizeFindingsNode = (value: unknown): FindingsNodePayload | null => {
  if (!isObject(value)) {
    return null
  }

  const nodeId = readString(value.nodeId) ?? readString(value.id)
  if (!nodeId) {
    return null
  }

  const metrics = normalizeMetrics(value.metrics)
  const table = normalizeTable(value.table)
  const logs = normalizeLogs(value.logs)
  const traces = normalizeTraces(value.traces)
  const unsupportedReason = readString(value.unsupportedReason)
  const inferredKind =
    unsupportedReason || (!hasMetrics(metrics) && !hasTable(table) && !hasLogs(logs) && !hasTraces(traces))
      ? unsupportedReason
        ? 'unsupported'
        : 'summary'
      : 'evidence'

  const nodeKind = normalizeFindingsNodeKind(value.nodeKind, inferredKind)
  const status = stabilizeFindingsNodeStatus({
    nodeKind,
    status: value.status,
    statusColor: value.statusColor,
  })

  return {
    nodeId,
    parentId: readString(value.parentId) ?? null,
    title: readString(value.title) ?? nodeId,
    subtitle: readString(value.subtitle),
    nodeKind,
    status,
    statusColor: status,
    summaryConclusion: readString(value.summaryConclusion),
    content: readString(value.content),
    metrics,
    table,
    logs,
    traces,
    unsupportedReason,
  }
}

export const normalizeFindingsDataResponse = (value: unknown): FindingsDataResponse | null => {
  if (!isObject(value) || !Array.isArray(value.nodes)) {
    return null
  }

  const nodes = value.nodes
    .map((node) => normalizeFindingsNode(node))
    .filter((node): node is FindingsNodePayload => node !== null)

  return {
    serviceName: readString(value.serviceName),
    analysisPeriod: readString(value.analysisPeriod),
    theme: readString(value.theme),
    timestamp: readString(value.timestamp),
    nodes,
  }
}

// 获取对话历史消息
export function getConversationMessagesApi(threadId: string) {
  return get(`/ai/api/v1/conversations/messages?threadId=${threadId}`)
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
  const response = await get(`/ai/api/v1/conversations/findings-data?${queryParams}`)
  return normalizeFindingsDataResponse(response)
}
