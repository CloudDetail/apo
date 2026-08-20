// @ts-nocheck
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const currentDir = path.dirname(fileURLToPath(import.meta.url))
const routesSource = fs.readFileSync(path.resolve(currentDir, '../../../../routes.js'), 'utf8')

test('work capability pages do not use route permission guards', () => {
  for (const page of ['CapabilityListPage', 'CapabilityDetailPage', 'CapabilityVersionPage']) {
    assert.match(routesSource, new RegExp(`const ${page} = React\\.lazy\\(`))
    assert.doesNotMatch(routesSource, new RegExp(`const ${page} = AuthRouter\\(`))
  }
})
