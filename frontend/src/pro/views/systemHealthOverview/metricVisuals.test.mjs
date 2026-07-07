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
      { status: 'green', range: '82-100分', label: '健康' },
      { status: 'yellow', range: '45-81分', label: '关注（关注系统）' },
      { status: 'red', range: '0-44分', label: '异常（异常系统）' },
      { status: 'gray', range: '无可评分数据', label: '无数据' },
    ],
  )
})

test('healthScoreLegendItems documents default score ranges without sample scores', () => {
  assert.deepEqual(healthScoreLegendItems, [
    { status: 'green', range: '80-100分', label: '健康' },
    { status: 'yellow', range: '60-79分', label: '关注（关注系统）' },
    { status: 'red', range: '0-59分', label: '异常（异常系统）' },
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
    undefined,
    { pendingIssueTotal: 5 },
  )

  assert.deepEqual(
    cards.map((card) => card.label),
    ['异常系统', '关注系统', '待处理问题'],
  )
  assert.deepEqual(
    cards.map((card) => card.value),
    ['1 / 12', '1 / 12', 5],
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

test('buildOverviewSummaryCards keeps zero system ratios colored by severity', () => {
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
  const attentionSystemsCard = cards.find((card) => card.key === 'attentionSystems')

  assert.equal(abnormalSystemsCard.value, '0 / 4')
  assert.equal(attentionSystemsCard.value, '0 / 4')
  assert.equal(abnormalSystemsCard.highlightNumerator, true)
  assert.equal(attentionSystemsCard.highlightNumerator, true)
})

test('buildOverviewSummaryCards presents pending issue counts as big numbers', () => {
  const cards = metricVisuals.buildOverviewSummaryCards(
    {
      totalSystems: 4,
      abnormalMetrics: 12,
      severeMetrics: 3,
    },
    [],
    undefined,
    { pendingIssueTotal: 6 },
  )

  const pendingIssuesCard = cards.find((card) => card.key === 'pendingIssues')

  assert.equal(pendingIssuesCard.label, '待处理问题')
  assert.equal(pendingIssuesCard.value, 6)
  assert.equal(pendingIssuesCard.visual, 'number')
  assert.equal(pendingIssuesCard.percent, undefined)
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

test('metricVisualFor labels aggregate resource values as the judging waterline', () => {
  const visual = metricVisualFor('resource', {
    status: 'red',
    value: '100.00%',
    note: '系统资源水位聚合',
    badges: [
      { label: '平均水位', value: '36.12%', status: 'green' },
      { label: '高水位实例', value: '1/14', status: 'red' },
      { label: 'P95水位', value: '100.00%', status: 'red' },
    ],
  })

  assert.equal(visual.helper, '判定水位')
  assert.equal(visual.note, '系统判定水位过高')
  assert.equal(visual.value, '100.00%')
  assert.equal(visual.percent, 100)
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
    '数据组内服务关联 Pod 或 Host 去重后的整体资源水位。最高水位 Pod：ts-train-service-787984cc59-8gbnh',
  )
  assert.equal(metricSourceTitleFor('errorRate'), '数据组内总失败请求数占总请求数的比例。')
  assert.equal(metricSourceTitleFor('latency'), '数据组整体请求延时，当前使用请求数加权平均值。')
  assert.equal(
    metricSourceTitleFor('errorLog'),
    '数据组内可见服务 error、critical 和 exception 日志总数。',
  )
  assert.equal(metricSourceTitleFor('badRequests'), '数据组内当前 firing 告警按级别汇总后的数量。')
  assert.equal(metricSourceTitleFor('lockCost'), '数据组内 futex 锁等待总耗时按总请求数均摊后的耗时。')
})

test('metricVisualFor scales resource shift against the abnormal threshold', () => {
  assert.deepEqual(
    metricVisualFor('resourceShift', {
      status: 'yellow',
      value: '45.00%',
      note: 'CPU/内存资源水位较 24h 前变化',
    }),
    {
      kind: 'deviation',
      status: 'yellow',
      tone: 'warning',
      value: '45.00%',
      note: 'CPU/内存资源水位较 24h 前变化',
      percent: 75,
      helper: '较 24h 前',
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
    badges: [
      { label: '红色告警', value: '0', status: 'green' },
      { label: '黄色告警', value: '3', status: 'yellow' },
    ],
  })

  assert.equal(visual.kind, 'number')
  assert.equal(visual.value, '3 个')
  assert.equal(visual.helper, '活跃告警')
  assert.deepEqual(visual.badges, [
    { label: '红色告警', value: '0', status: 'green' },
    { label: '黄色告警', value: '3', status: 'yellow' },
  ])
  assert.equal('distribution' in visual, false)
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

test('metricVisualFor compacts large lock cost values for metric tiles', () => {
  const visual = metricVisualFor('lockCost', {
    status: 'red',
    value: '12660.00 ms/请求',
    note: '由 futex 锁等待耗时按请求数均摊计算',
  })

  assert.equal(visual.kind, 'number')
  assert.equal(visual.value, '12660 ms/请求')
})

test('metricVisualFor exposes trend metadata for system aggregate line metrics', () => {
  const visual = metricVisualFor('errorRate', {
    status: 'yellow',
    value: '3.40%',
    note: '系统请求错误率',
    unit: '%',
    seriesUnit: '%',
    seriesLabel: '最高服务错误率',
    series: [
      { timestamp: 101, value: 1.2 },
      { timestamp: 102, value: 3.4 },
    ],
  })

  assert.equal(visual.kind, 'gauge')
  assert.equal(visual.chart.colorRole, 'errorRate')
  assert.equal(visual.chart.unit, '%')
  assert.equal(visual.chart.label, '最高服务错误率')
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

test('buildMetricTrendOption omits baseline mark lines for noisy metric trends', () => {
  assert.equal(typeof metricVisuals.buildMetricTrendOption, 'function')

  ;[
    ['resourceShift', '%', 'resourceShift'],
    ['errorRate', '%', 'errorRate'],
    ['latency', 'ms', 'latency'],
  ].forEach(([key, unit, colorRole]) => {
    const option = metricVisuals.buildMetricTrendOption(
      {
        key,
        unit,
        colorRole,
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
})

test('buildMetricTrendOption keeps empty trend buckets as gaps', () => {
  const option = metricVisuals.buildMetricTrendOption({
    key: 'latency',
    unit: 'ms',
    colorRole: 'latency',
    series: [
      { timestamp: 101, value: 20 },
      { timestamp: 102, value: null },
      { timestamp: 103, value: 35 },
    ],
  })

  assert.deepEqual(option.series[0].data, [20, null, 35])
  assert.equal(option.series[0].connectNulls, false)
  assert.equal(option.graphic, undefined)
})

test('buildMetricTrendOption shows an in-chart empty state when all buckets lack data', () => {
  const option = metricVisuals.buildMetricTrendOption({
    key: 'latency',
    unit: 'ms',
    colorRole: 'latency',
    series: [
      { timestamp: 101, value: null },
      { timestamp: 102, value: undefined },
    ],
  })

  assert.deepEqual(option.series[0].data, [null, null])
  assert.equal(option.series[0].connectNulls, false)
  assert.equal(option.graphic.style.text, '无数据')
})
