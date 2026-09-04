import assert from 'node:assert/strict'
import test from 'node:test'

import { dataIntegrationSaveMode } from './saveMode.ts'

test('new integrations use the create API even after the required id is entered', () => {
  assert.equal(dataIntegrationSaveMode(undefined), 'create')
  assert.equal(dataIntegrationSaveMode({}), 'create')
})

test('existing integrations use the update API', () => {
  assert.equal(dataIntegrationSaveMode({ id: 'cluster-69' }), 'update')
})
