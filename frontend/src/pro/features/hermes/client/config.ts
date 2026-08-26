import {
  HERMES_USER_ID_QUERY,
  requireHermesUserId,
} from './userScope.ts'

const DEFAULT_API_BASE_URL = '/hermes-api'
const DEFAULT_WS_URL = '/hermes-ws'

export type AgentEngine = 'hermes' | 'agentsre'

export function hermesApiBaseUrl(): string {
  const configured = String((import.meta as any).env?.VITE_HERMES_API_BASE_URL || '').trim()
  return trimTrailingSlash(configured || DEFAULT_API_BASE_URL)
}

export function hermesWebSocketUrl(
  userId?: unknown,
  locationValue: Location = window.location,
): string {
  const configured = String((import.meta as any).env?.VITE_HERMES_WS_URL || '').trim()
  const resolved = resolveWebSocketUrl(configured || DEFAULT_WS_URL, locationValue)
  const normalizedUserId = requireHermesUserId(userId)
  const url = new URL(resolved)
  url.searchParams.set(HERMES_USER_ID_QUERY, normalizedUserId)
  return url.toString()
}

export function configuredAgentEngine(): AgentEngine {
  const configured = String((import.meta as any).env?.VITE_AGENT_ENGINE || 'hermes').toLowerCase()
  return configured === 'agentsre' ? 'agentsre' : 'hermes'
}

export function resolveWebSocketUrl(value: string, locationValue: Location): string {
  if (/^wss?:\/\//i.test(value)) return value
  if (/^https?:\/\//i.test(value)) return value.replace(/^http/i, 'ws')

  const protocol = locationValue.protocol === 'https:' ? 'wss:' : 'ws:'
  const path = value.startsWith('/') ? value : `/${value}`
  return `${protocol}//${locationValue.host}${path}`
}

function trimTrailingSlash(value: string): string {
  return value.length > 1 ? value.replace(/\/+$/, '') : value
}
