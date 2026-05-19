import test from 'node:test'
import assert from 'node:assert/strict'

import { buildRelatedReportGroups, getRelatedReportLabel } from './RelatedReports'

const messages: Record<string, string> = {
  'evidencePanel.viewSlowReport': '查看慢调用报告',
  'evidencePanel.viewErrorReport': '查看错误报告',
}

const t = (key: string) => messages[key] ?? key

test('uses AgentSRE report title before default translated report label', () => {
  const label = getRelatedReportLabel({
    link: {
      type: 'error',
      title: '数据库错误报告',
      traceId: 'trace-1',
      url: '/#/cause/error-report/trace-1',
    },
    typeTotal: 2,
    typeIndex: 1,
    t,
  })

  assert.equal(label, '数据库错误报告')
})

test('falls back to indexed translated label when report title is absent', () => {
  const label = getRelatedReportLabel({
    link: {
      type: 'error',
      traceId: 'trace-1',
      url: '/#/cause/error-report/trace-1',
    },
    typeTotal: 2,
    typeIndex: 1,
    t,
  })

  assert.equal(label, '查看错误报告 1')
})

test('distinguishes duplicate titled reports inside the same group', () => {
  const label = getRelatedReportLabel({
    link: {
      type: 'error',
      title: 'ts-travel-plan-service error report',
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
    },
    typeTotal: 3,
    typeIndex: 2,
    groupTotal: 2,
    groupIndex: 2,
    t,
  })

  assert.equal(label, 'ts-travel-plan-service error report 2')
})

test('groups error reports by service and keeps all links visible with representative first', () => {
  const groups = buildRelatedReportGroups([
    {
      type: 'error',
      title: 'ts-travel-plan-service error report',
      serviceName: 'ts-travel-plan-service',
      errorGroupKey: 'TimeoutException',
      representative: false,
      traceId: 'trace-secondary',
      url: '/#/cause/error-report/trace-secondary',
    },
    {
      type: 'error',
      title: 'ts-travel-plan-service error report',
      serviceName: 'ts-travel-plan-service',
      errorGroupKey: 'TimeoutException',
      representative: true,
      traceId: 'trace-representative',
      url: '/#/cause/error-report/trace-representative',
    },
    {
      type: 'error',
      title: 'ts-travel-service error report',
      serviceName: 'ts-travel-service',
      representative: true,
      traceId: 'trace-travel',
      url: '/#/cause/error-report/trace-travel',
    },
  ])

  assert.equal(groups.length, 2)
  assert.deepEqual(
    groups.map((group) => group.title),
    ['ts-travel-plan-service / TimeoutException', 'ts-travel-service'],
  )
  assert.deepEqual(
    groups[0].links.map((link) => link.traceId),
    ['trace-representative', 'trace-secondary'],
  )
  assert.equal(groups[0].links.length, 2)
})

test('uses producer groupCount when it is larger than currently visible links', () => {
  const groups = buildRelatedReportGroups([
    {
      type: 'error',
      title: 'ts-travel-plan-service error report',
      serviceName: 'ts-travel-plan-service',
      groupCount: 5,
      traceId: 'trace-representative',
      url: '/#/cause/error-report/trace-representative',
    },
    {
      type: 'error',
      title: 'ts-travel-plan-service error report',
      serviceName: 'ts-travel-plan-service',
      groupCount: 3,
      traceId: 'trace-secondary',
      url: '/#/cause/error-report/trace-secondary',
    },
  ])

  assert.equal(groups[0].displayCount, 5)
  assert.equal(groups[0].links.length, 2)
})

test('does not append representative text to report labels', () => {
  const label = getRelatedReportLabel({
    link: {
      type: 'error',
      title: 'ts-travel-plan-service error report',
      representative: true,
      traceId: 'trace-representative',
      url: '/#/cause/error-report/trace-representative',
    },
    typeTotal: 2,
    typeIndex: 1,
    groupTotal: 2,
    groupIndex: 1,
    t,
  })

  assert.equal(label, 'ts-travel-plan-service error report 1')
})
