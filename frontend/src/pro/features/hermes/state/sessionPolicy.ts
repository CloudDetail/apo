import type { HermesSessionStatus } from '../client/types.ts'

export function shouldRestoreInBackground(status: HermesSessionStatus): boolean {
  return status === 'running' || status === 'waiting_confirmation'
}
