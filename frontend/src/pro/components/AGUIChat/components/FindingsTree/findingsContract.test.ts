import test from 'node:test'
import assert from 'node:assert/strict'

import { getFindingsDetailSections, toFindingsNodeData } from './findingsContract'
import type { FindingsNodePayload } from '../../../../api/conversation'

test('maps reportLinks from payload to selected node data', () => {
  const payload: FindingsNodePayload = {
    nodeId: 'code_cpu_check',
    title: 'Code CPU',
    nodeKind: 'evidence',
    status: 'abnormal',
    parentId: null,
    reportLinks: [
      {
        type: 'slow',
        traceId: 'trace-1',
        spanId: 'span-1',
        url: '/#/cause/report/trace-1/span-1',
      },
    ],
  }

  assert.deepEqual(toFindingsNodeData(payload).reportLinks, payload.reportLinks)
})

test('hydrating nodes do not expose normal detail sections', () => {
  const payload: FindingsNodePayload = {
    nodeId: 'code_cpu_check',
    title: 'Code CPU',
    nodeKind: 'evidence',
    status: 'normal',
    statusColor: 'normal',
    parentId: null,
    summaryConclusion: 'CPU is normal',
    content: 'Normal detail should wait for hydration',
    metrics: {
      abnormalMetrics: [],
      normalMetrics: [
        {
          title: 'CPU',
          unit: '%',
          timeseries: [],
        },
      ],
    },
    hydrationStatus: 'running',
  }

  assert.deepEqual(getFindingsDetailSections(toFindingsNodeData(payload)), {
    loading: true,
    summaryConclusion: false,
    content: false,
    abnormalMetrics: false,
    normalMetrics: false,
    table: false,
    logs: false,
    traces: false,
    unsupported: false,
  })
})
