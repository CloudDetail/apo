import { stabilizeFindingsNodeStatus } from './findingsPresentation'
import { normalizeFindingsReportLinks } from './findingsReportLinks'
import type {
  FindingsDataResponse,
  FindingsHydrationStatus,
  FindingsLogItem,
  FindingsMetricItem,
  FindingsMetrics,
  FindingsNodeKind,
  FindingsNodePayload,
  FindingsTable,
  FindingsTableHeader,
  FindingsTraceCall,
  FindingsTraceData,
  FindingsTraceOperation,
} from './conversation'

const isObject = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === 'object' && !Array.isArray(value)

const readString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

const readNumber = (value: unknown): number | undefined => {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    return undefined
  }

  return value
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

const normalizeTraceOperations = (value: unknown): FindingsTraceOperation[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined
  }

  const operations = value
    .filter(isObject)
    .map((operation) => {
      const name = readString(operation.name)

      if (!name) {
        return null
      }

      return {
        name,
        count: readNumber(operation.count) ?? 0,
        avgDuration: readNumber(operation.avgDuration) ?? 0,
      }
    })
    .filter((operation) => operation !== null) as FindingsTraceOperation[]

  return operations.length > 0 ? operations : undefined
}

const normalizeTraceCalls = (value: unknown): FindingsTraceCall[] => {
  if (!Array.isArray(value)) {
    return []
  }

  return value.filter(isObject).map((call) => ({
    target: readString(call.target) ?? '',
    type: readString(call.type) ?? '',
    count: readNumber(call.count) ?? 0,
    avgDuration: readNumber(call.avgDuration) ?? 0,
    maxDuration: readNumber(call.maxDuration) ?? 0,
    errorCount: readNumber(call.errorCount) ?? 0,
    operations: normalizeTraceOperations(call.operations),
  }))
}

const getRawTraceCallType = (operationName: string) =>
  /^(select|insert|update|delete|replace|with|show|explain)\b/i.test(operationName)
    ? 'database'
    : 'trace'

interface RawTraceStats {
  count: number
  durationSum: number
  maxDuration: number
  errorCount: number
}

const normalizeRawTraceRows = (rows: Record<string, unknown>[]): FindingsTraceData[] => {
  const serviceGroups = new Map<string, Map<string, RawTraceStats>>()

  rows.forEach((row) => {
    const serviceName = readString(row.serviceName)
    const operationName = readString(row.operationName)

    if (!serviceName || !operationName) {
      return
    }

    const duration = readNumber(row.duration) ?? 0
    let operationGroups = serviceGroups.get(serviceName)
    if (!operationGroups) {
      operationGroups = new Map<string, RawTraceStats>()
      serviceGroups.set(serviceName, operationGroups)
    }

    const stats = operationGroups.get(operationName) ?? {
      count: 0,
      durationSum: 0,
      maxDuration: 0,
      errorCount: 0,
    }

    stats.count += 1
    stats.durationSum += duration
    stats.maxDuration = Math.max(stats.maxDuration, duration)
    stats.errorCount += row.isError === true ? 1 : 0
    operationGroups.set(operationName, stats)
  })

  return Array.from(serviceGroups.entries()).map(([service, operationGroups]) => ({
    service,
    calls: Array.from(operationGroups.entries()).map(([operationName, stats]) => {
      const avgDuration = stats.count > 0 ? stats.durationSum / stats.count : 0

      return {
        target: operationName,
        type: getRawTraceCallType(operationName),
        count: stats.count,
        avgDuration,
        maxDuration: stats.maxDuration,
        errorCount: stats.errorCount,
        operations: [
          {
            name: operationName,
            count: stats.count,
            avgDuration,
          },
        ],
      }
    }),
  }))
}

const normalizeTraces = (value: unknown): FindingsTraceData[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined
  }

  const rawRows: Record<string, unknown>[] = []
  const traces = value
    .filter(isObject)
    .map((trace) => {
      const service = readString(trace.service)
      const calls = normalizeTraceCalls(trace.calls)

      if (service && calls.length > 0) {
        return {
          service,
          calls,
        }
      }

      rawRows.push(trace)
      return null
    })
    .filter((trace) => trace !== null) as FindingsTraceData[]

  const normalizedRawRows = normalizeRawTraceRows(rawRows)
  const normalized = [...traces, ...normalizedRawRows]

  return normalized.length > 0 ? normalized : undefined
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

const normalizeHydrationStatus = (value: unknown): FindingsHydrationStatus | undefined =>
  value === 'pending' ||
  value === 'running' ||
  value === 'success' ||
  value === 'failed' ||
  value === 'unsupported' ||
  value === 'skipped'
    ? value
    : undefined

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
    unsupportedReason ||
    (!hasMetrics(metrics) && !hasTable(table) && !hasLogs(logs) && !hasTraces(traces))
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
    reportLinks: normalizeFindingsReportLinks(value.reportLinks),
    unsupportedReason,
    hydrationStatus: normalizeHydrationStatus(value.hydrationStatus),
    attempt: readNumber(value.attempt),
    retryable: typeof value.retryable === 'boolean' ? value.retryable : undefined,
    errorType: readString(value.errorType),
    errorMessage: readString(value.errorMessage),
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
