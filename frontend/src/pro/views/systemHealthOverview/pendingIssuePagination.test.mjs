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
    tableSource.includes('<td colSpan={5}>') && tableSource.includes('<PendingIssueAnalysis'),
    'analysis row should span the remaining five columns',
  )
})

test('pending issue system column opens the corresponding system detail', () => {
  const systemCellStart = pendingIssueQueueSource.indexOf(
    '<td className="sho-pending-system-cell">',
  )
  assert.notEqual(systemCellStart, -1, 'pending issue system cell should exist')

  const systemCellEnd = pendingIssueQueueSource.indexOf('</td>', systemCellStart)
  assert.notEqual(systemCellEnd, -1, 'pending issue system cell should close')

  const systemCellSource = pendingIssueQueueSource.slice(systemCellStart, systemCellEnd)

  assert.ok(
    pendingIssueQueueSource.includes('onOpenSystemDetail,'),
    'pending issue queue should accept a system detail callback',
  )
  assert.ok(systemCellSource.includes('<button'), 'system name should be a button')
  assert.ok(
    pendingIssueQueueSource.includes('const pendingIssueSystemTargets = (issue = {}) =>'),
    'pending issue queue should derive one clickable target per system',
  )
  assert.ok(
    systemCellSource.includes('pendingIssueSystemTargets(issue).map((system, index) => ('),
    'pending issue system cell should render each system target separately',
  )
  assert.ok(systemCellSource.includes('{index > 0 && ('), 'multiple systems should be separated')
  assert.ok(systemCellSource.includes('className="sho-pending-system-separator"'))
  assert.ok(systemCellSource.includes('、'))
  assert.ok(systemCellSource.includes('type="button"'))
  assert.ok(systemCellSource.includes('className="sho-pending-system"'))
  assert.ok(systemCellSource.includes('title={system.name}'))
  assert.ok(systemCellSource.includes('aria-label={`查看${system.name}健康度详情`}'))
  assert.ok(systemCellSource.includes('onOpenSystemDetail(pendingIssueSystemDetailTarget(system))'))
  assert.ok(
    pendingIssueQueueSource.includes('onOpenSystemDetail: PropTypes.func.isRequired'),
    'system detail callback should be required',
  )
  assert.ok(
    viewSource.includes('const openPendingIssueSystemDetail = (system) => {'),
    'overview page should define a pending issue detail opener',
  )
  assert.ok(
    viewSource.includes('openDetail(system, { refreshHeader: true })'),
    'pending issue detail opener should refresh the detail header',
  )
  assert.ok(
    viewSource.includes('onOpenSystemDetail={openPendingIssueSystemDetail}'),
    'pending issue queue should receive the detail opener',
  )
})
