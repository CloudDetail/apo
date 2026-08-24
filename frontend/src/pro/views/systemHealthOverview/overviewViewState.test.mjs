import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

import * as overviewViewState from './overviewViewState.mjs'
import {
  buildOverviewRequestDecision,
  hasOverviewContent,
  isLatestOverviewRequest,
  shouldRenderInitialOverviewLoading,
} from './overviewViewState.mjs'

const viewSource = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')
const cssSource = readFileSync(new URL('./index.css', import.meta.url), 'utf8')
const overviewInitialLoadingSource = readFileSync(
  new URL('./components/OverviewInitialLoading.jsx', import.meta.url),
  'utf8',
)

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

test('system row refresh stays scoped to the system panel loading state', () => {
  const renderOverviewStart = viewSource.indexOf('const renderOverview = () => {')
  const renderOverviewEnd = viewSource.indexOf('const renderDetail = () => {', renderOverviewStart)
  assert.notEqual(renderOverviewStart, -1, 'renderOverview should exist')
  assert.notEqual(renderOverviewEnd, -1, 'renderDetail should exist after renderOverview')

  const renderOverviewSource = viewSource.slice(renderOverviewStart, renderOverviewEnd)
  assert.equal(
    renderOverviewSource.includes(
      'const effectiveOverviewLoading = !queryParams || overviewLoading || systemRowsLoading',
    ),
    false,
    'systemRowsLoading should not trigger the full overview loading shell',
  )
  assert.ok(
    renderOverviewSource.includes('overviewLoading: !queryParams || overviewLoading,'),
    'full overview loading should only depend on missing query params or overview loading',
  )

  const overviewStateStart = viewSource.indexOf('const renderOverviewState = () => {')
  const overviewStateEnd = viewSource.indexOf('const renderOverview = () => {', overviewStateStart)
  assert.notEqual(overviewStateStart, -1, 'renderOverviewState should exist')
  assert.notEqual(overviewStateEnd, -1, 'renderOverview should exist after renderOverviewState')

  const overviewStateSource = viewSource.slice(overviewStateStart, overviewStateEnd)
  assert.ok(
    overviewStateSource.includes('overviewLoading || systemRowsLoading'),
    'systemRowsLoading should still render the system panel loading state',
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

test('overview empty state distinguishes no matching systems from no health data', () => {
  assert.equal(typeof overviewViewState.overviewSystemEmptyMessage, 'function')

  assert.equal(
    overviewViewState.overviewSystemEmptyMessage({
      systems: [],
      systemPageTotal: 0,
      systemHealthFilter: 'abnormal',
      systemKeyword: '',
    }),
    '没有匹配的系统',
  )
  assert.equal(
    overviewViewState.overviewSystemEmptyMessage({
      systems: [],
      systemPageTotal: 0,
      systemHealthFilter: 'all',
      systemKeyword: '',
    }),
    '暂无系统健康度数据',
  )
  assert.equal(
    overviewViewState.overviewSystemEmptyMessage({
      systems: [],
      systemPageTotal: 0,
      systemHealthFilter: 'all',
      systemKeyword: '订单',
    }),
    '没有匹配的系统',
  )
})

test('service empty state distinguishes filters from missing service health data', () => {
  assert.equal(typeof overviewViewState.overviewServiceEmptyMessage, 'function')

  assert.equal(
    overviewViewState.overviewServiceEmptyMessage({
      services: [],
      servicePageTotal: 0,
      serviceHealthFilter: 'abnormal',
      serviceKeyword: '',
    }),
    '没有匹配的服务',
  )
  assert.equal(
    overviewViewState.overviewServiceEmptyMessage({
      services: [],
      servicePageTotal: 0,
      serviceHealthFilter: 'all',
      serviceKeyword: '',
    }),
    '暂无服务健康度数据',
  )
  assert.equal(
    overviewViewState.overviewServiceEmptyMessage({
      services: [],
      servicePageTotal: 4,
      serviceHealthFilter: 'all',
      serviceKeyword: '',
    }),
    '没有匹配的服务',
  )
  assert.equal(
    overviewViewState.overviewServiceEmptyMessage({
      services: [{ name: 'order-api' }],
      servicePageTotal: 1,
      serviceHealthFilter: 'all',
      serviceKeyword: '',
    }),
    '',
  )
})

test('overview responses only apply for the latest request', () => {
  assert.equal(isLatestOverviewRequest(2, 2), true)
  assert.equal(isLatestOverviewRequest(1, 2), false)
})

test('overview auto refresh waits while an overview request is loading', () => {
  assert.deepEqual(
    buildOverviewRequestDecision({
      requestKey: 'next-window',
      lastRequestKey: 'current-window',
      overviewLoading: true,
    }),
    {
      shouldLoad: false,
      nextLastRequestKey: 'current-window',
    },
  )
})

test('overview loading skeleton follows the current overview layout', () => {
  const loadingStart = overviewInitialLoadingSource.indexOf('export default function OverviewInitialLoading')
  const pageStart = viewSource.indexOf('export default function SystemHealthOverviewPage')
  assert.notEqual(loadingStart, -1, 'overview loading component should exist')
  assert.notEqual(pageStart, -1, 'overview page component should exist')

  const loadingSource = overviewInitialLoadingSource.slice(loadingStart)

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
  assert.ok(viewSource.includes('<OverviewInitialLoading metricLabels={metricLabels} />'))
})

test('overview loading summary uses the same narrower pending-issue column as summary cards', () => {
  const match = cssSource.match(/\.sho-loading-summary\s*\{([^}]+)\}/)
  assert.ok(match, 'loading summary rule should exist')
  assert.ok(
    match[1].includes(
      'grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) minmax(220px, 0.62fr)',
    ),
  )
})
