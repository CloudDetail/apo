import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildStatusDistribution,
  healthScoreLegendItems,
  healthScoreLegendItemsFor,
  healthScoreStatus,
  metricSourceTitleFor,
  metricVisualFor,
  normalizeHealthScoreThresholds,
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

test('buildStatusDistribution derives system severity from configured score thresholds', () => {
  const distribution = buildStatusDistribution(
    [
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
    ],
    {
      severeAbnormalScoreThreshold: 60,
      generalAbnormalScoreThreshold: 80,
    },
  )

  assert.deepEqual(distribution, {
    total: 4,
    red: 0,
    yellow: 1,
    green: 2,
    gray: 1,
  })
})

test('healthScoreStatus uses configured below-score thresholds', () => {
  const thresholds = {
    severeAbnormalScoreThreshold: 45,
    generalAbnormalScoreThreshold: 82,
  }

  assert.equal(healthScoreStatus(44, thresholds), 'red')
  assert.equal(healthScoreStatus(45, thresholds), 'yellow')
  assert.equal(healthScoreStatus(81, thresholds), 'yellow')
  assert.equal(healthScoreStatus(82, thresholds), 'green')
})

test('normalizeHealthScoreThresholds rejects impossible ranges', () => {
  assert.deepEqual(
    normalizeHealthScoreThresholds({
      severeAbnormalScoreThreshold: 45,
      generalAbnormalScoreThreshold: 82,
    }),
    {
      severeAbnormalScoreThreshold: 45,
      generalAbnormalScoreThreshold: 82,
    },
  )
  assert.deepEqual(
    normalizeHealthScoreThresholds({
      severeAbnormalScoreThreshold: 90,
      generalAbnormalScoreThreshold: 80,
    }),
    {
      severeAbnormalScoreThreshold: 60,
      generalAbnormalScoreThreshold: 80,
    },
  )
})

test('healthScoreLegendItemsFor documents configured score ranges without sample scores', () => {
  assert.deepEqual(
    healthScoreLegendItemsFor({
      severeAbnormalScoreThreshold: 45,
      generalAbnormalScoreThreshold: 82,
    }),
    [
      { status: 'green', range: '82-100分', label: '正常' },
      { status: 'yellow', range: '45-81分', label: '一般异常（关注系统）' },
      { status: 'red', range: '0-44分', label: '严重异常（异常系统）' },
      { status: 'gray', range: '无可评分数据', label: '无数据' },
    ],
  )
})

test('healthScoreLegendItems documents default score ranges without sample scores', () => {
  assert.deepEqual(healthScoreLegendItems, [
    { status: 'green', range: '80-100分', label: '正常' },
    { status: 'yellow', range: '60-79分', label: '一般异常（关注系统）' },
    { status: 'red', range: '0-59分', label: '严重异常（异常系统）' },
    { status: 'gray', range: '无可评分数据', label: '无数据' },
  ])
})

test('buildOverviewSummaryCards presents compact anomaly cards in priority order', () => {
  assert.equal(typeof metricVisuals.buildOverviewSummaryCards, 'function')

  const cards = metricVisuals.buildOverviewSummaryCards(
    {
      totalSystems: 12,
      abnormalSystems: 9,
      lowScoreSystems: 8,
      abnormalMetrics: 31,
      severeMetrics: 7,
    },
    [
      {
        name: '订单系统',
        score: 42,
        status: 'red',
        services: [{}],
        metrics: {
          errorRate: 'red',
          latency: 'yellow',
          errorLog: 'red',
        },
      },
      {
        name: '支付系统',
        score: 68,
        status: 'yellow',
        services: [{}],
        metrics: {
          latency: 'red',
          errorLog: 'yellow',
        },
      },
      {
        name: '会员系统',
        score: 84,
        status: 'green',
        services: [{}],
        metrics: {
          errorRate: 'yellow',
        },
      },
      {
        name: '无数据系统',
        score: 0,
        status: 'gray',
        services: [],
        metrics: {},
      },
    ],
  )

  assert.deepEqual(
    cards.map((card) => card.label),
    ['异常系统', '关注系统', '严重异常指标'],
  )
  assert.deepEqual(
    cards.map((card) => card.value),
    ['1 / 12', '1 / 12', 7],
  )
  assert.deepEqual(
    cards.map((card) => card.highlightNumerator),
    [true, true, undefined],
  )
  assert.deepEqual(
    cards.map((card) => card.caption),
    [undefined, undefined, undefined],
  )
})

test('buildOverviewSummaryCards presents abnormal systems as text without a ring', () => {
  const cards = metricVisuals.buildOverviewSummaryCards(
    {
      totalSystems: 4,
      abnormalSystems: 3,
      lowScoreSystems: 1,
    },
    [
      {
        score: 58,
        status: 'red',
        services: [{}],
      },
    ],
  )

  const abnormalSystemsCard = cards.find((card) => card.key === 'abnormalSystems')
  const attentionSystemsCard = cards.find((card) => card.key === 'attentionSystems')

  assert.equal(abnormalSystemsCard.visual, 'ratio')
  assert.equal(abnormalSystemsCard.help, '系统健康度低于 60 分，计为异常系统。')
  assert.equal(attentionSystemsCard.help, '系统健康度 60-79 分之间，计为关注系统。')
  assert.equal(abnormalSystemsCard.value, '1 / 4')
  assert.equal(abnormalSystemsCard.percent, undefined)
  assert.equal(abnormalSystemsCard.highlightNumerator, true)
})

test('buildOverviewSummaryCards describes configured abnormal score thresholds', () => {
  const cards = metricVisuals.buildOverviewSummaryCards(
    {
      totalSystems: 4,
      abnormalSystems: 3,
    },
    [
      {
        score: 44,
        status: 'red',
        services: [{}],
      },
    ],
    {
      severeAbnormalScoreThreshold: 45,
      generalAbnormalScoreThreshold: 82,
    },
  )

  const abnormalSystemsCard = cards.find((card) => card.key === 'abnormalSystems')
  const attentionSystemsCard = cards.find((card) => card.key === 'attentionSystems')

  assert.equal(abnormalSystemsCard.help, '系统健康度低于 45 分，计为异常系统。')
  assert.equal(attentionSystemsCard.help, '系统健康度 45-81 分之间，计为关注系统。')
  assert.equal(attentionSystemsCard.value, '0 / 4')
})

test('buildOverviewSummaryCards keeps zero abnormal system ratio neutral', () => {
  const cards = metricVisuals.buildOverviewSummaryCards(
    {
      totalSystems: 4,
      abnormalSystems: 0,
    },
    [
      {
        score: 88,
        status: 'green',
        services: [{}],
      },
    ],
  )

  const abnormalSystemsCard = cards.find((card) => card.key === 'abnormalSystems')

  assert.equal(abnormalSystemsCard.value, '0 / 4')
  assert.equal(abnormalSystemsCard.highlightNumerator, false)
})

test('buildOverviewSummaryCards presents severe metric counts as big numbers', () => {
  const cards = metricVisuals.buildOverviewSummaryCards({
    totalSystems: 4,
    abnormalMetrics: 12,
    severeMetrics: 3,
  })

  const severeMetricsCard = cards.find((card) => card.key === 'severeMetrics')

  assert.equal(severeMetricsCard.visual, 'number')
  assert.equal(severeMetricsCard.percent, undefined)
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
      showThresholdMarker: true,
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

test('metricVisualFor renders missing metric values as empty data text', () => {
  const visual = metricVisualFor('latency', {
    status: 'gray',
    value: '',
    note: '暂无数据',
  })
  const sentinelVisual = metricVisualFor('latency', {
    status: 'gray',
    value: '--',
    note: '暂无数据',
  })

  assert.equal(visual.value, '暂无数据')
  assert.equal(sentinelVisual.value, '暂无数据')
})

test('metricSourceTitleFor explains title tooltip values in readable language', () => {
  assert.equal(
    metricSourceTitleFor('resource', {
      source: 'dev/train-ticket/ts-train-service-787984cc59-8gbnh',
    }),
    '当前窗口内 CPU 或内存压力最高的 Pod 水位。最高水位 Pod：ts-train-service-787984cc59-8gbnh',
  )
  assert.equal(metricSourceTitleFor('errorRate'), '当前窗口内失败请求占全部请求的比例。')
  assert.equal(metricSourceTitleFor('latency'), '当前窗口内所有请求的平均响应时间。')
  assert.equal(
    metricSourceTitleFor('errorLog'),
    '当前窗口内 error、critical 和 exception 日志总数。',
  )
  assert.equal(metricSourceTitleFor('badRequests'), '当前仍在触发中的告警数量。')
  assert.equal(metricSourceTitleFor('lockCost'), '当前窗口内每个请求平均分摊到的锁等待耗时。')
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
      value: '12 条',
      note: '当前时间窗口错误日志数',
      percent: 92,
      helper: '窗口计数',
    },
  )
})

test('metricVisualFor renders active alerts as a plain big number with count unit', () => {
  const visual = metricVisualFor('badRequests', {
    status: 'yellow',
    value: '3',
    note: '当前活跃告警数',
  })

  assert.equal(visual.kind, 'number')
  assert.equal(visual.value, '3 个')
  assert.equal(visual.helper, '活跃告警')
  assert.equal('marker' in visual, false)
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
  assert.equal('marker' in visual, false)
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

test('buildMetricTrendOption omits baseline mark lines for resource shift', () => {
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
  assert.equal(option.series[0].markLine, undefined)
  assert.equal(option.xAxis.axisLabel.show, false)
  assert.equal(option.yAxis.axisLabel.show, false)
  assert.equal(option.yAxis.splitLine.show, false)
})
