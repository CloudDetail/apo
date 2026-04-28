export interface BuildThreadTransitionResetInput {
  currentRunId?: string
  loadRevision: number
  targetThreadId?: string
}

export interface ThreadTransitionRuntimeReset {
  isRunning: false
  currentRunId: undefined
  currentStep: undefined
  currentMessage: undefined
  currentToolCall: undefined
  loading: boolean
  error: null
}

export interface ThreadTransitionScopedReset<
  Message = unknown,
  ToolCall = unknown,
  AgentToolDisplay = unknown,
> {
  messages: Message[]
  state: Record<string, never>
  findingsToolCalls: Map<string, Map<string, ToolCall>>
  findingsAgentToolDisplay: Map<string, AgentToolDisplay[]>
}

export interface ThreadTransitionReset<
  Message = unknown,
  ToolCall = unknown,
  AgentToolDisplay = unknown,
> {
  shouldCancelRun: boolean
  shouldLoadHistory: boolean
  targetThreadId?: string
  nextLoadRevision: number
  runtime: ThreadTransitionRuntimeReset
  scoped: ThreadTransitionScopedReset<Message, ToolCall, AgentToolDisplay>
}

export const buildThreadTransitionReset = <
  Message = unknown,
  ToolCall = unknown,
  AgentToolDisplay = unknown,
>({
  currentRunId,
  loadRevision,
  targetThreadId,
}: BuildThreadTransitionResetInput): ThreadTransitionReset<Message, ToolCall, AgentToolDisplay> => ({
  shouldCancelRun: Boolean(currentRunId),
  shouldLoadHistory: Boolean(targetThreadId),
  targetThreadId,
  nextLoadRevision: loadRevision + 1,
  runtime: {
    isRunning: false,
    currentRunId: undefined,
    currentStep: undefined,
    currentMessage: undefined,
    currentToolCall: undefined,
    loading: Boolean(targetThreadId),
    error: null,
  },
  scoped: {
    messages: [],
    state: {},
    findingsToolCalls: new Map(),
    findingsAgentToolDisplay: new Map(),
  },
})

export const shouldApplyThreadLoadResult = ({
  activeLoadRevision,
  responseLoadRevision,
}: {
  activeLoadRevision: number
  responseLoadRevision: number
}) => activeLoadRevision === responseLoadRevision

export const hasThreadSelectionChanged = (
  previousThreadId?: string,
  nextThreadId?: string,
) => previousThreadId !== nextThreadId

export const shouldApplyRouteThreadSelection = ({
  previousRouteThreadId,
  nextRouteThreadId,
  activeThreadId,
}: {
  previousRouteThreadId?: string
  nextRouteThreadId?: string
  activeThreadId?: string
}) => {
  if (!hasThreadSelectionChanged(previousRouteThreadId, nextRouteThreadId)) {
    return false
  }

  if (nextRouteThreadId && nextRouteThreadId === activeThreadId) {
    return false
  }

  return true
}
