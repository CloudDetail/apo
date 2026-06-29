import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

const css = readFileSync(new URL('./index.css', import.meta.url), 'utf8')
const viewSource = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')

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

test('summary ratio value uses neutral text with optional red numerator', () => {
  const valueDeclarations = declarationsFor('.sho-summary-value')
  const numeratorDeclarations = declarationsFor('.sho-summary-value-critical')

  assert.ok(valueDeclarations.includes('color: var(--ant-color-text)'))
  assert.ok(valueDeclarations.includes('font-size: 34px'))
  assert.ok(valueDeclarations.includes('font-variant-numeric: tabular-nums'))
  assert.ok(numeratorDeclarations.includes('color: #f04438'))
})

test('summary ratio value is not truncated by the generic card value cap', () => {
  const declarations = declarationsFor(
    '.sho-summary-card-ratio .sho-summary-statistic .ant-statistic-content',
  )

  assert.ok(declarations.includes('max-width: none'))
})

test('summary metric cards render their large values through Statistic', () => {
  assert.ok(viewSource.includes('Statistic'))
  assert.ok(viewSource.includes('<Statistic'))
  assert.equal(viewSource.includes('<strong>{renderedValue}</strong>'), false)
})

test('abnormal system summary title uses a question tooltip while settings live in toolbar', () => {
  assert.equal(viewSource.includes('help={card.help}'), false)
  assert.match(viewSource, /help=\{card\.key === 'abnormalSystems' \? null : card\.help\}/)
  assert.ok(viewSource.includes('className="sho-toolbar-actions"'))
  assert.ok(viewSource.includes('className="sho-threshold-toolbar-trigger"'))
  assert.ok(viewSource.includes('content={renderThresholdSettings()}'))
  assert.ok(viewSource.includes('className="sho-summary-question-trigger"'))
  assert.ok(viewSource.includes('当前健康度低于'))
})

test('abnormal system help can open threshold settings from its link', () => {
  assert.ok(
    viewSource.includes(
      'const [thresholdSettingsOpen, setThresholdSettingsOpen] = useState(false)',
    ),
  )
  assert.ok(viewSource.includes('open={thresholdSettingsOpen}'))
  assert.ok(viewSource.includes('className="sho-abnormal-system-help-link"'))
  assert.ok(viewSource.includes('前往设置'))
  assert.ok(viewSource.includes('setThresholdSettingsOpen(true)'))
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
  const summaryQuestionDeclarations = declarationsFor('.sho-summary-question-trigger')

  assert.ok(viewSource.includes('className="sho-threshold-inputs"'))
  assert.ok(viewSource.includes('className="sho-threshold-abnormal-note"'))
  assert.ok(viewSource.includes('系统健康度低于 {generalThreshold} 分，计为异常系统'))
  assert.ok(viewSource.includes('className="sho-threshold-legend"'))
  assert.ok(viewSource.includes('<ThresholdStatusRing status="red" />'))
  assert.ok(viewSource.includes('<ThresholdStatusRing status="yellow" />'))
  assert.ok(viewSource.includes('<ThresholdStatusRing status="green" />'))
  assert.ok(viewSource.includes('severeRange'))
  assert.ok(viewSource.includes('一般异常（亚健康）'))
  assert.ok(viewSource.includes('健康'))
  assert.equal(viewSource.includes('icon={<Check'), false)
  assert.ok(popoverDeclarations.includes('width: 360px'))
  assert.ok(inputDeclarations.includes('grid-template-columns: auto 72px 1fr'))
  assert.ok(noteDeclarations.includes('background: var(--ant-color-fill-quaternary)'))
  assert.ok(noteDeclarations.includes('font-weight: 650'))
  assert.ok(legendDeclarations.includes('grid-template-columns: 30px 76px minmax(0, 1fr)'))
  assert.ok(ringDeclarations.some((declaration) => declaration.includes('conic-gradient')))
  assert.ok(actionDeclarations.includes('justify-content: flex-end'))
  assert.ok(saveDeclarations.includes('min-width: 64px'))
  assert.ok(toolbarTriggerDeclarations.includes('height: 30px'))
  assert.ok(summaryQuestionDeclarations.includes('width: 24px'))
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

  assert.ok(declarations.includes('min-height: 56px'))
  assert.ok(declarations.includes('display: flex'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('justify-content: center'))
  assert.ok(declarations.includes('margin-bottom: 10px'))
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
  assert.ok(declarations.includes('min-height: 56px'))
  assert.ok(declarations.includes('display: flex'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('justify-content: flex-start'))
  assert.ok(declarations.includes('background: var(--ant-color-bg-container)'))
  assert.ok(declarations.includes('border: 1px solid var(--ant-color-border-secondary)'))
  assert.equal(
    declarations.some((declaration) => declaration.startsWith('max-width')),
    false,
  )
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

test('compact detail title and exception label are enlarged without changing controls', () => {
  const titleDeclarations = declarationsFor('.sho-detail-compact-title strong')
  const exceptionDeclarations = declarationBlocksFor('.sho-detail-compact-exception-label').flat()
  const backDeclarations = declarationsFor('.sho-detail-compact-back.ant-btn')
  const tagDeclarations = declarationsFor('.sho-detail-compact-tags .ant-tag')

  assert.ok(titleDeclarations.includes('font-size: 18px'))
  assert.ok(titleDeclarations.includes('line-height: 26px'))
  assert.ok(exceptionDeclarations.includes('font-size: 16px'))
  assert.ok(exceptionDeclarations.includes('line-height: 22px'))
  assert.ok(exceptionDeclarations.includes('font-weight: 700'))
  assert.ok(backDeclarations.includes('font-size: 12px'))
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

  assert.ok(declarations.includes('grid-template-columns: minmax(0, 1fr) 38px'))
  assert.ok(declarations.includes('grid-template-rows: 118px'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(valueDeclarations.includes('align-self: center'))
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
