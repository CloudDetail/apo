import test from 'node:test'
import assert from 'node:assert/strict'

import { hasOverviewContent, shouldRenderInitialOverviewLoading } from './overviewViewState.mjs'

test('initial overview loading is used before any overview data is available', () => {
  assert.equal(
    shouldRenderInitialOverviewLoading({
      overviewLoading: true,
      systems: [],
      pendingIssues: [],
      summary: {},
    }),
    true,
  )
})

test('overview refresh keeps existing content visible while loading', () => {
  assert.equal(
    shouldRenderInitialOverviewLoading({
      overviewLoading: true,
      systems: [{ groupId: 1001, name: '订单系统' }],
      pendingIssues: [],
      summary: { totalSystems: 1 },
    }),
    false,
  )
})

test('overview does not use initial loading once loading finishes', () => {
  assert.equal(
    shouldRenderInitialOverviewLoading({
      overviewLoading: false,
      systems: [],
      pendingIssues: [],
      summary: {},
    }),
    false,
  )
})

test('overview content exists when systems, pending issues, or summary data exists', () => {
  assert.equal(
    hasOverviewContent({
      systems: [{ groupId: 1001 }],
      pendingIssues: [],
      summary: {},
    }),
    true,
  )
  assert.equal(
    hasOverviewContent({
      systems: [],
      pendingIssues: [{ id: 'issue-1' }],
      summary: {},
    }),
    true,
  )
  assert.equal(
    hasOverviewContent({
      systems: [],
      pendingIssues: [],
      summary: { totalSystems: 0 },
    }),
    true,
  )
})
