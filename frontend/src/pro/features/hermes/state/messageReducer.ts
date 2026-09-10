import { checkpointForTimeline, checkpointFromPayload, contentText } from '../client/history.ts'
import {
  isInterruptedTerminalEvent,
  isRunScopedEventType,
  runIdFromEventPayload,
} from '../client/eventPayload.ts'
import { isPlanToolPayload, planFingerprint, planFromToolPayload } from '../client/plan.ts'
import type {
  HermesActivity,
  HermesCheckpoint,
  HermesEvent,
  HermesHistorySnapshot,
  HermesMessage,
  HermesPlan,
  HermesRunStatus,
  HermesRunStatusTimelineItem,
  HermesTimelineItem,
  HermesToolActivity,
} from '../client/types.ts'

export interface HermesMessageState {
  timeline: HermesTimelineItem[]
  reasoning: string
  checkpoint: HermesCheckpoint | null
  busy: boolean
  stopping: boolean
  stopRetryAllowed: boolean
  statusText: string
  error: string | null
  activeRunId: string | null
  nextSequenceNo: number
  hydrationVersion: number
}

export type HermesMessageAction =
  | { type: 'RESET' }
  | { type: 'HYDRATE'; snapshot: HermesHistorySnapshot; busy?: boolean; statusText?: string }
  | { type: 'SUBMIT_STARTED'; text: string; runId: string }
  | { type: 'USER_ADDITION'; text: string; runId: string | null }
  | { type: 'CHECKPOINT_RESPONDED'; text: string }
  | { type: 'EVENT_RECEIVED'; event: HermesEvent }
  | { type: 'SEND_FAILED'; error: string }
  | { type: 'STOP_REQUESTED' }
  | { type: 'STOP_ACKNOWLEDGED' }
  | { type: 'STOP_RETRY_AVAILABLE' }
  | { type: 'STOP_FAILED'; error: string; runId: string | null }
  | { type: 'SET_STATUS'; statusText: string; busy?: boolean }
  | { type: 'CLEAR_ERROR' }

export const initialHermesMessageState: HermesMessageState = {
  timeline: [],
  reasoning: '',
  checkpoint: null,
  busy: false,
  stopping: false,
  stopRetryAllowed: false,
  statusText: '准备就绪',
  error: null,
  activeRunId: null,
  nextSequenceNo: 1,
  hydrationVersion: 0,
}

export function hermesMessageReducer(
  state: HermesMessageState,
  action: HermesMessageAction,
): HermesMessageState {
  if (action.type === 'RESET') {
    return { ...initialHermesMessageState, hydrationVersion: state.hydrationVersion + 1 }
  }
  if (action.type === 'HYDRATE') {
    const timeline = ensureCheckpointTimeline(action.snapshot.timeline, action.snapshot.checkpoint)
    return {
      ...initialHermesMessageState,
      timeline,
      checkpoint: action.snapshot.checkpoint,
      activeRunId: action.snapshot.activeRunId,
      busy: action.busy ?? false,
      statusText:
        action.statusText || (action.snapshot.checkpoint ? '等待用户回复' : '历史记录已加载'),
      nextSequenceNo: nextSequence(timeline),
      hydrationVersion: state.hydrationVersion + 1,
    }
  }
  if (action.type === 'SUBMIT_STARTED') {
    const occurredAt = new Date().toISOString()
    const userMessage = createMessage('user', action.text, action.runId, true)
    return {
      ...state,
      timeline: [
        ...state.timeline,
        messageTimelineItem(userMessage, state.nextSequenceNo, occurredAt, action.runId),
        {
          id: `run-status-${action.runId}`,
          kind: 'run_status',
          sequenceNo: state.nextSequenceNo + 1,
          occurredAt,
          runId: action.runId,
          status: 'running',
          label: '正在理解问题并制定分析步骤',
          detail: '任务已开始',
          startedAt: occurredAt,
        },
      ],
      reasoning: '',
      checkpoint: null,
      busy: true,
      stopping: false,
      stopRetryAllowed: false,
      statusText: '正在理解问题',
      error: null,
      activeRunId: action.runId,
      nextSequenceNo: state.nextSequenceNo + 2,
    }
  }
  if (action.type === 'USER_ADDITION') {
    const item = messageTimelineItem(
      createMessage('user', action.text, action.runId, true),
      state.nextSequenceNo,
      new Date().toISOString(),
      action.runId,
    )
    return {
      ...state,
      timeline: [...state.timeline, item],
      statusText: '正在处理补充要求',
      error: null,
      nextSequenceNo: state.nextSequenceNo + 1,
    }
  }
  if (action.type === 'CHECKPOINT_RESPONDED') {
    const occurredAt = new Date().toISOString()
    const timeline = resolveLatestCheckpoint(state.timeline, action.text)
    const item = messageTimelineItem(
      createMessage('user', action.text, state.activeRunId, true),
      state.nextSequenceNo,
      occurredAt,
      state.activeRunId,
    )
    return {
      ...state,
      timeline: updateRunStatus([...timeline, item], state.activeRunId, {
        status: 'running',
        label: '已收到确认，正在继续分析',
        detail: '继续执行当前任务',
      }),
      checkpoint: null,
      busy: true,
      stopping: false,
      stopRetryAllowed: false,
      statusText: '正在继续处理',
      error: null,
      nextSequenceNo: state.nextSequenceNo + 1,
    }
  }
  if (action.type === 'EVENT_RECEIVED') return applyHermesEvent(state, action.event)
  if (action.type === 'SEND_FAILED') {
    const occurredAt = new Date().toISOString()
    return {
      ...state,
      timeline: finishRunStatus(
        state.timeline,
        state.activeRunId,
        occurredAt,
        'failed',
        '消息发送失败',
        action.error,
      ),
      busy: false,
      stopping: false,
      stopRetryAllowed: false,
      checkpoint: null,
      activeRunId: null,
      statusText: '发送失败',
      error: action.error,
    }
  }
  if (action.type === 'STOP_REQUESTED') {
    return {
      ...state,
      stopping: true,
      stopRetryAllowed: false,
      statusText: '正在停止当前任务',
      error: null,
      timeline: updateRunStatus(state.timeline, state.activeRunId, {
        status: 'stopping',
        label: '正在停止当前任务',
        detail: '已请求停止，正在等待任务确认',
      }),
    }
  }
  if (action.type === 'STOP_ACKNOWLEDGED') {
    return {
      ...state,
      stopping: true,
      statusText: '已请求停止，等待任务结束',
      timeline: updateRunStatus(state.timeline, state.activeRunId, {
        status: 'stopping',
        label: '停止请求已发送',
        detail: '正在等待当前分析安全结束',
      }),
    }
  }
  if (action.type === 'STOP_RETRY_AVAILABLE') {
    if (!state.busy || !state.stopping) return state
    return {
      ...state,
      stopRetryAllowed: true,
      statusText: '停止请求尚未确认，可重试',
      timeline: updateRunStatus(state.timeline, state.activeRunId, {
        status: 'stopping',
        label: '停止请求尚未确认',
        detail: '可以再次发送停止请求，或等待当前步骤结束',
      }),
    }
  }
  if (action.type === 'STOP_FAILED') {
    if (!state.busy || (action.runId && state.activeRunId !== action.runId)) return state
    return {
      ...state,
      stopping: false,
      stopRetryAllowed: true,
      statusText: '停止失败，可重试',
      error: action.error,
      timeline: updateRunStatus(state.timeline, state.activeRunId, {
        status: 'running',
        label: '任务仍在运行',
        detail: '停止请求未成功，可重试',
      }),
    }
  }
  if (action.type === 'SET_STATUS') {
    return {
      ...state,
      statusText: action.statusText,
      ...(typeof action.busy === 'boolean' ? { busy: action.busy } : {}),
    }
  }
  if (action.type === 'CLEAR_ERROR') return { ...state, error: null }
  return state
}

export function applyHermesEvent(
  state: HermesMessageState,
  event: HermesEvent,
): HermesMessageState {
  if (shouldIgnoreEventAfterInterrupt(state, event)) return state
  const payload = event.payload
  const runId = runIdFromEventPayload(event.payload) || state.activeRunId
  const occurredAt = eventOccurredAt(event)

  if (event.type === 'message.delta') {
    const delta = contentText(payload.delta ?? payload.text ?? payload.content)
    if (!delta) return state
    const flushed = flushReasoning(state, occurredAt, runId)
    const appended = appendAssistant(
      flushed.timeline,
      delta,
      flushed.nextSequenceNo,
      occurredAt,
      runId,
    )
    return {
      ...flushed,
      timeline: updateRunStatus(appended.timeline, runId, {
        status: 'running',
        label: '正在整理分析结论',
        detail: '已完成主要查询，正在生成回答',
      }),
      statusText: '正在整理结论',
      nextSequenceNo: appended.nextSequenceNo,
    }
  }
  if (event.type === 'message.complete') {
    const content = contentText(payload.text ?? payload.content)
    const interrupted = isInterruptedTerminalEvent(event.type, payload)
    const flushed = flushReasoning(state, occurredAt, runId)
    const completed = completeAssistant(
      flushed.timeline,
      content,
      flushed.nextSequenceNo,
      occurredAt,
      runId,
      interrupted,
    )
    return {
      ...flushed,
      timeline: finishRunStatus(
        completed.timeline,
        runId,
        occurredAt,
        interrupted ? 'interrupted' : 'completed',
        interrupted ? '任务已停止' : '本轮分析已完成',
        interrupted ? '已保留停止前产生的内容' : '可以继续追问或补充分析范围',
      ),
      nextSequenceNo: completed.nextSequenceNo,
      busy: false,
      stopping: false,
      stopRetryAllowed: false,
      checkpoint: null,
      activeRunId: null,
      statusText: interrupted ? '处理已中断' : '已完成',
      error: null,
    }
  }
  if (event.type === 'reasoning.delta') {
    const delta = contentText(payload.delta ?? payload.text ?? payload.content)
    return delta
      ? {
          ...state,
          reasoning: state.reasoning + delta,
          statusText: '正在分析问题',
          timeline: updateRunStatus(state.timeline, runId, {
            status: 'running',
            label: '正在分析问题',
            detail: '梳理现象、范围和可能原因',
          }),
        }
      : state
  }
  if (event.type === 'message.interim') {
    const text = contentText(payload.message ?? payload.text ?? payload.content).trim()
    const flushed = flushReasoning(state, occurredAt, runId)
    if (!text) return flushed
    return {
      ...flushed,
      timeline: [
        ...flushed.timeline,
        activityTimelineItem(
          { id: createId('commentary'), kind: 'commentary', text },
          flushed.nextSequenceNo,
          occurredAt,
          runId,
        ),
      ],
      nextSequenceNo: flushed.nextSequenceNo + 1,
    }
  }
  if (isToolEvent(event.type)) {
    const flushed = flushReasoning(state, occurredAt, runId)
    if (isPlanToolPayload(payload)) {
      const plan = isCompletedToolEvent(event.type) ? planFromToolPayload(payload) : null
      if (!plan) return flushed
      const appended = appendPlan(flushed.timeline, plan, flushed.nextSequenceNo, occurredAt, runId)
      return {
        ...flushed,
        timeline: updateRunStatus(appended.timeline, runId, {
          status: 'running',
          label: '分析方案已生成',
          detail: '请确认方案，或提出修改要求',
        }),
        nextSequenceNo: appended.nextSequenceNo,
        statusText: '分析方案已生成',
      }
    }
    const updated = upsertTool(flushed.timeline, event, flushed.nextSequenceNo, occurredAt, runId)
    return {
      ...flushed,
      timeline: updateRunStatus(updated.timeline, runId, {
        status: 'running',
        label: '正在查询和校验数据',
        detail: event.type === 'tool.failed' ? '部分查询未成功，正在继续处理' : '分析步骤正在执行',
      }),
      nextSequenceNo: updated.nextSequenceNo,
      statusText: event.type === 'tool.failed' ? '部分查询失败，继续处理中' : '正在查询数据',
    }
  }
  if (event.type === 'approval.request' || event.type === 'clarify.request') {
    const flushed = flushReasoning(state, occurredAt, runId)
    const checkpoint = checkpointForTimeline(
      checkpointFromPayload(event.type, payload, runId),
      flushed.timeline,
    )
    const timeline: HermesTimelineItem[] = [
      ...flushed.timeline,
      {
        id: createId('timeline-checkpoint'),
        kind: 'checkpoint',
        sequenceNo: flushed.nextSequenceNo,
        occurredAt,
        runId,
        checkpoint,
        resolved: false,
      },
    ]
    return {
      ...flushed,
      timeline: updateRunStatus(timeline, runId, {
        status: 'waiting',
        label: event.type === 'approval.request' ? '等待确认后继续' : '等待补充信息',
        detail: '当前任务已暂停，不会继续执行后续步骤',
      }),
      nextSequenceNo: flushed.nextSequenceNo + 1,
      checkpoint,
      busy: false,
      stopping: false,
      stopRetryAllowed: false,
      statusText: event.type === 'approval.request' ? '等待确认' : '等待补充信息',
    }
  }
  if (event.type === 'run.completed') {
    return terminalState(state, runId, occurredAt, 'completed', '已完成', null)
  }
  if (event.type === 'run.failed' || event.type === 'error') {
    return terminalState(
      state,
      runId,
      occurredAt,
      'failed',
      '运行失败',
      contentText(payload.message) || '本次处理未能完成，请稍后重试。',
    )
  }
  if (isInterruptedTerminalEvent(event.type, payload)) {
    return terminalState(state, runId, occurredAt, 'interrupted', '处理已中断', null)
  }
  return state
}

function terminalState(
  state: HermesMessageState,
  runId: string | null,
  occurredAt: string,
  status: Extract<HermesRunStatus, 'completed' | 'failed' | 'interrupted'>,
  statusText: string,
  error: string | null,
): HermesMessageState {
  const labels = {
    completed: ['本轮分析已完成', '可以继续追问或补充分析范围'],
    failed: ['本轮分析未完成', error || '运行过程中发生异常'],
    interrupted: ['任务已停止', '已保留停止前产生的内容'],
  }
  return {
    ...state,
    timeline: finishRunStatus(
      state.timeline,
      runId,
      occurredAt,
      status,
      labels[status][0],
      labels[status][1],
    ),
    reasoning: '',
    busy: false,
    stopping: false,
    stopRetryAllowed: false,
    checkpoint: null,
    activeRunId: null,
    statusText,
    error,
  }
}

function flushReasoning(
  state: HermesMessageState,
  occurredAt: string,
  runId: string | null,
): HermesMessageState {
  const text = state.reasoning.trim()
  if (!text) return { ...state, reasoning: '' }
  const activity: HermesActivity = { id: createId('commentary'), kind: 'commentary', text }
  return {
    ...state,
    timeline: [
      ...state.timeline,
      activityTimelineItem(activity, state.nextSequenceNo, occurredAt, runId),
    ],
    reasoning: '',
    nextSequenceNo: state.nextSequenceNo + 1,
  }
}

function appendAssistant(
  timeline: HermesTimelineItem[],
  delta: string,
  sequenceNo: number,
  occurredAt: string,
  runId: string | null,
): { timeline: HermesTimelineItem[]; nextSequenceNo: number } {
  const index = findStreamingAssistant(timeline, runId)
  if (index >= 0) {
    const next = [...timeline]
    const item = next[index]
    if (item.kind === 'message') {
      next[index] = {
        ...item,
        message: { ...item.message, content: item.message.content + delta },
      }
    }
    return { timeline: next, nextSequenceNo: sequenceNo }
  }
  const message = createMessage('assistant', delta, runId, false)
  return {
    timeline: [...timeline, messageTimelineItem(message, sequenceNo, occurredAt, runId)],
    nextSequenceNo: sequenceNo + 1,
  }
}

function completeAssistant(
  timeline: HermesTimelineItem[],
  content: string,
  sequenceNo: number,
  occurredAt: string,
  runId: string | null,
  interrupted: boolean,
): { timeline: HermesTimelineItem[]; nextSequenceNo: number } {
  const index = findStreamingAssistant(timeline, runId)
  if (index >= 0) {
    const next = [...timeline]
    const item = next[index]
    if (item.kind === 'message') {
      if (interrupted && !content && !item.message.content) next.splice(index, 1)
      else {
        next[index] = {
          ...item,
          message: { ...item.message, content: content || item.message.content, complete: true },
        }
      }
    }
    return { timeline: next, nextSequenceNo: sequenceNo }
  }
  if (!content) return { timeline, nextSequenceNo: sequenceNo }
  const message = createMessage('assistant', content, runId, true)
  return {
    timeline: [...timeline, messageTimelineItem(message, sequenceNo, occurredAt, runId)],
    nextSequenceNo: sequenceNo + 1,
  }
}

function findStreamingAssistant(timeline: HermesTimelineItem[], runId: string | null): number {
  for (let index = timeline.length - 1; index >= 0; index -= 1) {
    const item = timeline[index]
    if (
      item.kind === 'message' &&
      item.message.role === 'assistant' &&
      item.message.complete === false &&
      (!runId || item.runId === runId)
    ) {
      return index
    }
  }
  return -1
}

function upsertTool(
  timeline: HermesTimelineItem[],
  event: HermesEvent,
  sequenceNo: number,
  occurredAt: string,
  runId: string | null,
): { timeline: HermesTimelineItem[]; nextSequenceNo: number } {
  const payload = event.payload
  const id = String(payload.tool_id ?? payload.id ?? createId('tool'))
  const completed = event.type === 'tool.complete' || event.type === 'tool.completed'
  const failed = event.type === 'tool.failed'
  const rawArgs = payload.args ?? payload.arguments
  const nextTool: HermesToolActivity = {
    id,
    kind: 'tool',
    name: String(payload.name ?? payload.tool ?? 'tool'),
    status: failed ? 'failed' : completed ? 'completed' : 'running',
    context: String(payload.context ?? payload.message ?? ''),
    args: isObject(rawArgs) ? rawArgs : undefined,
    result: payload.result ?? payload.output,
    evidence: Array.isArray(payload.evidence)
      ? (payload.evidence as HermesToolActivity['evidence'])
      : undefined,
    durationMs: typeof payload.duration_ms === 'number' ? payload.duration_ms : undefined,
  }
  const index = timeline.findIndex(
    (item) =>
      item.kind === 'activity' &&
      item.activity.kind === 'tool' &&
      item.activity.id === id &&
      (!runId || item.runId === runId),
  )
  if (index < 0) {
    return {
      timeline: [...timeline, activityTimelineItem(nextTool, sequenceNo, occurredAt, runId)],
      nextSequenceNo: sequenceNo + 1,
    }
  }
  const next = [...timeline]
  const previous = next[index]
  if (previous.kind === 'activity' && previous.activity.kind === 'tool') {
    next[index] = { ...previous, activity: { ...previous.activity, ...nextTool } }
  }
  return { timeline: next, nextSequenceNo: sequenceNo }
}

function appendPlan(
  timeline: HermesTimelineItem[],
  plan: HermesPlan,
  sequenceNo: number,
  occurredAt: string,
  runId: string | null,
): { timeline: HermesTimelineItem[]; nextSequenceNo: number } {
  const fingerprint = planFingerprint(plan)
  const exists = timeline.some(
    (item) =>
      item.kind === 'plan' && item.runId === runId && planFingerprint(item.plan) === fingerprint,
  )
  if (exists) return { timeline, nextSequenceNo: sequenceNo }
  return {
    timeline: [
      ...timeline,
      {
        id: createId('timeline-plan'),
        kind: 'plan',
        sequenceNo,
        occurredAt,
        runId,
        plan,
      },
    ],
    nextSequenceNo: sequenceNo + 1,
  }
}

function updateRunStatus(
  timeline: HermesTimelineItem[],
  runId: string | null,
  patch: Partial<HermesRunStatusTimelineItem>,
): HermesTimelineItem[] {
  if (!runId) return timeline
  for (let index = timeline.length - 1; index >= 0; index -= 1) {
    const item = timeline[index]
    if (item.kind === 'run_status' && item.runId === runId) {
      if (item.status === 'interrupted') return timeline
      const next = [...timeline]
      next[index] = { ...item, ...patch }
      return next
    }
  }
  return timeline
}

function finishRunStatus(
  timeline: HermesTimelineItem[],
  runId: string | null,
  occurredAt: string,
  status: HermesRunStatus,
  label: string,
  detail: string,
): HermesTimelineItem[] {
  if (!runId) return timeline
  for (let index = timeline.length - 1; index >= 0; index -= 1) {
    const item = timeline[index]
    if (item.kind === 'run_status' && item.runId === runId) {
      if (item.status === 'interrupted' && status !== 'interrupted') return timeline
      const startedAt = item.startedAt || item.occurredAt
      const durationMs = Math.max(0, new Date(occurredAt).getTime() - new Date(startedAt).getTime())
      const next = [...timeline]
      next[index] = {
        ...item,
        status,
        label,
        detail,
        finishedAt: occurredAt,
        durationMs: Number.isFinite(durationMs) ? durationMs : undefined,
      }
      return next
    }
  }
  return timeline
}

export function shouldIgnoreEventAfterInterrupt(
  state: HermesMessageState,
  event: HermesEvent,
): boolean {
  const runId = runIdFromEventPayload(event.payload)
  if (!isRunScopedEventType(event.type)) return false

  if (runId) {
    const matchingRunStatus = [...state.timeline]
      .reverse()
      .find(
        (item): item is HermesRunStatusTimelineItem =>
          item.kind === 'run_status' && item.runId === runId,
      )
    if (matchingRunStatus?.status === 'interrupted') return true
    return Boolean(state.activeRunId && state.activeRunId !== runId)
  }

  if (state.activeRunId) return false
  const latestRunStatus = [...state.timeline]
    .reverse()
    .find((item): item is HermesRunStatusTimelineItem => item.kind === 'run_status')
  return latestRunStatus?.status === 'interrupted'
}

function resolveLatestCheckpoint(
  timeline: HermesTimelineItem[],
  responseText: string,
): HermesTimelineItem[] {
  for (let index = timeline.length - 1; index >= 0; index -= 1) {
    const item = timeline[index]
    if (item.kind === 'checkpoint' && !item.resolved) {
      const next = [...timeline]
      next[index] = { ...item, resolved: true, responseText }
      return next
    }
  }
  return timeline
}

function ensureCheckpointTimeline(
  timeline: HermesTimelineItem[],
  checkpoint: HermesCheckpoint | null,
): HermesTimelineItem[] {
  if (!checkpoint || timeline.some((item) => item.kind === 'checkpoint' && !item.resolved)) {
    return timeline
  }
  return [
    ...timeline,
    {
      id: createId('timeline-checkpoint'),
      kind: 'checkpoint',
      sequenceNo: nextSequence(timeline),
      occurredAt: new Date().toISOString(),
      runId: null,
      checkpoint: checkpointForTimeline(checkpoint, timeline),
      resolved: false,
    },
  ]
}

function messageTimelineItem(
  message: HermesMessage,
  sequenceNo: number,
  occurredAt: string,
  runId: string | null,
): HermesTimelineItem {
  return {
    id: `timeline-${message.id}`,
    kind: 'message',
    sequenceNo,
    occurredAt,
    runId,
    message,
  }
}

function activityTimelineItem(
  activity: HermesActivity,
  sequenceNo: number,
  occurredAt: string,
  runId: string | null,
): HermesTimelineItem {
  return {
    id: `timeline-${activity.id}`,
    kind: 'activity',
    sequenceNo,
    occurredAt,
    runId,
    activity,
  }
}

function createMessage(
  role: HermesMessage['role'],
  content: string,
  runId: string | null,
  complete: boolean,
): HermesMessage {
  return {
    id: createId('message'),
    role,
    content,
    complete,
    ...(runId ? { runId } : {}),
  }
}

function eventOccurredAt(event: HermesEvent): string {
  const value = event.payload.occurred_at ?? event.payload.occurredAt
  return typeof value === 'string' && value ? value : new Date().toISOString()
}

function nextSequence(timeline: HermesTimelineItem[]): number {
  return Math.floor(Math.max(0, ...timeline.map((item) => item.sequenceNo))) + 1
}

function isToolEvent(type: string): boolean {
  return ['tool.start', 'tool.started', 'tool.complete', 'tool.completed', 'tool.failed'].includes(
    type,
  )
}

function isCompletedToolEvent(type: string): boolean {
  return type === 'tool.complete' || type === 'tool.completed'
}

function createId(prefix: string): string {
  return `${prefix}-${typeof crypto.randomUUID === 'function' ? crypto.randomUUID() : Date.now()}`
}

function isObject(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
