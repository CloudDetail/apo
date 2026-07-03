import test from 'node:test'
import assert from 'node:assert/strict'

import { buildPendingIssueAnalysisKey } from './pendingIssueAnalysis.mjs'

test('pending issue analysis key changes when current value or note changes', () => {
  const baseIssue = {
    groupId: 1001,
    serviceName: 'order-api',
    metricKey: 'latency',
    status: 'yellow',
    value: '380ms',
    note: '响应延时升高',
  }
  const params = { relativeTime: '15m' }

  assert.notEqual(
    buildPendingIssueAnalysisKey(baseIssue, params),
    buildPendingIssueAnalysisKey({ ...baseIssue, value: '420ms' }, params),
  )
  assert.notEqual(
    buildPendingIssueAnalysisKey(baseIssue, params),
    buildPendingIssueAnalysisKey({ ...baseIssue, note: '响应延时持续升高' }, params),
  )
})
