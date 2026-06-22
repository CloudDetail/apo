import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

const css = readFileSync(new URL('./index.css', import.meta.url), 'utf8')

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

test('stat label keeps help icon on the same line as the text', () => {
  const declarations = declarationsFor('.sho-stat > span')

  assert.ok(declarations.includes('display: inline-flex'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('gap: 4px'))
  assert.ok(declarations.includes('white-space: nowrap'))
})

test('stat label help icon does not shrink away from the label text', () => {
  const declarations = declarationsFor('.sho-stat > span .sho-help-icon')

  assert.ok(declarations.includes('flex: 0 0 auto'))
})

test('summary number cards reserve the right side for the large value', () => {
  const declarations = declarationsFor('.sho-summary-card-number')

  assert.ok(declarations.includes('grid-template-columns: minmax(0, 1fr) auto'))
  assert.ok(declarations.includes('grid-template-rows: auto auto'))
  assert.ok(!declarations.includes('justify-content: center'))
})

test('summary ring cards reserve the right side for the circular ratio', () => {
  const declarations = declarationsFor('.sho-summary-card-ring')

  assert.ok(declarations.includes('grid-template-columns: minmax(0, 1fr) auto'))
  assert.ok(declarations.includes('align-items: center'))
  assert.ok(declarations.includes('column-gap: 16px'))
})

test('summary ring uses circular conic fill and compact center text', () => {
  const ringDeclarations = declarationsFor('.sho-summary-ring')
  const valueDeclarations = declarationsFor('.sho-summary-ring strong')

  assert.ok(ringDeclarations.includes('width: 64px'))
  assert.ok(ringDeclarations.some((declaration) => declaration.includes('conic-gradient')))
  assert.ok(valueDeclarations.includes('font-size: 16px'))
  assert.ok(valueDeclarations.includes('font-variant-numeric: tabular-nums'))
})

test('summary number value uses compact tabular digits in the right column', () => {
  const declarations = declarationsFor('.sho-summary-card-number .sho-summary-card-head strong')

  assert.ok(declarations.includes('grid-column: 2'))
  assert.ok(declarations.includes('grid-row: 1 / 3'))
  assert.ok(declarations.includes('font-size: 34px'))
  assert.ok(declarations.includes('font-variant-numeric: tabular-nums'))
})

test('summary number caption stays in the left column without a divider', () => {
  const declarations = declarationsFor('.sho-summary-card-number small')

  assert.ok(declarations.includes('grid-column: 1'))
  assert.ok(declarations.includes('grid-row: 2'))
  assert.ok(!declarations.some((declaration) => declaration.startsWith('border-top')))
})

test('score card label style does not override the health ring score', () => {
  const labelDeclarations = declarationsFor('.sho-score-card > div:last-child > span')

  assert.ok(labelDeclarations.includes('font-size: 12px'))
  assert.equal(css.includes('.sho-score-card span {'), false)
})

test('large health ring keeps a prominent center score', () => {
  const declarations = declarationsFor('.sho-health-ring-large')

  assert.ok(declarations.includes('font-size: 26px'))
  assert.ok(declarations.includes('font-variant-numeric: tabular-nums'))
})

test('system metric tiles reserve consistent vertical slots', () => {
  const declarations = declarationsFor('.sho-metric-tile')

  assert.ok(declarations.includes('grid-template-rows: 22px 118px minmax(34px, auto)'))
  assert.ok(declarations.includes('align-content: stretch'))
})

test('system metric visuals align value and chart areas', () => {
  const declarations = declarationsFor('.sho-metric-visual')

  assert.ok(declarations.includes('grid-template-rows: 42px 72px'))
  assert.ok(declarations.includes('min-height: 118px'))
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
