import test from 'node:test'
import assert from 'node:assert/strict'

import { shouldRenderDetailBody } from './detailViewState.mjs'

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
