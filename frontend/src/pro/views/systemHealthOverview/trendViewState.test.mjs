import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildHealthScoreBaselineItems,
  buildDetailTrendRows,
  buildTrendChartRows,
  DEFAULT_TREND_RANGE,
  isLongTrendGranularity,
  readStoredTrendRange,
  SYSTEM_HEALTH_TREND_RANGE_STORAGE_KEY,
  trendAxisUsesDateOnly,
  trendMetricGroups,
  trendMetricConfigs,
  trendPrimaryValue,
  trendRangeOptions,
  trendTooltipStats,
  writeStoredTrendRange,
} from './trendViewState.mjs'

test('trend metrics expose only system risk count and ratio charts', () => {
  assert.deepEqual(
    trendMetricGroups.map((group) => ({
      label: group.label,
      chartLabel: group.chartLabel,
      metrics: group.metrics.map((metric) => metric.label),
    })),
    [
      { label: '异常系统', chartLabel: '异常系统趋势图', metrics: ['数量趋势图', '占比趋势图'] },
      { label: '关注系统', chartLabel: '关注系统趋势图', metrics: ['数量趋势图', '占比趋势图'] },
    ],
  )

  assert.deepEqual(
    trendMetricConfigs.map((metric) => metric.label),
    ['数量趋势图', '占比趋势图', '数量趋势图', '占比趋势图'],
  )
  assert.deepEqual(
    trendMetricConfigs.map((metric) => metric.seriesLabel),
    ['数量', '占比', '数量', '占比'],
  )
  assert.equal(
    trendMetricConfigs.some((metric) => /异常系统|关注系统/.test(`${metric.label}${metric.seriesLabel}`)),
    false,
  )
})

test('trend range options include past 15 days between 7 days and one month', () => {
  assert.deepEqual(
    trendRangeOptions.map((option) => [option.key, option.label]),
    [
      ['today', '今天'],
      ['last24h', '近24小时'],
      ['thisWeek', '本周'],
      ['last7d', '过去7天'],
      ['last15d', '过去15天'],
      ['last30d', '过去一个月'],
      ['last6mo', '过去半年'],
      ['last1y', '过去一年'],
    ],
  )
})

test('trend range storage uses one shared key and rejects unsupported ranges', () => {
  const entries = new Map()
  const storage = {
    getItem: (key) => entries.get(key) ?? null,
    setItem: (key, value) => {
      entries.set(key, value)
    },
  }

  assert.equal(readStoredTrendRange(storage), DEFAULT_TREND_RANGE)

  writeStoredTrendRange('last15d', storage)
  assert.equal(entries.get(SYSTEM_HEALTH_TREND_RANGE_STORAGE_KEY), 'last15d')
  assert.equal(readStoredTrendRange(storage), 'last15d')

  entries.set(SYSTEM_HEALTH_TREND_RANGE_STORAGE_KEY, 'unsupported')
  assert.equal(readStoredTrendRange(storage), DEFAULT_TREND_RANGE)
  assert.equal(writeStoredTrendRange('unsupported', storage), DEFAULT_TREND_RANGE)
  assert.equal(entries.get(SYSTEM_HEALTH_TREND_RANGE_STORAGE_KEY), DEFAULT_TREND_RANGE)
})

test('trendPrimaryValue keeps risk count and ratio peaks visible', () => {
  const point = {
    abnormalSystems: { avg: 0.5, peak: 2, min: 0, last: 1, activeRatio: 0.5 },
    abnormalSystemRatio: { avg: 0.125, peak: 0.5, min: 0, last: 0.25, activeRatio: 0.5 },
  }

  assert.equal(trendPrimaryValue(point, 'abnormalSystems'), 2)
  assert.equal(trendPrimaryValue(point, 'abnormalSystemRatio'), 0.5)
})

test('trendPrimaryValue still supports system health score sparklines', () => {
  assert.equal(trendPrimaryValue({ systemHealthScore: { min: 61, max: 96, last: 94 } }, 'systemHealthScore'), 61)
})

test('trendPrimaryValue keeps missing system health score as no data', () => {
  assert.equal(trendPrimaryValue({ systemHealthScore: { min: 0, last: 0, sampleCount: 0 } }, 'systemHealthScore'), null)
})

test('buildTrendChartRows sorts points and keeps tooltip stats available', () => {
  const rows = buildTrendChartRows(
    [
      { timestamp: 3000, abnormalSystems: { avg: 0, peak: 0, last: 0, activeRatio: 0 } },
      { timestamp: 1000, abnormalSystems: { avg: 0.5, peak: 2, last: 1, activeRatio: 0.5 } },
    ],
    'abnormalSystems',
  )

  assert.deepEqual(
    rows.map((row) => row.timestamp),
    [1000, 3000],
  )
  assert.equal(rows[0].value, 2)
  assert.deepEqual(rows[0].stats, {
    avg: 0.5,
    peak: 2,
    min: 0,
    last: 1,
    ratio: 0.5,
    sampleCount: 0,
  })
})

test('buildTrendChartRows carries missing snapshot markers as gaps', () => {
  const rows = buildTrendChartRows(
    [{ timestamp: 1000, missingSnapshot: true, abnormalSystems: { peak: 0 } }],
    'abnormalSystems',
  )

  assert.deepEqual(rows, [
    {
      timestamp: 1000,
      missingSnapshot: true,
      value: null,
      stats: {
        avg: null,
        peak: null,
        min: null,
        last: null,
        ratio: 0,
        sampleCount: 0,
      },
    },
  ])
})

test('trendTooltipStats keeps percentage stats for ratio charts', () => {
  assert.deepEqual(
    trendTooltipStats(
      {
        attentionSystemRatio: {
          avg: 0.1,
          peak: 0.25,
          min: 0,
          last: 0.2,
          activeRatio: 0.5,
          sampleCount: 24,
        },
      },
      'attentionSystemRatio',
    ),
    {
      avg: 0.1,
      peak: 0.25,
      min: 0,
      last: 0.2,
      ratio: 0.5,
      sampleCount: 24,
    },
  )
})

test('buildDetailTrendRows combines system score and service health ratios', () => {
  const rows = buildDetailTrendRows([
    {
      timestamp: 3000,
      systemHealthScore: { min: 61, last: 72, sampleCount: 1 },
      abnormalServices: { avg: 1, peak: 1, last: 1, sampleCount: 1 },
      attentionServices: { avg: 1, peak: 1, last: 1, sampleCount: 1 },
      healthyServices: { avg: 2, peak: 2, last: 2, sampleCount: 1 },
      noDataServices: { avg: 0, peak: 0, last: 0, sampleCount: 1 },
    },
    {
      timestamp: 1000,
      systemHealthScore: { min: 88, last: 91, sampleCount: 1 },
      abnormalServices: { avg: 0, peak: 0, last: 0, sampleCount: 1 },
      attentionServices: { avg: 2, peak: 2, last: 2, sampleCount: 1 },
      healthyServices: { avg: 2, peak: 2, last: 2, sampleCount: 1 },
      noDataServices: { avg: 1, peak: 1, last: 1, sampleCount: 1 },
    },
  ])

  assert.deepEqual(
    rows.map((row) => row.timestamp),
    [1000, 3000],
  )
  assert.deepEqual(rows[0], {
    timestamp: 1000,
    score: 88,
    serviceDistributionMode: 'snapshot',
    serviceCounts: { abnormal: 0, attention: 2, healthy: 2, noData: 1 },
    serviceRatios: { abnormal: 0, attention: 0.4, healthy: 0.4, noData: 0.2 },
  })
  assert.deepEqual(rows[1], {
    timestamp: 3000,
    score: 61,
    serviceDistributionMode: 'snapshot',
    serviceCounts: { abnormal: 1, attention: 1, healthy: 2, noData: 0 },
    serviceRatios: { abnormal: 0.25, attention: 0.25, healthy: 0.5, noData: 0 },
  })
})

const aggregatedServicePoint = {
  timestamp: 1000,
  systemHealthScore: { min: 88, last: 91, sampleCount: 6 },
  abnormalServices: { avg: 0.2, peak: 1, last: 0, sampleCount: 6 },
  attentionServices: { avg: 4.5, peak: 7, last: 6, sampleCount: 6 },
  healthyServices: { avg: 8.1, peak: 10, last: 7, sampleCount: 6 },
  noDataServices: { avg: 1.2, peak: 3, last: 2, sampleCount: 6 },
}

test('buildDetailTrendRows uses last snapshot service counts for hourly distribution', () => {
  const rows = buildDetailTrendRows([aggregatedServicePoint], 'hour')

  assert.deepEqual(rows, [
    {
      timestamp: 1000,
      score: 88,
      serviceDistributionMode: 'snapshot',
      serviceCounts: { abnormal: 0, attention: 6, healthy: 7, noData: 2 },
      serviceRatios: { abnormal: 0, attention: 6 / 15, healthy: 7 / 15, noData: 2 / 15 },
    },
  ])
})

test('buildDetailTrendRows uses average service ratios for aggregated distribution buckets', () => {
  const rows = buildDetailTrendRows([aggregatedServicePoint], '6h')
  const serviceTotal = 0.2 + 4.5 + 8.1 + 1.2

  assert.deepEqual(rows, [
    {
      timestamp: 1000,
      score: 88,
      serviceDistributionMode: 'average',
      serviceCounts: { abnormal: 0.2, attention: 4.5, healthy: 8.1, noData: 1.2 },
      serviceRatios: {
        abnormal: 0.2 / serviceTotal,
        attention: 4.5 / serviceTotal,
        healthy: 8.1 / serviceTotal,
        noData: 1.2 / serviceTotal,
      },
    },
  ])
})

test('buildDetailTrendRows preserves missing snapshots without service no-data ratios', () => {
  const rows = buildDetailTrendRows([{ timestamp: 1000, missingSnapshot: true }])

  assert.deepEqual(rows, [
    {
      timestamp: 1000,
      missingSnapshot: true,
      score: null,
      serviceCounts: { abnormal: null, attention: null, healthy: null, noData: null },
      serviceRatios: { abnormal: null, attention: null, healthy: null, noData: null },
    },
  ])
})

test('buildDetailTrendRows keeps gray system score as no data while preserving service distribution', () => {
  const rows = buildDetailTrendRows([
    {
      timestamp: 1000,
      systemHealthScore: { min: 0, last: 0, sampleCount: 0 },
      abnormalServices: { avg: 0, peak: 0, last: 0, sampleCount: 1 },
      attentionServices: { avg: 0, peak: 0, last: 0, sampleCount: 1 },
      healthyServices: { avg: 0, peak: 0, last: 0, sampleCount: 1 },
      noDataServices: { avg: 2, peak: 2, last: 2, sampleCount: 1 },
    },
  ])

  assert.deepEqual(rows, [
    {
      timestamp: 1000,
      score: null,
      serviceDistributionMode: 'snapshot',
      serviceCounts: { abnormal: 0, attention: 0, healthy: 0, noData: 2 },
      serviceRatios: { abnormal: 0, attention: 0, healthy: 0, noData: 1 },
    },
  ])
})

test('isLongTrendGranularity identifies aggregated day and week ranges', () => {
  assert.equal(isLongTrendGranularity('hour'), false)
  assert.equal(isLongTrendGranularity('6h'), false)
  assert.equal(isLongTrendGranularity('day'), true)
  assert.equal(isLongTrendGranularity('week'), true)
})

test('trendAxisUsesDateOnly follows daily-or-larger point intervals', () => {
  const day = 24 * 60 * 60 * 1000
  const rows = [
    { timestamp: Date.UTC(2026, 6, 1, 0, 0, 0) },
    { timestamp: Date.UTC(2026, 6, 2, 0, 0, 0) },
  ]

  assert.equal(trendAxisUsesDateOnly(rows, 'hour'), true)
  assert.equal(trendAxisUsesDateOnly(rows, '6h'), true)
  assert.equal(trendAxisUsesDateOnly(rows, 'day'), true)
  assert.equal(trendAxisUsesDateOnly(rows, 'week'), true)
  assert.equal(
    trendAxisUsesDateOnly(
      [
        { timestamp: rows[0].timestamp },
        { timestamp: rows[0].timestamp + day / 4 },
      ],
      '6h',
    ),
    false,
  )
  assert.equal(trendAxisUsesDateOnly([{ timestamp: rows[0].timestamp }], 'hour'), false)
})

test('health score trend baselines follow current score judgement thresholds', () => {
  assert.deepEqual(buildHealthScoreBaselineItems(), [
    { key: 'attentionBaseline', label: '关注基线', value: 80, tone: 'warning' },
    { key: 'abnormalBaseline', label: '异常基线', value: 60, tone: 'critical' },
  ])

  assert.deepEqual(
    buildHealthScoreBaselineItems({
      severeAbnormalScoreThreshold: 72,
      generalAbnormalScoreThreshold: 88,
    }),
    [
      { key: 'attentionBaseline', label: '关注基线', value: 88, tone: 'warning' },
      { key: 'abnormalBaseline', label: '异常基线', value: 72, tone: 'critical' },
    ],
  )
})
