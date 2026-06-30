import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

const source = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')

const between = (start, end) => {
  const startIndex = source.indexOf(start)
  assert.notEqual(startIndex, -1, `${start} should exist`)
  const endIndex = source.indexOf(end, startIndex)
  assert.notEqual(endIndex, -1, `${end} should exist after ${start}`)
  return source.slice(startIndex, endIndex)
}

test('health score legend keeps only primary status labels', () => {
  const statusMetaSource = between('const statusMeta = {', 'const statusOrder')
  const legendSource = between('const Legend = () => (', '</div>\n)')

  assert.match(statusMetaSource, /label: '健康'/)
  assert.match(statusMetaSource, /label: '关注'/)
  assert.match(statusMetaSource, /label: '异常'/)
  assert.match(statusMetaSource, /label: '无数据'/)
  assert.doesNotMatch(statusMetaSource, /description:/)
  assert.doesNotMatch(legendSource, /meta\.description/)
  assert.doesNotMatch(legendSource, /<small>/)
})

test('health score tooltip keeps only calculation copy without score ranges', () => {
  const tooltipSource = between(
    'const HealthScoreCalculationTooltip = () => (',
    'const SystemMetricTrendChart',
  )
  const systemHealthSource = between('<div className="sho-panel-title">', '<Space wrap size={8}>')

  assert.match(
    tooltipSource,
    /由资源水位、资源波动、错误率、响应延时、错误日志、活跃告警、锁等待综合计算/,
  )
  assert.doesNotMatch(tooltipSource, /healthScoreLegendItemsFor/)
  assert.doesNotMatch(tooltipSource, /sho-health-score-help-ring/)
  assert.doesNotMatch(tooltipSource, /sho-health-score-help-range/)
  assert.doesNotMatch(tooltipSource, /sho-legend-item/)
  assert.match(systemHealthSource, /sho-health-score-popover/)
  assert.match(systemHealthSource, /HelpCircle/)
})

test('health score range tooltip documents configured score bands with ring samples', () => {
  const rangeTooltipSource = between(
    'const healthScoreRangesFor = (',
    'HealthScoreRangeTooltip.propTypes',
  )
  const overviewTableSource = between('<th>系统</th>', '{metricLabels.map(([, label]) => (')
  const serviceTableSource = between(
    '<HealthTable className="sho-service-table">',
    '</HealthTable>',
  )

  assert.match(rangeTooltipSource, /healthScoreRangesFor\(scoreThresholds\)\.map/)
  assert.match(rangeTooltipSource, /label: '关注'/)
  assert.match(rangeTooltipSource, /label: '异常'/)
  assert.doesNotMatch(rangeTooltipSource, /（关注系统）/)
  assert.doesNotMatch(rangeTooltipSource, /（异常系统）/)
  assert.match(rangeTooltipSource, /<ThresholdStatusRing status=\{item\.status\} \/>/)
  assert.match(rangeTooltipSource, /sho-health-score-range-text/)
  assert.match(rangeTooltipSource, /sho-health-score-range-label/)
  assert.match(overviewTableSource, /HealthScoreColumnTitle scoreThresholds=\{scoreThresholds\}/)
  assert.match(serviceTableSource, /HealthScoreColumnTitle scoreThresholds=\{scoreThresholds\}/)
})

test('summary anomaly cards add icons to all summary titles', () => {
  const summarySource = between('{summaryCards.map((card) => (', '</SummaryMetricCard>')

  assert.match(summarySource, /summaryCardIcons\[card\.key\]/)
  assert.match(source, /abnormalSystems: \(/)
  assert.match(source, /attentionSystems: \(/)
  assert.match(source, /pendingIssues: \(/)
  assert.match(source, /<TriangleAlert/)
  assert.match(source, /<ListTodo/)
  assert.match(source, /sho-summary-title-icon-critical/)
  assert.match(source, /<CircleAlert/)
  assert.match(source, /sho-summary-title-icon-warning/)
})
