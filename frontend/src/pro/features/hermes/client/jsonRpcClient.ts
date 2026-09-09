import type { HermesConnectionStatus } from './types.ts'

export interface JsonRpcErrorPayload {
  code: number | string
  message: string
  data?: unknown
}

export interface JsonRpcNotification {
  jsonrpc: '2.0'
  method: string
  params?: unknown
}

type PendingRequest = {
  resolve: (value: unknown) => void
  reject: (reason: Error) => void
  timer: ReturnType<typeof setTimeout> | null
}

export interface JsonRpcRequestOptions {
  timeoutMs?: number
}

type WebSocketLike = Pick<WebSocket, 'readyState' | 'send' | 'close' | 'addEventListener'>

type SocketFactory = (url: string) => WebSocketLike

const SOCKET_OPEN = 1
const MAX_RECONNECT_DELAY = 10_000
const DEFAULT_REQUEST_TIMEOUT = 10_000
const CONNECT_TIMEOUT = 8_000

export class JsonRpcError extends Error {
  readonly payload: JsonRpcErrorPayload

  constructor(payload: JsonRpcErrorPayload) {
    super(payload.message || 'Agent 请求执行失败。')
    this.payload = payload
    this.name = 'JsonRpcError'
  }
}

export class JsonRpcTimeoutError extends Error {
  readonly method: string
  readonly timeoutMs: number

  constructor(method: string, timeoutMs: number) {
    super(`Agent 请求超时（${method}，${Math.round(timeoutMs / 1000)} 秒）。`)
    this.method = method
    this.timeoutMs = timeoutMs
    this.name = 'JsonRpcTimeoutError'
  }
}

export class JsonRpcClient {
  private readonly url: string
  private readonly socketFactory: SocketFactory
  private socket: WebSocketLike | null = null
  private connecting: Promise<void> | null = null
  private sequence = 0
  private reconnectAttempts = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private manuallyClosed = false
  private readonly pending = new Map<number, PendingRequest>()
  private readonly listeners = new Set<(notification: JsonRpcNotification) => void>()
  private readonly statusListeners = new Set<
    (status: HermesConnectionStatus, error?: Error) => void
  >()

  constructor(url: string, socketFactory: SocketFactory = (value) => new WebSocket(value)) {
    this.url = url
    this.socketFactory = socketFactory
  }

  connect(): Promise<void> {
    if (this.socket?.readyState === SOCKET_OPEN) return Promise.resolve()
    if (this.connecting) return this.connecting

    this.manuallyClosed = false
    const reconnecting = this.reconnectAttempts > 0
    this.publishStatus(reconnecting ? 'reconnecting' : 'connecting')
    this.connecting = new Promise<void>((resolve, reject) => {
      let settled = false
      const socket = this.socketFactory(this.url)
      this.socket = socket
      let connectTimer: ReturnType<typeof setTimeout>

      const failConnection = (message: string) => {
        if (settled) return
        settled = true
        clearTimeout(connectTimer)
        this.connecting = null
        const error = new Error(message)
        this.publishStatus('error', error)
        reject(error)
      }

      connectTimer = setTimeout(() => {
        failConnection('连接 Agent 服务超时，请稍后重试。')
        socket.close()
      }, CONNECT_TIMEOUT)

      socket.addEventListener('open', () => {
        if (settled) return
        settled = true
        clearTimeout(connectTimer)
        this.connecting = null
        this.reconnectAttempts = 0
        this.publishStatus('connected')
        resolve()
      })
      socket.addEventListener('error', () => {
        failConnection('暂时无法连接 Agent 服务，请稍后重试。')
      })
      socket.addEventListener('message', (event) => {
        this.receive(String((event as MessageEvent).data))
      })
      socket.addEventListener('close', () => {
        const wasCurrentSocket = this.socket === socket
        if (wasCurrentSocket) this.socket = null
        this.connecting = null
        if (!settled) {
          settled = true
          clearTimeout(connectTimer)
          reject(new Error('Agent 服务连接已断开。'))
        }
        const error = new Error('Agent 服务连接已断开。')
        for (const request of this.pending.values()) {
          if (request.timer) clearTimeout(request.timer)
          request.reject(error)
        }
        this.pending.clear()
        if (this.manuallyClosed) {
          this.publishStatus('disconnected')
          return
        }
        this.publishStatus('disconnected', error)
        this.scheduleReconnect()
      })
    })

    return this.connecting
  }

  async request<T>(
    method: string,
    params: Record<string, unknown> = {},
    options: JsonRpcRequestOptions = {},
  ): Promise<T> {
    await this.connect()
    const socket = this.socket
    if (!socket || socket.readyState !== SOCKET_OPEN) {
      throw new Error('Agent 服务暂时不可用，请稍后重试。')
    }

    const id = ++this.sequence
    const timeoutMs = options.timeoutMs ?? DEFAULT_REQUEST_TIMEOUT
    const response = new Promise<T>((resolve, reject) => {
      const pending: PendingRequest = {
        resolve: (value) => resolve(value as T),
        reject,
        timer: null,
      }
      if (timeoutMs > 0) {
        pending.timer = setTimeout(() => {
          if (this.pending.get(id) !== pending) return
          this.pending.delete(id)
          reject(new JsonRpcTimeoutError(method, timeoutMs))
        }, timeoutMs)
      }
      this.pending.set(id, pending)
    })
    try {
      socket.send(JSON.stringify({ jsonrpc: '2.0', id, method, params }))
    } catch (reason) {
      const pending = this.pending.get(id)
      if (pending?.timer) clearTimeout(pending.timer)
      this.pending.delete(id)
      throw reason
    }
    return response
  }

  subscribe(listener: (notification: JsonRpcNotification) => void): () => void {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  subscribeStatus(listener: (status: HermesConnectionStatus, error?: Error) => void): () => void {
    this.statusListeners.add(listener)
    return () => this.statusListeners.delete(listener)
  }

  close(): void {
    this.manuallyClosed = true
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.reconnectTimer = null
    const error = new Error('Agent 服务连接已关闭。')
    for (const request of this.pending.values()) {
      if (request.timer) clearTimeout(request.timer)
      request.reject(error)
    }
    this.pending.clear()
    this.socket?.close()
    this.socket = null
  }

  private scheduleReconnect(): void {
    if (this.manuallyClosed || this.reconnectTimer) return
    this.reconnectAttempts += 1
    const delay = Math.min(500 * 2 ** (this.reconnectAttempts - 1), MAX_RECONNECT_DELAY)
    this.publishStatus('reconnecting')
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      void this.connect().catch(() => this.scheduleReconnect())
    }, delay)
  }

  private publishStatus(status: HermesConnectionStatus, error?: Error): void {
    for (const listener of this.statusListeners) listener(status, error)
  }

  private receive(raw: string): void {
    let payload: unknown
    try {
      payload = JSON.parse(raw)
    } catch {
      return
    }
    if (!isObject(payload) || payload.jsonrpc !== '2.0') return

    if (typeof payload.id === 'number') {
      const request = this.pending.get(payload.id)
      if (!request) return
      this.pending.delete(payload.id)
      if (request.timer) clearTimeout(request.timer)
      if (isObject(payload.error)) {
        request.reject(
          new JsonRpcError({
            code:
              typeof payload.error.code === 'number' || typeof payload.error.code === 'string'
                ? payload.error.code
                : 'unknown',
            message: String(payload.error.message || 'Agent 请求执行失败。'),
            data: payload.error.data,
          }),
        )
      } else {
        request.resolve(payload.result)
      }
      return
    }

    if (typeof payload.method === 'string') {
      const notification = payload as unknown as JsonRpcNotification
      for (const listener of this.listeners) listener(notification)
    }
  }
}

function isObject(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
