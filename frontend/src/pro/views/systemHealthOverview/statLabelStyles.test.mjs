import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

const sourceFor = (path) => readFileSync(new URL(path, import.meta.url), 'utf8')

const css = sourceFor('./index.css')
const viewSource = sourceFor('./index.jsx')
const thresholdSettingsSource = sourceFor('./components/ThresholdSettingsContent.jsx')
const detailHeaderSource = sourceFor('./components/DetailHeader.jsx')
const overviewSummarySource = sourceFor('./components/OverviewSummary.jsx')
const trendCommonSource = sourceFor('./components/TrendCommon.jsx')
const pendingIssueQueueSource = sourceFor('./components/PendingIssueQueue.jsx')
const overviewInitialLoadingSource = sourceFor('./components/OverviewInitialLoading.jsx')

const sourceBetween = (start, end, targetSource = viewSource) => {
  const startIndex = targetSource.indexOf(start)
  assert.notEqual(startIndex, -1, `${start} should exist`)
  const endIndex = targetSource.indexOf(end, startIndex)
  assert.notEqual(endIndex, -1, `${end} should exist after ${start}`)
  return targetSource.slice(startIndex, endIndex)
}

const declarationsFor = (selector) => {
  const match = css.match(
    new RegExp(`${selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*\\{([^}]+)\\}`),
  )
  assert.ok(match, `${selector} rule should exist`)
  return match[1]
    .split(';')
    .map((line) => line.trim())
    .filter(Boolean)
}

const declarationBlocksFor = (selector) => {
  const pattern = new RegExp(
    `${selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*\\{([^}]+)\\}`,
    'g',
  )
  const blocks = [...css.matchAll(pattern)].map((match) =>
    match[1]
      .split(';')
      .map((line) => line.trim())
      .filter(Boolean),
  )
  assert.ok(blocks.length > 0, `${selector} rule should exist`)
  return blocks
}

test('summary cards center the title and value as a compact group', () => {
  const declarations = declarationsFor('.sho-summary-card')

  assert.ok(declarations.includes('grid-template-columns: auto auto'))
  assert.ok(declarations.includes('align-content: center'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('justify-content: center'))
  assert.ok(declarations.includes('column-gap: 16px'))
})

test('detail page removes outer padding so the header spans the row', () => {
  const declarations = declarationsFor('.system-health-overview-page-detail')

  assert.ok(declarations.includes('padding: 0'))
})

test('summary card title is enlarged after removing descriptions', () => {
  const declarations = declarationsFor('.sho-summary-card-head > span')

  assert.ok(declarations.includes('display: inline-flex'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('gap: 6px'))
  assert.ok(declarations.includes('min-width: 0'))
  assert.ok(declarations.includes('color: var(--ant-color-text)'))
  assert.ok(declarations.includes('font-size: 18px'))
  assert.ok(declarations.includes('font-weight: 700'))
})

test('summary cards with trend bodies vertically center title and value row', () => {
  const declarations = declarationsFor('.sho-summary-card-with-body')

  assert.ok(declarations.includes('grid-template-columns: minmax(0, 1fr) auto'))
  assert.ok(declarations.includes('align-content: start'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('justify-content: stretch'))
})

test('summary card title icons use severity colors without shrinking', () => {
  const iconDeclarations = declarationsFor('.sho-summary-title-icon')
  const criticalDeclarations = declarationsFor('.sho-summary-title-icon-critical')
  const warningDeclarations = declarationsFor('.sho-summary-title-icon-warning')

  assert.ok(iconDeclarations.includes('flex: 0 0 auto'))
  assert.ok(iconDeclarations.includes('stroke-width: 2.4'))
  assert.ok(criticalDeclarations.includes('color: #f04438'))
  assert.ok(warningDeclarations.includes('color: #f79009'))
})

test('summary cards no longer render captions or ring visuals', () => {
  assert.equal(css.includes('.sho-summary-ring'), false)
  assert.equal(css.includes('.sho-summary-card small'), false)
})

test('summary ratio value uses neutral text with severity numerator highlights', () => {
  const valueDeclarations = declarationsFor('.sho-summary-value')
  const highlightDeclarations = declarationsFor('.sho-summary-value-highlight')
  const criticalDeclarations = declarationsFor('.sho-summary-value-highlight-critical')
  const warningDeclarations = declarationsFor('.sho-summary-value-highlight-warning')

  assert.ok(valueDeclarations.includes('color: var(--ant-color-text)'))
  assert.ok(valueDeclarations.includes('font-size: 34px'))
  assert.ok(valueDeclarations.includes('font-variant-numeric: tabular-nums'))
  assert.ok(highlightDeclarations.includes('font-weight: inherit'))
  assert.ok(criticalDeclarations.includes('color: #f04438'))
  assert.ok(warningDeclarations.includes('color: #f79009'))
})

test('summary ratio value is not truncated by the generic card value cap', () => {
  const declarations = declarationsFor(
    '.sho-summary-card-ratio .sho-summary-statistic .ant-statistic-content',
  )

  assert.ok(declarations.includes('max-width: none'))
})

test('summary metric cards render their large values through Statistic', () => {
  assert.ok(overviewSummarySource.includes('Statistic'))
  assert.ok(overviewSummarySource.includes('<Statistic'))
  assert.equal(overviewSummarySource.includes('<strong>{renderedValue}</strong>'), false)
})

test('summary titles use direct icons while settings live in toolbar', () => {
  const summaryMetricCardSource = sourceBetween(
    'const SummaryMetricCard = ({',
    'SummaryMetricCard.propTypes',
    overviewSummarySource,
  )

  assert.ok(overviewSummarySource.includes('help={card.help}'))
  assert.match(summaryMetricCardSource, /<Tooltip title=\{help\}>/)
  assert.doesNotMatch(summaryMetricCardSource, /HelpCircle/)
  assert.ok(viewSource.includes('const summaryCardIcons = {'))
  assert.match(viewSource, /abnormalSystems: \(/)
  assert.match(viewSource, /attentionSystems: \(/)
  assert.match(viewSource, /pendingIssues: \(/)
  assert.ok(viewSource.includes('className="sho-toolbar-actions"'))
  assert.ok(viewSource.includes('className="sho-threshold-toolbar-trigger"'))
  assert.ok(viewSource.includes('<ThresholdSettingsContent'))
  assert.equal(viewSource.includes('className="sho-summary-question-trigger"'), false)
  assert.ok(thresholdSettingsSource.includes('系统健康度低于 {severeThreshold} 分，计为异常系统'))
  assert.ok(
    thresholdSettingsSource.includes(
      '系统健康度 {severeThreshold}-{generalThreshold - 1} 分之间，计为关注系统',
    ),
  )
})

test('pending issue section uses item and anomaly wording with the atomic analysis icon', () => {
  assert.ok(pendingIssueQueueSource.includes("import atomicIcon from 'src/pro/assets/atomic.png'"))
  assert.ok(pendingIssueQueueSource.includes("import { HelpCircle } from 'lucide-react'"))
  assert.equal(pendingIssueQueueSource.includes('Atom'), false)
  assert.equal(pendingIssueQueueSource.includes('search.png'), false)
  assert.equal(pendingIssueQueueSource.includes('searchIcon'), false)
  assert.equal(pendingIssueQueueSource.includes('BrainCircuit'), false)
  assert.equal(pendingIssueQueueSource.includes('analysis.png'), false)
  assert.equal(pendingIssueQueueSource.includes('Sparkles'), false)
  assert.ok(pendingIssueQueueSource.includes('待处理事项'))
  assert.equal(pendingIssueQueueSource.includes('待处理问题'), false)
  assert.ok(pendingIssueQueueSource.includes('类异常'))
  assert.ok(pendingIssueQueueSource.includes('<th>异常</th>'))
  assert.ok(pendingIssueQueueSource.includes('className="sho-pending-analysis-button-icon"'))
  assert.ok(pendingIssueQueueSource.includes('src={atomicIcon}'))
  assert.ok(pendingIssueQueueSource.includes(": '异常分析'"))
  assert.equal(pendingIssueQueueSource.includes('智能分析'), false)
  assert.equal(pendingIssueQueueSource.includes('AI 分析'), false)
  assert.ok(viewSource.includes("setPendingIssueError(error?.message || '待处理事项加载失败')"))
  assert.ok(viewSource.includes("error: error?.message || '异常分析失败'"))
  assert.equal(viewSource.includes('智能分析失败'), false)
  assert.ok(overviewInitialLoadingSource.includes('待处理事项'))
  assert.equal(overviewInitialLoadingSource.includes('待处理问题'), false)
})

test('pending issue analysis row only renders when there is visible analysis content', () => {
  assert.ok(pendingIssueQueueSource.includes('const shouldShowAnalysisRow ='))
  assert.ok(pendingIssueQueueSource.includes('Boolean(analysisState.error)'))
  assert.ok(
    pendingIssueQueueSource.includes(
      "Boolean(analysisState.data && analysisState.data.aiStatus !== 'succeeded')",
    ),
  )
  assert.ok(pendingIssueQueueSource.includes('{shouldShowAnalysisRow && ('))
  assert.equal(pendingIssueQueueSource.includes('{analysisNode && ('), false)
})

test('threshold settings can still be opened from the toolbar', () => {
  assert.ok(
    viewSource.includes(
      'const [thresholdSettingsOpen, setThresholdSettingsOpen] = useState(false)',
    ),
  )
  assert.ok(viewSource.includes('open={thresholdSettingsOpen}'))
  assert.ok(viewSource.includes('className="sho-threshold-toolbar-trigger"'))
})

test('threshold toolbar trigger vertically centers icon and label', () => {
  const triggerDeclarations = declarationsFor('.sho-threshold-toolbar-trigger.ant-btn')
  const iconDeclarations = declarationsFor('.sho-threshold-toolbar-trigger .ant-btn-icon')
  const iconSvgDeclarations = declarationsFor('.sho-threshold-toolbar-trigger .ant-btn-icon svg')

  assert.ok(triggerDeclarations.includes('display: inline-flex'))
  assert.ok(triggerDeclarations.includes('align-items: center'))
  assert.ok(triggerDeclarations.includes('justify-content: center'))
  assert.ok(triggerDeclarations.includes('line-height: 1'))
  assert.ok(iconDeclarations.includes('display: inline-flex'))
  assert.ok(iconDeclarations.includes('align-items: center'))
  assert.ok(iconDeclarations.includes('line-height: 0'))
  assert.ok(iconSvgDeclarations.includes('display: block'))
})

test('trend range selector lives in toolbar before threshold settings', () => {
  const toolbarActionsSource = sourceBetween(
    '<div className="sho-toolbar-actions">',
    '<Button\n                aria-label="健康异常判定设置"',
  )

  assert.ok(toolbarActionsSource.includes('趋势图历史范围'))
  assert.ok(toolbarActionsSource.includes('className="sho-trend-range-select"'))
  assert.ok(toolbarActionsSource.includes('value={trendRange}'))
  assert.ok(toolbarActionsSource.includes('onChange={handleTrendRangeChange}'))
  assert.ok(viewSource.includes("judgement: 'snapshot'"))
  assert.equal(viewSource.includes("judgement: 'current'"), false)
  assert.equal(viewSource.includes('判定口径：'), false)
  assert.equal(viewSource.includes('className="sho-history-range"'), false)
})

test('summary anomaly cards embed their trend charts instead of using a standalone history panel', () => {
  const summarySource = sourceBetween(
    '<div className="sho-summary">',
    'OverviewSummary.propTypes',
    overviewSummarySource,
  )

  assert.ok(summarySource.includes('summaryTrendGroupByKey[card.key]'))
  assert.ok(summarySource.includes('<SummaryTrendGroup'))
  assert.equal(summarySource.includes('renderHistoryTrends()'), false)
  assert.equal(viewSource.includes('const renderHistoryTrends = () => {'), false)
  assert.ok(overviewSummarySource.includes('className="sho-summary-trends"'))
  assert.ok(overviewSummarySource.includes('className="sho-summary-trend-item"'))
  assert.equal(viewSource.includes('className="sho-history-panel"'), false)
  assert.equal(css.includes('.sho-history-panel'), false)
})

test('trend loading failures render as retryable state blocks instead of plain red text', () => {
  const failureStateSource = sourceBetween(
    'export const TrendFailureState = ({',
    'TrendFailureState.propTypes',
    trendCommonSource,
  )

  assert.ok(failureStateSource.includes('CircleAlert'))
  assert.ok(failureStateSource.includes('className="sho-trend-failure-title"'))
  assert.ok(failureStateSource.includes('className="sho-trend-failure-description"'))
  assert.ok(failureStateSource.includes('onClick={onRetry}'))
  assert.ok(viewSource.includes('onTrendRetry={handleOverviewTrendRetry}'))
  assert.ok(overviewSummarySource.includes('description="无法读取历史快照"'))
  assert.equal(
    viewSource.includes(
      '<div className="sho-summary-trend-state sho-summary-trend-error">趋势加载失败</div>',
    ),
    false,
  )
  assert.equal(
    viewSource.includes(
      '<div className="sho-detail-trend-state sho-detail-trend-error">趋势加载失败</div>',
    ),
    false,
  )
})

test('trend failure visual treatment uses a compact full-border surface', () => {
  const shellDeclarations = declarationsFor('.sho-trend-failure')
  const innerDeclarations = declarationsFor('.sho-trend-failure-inner')
  const iconDeclarations = declarationsFor('.sho-trend-failure-icon')

  assert.ok(shellDeclarations.includes('display: flex'))
  assert.ok(shellDeclarations.includes('align-items: center'))
  assert.ok(shellDeclarations.includes('justify-content: center'))
  assert.ok(innerDeclarations.includes('border: 1px solid var(--ant-color-error-border)'))
  assert.equal(innerDeclarations.some((declaration) => declaration.startsWith('border-left:')), false)
  assert.ok(iconDeclarations.includes('color: var(--ant-color-error)'))
  assert.ok(iconDeclarations.includes('background: var(--ant-color-error-bg)'))
})

test('detail row cards suppress stray bottom borders', () => {
  const headerDeclarations = declarationsFor('.sho-detail-compact-inner')
  const trendPanelDeclarations = declarationBlocksFor('.sho-detail-trend-panel').flat()

  assert.ok(headerDeclarations.includes('border-bottom: 0'))
  assert.ok(trendPanelDeclarations.includes('border-bottom: 0'))
})

test('summary trend charts combine count and ratio without header values or grid lines', () => {
  const trendGroupSource = sourceBetween(
    'const SummaryTrendGroup = ({',
    'SummaryTrendGroup.propTypes',
    overviewSummarySource,
  )
  const trendOptionSource = sourceBetween(
    'const buildSummaryTrendOption = (rows, group, granularity) => {',
    'const SummaryTrendGroup =',
    overviewSummarySource,
  )

  assert.equal(viewSource.includes('const SummaryTrendItem = ({'), false)
  assert.equal(trendGroupSource.includes('group.metrics.map((metric) => ('), false)
  assert.ok(trendGroupSource.includes('buildSummaryCombinedTrendRows(points, group.metrics)'))
  assert.ok(trendGroupSource.includes('buildSummaryTrendOption(rows, group, granularity)'))
  assert.ok(trendGroupSource.includes('${trendRangeLabel}趋势图'))
  assert.equal(trendGroupSource.includes('${group.label}${trendRangeLabel}趋势图'), false)
  assert.ok(viewSource.includes('trendRangeLabel={systemTrendRangeLabel}'))
  assert.equal(css.includes('.sho-summary-trend-head b'), false)
  assert.ok(trendOptionSource.includes('axisLine: { show: false }'))
  assert.ok(trendOptionSource.includes('splitLine: { show: false }'))
  assert.ok(trendOptionSource.includes('data: trendAxisLabels(rows, granularity)'))
  assert.ok(trendOptionSource.includes('minInterval: 1'))
  assert.ok(trendOptionSource.includes('yAxisIndex: isTrendPercentMetric(metric) ? 1 : 0'))
  assert.ok(trendOptionSource.includes('showSymbol: false'))
  assert.equal(trendOptionSource.includes('showSymbol: rows.length'), false)
  assert.ok(trendOptionSource.includes('formatSummaryTrendTooltipMetricLabel(metric)'))
  assert.ok(trendOptionSource.includes('${formatSummaryTrendTooltipMetricLabel(metric)}：${formatTrendNumber(value, metric)}'))
  assert.equal(trendOptionSource.includes('current：'), false)
  assert.equal(trendOptionSource.includes('平均值'), false)
  assert.equal(trendOptionSource.includes('期末值'), false)
  assert.equal(trendOptionSource.includes('期末占比'), false)
  assert.equal(trendOptionSource.includes('持续占比'), false)
})

test('summary trend tooltip distinguishes missing snapshot buckets', () => {
  const trendOptionSource = sourceBetween(
    'const buildSummaryTrendOption = (rows, group, granularity) => {',
    'const SummaryTrendGroup =',
    overviewSummarySource,
  )

  assert.ok(trendOptionSource.includes('无数据'))
  assert.ok(trendOptionSource.includes('row.missingSnapshot'))
  assert.ok(trendOptionSource.includes('summaryTrendDataPointValue(row, metric.key)'))
  assert.ok(
    trendOptionSource.includes('graphic: summaryRowsHaveTrendData(rows, metrics) ? undefined : noDataTrendGraphic()'),
  )
})

test('threshold popover separates input rows from status legend rows', () => {
  const popoverDeclarations = declarationsFor('.sho-threshold-popover')
  const inputDeclarations = declarationsFor('.sho-threshold-input-row')
  const noteDeclarations = declarationsFor('.sho-threshold-abnormal-note')
  const legendDeclarations = declarationsFor('.sho-threshold-legend-row')
  const ringDeclarations = declarationsFor('.sho-threshold-status-ring')
  const actionDeclarations = declarationsFor('.sho-threshold-actions')
  const saveDeclarations = declarationsFor('.sho-threshold-save')
  const toolbarTriggerDeclarations = declarationsFor('.sho-threshold-toolbar-trigger')

  assert.ok(thresholdSettingsSource.includes('className="sho-threshold-inputs"'))
  assert.ok(thresholdSettingsSource.includes('className="sho-threshold-abnormal-note"'))
  assert.ok(thresholdSettingsSource.includes('健康异常判定'))
  assert.equal(thresholdSettingsSource.includes('健康度分值异常判定'), false)
  assert.ok(thresholdSettingsSource.includes('aria-label="关注系统健康度阈值"'))
  assert.ok(thresholdSettingsSource.includes('<span>分，关注</span>'))
  assert.equal(thresholdSettingsSource.includes('分，一般异常'), false)
  assert.equal(thresholdSettingsSource.includes('分，关注系统'), false)
  assert.ok(thresholdSettingsSource.includes('系统健康度低于 {severeThreshold} 分，计为异常系统'))
  assert.ok(
    thresholdSettingsSource.includes(
      '系统健康度 {severeThreshold}-{generalThreshold - 1} 分之间，计为关注系统',
    ),
  )
  assert.ok(thresholdSettingsSource.includes('className="sho-threshold-legend"'))
  assert.ok(thresholdSettingsSource.includes('<ThresholdStatusRing status="red" />'))
  assert.ok(thresholdSettingsSource.includes('<ThresholdStatusRing status="yellow" />'))
  assert.ok(thresholdSettingsSource.includes('<ThresholdStatusRing status="green" />'))
  assert.ok(thresholdSettingsSource.includes('severeRange'))
  assert.equal(thresholdSettingsSource.includes('亚健康'), false)
  assert.ok(thresholdSettingsSource.includes('>异常</span>'))
  assert.ok(thresholdSettingsSource.includes('>关注</span>'))
  assert.equal(thresholdSettingsSource.includes('严重异常（异常系统）'), false)
  assert.equal(thresholdSettingsSource.includes('一般异常（关注系统）'), false)
  assert.ok(thresholdSettingsSource.includes('健康'))
  assert.equal(thresholdSettingsSource.includes('icon={<Check'), false)
  assert.ok(popoverDeclarations.includes('width: 360px'))
  assert.ok(inputDeclarations.includes('grid-template-columns: auto 72px 1fr'))
  assert.ok(noteDeclarations.includes('background: var(--ant-color-fill-quaternary)'))
  assert.ok(noteDeclarations.includes('font-weight: 650'))
  assert.ok(legendDeclarations.includes('grid-template-columns: 30px 76px minmax(0, 1fr)'))
  assert.ok(ringDeclarations.some((declaration) => declaration.includes('conic-gradient')))
  assert.ok(actionDeclarations.includes('justify-content: flex-end'))
  assert.ok(saveDeclarations.includes('min-width: 64px'))
  assert.ok(toolbarTriggerDeclarations.includes('height: 30px'))
})

test('summary Statistic value uses compact tabular digits beside the centered title', () => {
  const statisticDeclarations = declarationsFor('.sho-summary-statistic')
  const contentDeclarations = declarationsFor('.sho-summary-statistic .ant-statistic-content')

  assert.ok(statisticDeclarations.includes('grid-column: 2'))
  assert.ok(statisticDeclarations.includes('justify-self: start'))
  assert.ok(statisticDeclarations.includes('min-width: max-content'))
  assert.ok(contentDeclarations.includes('max-width: none'))
  assert.ok(contentDeclarations.includes('overflow: visible'))
  assert.ok(contentDeclarations.includes('font-size: 34px'))
  assert.ok(contentDeclarations.includes('font-variant-numeric: tabular-nums'))
})

test('compact detail header keeps a single-row fixed monitor shape', () => {
  const declarations = declarationsFor('.sho-detail-compact-header')
  const allHeaderDeclarations = declarationBlocksFor('.sho-detail-compact-header').flat()

  assert.ok(declarations.includes('min-height: 88px'))
  assert.ok(declarations.includes('display: flex'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('justify-content: center'))
  assert.ok(declarations.includes('margin-bottom: 12px'))
  assert.equal(
    allHeaderDeclarations.some((declaration) => declaration.startsWith('padding')),
    false,
  )
  assert.equal(
    allHeaderDeclarations.some((declaration) => declaration.startsWith('background')),
    false,
  )
})

test('compact detail monitor aligns groups without changing theme colors', () => {
  const declarations = declarationsFor('.sho-detail-compact-inner')

  assert.ok(declarations.includes('width: 100%'))
  assert.ok(declarations.includes('min-height: 88px'))
  assert.ok(declarations.includes('display: flex'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('justify-content: flex-start'))
  assert.ok(declarations.includes('gap: 28px'))
  assert.ok(declarations.includes('background: var(--ant-color-bg-container)'))
  assert.ok(declarations.includes('border: 1px solid var(--ant-color-border-secondary)'))
  assert.equal(
    declarations.some((declaration) => declaration.startsWith('max-width')),
    false,
  )
})

test('detail page keeps equal spacing between stacked cards', () => {
  const declarations = declarationsFor('.system-health-overview-page-detail .sho-panel')

  assert.ok(declarations.includes('margin-bottom: 12px'))
})

test('compact detail score label style does not override the health ring score', () => {
  const labelDeclarations = declarationsFor('.sho-detail-compact-score-label')

  assert.ok(labelDeclarations.includes('display: inline-flex'))
  assert.ok(labelDeclarations.includes('align-items: center'))
  assert.ok(labelDeclarations.includes('font-size: 16px'))
  assert.ok(labelDeclarations.includes('line-height: 22px'))
  assert.ok(labelDeclarations.includes('font-weight: 700'))
  assert.equal(css.includes('.sho-detail-compact-score span {'), false)
})

test('compact detail service count drops card chrome and reuses summary ratio value styling', () => {
  const countDeclarations = declarationsFor('.sho-detail-compact-service-count')
  const labelDeclarations = declarationsFor('.sho-detail-compact-service-label')
  const detailValueDeclarations = declarationsFor('.sho-detail-compact-service-value')
  const valueDeclarations = declarationsFor('.sho-summary-value')

  assert.equal(
    countDeclarations.some((declaration) => declaration.startsWith('border')),
    false,
  )
  assert.equal(
    countDeclarations.some((declaration) => declaration.startsWith('background')),
    false,
  )
  assert.ok(valueDeclarations.includes('font-size: 34px'))
  assert.ok(valueDeclarations.includes('font-variant-numeric: tabular-nums'))
  assert.ok(labelDeclarations.includes('font-size: 16px'))
  assert.ok(labelDeclarations.includes('line-height: 22px'))
  assert.ok(detailValueDeclarations.includes('font-size: 24px'))
  assert.ok(detailValueDeclarations.includes('line-height: 28px'))
  assert.ok(detailValueDeclarations.includes('font-weight: 700'))
  assert.ok(detailHeaderSource.includes('className="sho-detail-compact-service-value sho-summary-value"'))
})

test('compact detail text uses 16px without changing controls', () => {
  const titleDeclarations = declarationsFor('.sho-detail-compact-title strong')
  const exceptionDeclarations = declarationBlocksFor('.sho-detail-compact-exception-label').flat()
  const backDeclarations = declarationsFor('.sho-detail-compact-back.ant-btn')
  const loadingDeclarations = declarationsFor('.sho-detail-loading-status')
  const tagDeclarations = declarationsFor('.sho-detail-compact-tags .ant-tag')

  assert.ok(titleDeclarations.includes('font-size: 16px'))
  assert.ok(titleDeclarations.includes('line-height: 22px'))
  assert.ok(exceptionDeclarations.includes('font-size: 16px'))
  assert.ok(exceptionDeclarations.includes('line-height: 22px'))
  assert.ok(exceptionDeclarations.includes('font-weight: 700'))
  assert.ok(backDeclarations.includes('font-size: 16px'))
  assert.ok(loadingDeclarations.includes('font-size: 16px'))
  assert.equal(
    tagDeclarations.some((declaration) => declaration.startsWith('font-size')),
    false,
  )
  assert.equal(
    tagDeclarations.some((declaration) => declaration.startsWith('line-height')),
    false,
  )
  assert.ok(tagDeclarations.includes('max-width: 118px'))
})

test('detail no longer uses separate summary cards below the title bar', () => {
  assert.equal(css.includes('.sho-detail-summary'), false)
  assert.equal(css.includes('.sho-score-card'), false)
})

test('mini health ring fits the compact monitor header', () => {
  const declarations = declarationsFor('.sho-health-ring-mini')

  assert.ok(declarations.includes('width: 52px'))
  assert.ok(declarations.includes('height: 52px'))
  assert.ok(declarations.includes('font-size: 16px'))
  assert.ok(declarations.includes('font-variant-numeric: tabular-nums'))
})

test('no-data health ring text is smaller and not bold', () => {
  const declarations = declarationsFor('.sho-health-ring-no-data')
  const miniDeclarations = declarationsFor('.sho-health-ring-mini.sho-health-ring-no-data')

  assert.ok(declarations.includes('font-size: 10px'))
  assert.ok(declarations.includes('font-weight: 400'))
  assert.ok(miniDeclarations.includes('font-size: 11px'))
  assert.ok(miniDeclarations.includes('font-weight: 400'))
})

test('compact detail exception icon uses error red', () => {
  const declarations = declarationsFor('.sho-detail-compact-exception-label svg')

  assert.ok(declarations.includes('width: 18px'))
  assert.ok(declarations.includes('height: 18px'))
  assert.ok(declarations.includes('color: var(--ant-color-error)'))
})

test('health score tooltip only keeps compact calculation copy', () => {
  const popoverDeclarations = declarationsFor('.sho-health-score-popover .ant-tooltip-inner')
  const helpDeclarations = declarationsFor('.sho-health-score-help')

  assert.ok(popoverDeclarations.includes('width: 260px'))
  assert.ok(helpDeclarations.includes('font-size: 12px'))
  assert.ok(helpDeclarations.includes('line-height: 1.55'))
  assert.equal(css.includes('.sho-health-score-help-ring'), false)
  assert.equal(css.includes('.sho-health-score-help-range'), false)
})

test('health score range tooltip shows score bands with ring samples', () => {
  const popoverDeclarations = declarationsFor('.sho-health-score-range-popover .ant-tooltip-inner')
  const titleDeclarations = declarationsFor('.sho-health-score-title')
  const rowDeclarations = declarationsFor('.sho-health-score-range-row')
  const ringDeclarations = declarationsFor('.sho-health-score-range-row .sho-threshold-status-ring')

  assert.ok(popoverDeclarations.includes('width: 210px'))
  assert.ok(titleDeclarations.includes('display: inline-flex'))
  assert.ok(titleDeclarations.includes('gap: 5px'))
  assert.ok(rowDeclarations.includes('grid-template-columns: 26px 70px minmax(0, 1fr)'))
  assert.ok(ringDeclarations.includes('width: 24px'))
  assert.ok(ringDeclarations.includes('height: 24px'))
  assert.ok(ringDeclarations.some((declaration) => declaration.includes('conic-gradient')))
})

test('system metric tiles reserve consistent vertical slots', () => {
  const declarations = declarationsFor('.sho-metric-tile')

  assert.ok(declarations.includes('grid-template-rows: 22px 118px'))
  assert.ok(declarations.includes('min-height: 168px'))
  assert.ok(declarations.includes('align-content: stretch'))
})

test('system metric tiles no longer reserve footer marker space', () => {
  assert.equal(css.includes('sho-metric-foot'), false)
})

test('system metric visuals align value and chart areas', () => {
  const declarations = declarationsFor('.sho-metric-visual')

  assert.ok(declarations.includes('grid-template-rows: 42px 72px'))
  assert.ok(declarations.includes('min-height: 118px'))
})

test('resource waterline metric centers value and waterline horizontally', () => {
  const declarations = declarationsFor('.sho-metric-visual-waterline')
  const valueDeclarations = declarationsFor('.sho-metric-visual-waterline .sho-metric-value-block')
  const waterlineDeclarations = declarationsFor('.sho-waterline')

  assert.ok(declarations.includes('grid-template-columns: minmax(0, 1fr) 55px'))
  assert.ok(declarations.includes('grid-template-rows: 118px'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(valueDeclarations.includes('align-self: center'))
  assert.ok(waterlineDeclarations.includes('width: 55px'))
  assert.ok(waterlineDeclarations.includes('height: 99px'))
})

test('plain number metric values are centered in the visual area', () => {
  const declarations = declarationsFor('.sho-metric-visual-number')
  const valueDeclarations = declarationsFor('.sho-metric-visual-number .sho-metric-value-block')

  assert.ok(declarations.includes('grid-template-rows: 118px'))
  assert.ok(declarations.includes('place-items: center'))
  assert.ok(valueDeclarations.includes('align-content: center'))
  assert.ok(valueDeclarations.includes('justify-items: center'))
  assert.ok(valueDeclarations.includes('text-align: center'))
})

test('system metric values can show compact units without ellipsis', () => {
  const declarations = declarationsFor('.sho-metric-value-block b')

  assert.ok(declarations.includes('display: inline-flex'))
  assert.ok(declarations.includes('white-space: normal'))
  assert.equal(
    declarations.some((declaration) => declaration.startsWith('text-overflow')),
    false,
  )
})

test('system metric value block centers visible values by default', () => {
  const declarations = declarationsFor('.sho-metric-value-block')

  assert.ok(declarations.includes('align-content: center'))
  assert.ok(declarations.includes('justify-items: center'))
  assert.ok(declarations.includes('text-align: center'))
})
