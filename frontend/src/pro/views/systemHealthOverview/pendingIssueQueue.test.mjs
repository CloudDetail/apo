import test from 'node:test'
import assert from 'node:assert/strict'

import { buildPendingIssueQueueRows } from './pendingIssueQueue.mjs'

test('buildPendingIssueQueueRows aggregates pending issues by service and metric', () => {
  const rows = buildPendingIssueQueueRows([
    {
      id: 'yellow-latency-1',
      groupId: 1001,
      systemName: '订单系统',
      serviceName: 'order-api',
      metricKey: 'latency',
      metricLabel: '响应延时',
      status: 'yellow',
      value: '820ms',
      note: 'P95 延时升高',
    },
    {
      id: 'red-latency-2',
      groupId: 1002,
      systemName: '支付系统',
      serviceName: 'order-api',
      metricKey: 'latency',
      metricLabel: '响应延时',
      status: 'red',
      value: '960ms',
      note: 'P99 延时异常',
    },
    {
      id: 'red-resource',
      groupId: 1003,
      systemName: '营销系统',
      serviceName: 'campaign-worker',
      metricKey: 'resource',
      metricLabel: '资源水位',
      status: 'red',
      value: '91%',
      note: '内存水位异常',
    },
  ])

  assert.equal(rows.length, 2)
  assert.deepEqual(
    rows.map((row) => row.queueId),
    ['campaign-worker:resource', 'order-api:latency'],
  )
  assert.equal(rows[1].status, 'red')
  assert.equal(rows[1].issueCount, 2)
  assert.equal(rows[1].value, '960ms、820ms')
  assert.equal(rows[1].note, 'P99 延时异常')
  assert.equal(rows[1].systemName, '支付系统、订单系统')
  assert.equal(rows[1].groupId, 1002)
})

test('buildPendingIssueQueueRows labels replica-only resource issues as availability', () => {
  const rows = buildPendingIssueQueueRows([
    {
      id: 'red-resource-replica',
      groupId: 1001,
      systemName: '订单系统',
      serviceName: 'ts-route-plan-service',
      metricKey: 'resource',
      metricLabel: '资源水位',
      status: 'red',
      value: '异常',
      note: '应用实例数小于2，存在服务不可用风险',
    },
  ])

  assert.equal(rows[0].metricLabel, '资源可用性')
})
