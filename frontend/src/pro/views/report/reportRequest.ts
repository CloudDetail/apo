export interface ReportViewRequest {
  startTime: number
  endTime: number
  tags: Record<string, string>
}

interface ReportTimeRange {
  startTime?: number | string | null
  endTime?: number | string | null
}

const toFiniteNumber = (value: unknown) => {
  if (value === null || value === undefined || value === '') return undefined

  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

const readNumberParam = (params: URLSearchParams, key: string) => {
  return toFiniteNumber(params.get(key))
}

const readStringParam = (params: URLSearchParams, key: string) => {
  const value = params.get(key)?.trim()
  return value || undefined
}

export function buildReportViewRequest(
  params: URLSearchParams,
  fallbackTimeRange?: ReportTimeRange,
): ReportViewRequest | undefined {
  const startTime = readNumberParam(params, 'startTime') ?? toFiniteNumber(fallbackTimeRange?.startTime)
  const endTime = readNumberParam(params, 'endTime') ?? toFiniteNumber(fallbackTimeRange?.endTime)
  const alertEventId = readStringParam(params, 'alertEventId')
  const agentChatId = readStringParam(params, 'agentChatId')

  if (!startTime || !endTime || (!alertEventId && !agentChatId)) {
    return undefined
  }

  return {
    startTime,
    endTime,
    tags: alertEventId ? { alertEventId } : { agentChatId: agentChatId as string },
  }
}
