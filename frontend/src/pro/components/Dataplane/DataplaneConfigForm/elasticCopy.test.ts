import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const frontendRoot = process.cwd()

const readFrontendSource = (relativePath: string) => {
  return fs.readFileSync(path.join(frontendRoot, 'src', relativePath), 'utf8')
}

test('elastic form keeps onboarding guidance out of the main form', () => {
  const configFormSource = readFrontendSource('pro/components/Dataplane/DataplaneConfigForm/index.tsx')

  assert.doesNotMatch(
    configFormSource,
    /elasticLogOnlyMode/,
    'Elastic main flow guidance should not render as an inline form alert',
  )
})

test('non-elastic dataplanes keep the original capabilitySpec behavior', () => {
  const configFormSource = readFrontendSource('pro/components/Dataplane/DataplaneConfigForm/index.tsx')

  assert.match(
    configFormSource,
    /const capability = isElastic\s*\?\s*ELASTIC_LOG_ONLY_CAPABILITIES\s*:\s*capabilitySpec \|\| \[\]/,
    'Only Elastic should override capabilities; non-Elastic dataplanes should keep using capabilitySpec',
  )
  assert.doesNotMatch(
    configFormSource,
    /isDataplaneItem\(dataplaneParams\)\s*\?\s*dataplaneParams\.capability/,
    'Non-Elastic dataplanes should not switch to instance capability in this Elastic-only change',
  )
})

test('elastic setup guide explains the simplified main flow and defaults', () => {
  const guideSource = readFrontendSource('pro/components/Dataplane/DataplaneGuide/guideMd.ts')

  assert.match(
    guideSource,
    /主流程只需要填写\s+\*\*名称\*\*、\*\*URL\*\* 和 \*\*API Key\*\*/,
    'Chinese guide should state the simplified Elastic main flow',
  )
  assert.match(
    guideSource,
    /Only fill in \*\*Name\*\*, \*\*URL\*\*, and \*\*API Key\*\*/,
    'English guide should state the simplified Elastic main flow',
  )
  assert.match(
    guideSource,
    /authType=apiKey/,
    'Guide should document the default auth type',
  )
  assert.match(
    guideSource,
    /logIndexPattern=logs-\*/,
    'Guide should document the default log index pattern',
  )
})

test('elastic dataplane uses the existing Elastic logo asset', () => {
  const constantsSource = readFrontendSource('pro/constants.ts')

  assert.match(
    constantsSource,
    /import elastic from 'src\/core\/assets\/dataCollector\/elastic\.svg'/,
    'Elastic should reuse the existing Elastic SVG asset',
  )
  assert.match(
    constantsSource,
    /'elastic': elastic/,
    'Elastic dataplane type should map to the Elastic logo',
  )
})
