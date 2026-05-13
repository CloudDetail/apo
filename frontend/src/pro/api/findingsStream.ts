import type { FindingsNodePayload } from './conversation'

export type FindingsStreamEvent =
  | {
      type: 'tree_shell'
      hydrationId: string
      threadId?: string
      serviceName?: string
      nodes: FindingsNodePayload[]
    }
  | {
      type: 'node_started'
      hydrationId: string
      nodeId: string
      attempt?: number
    }
  | {
      type: 'node_patch'
      hydrationId: string
      nodeId: string
      attempt?: number
      node: FindingsNodePayload
    }
  | {
      type: 'node_failed'
      hydrationId: string
      nodeId: string
      attempt?: number
      retryable?: boolean
      errorType?: string
      message?: string
    }
  | {
      type: 'node_skipped'
      hydrationId: string
      nodeId: string
      reason?: string
    }
  | {
      type: 'tree_done'
      hydrationId: string
      scope?: 'tree' | 'node_retry'
      nodeId?: string
      summary?: {
        success?: number
        failed?: number
        skipped?: number
        unsupported?: number
      }
    }
  | {
      type: 'tree_failed'
      hydrationId?: string
      errorType?: string
      message?: string
    }

export interface StreamFindingsDataHandlers {
  onEvent?: (event: FindingsStreamEvent) => void
  onTreeShell?: (event: Extract<FindingsStreamEvent, { type: 'tree_shell' }>) => void
  onNodeStarted?: (event: Extract<FindingsStreamEvent, { type: 'node_started' }>) => void
  onNodePatch?: (event: Extract<FindingsStreamEvent, { type: 'node_patch' }>) => void
  onNodeFailed?: (event: Extract<FindingsStreamEvent, { type: 'node_failed' }>) => void
  onNodeSkipped?: (event: Extract<FindingsStreamEvent, { type: 'node_skipped' }>) => void
  onTreeDone?: (event: Extract<FindingsStreamEvent, { type: 'tree_done' }>) => void
  onTreeFailed?: (event: Extract<FindingsStreamEvent, { type: 'tree_failed' }>) => void
}

const dispatchFindingsStreamEvent = (
  event: FindingsStreamEvent,
  handlers: StreamFindingsDataHandlers,
) => {
  handlers.onEvent?.(event)

  switch (event.type) {
    case 'tree_shell':
      handlers.onTreeShell?.(event)
      break
    case 'node_started':
      handlers.onNodeStarted?.(event)
      break
    case 'node_patch':
      handlers.onNodePatch?.(event)
      break
    case 'node_failed':
      handlers.onNodeFailed?.(event)
      break
    case 'node_skipped':
      handlers.onNodeSkipped?.(event)
      break
    case 'tree_done':
      handlers.onTreeDone?.(event)
      break
    case 'tree_failed':
      handlers.onTreeFailed?.(event)
      break
  }
}

const parseFindingsStreamLine = (line: string): FindingsStreamEvent => {
  try {
    return JSON.parse(line) as FindingsStreamEvent
  } catch (error) {
    throw new Error(
      `Invalid findings stream event: ${error instanceof Error ? error.message : String(error)}`,
    )
  }
}

export const consumeFindingsStreamResponse = async (
  response: Response,
  handlers: StreamFindingsDataHandlers,
) => {
  if (!response.ok) {
    throw new Error(`Findings stream request failed with status ${response.status}`)
  }

  if (!response.body) {
    throw new Error('Findings stream response body is empty')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() ?? ''

    for (const line of lines) {
      const trimmed = line.trim()
      if (!trimmed) {
        continue
      }
      dispatchFindingsStreamEvent(parseFindingsStreamLine(trimmed), handlers)
    }
  }

  buffer += decoder.decode()
  const trailing = buffer.trim()
  if (trailing) {
    dispatchFindingsStreamEvent(parseFindingsStreamLine(trailing), handlers)
  }
}
