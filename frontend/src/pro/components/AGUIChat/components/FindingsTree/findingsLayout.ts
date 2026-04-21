import { hierarchy, tree as d3tree } from 'd3-hierarchy'

import type { FindingsNodeData, FindingsTopologyData } from './findingsContract'

const NODE_WIDTH = 220
const NODE_HEIGHT = 72
const LAYER_GAP = 80
const SIBLING_GAP = 20

const RESERVED_CHILD_SLOT_COUNTS: Record<string, number> = {
  dependency_check: 1,
  db_query_check: 1,
}

type LayoutNode = FindingsTopologyData['nodes'][number]
type LayoutEdge = FindingsTopologyData['edges'][number]

interface InternalTreeNode extends LayoutNode {
  children?: InternalTreeNode[]
  __placeholder?: boolean
}

const PLACEHOLDER_PREFIX = '__placeholder__'
const VIRTUAL_ROOT_ID = '__virtual__'

const isPlaceholderId = (id: string) => id.startsWith(PLACEHOLDER_PREFIX)

const createPlaceholderNode = (parentId: string, index: number): InternalTreeNode => ({
  id: `${PLACEHOLDER_PREFIX}:${parentId}:${index}`,
  type: 'findingsNode',
  __placeholder: true,
  data: {
    nodeId: `${PLACEHOLDER_PREFIX}:${parentId}:${index}`,
    title: '',
    status: 'normal',
    visualStatus: 'normal',
    nodeKind: 'summary',
    disabled: true,
  } as FindingsNodeData,
})

const buildTree = (data: FindingsTopologyData) => {
  const nodes = [...(data?.nodes || [])]
  const edges = [...(data?.edges || [])]

  const outAdj = new Map<string, string[]>()
  const inDeg = new Map<string, number>()

  nodes.forEach((node) => inDeg.set(node.id, 0))

  edges.forEach((edge) => {
    if (!outAdj.has(edge.source)) outAdj.set(edge.source, [])
    outAdj.get(edge.source)!.push(edge.target)
    inDeg.set(edge.target, (inDeg.get(edge.target) || 0) + 1)
  })

  let roots = nodes.filter((node) => (inDeg.get(node.id) || 0) === 0).map((node) => node.id)
  if (!roots.length && nodes.length > 0) roots = [nodes[0].id]

  const hasVirtualRoot = roots.length > 1
  if (hasVirtualRoot) {
    nodes.push({
      id: VIRTUAL_ROOT_ID,
      type: 'findingsNode',
      data: {
        nodeId: VIRTUAL_ROOT_ID,
        title: 'root',
        status: 'normal',
        visualStatus: 'normal',
        nodeKind: 'summary',
      },
    })
    roots.forEach((rootId) => {
      edges.push({ id: `v-${rootId}`, source: VIRTUAL_ROOT_ID, target: rootId })
    })
    roots = [VIRTUAL_ROOT_ID]
  }

  const parent = new Map<string, string>()
  const children = new Map<string, string[]>()
  const queue = [...roots]

  while (queue.length > 0) {
    const current = queue.shift()!
    const nextChildren = outAdj.get(current) || []
    for (const childId of nextChildren) {
      if (!parent.has(childId)) {
        parent.set(childId, current)
        if (!children.has(current)) children.set(current, [])
        children.get(current)!.push(childId)
        queue.push(childId)
      }
    }
  }

  const id2node = new Map(nodes.map((node) => [node.id, node]))

  const buildSubtree = (id: string): InternalTreeNode => {
    const base = id2node.get(id) || {
      id,
      type: 'findingsNode',
      data: {
        nodeId: id,
        title: id,
        status: 'normal',
        visualStatus: 'normal',
        nodeKind: 'summary',
      },
    }

    const realChildren = (children.get(id) || []).map(buildSubtree)
    const reservedSlots = RESERVED_CHILD_SLOT_COUNTS[id] ?? 0
    const placeholders = Array.from({
      length: Math.max(reservedSlots - realChildren.length, 0),
    }).map((_, index) => createPlaceholderNode(id, index))

    return {
      ...base,
      children: [...realChildren, ...placeholders],
    }
  }

  return { treeData: buildSubtree(roots[0]), hasVirtualRoot }
}

export const buildFindingsLayout = (
  data: FindingsTopologyData,
  findingsId?: string,
  onNodeClick?: (node: any) => void,
) => {
  const { treeData } = buildTree(data)
  const root = hierarchy(treeData, (node: InternalTreeNode) => node.children || [])
  const layout = d3tree<InternalTreeNode>()
    .nodeSize([SIBLING_GAP + NODE_HEIGHT, LAYER_GAP + NODE_WIDTH])
    .separation((a, b) => (a.parent === b.parent ? 1 : 1.25))

  layout(root)

  const xs: number[] = []
  root.each((node) => xs.push(node.x))
  const xMin = Math.min(...xs)
  const xMax = Math.max(...xs)
  const vCenter = (xMin + xMax) / 2

  const nodes: any[] = []
  root.each((node) => {
    if (node.data.__placeholder) {
      return
    }

    nodes.push({
      id: node.data.id,
      type: node.data.type || 'findingsNode',
      data: {
        ...(node.data.data || node.data),
        nodeId: node.data.id,
        findingsId,
        onClick: onNodeClick,
      },
      position: { x: node.y, y: node.x - vCenter },
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
      sourcePosition: 'right',
      targetPosition: 'left',
    })
  })

  const realNodeIds = new Set(nodes.map((node) => node.id))
  const edges: LayoutEdge[] = []

  root.each((node) => {
    if (!node.children || !realNodeIds.has(node.data.id)) {
      return
    }

    node.children.forEach((child) => {
      if (child.data.__placeholder || !realNodeIds.has(child.data.id) || isPlaceholderId(child.data.id)) {
        return
      }

      edges.push({
        id: `${node.data.id}-${child.data.id}`,
        source: node.data.id,
        target: child.data.id,
      })
    })
  })

  return { nodes, edges }
}
