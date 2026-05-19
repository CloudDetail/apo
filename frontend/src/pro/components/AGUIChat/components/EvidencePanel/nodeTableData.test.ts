import test from 'node:test'
import assert from 'node:assert/strict'

import { buildNodeTableDataSource } from './nodeTableData'

test('preserves explicit row keys and falls back to index for table rows', () => {
  const rows = buildNodeTableDataSource([
    { key: 'pod-1', name: 'pod-a' },
    { name: 'pod-b' },
  ])

  assert.deepEqual(rows, [
    { key: 'pod-1', name: 'pod-a', _rowKey: 'pod-1' },
    { name: 'pod-b', _rowKey: 1 },
  ])
})
