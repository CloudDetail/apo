import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

import {
  buildServiceHealthAssistantIssues,
  buildAbnormalServiceRatio,
  buildDetailHeaderRequestKey,
  compareHealthScoreAscending,
  DEFAULT_HEALTH_FILTER_VALUE,
  DEFAULT_HEALTH_FILTER_VALUES,
  serviceAiAssistantActionFor,
  serviceAiAssistantActionLabel,
  shouldSyncDetailHeaderFromRows,
  targetMatchesHealthFilter,
  shouldRenderDetailBody,
} from './detailViewState.mjs'

const sourceFor = (path) => readFileSync(new URL(path, import.meta.url), 'utf8')

const viewSource = sourceFor('./index.jsx')
const detailLoadingSource = sourceFor('./components/DetailLoading.jsx')
const healthStatusSource = sourceFor('./components/HealthStatus.jsx')
const pendingIssueQueueSource = sourceFor('./components/PendingIssueQueue.jsx')
const overviewSystemTableSource = sourceFor('./components/OverviewSystemTable.jsx')
const detailHeaderSource = sourceFor('./components/DetailHeader.jsx')
const serviceHealthTableSource = sourceFor('./components/ServiceHealthTable.jsx')
const serviceHealthAssistantModalSource = sourceFor('./components/ServiceHealthAssistantModal.jsx')
const detailTrendSectionSource = sourceFor('./components/DetailTrendSection.jsx')
const metricVisualTileSource = sourceFor('./components/MetricVisualTile.jsx')
const trendCommonSource = sourceFor('./components/TrendCommon.jsx')

const selectBlockWithPlaceholder = (source, placeholder) => {
  const placeholderIndex = source.indexOf(`placeholder="${placeholder}"`)
  assert.notEqual(placeholderIndex, -1, `${placeholder} select should exist`)
  const selectStart = source.lastIndexOf('<Select', placeholderIndex)
  const selectEnd = source.indexOf('/>', placeholderIndex)
  assert.notEqual(selectStart, -1, `${placeholder} select should start`)
  assert.notEqual(selectEnd, -1, `${placeholder} select should close`)
  return source.slice(selectStart, selectEnd)
}

test('detail body stays hidden while detail header refreshes', () => {
  assert.equal(
    shouldRenderDetailBody({
      selectedSystem: {
        groupId: 1001,
        name: '订单系统',
        services: [],
        metrics: {},
      },
      detailHeaderLoading: true,
    }),
    false,
  )
})

test('detail body renders after detail header refresh finishes', () => {
  assert.equal(
    shouldRenderDetailBody({
      selectedSystem: {
        groupId: 1001,
        name: '订单系统',
        services: [],
        metrics: {},
      },
      detailHeaderLoading: false,
    }),
    true,
  )
})

test('detail body does not render without a selected system', () => {
  assert.equal(
    shouldRenderDetailBody({
      selectedSystem: null,
      detailHeaderLoading: false,
    }),
    false,
  )
})

test('detail header request key tracks selected system and time window', () => {
  const firstKey = buildDetailHeaderRequestKey('1001', {
    relativeTime: '15m',
  })
  const nextKey = buildDetailHeaderRequestKey('1001', {
    relativeTime: '1h',
  })

  assert.notEqual(firstKey, nextKey)
  assert.equal(buildDetailHeaderRequestKey('', { relativeTime: '15m' }), '')
})

test('detail header only syncs from system rows loaded for the current time window', () => {
  const row = { groupId: 1001, name: '订单系统' }

  assert.equal(shouldSyncDetailHeaderFromRows(row, '15m', '1h'), false)
  assert.equal(shouldSyncDetailHeaderFromRows(row, '1h', '1h'), true)
  assert.equal(shouldSyncDetailHeaderFromRows(null, '1h', '1h'), false)
})

test('detail view skips the header request on row click and refreshes it on time changes', () => {
  assert.ok(viewSource.includes('const lastDetailHeaderRequestKeyRef = useRef'))
  assert.ok(
    viewSource.includes('const [detailHeaderLoading, setDetailHeaderLoading] = useState(false)'),
  )
  assert.ok(
    viewSource.includes('lastDetailHeaderRequestKeyRef.current = buildDetailHeaderRequestKey'),
  )
  assert.ok(viewSource.includes('lastDetailHeaderRequestKeyRef.current === detailHeaderRequestKey'))
  assert.ok(viewSource.includes('setDetailHeaderLoading(true)'))
  assert.ok(viewSource.includes('setDetailHeaderLoading(false)'))
  assert.ok(viewSource.includes('getSystemHealthSystemDetailApi(selectedSystemId, {'))
  assert.ok(viewSource.includes('includeServices: false'))
  assert.equal(viewSource.includes('detailRefreshToken'), false)
})

test('detail header ratio counts red and yellow health score services separately', () => {
  assert.deepEqual(
    buildAbnormalServiceRatio(
      [
        { score: 59, status: 'green', metrics: { latency: 'green' } },
        { score: 60, status: 'red', metrics: { latency: 'red' } },
        { score: 79, status: 'red', metrics: { latency: 'red' } },
        { score: 80, status: 'yellow', metrics: { latency: 'yellow' } },
        { score: 95, status: 'yellow', metrics: { latency: 'yellow' } },
        { score: 90, status: 'green', metrics: { latency: 'red' } },
      ],
      {
        severeAbnormalScoreThreshold: 60,
        generalAbnormalScoreThreshold: 80,
      },
    ),
    { abnormal: 1, attention: 2, total: 6 },
  )
})

test('detail header ratio handles missing services', () => {
  assert.deepEqual(buildAbnormalServiceRatio(null), { abnormal: 0, attention: 0, total: 0 })
})

test('detail header ratio follows custom red health score threshold', () => {
  assert.deepEqual(
    buildAbnormalServiceRatio(
      [
        { score: 84, status: 'green', metrics: { latency: 'green' } },
        { score: 85, status: 'green', metrics: { latency: 'green' } },
      ],
      {
        severeAbnormalScoreThreshold: 85,
        generalAbnormalScoreThreshold: 90,
      },
    ),
    { abnormal: 1, attention: 1, total: 2 },
  )
})

test('detail ratio and filter treat gray zero as no data, not computed zero', () => {
  const scoreThresholds = {
    severeAbnormalScoreThreshold: 60,
    generalAbnormalScoreThreshold: 80,
  }
  const computedZero = { score: 0, status: 'red', metrics: { latency: 'red' } }
  const noDataZero = { score: 0, status: 'gray', metrics: { latency: 'gray' } }

  assert.deepEqual(
    buildAbnormalServiceRatio(
      [
        computedZero,
        noDataZero,
        { score: 59, status: 'red', metrics: { latency: 'red' } },
        { score: 79, status: 'yellow', metrics: { latency: 'yellow' } },
      ],
      scoreThresholds,
    ),
    { abnormal: 2, attention: 1, total: 4 },
  )
  assert.equal(
    targetMatchesHealthFilter(computedZero, DEFAULT_HEALTH_FILTER_VALUES.abnormal, scoreThresholds),
    true,
  )
  assert.equal(
    targetMatchesHealthFilter(noDataZero, DEFAULT_HEALTH_FILTER_VALUES.abnormal, scoreThresholds),
    false,
  )
  assert.equal(
    targetMatchesHealthFilter(noDataZero, DEFAULT_HEALTH_FILTER_VALUES.noData, scoreThresholds),
    true,
  )
  assert.equal(
    targetMatchesHealthFilter(noDataZero, DEFAULT_HEALTH_FILTER_VALUES.all, scoreThresholds),
    true,
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

  const metricsStart = viewSource.indexOf(
    '<section className="sho-panel sho-system-metrics-panel">',
  )

  assert.ok(viewSource.includes('<DetailHeader'))
  assert.notEqual(metricsStart, -1, 'compact detail header should render before metric panels')

  const headerSource = detailHeaderSource
  const healthLabelStart = headerSource.indexOf('<div className="sho-detail-compact-score-label">')
  const healthScoreStart = headerSource.indexOf('<HealthRing')
  assert.match(headerSource, /<Button\s+className="sho-detail-compact-back"/)
  assert.ok(headerSource.includes('<div className="sho-detail-compact-title">'))
  assert.ok(headerSource.includes('score={selectedSystem.score}'))
  assert.ok(headerSource.includes('scoreThresholds={scoreThresholds}'))
  assert.ok(headerSource.includes('size="mini"'))
  assert.ok(headerSource.includes('<div className="sho-detail-compact-score-label">'))
  assert.ok(healthLabelStart < healthScoreStart)
  assert.ok(headerSource.includes('<div className="sho-detail-compact-service-count">'))
  assert.ok(
    headerSource.indexOf('<span>异常服务</span>') < headerSource.indexOf('<span>关注服务</span>'),
  )
  assert.ok(headerSource.includes('<TriangleAlert size={18} />'))
  assert.ok(headerSource.includes('<span>异常服务</span>'))
  assert.ok(
    headerSource.includes(
      'aria-label={`异常服务 ${abnormalServiceRatio.abnormal} / ${abnormalServiceRatio.total}`}',
    ),
  )
  assert.ok(headerSource.includes('<CircleAlert size={18} />'))
  assert.ok(headerSource.includes('<span>关注服务</span>'))
  assert.ok(
    headerSource.includes(
      'aria-label={`关注服务 ${abnormalServiceRatio.attention} / ${abnormalServiceRatio.total}`}',
    ),
  )
  assert.equal(headerSource.includes('<span>异常服</span>'), false)
  assert.equal(headerSource.includes('异常服务数'), false)
  assert.ok(viewSource.includes('buildSystemServiceRiskCounts(selectedSystem, scoreThresholds)'))
  assert.ok(viewSource.includes('buildAbnormalServiceRatio(system.services, scoreThresholds)'))
  assert.ok(headerSource.includes('abnormalServiceRatio.abnormal'))
  assert.ok(headerSource.includes('abnormalServiceRatio.attention'))
  assert.ok(headerSource.includes('abnormalServiceRatio.total'))
  assert.ok(headerSource.includes('className="sho-detail-compact-service-value sho-summary-value"'))
  assert.ok(headerSource.includes('sho-summary-value-highlight-critical'))
  assert.ok(headerSource.includes('sho-summary-value-highlight-warning'))
  assert.equal(headerSource.includes('abnormalServiceRatio.abnormal > 0'), false)
  assert.equal(headerSource.includes('abnormalServiceRatio.attention > 0'), false)
  assert.ok(headerSource.includes('<div className="sho-detail-compact-exception-label">'))
  assert.ok(headerSource.includes('<Activity size={18} />'))
  assert.ok(headerSource.includes('<span>异常指标</span>'))
  assert.equal(headerSource.includes('严重异常指标'), false)
  assert.equal(headerSource.includes('主要异常'), false)
  assert.equal(headerSource.includes('SYSTEM'), false)
  assert.equal(headerSource.includes('SCORE'), false)
  assert.equal(headerSource.includes('sho-detail-compact-update'), false)
  assert.equal(headerSource.includes('Last Update'), false)
  assert.equal(headerSource.includes('RefreshCw'), false)
  assert.equal(headerSource.includes('formatDetailGeneratedAt(detailGeneratedAt)'), false)
})

test('detail header renders loading state instead of stale score while detail refreshes', () => {
  const loadingStart = detailHeaderSource.indexOf('{detailHeaderLoading ? (')

  assert.ok(viewSource.includes('<DetailHeader'))
  assert.notEqual(loadingStart, -1, 'compact detail header should branch on detailHeaderLoading')

  const headerSource = detailHeaderSource
  const headerLoadingSource = detailHeaderSource.slice(loadingStart)
  assert.ok(headerSource.includes('score={selectedSystem.score}'))
  assert.ok(
    headerLoadingSource.includes('<DetailLoadingStatus className="sho-detail-compact-loading" />'),
  )
})

test('detail body loading state uses a structured skeleton panel', () => {
  assert.ok(detailLoadingSource.includes('export const DetailLoadingPanel = () => ('))
  assert.ok(detailLoadingSource.includes('className="sho-panel sho-detail-loading"'))
  assert.ok(detailLoadingSource.includes('className="sho-detail-loading-metrics"'))
  assert.match(viewSource, /detailHeaderLoading\s*\?\s*(?:\(\s*)?<DetailLoadingPanel\s*\/>/)
  assert.ok(detailLoadingSource.includes('正在加载系统详情'))
})

test('detail page renders separate score and service distribution trend charts before health metrics', () => {
  const trendPanelStart = viewSource.indexOf('<DetailTrendSection')
  const metricsStart = viewSource.indexOf(
    '<section className="sho-panel sho-system-metrics-panel">',
  )
  const servicePanelStart = viewSource.indexOf('<ServiceHealthTable')

  assert.notEqual(trendPanelStart, -1, 'detail trend panel should exist')
  assert.notEqual(metricsStart, -1, 'health metric panel should exist')
  assert.notEqual(servicePanelStart, -1, 'service table component should exist')
  assert.ok(trendPanelStart < metricsStart, 'trend panel should render before metric panels')
  assert.ok(metricsStart < servicePanelStart, 'metric panel should render before service table')
  assert.ok(viewSource.includes("buildDetailTrendRows(resp?.points, resp?.granularity || '')"))
  assert.ok(viewSource.includes('groupId: selectedSystemId'))
  assert.ok(
    detailTrendSectionSource.includes('<section className="sho-panel sho-detail-trend-panel">'),
  )
  assert.ok(detailTrendSectionSource.includes('<DetailHealthScoreTrendChart'))
  assert.ok(detailTrendSectionSource.includes('<DetailServiceStatusTrendChart'))
  assert.equal(viewSource.includes('<DetailTrendChart'), false)
  assert.ok(detailTrendSectionSource.includes('健康度分数趋势'))
  assert.ok(detailTrendSectionSource.includes('服务状态分布趋势'))
  const serviceTrendLegendSource = detailTrendSectionSource.slice(
    detailTrendSectionSource.indexOf('const detailServiceTrendLegend = ['),
    detailTrendSectionSource.indexOf('const formatDetailHealthScoreTrendTooltip ='),
  )
  assert.ok(serviceTrendLegendSource.includes("label: '异常'"))
  assert.ok(serviceTrendLegendSource.includes("label: '关注'"))
  assert.ok(serviceTrendLegendSource.includes("label: '健康'"))
  assert.ok(serviceTrendLegendSource.includes("label: '无数据'"))
  assert.equal(serviceTrendLegendSource.includes('服务平均占比'), false)
  assert.ok(
    detailTrendSectionSource.includes(
      'const total = formatServiceTrendCount(detailServiceTrendTotalCount(row))',
    ),
  )
  assert.ok(detailTrendSectionSource.includes('总服务数量：${total} 个'))
  assert.ok(trendCommonSource.includes("abnormal: '#E24B4A'"))
  assert.ok(trendCommonSource.includes("attention: '#EF9F27'"))
  assert.ok(trendCommonSource.includes("healthy: '#97C459'"))
  assert.ok(trendCommonSource.includes("noData: '#98A2B3'"))
  assert.ok(detailTrendSectionSource.includes('className="sho-detail-trend-grid"'))

  const trendHeadSource = detailTrendSectionSource.slice(
    detailTrendSectionSource.indexOf('<section className="sho-panel sho-detail-trend-panel">'),
    detailTrendSectionSource.indexOf('<div className="sho-detail-trend-grid">'),
  )
  assert.ok(trendHeadSource.includes('className="sho-detail-trend-range-select"'))
  assert.ok(trendHeadSource.includes('options={trendRangeOptions.map((option) => ({'))
  assert.ok(trendHeadSource.includes('value={trendRange}'))
  assert.ok(trendHeadSource.includes('onChange={onRangeChange}'))
  assert.ok(viewSource.includes('onRangeChange={handleTrendRangeChange}'))
  assert.equal(trendHeadSource.includes('<span>{systemTrendRangeLabel}</span>'), false)
})

test('detail service status trend tooltip renders colored status dots', () => {
  const formatterSource = detailTrendSectionSource.slice(
    detailTrendSectionSource.indexOf('const formatDetailServiceStatusTrendTooltip ='),
    detailTrendSectionSource.indexOf('const buildDetailHealthScoreTrendOption ='),
  )

  assert.ok(formatterSource.includes('formatDetailServiceStatusTooltipLine(item, row)'))
  assert.ok(formatterSource.includes('formatDetailServiceTotalTooltipLine(row)'))
  assert.ok(formatterSource.includes('总服务数量'))
  assert.equal(formatterSource.includes('（平均 ${count} 个）'), false)
  assert.ok(formatterSource.includes("row.serviceDistributionMode === 'average'"))
  assert.ok(formatterSource.includes('统计口径：时段平均占比'))
  assert.ok(formatterSource.includes('${item.label}：${ratio}</span>'))
  assert.ok(
    formatterSource.includes(
      'const count = formatServiceTrendCount(row.serviceCounts?.[item.key])',
    ),
  )
  assert.ok(formatterSource.includes('${item.label}：${count} 个（${ratio}）'))
  assert.ok(formatterSource.includes('sho-detail-service-tooltip-dot'))
  assert.ok(formatterSource.includes('background:${item.color}'))
  assert.ok(formatterSource.includes('detailServiceTrendLegend.map((item) =>'))
})

test('detail trend charts distinguish missing snapshot buckets from service no-data', () => {
  const tooltipSource = detailTrendSectionSource.slice(
    detailTrendSectionSource.indexOf('const formatDetailServiceStatusTrendTooltip ='),
    detailTrendSectionSource.indexOf('const formatDetailServiceStatusTooltipLine ='),
  )
  const optionSource = detailTrendSectionSource.slice(
    detailTrendSectionSource.indexOf('const buildDetailServiceStatusTrendOption ='),
    detailTrendSectionSource.indexOf('const DetailTrendState ='),
  )

  assert.ok(tooltipSource.includes('无数据'))
  assert.ok(tooltipSource.includes('row.missingSnapshot'))
  assert.ok(optionSource.includes('row.missingSnapshot ? null'))
})

test('detail health score trend keeps missing buckets as gaps', () => {
  const optionSource = detailTrendSectionSource.slice(
    detailTrendSectionSource.indexOf('const buildDetailHealthScoreTrendOption ='),
    detailTrendSectionSource.indexOf('const buildDetailServiceStatusTrendOption ='),
  )

  assert.ok(optionSource.includes("data: rows.map((row) => trendDataPointValue(row, 'score'))"))
  assert.ok(optionSource.includes('connectNulls: false'))
  assert.ok(
    optionSource.includes(
      "graphic: rowsHaveTrendData(rows, 'score') ? undefined : noDataTrendGraphic()",
    ),
  )
})

test('detail health score trend chart renders score judgement baselines', () => {
  const optionStart = detailTrendSectionSource.indexOf('const buildDetailHealthScoreTrendOption =')
  const serviceOptionStart = detailTrendSectionSource.indexOf(
    'const buildDetailServiceStatusTrendOption =',
  )
  const chartStart = detailTrendSectionSource.indexOf('const DetailHealthScoreTrendChart =')
  const serviceChartStart = detailTrendSectionSource.indexOf(
    'const DetailServiceStatusTrendChart =',
  )

  assert.notEqual(optionStart, -1, 'detail health score trend option should exist')
  assert.notEqual(serviceOptionStart, -1, 'service status trend option should exist')
  assert.notEqual(chartStart, -1, 'detail health score trend chart should exist')
  assert.notEqual(serviceChartStart, -1, 'service status trend chart should exist')

  const optionSource = detailTrendSectionSource.slice(optionStart, serviceOptionStart)
  const chartSource = detailTrendSectionSource.slice(chartStart, serviceChartStart)

  assert.ok(detailTrendSectionSource.includes('buildHealthScoreBaselineItems'))
  assert.ok(optionSource.includes('buildHealthScoreBaselineItems(scoreThresholds)'))
  assert.ok(optionSource.includes('markLine'))
  assert.ok(optionSource.includes("kind: 'baseline'"))
  assert.ok(optionSource.includes('item.value'))
  assert.ok(chartSource.includes('scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS'))
  assert.ok(
    chartSource.includes('buildDetailHealthScoreTrendOption(rows, granularity, scoreThresholds)'),
  )
  assert.equal(chartSource.includes('<DetailTrendLegend'), false)
  assert.equal(detailTrendSectionSource.includes('const detailHealthScoreTrendLegend ='), false)
  assert.ok(viewSource.includes('scoreThresholds={scoreThresholds}'))
})

test('health page uses active alert wording and health metric section title', () => {
  assert.equal(viewSource.includes('异常请求'), false)
  assert.ok(viewSource.includes("['badRequests', '活跃告警']"))
  assert.ok(viewSource.includes('<div className="sho-panel-title">健康指标</div>'))
  assert.ok(serviceHealthTableSource.includes('<div className="sho-panel-title">服务健康度</div>'))
  assert.ok(pendingIssueQueueSource.includes('<th>当前态</th>'))
  assert.equal(pendingIssueQueueSource.includes('<th>当前值</th>'), false)
})

test('service health table exposes AI assistant actions by health status', () => {
  assert.equal(serviceAiAssistantActionLabel('red'), '异常分析')
  assert.equal(serviceAiAssistantActionLabel('yellow'), '关注调优')
  assert.equal(serviceAiAssistantActionLabel('green'), '持续保持')
  assert.equal(serviceAiAssistantActionLabel('gray'), '无数据')
  assert.deepEqual(
    serviceAiAssistantActionFor({
      status: 'yellow',
      metrics: { resource: 'green', latency: 'red' },
    }),
    { label: '异常分析', actionable: true, status: 'red' },
  )
  assert.deepEqual(
    serviceAiAssistantActionFor({
      status: 'green',
      metrics: { resource: 'green', latency: 'yellow' },
    }),
    { label: '关注调优', actionable: true, status: 'yellow' },
  )
  assert.deepEqual(
    serviceAiAssistantActionFor({
      status: 'green',
      metrics: { resource: 'green', latency: 'green' },
    }),
    { label: '持续保持', actionable: false, status: 'green' },
  )
  assert.deepEqual(
    serviceAiAssistantActionFor({
      status: 'green',
      metrics: { resource: 'green', latency: 'gray' },
    }),
    { label: '持续保持', actionable: false, status: 'green' },
  )
  assert.deepEqual(
    serviceAiAssistantActionFor({
      status: 'gray',
      metrics: { resource: 'gray', latency: 'gray' },
    }),
    { label: '无数据', actionable: false, status: 'gray' },
  )

  assert.ok(serviceHealthTableSource.includes('<th>AI助手</th>'))
  assert.ok(serviceHealthTableSource.includes('serviceAiAssistantActionFor(service)'))
  assert.ok(serviceHealthTableSource.includes('onOpenServiceAssistant(service)'))
  assert.ok(viewSource.includes('<ServiceHealthAssistantModal'))
  assert.ok(viewSource.includes('onOpenServiceAssistant={openServiceAssistant}'))
  assert.ok(viewSource.includes('if (!serviceAiAssistantActionFor(service).actionable) return'))
})

test('service assistant builds red and yellow metric analysis issues only', () => {
  const issues = buildServiceHealthAssistantIssues(
    {
      name: 'order-api',
      score: 71,
      status: 'yellow',
      metrics: {
        resource: 'green',
        errorRate: 'red',
        latency: 'yellow',
        errorLog: 'gray',
      },
      metricDetails: {
        errorRate: { value: '5.7%', note: '错误率高于阈值' },
        latency: { value: '850ms', note: 'P99 延时升高' },
      },
    },
    {
      groupId: 1001,
      name: '订单系统',
      score: 62,
    },
    [
      ['resource', '资源水位'],
      ['errorRate', '错误率'],
      ['latency', '响应延时'],
      ['errorLog', '错误日志'],
    ],
  )

  assert.deepEqual(
    issues.map((issue) => [issue.metricKey, issue.metricLabel, issue.status, issue.value]),
    [
      ['errorRate', '错误率', 'red', '5.7%'],
      ['latency', '响应延时', 'yellow', '850ms'],
    ],
  )
  assert.equal(issues[0].groupId, 1001)
  assert.equal(issues[0].systemName, '订单系统')
  assert.equal(issues[0].serviceName, 'order-api')
  assert.equal(issues[0].serviceScore, 71)
  assert.equal(issues[0].systemScore, 62)
})

test('service assistant modal shows health lamps and routes each metric to Agent', () => {
  assert.ok(serviceHealthAssistantModalSource.includes('HealthRing'))
  assert.ok(serviceHealthAssistantModalSource.includes('metricLabels.map(([key, label]) => ('))
  assert.ok(serviceHealthAssistantModalSource.includes('<Lamp status={service?.metrics?.[key] ||'))
  assert.ok(serviceHealthAssistantModalSource.includes('ServiceAssistantIssue'))
  assert.ok(serviceHealthAssistantModalSource.includes('issues.length ? ('))
  assert.ok(serviceHealthAssistantModalSource.includes('<Tabs'))
  assert.equal(serviceHealthAssistantModalSource.includes('issues.length > 1 ? ('), false)
  assert.ok(serviceHealthAssistantModalSource.includes('onAnalyzeIssue(issue)'))
  assert.ok(serviceHealthAssistantModalSource.includes('在 Agent 中分析'))
  assert.equal(serviceHealthAssistantModalSource.includes('PendingIssueAnalysis'), false)
  assert.equal(serviceHealthAssistantModalSource.includes('AI 分析中'), false)
})

test('service assistant modal title follows the service health status instead of active metric status', () => {
  assert.ok(
    serviceHealthAssistantModalSource.includes(
      "import { serviceAiAssistantActionFor } from '../detailViewState.mjs'",
    ),
  )
  assert.ok(serviceHealthAssistantModalSource.includes('serviceAssistantTitle(service)'))
  assert.ok(
    serviceHealthAssistantModalSource.includes('serviceAiAssistantActionFor(service).label'),
  )
  assert.equal(
    serviceHealthAssistantModalSource.includes(
      'serviceAiAssistantActionLabel(activeIssue?.status || service?.status)',
    ),
    false,
  )
  assert.equal(serviceAiAssistantActionLabel('red'), '异常分析')
  assert.equal(serviceAiAssistantActionLabel('yellow'), '关注调优')
  assert.equal(serviceAiAssistantActionLabel('green'), '持续保持')
  assert.ok(
    serviceHealthAssistantModalSource.includes('serviceAiAssistantActionFor(service).label'),
  )
  assert.equal(serviceHealthAssistantModalSource.includes("yellow: '调优分析'"), false)
  assert.equal(serviceHealthAssistantModalSource.includes('AI助手'), false)
  assert.ok(serviceHealthAssistantModalSource.includes('width={1040}'))
})

test('system and service health filters use compact single-select options', () => {
  const systemOptionsStart = viewSource.indexOf('const systemHealthFilterOptions = [')
  const serviceOptionsStart = viewSource.indexOf('const serviceHealthFilterOptions = [')
  assert.ok(serviceHealthTableSource.includes('<div className="sho-panel-title">服务健康度</div>'))
  assert.ok(serviceHealthTableSource.includes('<HealthTable className="sho-service-table">'))
  assert.notEqual(systemOptionsStart, -1, 'system filter options should exist')
  assert.notEqual(serviceOptionsStart, -1, 'service filter options should exist')

  const servicePanelSource = serviceHealthTableSource
  const systemOptionsSource = viewSource.slice(systemOptionsStart, serviceOptionsStart)
  const serviceOptionsSource = viewSource.slice(
    serviceOptionsStart,
    viewSource.indexOf('const summaryCardIcons'),
  )
  const oldFilterPrefix = '只看'
  assert.equal(viewSource.includes('<span className="sho-muted">异常和关注系统</span>'), false)
  assert.equal(viewSource.includes(`${oldFilterPrefix}异常和关注系统`), false)
  assert.equal(
    servicePanelSource.includes('<span className="sho-muted">异常和关注服务</span>'),
    false,
  )
  assert.equal(servicePanelSource.includes(`${oldFilterPrefix}异常和关注服务`), false)
  assert.ok(
    viewSource.includes(
      'const [systemHealthFilter, setSystemHealthFilter] = useState(DEFAULT_HEALTH_FILTER_VALUE)',
    ),
  )
  assert.ok(
    viewSource.includes(
      'const [serviceHealthFilter, setServiceHealthFilter] = useState(DEFAULT_HEALTH_FILTER_VALUE)',
    ),
  )
  assert.equal(viewSource.includes('mode="multiple"'), false)
  assert.equal(viewSource.includes('maxTagCount="responsive"'), false)
  assert.ok(overviewSystemTableSource.includes('options={systemHealthFilterOptions}'))
  assert.ok(servicePanelSource.includes('options={serviceHealthFilterOptions}'))
  assert.ok(servicePanelSource.includes('value={serviceHealthFilter}'))
  assert.ok(servicePanelSource.includes('onChange={onServiceHealthFilterChange}'))
  assert.ok(viewSource.includes('onServiceHealthFilterChange={setServiceHealthFilter}'))
  assert.ok(viewSource.includes('setServiceHealthFilter(DEFAULT_HEALTH_FILTER_VALUE)'))
  assert.ok(servicePanelSource.includes('<Select'))
  assert.equal(servicePanelSource.includes('serviceFilter'), false)
  assert.equal(
    selectBlockWithPlaceholder(overviewSystemTableSource, '筛选系统').includes('allowClear'),
    false,
    'system health filter should not be clearable',
  )
  assert.equal(
    selectBlockWithPlaceholder(serviceHealthTableSource, '筛选服务').includes('allowClear'),
    false,
    'service health filter should not be clearable',
  )
  assert.ok(
    systemOptionsSource.indexOf("label: '异常系统'") <
      systemOptionsSource.indexOf("label: '关注系统'"),
  )
  assert.ok(
    systemOptionsSource.indexOf("label: '关注系统'") <
      systemOptionsSource.indexOf("label: '健康系统'"),
  )
  assert.ok(
    systemOptionsSource.indexOf("label: '健康系统'") <
      systemOptionsSource.indexOf("label: '无数据系统'"),
  )
  assert.ok(
    systemOptionsSource.indexOf("label: '无数据系统'") <
      systemOptionsSource.indexOf("label: '全部'"),
  )
  assert.ok(
    serviceOptionsSource.indexOf("label: '异常服务'") <
      serviceOptionsSource.indexOf("label: '关注服务'"),
  )
  assert.ok(
    serviceOptionsSource.indexOf("label: '关注服务'") <
      serviceOptionsSource.indexOf("label: '健康服务'"),
  )
  assert.ok(
    serviceOptionsSource.indexOf("label: '健康服务'") <
      serviceOptionsSource.indexOf("label: '无数据服务'"),
  )
  assert.ok(
    serviceOptionsSource.indexOf("label: '无数据服务'") <
      serviceOptionsSource.indexOf("label: '全部'"),
  )
  assert.equal(DEFAULT_HEALTH_FILTER_VALUE, DEFAULT_HEALTH_FILTER_VALUES.abnormal)
})

test('health score sorting keeps computed zero first and no-data rows last', () => {
  const rows = [
    { name: 'healthy', score: 95, status: 'green' },
    { name: 'no-data-zero', score: 0, status: 'gray' },
    { name: 'missing-score' },
    { name: 'warning-string', score: '70', status: 'yellow' },
    { name: 'computed-zero', score: 0, status: 'red' },
    { name: 'critical', score: 20, status: 'red' },
    { name: 'same-score-a', score: 70, status: 'yellow' },
    { name: 'same-score-b', score: 70, status: 'yellow' },
  ]

  const sortedNames = [...rows].sort(compareHealthScoreAscending).map((row) => row.name)

  assert.deepEqual(sortedNames, [
    'computed-zero',
    'critical',
    'warning-string',
    'same-score-a',
    'same-score-b',
    'healthy',
    'no-data-zero',
    'missing-score',
  ])
})

test('system and service health tables sort lower health scores first', () => {
  const overviewStart = viewSource.indexOf('const overviewSystems = useMemo(')
  const serviceStart = viewSource.indexOf('const filteredServices = serviceRows')
  const openDetailStart = viewSource.indexOf('const openDetail = (system')

  assert.notEqual(overviewStart, -1, 'overview systems memo should exist')
  assert.notEqual(serviceStart, -1, 'filtered services state should exist')
  assert.notEqual(openDetailStart, -1, 'open detail handler should exist')

  const overviewSource = viewSource.slice(overviewStart, serviceStart)
  const serviceSource = viewSource.slice(serviceStart, openDetailStart)

  assert.ok(viewSource.includes("sort: 'score_asc'"))
  assert.ok(viewSource.includes('getSystemHealthSystemsApi'))
  assert.ok(viewSource.includes('getSystemHealthServicesApi'))
  assert.equal(serviceSource.includes('serviceSeverity'), false)
})

test('service abnormal attention filter follows health score thresholds', async () => {
  const { serviceMatchesAbnormalAttentionFilter } = await import('./detailViewState.mjs')
  assert.equal(typeof serviceMatchesAbnormalAttentionFilter, 'function')

  const scoreThresholds = {
    severeAbnormalScoreThreshold: 60,
    generalAbnormalScoreThreshold: 80,
  }

  assert.equal(serviceMatchesAbnormalAttentionFilter({ score: 59 }, true, scoreThresholds), true)
  assert.equal(serviceMatchesAbnormalAttentionFilter({ score: 79 }, true, scoreThresholds), true)
  assert.equal(serviceMatchesAbnormalAttentionFilter({ score: 80 }, true, scoreThresholds), false)
  assert.equal(serviceMatchesAbnormalAttentionFilter({ score: 95 }, false, scoreThresholds), true)
})

test('health single-select filter distinguishes abnormal and attention rows', () => {
  const scoreThresholds = {
    severeAbnormalScoreThreshold: 60,
    generalAbnormalScoreThreshold: 80,
  }

  assert.equal(
    targetMatchesHealthFilter(
      { score: 59 },
      DEFAULT_HEALTH_FILTER_VALUES.abnormal,
      scoreThresholds,
    ),
    true,
  )
  assert.equal(
    targetMatchesHealthFilter(
      { score: 79 },
      DEFAULT_HEALTH_FILTER_VALUES.abnormal,
      scoreThresholds,
    ),
    false,
  )
  assert.equal(
    targetMatchesHealthFilter(
      { score: 79 },
      DEFAULT_HEALTH_FILTER_VALUES.attention,
      scoreThresholds,
    ),
    true,
  )
  assert.equal(
    targetMatchesHealthFilter(
      { score: 79 },
      DEFAULT_HEALTH_FILTER_VALUES.attention,
      scoreThresholds,
    ),
    true,
  )
  assert.equal(
    targetMatchesHealthFilter({ score: 95 }, DEFAULT_HEALTH_FILTER_VALUES.healthy, scoreThresholds),
    true,
  )
  assert.equal(
    targetMatchesHealthFilter({ score: 59 }, DEFAULT_HEALTH_FILTER_VALUES.healthy, scoreThresholds),
    false,
  )
  assert.equal(
    targetMatchesHealthFilter({ score: 95 }, DEFAULT_HEALTH_FILTER_VALUES.all, scoreThresholds),
    true,
  )
  assert.equal(
    targetMatchesHealthFilter(
      { score: 0, status: 'gray' },
      DEFAULT_HEALTH_FILTER_VALUES.noData,
      scoreThresholds,
    ),
    true,
  )
  assert.equal(
    targetMatchesHealthFilter(
      { score: 0, status: 'gray' },
      DEFAULT_HEALTH_FILTER_VALUES.abnormal,
      scoreThresholds,
    ),
    false,
  )
  assert.equal(
    targetMatchesHealthFilter(
      { score: 0, status: 'red' },
      DEFAULT_HEALTH_FILTER_VALUES.abnormal,
      scoreThresholds,
    ),
    true,
  )
  assert.equal(targetMatchesHealthFilter({ score: 95 }, [], scoreThresholds), true)
})

test('health score rings receive status for no-data display', () => {
  assert.ok(overviewSystemTableSource.includes('status={system.status}'))
  assert.ok(detailHeaderSource.includes('status={selectedSystem.status}'))
  assert.ok(serviceHealthTableSource.includes('status={service.status}'))
  assert.ok(healthStatusSource.includes("const hasHealthData = status !== 'gray'"))
  assert.ok(
    healthStatusSource.includes("const displayScore = hasHealthData ? safeScore : '无数据'"),
  )
  assert.ok(healthStatusSource.includes('{displayScore}'))
})

test('metric cards do not render gray helper descriptions under values', () => {
  assert.equal(viewSource.includes('<span>{visual.helper}</span>'), false)
})

test('metric cards keep evidence in hover tooltip instead of inline badges', () => {
  assert.equal(metricVisualTileSource.includes('<MetricBadges badges={visual.badges} />'), false)
  assert.equal(metricVisualTileSource.includes('const MetricBadges ='), false)
  assert.ok(metricVisualTileSource.includes('MetricDecisionTooltip'))
  assert.ok(metricVisualTileSource.includes('overlayClassName="sho-metric-decision-tooltip"'))
})

test('metric decision tooltip keeps health explanation compact', () => {
  assert.ok(metricVisualTileSource.includes('metricTooltipConclusion'))
  assert.ok(metricVisualTileSource.includes('splitMetricTooltipBadges'))
  assert.ok(metricVisualTileSource.includes('sho-metric-decision-list'))
  assert.ok(metricVisualTileSource.includes("return '水位变化'"))
  assert.ok(
    metricVisualTileSource.includes("const evidenceLimit = metricKey === 'resourceShift' ? 3 : 2"),
  )
  assert.equal(metricVisualTileSource.includes('metricTooltipReason'), false)
  assert.equal(metricVisualTileSource.includes('sho-metric-decision-section'), false)
  assert.equal(metricVisualTileSource.includes('整体情况'), false)
  assert.equal(metricVisualTileSource.includes('为什么这样判定'), false)
  assert.equal(metricVisualTileSource.includes('仅用于定位，不参与整体判定'), false)
  assert.equal(metricVisualTileSource.includes('系统主值'), false)
  assert.equal(metricVisualTileSource.includes('判定口径'), false)
  assert.equal(metricVisualTileSource.includes('补充依据'), false)
})

test('active alert metric card does not render a distribution chart', () => {
  assert.equal(viewSource.includes('AlertDistributionBar'), false)
  assert.equal(viewSource.includes('sho-alert-distribution'), false)
  assert.equal(viewSource.includes('visual.distribution'), false)
})
