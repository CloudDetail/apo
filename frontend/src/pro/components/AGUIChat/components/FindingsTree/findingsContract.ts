import type {
  FindingsLogItem,
  FindingsMetrics,
  FindingsHydrationStatus,
  FindingsNodeKind,
  FindingsNodePayload,
  FindingsReportLink,
  FindingsStatus,
  FindingsTable,
  FindingsTraceData,
} from '../../../../api/conversation'
import { getFindingsTextSectionVisibility } from './detailSectionPolicy'

export interface FindingsNodeData {
  title: string
  subtitle?: string
  content?: string
  status: FindingsStatus
  visualStatus: FindingsStatus
  badge?: string
  active?: boolean
  disabled?: boolean
  nodeId: string
  findingsId?: string
  onClick?: (data: FindingsNodeData) => void
  result?: string
  summaryConclusion?: string
  nodeKind: FindingsNodeKind
  unsupportedReason?: string
  metrics?: FindingsMetrics
  traces?: FindingsTraceData[]
  reportLinks?: FindingsReportLink[]
  logs?: FindingsLogItem[]
  table?: FindingsTable
  hydrationStatus?: FindingsHydrationStatus
  attempt?: number
  retryable?: boolean
  errorType?: string
  errorMessage?: string
}

export interface FindingsTopologyNode {
  id: string
  type: 'findingsNode'
  data: FindingsNodeData
}

export interface FindingsTopologyEdge {
  id: string
  source: string
  target: string
}

export interface FindingsTopologyData {
  nodes: FindingsTopologyNode[]
  edges: FindingsTopologyEdge[]
}

export interface FindingsDetailSections {
  loading: boolean
  summaryConclusion: boolean
  content: boolean
  abnormalMetrics: boolean
  normalMetrics: boolean
  table: boolean
  logs: boolean
  traces: boolean
  unsupported: boolean
}

const hasTableData = (table?: FindingsTable) =>
  Boolean(table && ((table.headers?.length ?? 0) > 0 || (table.data?.length ?? 0) > 0))

const isHydrating = (status?: FindingsHydrationStatus) =>
  status === 'pending' || status === 'running'

const getVisualStatus = ({
  nodeKind,
  status,
}: {
  nodeKind: FindingsNodeKind
  status: FindingsStatus
}): FindingsStatus => (nodeKind === 'unsupported' ? 'normal' : status)

export const toFindingsNodeData = (node: FindingsNodePayload): FindingsNodeData => ({
  nodeId: node.nodeId,
  title: node.title,
  subtitle: node.subtitle,
  content: node.content,
  status: node.status,
  visualStatus: getVisualStatus({
    nodeKind: node.nodeKind,
    status: node.status,
  }),
  summaryConclusion: node.summaryConclusion,
  nodeKind: node.nodeKind,
  unsupportedReason: node.unsupportedReason,
  metrics: node.metrics,
  traces: node.traces,
  reportLinks: node.reportLinks,
  logs: node.logs,
  table: node.table,
  hydrationStatus: node.hydrationStatus,
  attempt: node.attempt,
  retryable: node.retryable,
  errorType: node.errorType,
  errorMessage: node.errorMessage,
})

export const buildFindingsTopologyData = (nodes: FindingsNodePayload[]): FindingsTopologyData => {
  if (!nodes || nodes.length === 0) {
    return { nodes: [], edges: [] }
  }

  return {
    nodes: nodes.map((node) => ({
      id: node.nodeId,
      type: 'findingsNode' as const,
      data: toFindingsNodeData(node),
    })),
    edges: nodes
      .filter((node) => typeof node.parentId === 'string' && node.parentId.length > 0)
      .map((node) => ({
        id: `${node.parentId}-${node.nodeId}`,
        source: node.parentId as string,
        target: node.nodeId,
      })),
  }
}

export const getFindingsDetailSections = (node: FindingsNodeData): FindingsDetailSections => {
  if (isHydrating(node.hydrationStatus)) {
    return {
      loading: true,
      summaryConclusion: false,
      content: false,
      abnormalMetrics: false,
      normalMetrics: false,
      table: false,
      logs: false,
      traces: false,
      unsupported: false,
    }
  }

  const abnormalMetrics = (node.metrics?.abnormalMetrics.length ?? 0) > 0
  const normalMetrics = (node.metrics?.normalMetrics.length ?? 0) > 0
  const table = hasTableData(node.table)
  const logs = (node.logs?.length ?? 0) > 0
  const traces = (node.traces?.length ?? 0) > 0
  const unsupported = node.nodeKind === 'unsupported' || Boolean(node.unsupportedReason)
  const textSections = getFindingsTextSectionVisibility({
    summaryConclusion: node.summaryConclusion,
    content: node.content,
  })

  return {
    loading: false,
    summaryConclusion: textSections.summaryConclusion,
    content: textSections.content,
    abnormalMetrics,
    normalMetrics,
    table,
    logs,
    traces,
    unsupported,
  }
}
