import { useEffect, useRef, useState } from 'react'
import Topology from 'src/core/components/ReactFlow/Topology'
import MetricsNode from './MetricsNode'
import Empty from 'src/core/components/Empty/Empty'

function escapeId(id) {
  return id.replace(/[^a-zA-Z0-9-_]/g, '_')
}
function contactLevelService(level, service) {
  return escapeId(level + '-' + service)
}
const AffectedEndpoints = (props) => {
  console.log(props)
  const { topology: topologyProps } = props
  const [topologyData, setTopologyData] = useState({
    nodes: [],
    edges: [],
  })
  // current parent children的数据处理

  const prepareTopologyData = (data) => {
    console.log(data)
    if (!data) {
      setTopologyData({ nodes: [], edges: [] })
      return
    }
    // 当 props 形如 { children: [], current: {}, parents: [] } 时，展示 Empty
    const isEmptyTopology =
      (!data.current || Object.keys(data.current).length === 0 || !data.current.service) &&
      (!data.parents || data.parents.length === 0) &&
      (!data.children || data.children.length === 0)
    if (isEmptyTopology) {
      return { nodes: [], edges: [] }
    }

    const nodeSet = new Set() // 用于存储已添加节点的 ID
    const edgeSet = new Set() // 用于存储已添加边的 ID
    const current = data.current
    const currentId = contactLevelService('current', current.service)
    const nodes = [
      {
        id: currentId,
        data: {
          label: current.service,
          service: current.service,
          description: current.description,
          metrics: current.metrics,
        },
        position: { x: 0, y: 0 },
        type: 'metricsNode',
      },
    ]
    nodeSet.add(currentId)
    const edges = []
    data.parents?.forEach((parent) => {
      const nodeId = contactLevelService('parent', parent.service)
      if (!nodeSet.has(nodeId)) {
        nodes.push({
          id: nodeId,
          data: {
            label: parent.service,
            service: parent.service,
            description: parent.description,
            metrics: parent.metrics,
          },
          position: { x: 0, y: 0 },
          type: 'metricsNode',
        })
        nodeSet.add(nodeId)
      }
      const edgeId = nodeId + '-' + currentId
      if (!edgeSet.has(edgeId)) {
        edges.push({
          id: edgeId,
          source: nodeId,
          target: currentId,
          type: nodeId === currentId ? 'loop' : 'smart',
          markerEnd: 'url(#arrowhead)',
        })
        edgeSet.add(edgeId)
      }
    })
    data.children?.forEach((child, index) => {
      const childId = contactLevelService('child', child.service)
      if (!nodeSet.has(childId)) {
        nodes.push({
          id: childId,
          data: {
            label: child.service,
            service: child.service,
            description: child.description,
            metrics: child.metrics,
          },
          position: { x: 0, y: 0 },
          type: 'metricsNode',
        })
        nodeSet.add(childId)
      }

      const childEdgeId = currentId + '-' + childId
      if (!edgeSet.has(childEdgeId)) {
        edges.push({
          id: childEdgeId,
          source: currentId,
          target: childId,
          type: currentId === childId ? 'loop' : 'smart',
        })
        edgeSet.add(childEdgeId)
      }
    })
    let resultNodes = [...nodes]
    // prepareDataForChartAndTopologyNode(result)
    console.log({ nodes: resultNodes, edges })
    return { nodes: resultNodes, edges }
  }

  useEffect(() => {
    if (topologyProps && Object.keys(topologyProps).length > 0) {
      setTopologyData(prepareTopologyData(topologyProps))
    }
  }, [topologyProps])
  const wrapperRef = useRef<HTMLDivElement>(null)
  const preRef = useRef<HTMLDivElement>(null)
  const [wrapActive, setWrapActive] = useState(false)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setWrapActive(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])
  useEffect(() => {
    console.log(wrapActive)
    if (wrapperRef.current && preRef.current) {
      wrapperRef.current.style.pointerEvents = wrapActive ? 'auto' : 'none'
      preRef.current.style.outline = wrapActive ? 'rgb(110, 159, 255) solid 2px' : 'none'
    }
  }, [wrapActive])
  return topologyData?.nodes.length === 0 ? (
    <Empty />
  ) : (
    <div ref={wrapperRef} className="w-full h-[500px] z-40 relative">
      <div
        ref={preRef}
        className="w-full h-[500px] z-50 relative"
        onClick={() => setWrapActive(true)}
      >
        <Topology
          data={topologyData}
          nodeHeight={260}
          nodeWidth={200}
          nodeTypes={{ metricsNode: MetricsNode }}
          active={wrapActive}
        />
      </div>
    </div>
  )
}
export default AffectedEndpoints
