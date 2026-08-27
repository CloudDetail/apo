import test from 'node:test'
import assert from 'node:assert/strict'

import {
  getPreloadRuntimeSection,
  parsePreloadConfig,
  setPreloadRuntimeEnabled,
} from './preloadConfig.ts'

const config = `[java]
JAVA_TOOL_OPTIONS=-javaagent:agent.jar

[nodejs]
NODE_OPTIONS=--require agent.js

[python]
PYTHONPATH=/opt/apo/python

[dotnet]
CORECLR_ENABLE_PROFILING=1

[custom]
keep=this

[basename-blacklist]
full=["java -version"]
`

test('detects enabled runtimes from the sections consumed by preload', () => {
  assert.deepEqual(parsePreloadConfig(config), {
    enabled: { java: true, python: true, nodejs: true, dotnet: true },
    issues: [],
  })
})

test('disables and restores a runtime without changing its configuration', () => {
  const disabled = setPreloadRuntimeEnabled(config, 'python', false)

  assert.equal(disabled.issues.length, 0)
  assert.equal(parsePreloadConfig(disabled.content).enabled.python, false)
  assert.doesNotMatch(disabled.content, /^\[python\]$/m)
  assert.match(disabled.content, /^\[custom\]$/m)
  assert.match(getPreloadRuntimeSection(disabled.content, 'python').content, /^\[python\]$/m)

  const restored = setPreloadRuntimeEnabled(disabled.content, 'python', true)
  assert.equal(restored.content, config)
})

test('adds a package-compatible default section when enabling an absent runtime', () => {
  const minimal = `[basename-blacklist]\nfull=[]\n`
  const result = setPreloadRuntimeEnabled(minimal, 'nodejs', true)

  assert.match(result.content, /^\[nodejs\]$/m)
  assert.match(
    result.content,
    /NODE_OPTIONS = --require \/etc\/apo\/instrumentations\/nodejs\/autoinstrumentation\.js/,
  )
  assert.ok(result.content.indexOf('[nodejs]') < result.content.indexOf('[basename-blacklist]'))
})

test('preserves ambiguous input and reports duplicate runtime sections', () => {
  const duplicated = `[java]\na=1\n[java]\na=2\n`
  const result = setPreloadRuntimeEnabled(duplicated, 'java', false)

  assert.equal(result.content, duplicated)
  assert.match(result.issues.join(','), /重复 section/)
})
