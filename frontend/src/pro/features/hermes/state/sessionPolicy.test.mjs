// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { shouldRestoreInBackground } from './sessionPolicy.ts'

test('only live workflow states restore in the background', () => {
  assert.equal(shouldRestoreInBackground('running'), true)
  assert.equal(shouldRestoreInBackground('waiting_confirmation'), true)
  assert.equal(shouldRestoreInBackground('ready'), false)
  assert.equal(shouldRestoreInBackground('failed'), false)
  assert.equal(shouldRestoreInBackground('interrupted'), false)
})
