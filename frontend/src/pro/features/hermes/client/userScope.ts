export const HERMES_USER_ID_HEADER = 'X-Hermes-User-Id'
export const HERMES_USER_ID_QUERY = 'user_id'
const HERMES_USER_ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/
const RESERVED_LOCAL_OWNER_ID = 'local-host'

export type HermesUserIdReader = () => unknown

export class HermesUserScopeError extends Error {
  constructor(message = 'APO 用户信息尚未加载，暂时无法访问 Hermes。') {
    super(message)
    this.name = 'HermesUserScopeError'
  }
}

export function canonicalHermesUserId(value: unknown): string | null {
  if (typeof value !== 'string' && typeof value !== 'number') return null
  if (typeof value === 'number' && !Number.isFinite(value)) return null
  const normalized = String(value).trim()
  if (normalized === RESERVED_LOCAL_OWNER_ID) return null
  return HERMES_USER_ID_PATTERN.test(normalized) ? normalized : null
}

export function requireHermesUserId(value: unknown): string {
  const userId = canonicalHermesUserId(value)
  if (!userId) throw new HermesUserScopeError()
  return userId
}

export interface HermesUserScopeSnapshot<T> {
  generation: number
  identity: T
  userId: string
}

export class HermesUserScopeGeneration {
  private generation = 0
  private userId: string | null = null

  update(userId: string | null): void {
    if (this.userId === userId) return
    this.userId = userId
    this.generation += 1
  }

  capture<T>(identity: T): HermesUserScopeSnapshot<T> {
    if (!this.userId) throw new HermesUserScopeError()
    return { generation: this.generation, identity, userId: this.userId }
  }

  isCurrent<T>(snapshot: HermesUserScopeSnapshot<T>, identity: T): boolean {
    return (
      snapshot.generation === this.generation &&
      snapshot.userId === this.userId &&
      snapshot.identity === identity
    )
  }
}
