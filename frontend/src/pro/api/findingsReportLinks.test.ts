import test from 'node:test'
import assert from 'node:assert/strict'

import { normalizeFindingsReportLinks } from './findingsReportLinks'

test('accepts exact slow report hash route with traceId spanId and mutatedType', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'slow',
      title: 'CPU slow report',
      traceId: 'trace-1',
      spanId: 'span-1',
      mutatedType: 'cpu_time',
      url: '/#/cause/report/trace-1/span-1?mutatedType=cpu_time',
      source: 'anomaly-span-list',
    },
  ])

  assert.deepEqual(links, [
    {
      type: 'slow',
      title: 'CPU slow report',
      traceId: 'trace-1',
      spanId: 'span-1',
      mutatedType: 'cpu_time',
      url: '/#/cause/report/trace-1/span-1?mutatedType=cpu_time',
      source: 'anomaly-span-list',
    },
  ])
})

test('accepts slow report mutatedType from url query when field is absent', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      url: '/#/cause/report/trace-1/span-1?mutatedType=cpu_time',
    },
  ])

  assert.deepEqual(links, [
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      mutatedType: 'cpu_time',
      url: '/#/cause/report/trace-1/span-1?mutatedType=cpu_time',
    },
  ])
})

test('accepts exact error report hash route with traceId', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
    },
  ])

  assert.deepEqual(links, [
    {
      type: 'error',
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
    },
  ])
})

test('preserves error report grouping metadata from producer', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      title: 'ts-travel-plan-service error report',
      serviceName: 'ts-travel-plan-service',
      errorGroupKey: 'TimeoutException',
      representative: true,
      groupCount: 2,
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
      source: 'dependency-red-error-trace',
    },
  ])

  assert.deepEqual(links, [
    {
      type: 'error',
      title: 'ts-travel-plan-service error report',
      serviceName: 'ts-travel-plan-service',
      errorGroupKey: 'TimeoutException',
      representative: true,
      groupCount: 2,
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
      source: 'dependency-red-error-trace',
    },
  ])
})

test('rejects slow report links without spanId', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'slow',
      traceId: 'trace-1',
      url: '/#/cause/report/trace-1/',
    },
  ])

  assert.equal(links, undefined)
})

test('rejects error report links without producer supplied url', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      traceId: 'trace-1',
    },
  ])

  assert.equal(links, undefined)
})

test('rejects slow report links without producer supplied url', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
    },
  ])

  assert.equal(links, undefined)
})

test('rejects slow report links with query params when mutatedType is absent', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      url: '/#/cause/report/trace-1/span-1?x=1',
    },
  ])

  assert.equal(links, undefined)
})

test('rejects slow report links with extra query params when mutatedType is present', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      mutatedType: 'cpu_time',
      url: '/#/cause/report/trace-1/span-1?mutatedType=cpu_time&x=1',
    },
  ])

  assert.equal(links, undefined)
})

test('rejects error report links with query params', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      traceId: 'trace-1',
      url: '/#/cause/error-report/trace-1?x=1',
    },
  ])

  assert.equal(links, undefined)
})

test('rejects external URLs and routes that only contain cause report text', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      traceId: 'trace-1',
      url: 'https://example.com/#/cause/error-report/trace-1',
    },
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      url: '/#/foo/cause/report/trace-1/span-1',
    },
  ])

  assert.equal(links, undefined)
})

test('orders slow report links before error report links', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
    },
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      url: '/#/cause/report/trace-1/span-1',
    },
  ])

  assert.deepEqual(
    links?.map((link) => link.type),
    ['slow', 'error'],
  )
})
