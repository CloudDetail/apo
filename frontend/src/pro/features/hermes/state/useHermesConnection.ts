import { useEffect, useMemo, useState } from 'react'
import { useSelector } from 'react-redux'
import { hermesApiBaseUrl, hermesWebSocketUrl } from '../client/config.ts'
import { JsonRpcClient } from '../client/jsonRpcClient.ts'
import { HermesSessionApi } from '../client/sessionApi.ts'
import { HermesSessionClient } from '../client/sessionClient.ts'
import type { HermesConnectionStatus } from '../client/types.ts'
import { canonicalHermesUserId, HermesUserScopeError } from '../client/userScope.ts'

export function createHermesConnection(userId: string | null) {
  const sessionApi = new HermesSessionApi(
    hermesApiBaseUrl(),
    undefined,
    () => userId,
  )
  if (!userId) {
    return { rpc: null, sessionClient: null, sessionApi }
  }
  const rpc = new JsonRpcClient(hermesWebSocketUrl(userId))
  return {
    rpc,
    sessionClient: new HermesSessionClient(rpc),
    sessionApi,
  }
}

export function useHermesConnection() {
  const rawUserId = useSelector((state: any) => state.userReducer?.user?.userId)
  const userId = canonicalHermesUserId(rawUserId)
  const [status, setStatus] = useState<HermesConnectionStatus>('idle')
  const [error, setError] = useState<string | null>(null)
  const connection = useMemo(() => createHermesConnection(userId), [userId])

  useEffect(() => {
    const { rpc } = connection
    let active = true
    if (!userId || !rpc) {
      setStatus('idle')
      setError(new HermesUserScopeError().message)
      return () => {
        active = false
        rpc?.close()
      }
    }
    setError(null)
    const unsubscribe = rpc.subscribeStatus((nextStatus, reason) => {
      if (!active) return
      setStatus(nextStatus)
      setError(reason?.message || null)
    })
    void rpc.connect().catch((reason) => {
      if (!active) return
      setError(reason instanceof Error ? reason.message : '无法连接 Agent 服务。')
    })
    return () => {
      active = false
      unsubscribe()
      rpc.close()
    }
  }, [connection, userId])

  return {
    ...connection,
    userScopeKey: userId,
    userScopeReady: Boolean(userId),
    connectionStatus: status,
    connectionError: error,
  }
}
