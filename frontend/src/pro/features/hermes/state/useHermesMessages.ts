import { useCallback, useReducer } from 'react'
import type { HermesEvent, HermesHistorySnapshot } from '../client/types.ts'
import {
  hermesMessageReducer,
  initialHermesMessageState,
} from './messageReducer.ts'

export function useHermesMessages() {
  const [state, dispatch] = useReducer(hermesMessageReducer, initialHermesMessageState)

  const reset = useCallback(() => dispatch({ type: 'RESET' }), [])
  const hydrate = useCallback(
    (snapshot: HermesHistorySnapshot, busy = false, statusText?: string) =>
      dispatch({ type: 'HYDRATE', snapshot, busy, statusText }),
    [],
  )
  const submitStarted = useCallback(
    (text: string, runId: string) => dispatch({ type: 'SUBMIT_STARTED', text, runId }),
    [],
  )
  const addUserMessage = useCallback(
    (text: string, runId: string | null) =>
      dispatch({ type: 'USER_ADDITION', text, runId }),
    [],
  )
  const checkpointResponded = useCallback(
    (text: string) => dispatch({ type: 'CHECKPOINT_RESPONDED', text }),
    [],
  )
  const applyEvent = useCallback(
    (event: HermesEvent) => dispatch({ type: 'EVENT_RECEIVED', event }),
    [],
  )
  const sendFailed = useCallback(
    (error: string) => dispatch({ type: 'SEND_FAILED', error }),
    [],
  )
  const stopRequested = useCallback(() => dispatch({ type: 'STOP_REQUESTED' }), [])
  const stopAcknowledged = useCallback(() => dispatch({ type: 'STOP_ACKNOWLEDGED' }), [])
  const stopRetryAvailable = useCallback(
    () => dispatch({ type: 'STOP_RETRY_AVAILABLE' }),
    [],
  )
  const stopFailed = useCallback(
    (error: string, runId: string | null) => dispatch({ type: 'STOP_FAILED', error, runId }),
    [],
  )
  const setStatus = useCallback(
    (statusText: string, busy?: boolean) =>
      dispatch({ type: 'SET_STATUS', statusText, busy }),
    [],
  )
  const clearError = useCallback(() => dispatch({ type: 'CLEAR_ERROR' }), [])

  return {
    state,
    reset,
    hydrate,
    submitStarted,
    addUserMessage,
    checkpointResponded,
    applyEvent,
    sendFailed,
    stopRequested,
    stopAcknowledged,
    stopRetryAvailable,
    stopFailed,
    setStatus,
    clearError,
  }
}
