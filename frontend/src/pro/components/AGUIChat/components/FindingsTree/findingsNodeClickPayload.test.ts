import test from 'node:test'
import assert from 'node:assert/strict'

import { buildFindingsNodeClickPayload } from './FindingsNode'
import type { FindingsNodeData } from './findingsContract'

test('node click payload preserves report links after switching nodes', () => {
  const reportLinks: FindingsNodeData['reportLinks'] = [
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      mutatedType: 'cpu_time',
      url: '/#/cause/report/trace-1/span-1?mutatedType=cpu_time',
    },
  ]

  const clicked = buildFindingsNodeClickPayload(
    {
      nodeId: 'ebpf_check_cpu',
      title: 'CPU检查',
      status: 'abnormal',
      visualStatus: 'abnormal',
      nodeKind: 'evidence',
      reportLinks,
      active: true,
      disabled: false,
      onClick: () => {},
    },
    'abnormal',
  )

  assert.deepEqual(clicked.reportLinks, reportLinks)
  assert.equal(clicked.onClick, undefined)
  assert.equal(clicked.active, undefined)
  assert.equal(clicked.disabled, undefined)
})
