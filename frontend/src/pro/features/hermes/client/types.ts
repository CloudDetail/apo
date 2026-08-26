export type HermesConnectionStatus =
  | 'idle'
  | 'connecting'
  | 'connected'
  | 'reconnecting'
  | 'disconnected'
  | 'error'

export type HermesSessionStatus =
  | 'ready'
  | 'running'
  | 'waiting_confirmation'
  | 'failed'
  | 'interrupted'

export interface HermesSession {
  liveSessionId: string
  storedSessionId: string
  title?: string
  running?: boolean
  status?: string
}

export interface HermesSessionBookmark {
  storedSessionId: string
  liveSessionId: string | null
  title: string
  createdAt: string
  updatedAt: string
  status: HermesSessionStatus
  activeRunId: string | null
  checkpoint: HermesCheckpoint | null
}

export interface HermesMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  runId?: string
  complete?: boolean
}

export interface HermesCheckpointChoice {
  value: string
  label: string
  description?: string
}

export interface HermesCheckpoint {
  type: 'approval' | 'clarification'
  requestId: string
  title: string
  description: string
  choices: HermesCheckpointChoice[]
  purpose?: 'plan_confirmation'
}

export interface HermesPlanStep {
  id: string
  title: string
}

export interface HermesPlan {
  steps: HermesPlanStep[]
}

export interface ApoMetricItem {
  title: string
  unit: string
  timeseries: Array<{
    legend: string
    labels?: Record<string, unknown>
    chart: { chartData: Record<string, number> }
  }>
  detection?: string
}

export interface ApoLogRecord {
  timestamp: number | string
  message: string
  level?: string
  service?: string
  instance?: string
  traceId?: string
}

export interface ApoTopologyNode {
  id: string
  type?: string
  position: { x: number; y: number }
  data: {
    label: string
    service?: string
    endpoint?: string
    instance?: string
    isTraced?: boolean
    outOfGroup?: boolean
    status?: 'normal' | 'warning' | 'critical'
    [key: string]: unknown
  }
}

export interface ApoTopologyEdge {
  id: string
  source: string
  target: string
  type?: string
  data?: Record<string, unknown>
}

export interface HermesEvidence<T = unknown> {
  id: string
  kind: 'apo_metrics' | 'apo_topology' | 'apo_logs' | 'apo_json' | (string & {})
  title: string
  summary?: string
  source: { sourceId: string; operationId: string }
  query?: {
    startTime?: number
    endTime?: number
    service?: string
    instance?: string
    host?: string
  }
  data: T
  total?: number
  conversionError?: string
}

export interface HermesToolActivity {
  id: string
  kind: 'tool'
  name: string
  status: 'running' | 'completed' | 'failed'
  context?: string
  args?: Record<string, unknown>
  result?: unknown
  evidence?: HermesEvidence[]
  durationMs?: number
}

export interface HermesCommentaryActivity {
  id: string
  kind: 'commentary'
  text: string
}

export type HermesActivity = HermesToolActivity | HermesCommentaryActivity

export type HermesRunStatus =
  | 'running'
  | 'waiting'
  | 'stopping'
  | 'completed'
  | 'failed'
  | 'interrupted'

interface HermesTimelineBase {
  id: string
  sequenceNo: number
  occurredAt: string
  runId: string | null
}

export interface HermesMessageTimelineItem extends HermesTimelineBase {
  kind: 'message'
  message: HermesMessage
}

export interface HermesActivityTimelineItem extends HermesTimelineBase {
  kind: 'activity'
  activity: HermesActivity
}

export interface HermesCheckpointTimelineItem extends HermesTimelineBase {
  kind: 'checkpoint'
  checkpoint: HermesCheckpoint
  resolved: boolean
  responseText?: string
}

export interface HermesPlanTimelineItem extends HermesTimelineBase {
  kind: 'plan'
  plan: HermesPlan
}

export interface HermesRunStatusTimelineItem extends HermesTimelineBase {
  kind: 'run_status'
  status: HermesRunStatus
  label: string
  detail?: string
  startedAt?: string
  finishedAt?: string
  durationMs?: number
}

export type HermesTimelineItem =
  | HermesMessageTimelineItem
  | HermesActivityTimelineItem
  | HermesCheckpointTimelineItem
  | HermesPlanTimelineItem
  | HermesRunStatusTimelineItem

export interface HermesEvent {
  type: string
  sessionId: string | null
  payload: Record<string, unknown>
}

export interface PersistedConversationEvent {
  eventId: string
  sequenceNo: number
  direction: 'browser_to_gateway' | 'gateway_to_browser'
  eventType: string
  runId: string | null
  payload: Record<string, unknown>
  occurredAt: string
}

export interface HermesHistorySnapshot {
  timeline: HermesTimelineItem[]
  checkpoint: HermesCheckpoint | null
  activeRunId: string | null
}
