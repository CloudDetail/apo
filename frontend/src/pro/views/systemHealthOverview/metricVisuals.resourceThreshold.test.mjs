import test from 'node:test'
import assert from 'node:assert/strict'

import { metricVisualFor } from './metricVisuals.mjs'

test('resource waterline hides the threshold marker when status is green', () => {
  assert.deepEqual(
    metricVisualFor('resource', {
      status: 'green',
      value: '24.00%',
      note: '当前 CPU 资源水位',
    }),
    {
      kind: 'waterline',
      status: 'green',
      tone: 'stable',
      value: '24.00%',
      note: '当前 CPU 水位正常',
      percent: 24,
      helper: '当前水位',
      showThresholdMarker: false,
    },
  )
})

test('resource waterline keeps the threshold marker when status is yellow', () => {
  assert.deepEqual(
    metricVisualFor('resource', {
      status: 'yellow',
      value: '83.40%',
      note: '当前 CPU 资源水位',
    }),
    {
      kind: 'waterline',
      status: 'yellow',
      tone: 'warning',
      value: '83.40%',
      note: '当前 CPU 水位偏高',
      percent: 83,
      helper: '当前水位',
      showThresholdMarker: true,
    },
  )
})
