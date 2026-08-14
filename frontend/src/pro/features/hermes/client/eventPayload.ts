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
