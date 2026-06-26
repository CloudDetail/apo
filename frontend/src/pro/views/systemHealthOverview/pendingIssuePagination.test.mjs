import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

import { buildPendingIssuesFromSystems, paginatePendingIssues } from './pendingIssuePagination.mjs'

const viewSource = readFileSync(new URL('./index.jsx', import.meta.url), 'utf8')

test('buildPendingIssuesFromSystems derives sorted issues from overview systems', () => {
  const issues = buildPendingIssuesFromSystems([
    {
      groupId: 202,
      name: '支付系统',
      score: 72,
      services: [
        {
          name: 'payment-api',
          score: 65,
          metrics: {
            latency: 'yellow',
            errorRate: 'green',
          },
          metricDetails: {
            latency: { value: '420ms', note: '响应延时升高' },
          },
        },
      ],
    },
    {
      groupId: 101,
      name: '订单系统',
      score: 38,
      services: [
        {
          name: 'order-api',
          score: 35,
          metrics: {
            errorRate: 'red',
            latency: 'yellow',
          },
          metricDetails: {
            errorRate: { value: '5.70%', note: '错误率超过异常阈值' },
            latency: { value: '380ms', note: '响应延时升高' },
          },
        },
      ],
    },
  ])

  assert.equal(issues.length, 3)
  assert.deepEqual(
    issues.map((issue) => issue.id),
    ['101:order-api:errorRate', '101:order-api:latency', '202:payment-api:latency'],
  )
  assert.equal(issues[0].metricLabel, '错误率')
  assert.equal(issues[0].value, '5.70%')
  assert.equal(issues[0].note, '错误率超过异常阈值')
})

test('paginatePendingIssues slices local issues without request params', () => {
  const issues = Array.from({ length: 25 }, (_, index) => ({ id: `issue-${index + 1}` }))

  const page = paginatePendingIssues(issues, 2, 10)

  assert.equal(page.total, 25)
  assert.equal(page.page, 2)
  assert.equal(page.pageSize, 10)
  assert.deepEqual(
    page.issues.map((issue) => issue.id),
    [
      'issue-11',
      'issue-12',
      'issue-13',
      'issue-14',
      'issue-15',
      'issue-16',
      'issue-17',
      'issue-18',
      'issue-19',
      'issue-20',
    ],
  )
})

test('system health page renders pending issue pagination controls', () => {
  assert.ok(viewSource.includes('Pagination'), 'page should import and render AntD Pagination')
  assert.ok(
    viewSource.includes('pendingIssuePageData.total'),
    'page should track pending issue total from local pagination',
  )
  assert.ok(
    viewSource.includes('setPendingIssuePage'),
    'page should update pending issue page when pagination changes',
  )
  assert.ok(
    viewSource.includes('skipNewAbnormalMetrics: true'),
    'overview loading should skip expensive new abnormal metric calculation',
  )
  assert.ok(
    !viewSource.includes('getSystemHealthOverviewApi(overviewQueryParams)'),
    'pending issue pagination must not re-fetch the full overview',
  )
})
