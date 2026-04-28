import { store } from 'src/core/store/store'

export const normalizeAgentUserId = (value: unknown) => {
  if (typeof value !== 'string' && typeof value !== 'number') {
    return undefined
  }

  const normalized = String(value).trim()
  if (!normalized) {
    return undefined
  }

  return normalized
}

export const getCurrentAgentUserId = () => {
  const state = store.getState() as {
    userReducer?: {
      user?: {
        userId?: string | number
      }
    }
  }

  return normalizeAgentUserId(state.userReducer?.user?.userId)
}
