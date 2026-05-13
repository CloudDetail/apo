import test from 'node:test'
import assert from 'node:assert/strict'

import { getFindingsNodePresentation } from './findingsNodePresentation'

test('running nodes use loading presentation and loading badge', () => {
  const presentation = getFindingsNodePresentation({
    status: 'normal',
    nodeKind: 'evidence',
    hydrationStatus: 'running',
  })

  assert.equal(presentation.themeKey, 'loading')
  assert.equal(presentation.resolvedStatus, 'normal')
  assert.equal(presentation.badgeLabelKey, 'findingsTree.hydrationLoading')
})

test('failed nodes keep warning presentation without changing business status', () => {
  const presentation = getFindingsNodePresentation({
    status: 'normal',
    nodeKind: 'evidence',
    hydrationStatus: 'failed',
  })

  assert.equal(presentation.themeKey, 'warning')
  assert.equal(presentation.resolvedStatus, 'warning')
  assert.equal(presentation.badgeLabelKey, 'findingsTree.hydrationFailed')
})
