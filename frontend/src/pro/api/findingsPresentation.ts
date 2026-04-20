export type FindingsStatus = 'normal' | 'warning' | 'abnormal'

export type FindingsNodeKind = 'summary' | 'evidence' | 'unsupported'

const STATUS_MAP: Record<string, FindingsStatus> = {
  normal: 'normal',
  healthy: 'normal',
  ok: 'normal',
  success: 'normal',
  warning: 'warning',
  warn: 'warning',
  yellow: 'warning',
  info: 'warning',
  muted: 'warning',
  unsupported: 'warning',
  'no-valid-evidence': 'warning',
  abnormal: 'abnormal',
  red: 'abnormal',
  error: 'abnormal',
  critical: 'abnormal',
  failed: 'abnormal',
}

const UNSUPPORTED_REASON_TRANSLATION_KEYS = {
  'no-downstream-services': 'evidencePanel.noDownstreamServices',
  'red-metrics-unavailable': 'evidencePanel.redMetricsUnavailable',
  'red-query-failed': 'evidencePanel.redQueryFailed',
  'no-valid-evidence': 'evidencePanel.noValidEvidence',
} as const

const readString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

export const normalizeFindingsStatus = (
  value: unknown,
  fallback: FindingsStatus = 'warning',
): FindingsStatus => {
  const normalized = readString(value)?.toLowerCase()
  if (!normalized) {
    return fallback
  }

  return STATUS_MAP[normalized] ?? fallback
}

export const stabilizeFindingsNodeStatus = ({
  nodeKind,
  status,
  statusColor,
}: {
  nodeKind: FindingsNodeKind
  status?: unknown
  statusColor?: unknown
}): FindingsStatus => {
  const fallbackStatus = nodeKind === 'unsupported' ? 'warning' : 'normal'
  return normalizeFindingsStatus(status ?? statusColor, fallbackStatus)
}

export const getUnsupportedReasonTranslationKey = (reason?: string): string | null => {
  if (!reason) {
    return null
  }

  return (
    UNSUPPORTED_REASON_TRANSLATION_KEYS[
      reason as keyof typeof UNSUPPORTED_REASON_TRANSLATION_KEYS
    ] ?? null
  )
}
