import { useMemo } from 'react'
import { useUserContext } from 'src/core/contexts/UserContext'
import { canonicalHermesUserId } from '../client/userScope.ts'
import { WorkCapabilityApi } from './client/workCapabilityApi.ts'

export function useWorkCapabilityUserScopeKey(): string | null {
  const { user } = useUserContext() as {
    user?: { userId?: string | number }
  }
  return canonicalHermesUserId(user?.userId)
}

export function useWorkCapabilityApi(): WorkCapabilityApi {
  const userId = useWorkCapabilityUserScopeKey()
  return useMemo(
    () => new WorkCapabilityApi({ readUserId: () => userId }),
    [userId],
  )
}
