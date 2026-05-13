import test from 'node:test'
import assert from 'node:assert/strict'

import { getNodeVisualStatus, getStatusColor } from './findingsTopologyPresentation'

test('hydrating downstream nodes use loading edge color instead of normal green', () => {
  assert.equal(
    getNodeVisualStatus({
      status: 'normal',
      visualStatus: 'normal',
      hydrationStatus: 'running',
    }),
    'loading',
  )
  assert.equal(
    getNodeVisualStatus({
      status: 'normal',
      visualStatus: 'normal',
      hydrationStatus: 'pending',
    }),
    'loading',
  )
  assert.notEqual(getStatusColor('loading'), getStatusColor('normal'))
})

test('completed nodes keep business visual status for edge color', () => {
  assert.equal(
    getNodeVisualStatus({
      status: 'normal',
      visualStatus: 'abnormal',
      hydrationStatus: 'success',
    }),
    'abnormal',
  )
})
