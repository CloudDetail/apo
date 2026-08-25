import type {
  HermesActivity,
  HermesCheckpoint,
  HermesCheckpointTimelineItem,
  HermesHistorySnapshot,
  HermesMessage,
  HermesPlan,
  HermesPlanTimelineItem,
  HermesRunStatus,
  HermesRunStatusTimelineItem,
  HermesTimelineItem,
  HermesToolActivity,
  PersistedConversationEvent,
} from './types.ts'
import {
  isInterruptedTerminalEvent,
  isRunScopedEventType,
  runIdFromEventPayload,
  runStartEventDecision,
} from './eventPayload.ts'
import { isPlanToolPayload, planFingerprint, planFromToolPayload } from './plan.ts'

export function historyFromPersistedEvents(
  records: PersistedConversationEvent[],
): HermesHistorySnapshot {
  const timeline: HermesTimelineItem[] = []
  const toolIndexes = new Map<string, number>()
  const planFingerprints = new Set<string>()
  const runStatusIndexes = new Map<string, number>()
  const interruptedRunIds = new Set<string>()
  let checkpoint: HermesCheckpoint | null = null
  let activeRunId: string | null = null
  let interruptedRunAwaitingNextStart = false
  let pendingRunStartId: string | null = null
  let reasoning = ''
  let reasoningRecord: PersistedConversationEvent | null = null

  const flushReasoning = () => {
    const text = reasoning.trim()
    reasoning = ''
    if (!text || !reasoningRecord) return
    pushActivity(
      timeline,
      { id: createId('commentary'), kind: 'commentary', text },
      reasoningRecord,
    )
    reasoningRecord = null
  }

  for (const record of [...records].sort((left, right) => left.sequenceNo - right.sequenceNo)) {
    const payload = record.payload
    const runId = record.runId || activeRunId
    const payloadRunId = runIdFromEventPayload(payload)

    if (record.direction === 'browser_to_gateway') {
      if (record.eventType === 'prompt.submit') {
        activeRunId = record.runId
        checkpoint = null
        if (interruptedRunAwaitingNextStart) {
          pendingRunStartId = record.runId || payloadRunId
          interruptedRunAwaitingNextStart = false
        }
        const content = visibleMessageText(contentText(payload.text))
        if (content) pushMessage(timeline, 'user', content, record, true)
        if (record.runId) {
          pushRunStatus(timeline, runStatusIndexes, record.runId, record, {
            status: 'running',
            label: '正在理解问题并制定分析步骤',
            detail: '任务已开始',
            startedAt: record.occurredAt,
          })
        }
      } else if (record.eventType === 'session.steer') {
        const content = contentText(payload.text)
        if (content) pushMessage(timeline, 'user', content, record, true)
      } else if (
        record.eventType === 'approval.respond' ||
        record.eventType === 'clarify.respond'
      ) {
        const content = contentText(payload.label ?? payload.answer ?? payload.choice)
        if (content) pushMessage(timeline, 'user', content, record, true)
        resolveLatestCheckpoint(timeline, content)
        checkpoint = null
        if (runId) {
          updateRunStatus(timeline, runStatusIndexes, runId, {
            status: 'running',
            label: '已收到确认，正在继续分析',
            detail: '继续执行当前任务',
          })
        }
      }
      continue
    }

    const runStartDecision = runStartEventDecision(
      pendingRunStartId,
      record.eventType,
      payloadRunId,
    )
    if (runStartDecision === 'ignore') continue
    if (runStartDecision === 'clear') pendingRunStartId = null

    const interruptedTerminal = isInterruptedTerminalEvent(record.eventType, payload)
    const eventIdentityRunId = payloadRunId || runId
    if (
      eventIdentityRunId &&
      isRunScopedEventType(record.eventType) &&
      interruptedRunIds.has(eventIdentityRunId)
    ) continue

    if (record.eventType === 'message.delta') {
      flushReasoning()
      appendAssistant(
        timeline,
        contentText(payload.delta ?? payload.text ?? payload.content),
        record,
      )
      if (runId) {
        updateRunStatus(timeline, runStatusIndexes, runId, {
          status: 'running',
          label: '正在整理分析结论',
          detail: '已完成主要查询，正在生成回答',
        })
      }
      continue
    }
    if (record.eventType === 'message.complete') {
      flushReasoning()
      completeAssistant(timeline, contentText(payload.text ?? payload.content), record)
      const interrupted = interruptedTerminal
      const interruptedActiveRun = interrupted && Boolean(runId) && activeRunId === runId
      if (runId) {
        finishRunStatus(
          timeline,
          runStatusIndexes,
          runId,
          record,
          interrupted ? 'interrupted' : 'completed',
          interrupted ? '任务已停止' : '本轮分析已完成',
          interrupted ? '已保留停止前产生的内容' : '可以继续追问或补充分析范围',
        )
        if (interrupted) interruptedRunIds.add(runId)
      }
      if (interruptedActiveRun) interruptedRunAwaitingNextStart = true
      if (activeRunId === runId) {
        activeRunId = null
        checkpoint = null
      }
      continue
    }
    if (record.eventType === 'reasoning.delta') {
      reasoningRecord ||= record
      reasoning += contentText(payload.delta ?? payload.text ?? payload.content)
      if (runId) {
        updateRunStatus(timeline, runStatusIndexes, runId, {
          status: 'running',
          label: '正在分析问题',
          detail: '梳理现象、范围和可能原因',
        })
      }
      continue
    }
    if (record.eventType === 'message.interim') {
      flushReasoning()
      const text = contentText(payload.message ?? payload.text ?? payload.content)
      if (text) {
        pushActivity(timeline, { id: createId('commentary'), kind: 'commentary', text }, record)
      }
      continue
    }
    if (isToolEvent(record.eventType)) {
      flushReasoning()
      if (isPlanToolPayload(payload)) {
        if (isCompletedToolEvent(record.eventType)) {
          const plan = planFromToolPayload(payload)
          if (plan) pushPlan(timeline, planFingerprints, plan, record, runId)
        }
        continue
      }
      upsertTool(timeline, toolIndexes, record)
      if (runId) {
        updateRunStatus(timeline, runStatusIndexes, runId, {
          status: 'running',
          label: '正在查询和校验数据',
          detail:
            record.eventType === 'tool.failed'
              ? '部分查询未成功，正在继续处理'
              : '分析步骤正在执行',
        })
      }
      continue
    }
    if (record.eventType === 'approval.request' || record.eventType === 'clarify.request') {
      flushReasoning()
      checkpoint = checkpointForTimeline(checkpointFromPayload(record.eventType, payload), timeline)
      pushCheckpoint(timeline, checkpoint, record)
      if (runId) {
        updateRunStatus(timeline, runStatusIndexes, runId, {
          status: 'waiting',
          label: record.eventType === 'approval.request' ? '等待确认后继续' : '等待补充信息',
          detail: '当前任务已暂停，不会继续执行后续步骤',
        })
      }
      continue
    }
    if (record.eventType === 'run.completed') {
      if (runId) {
        finishRunStatus(
          timeline,
          runStatusIndexes,
          runId,
          record,
          'completed',
          '本轮分析已完成',
          '可以继续追问或补充分析范围',
        )
      }
      if (activeRunId === runId) {
        activeRunId = null
        checkpoint = null
      }
      continue
    }
    if (record.eventType === 'run.failed' || record.eventType === 'error') {
      if (runId) {
        finishRunStatus(
          timeline,
          runStatusIndexes,
          runId,
          record,
          'failed',
          '本轮分析未完成',
          contentText(payload.message) || '运行过程中发生异常',
        )
      }
      if (activeRunId === runId) {
        activeRunId = null
        checkpoint = null
      }
      continue
    }
    if (interruptedTerminal) {
      const interruptedActiveRun = Boolean(runId) && activeRunId === runId
      if (runId) {
        finishRunStatus(
          timeline,
          runStatusIndexes,
          runId,
          record,
          'interrupted',
          '任务已停止',
          '已保留停止前产生的内容',
        )
        interruptedRunIds.add(runId)
      }
      if (interruptedActiveRun) interruptedRunAwaitingNextStart = true
      if (activeRunId === runId) {
        activeRunId = null
        checkpoint = null
      }
    }
  }

  flushReasoning()
  return { timeline, checkpoint, activeRunId }
}

export function historyFromGateway(value: unknown): HermesHistorySnapshot {
  if (!isObject(value)) return emptyHistory()
  const rows = Array.isArray(value.messages)
    ? value.messages
    : Array.isArray(value.history)
      ? value.history
      : []
  const timeline: HermesTimelineItem[] = []
  const planFingerprints = new Set<string>()

  rows.forEach((row, index) => {
    if (!isObject(row)) return
    const role = String(row.role || '')
    const content = visibleMessageText(contentText(row.content ?? row.text))
    const occurredAt = String(row.occurredAt ?? row.created_at ?? new Date().toISOString())
    const runId = typeof row.runId === 'string' ? row.runId : null
    if ((role === 'user' || role === 'assistant') && content) {
      const messageItem: HermesTimelineItem = {
        id: createId('timeline-message'),
        kind: 'message',
        sequenceNo: index + 1,
        occurredAt,
        runId,
        message: {
          id: createId('message'),
          role,
          content,
          complete: true,
          ...(runId ? { runId } : {}),
        },
      }
      timeline.push(messageItem)
    }
    if (role === 'tool') {
      const plan = planFromToolPayload(row)
      if (plan) {
        pushGatewayPlan(timeline, planFingerprints, plan, index + 1, occurredAt, runId)
        return
      }
      timeline.push({
        id: createId('timeline-activity'),
        kind: 'activity',
        sequenceNo: index + 1,
        occurredAt,
        runId,
        activity: {
          id: createId('tool'),
          kind: 'tool',
          name: String(row.name || 'tool'),
          status: 'completed',
          context: String(row.context || ''),
          args: isObject(row.args ?? row.arguments) ? row.args ?? row.arguments : undefined,
          result: row.result ?? row.output,
        },
      })
    }
  })

  return { timeline, checkpoint: null, activeRunId: null }
}

export function checkpointFromPayload(
  eventType: string,
  payload: Record<string, unknown>,
): HermesCheckpoint {
  const approval = eventType === 'approval.request'
  const defaults = approval ? ['once', 'deny'] : []
  const checkpoint: HermesCheckpoint = {
    type: approval ? 'approval' : 'clarification',
    requestId: String(payload.request_id ?? payload.requestId ?? ''),
    title: String(payload.title ?? (approval ? '需要确认后才能继续' : '需要补充信息')),
    description: String(
      payload.description ??
        payload.question ??
        payload.reason ??
        payload.command ??
        (approval ? '请确认是否继续。' : '请提供补充信息。'),
    ),
    choices: normalizeChoices(payload.choices ?? payload.options, defaults),
  }
  return isPlanConfirmation(checkpoint)
    ? { ...checkpoint, purpose: 'plan_confirmation' }
    : checkpoint
}

export function checkpointForTimeline(
  checkpoint: HermesCheckpoint,
  timeline: HermesTimelineItem[],
): HermesCheckpoint {
  if (checkpoint.purpose !== 'plan_confirmation') return checkpoint
  const existing = timeline.find(
    (item) =>
      item.kind === 'checkpoint' &&
      item.checkpoint.purpose === 'plan_confirmation' &&
      item.checkpoint.requestId === checkpoint.requestId,
  )
  if (existing?.kind === 'checkpoint') return existing.checkpoint
  return checkpoint
}

export function contentText(value: unknown): string {
  if (typeof value === 'string') return value
  if (!Array.isArray(value)) return ''
  return value
    .map((part) => (isObject(part) ? String(part.text || '') : ''))
    .filter(Boolean)
    .join('\n')
}

export function visibleMessageText(value: string): string {
  const markers = [
    '\n\n<hermes_qa_document_context>',
    '\n\n<hermes_foundation_runtime_context>',
    '\n\n<hermes_qa_regenerate_context>',
    '\n\n<hermes_qa_workflow_selection>',
    '\n\n<hermes_qa_playbook_selection>',
  ]
  const firstMarker = markers
    .map((marker) => value.indexOf(marker))
    .filter((index) => index >= 0)
    .sort((left, right) => left - right)[0]
  return (firstMarker === undefined ? value : value.slice(0, firstMarker)).trim()
}

function normalizeChoices(value: unknown, defaults: string[]) {
  const raw = Array.isArray(value) && value.length ? value : defaults
  return raw.flatMap((choice) => {
    if (typeof choice === 'string') {
      const normalized = choice.trim()
      return normalized ? [{ value: normalized, label: choiceLabel(normalized) }] : []
    }
    if (!isObject(choice)) return []
    const choiceValue = String(choice.value ?? choice.label ?? '').trim()
    if (!choiceValue) return []
    const description = String(choice.description ?? '').trim()
    return [
      {
        value: choiceValue,
        label: String(choice.label || choiceLabel(choiceValue)),
        ...(description ? { description } : {}),
      },
    ]
  })
}

function choiceLabel(value: string): string {
  const labels: Record<string, string> = {
    once: '允许本次执行',
    always: '始终允许',
    deny: '拒绝',
    confirm: '确认并继续',
  }
  return labels[value] || value
}

function isPlanConfirmation(checkpoint: HermesCheckpoint): boolean {
  if (checkpoint.type !== 'clarification' || checkpoint.choices.length === 0) return false
  const content = `${checkpoint.title}\n${checkpoint.description}`
  const choices = checkpoint.choices.map((choice) => `${choice.value} ${choice.label}`).join('\n')
  return (
    /方案|计划|步骤|plan/i.test(content) &&
    /确认|继续|执行|confirm|continue/i.test(choices) &&
    /修改|调整|修订|modify|revise|edit/i.test(choices)
  )
}

function pushMessage(
  timeline: HermesTimelineItem[],
  role: HermesMessage['role'],
  content: string,
  record: PersistedConversationEvent,
  complete: boolean,
): void {
  const itemId = `message-${record.eventId || createId('history')}`
  timeline.push({
    id: `timeline-${itemId}`,
    kind: 'message',
    sequenceNo: record.sequenceNo,
    occurredAt: record.occurredAt,
    runId: record.runId,
    message: {
      id: itemId,
      role,
      content,
      complete,
      ...(record.runId ? { runId: record.runId } : {}),
    },
  })
}

function appendAssistant(
  timeline: HermesTimelineItem[],
  delta: string,
  record: PersistedConversationEvent,
): void {
  if (!delta) return
  const index = findStreamingAssistant(timeline, record.runId)
  if (index >= 0) {
    const item = timeline[index]
    if (item.kind === 'message') {
      timeline[index] = {
        ...item,
        message: { ...item.message, content: item.message.content + delta },
      }
    }
    return
  }
  pushMessage(timeline, 'assistant', delta, record, false)
}

function completeAssistant(
  timeline: HermesTimelineItem[],
  content: string,
  record: PersistedConversationEvent,
): void {
  const index = findStreamingAssistant(timeline, record.runId)
  if (index >= 0) {
    const item = timeline[index]
    if (item.kind === 'message') {
      timeline[index] = {
        ...item,
        message: { ...item.message, content: content || item.message.content, complete: true },
      }
    }
    return
  }
  if (content) pushMessage(timeline, 'assistant', content, record, true)
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

function pushActivity(
  timeline: HermesTimelineItem[],
  activity: HermesActivity,
  record: PersistedConversationEvent,
): void {
  timeline.push({
    id: `timeline-${activity.id}`,
    kind: 'activity',
    sequenceNo: record.sequenceNo,
    occurredAt: record.occurredAt,
    runId: record.runId,
    activity,
  })
}

function pushPlan(
  timeline: HermesTimelineItem[],
  fingerprints: Set<string>,
  plan: HermesPlan,
  record: PersistedConversationEvent,
  runId: string | null,
): void {
  pushGatewayPlan(
    timeline,
    fingerprints,
    plan,
    record.sequenceNo,
    record.occurredAt,
    runId,
    record.eventId,
  )
}

function pushGatewayPlan(
  timeline: HermesTimelineItem[],
  fingerprints: Set<string>,
  plan: HermesPlan,
  sequenceNo: number,
  occurredAt: string,
  runId: string | null,
  eventId?: string,
): void {
  const fingerprint = `${runId || 'session'}:${planFingerprint(plan)}`
  if (fingerprints.has(fingerprint)) return
  fingerprints.add(fingerprint)
  const item: HermesPlanTimelineItem = {
    id: `timeline-plan-${eventId || createId('history')}`,
    kind: 'plan',
    sequenceNo,
    occurredAt,
    runId,
    plan,
  }
  timeline.push(item)
}

function upsertTool(
  timeline: HermesTimelineItem[],
  indexes: Map<string, number>,
  record: PersistedConversationEvent,
): void {
  const payload = record.payload
  const id = String(payload.tool_id ?? payload.id ?? createId('tool'))
  const key = `${record.runId || 'session'}:${id}`
  const failed = record.eventType === 'tool.failed'
  const completed = record.eventType === 'tool.complete' || record.eventType === 'tool.completed'
  const rawArgs = payload.args ?? payload.arguments
  const activity: HermesToolActivity = {
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
  const index = indexes.get(key)
  if (index === undefined) {
    indexes.set(key, timeline.length)
    pushActivity(timeline, activity, record)
  } else {
    const item = timeline[index]
    if (item.kind === 'activity' && item.activity.kind === 'tool') {
      timeline[index] = { ...item, activity: { ...item.activity, ...activity } }
    }
  }
}

function pushCheckpoint(
  timeline: HermesTimelineItem[],
  checkpoint: HermesCheckpoint,
  record: PersistedConversationEvent,
): void {
  timeline.push({
    id: `timeline-checkpoint-${record.eventId || createId('history')}`,
    kind: 'checkpoint',
    sequenceNo: record.sequenceNo,
    occurredAt: record.occurredAt,
    runId: record.runId,
    checkpoint,
    resolved: false,
  })
}

function resolveLatestCheckpoint(timeline: HermesTimelineItem[], responseText: string): void {
  for (let index = timeline.length - 1; index >= 0; index -= 1) {
    const item = timeline[index]
    if (item.kind === 'checkpoint' && !item.resolved) {
      const resolved: HermesCheckpointTimelineItem = {
        ...item,
        resolved: true,
        ...(responseText ? { responseText } : {}),
      }
      timeline[index] = resolved
      return
    }
  }
}

function pushRunStatus(
  timeline: HermesTimelineItem[],
  indexes: Map<string, number>,
  runId: string,
  record: PersistedConversationEvent,
  status: Pick<HermesRunStatusTimelineItem, 'status' | 'label' | 'detail' | 'startedAt'>,
): void {
  const item: HermesRunStatusTimelineItem = {
    id: `run-status-${runId}`,
    kind: 'run_status',
    sequenceNo: record.sequenceNo,
    occurredAt: record.occurredAt,
    runId,
    ...status,
  }
  indexes.set(runId, timeline.length)
  timeline.push(item)
}

function updateRunStatus(
  timeline: HermesTimelineItem[],
  indexes: Map<string, number>,
  runId: string,
  patch: Partial<HermesRunStatusTimelineItem>,
): void {
  const index = indexes.get(runId)
  if (index === undefined) return
  const item = timeline[index]
  if (item.kind !== 'run_status' || item.status === 'interrupted') return
  timeline[index] = { ...item, ...patch }
}

function finishRunStatus(
  timeline: HermesTimelineItem[],
  indexes: Map<string, number>,
  runId: string,
  record: PersistedConversationEvent,
  status: HermesRunStatus,
  label: string,
  detail: string,
): void {
  const index = indexes.get(runId)
  if (index === undefined) return
  const item = timeline[index]
  if (item.kind !== 'run_status') return
  if (item.status === 'interrupted' && status !== 'interrupted') return
  const durationMs = Math.max(
    0,
    new Date(record.occurredAt).getTime() - new Date(item.startedAt || item.occurredAt).getTime(),
  )
  timeline[index] = {
    ...item,
    status,
    label,
    detail,
    finishedAt: record.occurredAt,
    durationMs: Number.isFinite(durationMs) ? durationMs : undefined,
  }
}

function isToolEvent(type: string): boolean {
  return ['tool.start', 'tool.started', 'tool.complete', 'tool.completed', 'tool.failed'].includes(
    type,
  )
}

function isCompletedToolEvent(type: string): boolean {
  return type === 'tool.complete' || type === 'tool.completed'
}

function emptyHistory(): HermesHistorySnapshot {
  return { timeline: [], checkpoint: null, activeRunId: null }
}

function createId(prefix: string): string {
  return `${prefix}-${typeof crypto.randomUUID === 'function' ? crypto.randomUUID() : Date.now()}`
}

function isObject(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
