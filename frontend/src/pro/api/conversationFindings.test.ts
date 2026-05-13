import test from 'node:test'
import assert from 'node:assert/strict'

import { normalizeFindingsNode } from './findingsDataNormalizer'

test('normalizes raw trace rows into NodeTraces call groups', () => {
  const node = normalizeFindingsNode({
    nodeId: 'db_traces_check',
    title: 'DB traces',
    nodeKind: 'evidence',
    status: 'normal',
    parentId: 'db_query_check',
    traces: [
      {
        serviceName: 'ts-order-service',
        operationName: 'SELECT Order',
        duration: 100,
        isError: false,
      },
      {
        serviceName: 'ts-order-service',
        operationName: 'SELECT Order',
        duration: 300,
        isError: true,
      },
      {
        serviceName: 'ts-price-service',
        operationName: 'SELECT price_config',
        duration: 200,
        isError: false,
      },
    ],
  })

  assert.deepEqual(node?.traces, [
    {
      service: 'ts-order-service',
      calls: [
        {
          target: 'SELECT Order',
          type: 'database',
          count: 2,
          avgDuration: 200,
          maxDuration: 300,
          errorCount: 1,
          operations: [
            {
              name: 'SELECT Order',
              count: 2,
              avgDuration: 200,
            },
          ],
        },
      ],
    },
    {
      service: 'ts-price-service',
      calls: [
        {
          target: 'SELECT price_config',
          type: 'database',
          count: 1,
          avgDuration: 200,
          maxDuration: 200,
          errorCount: 0,
          operations: [
            {
              name: 'SELECT price_config',
              count: 1,
              avgDuration: 200,
            },
          ],
        },
      ],
    },
  ])
})
