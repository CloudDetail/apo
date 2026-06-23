import test from 'node:test'
import assert from 'node:assert/strict'

import { buildDetailRequestKey, shouldRenderDetailBody } from './detailViewState.mjs'

test('detail body stays hidden while detail data is loading', () => {
  assert.equal(
    shouldRenderDetailBody({
      selectedSystem: {
        groupId: 1001,
        name: '订单系统',
        services: [],
        metrics: {},
      },
      detailLoading: true,
    }),
    false,
  )
})

test('detail body renders after detail loading finishes', () => {
  assert.equal(
    shouldRenderDetailBody({
      selectedSystem: {
        groupId: 1001,
        name: '订单系统',
        services: [],
        metrics: {},
      },
      detailLoading: false,
    }),
    true,
  )
})

test('detail body does not render without a selected system', () => {
  assert.equal(
    shouldRenderDetailBody({
      selectedSystem: null,
      detailLoading: false,
    }),
    false,
  )
})

test('detail request key changes when the selected system time window changes', () => {
  const firstKey = buildDetailRequestKey('1001', {
    relativeTime: '15m',
  })
  const nextKey = buildDetailRequestKey('1001', {
    relativeTime: '1h',
  })

  assert.notEqual(firstKey, nextKey)
})

test('detail request key stays empty without a selected system', () => {
  assert.equal(
    buildDetailRequestKey('', {
      relativeTime: '15m',
    }),
    '',
  )
})
