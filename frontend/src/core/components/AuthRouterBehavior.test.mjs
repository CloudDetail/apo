/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const authRouterSource = fs.readFileSync(path.join(currentDir, 'AuthRouter.tsx'), 'utf8')

test('auth router waits for permission result before allowing a protected route', () => {
  assert.match(authRouterSource, /useState\(null\)/)
  assert.match(authRouterSource, /setIsAllowed\(hasPermission\)/)
  assert.doesNotMatch(authRouterSource, /\/\/\s*setIsAllowed\(hasPermission\)/)
})
