/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export const chartTypes = [
  { value: 'line', labelKey: 'line' },
  { value: 'area', labelKey: 'area' },
  { value: 'bar', labelKey: 'bar' },
  { value: 'table', labelKey: 'table' },
  { value: 'stat', labelKey: 'stat' },
  { value: 'barGauge', labelKey: 'barGauge' },
]

export const chartLegendModes = [
  { value: 'list', labelKey: 'legendList' },
  { value: 'table', labelKey: 'legendTable' },
]

const chartTypesWithLegendMode = new Set(['line', 'area', 'bar'])

export const chartSupportsLegendMode = (type) => chartTypesWithLegendMode.has(type)

export const chartSupportsTimeBrush = (type) => type === 'line' || type === 'area'

export const defaultChartSeriesLimit = 50

export const defaultBarChartSeriesLimit = 20

export const defaultChartSeriesLimitForWidget = (widget = {}) => (
  widget.type === 'bar' ? defaultBarChartSeriesLimit : defaultChartSeriesLimit
)

export const chartSeriesLimitOptions = [50, 100, 200, 500]

const maxChartSeriesLimit = 500

const chartSeriesScore = (series = {}) => {
  let score = -Infinity
  ;(series.values || []).forEach((point) => {
    const value = Number(point.Value ?? point.value)
    if (Number.isFinite(value) && value > score) score = value
  })
  return score
}

export const normalizeChartSeriesLimit = (value, fallback = defaultChartSeriesLimit) => {
  const limit = Number(value)
  return Number.isFinite(limit) && limit > 0
    ? Math.min(Math.floor(limit), maxChartSeriesLimit)
    : fallback
}

export const chartSeriesLimitForWidget = (widget = {}) => (
  normalizeChartSeriesLimit(
    widget.display?.seriesLimit ?? widget.query?.seriesLimit,
    defaultChartSeriesLimitForWidget(widget),
  )
)

export const chartPreviewSeriesLimitInfo = (widget = {}, preview = {}) => {
  const series = preview?.series || []
  const limit = chartSeriesLimitForWidget(widget)
  const limited = chartSupportsLegendMode(widget.type) && series.length > limit
  return {
    limited,
    total: series.length,
    shown: limited ? limit : series.length,
    limit,
  }
}

export const chartSeriesLimitControlInfo = (widget = {}, preview = {}) => {
  if (!chartSupportsLegendMode(widget.type)) return null
  const fallback = chartPreviewSeriesLimitInfo(widget, preview)
  const source = preview?.seriesLimit || fallback
  const limit = chartSeriesLimitForWidget(widget)
  const total = Number(source.total ?? fallback.total)
  const shown = Number(source.shown ?? fallback.shown)
  const limited = Boolean(source.limited || total > limit)
  const defaultLimit = defaultChartSeriesLimitForWidget(widget)
  const expandedPastDefault = limit > defaultLimit && total > defaultLimit
  if (!limited && !expandedPastDefault) return null
  return {
    limited,
    total,
    shown,
    limit,
  }
}

export const limitChartPreviewSeries = (widget = {}, preview = {}) => {
  if (!preview) return preview
  const info = chartPreviewSeriesLimitInfo(widget, preview)
  if (!info.limited) return preview
  const series = (preview.series || [])
    .map((item, index) => ({ item, index, score: chartSeriesScore(item) }))
    .sort((a, b) => {
      const aFinite = Number.isFinite(a.score)
      const bFinite = Number.isFinite(b.score)
      if (aFinite && bFinite && b.score !== a.score) return b.score - a.score
      if (aFinite !== bFinite) return aFinite ? -1 : 1
      return a.index - b.index
    })
    .slice(0, info.limit)
    .map(({ item }) => item)
  return {
    ...preview,
    series,
    seriesLimit: {
      ...info,
      shown: series.length,
    },
  }
}

export const normalizeChartLegendMode = (legendMode) => (
  legendMode === 'table' ? 'table' : 'list'
)

export const unitPresetOptions = [
  { value: 'bytes', labelKey: 'unitPresetBytes', unit: 'B', supportsAutoScale: true, autoScaleDefault: true },
  { value: 'milliseconds', labelKey: 'unitPresetMilliseconds', unit: 'ms', supportsAutoScale: true, autoScaleDefault: false },
  { value: 'seconds', labelKey: 'unitPresetSeconds', unit: 's', supportsAutoScale: true, autoScaleDefault: false },
  { value: 'percent', labelKey: 'unitPresetPercent', unit: '%', supportsAutoScale: true, autoScaleDefault: false },
]

const dashboardGridColumns = 4
const defaultDashboardWidgetWidth = 2
const legacyDashboardColumns = 12
const legacyTwoColumnCount = 2
const unitPresetMap = new Map(unitPresetOptions.map((preset) => [preset.value, preset]))
const legacyUnitPresetMap = new Map([
  ['b', 'bytes'],
  ['bytes', 'bytes'],
  ['ms', 'milliseconds'],
  ['s', 'seconds'],
  ['%', 'percent'],
])

export const queryItemsForQuery = (query = {}) => {
  const items = Array.isArray(query.queries) && query.queries.length
    ? query.queries
    : [{ id: 'q-1', expression: query.expression || '', legend: query.legend || '' }]
  return items.map((item, index) => ({
    id: item.id || `q-${index + 1}`,
    expression: item.expression || '',
    legend: item.legend || '',
    ...(item.hidden === true ? { hidden: true } : {}),
  }))
}

export const normalizeWidgetQuery = (query = {}) => {
  const queries = queryItemsForQuery(query)
  const first = queries[0] || { expression: '', legend: '' }
  return {
    mode: query.mode || 'custom',
    presetId: query.presetId || '',
    expression: first.expression || '',
    params: query.params || {},
    legend: first.legend || '',
    queries,
  }
}

export const defaultWidget = (index = 0) => ({
  id: `w-${Date.now()}-${index}`,
  title: 'New Widget',
  description: '',
  type: 'line',
  layout: { x: (index % legacyTwoColumnCount) * defaultDashboardWidgetWidth, y: Math.floor(index / legacyTwoColumnCount), w: defaultDashboardWidgetWidth, h: 1 },
  query: normalizeWidgetQuery({ mode: 'custom', presetId: '', expression: 'up', params: {} }),
  display: { unit: '', unitPreset: '', autoScale: false, decimals: 2, legendMode: 'list', warningThreshold: undefined, criticalThreshold: undefined },
})

export const mockPaletteWidget = (type = 'line', title = '') => ({
  id: `palette-${type}`,
  title,
  type,
  display: type === 'table'
    ? { unit: 'ms', unitPreset: 'milliseconds', autoScale: false, decimals: 1 }
    : {
      unit: '%',
      unitPreset: 'percent',
      autoScale: false,
      decimals: 1,
      ...(chartSupportsLegendMode(type) ? { legendMode: 'list' } : {}),
    },
  query: { mode: 'custom', expression: '' },
})

export const mockPalettePreview = (type = 'line') => ({
  status: 'success',
  series: [
    {
      metric: { svc_name: 'checkout' },
      values: [
        { TimeStamp: 1_000_000, Value: type === 'bar' ? 18 : 12 },
        { TimeStamp: 1_060_000, Value: type === 'bar' ? 32 : 24 },
        { TimeStamp: 1_120_000, Value: type === 'bar' ? 24 : 18 },
      ],
    },
    {
      metric: { svc_name: 'cart' },
      values: [
        { TimeStamp: 1_000_000, Value: 9 },
        { TimeStamp: 1_060_000, Value: 16 },
        { TimeStamp: 1_120_000, Value: 28 },
      ],
    },
  ],
})

const normalizeUnitText = (unit) => String(unit || '').trim()

const normalizeTableColumns = (columns = []) => (
  Array.isArray(columns)
    ? columns
      .map((column) => ({
        name: String(column?.name || '').trim(),
        label: String(column?.label || column?.title || column?.name || '').trim(),
      }))
      .filter((column) => column.name)
    : []
)

const normalizeTableHiddenColumns = (columns = []) => (
  Array.isArray(columns)
    ? columns.map((column) => String(column || '').trim()).filter(Boolean)
    : []
)

const resolveDisplayUnit = (display = {}) => {
  const configuredPreset = String(display.unitPreset || '')
  const unit = normalizeUnitText(display.unit)
  if (unitPresetMap.has(configuredPreset)) {
    const preset = unitPresetMap.get(configuredPreset)
    return {
      unitPreset: configuredPreset,
      unit: unit || preset.unit,
      supportsAutoScale: preset.supportsAutoScale,
      autoScale: display.autoScale === undefined ? preset.autoScaleDefault : Boolean(display.autoScale),
    }
  }
  if (configuredPreset === 'custom') {
    return { unitPreset: 'custom', unit, supportsAutoScale: false, autoScale: false }
  }
  const legacyPreset = legacyUnitPresetMap.get(unit.toLowerCase())
  if (legacyPreset) {
    const preset = unitPresetMap.get(legacyPreset)
    return {
      unitPreset: legacyPreset,
      unit: unit || preset.unit,
      supportsAutoScale: preset.supportsAutoScale,
      autoScale: legacyPreset === 'bytes',
    }
  }
  return { unitPreset: unit ? 'custom' : '', unit, supportsAutoScale: false, autoScale: false }
}

export const normalizeWidgetDisplay = (display = {}) => {
  const resolved = resolveDisplayUnit(display)
  const tableColumns = normalizeTableColumns(display.tableColumns)
  const tableHiddenColumns = normalizeTableHiddenColumns(display.tableHiddenColumns)
  return {
    unit: resolved.unit,
    unitPreset: resolved.unitPreset,
    autoScale: resolved.autoScale,
    decimals: Number.isFinite(Number(display.decimals)) ? Number(display.decimals) : 2,
    ...(tableColumns.length ? { tableColumns } : {}),
    ...(tableHiddenColumns.length ? { tableHiddenColumns } : {}),
    ...(display.showValueColumn === false || display.showValueColumn === true ? { showValueColumn: display.showValueColumn } : {}),
    warningThreshold: optionalNumber(display.warningThreshold),
    criticalThreshold: optionalNumber(display.criticalThreshold),
  }
}

export const displayForUnitPreset = (unitPreset = '', fallbackUnit = '') => {
  const preset = unitPresetMap.get(unitPreset)
  if (!preset) {
    return {
      unit: fallbackUnit,
      unitPreset: fallbackUnit ? 'custom' : '',
      autoScale: false,
    }
  }
  return {
    unit: preset.unit,
    unitPreset: preset.value,
    autoScale: preset.autoScaleDefault,
  }
}

export const unitPresetSupportsAutoScale = (display = {}) => resolveDisplayUnit(display).supportsAutoScale

const formatNumberWithUnit = (value, unit, decimals) => {
  const formatted = Number(value).toFixed(decimals)
  return unit ? `${formatted} ${unit}` : formatted
}

const scaledUnit = (numeric, units) => units.find((item) => Math.abs(numeric) >= item.value) || units[units.length - 1]

const byteUnits = [
  { value: 1024 ** 4, suffix: 'TiB' },
  { value: 1024 ** 3, suffix: 'GiB' },
  { value: 1024 ** 2, suffix: 'MiB' },
  { value: 1024, suffix: 'KiB' },
  { value: 1, suffix: 'B' },
]

const millisecondUnits = [
  { value: 60000, suffix: 'min' },
  { value: 1000, suffix: 's' },
  { value: 1, suffix: 'ms' },
  { value: 0.001, suffix: 'us' },
  { value: 0.000001, suffix: 'ns' },
]

const secondUnits = [
  { value: 3600, suffix: 'h' },
  { value: 60, suffix: 'min' },
  { value: 1, suffix: 's' },
  { value: 0.001, suffix: 'ms' },
  { value: 0.000001, suffix: 'us' },
  { value: 0.000000001, suffix: 'ns' },
]

const formatScaledUnitValue = (numeric, units, decimals) => {
  const unit = scaledUnit(numeric, units)
  return `${(numeric / unit.value).toFixed(decimals)} ${unit.suffix}`
}

export const displayThresholdToRaw = (value, display = {}) => {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return undefined
  const resolved = resolveDisplayUnit(display)
  if (resolved.unitPreset === 'percent' && resolved.autoScale) return numeric / 100
  return numeric
}

export const formatWidgetValue = (value, display = {}) => {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return value ?? '-'
  const decimals = Number.isFinite(Number(display.decimals)) ? Number(display.decimals) : 2
  const resolved = resolveDisplayUnit(display)
  if (resolved.unitPreset === 'bytes' && resolved.autoScale) return formatScaledUnitValue(numeric, byteUnits, decimals)
  if (resolved.unitPreset === 'milliseconds' && resolved.autoScale) return formatScaledUnitValue(numeric, millisecondUnits, decimals)
  if (resolved.unitPreset === 'seconds' && resolved.autoScale) return formatScaledUnitValue(numeric, secondUnits, decimals)
  if (resolved.unitPreset === 'percent' && resolved.autoScale) return formatNumberWithUnit(numeric * 100, resolved.unit, decimals)
  return formatNumberWithUnit(numeric, resolved.unit, decimals)
}

export const barGaugeProgressPercent = (value, display = {}) => {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return 0
  const resolved = resolveDisplayUnit(display)
  const displayPercent = resolved.unitPreset === 'percent' && resolved.autoScale ? numeric * 100 : numeric
  return Math.min(100, Math.max(0, Number(displayPercent.toFixed(6))))
}

const trimTrailingZeros = (value) => String(value).replace(/\.0+$/, '').replace(/(\.\d*[1-9])0+$/, '$1')

const compactAxisUnits = [
  { value: 1e12, suffix: 'T' },
  { value: 1e9, suffix: 'B' },
  { value: 1e6, suffix: 'M' },
  { value: 1e3, suffix: 'K' },
]

export const formatChartValue = (value, display = {}) => {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return value ?? '-'
  const resolved = resolveDisplayUnit(display)
  if (resolved.autoScale) return formatWidgetValue(value, display)
  const compactUnit = compactAxisUnits.find((item) => Math.abs(numeric) >= item.value)
  if (!compactUnit) return formatWidgetValue(value, display)

  const compactValue = `${trimTrailingZeros((numeric / compactUnit.value).toFixed(2))}${compactUnit.suffix}`
  return resolved.unit ? `${compactValue} ${resolved.unit}` : compactValue
}

export const formatAxisValue = formatChartValue

const escapeTooltipText = (value) => String(value ?? '')
  .replace(/&/g, '&amp;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')
  .replace(/"/g, '&quot;')
  .replace(/'/g, '&#39;')

const tooltipParamValue = (item = {}) => {
  if (item.value !== undefined && item.value !== null) {
    return Array.isArray(item.value) ? item.value[item.value.length - 1] : item.value
  }
  if (item.data !== undefined && item.data !== null) {
    return Array.isArray(item.data) ? item.data[item.data.length - 1] : item.data
  }
  return undefined
}

export const formatChartTooltip = (params = [], display = {}) => {
  const items = Array.isArray(params) ? params : [params]
  const title = escapeTooltipText(items[0]?.axisValueLabel || items[0]?.name || '')
  const rows = items
    .map((item) => ({ item, value: tooltipParamValue(item) }))
    .filter(({ value }) => value !== undefined && value !== null)

  const titleHtml = title
    ? `<div style="margin-bottom: 8px; color: inherit; font-weight: 600;">${title}</div>`
    : ''
  const rowsHtml = rows.map(({ item, value }) => `
    <div style="display: grid; grid-template-columns: minmax(120px, 1fr) auto; align-items: center; gap: 16px; min-width: 220px; max-width: 360px; margin-top: 6px;">
      <span style="display: inline-flex; align-items: center; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">${item.marker || ''}<span style="overflow: hidden; text-overflow: ellipsis;">${escapeTooltipText(item.seriesName || item.name || '')}</span></span>
      <strong style="font-variant-numeric: tabular-nums; font-weight: 600; text-align: right; white-space: nowrap;">${escapeTooltipText(formatChartValue(value, display))}</strong>
    </div>
  `).join('')

  return `<div style="min-width: 220px; max-width: 360px;">${titleHtml}${rowsHtml}</div>`
}

const chartEventOffset = (params = {}) => {
  const event = params.event || {}
  const sourceEvent = event.event || event
  const offsetX = Number(sourceEvent.offsetX ?? sourceEvent.zrX ?? params.offsetX ?? params.zrX)
  const offsetY = Number(sourceEvent.offsetY ?? sourceEvent.zrY ?? params.offsetY ?? params.zrY)
  return Number.isFinite(offsetX) && Number.isFinite(offsetY) ? [offsetX, offsetY] : null
}

const lineTooltipHitPixelTolerance = 28
const chartTooltipOffset = 12
const chartTooltipEdgePadding = 22

const clampChartTooltipPosition = (value, min, max) => Math.max(min, Math.min(value, Math.max(min, max)))

export const chartTooltipPosition = (point = [], _params, _dom, _rect, size = {}) => {
  const x = Number(point[0])
  const y = Number(point[1])
  const contentWidth = Number(size.contentSize?.[0])
  const contentHeight = Number(size.contentSize?.[1])
  const viewWidth = Number(size.viewSize?.[0])
  const viewHeight = Number(size.viewSize?.[1])
  if (
    !Number.isFinite(x)
    || !Number.isFinite(y)
    || !Number.isFinite(contentWidth)
    || !Number.isFinite(contentHeight)
    || !Number.isFinite(viewWidth)
    || !Number.isFinite(viewHeight)
  ) {
    return [chartTooltipOffset, chartTooltipOffset]
  }
  return [
    clampChartTooltipPosition(x + chartTooltipOffset, chartTooltipOffset, viewWidth - contentWidth - chartTooltipEdgePadding),
    clampChartTooltipPosition(y + chartTooltipOffset, chartTooltipOffset, viewHeight - contentHeight - chartTooltipEdgePadding),
  ]
}

const chartSeriesPointValue = (point) => {
  if (Array.isArray(point)) return Number(point[1])
  if (point && typeof point === 'object') return chartSeriesPointValue(point.value)
  return Number(point)
}

export const showNearestLineTooltip = (params = {}, chart, lineHoverSeries = [], onLineHoverPoint) => {
  if (params.dataIndex !== undefined && params.dataIndex !== null) return false
  if (!Array.isArray(lineHoverSeries) || !lineHoverSeries.length) return false
  const point = chartEventOffset(params)
  if (!point || typeof chart?.convertFromPixel !== 'function' || typeof chart?.dispatchAction !== 'function') return false
  const gridPoint = chart.convertFromPixel('grid', point)
  if (!Array.isArray(gridPoint)) return false
  const dataIndex = Math.round(Number(gridPoint[0]))
  const yValue = Number(gridPoint[1])
  if (!Number.isInteger(dataIndex) || !Number.isFinite(yValue)) return false
  let nearest = null
  for (const series of lineHoverSeries) {
    const seriesIndex = Number(series?.seriesIndex)
    const data = Array.isArray(series?.data) ? series.data : []
    if (series?.selected === false || !Number.isInteger(seriesIndex) || dataIndex < 0 || dataIndex >= data.length) continue
    const value = chartSeriesPointValue(data[dataIndex])
    if (!Number.isFinite(value)) continue
    const distance = Math.abs(value - yValue)
    if (!nearest || distance < nearest.distance) nearest = { seriesIndex, value, distance, color: series?.color }
  }
  if (!nearest) {
    onLineHoverPoint?.(null)
    return false
  }
  const nearestPoint = chart.convertToPixel?.('grid', [dataIndex, nearest.value])
  if (
    Array.isArray(nearestPoint)
    && Number.isFinite(nearestPoint[1])
    && Math.abs(nearestPoint[1] - point[1]) > lineTooltipHitPixelTolerance
  ) {
    onLineHoverPoint?.(null)
    chart.dispatchAction({ type: 'updateAxisPointer', currTrigger: 'leave' })
    chart.dispatchAction({ type: 'hideTip' })
    return false
  }
  if (Array.isArray(nearestPoint) && Number.isFinite(nearestPoint[0]) && Number.isFinite(nearestPoint[1])) {
    onLineHoverPoint?.({
      x: nearestPoint[0],
      y: nearestPoint[1],
      color: nearest.color,
      seriesIndex: nearest.seriesIndex,
      dataIndex,
    })
    chart.dispatchAction({
      type: 'updateAxisPointer',
      currTrigger: 'mousemove',
      x: nearestPoint[0],
      y: nearestPoint[1],
    })
  } else {
    onLineHoverPoint?.(null)
  }
  chart.dispatchAction({ type: 'showTip', seriesIndex: nearest.seriesIndex, dataIndex })
  return true
}

const brushAxisPosition = (value) => {
  const position = Array.isArray(value) ? value[0] : value
  const number = Number(position)
  return Number.isFinite(number) ? number : null
}

const clampBrushIndex = (index, length) => Math.max(0, Math.min(length - 1, index))

export const chartTimeRangeFromBrush = (params = {}, chart, timeValues = []) => {
  const values = timeValues
    .map((value) => Number(value))
    .filter((value) => Number.isFinite(value))
  if (values.length < 2 || typeof chart?.convertFromPixel !== 'function') return null
  const brushArea = (params.areas || []).find((area) => (
    area?.brushType === 'lineX'
    && Array.isArray(area.range)
    && area.range.length >= 2
  ))
  if (!brushArea) return null

  const leftPixel = Number(brushArea.range[0])
  const rightPixel = Number(brushArea.range[1])
  if (!Number.isFinite(leftPixel) || !Number.isFinite(rightPixel)) return null

  const leftPosition = brushAxisPosition(chart.convertFromPixel({ xAxisIndex: 0 }, leftPixel))
  const rightPosition = brushAxisPosition(chart.convertFromPixel({ xAxisIndex: 0 }, rightPixel))
  if (leftPosition === null || rightPosition === null) return null

  const minPosition = Math.min(leftPosition, rightPosition)
  const maxPosition = Math.max(leftPosition, rightPosition)
  const startIndex = clampBrushIndex(Math.floor(minPosition), values.length)
  const endIndex = clampBrushIndex(Math.ceil(maxPosition), values.length)
  const startTime = Math.round(values[Math.min(startIndex, endIndex)])
  const endTime = Math.round(values[Math.max(startIndex, endIndex)])
  if (!Number.isFinite(startTime) || !Number.isFinite(endTime) || endTime <= startTime) return null
  return { startTime, endTime }
}

export const autoLegendValue = '__auto'
const autoLegendInternalLabels = new Set(['__name__', 'query_id', 'query_legend', 'queryId', 'queryLegend'])

const escapeLegendLabelValue = (value) => String(value)
  .replace(/\\/g, '\\\\')
  .replace(/"/g, '\\"')
  .replace(/\n/g, '\\n')

const grafanaAutoLegendName = (metric = {}) => {
  const labels = Object.entries(metric)
    .filter(([name, value]) => name && !autoLegendInternalLabels.has(name) && value !== undefined && value !== null)
    .sort(([left], [right]) => left.localeCompare(right))
  if (!labels.length) return ''
  return `{${labels.map(([name, value]) => `${name}="${escapeLegendLabelValue(value)}"`).join(', ')}}`
}

export const seriesDisplayName = (metric = {}, index = 0, legend = '') => {
  const fallback = metric.svc_name || metric.pod || metric.instance || `series-${index + 1}`
  const legendTemplate = metric.query_legend || metric.queryLegend || legend
  if (!legendTemplate) return fallback
  if (String(legendTemplate).trim() === autoLegendValue) return grafanaAutoLegendName(metric) || fallback
  const resolved = legendTemplate
    .replace(/\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}/g, (_, key) => metric[key] ?? '')
    .replace(/\$([A-Za-z_][A-Za-z0-9_]*)/g, (_, key) => metric[key] ?? '')
    .trim()
  return resolved || fallback
}

const internalTableLabels = new Set(['__name__'])

const metricLabelNamesForPreview = (preview = {}) => (
  Array.from(new Set((preview.series || []).flatMap((item) => Object.keys(item.metric || {}))))
    .filter((name) => name && !internalTableLabels.has(name))
    .sort()
)

export const tableColumnsForPreview = (preview = {}, display = {}) => {
  const hidden = new Set(normalizeTableHiddenColumns(display.tableHiddenColumns))
  const configuredColumns = normalizeTableColumns(display.tableColumns)
    .filter((column) => !hidden.has(column.name))
    .map((column) => ({ kind: 'label', name: column.name, title: column.label || column.name }))
  const configuredNames = new Set(configuredColumns.map((column) => column.name))
  const extraColumns = metricLabelNamesForPreview(preview)
    .filter((name) => !configuredNames.has(name) && !hidden.has(name))
    .map((name) => ({ kind: 'label', name, title: name }))
  const labelColumns = [...configuredColumns, ...extraColumns]
  const baseColumns = labelColumns.length
    ? labelColumns
    : [{ kind: 'name', name: 'name', title: 'seriesName' }]
  return display.showValueColumn === false
    ? baseColumns
    : [...baseColumns, { kind: 'value', name: 'value', title: 'seriesValue' }]
}

export const tableRowsForPreview = (preview = {}, widget = {}) => (preview.series || []).map((item, index) => {
  const values = item.values || []
  const last = values[values.length - 1]
  const rawValue = last ? last.Value ?? last.value : undefined
  return {
    key: index,
    labels: item.metric || {},
    rawValue,
    name: seriesDisplayName(item.metric || {}, index, widget.query?.legend),
    value: last ? formatWidgetValue(rawValue, widget.display) : '-',
  }
})

export const chartGridLeft = () => 8

const optionalNumber = (value) => {
  if (value === undefined || value === null || value === '') return undefined
  const number = Number(value)
  return Number.isFinite(number) ? number : undefined
}

const normalizeThemeMode = (themeMode = 'light') => {
  if (themeMode === true) return 'dark'
  if (themeMode === false) return 'light'
  return String(themeMode).toLowerCase() === 'dark' ? 'dark' : 'light'
}

const grafanaDarkPalette = [
  '#7eb26d',
  '#e9b839',
  '#6ed0e0',
  '#ef843c',
  '#e24d42',
  '#1f78c1',
  '#ba43a9',
  '#705da0',
  '#4e8340',
]

const grafanaLightPalette = [
  '#7eb26d',
  '#eab839',
  '#6ed0e0',
  '#ef843c',
  '#e24d42',
  '#1f78c1',
  '#ba43a9',
  '#705da0',
  '#4e8340',
]

export const getChartThemeTokens = (themeMode = 'light') => {
  const mode = normalizeThemeMode(themeMode)
  const isDark = mode === 'dark'
  return {
    mode,
    palette: isDark ? grafanaDarkPalette : grafanaLightPalette,
    textPrimary: isDark ? '#e5e7eb' : '#1f2937',
    textSecondary: isDark ? '#a9b4c7' : '#637083',
    textMuted: isDark ? '#8b98ad' : '#7a8699',
    tooltipBg: isDark ? 'rgba(15, 23, 42, 0.94)' : 'rgba(255, 255, 255, 0.96)',
    tooltipBorder: isDark ? '#334155' : '#d9e2ef',
    tooltipShadow: isDark ? '0 12px 32px rgba(0, 0, 0, 0.34)' : '0 12px 32px rgba(15, 23, 42, 0.14)',
    axisPointer: isDark ? 'rgba(122, 162, 255, 0.45)' : '#9fb5d0',
    axisShadow: isDark ? 'rgba(122, 162, 255, 0.12)' : 'rgba(22, 119, 255, 0.08)',
    splitLine: isDark ? 'rgba(148, 163, 184, 0.14)' : '#edf1f7',
    pageIcon: isDark ? '#8b98ad' : '#8aa0b6',
    pageIconInactive: isDark ? '#475569' : '#cbd5e1',
    symbolBorder: isDark ? '#111827' : '#ffffff',
    warning: isDark ? '#e9b839' : '#eab839',
    critical: '#e24d42',
  }
}

export const chartLegendRows = (widget, preview, themeMode = 'light') => {
  const tokens = getChartThemeTokens(themeMode)
  const renderPreview = limitChartPreviewSeries(widget, preview)
  return (renderPreview?.series || []).map((item, index) => {
    const values = (item.values || [])
      .map((point) => point.Value ?? point.value)
      .filter((value) => value !== undefined && value !== null)
      .map((value) => Number(value))
      .filter((value) => Number.isFinite(value))
    const min = values.length ? Math.min(...values) : undefined
    const max = values.length ? Math.max(...values) : undefined
    const mean = values.length ? values.reduce((sum, value) => sum + value, 0) / values.length : undefined

    return {
      key: index,
      color: tokens.palette[index % tokens.palette.length],
      name: seriesDisplayName(item.metric || {}, index, widget.query?.legend),
      min: min === undefined ? '-' : formatAxisValue(min, widget.display),
      max: max === undefined ? '-' : formatAxisValue(max, widget.display),
      mean: mean === undefined ? '-' : formatAxisValue(mean, widget.display),
    }
  })
}

const hexToRgba = (hex, alpha) => {
  const normalized = String(hex || '').replace('#', '')
  if (!/^[0-9a-fA-F]{6}$/.test(normalized)) return hex
  const red = parseInt(normalized.slice(0, 2), 16)
  const green = parseInt(normalized.slice(2, 4), 16)
  const blue = parseInt(normalized.slice(4, 6), 16)
  return `rgba(${red}, ${green}, ${blue}, ${alpha})`
}

const gradientAreaStyle = (widgetType, color, mode) => {
  if (widgetType !== 'area') return undefined
  return {
    opacity: mode === 'dark' ? 0.2 : 0.24,
    color: {
      type: 'linear',
      x: 0,
      y: 0,
      x2: 0,
      y2: 1,
      colorStops: [
        { offset: 0, color: hexToRgba(color, mode === 'dark' ? 0.42 : 0.34) },
        { offset: 0.58, color: hexToRgba(color, mode === 'dark' ? 0.18 : 0.14) },
        { offset: 1, color: hexToRgba(color, 0) },
      ],
    },
  }
}

const lineSeriesStyle = (color, mode) => ({
  width: 2.4,
  color,
})

const barItemStyle = (color) => ({
  borderRadius: [8, 8, 2, 2],
  color,
})

const chartTooltipAxisPointer = (widgetType, tokens) => {
  if (widgetType === 'bar') {
    return {
      type: 'shadow',
      shadowStyle: { color: tokens.axisShadow },
    }
  }
  return {
    type: 'cross',
    snap: true,
    label: { show: false },
    lineStyle: { color: tokens.axisPointer, type: 'dashed' },
    crossStyle: { color: tokens.axisPointer, type: 'dashed' },
  }
}

export const thresholdMarkLine = (display = {}, themeMode = 'light') => {
  const tokens = getChartThemeTokens(themeMode)
  const warning = displayThresholdToRaw(display.warningThreshold, display)
  const critical = displayThresholdToRaw(display.criticalThreshold, display)
  const data = []
  if (warning !== undefined) {
    data.push({ yAxis: warning, lineStyle: { color: tokens.warning, type: 'dashed' }, label: { show: false } })
  }
  if (critical !== undefined) {
    data.push({ yAxis: critical, lineStyle: { color: tokens.critical, type: 'dashed' }, label: { show: false } })
  }
  return data.length ? { silent: true, symbol: 'none', data } : undefined
}

export const variableToken = (name) => `$${name}`

export const buildChartOption = (widget, preview, themeMode = 'light', selectedSeries = {}) => {
  const tokens = getChartThemeTokens(themeMode)
  const renderPreview = limitChartPreviewSeries(widget, preview)
  const series = renderPreview?.series || []
  const type = widget.type === 'bar' ? 'bar' : 'line'
  const names = series.map((item, index) => seriesDisplayName(item.metric || {}, index, widget.query?.legend))
  const legendSelected = Object.fromEntries(names.map((name) => [name, selectedSeries[name] !== false]))
  const markLine = thresholdMarkLine(widget.display, tokens.mode)
  const xSet = new Set()
  series.forEach((item) => item.values?.forEach((point) => {
    const timestamp = Number(point.TimeStamp ?? point.timeStamp)
    if (Number.isFinite(timestamp)) xSet.add(timestamp)
  }))
  const xAxis = Array.from(xSet).sort((a, b) => a - b)
  return {
    animation: false,
    color: tokens.palette,
    toolbox: { show: false },
    tooltip: {
      trigger: widget.type === 'bar' ? 'axis' : 'item',
      backgroundColor: tokens.tooltipBg,
      borderColor: tokens.tooltipBorder,
      borderWidth: 1,
      padding: [10, 12],
      confine: true,
      position: chartTooltipPosition,
      triggerOn: widget.type === 'bar' ? 'click' : 'none',
      extraCssText: `border-radius: 12px; box-shadow: ${tokens.tooltipShadow};`,
      textStyle: { color: tokens.textPrimary, fontSize: 12 },
      axisPointer: chartTooltipAxisPointer(widget.type, tokens),
      formatter: (params) => formatChartTooltip(params, widget.display),
      valueFormatter: (value) => formatChartValue(value, widget.display),
    },
    grid: { left: chartGridLeft(widget.display), right: 12, top: 12, bottom: 24, containLabel: true },
    legend: {
      show: false,
      data: names,
      selected: legendSelected,
      type: 'scroll',
      top: 0,
      icon: 'circle',
      itemWidth: 7,
      itemHeight: 7,
      textStyle: { fontSize: 11, color: tokens.textSecondary },
      pageIconColor: tokens.pageIcon,
      pageIconInactiveColor: tokens.pageIconInactive,
    },
    xAxis: {
      type: 'category',
      boundaryGap: widget.type === 'bar',
      data: xAxis.map((ts) => new Date(ts / 1000).toLocaleTimeString()),
      customDashboardTimeValues: xAxis,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: tokens.textMuted, fontSize: 11, margin: 6 },
      splitLine: { show: false },
    },
    brush: chartSupportsTimeBrush(widget.type) ? {
      toolbox: [],
      xAxisIndex: 'all',
      brushStyle: {
        borderWidth: 1,
        color: tokens.axisShadow,
        borderColor: tokens.axisPointer,
      },
    } : undefined,
    yAxis: {
      type: 'value',
      scale: true,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: {
        color: tokens.textMuted,
        fontSize: 11,
        formatter: (value) => formatChartValue(value, widget.display),
      },
      splitLine: { show: true, lineStyle: { color: tokens.splitLine, type: 'dashed' } },
    },
    series: series.map((item, index) => {
      const color = tokens.palette[index % tokens.palette.length]
      const valueMap = new Map((item.values || []).map((point) => [point.TimeStamp ?? point.timeStamp, point.Value ?? point.value]))
      return {
        name: names[index],
        type,
        areaStyle: type === 'line' ? gradientAreaStyle(widget.type, color, tokens.mode) : undefined,
        lineStyle: type === 'line' ? lineSeriesStyle(color, tokens.mode) : undefined,
        markLine: index === 0 ? markLine : undefined,
        showSymbol: type === 'line',
        showAllSymbol: false,
        symbol: 'circle',
        symbolSize: type === 'line' ? 0 : 5,
        triggerLineEvent: false,
        smooth: type === 'line' ? 0.35 : undefined,
        connectNulls: type === 'line',
        sampling: type === 'line' ? 'lttb' : undefined,
        emphasis: { disabled: true },
        barMaxWidth: type === 'bar' ? 22 : undefined,
        barCategoryGap: type === 'bar' ? '48%' : undefined,
        itemStyle: type === 'bar' ? barItemStyle(color) : { color, borderColor: tokens.symbolBorder, borderWidth: 1 },
        data: xAxis.map((ts) => valueMap.get(ts) ?? null),
      }
    }),
  }
}

export const getSingleValue = (preview) => {
  const first = preview?.series?.[0]?.values || []
  const last = first[first.length - 1]
  return last ? Number(last.Value ?? last.value).toFixed(2) : '-'
}

export const normalizeWidgetLayout = (layout = {}, index = 0) => {
  let x = Number(layout.x)
  let y = Number(layout.y)
  let width = Number(layout.w)
  let height = Number(layout.h)
  if (!Number.isFinite(x)) x = (index % legacyTwoColumnCount) * defaultDashboardWidgetWidth
  if (!Number.isFinite(y)) y = Math.floor(index / legacyTwoColumnCount)
  const legacyLayout = width > dashboardGridColumns || x >= dashboardGridColumns
  if (legacyLayout) {
    width = Number.isFinite(width) && width > 0
      ? Math.max(1, Math.round(width * dashboardGridColumns / legacyDashboardColumns))
      : defaultDashboardWidgetWidth
    width = Math.min(width, dashboardGridColumns)
    x = Math.max(0, Math.min(Math.floor(x * dashboardGridColumns / legacyDashboardColumns), dashboardGridColumns - width))
    const rowHeight = height > 0 ? height : 4
    y = Math.floor(Math.max(0, y) / rowHeight)
    height = 1
  } else if (!Number.isFinite(width) || width <= 0) {
    const legacyColumn = Math.max(0, Math.min(Math.round(x), legacyTwoColumnCount - 1))
    x = legacyColumn * defaultDashboardWidgetWidth
    width = defaultDashboardWidgetWidth
  }
  const normalizedX = Math.max(0, Math.min(dashboardGridColumns - 1, x))
  const maxWidth = dashboardGridColumns - normalizedX
  return {
    x: normalizedX,
    y: Math.max(0, y),
    w: Number.isFinite(width) && width > 0 ? Math.max(1, Math.min(Math.round(width), maxWidth)) : 1,
    h: Number.isFinite(height) && height > 0 ? Math.max(1, Math.round(height)) : 1,
  }
}

export const normalizeWidget = (widget = {}, index = 0) => {
  const type = widget.type === 'topn' ? 'table' : (widget.type || 'line')
  return {
    ...widget,
    type,
    legacyType: widget.type === 'topn' ? 'topn' : widget.legacyType,
    unsupportedType: widget.type === 'single' || widget.unsupportedType || false,
    layout: normalizeWidgetLayout(widget.layout, index),
    query: normalizeWidgetQuery(widget.query),
    display: {
      ...normalizeWidgetDisplay(widget.display),
      ...(chartSupportsLegendMode(type) ? { legendMode: normalizeChartLegendMode(widget.display?.legendMode) } : {}),
    },
  }
}

const isGlobalContextVariable = (variable = {}) => (
  variable?.type === 'timeRange'
  || variable?.name === 'timeRange'
  || variable?.type === 'dataGroup'
  || variable?.name === 'dataGroup'
)

const builtinResourceVariableAliases = {
  cluster: 'cluster',
  cluster_id: 'cluster',
  clusterId: 'cluster',
  clusterID: 'cluster',
  cluster_name: 'cluster',
  namespace: 'namespace',
  namespace_name: 'namespace',
  namespaceName: 'namespace',
  src_namespace: 'namespace',
  dst_namespace: 'namespace',
  svc: 'service',
  svc_name: 'service',
  service: 'service',
  service_name: 'service',
  endpoint: 'endpoint',
  endpoint_name: 'endpoint',
  content_key: 'endpoint',
  contentKey: 'endpoint',
  pod: 'pod',
  pod_name: 'pod',
  src_pod: 'pod',
  dst_pod: 'pod',
  instance: 'instance',
  instance_name: 'instance',
  node_name: 'instance',
  nodeName: 'instance',
  host_name: 'instance',
  hostName: 'instance',
}

const customLabelVariableNames = new Set([
  'related_nodes',
  'related_pid',
  'jvm_instance',
  'job_name',
  'cadvisor_job_name',
])

const builtinResourceVariableType = (variable = {}) => {
  if (customLabelVariableNames.has(String(variable.name || ''))) return ''
  return [variable.type, variable.name, ...(variable.labels || [])]
    .filter(Boolean)
    .map(String)
    .map((value) => builtinResourceVariableAliases[value] || '')
    .find(Boolean) || ''
}

const normalizeDashboardVariable = (variable = {}) => {
  const normalized = {
    ...variable,
    multi: variable.multi === false ? false : true,
  }
  const resourceType = builtinResourceVariableType(variable)
  if (!resourceType) return normalized
  const allValue = variable.allValue || '.*'
  return {
    ...normalized,
    type: resourceType,
    includeAll: true,
    allValue,
    value: allValue,
    defaultValue: allValue,
  }
}

export const normalizeDashboard = (dashboard = {}) => ({
  ...dashboard,
  variables: (dashboard?.variables || [])
    .filter((variable) => !isGlobalContextVariable(variable))
    .map(normalizeDashboardVariable),
  widgets: (dashboard?.widgets || []).map(normalizeWidget),
})

export const serializeDashboard = (dashboard = {}) => {
  const rest = { ...dashboard }
  delete rest.timeRange
  delete rest.refreshSeconds
  delete rest.updatedAt
  delete rest.lastVisitedAt
  return {
    ...rest,
    variables: (dashboard.variables || [])
      .filter((variable) => !isGlobalContextVariable(variable))
      .map(normalizeDashboardVariable),
    widgets: (dashboard.widgets || []).map(normalizeWidget),
  }
}
