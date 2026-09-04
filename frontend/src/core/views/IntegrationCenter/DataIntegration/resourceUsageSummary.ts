type ProbeResourceUsage = {
  cpuPercentAvg?: number
  memoryBytesAvg?: number
}

type NodeResourceUsage = {
  cpuPercent?: number
  memoryBytes?: number
}

const warnedInvalidValues = new Set<string>()

const sumValidValues = (field: string, values: Array<number | undefined>) => {
  const validValues = values.filter((value): value is number => {
    if (value === undefined) return false
    if (Number.isFinite(value)) return true

    const warningKey = `${field}:${String(value)}`
    if (!warnedInvalidValues.has(warningKey)) {
      warnedInvalidValues.add(warningKey)
      console.error('[ProbeManagement] ignored invalid resource usage', { field, value })
    }
    return false
  })

  return validValues.length
    ? validValues.reduce((total, value) => total + value, 0)
    : undefined
}

export const summarizeProbeResourceUsage = (
  usages: Array<ProbeResourceUsage | undefined>,
) => ({
  cpuPercent: sumValidValues(
    'cpuPercentAvg',
    usages.map((usage) => usage?.cpuPercentAvg),
  ),
  memoryBytes: sumValidValues(
    'memoryBytesAvg',
    usages.map((usage) => usage?.memoryBytesAvg),
  ),
})

export const summarizeNodeResourceUsage = (nodes: NodeResourceUsage[]) => ({
  cpuPercent: sumValidValues(
    'cpuPercent',
    nodes.map((node) => node.cpuPercent),
  ),
  memoryBytes: sumValidValues(
    'memoryBytes',
    nodes.map((node) => node.memoryBytes),
  ),
})
