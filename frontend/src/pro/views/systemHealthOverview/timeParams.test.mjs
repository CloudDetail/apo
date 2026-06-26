import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

import {
  buildOverviewRequestKey,
  buildOverviewRequestParams,
  buildStoreTimeParams,
} from './timeParams.mjs'

const viewSource = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')
const proRoutesSource = readFileSync(new URL('../../routes.js', import.meta.url), 'utf8')
const appHeaderSource = readFileSync(
  new URL('../../../core/components/AppHeader.js', import.meta.url),
  'utf8',
)
const dateTimeCombineSource = readFileSync(
  new URL('../../../core/components/DateTime/DateTimeCombine.jsx', import.meta.url),
  'utf8',
)

test('system health page uses shared concrete time request params', () => {
  assert.ok(
    viewSource.includes("from './timeParams.mjs'"),
    'page should use the tested time parameter builder',
  )
})

test('store time params use the concrete header time window even for relative ranges', () => {
  assert.deepEqual(
    buildStoreTimeParams({
      rangeTypeKey: '15m',
      startTime: 1719300000000000,
      endTime: 1719300900000000,
    }),
    {
      startTime: 1719300000000000,
      endTime: 1719300900000000,
    },
  )
})

test('overview request params prefer refreshed header timestamps over relative URL time', () => {
  assert.deepEqual(
    buildOverviewRequestParams(
      '?relativeTime=1h',
      {
        rangeTypeKey: '15m',
        startTime: 1719300000000000,
        endTime: 1719300900000000,
      },
      { preferStore: true },
    ),
    {
      startTime: 1719300000000000,
      endTime: 1719300900000000,
    },
  )
})

test('overview request params use header timestamps when URL only contains relative time', () => {
  assert.deepEqual(
    buildOverviewRequestParams('?relativeTime=1h', {
      rangeTypeKey: '1h',
      startTime: 1719297300000000,
      endTime: 1719300900000000,
    }),
    {
      startTime: 1719297300000000,
      endTime: 1719300900000000,
    },
  )
})

test('overview request params keep absolute URL time before the header time changes', () => {
  assert.deepEqual(
    buildOverviewRequestParams('?from=2026-06-25T10:00:00.000000Z&to=2026-06-25T10:15:00.000000Z', {
      rangeTypeKey: '15m',
      startTime: 1719300000000000,
      endTime: 1719300900000000,
    }),
    {
      startTime: 1782381600000000,
      endTime: 1782382500000000,
    },
  )
})

test('overview request params wait for initialized header time without absolute URL time', () => {
  assert.equal(buildOverviewRequestParams('', {}, { requireResolvedTimeRange: true }), null)
  assert.equal(
    buildOverviewRequestParams('?relativeTime=1h', {}, { requireResolvedTimeRange: true }),
    null,
  )
})

test('overview request params wait while persisted header time is rehydrating', () => {
  assert.equal(
    buildOverviewRequestParams(
      '',
      {
        rangeTypeKey: '15m',
        startTime: 1719300000000000,
        endTime: 1719300900000000,
        _persist: { rehydrated: false },
      },
      { requireResolvedTimeRange: true },
    ),
    null,
  )
  assert.deepEqual(
    buildOverviewRequestParams(
      '',
      {
        rangeTypeKey: '15m',
        startTime: 1719300000000000,
        endTime: 1719300900000000,
        _persist: { rehydrated: true },
      },
      { requireResolvedTimeRange: true },
    ),
    {
      startTime: 1719300000000000,
      endTime: 1719300900000000,
    },
  )
})

test('overview request key stays stable when URL and header resolve to the same window', () => {
  const urlParams = buildOverviewRequestParams(
    '?from=2026-06-25T10:00:00.000000Z&to=2026-06-25T10:15:00.000000Z',
    {},
    { requireResolvedTimeRange: true },
  )
  const storeParams = buildOverviewRequestParams(
    '?from=2026-06-25T10:00:00.000000Z&to=2026-06-25T10:15:00.000000Z',
    {
      rangeTypeKey: null,
      startTime: 1782381600000000,
      endTime: 1782382500000000,
    },
    { preferStore: true, requireResolvedTimeRange: true },
  )

  assert.equal(buildOverviewRequestKey(urlParams), buildOverviewRequestKey(storeParams))
})

test('system health page skips duplicate overview loads for the same request key', () => {
  assert.ok(
    viewSource.includes('const lastOverviewRequestKeyRef = useRef'),
    'page should remember the last overview request key',
  )
  assert.ok(
    viewSource.includes('buildOverviewRequestDecision({'),
    'page should use the overview request decision helper before loading',
  )
})

test('system health page debounces overview loads while time params settle', () => {
  assert.ok(
    viewSource.includes('const overviewRequestTimerRef = useRef'),
    'page should keep the pending overview request timer',
  )
  assert.ok(
    viewSource.includes('window.clearTimeout(overviewRequestTimerRef.current)'),
    'page should cancel a pending overview request when params change',
  )
  assert.ok(
    viewSource.includes('window.setTimeout(() => {'),
    'page should debounce the overview request instead of firing immediately',
  )
})

test('system health route disables header auto refresh', () => {
  const routeStart = proRoutesSource.indexOf("path: '/healthy-service'")
  assert.notEqual(routeStart, -1, 'healthy-service route should exist')
  const routeSource = proRoutesSource.slice(routeStart, routeStart + 220)

  assert.ok(routeSource.includes('disableSystemTimeAutoRefresh: true'))
  assert.ok(
    appHeaderSource.includes('disableAutoRefresh={currentRoute?.disableSystemTimeAutoRefresh}'),
  )
  assert.ok(
    dateTimeCombineSource.includes('<RefreshDateTime disableAutoRefresh={disableAutoRefresh} />'),
    'manual refresh should remain visible while auto refresh is disabled',
  )
})
