import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

const source = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')
const tableSource = readFileSync(new URL('./components/OverviewSystemTable.jsx', import.meta.url), 'utf8')
const systemScoreTrendCellSource = readFileSync(
  new URL('./components/SystemHealthScoreTrendCell.jsx', import.meta.url),
  'utf8',
)
const detailTrendSectionSource = readFileSync(
  new URL('./components/DetailTrendSection.jsx', import.meta.url),
  'utf8',
)
const cssSource = readFileSync(new URL('./index.css', import.meta.url), 'utf8')

const sliceBetweenIn = (targetSource, start, end) => {
  const startIndex = targetSource.indexOf(start)
  assert.notEqual(startIndex, -1, `${start} should exist`)
  const endIndex = targetSource.indexOf(end, startIndex)
  assert.notEqual(endIndex, -1, `${end} should exist after ${start}`)
  return targetSource.slice(startIndex, endIndex)
}

const sliceBetween = (start, end) => sliceBetweenIn(source, start, end)
const sliceTableBetween = (start, end) => sliceBetweenIn(tableSource, start, end)
const sliceSystemScoreTrendCellBetween = (start, end) =>
  sliceBetweenIn(systemScoreTrendCellSource, start, end)
const sliceDetailTrendSectionBetween = (start, end) =>
  sliceBetweenIn(detailTrendSectionSource, start, end)

test('overview system names do not render an aggregate status lamp', () => {
  const overviewRowsSource = sliceTableBetween(
    '{overviewSystems.map((system) => (',
    '{metricLabels.map(([key]) => (',
  )
  const nameCellSource = sliceTableBetween(
    '<div className="sho-name-cell">',
    '<SystemHealthScoreTrendCell',
  )

  assert.ok(overviewRowsSource.includes('<strong>{system.name}</strong>'))
  assert.ok(nameCellSource.includes('<strong>{system.name}</strong>'))
  assert.equal(nameCellSource.includes('<Lamp'), false)
})

test('overview system name opens detail through a keyboard-focusable control', () => {
  const rowSource = sliceTableBetween(
    '{overviewSystems.map((system) => (',
    '{metricLabels.map(([key]) => (',
  )
  const nameCellSource = sliceTableBetween(
    '<div className="sho-name-cell">',
    '<SystemHealthScoreTrendCell',
  )

  assert.ok(rowSource.includes('<tr key={system.groupId} onClick={() => onOpenDetail(system)}>'))
  assert.ok(nameCellSource.includes('<button'))
  assert.ok(nameCellSource.includes('type="button"'))
  assert.ok(nameCellSource.includes('className="sho-system-name-button"'))
  assert.ok(nameCellSource.includes('aria-label={`查看${system.name}健康度详情`}'))
  assert.ok(nameCellSource.includes('event.stopPropagation()'))
  assert.ok(nameCellSource.includes('onOpenDetail(system)'))
  assert.ok(cssSource.includes('.sho-system-name-button:focus-visible'))
})

test('overview system clickable rows use a pointer cursor', () => {
  const rowSource = sliceTableBetween(
    '{overviewSystems.map((system) => (',
    '{metricLabels.map(([key]) => (',
  )
  const rowCursorRuleStart = cssSource.indexOf('.sho-system-table tbody tr {')

  assert.ok(rowSource.includes('<tr key={system.groupId} onClick={() => onOpenDetail(system)}>'))
  assert.notEqual(rowCursorRuleStart, -1, 'system table row cursor rule should exist')

  const rowCursorRuleEnd = cssSource.indexOf('}', rowCursorRuleStart)
  const rowCursorRule = cssSource.slice(rowCursorRuleStart, rowCursorRuleEnd)
  assert.ok(rowCursorRule.includes('cursor: pointer'))
})

test('overview system history value cell keeps the default cursor', () => {
  const rowSource = sliceTableBetween(
    '{overviewSystems.map((system) => (',
    '{metricLabels.map(([key]) => (',
  )
  const trendCellCursorRuleStart = cssSource.indexOf('.sho-system-trend-cell-td {')

  assert.ok(rowSource.includes('<td className="sho-system-trend-cell-td">'))
  assert.notEqual(trendCellCursorRuleStart, -1, 'system trend cell cursor rule should exist')

  const trendCellCursorRuleEnd = cssSource.indexOf('}', trendCellCursorRuleStart)
  const trendCellCursorRule = cssSource.slice(trendCellCursorRuleStart, trendCellCursorRuleEnd)
  assert.ok(trendCellCursorRule.includes('cursor: default'))
})

test('overview system table shows history value trend after health score', () => {
  const headerSource = sliceTableBetween('<th>系统</th>', '{metricLabels.map(([, label]) => (')
  const rowSource = sliceTableBetween(
    '{overviewSystems.map((system) => (',
    '{metricLabels.map(([key]) => (',
  )

  const scoreHeaderIndex = headerSource.indexOf(
    '<HealthScoreColumnTitle scoreThresholds={scoreThresholds} />',
  )
  const trendHeaderIndex = headerSource.indexOf('历史值')
  const scoreCellIndex = rowSource.indexOf('<HealthRing')
  const trendCellIndex = rowSource.indexOf('<SystemHealthScoreTrendCell')

  assert.notEqual(scoreHeaderIndex, -1, 'health score header should exist')
  assert.notEqual(trendHeaderIndex, -1, 'history value trend header should exist')
  assert.equal(headerSource.includes('健康度历史趋势图'), false)
  assert.ok(scoreHeaderIndex < trendHeaderIndex)
  assert.notEqual(scoreCellIndex, -1, 'health score cell should exist')
  assert.ok(rowSource.includes('status={system.status}'))
  assert.notEqual(trendCellIndex, -1, 'health score trend cell should exist')
  assert.ok(scoreCellIndex < trendCellIndex)
  assert.ok(rowSource.includes('systemTrendRows[String(system.groupId)]'))
})

test('overview system history value trend uses a larger hover chart tooltip', () => {
  const trendCellSource = sliceSystemScoreTrendCellBetween(
    'export default function SystemHealthScoreTrendCell({',
    'SystemHealthScoreTrendCell.propTypes = {',
  )
  const rowSource = sliceTableBetween(
    '{overviewSystems.map((system) => (',
    '{metricLabels.map(([key]) => (',
  )

  assert.ok(trendCellSource.includes('overlayClassName="sho-system-trend-tooltip"'))
  assert.ok(trendCellSource.includes('title={'))
  assert.ok(trendCellSource.includes('className="sho-system-trend-tooltip-chart"'))
  assert.ok(trendCellSource.includes('className="sho-system-trend-tooltip-title"'))
  assert.ok(trendCellSource.includes("const safeSystemName = safeText(systemName) || '系统'"))
  assert.ok(trendCellSource.includes('${safeSystemName}${trendRangeLabel}健康度趋势图'))
  assert.ok(trendCellSource.includes('buildSystemHealthScoreTooltipOption(rows, granularity, scoreThresholds)'))
  assert.ok(rowSource.includes('systemName={system.name}'))
  assert.ok(rowSource.includes('trendRangeLabel={systemTrendRangeLabel}'))
  assert.ok(rowSource.includes('scoreThresholds={scoreThresholds}'))
  assert.ok(trendCellSource.includes("height: 220"))
  assert.ok(trendCellSource.includes('viewBox="0 0 116 34"'))
})

test('overview system history value sparkline uses svg without nested echarts tooltip', () => {
  const sparklineSource = sliceSystemScoreTrendCellBetween(
    'const buildSystemHealthScoreSparklineSegments =',
    'const buildSystemHealthScoreTooltipOption =',
  )

  assert.ok(sparklineSource.includes('segments'))
  assert.ok(sparklineSource.includes("trendDataPointValue(row, 'value')"))
  assert.equal(sparklineSource.includes("trigger: 'axis'"), false)
  assert.equal(sparklineSource.includes('formatSystemHealthScoreTooltip'), false)
})

test('overview system history value sparkline renders an area under the line', () => {
  const sparklineSource = sliceSystemScoreTrendCellBetween(
    'const buildSystemHealthScoreSparklineSegments =',
    'const buildSystemHealthScoreTooltipOption =',
  )
  const trendCellSource = sliceSystemScoreTrendCellBetween(
    'export default function SystemHealthScoreTrendCell({',
    'SystemHealthScoreTrendCell.propTypes = {',
  )

  assert.ok(sparklineSource.includes('areaPath'))
  assert.ok(sparklineSource.includes('linePoints'))
  assert.ok(trendCellSource.includes('className="sho-system-trend-svg-area"'))
  assert.ok(trendCellSource.indexOf('sho-system-trend-svg-area') < trendCellSource.indexOf('sho-system-trend-svg-line'))
  assert.ok(cssSource.includes('.sho-system-trend-svg-area'))
})

test('overview system history value hover chart shows readable x and y axes', () => {
  const tooltipOptionSource = sliceSystemScoreTrendCellBetween(
    'const buildSystemHealthScoreTooltipOption = (',
    'export default function SystemHealthScoreTrendCell',
  )

  assert.ok(tooltipOptionSource.includes('scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS'))
  assert.ok(tooltipOptionSource.includes('grid: { left: 58, right: 72, top: 18, bottom: 36 }'))
  assert.ok(tooltipOptionSource.includes("axisLabel: { color: '#667085', fontSize: 11, hideOverlap: true }"))
  assert.ok(tooltipOptionSource.includes("axisLine: { lineStyle: { color: 'rgba(102, 112, 133, 0.35)' } }"))
  assert.ok(tooltipOptionSource.includes("name: '健康度'"))
  assert.ok(tooltipOptionSource.includes("nameLocation: 'middle'"))
  assert.ok(tooltipOptionSource.includes('nameRotate: 90'))
  assert.ok(tooltipOptionSource.includes('nameGap: 42'))
  assert.equal(tooltipOptionSource.includes('padding: [0, 0, 0, -18]'), false)
  assert.ok(tooltipOptionSource.includes("splitLine: { lineStyle: { color: 'rgba(102, 112, 133, 0.14)' } }"))
})

test('overview system history value hover chart keeps missing buckets as gaps', () => {
  const tooltipOptionSource = sliceSystemScoreTrendCellBetween(
    'const buildSystemHealthScoreTooltipOption = (',
    'export default function SystemHealthScoreTrendCell',
  )

  assert.ok(tooltipOptionSource.includes("data: rows.map((row) => trendDataPointValue(row, 'value'))"))
  assert.ok(tooltipOptionSource.includes('connectNulls: false'))
  assert.ok(
    tooltipOptionSource.includes("graphic: rowsHaveTrendData(rows, 'value') ? undefined : noDataTrendGraphic()"),
  )
})

test('overview system history value sparkline shows a compact empty state', () => {
  const trendCellSource = sliceSystemScoreTrendCellBetween(
    'export default function SystemHealthScoreTrendCell({',
    'SystemHealthScoreTrendCell.propTypes = {',
  )

  assert.ok(trendCellSource.includes("const sparklineHasData = rowsHaveTrendData(rows, 'value')"))
  assert.ok(trendCellSource.includes('sho-system-trend-svg-empty-line'))
  assert.ok(trendCellSource.includes('sho-system-trend-svg-empty-text'))
  assert.ok(trendCellSource.includes('无数据'))
})

test('overview system history value hover chart renders score judgement baselines', () => {
  const tooltipOptionSource = sliceSystemScoreTrendCellBetween(
    'const buildSystemHealthScoreTooltipOption = (',
    'export default function SystemHealthScoreTrendCell',
  )

  assert.ok(tooltipOptionSource.includes('const baselineItems = buildHealthScoreBaselineItems(scoreThresholds)'))
  assert.ok(tooltipOptionSource.includes('markLine'))
  assert.ok(tooltipOptionSource.includes('baselineItems.map((item) => ({'))
  assert.ok(tooltipOptionSource.includes('name: `${item.label} ${item.value}`'))
  assert.ok(tooltipOptionSource.includes('yAxis: item.value'))
  assert.ok(tooltipOptionSource.includes("lineStyle: { color: healthScoreBaselineColor(item), type: 'dashed', width: 1.4 }"))
  assert.ok(tooltipOptionSource.includes('color: healthScoreBaselineColor(item)'))
})

test('overview system history value tooltip shows health score value only', () => {
  const formatterSource = sliceSystemScoreTrendCellBetween(
    'const formatSystemHealthScoreTooltip = (rows, granularity, params = []) => {',
    'const buildSystemHealthScoreSparklineSegments =',
  )

  assert.ok(formatterSource.includes("健康度：${formatTrendNumber(row.value, 'systemHealthScore')}"))
  assert.equal(formatterSource.includes('current'), false)
  assert.equal(formatterSource.includes('平均值'), false)
  assert.equal(formatterSource.includes('期末值'), false)
  assert.equal(formatterSource.includes('stats.avg'), false)
  assert.equal(formatterSource.includes('stats.last'), false)
})

test('overview system table uses server pagination and batch trend loading', () => {
  assert.ok(source.includes('DEFAULT_SYSTEM_PAGE_SIZE'))
  assert.ok(source.includes('getSystemHealthSystemsApi'))
  assert.ok(source.includes('systemPageTotal'))
  assert.ok(source.includes('setSystemPage'))
  assert.ok(source.includes('getSystemHealthSystemsTrendApi'))
  assert.ok(source.includes('buildTrendChartRows,'))
  assert.ok(source.includes('groupIds: overviewSystemTrendGroupIds'))
  assert.equal(
    source.includes('Promise.all(\n          overviewSystemTrendGroupIds.map'),
    false,
    'system trend rows should not issue one trend request per visible system',
  )
})

test('overview system search debounces server pagination requests', () => {
  assert.ok(source.includes("import { useDebounce } from 'react-use'"))
  assert.ok(source.includes('SYSTEM_SEARCH_DEBOUNCE_MS'))
  assert.ok(source.includes('const [debouncedSystemKeyword, setDebouncedSystemKeyword] = useState'))
  assert.ok(source.includes('setDebouncedSystemKeyword(systemKeyword.trim())'))
  assert.ok(source.includes('keyword: debouncedSystemKeyword'))
  assert.equal(
    source.includes('keyword: systemKeyword.trim()'),
    false,
    'system search should not request on every input character',
  )
  assert.ok(source.includes('}, [queryParams, systemHealthFilter, debouncedSystemKeyword])'))
  assert.equal(
    source.includes('}, [queryParams, systemHealthFilter, systemKeyword])'),
    false,
    'page reset should wait for the debounced search term',
  )
})

test('detail service table uses server pagination separate from the detail header', () => {
  assert.ok(source.includes('DEFAULT_SERVICE_PAGE_SIZE'))
  assert.ok(source.includes('getSystemHealthServicesApi'))
  assert.ok(source.includes('servicePageTotal'))
  assert.ok(source.includes('setServicePage'))
  assert.equal(
    source.includes('[...selectedSystem.services]'),
    false,
    'service rows should not be filtered and paged from the selected system payload',
  )
})

test('overview and detail trend ranges share one persisted selection', () => {
  assert.ok(source.includes('trendRange'))
  assert.ok(source.includes('handleTrendRangeChange'))
  assert.ok(source.includes('readStoredTrendRange'))
  assert.equal(source.includes('detailTrendRange'), false)
  assert.equal(source.includes('setDetailTrendRange'), false)

  const detailTrendSelectSource = sliceDetailTrendSectionBetween(
    'aria-label="趋势分析时间范围"',
    '<DetailHealthScoreTrendChart',
  )
  assert.ok(
    detailTrendSelectSource.includes('value={trendRange}'),
    'detail range selector should use the shared trendRange',
  )
  assert.ok(
    detailTrendSelectSource.includes('onChange={onRangeChange}'),
    'detail range selector should update the shared persisted trend range',
  )
  assert.ok(source.includes('onRangeChange={handleTrendRangeChange}'))
  assert.ok(source.includes('range: trendRange'))
  assert.ok(source.includes('}, [selectedSystemId, trendRange, detailTrendRefreshToken])'))
})

test('overview and detail trend retries are independent', () => {
  assert.ok(source.includes('overviewTrendRefreshToken'))
  assert.ok(source.includes('setOverviewTrendRefreshToken'))
  assert.ok(source.includes('detailTrendRefreshToken'))
  assert.ok(source.includes('setDetailTrendRefreshToken'))
  assert.equal(
    source.includes('const handleTrendRetry ='),
    false,
    'detail trend retry should not refetch hidden overview trends',
  )
  assert.ok(source.includes('onTrendRetry={handleOverviewTrendRetry}'))
  assert.ok(source.includes('onRetry={handleDetailTrendRetry}'))
  assert.ok(source.includes('}, [trendRange, overviewTrendRefreshToken, selectedSystemId])'))
  assert.ok(source.includes('}, [selectedSystemId, trendRange, detailTrendRefreshToken])'))
})

test('detail view pauses hidden overview and system trend requests', () => {
  assert.ok(source.includes('if (selectedSystemId) {'))
  assert.ok(source.includes('setTrendLoading(false)'))
  assert.ok(source.includes('if (selectedSystemId || !overviewSystemTrendGroupIds.length) {'))
})

test('detail service search debounces server pagination requests', () => {
  assert.ok(source.includes('SERVICE_SEARCH_DEBOUNCE_MS'))
  assert.ok(source.includes('const [debouncedServiceKeyword, setDebouncedServiceKeyword] = useState'))
  assert.ok(source.includes('setDebouncedServiceKeyword(serviceKeyword.trim())'))
  assert.ok(source.includes('keyword: debouncedServiceKeyword'))
  assert.equal(
    source.includes('keyword: serviceKeyword.trim()'),
    false,
    'service search should not request on every input character',
  )
  assert.ok(
    source.includes('}, [queryParams, selectedSystemId, serviceHealthFilter, debouncedServiceKeyword])'),
  )
  assert.equal(
    source.includes('}, [queryParams, selectedSystemId, serviceHealthFilter, serviceKeyword])'),
    false,
    'service page reset should wait for the debounced search term',
  )
})

test('overview request is slim and detail view only refetches the selected header after time changes', () => {
  const overviewLoadSource = sliceBetween('const loadOverview = useCallback(async () => {', '  useEffect(() => {')

  assert.ok(overviewLoadSource.includes('summaryOnly: true'))
  assert.equal(source.includes('const loadDetail = async () => {'), false)
  assert.ok(source.includes('const loadDetailHeader = async () => {'))
  assert.ok(source.includes('lastDetailHeaderRequestKeyRef.current === detailHeaderRequestKey'))
  assert.ok(source.includes('includeServices: false'))
})

test('detail header refreshes from loaded system rows before using its fallback header request', () => {
  assert.ok(source.includes('const selectedSystemFromRows = useMemo('))
  assert.ok(
    source.includes('systems.find((system) => String(system.groupId) === selectedSystemId)'),
  )
  assert.ok(source.includes('setSelectedSystem(selectedSystemFromRows)'))
  assert.ok(source.includes('getSystemHealthSystemDetailApi(selectedSystemId, {'))
})

test('system score trend cell uses lightweight svg sparkline and lazy hover chart', () => {
  const trendCellSource = sliceSystemScoreTrendCellBetween(
    'export default function SystemHealthScoreTrendCell({',
    'SystemHealthScoreTrendCell.propTypes = {',
  )

  assert.ok(trendCellSource.includes('sho-system-trend-svg'))
  assert.ok(trendCellSource.includes('tooltipOpen'))
  assert.ok(trendCellSource.includes('onOpenChange={setTooltipOpen}'))
  assert.ok(trendCellSource.includes('tooltipOpen && tooltipOption ? ('))
  assert.equal(
    trendCellSource.includes('<ReactECharts notMerge option={cellOption}'),
    false,
    'table row sparkline should not create an ECharts instance',
  )
})

test('system score trend tooltip keeps a non-empty title while closed', () => {
  const trendCellSource = sliceSystemScoreTrendCellBetween(
    'export default function SystemHealthScoreTrendCell({',
    'SystemHealthScoreTrendCell.propTypes = {',
  )

  assert.equal(
    trendCellSource.includes('const tooltipChart = tooltipOpen ? ('),
    false,
    'AntD Tooltip suppresses onOpenChange when title is initially null',
  )
  assert.ok(trendCellSource.includes('const tooltipChart = ('))
  assert.ok(trendCellSource.includes('title={tooltipChart}'))
  assert.ok(trendCellSource.includes('{tooltipOpen && tooltipOption ? ('))
})
