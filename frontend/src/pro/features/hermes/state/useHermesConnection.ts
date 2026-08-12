import { useEffect, useRef, useState } from 'react'
import { hermesApiBaseUrl, hermesWebSocketUrl } from '../client/config.ts'
import { JsonRpcClient } from '../client/jsonRpcClient.ts'
import { HermesSessionApi } from '../client/sessionApi.ts'
import { HermesSessionClient } from '../client/sessionClient.ts'
import type { HermesConnectionStatus } from '../client/types.ts'

export function useHermesConnection() {
  const rpcRef = useRef<JsonRpcClient | null>(null)
  const sessionClientRef = useRef<HermesSessionClient | null>(null)
  const sessionApiRef = useRef<HermesSessionApi | null>(null)
  const [status, setStatus] = useState<HermesConnectionStatus>('idle')
  const [error, setError] = useState<string | null>(null)

  if (!rpcRef.current) {
    rpcRef.current = new JsonRpcClient(hermesWebSocketUrl())
    sessionClientRef.current = new HermesSessionClient(rpcRef.current)
    sessionApiRef.current = new HermesSessionApi(hermesApiBaseUrl())
  }

  useEffect(() => {
    const rpc = rpcRef.current!
    const unsubscribe = rpc.subscribeStatus((nextStatus, reason) => {
      setStatus(nextStatus)
      setError(reason?.message || null)
    })
    void rpc.connect().catch((reason) => {
      setError(reason instanceof Error ? reason.message : '无法连接 Hermes 服务。')
    })
    return () => {
      unsubscribe()
      rpc.close()
    }
  }, [])

  return {
    rpc: rpcRef.current,
    sessionClient: sessionClientRef.current!,
    sessionApi: sessionApiRef.current!,
    connectionStatus: status,
    connectionError: error,
  }
}
