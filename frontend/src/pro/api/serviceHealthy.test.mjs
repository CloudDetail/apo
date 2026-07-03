import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = fs.readFileSync(path.join(__dirname, 'serviceHealthy.js'), 'utf8')

test('system health trend API defaults to historical snapshot judgement', () => {
  const start = source.indexOf('export const getSystemHealthTrendApi = (params) => {')
  const end = source.indexOf('export const getSystemHealthSystemDetailApi', start)
  const fnSource = source.slice(start, end)

  assert.ok(fnSource.includes("judgement: 'snapshot'"))
  assert.ok(fnSource.includes("...params"))
  assert.ok(fnSource.indexOf("judgement: 'snapshot'") < fnSource.indexOf('...params'))
})

test('system health scalable API helpers expose paged rows and batch trends', () => {
  assert.ok(source.includes('export const getSystemHealthSystemsApi = (params) => {'))
  assert.ok(source.includes("get('/api/system-health/systems', params)"))
  assert.ok(source.includes('export const getSystemHealthServicesApi = (groupId, params) => {'))
  assert.ok(source.includes('`/api/system-health/overview/${groupId}/services`'))
  assert.ok(source.includes('export const getSystemHealthPendingIssuesApi = (params) => {'))
  assert.ok(source.includes("get('/api/system-health/pending-issues', params)"))
  assert.ok(source.includes('export const getSystemHealthSystemsTrendApi = (data) => {'))
  assert.ok(source.includes("post('/api/system-health/trend/systems'"))
})
