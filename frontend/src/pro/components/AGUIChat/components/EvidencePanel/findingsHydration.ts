import type {
  FindingsNodePayload,
  FindingsStreamEvent,
  StreamFindingsDataHandlers,
  StreamFindingsDataParams,
} from '../../../../api/conversation'

export type FindingsHydrationPhase = 'idle' | 'streaming' | 'done' | 'failed'

export interface FindingsCacheEntry {
  nodes: FindingsNodePayload[]
  phase: FindingsHydrationPhase
  hydrationId?: string
  errorMessage?: string
}

export const isTerminalFindingsPhase = (phase?: FindingsHydrationPhase) =>
  phase === 'done' || phase === 'failed'

export const isWholeTreeDoneEvent = (event: FindingsStreamEvent) =>
  event.type === 'tree_done' && event.scope !== 'node_retry'

export const shouldStartFindingsStream = ({
  cachedPhase,
  isLoading,
  refresh = false,
}: {
  cachedPhase?: FindingsHydrationPhase
  isLoading: boolean
  refresh?: boolean
}) => {
  if (isLoading) {
    return false
  }
  if (!refresh && isTerminalFindingsPhase(cachedPhase)) {
    return false
  }
  return true
}

export const getFindingsStreamOverflowKey = (
  loadingKeys: Iterable<string>,
  activeKey: string | null,
  maxStreams: number,
) => {
  const keys = Array.from(loadingKeys)
  if (keys.length < maxStreams) {
    return null
  }
  return keys.find((key) => key !== activeKey) ?? null
}

export const getStaleFindingsStreamKeys = (loadingKeys: Iterable<string>, activeKey: string) =>
  Array.from(loadingKeys).filter((key) => key !== activeKey)

export const abortStaleFindingsStreams = ({
  loadingKeys,
  activeKey,
  abort,
}: {
  loadingKeys: Iterable<string>
  activeKey: string
  abort: (cacheKey: string) => void
}) => {
  const staleKeys = getStaleFindingsStreamKeys(loadingKeys, activeKey)
  staleKeys.forEach((cacheKey) => abort(cacheKey))
  return staleKeys
}

export const buildFindingsCacheKey = (threadId: string, serviceName: string) =>
  `${threadId}:${serviceName}`

export const buildFindingsNodeRetryParams = (
  threadId: string,
  serviceName: string,
  nodeId: string,
  signal?: AbortSignal,
): StreamFindingsDataParams => ({
  threadId,
  serviceName,
  nodeId,
  refresh: true,
  signal,
})

export const runFindingsStreamRequest = async ({
  streamFindingsData,
  params,
  cacheKey,
  getActiveKey,
  onActiveEvent,
}: {
  streamFindingsData: (
    params: StreamFindingsDataParams,
    handlers: StreamFindingsDataHandlers,
  ) => Promise<void>
  params: StreamFindingsDataParams
  cacheKey: string
  getActiveKey: () => string | null
  onActiveEvent: (event: FindingsStreamEvent) => void
}) => {
  await streamFindingsData(params, {
    onEvent: (event) => {
      if (getActiveKey() !== cacheKey) {
        return
      }
      onActiveEvent(event)
    },
  })
}

export const shouldSyncSelectedFindings = ({
  cacheKey,
  activeKey,
  currentThreadId,
  selectedServiceName,
}: {
  cacheKey: string
  activeKey: string | null
  currentThreadId?: string | null
  selectedServiceName?: string | null
}) => {
  if (activeKey !== cacheKey || !currentThreadId || !selectedServiceName) {
    return false
  }

  return buildFindingsCacheKey(currentThreadId, selectedServiceName) === cacheKey
}

export const createPendingFindingsShell = (
  nodes: FindingsNodePayload[],
): FindingsNodePayload[] =>
  nodes.map((node) => ({
    nodeId: node.nodeId,
    title: node.title,
    subtitle: node.subtitle,
    nodeKind: node.nodeKind,
    status: node.status,
    statusColor: node.statusColor,
    parentId: node.parentId,
    summaryConclusion: node.summaryConclusion,
    hydrationStatus: 'pending',
  }))

export const removeNodeAndDescendants = (
  nodes: FindingsNodePayload[],
  nodeId: string,
): FindingsNodePayload[] => {
  const removed = new Set<string>([nodeId])
  let changed = true

  while (changed) {
    changed = false
    for (const node of nodes) {
      if (typeof node.parentId === 'string' && removed.has(node.parentId) && !removed.has(node.nodeId)) {
        removed.add(node.nodeId)
        changed = true
      }
    }
  }

  return nodes.filter((node) => !removed.has(node.nodeId))
}

const mergeNodePatch = (
  nodes: FindingsNodePayload[],
  patchedNode: FindingsNodePayload,
): FindingsNodePayload[] => {
  let replaced = false
  const merged = nodes.map((node) => {
    if (node.nodeId !== patchedNode.nodeId) {
      return node
    }
    replaced = true
    return patchedNode
  })

  return replaced ? merged : [...nodes, patchedNode]
}

export const applyFindingsStreamEvent = (
  nodes: FindingsNodePayload[],
  event: FindingsStreamEvent,
): FindingsNodePayload[] => {
  switch (event.type) {
    case 'tree_shell':
      return event.nodes.map((node) => ({
        ...node,
        hydrationStatus: node.hydrationStatus ?? 'pending',
      }))

    case 'node_started':
      return nodes.map((node) =>
        node.nodeId === event.nodeId
          ? {
              ...node,
              hydrationStatus: 'running',
              attempt: event.attempt,
            }
          : node,
      )

    case 'node_patch':
      return mergeNodePatch(nodes, {
        ...event.node,
        hydrationStatus: 'success',
        attempt: event.attempt,
      })

    case 'node_failed':
      return nodes.map((node) =>
        node.nodeId === event.nodeId
          ? {
              ...node,
              hydrationStatus: 'failed',
              attempt: event.attempt,
              retryable: event.retryable,
              errorType: event.errorType,
              errorMessage: event.message,
            }
          : node,
      )

    case 'node_skipped':
      return removeNodeAndDescendants(nodes, event.nodeId)

    default:
      return nodes
  }
}
