export function normalizeHermesEventPayload(value: unknown): Record<string, unknown> {
  if (!isObject(value)) return {}
  const payload = { ...value }
  if (!Array.isArray(payload.evidence)) {
    const evidence = nestedEvidence(payload)
    if (evidence) payload.evidence = normalizeEvidence(evidence, payload)
  } else {
    payload.evidence = normalizeEvidence(payload.evidence, payload)
  }
  return payload
}

export function isInterruptedTerminalEvent(
  eventType: string,
  payload: Record<string, unknown>,
): boolean {
  if (['session.interrupted', 'run.cancelled', 'run.canceled'].includes(eventType)) return true
  if (eventType !== 'message.complete') return false
  return ['interrupted', 'cancelled', 'canceled'].includes(
    String(payload.status || '').trim().toLowerCase(),
  )
}

export type HermesRunStartEventDecision = 'pass' | 'ignore' | 'clear'

const RUN_SCOPED_EVENT_TYPES = new Set([
  'message.delta',
  'message.complete',
  'message.interim',
  'reasoning.delta',
  'tool.start',
  'tool.started',
  'tool.complete',
  'tool.completed',
  'tool.failed',
  'approval.request',
  'clarify.request',
  'run.completed',
  'run.failed',
  'run.cancelled',
  'run.canceled',
  'session.interrupted',
  'error',
])

export function isRunScopedEventType(eventType: string): boolean {
  return RUN_SCOPED_EVENT_TYPES.has(eventType)
}

export function runIdFromEventPayload(payload: Record<string, unknown>): string | null {
  const value = payload.run_id ?? payload.runId ?? payload.client_run_id ?? payload.clientRunId
  return typeof value === 'string' && value ? value : null
}

export function runStartEventDecision(
  expectedRunId: string | null,
  eventType: string,
  eventRunId: string | null,
): HermesRunStartEventDecision {
  if (!expectedRunId) return 'pass'
  if (eventType === 'message.start') {
    return !eventRunId || eventRunId === expectedRunId ? 'clear' : 'pass'
  }
  if (eventRunId) return 'pass'
  return isRunScopedEventType(eventType) ? 'ignore' : 'pass'
}

function normalizeEvidence(
  evidence: unknown[],
  payload: Record<string, unknown>,
): Record<string, unknown>[] {
  const result = capabilityResult(payload)
  const projection = result && isObject(result.modelProjection) ? result.modelProjection : null
  const projectionData = projection?.data
  return evidence.flatMap((item, index) => {
    if (!isObject(item)) return []
    const locator = isObject(item.locator) ? item.locator : {}
    const hasData = Object.prototype.hasOwnProperty.call(item, 'data')
    return [{
      ...item,
      id: String(item.id ?? item.evidenceId ?? `evidence-${index + 1}`),
      title: String(item.title ?? item.summary ?? '查询依据'),
      source: isObject(item.source) ? item.source : {
        sourceId: String(locator.sourceId ?? ''),
        operationId: String(locator.operationId ?? ''),
      },
      ...(hasData ? {} : { data: projectionData }),
      ...(item.conversionError || !locator.conversionError
        ? {}
        : { conversionError: String(locator.conversionError) }),
    }]
  })
}

function capabilityResult(payload: Record<string, unknown>): Record<string, any> | null {
  for (const candidate of [payload.result, payload.output, payload.capabilityResult, payload.capability_result]) {
    if (!isObject(candidate)) continue
    if (candidate.schemaVersion === 'capability.result.v1') return candidate
    const nested = candidate.capabilityResult ?? candidate.capability_result
    if (isObject(nested)) return nested
  }
  return null
}

function nestedEvidence(payload: Record<string, unknown>): unknown[] | null {
  const candidates = [
    payload.result,
    payload.output,
    payload.capability_result,
    payload.capabilityResult,
  ]
  for (const candidate of candidates) {
    if (!isObject(candidate)) continue
    if (Array.isArray(candidate.evidence)) return candidate.evidence
    const capabilityResult = candidate.capabilityResult ?? candidate.capability_result
    if (isObject(capabilityResult) && Array.isArray(capabilityResult.evidence)) {
      return capabilityResult.evidence
    }
  }
  return null
}

function isObject(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
