import type {
  FindingsHydrationStatus,
  FindingsNodeKind,
  FindingsStatus,
} from '../../../../api/conversation'

export type FindingsNodeThemeKey = FindingsStatus | 'loading'

const VALID_STATUS_SET = new Set<FindingsStatus>(['normal', 'warning', 'abnormal'])

const resolveStatus = (value: unknown, fallback: FindingsStatus = 'warning'): FindingsStatus => {
  if (typeof value === 'string' && VALID_STATUS_SET.has(value as FindingsStatus)) {
    return value as FindingsStatus
  }

  return fallback
}

export const getFindingsNodePresentation = ({
  status = 'warning',
  visualStatus,
  nodeKind,
  hydrationStatus,
  badge,
}: {
  status?: FindingsStatus
  visualStatus?: FindingsStatus
  nodeKind?: FindingsNodeKind
  hydrationStatus?: FindingsHydrationStatus
  badge?: string
}) => {
  const isRunning = hydrationStatus === 'running'
  const isPending = hydrationStatus === 'pending'
  const isFailed = hydrationStatus === 'failed'
  const isLoading = isRunning || isPending
  const resolvedStatus = resolveStatus(
    isFailed ? 'warning' : visualStatus ?? (nodeKind === 'unsupported' ? 'normal' : status),
  )

  return {
    isRunning,
    isPending,
    isFailed,
    isLoading,
    resolvedStatus,
    themeKey: isLoading ? 'loading' : resolvedStatus,
    badgeLabel: isRunning || isFailed ? undefined : badge,
    badgeLabelKey: isRunning
      ? 'findingsTree.hydrationLoading'
      : isFailed
        ? 'findingsTree.hydrationFailed'
        : undefined,
  }
}
