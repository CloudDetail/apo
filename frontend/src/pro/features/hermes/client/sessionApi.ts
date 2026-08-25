import type {
  HermesCheckpoint,
  HermesSessionBookmark,
  HermesSessionStatus,
  PersistedConversationEvent,
} from './types.ts'
import { normalizeHermesEventPayload } from './eventPayload.ts'
import {
  HERMES_USER_ID_HEADER,
  requireHermesUserId,
  type HermesUserIdReader,
} from './userScope.ts'

type Fetcher = typeof fetch

export class HermesSessionApi {
  private readonly baseUrl: string
  private readonly request: Fetcher
  private readonly readUserId: HermesUserIdReader

  constructor(
    baseUrl: string,
    request: Fetcher = (input, init) => fetch(input, init),
    readUserId: HermesUserIdReader = () => null,
  ) {
    this.baseUrl = baseUrl
    this.request = request
    this.readUserId = readUserId
  }

  async list(): Promise<HermesSessionBookmark[]> {
    const payload = await this.send<{ sessions?: unknown[] }>('/sessions', { method: 'GET' })
    return Array.isArray(payload.sessions)
      ? payload.sessions.flatMap((item) => {
          const session = normalizeBookmark(item)
          return session ? [session] : []
        })
      : []
  }

  async get(storedSessionId: string): Promise<HermesSessionBookmark> {
    const payload = await this.send<unknown>(`/sessions/${encodeURIComponent(storedSessionId)}`, {
      method: 'GET',
    })
    const session = normalizeBookmark(payload)
    if (!session) throw new Error('Hermes 返回的会话信息不完整。')
    return session
  }

  async upsert(session: HermesSessionBookmark): Promise<HermesSessionBookmark> {
    const payload = await this.send<unknown>(
      `/sessions/${encodeURIComponent(session.storedSessionId)}`,
      {
        method: 'PUT',
        body: JSON.stringify(toServerBookmark(session)),
      },
    )
    const saved = normalizeBookmark(payload)
    if (!saved) throw new Error('Hermes 没有成功保存会话。')
    return saved
  }

  async remove(storedSessionId: string): Promise<void> {
    const response = await this.sendRequest(
      `/sessions/${encodeURIComponent(storedSessionId)}`,
      {
        method: 'DELETE',
        headers: requestHeaders(false),
      },
    )
    if (!response.ok) throw new Error('无法删除该会话，请稍后重试。')
  }

  async history(storedSessionId: string): Promise<PersistedConversationEvent[]> {
    const payload = await this.send<{ events?: unknown[] }>(
      `/sessions/${encodeURIComponent(storedSessionId)}/history`,
      { method: 'GET' },
    )
    return Array.isArray(payload.events)
      ? payload.events.flatMap((item) => {
          const event = normalizeConversationEvent(item)
          return event ? [event] : []
        })
      : []
  }

  async interrupt(
    storedSessionId: string,
    activeRunId: string | null,
  ): Promise<HermesSessionBookmark> {
    const payload = await this.send<unknown>(
      `/sessions/${encodeURIComponent(storedSessionId)}/interrupt`,
      {
        method: 'POST',
        body: JSON.stringify({ active_run_id: activeRunId, cause: 'user_requested' }),
      },
    )
    const session = normalizeBookmark(payload)
    if (!session) throw new Error('Hermes 没有成功保存停止状态。')
    return session
  }

  private async send<T>(path: string, init: RequestInit): Promise<T> {
    const response = await this.sendRequest(path, init)
    if (!response.ok) {
      const message = await response.text().catch(() => '')
      throw new Error(message || `Hermes 请求失败（${response.status}）。`)
    }
    return (await response.json()) as T
  }

  private async sendRequest(path: string, init: RequestInit): Promise<Response> {
    const userId = requireHermesUserId(this.readUserId())
    const headers = requestHeaders(Boolean(init.body), init.headers)
    headers.set(HERMES_USER_ID_HEADER, userId)
    return this.request(this.url(path), {
      ...init,
      headers,
      credentials: 'include',
    })
  }

  private url(path: string): string {
    return `${this.baseUrl}${path}`
  }
}

export function createSerializedSessionWriter(
  api: Pick<HermesSessionApi, 'upsert'>,
): (session: HermesSessionBookmark) => Promise<HermesSessionBookmark> {
  const pendingBySession = new Map<string, Promise<void>>()

  return async (session) => {
    const previous = pendingBySession.get(session.storedSessionId) || Promise.resolve()
    const request = previous.catch(() => undefined).then(() => api.upsert(session))
    const settled = request.then(
      () => undefined,
      () => undefined,
    )
    pendingBySession.set(session.storedSessionId, settled)

    try {
      return await request
    } finally {
      if (pendingBySession.get(session.storedSessionId) === settled) {
        pendingBySession.delete(session.storedSessionId)
      }
    }
  }
}

function requestHeaders(hasBody: boolean, initial?: HeadersInit): Headers {
  const headers = new Headers(initial)
  headers.set('Accept', 'application/json')
  if (hasBody) headers.set('Content-Type', 'application/json')
  return headers
}

function normalizeBookmark(value: unknown): HermesSessionBookmark | null {
  if (!isObject(value)) return null
  const storedSessionId = String(value.stored_session_id || '').trim()
  const title = String(value.title || '').trim()
  const createdAt = String(value.created_at || '').trim()
  const updatedAt = String(value.updated_at || '').trim()
  if (!storedSessionId || !title || !createdAt || !updatedAt) return null
  return {
    storedSessionId,
    liveSessionId: String(value.live_session_id || '').trim() || null,
    title,
    createdAt,
    updatedAt,
    status: normalizeSessionStatus(value.status),
    activeRunId: String(value.active_run_id || '').trim() || null,
    checkpoint: normalizeCheckpoint(value.checkpoint),
  }
}

function normalizeSessionStatus(value: unknown): HermesSessionStatus {
  return ['running', 'waiting_confirmation', 'failed', 'interrupted'].includes(String(value))
    ? (String(value) as HermesSessionStatus)
    : 'ready'
}

function normalizeCheckpoint(value: unknown): HermesCheckpoint | null {
  if (!isObject(value)) return null
  const type = value.type === 'approval' ? 'approval' : 'clarification'
  const purpose = value.purpose === 'plan_confirmation' ? 'plan_confirmation' : undefined
  return {
    type,
    requestId: String(value.request_id ?? value.requestId ?? ''),
    title: String(value.title || (type === 'approval' ? '需要确认' : '需要补充信息')),
    description: String(value.description || ''),
    choices: Array.isArray(value.choices)
      ? value.choices.flatMap((choice) => {
          if (typeof choice === 'string') {
            const choiceValue = choice.trim()
            return choiceValue ? [{ value: choiceValue, label: choiceValue }] : []
          }
          if (!isObject(choice)) return []
          const choiceValue = String(choice.value ?? choice.label ?? '').trim()
          return choiceValue
            ? [
                {
                  value: choiceValue,
                  label: String(choice.label || choiceValue),
                  ...(String(choice.description || '').trim()
                    ? { description: String(choice.description).trim() }
                    : {}),
                },
              ]
            : []
        })
      : [],
    ...(purpose ? { purpose } : {}),
  }
}

function normalizeConversationEvent(value: unknown): PersistedConversationEvent | null {
  if (!isObject(value)) return null
  const eventId = String(value.event_id || '').trim()
  const eventType = String(value.event_type || '').trim()
  const occurredAt = String(value.occurred_at || '').trim()
  const direction =
    value.direction === 'browser_to_gateway' ? 'browser_to_gateway' : 'gateway_to_browser'
  if (!eventId || !eventType || !occurredAt) return null
  return {
    eventId,
    sequenceNo: Number(value.sequence_no || 0),
    direction,
    eventType,
    runId: String(value.run_id || '').trim() || null,
    payload: normalizeHermesEventPayload(value.payload),
    occurredAt,
  }
}

function toServerBookmark(session: HermesSessionBookmark) {
  return {
    stored_session_id: session.storedSessionId,
    live_session_id: session.liveSessionId,
    title: session.title,
    created_at: session.createdAt,
    updated_at: session.updatedAt,
    status: session.status,
    active_run_id: session.activeRunId,
    processing_duration_ms: 0,
    checkpoint: session.checkpoint,
  }
}

function isObject(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
