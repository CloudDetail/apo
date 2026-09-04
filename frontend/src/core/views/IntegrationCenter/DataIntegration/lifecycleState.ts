export type LifecycleProbeState = {
  nodeId: string
  probeType: string
  status: string
  circuit?: { state?: string }
  runtime?: {
    childRunning?: boolean
    circuit?: { state?: string }
  }
}

export type LifecycleTarget = Pick<LifecycleProbeState, 'nodeId' | 'probeType'>

export const lifecycleTargetKey = (target: LifecycleTarget) =>
  `${target.nodeId}\u0000${target.probeType}`

export const isLifecycleRefreshCurrent = (startedVersion: number, currentVersion: number) =>
  startedVersion === currentVersion

export const isLifecycleStateReached = (state: LifecycleProbeState, running: boolean) => {
  const circuitState = state.runtime?.circuit?.state || state.circuit?.state
  if (running) {
    return (
      state.status === 'RUNNING' &&
      circuitState !== 'FUSED' &&
      state.runtime?.childRunning !== false
    )
  }
  return (
    state.status === 'FUSED' && circuitState === 'FUSED' && state.runtime?.childRunning === false
  )
}

export async function waitForLifecycleStates(
  loadStates: () => Promise<LifecycleProbeState[]>,
  targets: LifecycleTarget[],
  running: boolean,
  options: {
    maxAttempts?: number
    intervalMs?: number
    sleep?: (milliseconds: number) => Promise<void>
  } = {},
) {
  const maxAttempts = options.maxAttempts ?? 30
  const intervalMs = options.intervalMs ?? 1000
  const sleep =
    options.sleep ?? ((milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds)))
  const targetKeys = new Set(targets.map(lifecycleTargetKey))
  let states: LifecycleProbeState[] = []

  for (let attempt = 0; attempt < maxAttempts; attempt += 1) {
    states = await loadStates()
    const reachedKeys = new Set(
      states
        .filter((state) => targetKeys.has(lifecycleTargetKey(state)))
        .filter((state) => isLifecycleStateReached(state, running))
        .map(lifecycleTargetKey),
    )
    const pending = targets.filter((target) => !reachedKeys.has(lifecycleTargetKey(target)))
    if (!pending.length || attempt === maxAttempts - 1) return { states, pending }
    await sleep(intervalMs)
  }

  return { states, pending: targets }
}
