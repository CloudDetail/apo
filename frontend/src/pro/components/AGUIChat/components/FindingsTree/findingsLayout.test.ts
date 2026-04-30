import test from 'node:test'
import assert from 'node:assert/strict'

import type { FindingsTopologyData } from './findingsContract'
import { buildFindingsLayout } from './findingsLayout'

const buildTopology = (withReservedChildren: boolean): FindingsTopologyData => {
  const nodes = [
    { id: 'root', type: 'findingsNode' as const, data: { nodeId: 'root', title: 'root', status: 'normal', visualStatus: 'normal', nodeKind: 'summary' as const } },
    { id: 'dependency_check', type: 'findingsNode' as const, data: { nodeId: 'dependency_check', title: 'dependency_check', status: 'normal', visualStatus: 'normal', nodeKind: 'evidence' as const } },
    { id: 'db_query_check', type: 'findingsNode' as const, data: { nodeId: 'db_query_check', title: 'db_query_check', status: 'normal', visualStatus: 'normal', nodeKind: 'evidence' as const } },
    { id: 'code_check', type: 'findingsNode' as const, data: { nodeId: 'code_check', title: 'code_check', status: 'normal', visualStatus: 'normal', nodeKind: 'evidence' as const } },
    { id: 'code_child_1', type: 'findingsNode' as const, data: { nodeId: 'code_child_1', title: 'code_child_1', status: 'normal', visualStatus: 'normal', nodeKind: 'evidence' as const } },
    { id: 'code_child_2', type: 'findingsNode' as const, data: { nodeId: 'code_child_2', title: 'code_child_2', status: 'normal', visualStatus: 'normal', nodeKind: 'evidence' as const } },
  ]

  if (withReservedChildren) {
    nodes.push(
      { id: 'dependency_child', type: 'findingsNode' as const, data: { nodeId: 'dependency_child', title: 'dependency_child', status: 'normal', visualStatus: 'normal', nodeKind: 'evidence' as const } },
      { id: 'db_query_child', type: 'findingsNode' as const, data: { nodeId: 'db_query_child', title: 'db_query_child', status: 'normal', visualStatus: 'normal', nodeKind: 'evidence' as const } },
    )
  }

  const edges = [
    { id: 'root-dependency', source: 'root', target: 'dependency_check' },
    { id: 'root-db-query', source: 'root', target: 'db_query_check' },
    { id: 'root-code', source: 'root', target: 'code_check' },
    { id: 'code-child-1', source: 'code_check', target: 'code_child_1' },
    { id: 'code-child-2', source: 'code_check', target: 'code_child_2' },
  ]

  if (withReservedChildren) {
    edges.push(
      { id: 'dependency-child', source: 'dependency_check', target: 'dependency_child' },
      { id: 'db-query-child', source: 'db_query_check', target: 'db_query_child' },
    )
  }

  return { nodes, edges }
}

const getNodeY = (nodes: ReturnType<typeof buildFindingsLayout>['nodes'], id: string) => {
  const match = nodes.find((node) => node.id === id)
  assert.ok(match, `expected node ${id} to exist`)
  return match.position.y
}

test('preserves sibling subtree slots when dependency and db_query checks have no children', () => {
  const fullLayout = buildFindingsLayout(buildTopology(true))
  const sparseLayout = buildFindingsLayout(buildTopology(false))

  assert.equal(getNodeY(sparseLayout.nodes, 'code_check'), getNodeY(fullLayout.nodes, 'code_check'))
  assert.equal(getNodeY(sparseLayout.nodes, 'code_child_1'), getNodeY(fullLayout.nodes, 'code_child_1'))
  assert.equal(getNodeY(sparseLayout.nodes, 'code_child_2'), getNodeY(fullLayout.nodes, 'code_child_2'))
})
