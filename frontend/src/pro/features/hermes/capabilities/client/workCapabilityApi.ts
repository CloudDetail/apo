import type {
  CapabilityListParams,
  CapabilityListResult,
  CapabilityProjection,
  CapabilityRunsParams,
  CapabilityRunsResult,
  WorkflowDeleteResult,
  WorkflowEditValues,
  WorkflowInput,
  WorkflowInputSource,
  WorkflowProjectionEvidence,
  WorkflowProjectionNode,
  WorkflowProjectionResult,
  WorkflowProjectionRun,
  WorkflowProjectionWorkflow,
  WorkflowRun,
  WorkflowStep,
  WorkflowUpdateInput,
  WorkflowVersion,
  WorkCapability,
} from './types.ts'
import { redactSensitiveText } from './sensitive.ts'
import { hermesApiBaseUrl } from '../../client/config.ts'
import {
  HERMES_USER_ID_HEADER,
  requireHermesUserId,
  type HermesUserIdReader,
} from '../../client/userScope.ts'

const WORKFLOW_PREFIX = '/workflows'
const WORKFLOW_RUN_PREFIX = '/workflow-runs'
const WORKFLOW_SCHEMA_VERSION = 'hermes-qa.workflow.v1'
const WORKFLOW_INPUT_SOURCES = new Set<WorkflowInputSource>([
  'conversation_or_prompt',
  'conversation',
  'user',
  'system',
  'step_output',
])

type RequestLike = typeof fetch

export interface WorkCapabilityApiOptions {
  request?: RequestLike
  baseUrl?: string
  readUserId?: HermesUserIdReader
}

export class WorkCapabilityApiError extends Error {
  readonly status?: number
  readonly code?: string

  constructor(message: string, options: { status?: number; code?: string } = {}) {
    super(message)
    this.name = 'WorkCapabilityApiError'
    this.status = options.status
    this.code = options.code
  }
}

export class HermesWorkflowUnavailableError extends WorkCapabilityApiError {
  constructor(message = '无法连接 Agent 运维任务库。', status?: number) {
    super(message, { status, code: 'hermes_workflow_unavailable' })
    this.name = 'HermesWorkflowUnavailableError'
  }
}

const isRecord = (value: unknown): value is Record<string, any> =>
  Boolean(value && typeof value === 'object' && !Array.isArray(value))

const record = (value: unknown): Record<string, any> => (isRecord(value) ? value : {})

const text = (value: unknown): string => (value === undefined || value === null ? '' : String(value).trim())

const safeText = (value: unknown): string => redactSensitiveText(text(value))

const nullableText = (value: unknown): string | null => text(value) || null

const numberOrNull = (value: unknown): number | null => {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() && Number.isFinite(Number(value))) return Number(value)
  return null
}

const numberOr = (value: unknown, fallback: number): number => numberOrNull(value) ?? fallback

const stringList = (value: unknown, redact = false): string[] => {
  if (!Array.isArray(value)) return []
  return value
    .flatMap((item) => {
      if (typeof item !== 'string' && typeof item !== 'number') return []
      const normalized = text(item)
      return normalized ? [redact ? redactSensitiveText(normalized) : normalized] : []
    })
}

const trimTrailingSlash = (value: string): string =>
  value.length > 1 ? value.replace(/\/+$/, '') : value

const endpointUrl = (base: string, path: string): string => {
  if (/^https?:\/\//i.test(path)) return path
  if (!base) return path
  return `${trimTrailingSlash(base)}${path.startsWith('/') ? path : `/${path}`}`
}

function requestHeaders(hasBody: boolean, initial?: HeadersInit): Headers {
  const headers = new Headers(initial)
  headers.set('Accept', 'application/json')
  if (hasBody) headers.set('Content-Type', 'application/json')
  return headers
}

function errorMessage(payload: unknown, fallback: string): string {
  const value = record(payload)
  if (typeof value.detail === 'string') return safeText(value.detail).slice(0, 500)
  if (isRecord(value.detail) && typeof value.detail.message === 'string') return safeText(value.detail.message).slice(0, 500)
  if (Array.isArray(value.detail)) {
    const first = value.detail.find((item) => isRecord(item) && typeof item.msg === 'string')
    if (isRecord(first)) return safeText(first.msg).slice(0, 500)
  }
  if (typeof value.message === 'string') return safeText(value.message).slice(0, 500)
  if (isRecord(value.error) && typeof value.error.message === 'string') return safeText(value.error.message).slice(0, 500)
  return safeText(fallback)
}

function normalizeWorkflowInput(value: unknown): WorkflowInput {
  const raw = record(value)
  const sourceName = text(raw.name)
  const name = redactSensitiveText(sourceName)
  const source = text(raw.source) as WorkflowInputSource
  if (!name || !WORKFLOW_INPUT_SOURCES.has(source)) {
    throw new WorkCapabilityApiError('Agent Workflow 输入契约不完整。', { code: 'invalid_response' })
  }
  return {
    name,
    required: raw.required === true,
    source,
    ...(sourceName !== name ? { sourceName } : {}),
  }
}

function normalizeWorkflowStep(value: unknown): WorkflowStep {
  const raw = record(value)
  const id = text(raw.id)
  const sourcePurpose = text(raw.purpose)
  const sourceExpectedOutput = text(raw.expectedOutput)
  const purpose = redactSensitiveText(sourcePurpose)
  const expectedOutput = redactSensitiveText(sourceExpectedOutput)
  const providerRef = nullableText(raw.providerRef)
  const capabilityRef = nullableText(raw.capabilityRef)
  if (!id || !purpose || !expectedOutput || Boolean(providerRef) !== Boolean(capabilityRef) || !isRecord(raw.inputMap)) {
    throw new WorkCapabilityApiError('Agent Workflow 步骤契约不完整。', { code: 'invalid_response' })
  }
  return {
    id,
    purpose,
    providerRef,
    capabilityRef,
    inputMap: { ...raw.inputMap },
    expectedOutput,
    effects: stringList(raw.effects),
    ...(sourcePurpose !== purpose ? { sourcePurpose } : {}),
    ...(sourceExpectedOutput !== expectedOutput ? { sourceExpectedOutput } : {}),
  }
}

export function toWorkflowRun(value: unknown): WorkflowRun {
  const raw = record(value)
  return {
    runId: text(raw.runId),
    sessionId: nullableText(raw.sessionId),
    version: numberOrNull(raw.version),
    status: nullableText(raw.status),
    relationship: nullableText(raw.relationship),
    title: safeText(raw.title || raw.goal) || null,
    startedAt: nullableText(raw.startedAt),
    completedAt: nullableText(raw.completedAt),
    createdAt: nullableText(raw.createdAt),
    updatedAt: nullableText(raw.updatedAt),
  }
}

export function toWorkflowVersion(value: unknown, currentVersion?: number): WorkflowVersion | null {
  const raw = record(value)
  const version = numberOrNull(raw.version ?? raw.version_number)
  if (version === null) return null
  return {
    version,
    digest: nullableText(raw.digest ?? raw.content_digest),
    createdAt: nullableText(raw.createdAt ?? raw.created_at),
    createdBy: nullableText(raw.createdBy ?? raw.created_by),
    sourceRunId: nullableText(raw.sourceRunId ?? raw.source_run_id),
    isCurrent: currentVersion === version,
  }
}

function projectionNodes(value: unknown): WorkflowProjectionNode[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((item, index) => {
    if (!isRecord(item)) return []
    const id = text(item.id)
    if (!id) return []
    const title = safeText(item.title || item.name || id) || `步骤 ${index + 1}`
    const status = safeText(item.status)
    const summary = safeText(item.summary || item.description)
    return [{
      id,
      title,
      ...(status ? { status } : {}),
      ...(summary ? { summary } : {}),
      current: item.current === true,
      requiresUserAction: item.requiresUserAction === true,
      dependencies: stringList(item.dependencies),
      facts: stringList(item.facts, true),
      gaps: projectionGaps(item.gaps),
      ...(projectionError(item.failure) ? { failure: projectionError(item.failure) } : {}),
    }]
  })
}

function projectionEvidence(value: unknown): WorkflowProjectionEvidence[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((item, index) => {
    if (!isRecord(item)) return []
    const id = text(item.evidenceId || item.id)
    if (!id) return []
    const kind = safeText(item.kind)
    const title = safeText(item.title || item.name || kind || id) || `Evidence ${index + 1}`
    const stepId = safeText(item.stepId)
    const status = safeText(item.status)
    const summary = safeText(item.summary || item.description)
    const trust = safeText(item.trust)
    const receiptId = safeText(item.receiptId)
    return [{
      id,
      title,
      ...(stepId ? { stepId } : {}),
      ...(kind ? { kind } : {}),
      ...(status ? { status } : {}),
      ...(summary ? { summary } : {}),
      ...(trust ? { trust } : {}),
      ...(receiptId ? { receiptId } : {}),
    }]
  })
}

function projectionError(value: unknown): string | undefined {
  if (typeof value === 'string' || typeof value === 'number') return safeText(value) || undefined
  if (!isRecord(value)) return undefined
  return safeText(value.message || value.detail || value.reason || value.code) || undefined
}

function projectionGaps(value: unknown): string[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((item) => {
    if (typeof item === 'string' || typeof item === 'number') {
      const gap = safeText(item)
      return gap ? [gap] : []
    }
    if (!isRecord(item)) return []
    const gap = safeText(item.message || item.reason || item.detail || item.code || item.label || item.title)
    return gap ? [gap] : []
  })
}

function projectionResult(value: unknown): WorkflowProjectionResult | undefined {
  const raw = record(value)
  const status = safeText(raw.status)
  const summary = safeText(raw.summary)
  const confidence = safeText(raw.confidence)
  const nextAction = safeText(raw.nextAction)
  const evidenceRefs = stringList(raw.evidenceRefs)
  if (!status && !summary && !confidence && !nextAction && !evidenceRefs.length) return undefined
  return {
    ...(status ? { status } : {}),
    ...(summary ? { summary } : {}),
    ...(confidence ? { confidence } : {}),
    ...(nextAction ? { nextAction } : {}),
    evidenceRefs,
  }
}

function projectionWorkflow(value: unknown): WorkflowProjectionWorkflow | undefined {
  const raw = record(value)
  const id = text(raw.id)
  if (!id) return undefined
  const name = safeText(raw.name || id) || id
  const version = numberOrNull(raw.version)
  const relationship = safeText(raw.relationship)
  const executionMode = safeText(raw.executionMode)
  return {
    id,
    name,
    ...(version !== null ? { version } : {}),
    ...(relationship ? { relationship } : {}),
    ...(executionMode ? { executionMode } : {}),
  }
}

function projectionRun(value: unknown): WorkflowProjectionRun | undefined {
  const raw = record(value)
  const runId = text(raw.runId)
  if (!runId) return undefined
  const sessionId = safeText(raw.sessionId)
  const goal = safeText(raw.goal)
  const executionMode = safeText(raw.executionMode)
  const status = safeText(raw.status)
  const stateRevision = numberOrNull(raw.stateRevision)
  const createdAt = nullableText(raw.createdAt)
  const updatedAt = nullableText(raw.updatedAt)
  const finishedAt = nullableText(raw.finishedAt)
  return {
    runId,
    ...(sessionId ? { sessionId } : {}),
    ...(goal ? { goal } : {}),
    ...(executionMode ? { executionMode } : {}),
    ...(status ? { status } : {}),
    ...(stateRevision !== null ? { stateRevision } : {}),
    ...(createdAt ? { createdAt } : {}),
    ...(updatedAt ? { updatedAt } : {}),
    ...(finishedAt ? { finishedAt } : {}),
  }
}

/** Normalize a Hermes Workflow projection into safe display summaries. */
export function toCapabilityProjection(value: unknown): CapabilityProjection {
  const root = record(value)
  const nodes = projectionNodes(root.nodes)
  const evidence = projectionEvidence(root.evidence)
  const workflow = projectionWorkflow(root.workflow)
  const workflowRun = projectionRun(root.workflowRun)
  const result = projectionResult(root.result || record(root.workflowRun).outcome)
  const gaps = nodes.flatMap((node) => node.gaps.map((gap) => `${node.title}：${gap}`))
  const failures = nodes.flatMap((node) => node.failure ? [`${node.title}：${node.failure}`] : [])
  const revision = numberOrNull(root.revision)
  const eventCount = numberOrNull(root.eventCount)
  return {
    runId: safeText(root.runId) || undefined,
    profile: safeText(root.profile) || undefined,
    displayType: safeText(root.displayType) || undefined,
    goal: safeText(root.goal) || undefined,
    ...(revision !== null ? { revision } : {}),
    ...(typeof root.hasPlan === 'boolean' ? { hasPlan: root.hasPlan } : {}),
    ...(eventCount !== null ? { eventCount } : {}),
    status: result?.status || workflowRun?.status,
    workflow,
    workflowRun,
    nodes: nodes.length ? nodes : undefined,
    evidence: evidence.length ? evidence : undefined,
    result,
    gaps: gaps.length ? gaps : undefined,
    failures: failures.length ? failures : undefined,
  }
}

/** Normalize the current Hermes Workflow library DTO. */
export function toWorkflow(value: unknown): WorkCapability {
  const raw = record(value)
  const schemaVersion = text(raw.schemaVersion)
  if (schemaVersion !== WORKFLOW_SCHEMA_VERSION) {
    throw new WorkCapabilityApiError('Agent 返回了不受支持的 Workflow 版本。', { code: 'invalid_response' })
  }
  const id = text(raw.id)
  if (!id) throw new WorkCapabilityApiError('Agent Workflow 响应缺少稳定 ID。', { code: 'invalid_response' })

  const status = text(raw.status) || 'unknown'
  const currentVersion = numberOr(raw.version, 1)
  const inputs = Array.isArray(raw.inputs) ? raw.inputs.map(normalizeWorkflowInput) : []
  const steps = Array.isArray(raw.steps) ? raw.steps.map(normalizeWorkflowStep) : []
  const runs = Array.isArray(raw.recentRuns)
    ? raw.recentRuns.map(toWorkflowRun).filter((item) => Boolean(item.runId))
    : []
  const versions = Array.isArray(raw.versions)
    ? raw.versions
      .map((item) => toWorkflowVersion(item, currentVersion))
      .filter((item): item is WorkflowVersion => Boolean(item))
    : []
  const sourceRunId = nullableText(raw.sourceRunId)
  const sourceRun = sourceRunId ? runs.find((item) => item.runId === sourceRunId) || { runId: sourceRunId } : null
  const enabled = status === 'active' ? true : status === 'disabled' || status === 'archived' ? false : null
  const recentFailures = numberOr(raw.recentFailures, 0)
  const lastRunStatus = nullableText(raw.lastRunStatus)
  const reusable = status === 'active'
    && recentFailures < 3
    && !['failed', 'blocked'].includes(lastRunStatus || '')
  const allowedActions = status === 'archived'
    ? ['restore', 'delete']
    : status === 'active' || status === 'disabled'
      ? [
        ...(reusable ? ['use'] : []),
        'edit',
        enabled === true ? 'disable' : 'enable',
        'archive',
        raw.pinned === true ? 'unpin' : 'pin',
        'delete',
      ]
      : []
  const actionBlockers = status === 'active' && !reusable
    ? [{
      action: 'use',
      code: 'workflow_unhealthy',
      reason: recentFailures >= 3
        ? `Agent 已记录 ${recentFailures} 次近期失败，当前不会复用该任务。`
        : `最近一次运行状态为${lastRunStatus === 'blocked' ? '已阻断' : '失败'}，Agent 当前不会复用该任务。`,
    }]
    : []
  const capabilityRefs = steps
    .flatMap((step) => step.capabilityRef ? [step.capabilityRef] : [])
    .filter((item, index, all) => all.indexOf(item) === index)

  const sourceExamples = stringList(raw.examples)

  return {
    schemaVersion,
    id,
    name: safeText(raw.name) || id,
    goal: safeText(raw.goal),
    status,
    enabled,
    pinned: raw.pinned === true,
    currentVersion,
    examples: sourceExamples.map(redactSensitiveText),
    ...(sourceExamples.some((item) => item !== redactSensitiveText(item)) ? { sourceExamples } : {}),
    inputs,
    inputHints: inputs.map((item) => item.name),
    steps,
    notes: stringList(raw.notes, true),
    effects: stringList(raw.effects),
    expectedResult: safeText(raw.expectedResult) || null,
    capabilityRefs,
    sourceRunId,
    sourceRun,
    createdAt: nullableText(raw.createdAt),
    updatedAt: nullableText(raw.updatedAt),
    lastUsedAt: nullableText(raw.lastUsedAt),
    successCount: numberOr(raw.successCount, 0),
    reuseCount: numberOr(raw.useCount, 0),
    recentFailures,
    lastRunStatus,
    successRate: null,
    lastRun: runs[0] || null,
    runs,
    versions,
    allowedActions,
    actionBlockers,
    integrationMode: 'hermes',
  }
}

function workflowListItems(payload: unknown): unknown[] {
  const root = record(payload)
  if (!Array.isArray(root.workflows)) {
    throw new WorkCapabilityApiError('Agent Workflow 列表响应格式无效。', { code: 'invalid_response' })
  }
  return root.workflows
}

function normalizeList(payload: unknown, params: CapabilityListParams): CapabilityListResult {
  const search = text(params.search).toLowerCase()
  const status = text(params.status).toLowerCase()
  const recentRun = text(params.recentRun).toLowerCase()
  const allItems = workflowListItems(payload).map(toWorkflow)
  const filtered = allItems.filter((item) => {
    if (search) {
      const searchable = [
        item.name,
        item.goal,
        ...item.examples,
        ...item.inputHints,
        ...item.steps.flatMap((step) => [step.purpose, step.expectedOutput]),
      ].join('\n').toLowerCase()
      if (!searchable.includes(search)) return false
    }
    if (status && item.status.toLowerCase() !== status) return false
    if (!matchesRecentRun(item, recentRun)) return false
    return true
  })
  const page = positiveInteger(params.page, 1)
  const pageSize = positiveInteger(params.pageSize, 20)
  const start = (page - 1) * pageSize
  return {
    items: filtered.slice(start, start + pageSize),
    total: filtered.length,
    page,
    pageSize,
  }
}

function matchesRecentRun(item: WorkCapability, recentRun: string): boolean {
  if (!recentRun) return true
  if (recentRun === 'never') return !item.lastUsedAt
  const days = recentRun === '7d' ? 7 : recentRun === '30d' ? 30 : 0
  if (!days || !item.lastUsedAt) return false
  const timestamp = Date.parse(item.lastUsedAt)
  return Number.isFinite(timestamp) && timestamp >= Date.now() - days * 24 * 60 * 60 * 1000
}

export class WorkCapabilityApi {
  private readonly request: RequestLike
  private readonly baseUrl: string
  private readonly readUserId: HermesUserIdReader

  constructor(options: WorkCapabilityApiOptions = {}) {
    this.request = options.request || ((input, init) => fetch(input, init))
    this.baseUrl = trimTrailingSlash(options.baseUrl ?? hermesApiBaseUrl())
    this.readUserId = options.readUserId || (() => null)
  }

  async list(params: CapabilityListParams = {}): Promise<CapabilityListResult> {
    const normalizedParams = {
      ...params,
      page: positiveInteger(params.page, 1),
      pageSize: positiveInteger(params.pageSize, 20),
    }
    const payload = await this.getJson(`${WORKFLOW_PREFIX}?includeArchived=true`)
    return normalizeList(payload, normalizedParams)
  }

  async get(id: string): Promise<WorkCapability> {
    const value = await this.getJson(`${WORKFLOW_PREFIX}/${encodeURIComponent(assertId(id))}`)
    return toWorkflow(value)
  }

  async versions(id: string): Promise<WorkflowVersion[]> {
    const detail = await this.get(id)
    return [...detail.versions].sort((left, right) => right.version - left.version)
  }

  async runs(id: string, params: CapabilityRunsParams = {}): Promise<CapabilityRunsResult> {
    assertId(id)
    const normalizedParams = {
      page: positiveInteger(params.page, 1),
      pageSize: positiveInteger(params.pageSize, 10),
      version: optionalPositiveInteger(params.version),
    }
    const detail = await this.get(id)
    const items = detail.runs.filter((item) => (
      normalizedParams.version === undefined || item.version === null || item.version === normalizedParams.version
    ))
    const start = (normalizedParams.page - 1) * normalizedParams.pageSize
    return {
      items: items.slice(start, start + normalizedParams.pageSize),
      total: items.length,
      page: normalizedParams.page,
      pageSize: normalizedParams.pageSize,
    }
  }

  async projection(runId: string): Promise<CapabilityProjection> {
    const value = await this.getJson(`${WORKFLOW_RUN_PREFIX}/${encodeURIComponent(assertId(runId))}`)
    return toCapabilityProjection(value)
  }

  async update(id: string, input: WorkflowUpdateInput): Promise<WorkCapability> {
    const value = await this.sendJson(`${WORKFLOW_PREFIX}/${encodeURIComponent(assertId(id))}`, {
      method: 'PUT',
      body: JSON.stringify(normalizeWorkflowUpdate(input)),
    })
    return toWorkflow(value)
  }

  async setEnabled(id: string, enabled: boolean): Promise<WorkCapability> {
    return this.update(id, { enabled })
  }

  async setPinned(id: string, pinned: boolean): Promise<WorkCapability> {
    const value = await this.sendJson(
      `${WORKFLOW_PREFIX}/${encodeURIComponent(assertId(id))}:${pinned ? 'pin' : 'unpin'}`,
      { method: 'POST' },
    )
    return toWorkflow(value)
  }

  async setArchived(id: string, archived: boolean): Promise<WorkCapability> {
    const value = await this.sendJson(
      `${WORKFLOW_PREFIX}/${encodeURIComponent(assertId(id))}:${archived ? 'archive' : 'restore'}`,
      { method: 'POST' },
    )
    return toWorkflow(value)
  }

  async delete(id: string): Promise<WorkflowDeleteResult> {
    const value = await this.sendJson(`${WORKFLOW_PREFIX}/${encodeURIComponent(assertId(id))}`, {
      method: 'DELETE',
    })
    return normalizeDeleteResult(value)
  }

  async deleteMany(ids: string[]): Promise<WorkflowDeleteResult> {
    const workflowIds = uniqueStrings(ids)
    if (!workflowIds.length || workflowIds.length > 100) {
      throw new WorkCapabilityApiError('请选择 1 至 100 个任务方案。', { code: 'invalid_delete_selection' })
    }
    const value = await this.sendJson(`${WORKFLOW_PREFIX}:delete`, {
      method: 'POST',
      body: JSON.stringify({ workflowIds }),
    })
    return normalizeDeleteResult(value)
  }

  async restore(id: string, version: number): Promise<WorkCapability> {
    const normalizedVersion = optionalPositiveInteger(version)
    if (!normalizedVersion) throw new WorkCapabilityApiError('版本号无效。', { code: 'invalid_version' })
    const value = await this.sendJson(
      `${WORKFLOW_PREFIX}/${encodeURIComponent(assertId(id))}/versions/${normalizedVersion}:restore`,
      { method: 'POST' },
    )
    return toWorkflow(value)
  }

  private async getJson(path: string): Promise<unknown> {
    return this.sendJson(path, { method: 'GET' })
  }

  private async sendJson(path: string, init: RequestInit): Promise<unknown> {
    const userId = requireHermesUserId(this.readUserId())
    let response: Response
    const headers = requestHeaders(Boolean(init.body), init.headers)
    headers.set(HERMES_USER_ID_HEADER, userId)
    try {
      response = await this.request(endpointUrl(this.baseUrl, path), {
        ...init,
        credentials: 'include',
        headers,
      })
    } catch (reason) {
      throw new HermesWorkflowUnavailableError(
        reason instanceof Error ? `Agent 运维任务库暂不可达：${redactSensitiveText(reason.message).slice(0, 500)}` : undefined,
      )
    }

    const payload = await response.json().catch(async () => {
      const body = await response.text().catch(() => '')
      return body ? { message: body } : null
    })
    if (!response.ok) {
      const message = errorMessage(payload, `Agent 运维任务库请求失败（${response.status}）。`)
      const payloadCode = text(record(payload).code || record(record(payload).error).code)
      const code = payloadCode || (
        response.status === 404
          ? 'not_found'
          : response.status === 409
            ? 'conflict'
            : response.status === 422
              ? 'invalid_workflow'
              : response.status === 403
                ? 'permission_denied'
                : undefined
      )
      throw new WorkCapabilityApiError(message, { status: response.status, code })
    }
    return payload
  }
}

function normalizeDeleteResult(value: unknown): WorkflowDeleteResult {
  const raw = record(value)
  const deletedIds = stringList(raw.deletedIds)
  const deletedCount = numberOr(raw.deletedCount, deletedIds.length)
  if (deletedCount !== deletedIds.length) {
    throw new WorkCapabilityApiError('Agent 删除响应不完整。', { code: 'invalid_response' })
  }
  return { deletedIds, deletedCount }
}

function assertId(value: string): string {
  const id = text(value)
  if (!id) throw new WorkCapabilityApiError('Workflow ID 不能为空。', { code: 'invalid_id' })
  return id
}

function positiveInteger(value: unknown, fallback: number): number {
  const candidate = numberOrNull(value)
  return candidate !== null && Number.isInteger(candidate) && candidate > 0 ? candidate : fallback
}

function optionalPositiveInteger(value: unknown): number | undefined {
  const candidate = numberOrNull(value)
  return candidate !== null && Number.isInteger(candidate) && candidate > 0 ? candidate : undefined
}

export function normalizeWorkflowUpdate(input: WorkflowUpdateInput) {
  const source = record(input)
  const body: {
    name?: string
    goal?: string
    examples?: string[]
    inputs?: WorkflowInput[]
    steps?: WorkflowStep[]
    enabled?: boolean
    pinned?: boolean
  } = {}

  if (hasOwn(source, 'name')) {
    const name = text(source.name)
    assertTextLength(name, 2, 120, '任务名称')
    assertNoRedactedPlaceholder(name, '任务名称')
    body.name = name
  }
  if (hasOwn(source, 'goal')) {
    const goal = text(source.goal)
    assertTextLength(goal, 2, 800, '任务目标')
    assertNoRedactedPlaceholder(goal, '任务目标')
    body.goal = goal
  }
  if (hasOwn(source, 'examples')) {
    const examples = uniqueStrings(source.examples)
    if (examples.length > 12) {
      throw new WorkCapabilityApiError('典型提问最多 12 项。', { code: 'invalid_examples' })
    }
    examples.forEach((item) => {
      assertTextLength(item, 1, 500, '典型提问')
      assertNoRedactedPlaceholder(item, '典型提问')
    })
    body.examples = examples
  }
  if (hasOwn(source, 'inputs')) {
    const inputs = Array.isArray(source.inputs) ? source.inputs.map(normalizeEditableInput) : []
    if (inputs.length > 12) {
      throw new WorkCapabilityApiError('输入提示最多 12 项。', { code: 'invalid_inputs' })
    }
    if (new Set(inputs.map((item) => item.name)).size !== inputs.length) {
      throw new WorkCapabilityApiError('输入提示名称不能重复。', { code: 'duplicate_inputs' })
    }
    body.inputs = inputs
  }
  if (hasOwn(source, 'steps')) {
    const steps = Array.isArray(source.steps) ? source.steps.map(normalizeEditableStep) : []
    if (steps.length < 2 || steps.length > 6) {
      throw new WorkCapabilityApiError('稳定步骤需要保留 2 至 6 项。', { code: 'invalid_steps' })
    }
    if (new Set(steps.map((step) => step.id)).size !== steps.length) {
      throw new WorkCapabilityApiError('稳定步骤 ID 不能重复。', { code: 'duplicate_steps' })
    }
    body.steps = steps
  }
  if (hasOwn(source, 'enabled')) {
    if (typeof source.enabled !== 'boolean') {
      throw new WorkCapabilityApiError('自动匹配状态无效。', { code: 'invalid_enabled' })
    }
    body.enabled = source.enabled
  }
  if (hasOwn(source, 'pinned')) {
    if (typeof source.pinned !== 'boolean') {
      throw new WorkCapabilityApiError('固定状态无效。', { code: 'invalid_pinned' })
    }
    body.pinned = source.pinned
  }

  if (!Object.keys(body).length) {
    throw new WorkCapabilityApiError('任务内容没有发生变化。', { code: 'no_changes' })
  }
  return body
}

function normalizeEditableInput(value: unknown): WorkflowInput {
  const raw = record(value)
  const name = text(raw.name)
  const source = text(raw.source) as WorkflowInputSource
  assertTextLength(name, 1, 100, '输入提示')
  assertNoRedactedPlaceholder(name, '输入提示')
  if (!WORKFLOW_INPUT_SOURCES.has(source)) {
    throw new WorkCapabilityApiError('输入提示来源无效。', { code: 'invalid_input_source' })
  }
  return { name, required: raw.required === true, source }
}

function normalizeEditableStep(value: unknown): WorkflowStep {
  const raw = record(value)
  const id = text(raw.id)
  const purpose = text(raw.purpose)
  const expectedOutput = text(raw.expectedOutput)
  const providerRef = nullableText(raw.providerRef)
  const capabilityRef = nullableText(raw.capabilityRef)
  assertTextLength(id, 1, 100, '步骤 ID')
  assertTextLength(purpose, 1, 500, '步骤内容')
  assertTextLength(expectedOutput, 1, 500, '步骤预期输出')
  assertNoRedactedPlaceholder(purpose, '步骤内容')
  assertNoRedactedPlaceholder(expectedOutput, '步骤预期输出')
  if (Boolean(providerRef) !== Boolean(capabilityRef)) {
    throw new WorkCapabilityApiError('步骤的 Provider 和 Capability 绑定不完整。', { code: 'invalid_step_binding' })
  }
  if (!isRecord(raw.inputMap)) {
    throw new WorkCapabilityApiError('步骤输入映射无效。', { code: 'invalid_step_input_map' })
  }
  const effects = uniqueStrings(raw.effects)
  if (effects.length > 12) {
    throw new WorkCapabilityApiError('每个步骤最多保留 12 个作用类型。', { code: 'invalid_step_effects' })
  }
  return {
    id,
    purpose,
    providerRef,
    capabilityRef,
    inputMap: { ...raw.inputMap },
    expectedOutput,
    effects,
  }
}

export function toWorkflowEditValues(capability: WorkCapability): WorkflowEditValues {
  return {
    name: capability.name,
    goal: capability.goal,
    examples: [...capability.examples],
    ...(capability.sourceExamples ? { sourceExamples: [...capability.sourceExamples] } : {}),
    inputs: capability.inputs.map((input) => ({ ...input })),
    steps: capability.steps.map((step) => ({
      ...step,
      inputMap: { ...step.inputMap },
      effects: [...step.effects],
    })),
  }
}

export function diffWorkflowUpdate(
  initial: WorkflowEditValues,
  current: WorkflowEditValues,
): WorkflowUpdateInput {
  const before = normalizeEditValues(initial)
  const after = normalizeEditValues(current)
  const update: WorkflowUpdateInput = {}
  if (before.name !== after.name) update.name = after.name
  if (before.goal !== after.goal) update.goal = after.goal
  if (!sameValue(before.examples, after.examples)) update.examples = restoreExampleValues(initial, after.examples)
  if (!sameValue(before.inputs, after.inputs)) update.inputs = restoreInputValues(initial, after.inputs)
  if (!sameValue(before.steps, after.steps)) update.steps = restoreStepValues(initial, after.steps)
  return update
}

function restoreExampleValues(initial: WorkflowEditValues, current: string[]): string[] {
  const sourceValues = initial.sourceExamples || initial.examples
  const used = new Set<number>()
  return current.map((value) => {
    const index = initial.examples.findIndex((candidate, candidateIndex) => candidate === value && !used.has(candidateIndex))
    if (index < 0) return value
    used.add(index)
    return sourceValues[index] || value
  })
}

function restoreInputValues(initial: WorkflowEditValues, current: WorkflowInput[]): WorkflowInput[] {
  return current.map((input) => {
    const previous = initial.inputs.find((candidate) => candidate.name === input.name)
    return {
      ...input,
      name: previous?.sourceName || input.name,
    }
  })
}

function restoreStepValues(initial: WorkflowEditValues, current: WorkflowStep[]): WorkflowStep[] {
  return current.map((step) => {
    const previous = initial.steps.find((candidate) => candidate.id === step.id)
    return {
      ...step,
      purpose: previous && previous.purpose === step.purpose
        ? previous.sourcePurpose || step.purpose
        : step.purpose,
      expectedOutput: previous && previous.expectedOutput === step.expectedOutput
        ? previous.sourceExpectedOutput || step.expectedOutput
        : step.expectedOutput,
    }
  })
}

function uniqueStrings(value: unknown): string[] {
  if (!Array.isArray(value)) return []
  return [...new Set(value.map(text).filter(Boolean))]
}

function normalizeEditValues(value: WorkflowEditValues): WorkflowEditValues {
  const source = record(value)
  return {
    name: text(source.name),
    goal: text(source.goal),
    examples: uniqueStrings(source.examples),
    inputs: Array.isArray(source.inputs)
      ? source.inputs.map((input) => ({
        name: text(input?.name),
        required: input?.required === true,
        source: text(input?.source) as WorkflowInputSource,
      }))
      : [],
    steps: Array.isArray(source.steps)
      ? source.steps.map((step) => ({
        id: text(step?.id),
        purpose: text(step?.purpose),
        providerRef: nullableText(step?.providerRef),
        capabilityRef: nullableText(step?.capabilityRef),
        inputMap: isRecord(step?.inputMap) ? { ...step.inputMap } : {},
        expectedOutput: text(step?.expectedOutput),
        effects: uniqueStrings(step?.effects),
      }))
      : [],
  }
}

function sameValue(left: unknown, right: unknown): boolean {
  return JSON.stringify(left) === JSON.stringify(right)
}

function hasOwn(value: Record<string, unknown>, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(value, key)
}

function assertNoRedactedPlaceholder(value: string, label: string) {
  if (value.includes('[已隐藏]')) {
    throw new WorkCapabilityApiError(
      `${label}包含已脱敏内容，请完整替换该内容后再保存。`,
      { code: 'redacted_value' },
    )
  }
}

function assertTextLength(value: string, min: number, max: number, label: string) {
  if (value.length < min || value.length > max) {
    throw new WorkCapabilityApiError(`${label}长度需要在 ${min} 至 ${max} 个字符之间。`, { code: 'invalid_update' })
  }
}


// Route/page names still use "capabilities" for URL compatibility.
export const toCapability = toWorkflow
export const toRun = toWorkflowRun
export const toVersion = toWorkflowVersion
