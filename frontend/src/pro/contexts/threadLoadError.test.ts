import test from 'node:test'
import assert from 'node:assert/strict'

import { classifyThreadLoadError, formatThreadLoadErrorMessage } from './threadLoadError'

test('formats a thread load error with the localized prefix and original error message', () => {
  const message = formatThreadLoadErrorMessage({
    error: new Error('thread not found'),
    messages: {
      loadFailed: 'Failed to load conversation history',
      unknownError: 'Unknown error',
    },
  })

  assert.equal(message, 'Failed to load conversation history: thread not found')
})

test('falls back to the localized unknown-error copy when the thrown value is not an Error', () => {
  const message = formatThreadLoadErrorMessage({
    error: { message: '' },
    messages: {
      loadFailed: 'Failed to load conversation history',
      unknownError: 'Unknown error',
    },
  })

  assert.equal(message, 'Failed to load conversation history: Unknown error')
})

test('classifies explicit missing-thread failures as recoverable not found errors', () => {
  assert.equal(classifyThreadLoadError(new Error('thread not found')), 'not_found')
  assert.equal(classifyThreadLoadError(new Error('CONVERSATION_NOT_FOUND')), 'not_found')
})

test('does not classify transient load failures as recoverable not found errors', () => {
  assert.equal(classifyThreadLoadError(new Error('network timeout')), 'transient')
  assert.equal(classifyThreadLoadError(new Error('Request failed with status code 500')), 'transient')
})
