import test from 'node:test'
import assert from 'node:assert/strict'

import { describePendingIssue } from './pendingIssueDescription.mjs'

test('describePendingIssue describes red resource issues as current state only', () => {
  assert.equal(
    describePendingIssue({
      systemName: 'train-ticket',
      serviceName: 'ts-gateway-service',
      metricKey: 'resource',
      metricLabel: '资源水位',
      status: 'red',
      value: '18.90%',
      note: '当前内存资源水位；资源副本健康存在风险',
    }),
    'train-ticket 中 ts-gateway-service 的资源水位当前值为 18.90%，已处于严重异常状态；当前内存资源水位；资源副本健康存在风险。',
  )
})

test('describePendingIssue describes lock wait without remediation guidance', () => {
  assert.equal(
    describePendingIssue({
      systemName: '订单系统',
      serviceName: 'order-api',
      metricKey: 'lockCost',
      metricLabel: '锁等待',
      status: 'yellow',
      value: '12ms',
      note: '由 futex 锁等待耗时按请求数均摊计算',
    }),
    '订单系统 中 order-api 的锁等待当前值为 12ms，已进入一般异常状态；由 futex 锁等待耗时按请求数均摊计算。',
  )
})

test('describePendingIssue falls back to note when metric type is unknown', () => {
  assert.equal(
    describePendingIssue({
      systemName: '用户管理系统',
      serviceName: 'user-api',
      metricKey: 'customMetric',
      metricLabel: '自定义指标',
      status: 'yellow',
      value: '--',
      note: '指标接近异常阈值',
    }),
    '用户管理系统 中 user-api 的自定义指标已进入一般异常状态；指标接近异常阈值。',
  )
})
