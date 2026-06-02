/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))

test('request interceptor suppresses user-visible notifications for cancelled requests', () => {
  const requestSource = readFileSync(resolve(currentDir, 'request.js'), 'utf8')

  assert.match(requestSource, /const isRequestCanceled = \(error = \{\}\) =>/)
  assert.match(requestSource, /axios\.isCancel\?\.\(error\)/)
  assert.match(requestSource, /error\.code === 'ERR_CANCELED'/)
  assert.match(requestSource, /error\.name === 'AbortError'/)
  assert.match(requestSource, /if \(isRequestCanceled\(error\)\) \{[\s\S]*return Promise\.reject\(error\)[\s\S]*\}[\s\S]*if \(error\.response\)/)
  assert.doesNotMatch(requestSource, /if \(error\.response\)[\s\S]*else \{[\s\S]*message: error\.message,[\s\S]*isRequestCanceled/)
})
