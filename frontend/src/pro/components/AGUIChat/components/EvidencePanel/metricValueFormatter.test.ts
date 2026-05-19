import test from 'node:test'
import assert from 'node:assert/strict'

import { formatMetricValue } from './metricValueFormatter'

test('formats log metrics with translated unit suffix', () => {
  const result = formatMetricValue(3.8, 'logs', {
    unit: '条',
    timesPerMinute: '次/分',
  })

  assert.equal(result, '3条')
})

test('formats latency values below 10ms as sub-cent precision', () => {
  const result = formatMetricValue(9.9, 'latency', {
    unit: '条',
    timesPerMinute: '次/分',
  })

  assert.equal(result, '< 0.01 ms')
})

test('formats tiny error rate values with lower-bound copy', () => {
  const result = formatMetricValue(0.009, 'errorRate', {
    unit: '条',
    timesPerMinute: '次/分',
  })

  assert.equal(result, '< 0.01%')
})

test('formats tiny throughput values with translated per-minute suffix', () => {
  const result = formatMetricValue(0.009, 'tps', {
    unit: '条',
    timesPerMinute: '次/分',
  })

  assert.equal(result, '< 0.01次/分')
})
