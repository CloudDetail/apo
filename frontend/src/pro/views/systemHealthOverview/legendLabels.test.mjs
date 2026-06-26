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

  assert.match(statusMetaSource, /label: '正常'/)
  assert.match(statusMetaSource, /label: '关注'/)
  assert.match(statusMetaSource, /label: '异常'/)
  assert.match(statusMetaSource, /label: '无数据'/)
  assert.doesNotMatch(statusMetaSource, /description:/)
  assert.doesNotMatch(legendSource, /meta\.description/)
  assert.doesNotMatch(legendSource, /<small>/)
})

test('health score tooltip uses score-ring samples without numbers and existing legend tags', () => {
  const tooltipSource = between(
    'const HealthScoreTooltip = () => (',
    'const SystemMetricTrendChart',
  )

  assert.match(tooltipSource, /sho-health-score-help-ring/)
  assert.match(
    tooltipSource,
    /className=\{`sho-legend-item sho-health-score-help-tag sho-health-score-help-tag-\$\{item\.status\}`\}/,
  )
  assert.doesNotMatch(tooltipSource, /HealthRing/)
  assert.doesNotMatch(tooltipSource, /sho-tag/)
  assert.doesNotMatch(tooltipSource, /sampleScore/)
})

test('summary anomaly cards add severity icons to severe and general anomaly titles', () => {
  const summarySource = between('{summaryCards.map((card) => (', '</SummaryMetricCard>')

  assert.match(summarySource, /card\.key === 'severeMetrics'/)
  assert.match(summarySource, /<TriangleAlert/)
  assert.match(summarySource, /sho-summary-title-icon-critical/)
  assert.match(summarySource, /card\.key === 'generalMetrics'/)
  assert.match(summarySource, /<CircleAlert/)
  assert.match(summarySource, /sho-summary-title-icon-warning/)
})
