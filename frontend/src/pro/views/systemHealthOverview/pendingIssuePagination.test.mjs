import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

import { buildPendingIssuesFromSystems, paginatePendingIssues } from './pendingIssuePagination.mjs'

const sourceFor = (path) => readFileSync(new URL(path, import.meta.url), 'utf8')

const viewSource = sourceFor('./index.jsx')
const pendingIssueQueueSource = sourceFor('./components/PendingIssueQueue.jsx')

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
  assert.ok(
    pendingIssueQueueSource.includes('Pagination'),
    'pending issue queue should render AntD Pagination',
  )
  assert.ok(
    viewSource.includes('pendingIssueTotal'),
    'page should track pending issue total from the server response',
  )
  assert.ok(
    viewSource.includes('setPendingIssuePage'),
    'page should update pending issue page when pagination changes',
  )
  assert.ok(
    viewSource.includes('getSystemHealthPendingIssuesApi'),
    'page should use the server-side pending issue endpoint',
  )
  assert.ok(
    viewSource.includes('pendingIssueLoading'),
    'pending issue pagination should have independent loading state',
  )
  assert.equal(
    viewSource.includes('buildPendingIssuesFromSystems(systems)'),
    false,
    'pending issue pagination must not derive issues from loaded overview system rows',
  )
  assert.equal(
    viewSource.includes('paginatePendingIssues(allPendingIssues'),
    false,
    'pending issue pagination must not slice local overview-derived issues',
  )
})

test('system health pending issues request only asks for abnormal issues', () => {
  const requestStart = viewSource.indexOf('getSystemHealthPendingIssuesApi({')
  assert.notEqual(requestStart, -1, 'pending issue API request should exist')

  const requestEnd = viewSource.indexOf('})', requestStart)
  assert.notEqual(requestEnd, -1, 'pending issue API request should close')

  const requestSource = viewSource.slice(requestStart, requestEnd)

  assert.ok(
    requestSource.includes("status: 'abnormal'"),
    'pending issue count and table should request only abnormal issues',
  )
  assert.equal(
    requestSource.includes("status: 'all'"),
    false,
    'pending issue request must not include general or attention issues',
  )
})

test('pending issue table uses AI assistant column without a separate status column', () => {
  const tableStart = pendingIssueQueueSource.indexOf('<table className="sho-pending-table">')
  assert.notEqual(tableStart, -1, 'pending issue table should exist')

  const tableEnd = pendingIssueQueueSource.indexOf('</table>', tableStart)
  assert.notEqual(tableEnd, -1, 'pending issue table should close')

  const tableSource = pendingIssueQueueSource.slice(tableStart, tableEnd)

  assert.ok(tableSource.includes('<th>AI助手</th>'), 'action column title should be AI助手')
  assert.equal(tableSource.includes('<th>状态</th>'), false, 'status header should be removed')
  assert.equal(
    tableSource.includes('sho-pending-col-status'),
    false,
    'status colgroup entry should be removed',
  )
  assert.equal(
    tableSource.includes('statusLabel'),
    false,
    'status label rendering should be removed from pending issue rows',
  )
  assert.ok(
    tableSource.includes('<td colSpan={5}>{analysisNode}</td>'),
    'analysis row should span the remaining five columns',
  )
})
