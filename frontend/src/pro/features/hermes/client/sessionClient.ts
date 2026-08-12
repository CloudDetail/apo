import { historyFromGateway } from './history.ts'
import { JsonRpcClient, JsonRpcError, type JsonRpcNotification } from './jsonRpcClient.ts'
import type { HermesEvent, HermesHistorySnapshot, HermesSession } from './types.ts'

const CREATE_TIMEOUT = 5_000
const ACTIVATE_TIMEOUT = 5_000
const RESUME_TIMEOUT = 8_000
const SESSION_ACTION_TIMEOUT = 5_000
const SUBMIT_TIMEOUT = 10_000

export class HermesSessionClient {
  private readonly rpc: JsonRpcClient

  constructor(rpc: JsonRpcClient) {
    this.rpc = rpc
  }

  async create(): Promise<HermesSession> {
    const result = await this.rpc.request<unknown>(
      'session.create',
      {
        source: 'apo-front',
      },
      { timeoutMs: CREATE_TIMEOUT },
    )
    return normalizeSession(result)
  }

  async resume(storedSessionId: string): Promise<HermesSession> {
    const result = await this.rpc.request<unknown>(
      'session.resume',
      {
        session_id: storedSessionId,
      },
      { timeoutMs: RESUME_TIMEOUT },
    )
    const resumed = normalizeSession(result, storedSessionId)
    await this.rpc.request(
      'session.activate',
      { session_id: resumed.liveSessionId },
      { timeoutMs: ACTIVATE_TIMEOUT },
    )
    return resumed
  }

  async restore(storedSessionId: string, liveSessionId: string | null): Promise<HermesSession> {
    if (liveSessionId) {
      try {
        const result = await this.rpc.request<unknown>(
          'session.activate',
          {
            session_id: liveSessionId,
          },
          { timeoutMs: ACTIVATE_TIMEOUT },
        )
        return normalizeSession(result, storedSessionId)
      } catch (reason) {
        if (!isMissingSession(reason)) throw reason
      }
    }
    return this.resume(storedSessionId)
  }

  async historySnapshot(liveSessionId: string): Promise<HermesHistorySnapshot> {
    const result = await this.rpc.request<unknown>(
      'session.history',
      {
        session_id: liveSessionId,
      },
      { timeoutMs: SESSION_ACTION_TIMEOUT },
    )
    return historyFromGateway(result)
  }

  createRunId(): string {
    return typeof crypto.randomUUID === 'function' ? crypto.randomUUID() : `run-${Date.now()}`
  }

  submit(liveSessionId: string, text: string, clientRunId = this.createRunId()) {
    return this.rpc.request(
      'prompt.submit',
      {
        session_id: liveSessionId,
        text,
        client_run_id: clientRunId,
      },
      { timeoutMs: SUBMIT_TIMEOUT },
    )
  }

  steer(liveSessionId: string, text: string) {
    return this.rpc.request(
      'session.steer',
      {
        session_id: liveSessionId,
        text,
      },
      { timeoutMs: SESSION_ACTION_TIMEOUT },
    )
  }

  interrupt(liveSessionId: string) {
    return this.rpc.request(
      'session.interrupt',
      { session_id: liveSessionId },
      { timeoutMs: SESSION_ACTION_TIMEOUT },
    )
  }

  respondApproval(liveSessionId: string, requestId: string, choice: string) {
    return this.rpc.request(
      'approval.respond',
      {
        session_id: liveSessionId,
        request_id: requestId,
        choice,
      },
      { timeoutMs: SESSION_ACTION_TIMEOUT },
    )
  }

  respondClarification(requestId: string, answer: string) {
    return this.rpc.request(
      'clarify.respond',
      {
        request_id: requestId,
        answer,
      },
      { timeoutMs: SESSION_ACTION_TIMEOUT },
    )
  }

  subscribe(listener: (event: HermesEvent) => void): () => void {
    return this.rpc.subscribe((notification) => {
      const event = normalizeEvent(notification)
      if (event) listener(event)
    })
  }
}

export function normalizeEvent(notification: JsonRpcNotification): HermesEvent | null {
  if (notification.method !== 'event' || !isObject(notification.params)) return null
  return {
    type: String(notification.params.type || 'unknown'),
    sessionId: String(notification.params.session_id || '').trim() || null,
    payload: isObject(notification.params.payload) ? notification.params.payload : {},
  }
}

export function normalizeSession(value: unknown, fallbackStoredId?: string): HermesSession {
  if (!isObject(value)) throw new Error('Hermes 没有成功创建会话，请重试。')
  const liveSessionId = String(value.session_id || '').trim()
  const storedSessionId = String(
    value.stored_session_id ||
      value.session_key ||
      value.resumed ||
      fallbackStoredId ||
      liveSessionId,
  ).trim()
  if (!liveSessionId || !storedSessionId) {
    throw new Error('Hermes 返回的会话信息不完整，请重试。')
  }
  const info = isObject(value.info) ? value.info : null
  return {
    liveSessionId,
    storedSessionId,
    ...(info && String(info.title || '').trim() ? { title: String(info.title).trim() } : {}),
    ...(typeof value.running === 'boolean' ? { running: value.running } : {}),
    ...(String(value.status || '').trim() ? { status: String(value.status).trim() } : {}),
  }
}

function isMissingSession(reason: unknown): boolean {
  if (!(reason instanceof JsonRpcError)) return false
  const code = Number(reason.payload.code)
  return code === 4001 || code === 4007
}

function isObject(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
