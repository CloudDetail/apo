import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

import {
  buildOverviewRequestDecision,
  hasOverviewContent,
  isLatestOverviewRequest,
  shouldRenderInitialOverviewLoading,
} from './overviewViewState.mjs'

const viewSource = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')
const cssSource = readFileSync(new URL('./index.css', import.meta.url), 'utf8')

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

test('overview refresh uses full-page loading even when stale content exists', () => {
  assert.equal(
    shouldRenderInitialOverviewLoading({
      overviewLoading: true,
      systems: [{ groupId: 1001, name: '订单系统' }],
      pendingIssues: [],
      summary: { totalSystems: 1 },
    }),
    true,
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

test('overview responses only apply for the latest request', () => {
  assert.equal(isLatestOverviewRequest(2, 2), true)
  assert.equal(isLatestOverviewRequest(1, 2), false)
})

test('overview auto refresh is consumed while an overview request is loading', () => {
  assert.deepEqual(
    buildOverviewRequestDecision({
      requestKey: 'next-window',
      lastRequestKey: 'current-window',
      overviewLoading: true,
    }),
    {
      shouldLoad: false,
      nextLastRequestKey: 'next-window',
    },
  )
})

test('overview loading skeleton follows the current overview layout', () => {
  const loadingStart = viewSource.indexOf('const OverviewInitialLoading')
  const pageStart = viewSource.indexOf('export default function SystemHealthOverviewPage')
  assert.notEqual(loadingStart, -1, 'overview loading component should exist')
  assert.notEqual(pageStart, -1, 'overview page component should exist')

  const loadingSource = viewSource.slice(loadingStart, pageStart)

  assert.ok(loadingSource.includes('{[0, 1, 2].map((item) => ('))
  assert.equal(loadingSource.includes('{[0, 1, 2, 3].map((item) => ('), false)
  assert.equal(loadingSource.includes('sho-loading-line-meter'), false)
  assert.equal(loadingSource.includes('sho-loading-line-caption'), false)
  assert.ok(loadingSource.includes('{metricLabels.map(([key]) => ('))
  assert.ok(
    loadingSource.includes(
      '<span className="sho-loading-table-cell sho-loading-table-cell-system" />',
    ),
  )
  assert.ok(
    loadingSource.includes(
      '<span className="sho-loading-table-cell sho-loading-table-cell-score" />',
    ),
  )
})

test('overview loading summary uses the same three-column grid as summary cards', () => {
  const match = cssSource.match(/\.sho-loading-summary\s*\{([^}]+)\}/)
  assert.ok(match, 'loading summary rule should exist')
  assert.ok(match[1].includes('grid-template-columns: repeat(3, minmax(0, 1fr))'))
})
