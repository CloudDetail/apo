import { runStartEventDecision } from '../client/eventPayload.ts'
import type {
  HermesCheckpoint,
  HermesHistorySnapshot,
  HermesSessionBookmark,
  HermesSessionStatus,
} from '../client/types.ts'

export interface HermesSessionOperationIdentity {
  generation: number
  storedSessionId: string | null
}

export interface HermesRunStartBarrier {
  storedSessionId: string
  runId: string
}

export interface HermesHydrationRuntime {
  status: HermesSessionStatus
  activeRunId: string | null
  checkpoint: HermesCheckpoint | null
  restoreInBackground: boolean
}

export interface HermesPendingSessionCreation<TSession, TScope> {
  generation: number
  promise: Promise<TSession>
  scope: TScope
}

export function shouldRestoreInBackground(status: HermesSessionStatus): boolean {
  return status === 'running' || status === 'waiting_confirmation'
}

export function canBindNewSessionOperation(
  operationGeneration: number,
  currentGeneration: number,
  currentStoredSessionId: string | null,
): boolean {
  return operationGeneration === currentGeneration && currentStoredSessionId === null
}

export function currentSessionCreation<TSession, TScope>(
  pending: HermesPendingSessionCreation<TSession, TScope> | null,
  currentGeneration: number,
  isScopeCurrent: (scope: TScope) => boolean,
): Promise<TSession> | null {
  if (
    !pending ||
    pending.generation !== currentGeneration ||
    !isScopeCurrent(pending.scope)
  ) {
    return null
  }
  return pending.promise
}

export function hydrationRuntimeFromHistory(
  bookmark: HermesSessionBookmark,
  snapshot: HermesHistorySnapshot,
): HermesHydrationRuntime {
  const latestRunStatus = [...snapshot.timeline]
    .reverse()
    .find((item) => item.kind === 'run_status')
  const activeHistoryStatus =
    snapshot.activeRunId && latestRunStatus?.runId === snapshot.activeRunId
      ? latestRunStatus.status === 'waiting'
        ? 'waiting_confirmation'
        : ['running', 'stopping'].includes(latestRunStatus.status)
          ? 'running'
          : null
      : null
  if (activeHistoryStatus) {
    return {
      status: activeHistoryStatus,
      activeRunId: snapshot.activeRunId,
      checkpoint: snapshot.checkpoint,
      restoreInBackground: true,
    }
  }
  const terminalStatus: HermesSessionStatus | null =
    latestRunStatus?.status === 'completed'
      ? 'ready'
      : latestRunStatus?.status === 'failed'
        ? 'failed'
        : latestRunStatus?.status === 'interrupted'
          ? 'interrupted'
          : null

  if (terminalStatus) {
    return {
      status: terminalStatus,
      activeRunId: null,
      checkpoint: null,
      restoreInBackground: false,
    }
  }

  return {
    status: bookmark.status,
    activeRunId:
      snapshot.activeRunId ||
      (shouldRestoreInBackground(bookmark.status) ? bookmark.activeRunId : null),
    checkpoint: snapshot.checkpoint || bookmark.checkpoint,
    restoreInBackground: shouldRestoreInBackground(bookmark.status),
  }
}

export function isCurrentSessionOperation(
  operation: HermesSessionOperationIdentity,
  currentGeneration: number,
  currentStoredSessionId: string | null,
): boolean {
  return (
    operation.generation === currentGeneration &&
    (operation.storedSessionId === null || operation.storedSessionId === currentStoredSessionId)
  )
}

export function isKnownInterruptedRun(
  interruptedRunIds: ReadonlySet<string>,
  eventRunId: string | null,
  activeRunId: string | null,
): boolean {
  const runId = eventRunId || activeRunId
  return Boolean(runId && interruptedRunIds.has(runId))
}

export function shouldIgnoreKnownInterruptedRunEvent(
  interruptedRunIds: ReadonlySet<string>,
  eventRunId: string | null,
  activeRunId: string | null,
  interruptedTerminal: boolean,
): boolean {
  if (eventRunId && interruptedRunIds.has(eventRunId)) {
    return !interruptedTerminal || eventRunId !== activeRunId
  }
  if (interruptedTerminal) return false
  return isKnownInterruptedRun(interruptedRunIds, eventRunId, activeRunId)
}

export async function reconcileInterruptedBookmark(
  request: () => Promise<HermesSessionBookmark>,
): Promise<HermesSessionBookmark | null> {
  try {
    const bookmark = await request()
    return bookmark.status === 'interrupted' ? bookmark : null
  } catch {
    return null
  }
}

export function canApplyInterruptReconciliation(
  interruptedRunId: string | null,
  messageBusy: boolean,
  messageActiveRunId: string | null,
  bookmarkStatus: HermesSessionStatus | null,
  bookmarkActiveRunId: string | null,
): boolean {
  const currentRunId =
    (messageBusy ? messageActiveRunId : null) ||
    (bookmarkStatus === 'running' ? bookmarkActiveRunId : null)
  return !currentRunId || currentRunId === interruptedRunId
}

export function runStartBarrierDecision(
  barrier: HermesRunStartBarrier | null,
  currentStoredSessionId: string | null,
  eventType: string,
  eventRunId: string | null,
): ReturnType<typeof runStartEventDecision> {
  if (!barrier || barrier.storedSessionId !== currentStoredSessionId) return 'pass'
  return runStartEventDecision(barrier.runId, eventType, eventRunId)
}
