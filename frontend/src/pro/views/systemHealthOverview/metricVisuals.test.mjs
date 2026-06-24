import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildStatusDistribution,
  metricSourceTitleFor,
  metricVisualFor,
  safePercent,
  statusRatio,
} from './metricVisuals.mjs'
import * as metricVisuals from './metricVisuals.mjs'

test('safePercent clamps invalid and out-of-range ratios', () => {
  assert.equal(safePercent(3, 10), 30)
  assert.equal(safePercent(0, 0), 0)
  assert.equal(safePercent(-2, 10), 0)
  assert.equal(safePercent(15, 10), 100)
})

test('statusRatio maps health statuses to comparable risk fill', () => {
  assert.equal(statusRatio('red'), 92)
  assert.equal(statusRatio('yellow'), 64)
  assert.equal(statusRatio('green'), 24)
  assert.equal(statusRatio('gray'), 0)
  assert.equal(statusRatio('unknown'), 0)
})

test('buildStatusDistribution derives system severity from score, data, and metric status', () => {
  const distribution = buildStatusDistribution([
    {
      score: 95,
      status: 'green',
      services: [{}],
      metrics: { latency: 'green' },
    },
    {
      score: 88,
      status: 'green',
      services: [{}],
      metrics: { errorRate: 'red' },
    },
    {
      score: 74,
      status: 'yellow',
      services: [{}],
      metrics: { latency: 'green' },
    },
    {
      score: 0,
      status: 'gray',
      services: [],
      metrics: {},
    },
  ])

  assert.deepEqual(distribution, {
    total: 4,
    red: 1,
    yellow: 1,
    green: 1,
    gray: 1,
  })
})

test('buildOverviewSummaryCards restores prototype overview metrics', () => {
  assert.equal(typeof metricVisuals.buildOverviewSummaryCards, 'function')

  const cards = metricVisuals.buildOverviewSummaryCards(
    {
      totalSystems: 12,
      abnormalSystems: 4,
      lowScoreSystems: 2,
      abnormalMetrics: 31,
      severeMetrics: 7,
      newAbnormalMetrics: 5,
    },
    [
      {
        name: '订单系统',
        metrics: {
          errorRate: 'red',
          latency: 'yellow',
          errorLog: 'red',
        },
      },
      {
        name: '支付系统',
        metrics: {
          latency: 'red',
          errorLog: 'yellow',
        },
      },
      {
        name: '会员系统',
        metrics: {
          errorRate: 'yellow',
        },
      },
    ],
    [
      ['errorRate', '错误率'],
      ['latency', '响应延时'],
      ['errorLog', '错误日志'],
    ],
  )

  assert.deepEqual(
    cards.map((card) => card.label),
    ['异常系统', '异常指标', '严重异常', '新增异常'],
  )
  assert.deepEqual(
    cards.map((card) => card.value),
    ['4 / 12', 31, 7, 5],
  )
  assert.equal(cards[0].caption, '2 个健康度低于 60')
  assert.equal(cards[1].caption, '错误率、响应延时、错误日志占比最高')
  assert.equal(cards[2].caption, '集中在订单系统、支付系统')
  assert.equal(cards[3].caption, '最近 10 分钟新增')
})

test('buildOverviewSummaryCards presents abnormal systems as a right-side ring ratio', () => {
  const cards = metricVisuals.buildOverviewSummaryCards({
    totalSystems: 4,
    abnormalSystems: 3,
    lowScoreSystems: 1,
  })

  const abnormalSystemsCard = cards.find((card) => card.key === 'abnormalSystems')

  assert.equal(abnormalSystemsCard.visual, 'ring')
  assert.equal(abnormalSystemsCard.value, '3 / 4')
  assert.equal(abnormalSystemsCard.percent, 75)
})

test('buildOverviewSummaryCards presents absolute anomaly counts as big numbers', () => {
  const cards = metricVisuals.buildOverviewSummaryCards(
    {
      totalSystems: 4,
      abnormalMetrics: 12,
      severeMetrics: 3,
      newAbnormalMetrics: 8,
    },
    [],
    [
      ['errorRate', '错误率'],
      ['latency', '响应延时'],
    ],
  )

  const abnormalMetricsCard = cards.find((card) => card.key === 'abnormalMetrics')
  const severeMetricsCard = cards.find((card) => card.key === 'severeMetrics')

  assert.equal(abnormalMetricsCard.visual, 'number')
  assert.equal(abnormalMetricsCard.percent, undefined)
  assert.equal(severeMetricsCard.visual, 'number')
  assert.equal(severeMetricsCard.percent, undefined)
})

test('buildOverviewSummaryCards presents new abnormal metrics as a number instead of progress', () => {
  const cards = metricVisuals.buildOverviewSummaryCards(
    {
      totalSystems: 4,
      abnormalMetrics: 12,
      newAbnormalMetrics: 8,
    },
    [],
    [
      ['errorRate', '错误率'],
      ['latency', '响应延时'],
    ],
  )

  const newAbnormalCard = cards.find((card) => card.key === 'newAbnormalMetrics')

  assert.equal(newAbnormalCard.visual, 'number')
  assert.equal(newAbnormalCard.percent, undefined)
})

test('metricVisualFor turns resource metrics into waterline visuals', () => {
  assert.deepEqual(
    metricVisualFor('resource', {
      status: 'yellow',
      value: '83.40%',
      note: '当前 CPU 资源水位；资源副本健康存在风险',
    }),
    {
      kind: 'waterline',
      status: 'yellow',
      tone: 'warning',
      value: '83.40%',
      note: '当前 CPU 水位偏高',
      percent: 83,
      helper: '当前水位',
    },
  )
})

test('metricVisualFor keeps resource notes readable and does not expose source paths in the footnote', () => {
  const visual = metricVisualFor('resource', {
    status: 'red',
    value: '93.00%',
    note: '当前内存资源水位；资源副本健康存在风险',
    source: 'dev/train-ticket/ts-train-service-787984cc59-8gbnh',
  })

  assert.equal(visual.note, '当前内存水位偏高')
  assert.equal(visual.source, 'dev/train-ticket/ts-train-service-787984cc59-8gbnh')
  assert.equal(visual.note.includes('资源副本'), false)
  assert.equal(visual.note.includes('dev/train-ticket'), false)
})

test('metricVisualFor does not infer current resource waterline from status when value is missing', () => {
  const visual = metricVisualFor('resource', {
    status: 'red',
    value: '--',
    note: '资源副本健康存在风险',
  })

  assert.equal(visual.percent, 0)
  assert.equal(visual.note, '资源副本健康存在风险')
  assert.equal('marker' in visual, false)
})

test('metricSourceTitleFor summarizes metric data sources for title tooltips', () => {
  assert.equal(
    metricSourceTitleFor('resource', {
      source: 'dev/train-ticket/ts-train-service-787984cc59-8gbnh',
    }),
    '数据来源：Pod 资源水位，最高水位 Pod：ts-train-service-787984cc59-8gbnh',
  )
  assert.equal(
    metricSourceTitleFor('errorRate'),
    '数据来源：Trace 请求指标，系统错误请求数 / 总请求数',
  )
})

test('metricVisualFor scales resource shift against the abnormal threshold', () => {
  assert.deepEqual(
    metricVisualFor('resourceShift', {
      status: 'yellow',
      value: '45%',
      note: 'CPU/内存资源相对 24h 前基线波动',
    }),
    {
      kind: 'deviation',
      status: 'yellow',
      tone: 'warning',
      value: '45%',
      note: 'CPU/内存资源相对 24h 前基线波动',
      percent: 75,
      helper: '相对 24h 基线',
      marker: '60%',
    },
  )
})

test('metricVisualFor keeps latency as a big value with status risk gauge', () => {
  assert.deepEqual(
    metricVisualFor('latency', {
      status: 'yellow',
      value: '420ms',
      note: '当前平均响应延时',
    }),
    {
      kind: 'gauge',
      status: 'yellow',
      tone: 'warning',
      value: '420ms',
      note: '当前平均响应延时',
      percent: 64,
      helper: '延时风险',
      marker: '状态',
    },
  )
})

test('metricVisualFor formats count metrics as counter visuals', () => {
  assert.deepEqual(
    metricVisualFor('errorLog', {
      status: 'red',
      value: '12',
      note: '当前时间窗口错误日志数',
    }),
    {
      kind: 'counter',
      status: 'red',
      tone: 'critical',
      value: '12',
      note: '当前时间窗口错误日志数',
      percent: 92,
      helper: '窗口计数',
      marker: '严重度',
    },
  )
})

test('metricVisualFor renders bad requests as a plain big number without a meter', () => {
  const visual = metricVisualFor('badRequests', {
    status: 'yellow',
    value: '3',
    note: '当前活动告警数',
  })

  assert.equal(visual.kind, 'number')
  assert.equal(visual.value, '3')
  assert.equal(visual.helper, '活动告警')
  assert.equal(visual.marker, '当前值')
  assert.equal('percent' in visual, false)
})

test('metricVisualFor renders lock cost as a plain big number without a meter', () => {
  const visual = metricVisualFor('lockCost', {
    status: 'yellow',
    value: '12.00 ms/请求',
    note: '由 futex 锁等待耗时按请求数均摊计算',
  })

  assert.equal(visual.kind, 'number')
  assert.equal(visual.value, '12.00 ms/请求')
  assert.equal(visual.helper, '锁等待均摊')
  assert.equal(visual.marker, '当前值')
  assert.equal('percent' in visual, false)
})

test('metricVisualFor exposes trend metadata for system aggregate line metrics', () => {
  const visual = metricVisualFor('errorRate', {
    status: 'yellow',
    value: '3.40%',
    note: '系统请求错误率',
    unit: '%',
    series: [
      { timestamp: 101, value: 1.2 },
      { timestamp: 102, value: 3.4 },
    ],
  })

  assert.equal(visual.kind, 'gauge')
  assert.equal(visual.chart.colorRole, 'errorRate')
  assert.equal(visual.chart.unit, '%')
  assert.deepEqual(visual.chart.series, [
    { timestamp: 101, value: 1.2 },
    { timestamp: 102, value: 3.4 },
  ])
})

test('metricVisualFor keeps resource waterline as a current snapshot even if series is present', () => {
  const visual = metricVisualFor('resource', {
    status: 'red',
    value: '93.00%',
    note: '当前内存资源水位',
    source: 'dev/prod/order-b',
    series: [
      { timestamp: 101, value: 80 },
      { timestamp: 102, value: 93 },
    ],
  })

  assert.equal(visual.kind, 'waterline')
  assert.equal(visual.chart, null)
  assert.equal(visual.source, 'dev/prod/order-b')
})

test('buildMetricTrendOption adds threshold mark lines for resource shift', () => {
  assert.equal(typeof metricVisuals.buildMetricTrendOption, 'function')

  const option = metricVisuals.buildMetricTrendOption(
    {
      key: 'resourceShift',
      unit: '%',
      colorRole: 'resourceShift',
      series: [
        { timestamp: 101, value: 35 },
        { timestamp: 102, value: 75 },
      ],
      thresholds: [
        { label: '关注线', value: 30, status: 'yellow' },
        { label: '异常线', value: 60, status: 'red' },
      ],
    },
    {
      color: '#ff8a45',
      warningColor: '#f79009',
      criticalColor: '#f04438',
      textColor: '#475467',
      gridColor: '#eaecf0',
    },
  )

  assert.deepEqual(option.series[0].data, [35, 75])
  assert.deepEqual(
    option.series[0].markLine.data.map((line) => [line.name, line.yAxis]),
    [
      ['关注线', 30],
      ['异常线', 60],
    ],
  )
  assert.equal(option.xAxis.axisLabel.show, false)
  assert.equal(option.yAxis.axisLabel.show, false)
  assert.equal(option.yAxis.splitLine.show, false)
})
