import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

import { buildDetailRequestKey, shouldRenderDetailBody } from './detailViewState.mjs'

const viewSource = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')

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

test('detail header omits removed guidance and summary cards', () => {
  assert.equal(viewSource.includes('按服务查看健康状态，优先处理红灯和持续黄灯指标。'), false)
  assert.equal(viewSource.includes('关注指标'), false)
  assert.equal(viewSource.includes('系统关键服务'), false)
  assert.equal(viewSource.includes('sho-detail-summary'), false)
  assert.equal(viewSource.includes('sho-score-card'), false)
})

test('detail header uses a compact single-row health monitor layout', () => {
  assert.ok(
    viewSource.includes(
      "className={`system-health-overview-page${selectedSystemId ? ' system-health-overview-page-detail' : ''}`}",
    ),
  )

  const headerStart = viewSource.indexOf('<div className="sho-detail-compact-header">')
  const metricsStart = viewSource.indexOf(
    '<section className="sho-panel sho-system-metrics-panel">',
  )

  assert.notEqual(headerStart, -1, 'compact detail header markup should exist')
  assert.notEqual(metricsStart, -1, 'compact detail header should render before metric panels')

  const headerSource = viewSource.slice(headerStart, metricsStart)
  assert.match(headerSource, /<Button\s+className="sho-detail-compact-back"/)
  assert.ok(headerSource.includes('<div className="sho-detail-compact-title">'))
  assert.ok(headerSource.includes('score={selectedSystem.score}'))
  assert.ok(headerSource.includes('scoreThresholds={scoreThresholds}'))
  assert.ok(headerSource.includes('size="mini"'))
  assert.ok(headerSource.includes('<div className="sho-detail-compact-score-label">'))
  assert.ok(headerSource.includes('<div className="sho-detail-compact-exception-label">'))
  assert.equal(headerSource.includes('SYSTEM'), false)
  assert.equal(headerSource.includes('SCORE'), false)
  assert.equal(headerSource.includes('sho-detail-compact-update'), false)
  assert.equal(headerSource.includes('Last Update'), false)
  assert.equal(headerSource.includes('RefreshCw'), false)
  assert.equal(headerSource.includes('formatDetailGeneratedAt(detailGeneratedAt)'), false)
})

test('detail header renders loading state instead of stale score while detail loads', () => {
  const headerStart = viewSource.indexOf('<div className="sho-detail-compact-header">')
  const loadingStart = viewSource.indexOf('{detailLoading ? (', headerStart)
  const metricsStart = viewSource.indexOf(
    '<section className="sho-panel sho-system-metrics-panel">',
  )

  assert.notEqual(headerStart, -1, 'compact detail header markup should exist')
  assert.notEqual(loadingStart, -1, 'compact detail header should branch on detailLoading')
  assert.ok(loadingStart < metricsStart, 'detail loading branch should live in the compact header')

  const headerLoadingSource = viewSource.slice(loadingStart, metricsStart)
  assert.ok(
    headerLoadingSource.includes('<DetailLoadingStatus className="sho-detail-compact-loading" />'),
  )
  assert.equal(headerLoadingSource.includes('<Spin size="small" />'), false)
  assert.ok(viewSource.includes('className="sho-detail-loading-pulse"'))
  assert.ok(viewSource.includes('正在加载系统详情'))
})

test('detail body loading state uses a structured skeleton panel', () => {
  assert.ok(viewSource.includes('const DetailLoadingPanel = () => ('))
  assert.ok(viewSource.includes('className="sho-panel sho-detail-loading"'))
  assert.ok(viewSource.includes('className="sho-detail-loading-metrics"'))
  assert.ok(viewSource.includes('{detailLoading && <DetailLoadingPanel />}'))
  assert.equal(viewSource.includes('<div className="sho-state-panel sho-detail-loading">'), false)
})

test('health page uses active alert wording and health metric section title', () => {
  assert.equal(viewSource.includes('异常请求'), false)
  assert.ok(viewSource.includes("['badRequests', '活跃告警']"))
  assert.ok(viewSource.includes('<div className="sho-panel-title">健康指标</div>'))
  assert.ok(viewSource.includes('<div className="sho-panel-title">服务健康度</div>'))
})

test('metric cards do not render gray helper descriptions under values', () => {
  assert.equal(viewSource.includes('<span>{visual.helper}</span>'), false)
})
