import type { FindingsHydrationStatus, FindingsStatus } from '../../../../api/conversation'

export type TopologyVisualStatus = FindingsStatus | 'loading'

const isFindingsStatus = (value: unknown): value is FindingsStatus =>
  value === 'normal' || value === 'warning' || value === 'abnormal'

const isHydratingStatus = (value?: FindingsHydrationStatus) =>
  value === 'pending' || value === 'running'

export const getNodeVisualStatus = (data?: {
  visualStatus?: unknown
  nodeKind?: unknown
  status?: unknown
  hydrationStatus?: FindingsHydrationStatus
}): TopologyVisualStatus => {
  if (isHydratingStatus(data?.hydrationStatus)) {
    return 'loading'
  }
  if (isFindingsStatus(data?.visualStatus)) {
    return data.visualStatus
  }
  if (data?.nodeKind === 'unsupported') {
    return 'normal'
  }
  if (isFindingsStatus(data?.status)) {
    return data.status
  }
  return 'normal'
}

export const getStatusColor = (status: TopologyVisualStatus) =>
  status === 'loading'
    ? '#94a3b8'
    : status === 'abnormal'
      ? '#ff4d6d'
      : status === 'warning'
        ? '#f59e0b'
        : '#22c55e'

export const getStatusStrokeColor = (status: TopologyVisualStatus) =>
  status === 'loading'
    ? '#64748b'
    : status === 'abnormal'
      ? '#dc2626'
      : status === 'warning'
        ? '#d97706'
        : '#059669'
