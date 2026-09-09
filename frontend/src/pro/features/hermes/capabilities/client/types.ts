/**
 * Product-facing view of Hermes Workflow library contracts.
 *
 * Hermes owns the wire format. APO keeps only fields used by the task library
 * and preserves the complete structured input/step definitions when editing.
 */

export type WorkflowStatus = 'active' | 'disabled' | 'archived' | 'unknown' | string

export type WorkflowInputSource =
  | 'conversation_or_prompt'
  | 'conversation'
  | 'user'
  | 'system'
  | 'step_output'

export interface WorkflowInput {
  name: string
  required: boolean
  source: WorkflowInputSource
  /** Original wire value retained only to avoid writing a redaction placeholder. */
  sourceName?: string
}

export interface WorkflowStep {
  id: string
  purpose: string
  providerRef: string | null
  capabilityRef: string | null
  inputMap: Record<string, unknown>
  expectedOutput: string
  effects: string[]
  /** Original wire values retained only for unchanged fields in whole-step updates. */
  sourcePurpose?: string
  sourceExpectedOutput?: string
}

export interface WorkflowRun {
  runId: string
  sessionId?: string | null
  version?: number | null
  status?: string | null
  relationship?: string | null
  title?: string | null
  startedAt?: string | null
  completedAt?: string | null
  createdAt?: string | null
  updatedAt?: string | null
  sourceRun?: boolean | WorkflowRun
  evidence?: Array<Record<string, unknown>>
  [key: string]: unknown
}

export interface WorkflowVersion {
  version: number
  digest?: string | null
  createdAt?: string | null
  createdBy?: string | null
  sourceRunId?: string | null
  isCurrent?: boolean
  [key: string]: unknown
}

export interface WorkCapability {
  schemaVersion: string
  id: string
  name: string
  goal: string
  status: WorkflowStatus
  enabled: boolean | null
  pinned: boolean
  currentVersion: number
  examples: string[]
  sourceExamples?: string[]
  inputs: WorkflowInput[]
  inputHints: string[]
  steps: WorkflowStep[]
  notes: string[]
  effects: string[]
  expectedResult: string | null
  capabilityRefs: string[]
  sourceRunId?: string | null
  sourceRun?: WorkflowRun | null
  createdAt?: string | null
  updatedAt?: string | null
  lastUsedAt?: string | null
  successCount: number
  reuseCount: number
  recentFailures: number
  lastRunStatus?: string | null
  successRate: number | null
  lastRun?: WorkflowRun | null
  runs: WorkflowRun[]
  versions: WorkflowVersion[]
  allowedActions: string[]
  actionBlockers: Array<string | { action?: string; code?: string; reason?: string; message?: string }>
  integrationMode: 'hermes'
}

export interface CapabilityListParams {
  page?: number
  pageSize?: number
  search?: string
  status?: string
  recentRun?: string
}

export interface CapabilityListResult {
  items: WorkCapability[]
  total: number
  page: number
  pageSize: number
  warnings?: string[]
}

export interface CapabilityRunsResult {
  items: WorkflowRun[]
  total: number
  page: number
  pageSize: number
}

export interface CapabilityRunsParams {
  page?: number
  pageSize?: number
  version?: number
}

export interface WorkflowEditValues {
  name: string
  goal: string
  examples: string[]
  sourceExamples?: string[]
  inputs: WorkflowInput[]
  steps: WorkflowStep[]
}

export type WorkflowUpdateInput = Partial<WorkflowEditValues> & {
  enabled?: boolean
  pinned?: boolean
}

export interface WorkflowDeleteResult {
  deletedIds: string[]
  deletedCount: number
}

export interface WorkflowProjectionNode {
  id: string
  title: string
  status?: string
  summary?: string
  current: boolean
  requiresUserAction: boolean
  dependencies: string[]
  facts: string[]
  gaps: string[]
  failure?: string
}

export interface WorkflowProjectionEvidence {
  id: string
  title: string
  stepId?: string
  kind?: string
  status?: string
  summary?: string
  trust?: string
  receiptId?: string
}

export interface WorkflowProjectionResult {
  status?: string
  summary?: string
  confidence?: string
  nextAction?: string
  evidenceRefs: string[]
}

export interface WorkflowProjectionWorkflow {
  id: string
  name: string
  version?: number
  relationship?: string
  executionMode?: string
}

export interface WorkflowProjectionRun {
  runId: string
  sessionId?: string
  goal?: string
  executionMode?: string
  status?: string
  stateRevision?: number
  createdAt?: string
  updatedAt?: string
  finishedAt?: string
}

export interface CapabilityProjection {
  runId?: string
  profile?: string
  displayType?: string
  goal?: string
  revision?: number
  hasPlan?: boolean
  eventCount?: number
  status?: string
  workflow?: WorkflowProjectionWorkflow
  workflowRun?: WorkflowProjectionRun
  nodes?: WorkflowProjectionNode[]
  evidence?: WorkflowProjectionEvidence[]
  result?: WorkflowProjectionResult
  gaps?: string[]
  failures?: string[]
}

/** State handed to the existing Agent page for an explicit Workflow run. */
export interface WorkflowLaunchContext {
  workflowId: string
  workflowName: string
  version: number
  goal: string
  source: 'workflow_library'
}

/** State handed to the Agent page for a health issue investigation. */
export interface AgentLaunchContext {
  goal: string
  source: 'system_health_pending_issue' | 'system_health_service_issue'
  autoSend?: boolean
}
