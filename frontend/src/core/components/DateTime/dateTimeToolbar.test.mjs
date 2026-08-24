import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

const readText = (relativePath) => readFileSync(new URL(relativePath, import.meta.url), 'utf8')
const readJson = (relativePath) => JSON.parse(readText(relativePath))

const combineSource = readText('./DateTimeCombine.jsx')
const rangePickerSource = readText('./DateTimeRangePicker.jsx')
const zhLocale = readJson('../../../../public/locales/zh/core/dateTime.json')
const enLocale = readJson('../../../../public/locales/en/core/dateTime.json')

const quickRangeKeys = [
  '5m',
  '15m',
  '30m',
  '1h',
  '3h',
  '6h',
  '12h',
  'today',
  'yesterday',
  'thisWeek',
  'last7Days',
]

test('date time toolbar keeps consistent gaps between refresh age and controls', () => {
  assert.match(combineSource, /className="flex items-center gap-2"/)
})

test('quick range menu labels are localized instead of using hard-coded English names', () => {
  assert.match(rangePickerSource, /t\(`dateTimeRangePicker\.quickRanges\.\$\{key\}`/)
  assert.doesNotMatch(rangePickerSource, /<Menu\.Item key=\{key\}>\{timeRangeMap\[key\]\.name\}<\/Menu\.Item>/)

  for (const key of quickRangeKeys) {
    assert.equal(typeof zhLocale.dateTimeRangePicker.quickRanges[key], 'string')
    assert.equal(typeof enLocale.dateTimeRangePicker.quickRanges[key], 'string')
  }
})

test('selected quick range labels are localized in trigger and share popover', () => {
  assert.doesNotMatch(rangePickerSource, /timeRangeMap\[storeTimeRange\?\.rangeTypeKey\]\.name/)
  assert.doesNotMatch(combineSource, /timeRangeMap\[rangeTypeKey\]\?\.name/)
  assert.match(combineSource, /getQuickRangeLabel\(rangeTypeKey\)/)
})
