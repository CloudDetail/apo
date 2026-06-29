import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

const source = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')

const sliceBetween = (start, end) => {
  const startIndex = source.indexOf(start)
  assert.notEqual(startIndex, -1, `${start} should exist`)
  const endIndex = source.indexOf(end, startIndex)
  assert.notEqual(endIndex, -1, `${end} should exist after ${start}`)
  return source.slice(startIndex, endIndex)
}

test('overview system names do not render an aggregate status lamp', () => {
  const overviewRowsSource = sliceBetween(
    '{overviewSystems.map((system) => (',
    '{metricLabels.map(([key]) => (',
  )
  const nameCellSource = sliceBetween(
    '<div className="sho-name-cell">',
    '<HealthRing score={system.score} scoreThresholds={scoreThresholds} />',
  )

  assert.ok(overviewRowsSource.includes('<strong>{system.name}</strong>'))
  assert.ok(nameCellSource.includes('<strong>{system.name}</strong>'))
  assert.equal(nameCellSource.includes('<Lamp'), false)
})
