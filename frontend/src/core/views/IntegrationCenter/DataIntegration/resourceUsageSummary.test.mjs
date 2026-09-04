import assert from 'node:assert/strict'
import test from 'node:test'

import {
  summarizeNodeResourceUsage,
  summarizeProbeResourceUsage,
} from './resourceUsageSummary.ts'

test('summarizes healthy samples even when another probe has no sample', () => {
  assert.deepEqual(
    summarizeProbeResourceUsage([
      { cpuPercentAvg: 1.25, memoryBytesAvg: 100 },
      undefined,
      { cpuPercentAvg: 0.75, memoryBytesAvg: 200 },
    ]),
    { cpuPercent: 2, memoryBytes: 300 },
  )
})

test('summarizes CPU and memory independently', () => {
  assert.deepEqual(
    summarizeProbeResourceUsage([
      { cpuPercentAvg: 2 },
      { memoryBytesAvg: 512 },
    ]),
    { cpuPercent: 2, memoryBytes: 512 },
  )
})

test('keeps cluster usage when one node has no samples', () => {
  assert.deepEqual(
    summarizeNodeResourceUsage([
      { cpuPercent: 3, memoryBytes: 1024 },
      {},
      { cpuPercent: 2, memoryBytes: 2048 },
    ]),
    { cpuPercent: 5, memoryBytes: 3072 },
  )
})

test('returns unavailable only when no valid sample exists', () => {
  assert.deepEqual(summarizeProbeResourceUsage([undefined, {}]), {
    cpuPercent: undefined,
    memoryBytes: undefined,
  })
})
