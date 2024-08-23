import React, { useState, useEffect, useCallback, useRef } from 'react'
import ReactFlow, {
  ReactFlowProvider,
  addEdge,
  useNodesState,
  useEdgesState,
  MarkerType,
  useReactFlow,
} from 'reactflow'
import { graphlib, layout as dagreLayout } from 'dagre'
import * as d3 from 'd3'
import 'reactflow/dist/style.css'
import ServiceNode from './ServiceNode'
import { useDispatch, useSelector } from 'react-redux'
import MoreNode from './MoreNode'
import { SmartBezierEdge } from '@tisoap/react-flow-smart-edge'
import 'reactflow/dist/style.css'
import { AiOutlineRollback } from 'react-icons/ai'
import { Tooltip } from 'antd'
const nodeWidth = 200
const nodeHeight = 60
const nodeTypes = { serviceNode: ServiceNode, moreNode: MoreNode } // 定义在组件外部
const edgeTypes = {
  smart: SmartBezierEdge, // 或者使用 SmartBezierEdge 等
}
const LayoutFlow = (props) => {
  const { data } = props
  // 所有链路
  const reactFlowInstance = useRef(null)
  const { canZoom } = props
  // const [initialNodes, setInitialNodes] = useState([])
  // const [initialEdges, setInitialEdges] = useState([])
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const dispatch = useDispatch()

  const setModalData = (value) => {
    dispatch({ type: 'setModalData', payload: value })
  }
  const { fitView } = useReactFlow()
  const prepareData = () => {
    const initialNodes = data.nodes ?? []
    const initialEdges = []
    data.edges.forEach((edge) => {
      initialEdges.push({
        ...edge,
        markerEnd: markerEnd,
        style: {
          stroke: '#6293FF',
        },
      })
    })
    return { initialNodes, initialEdges }
  }
  const markerEnd = {
    type: MarkerType.ArrowClosed,
    strokeWidth: 5,
    width: 25,
    height: 25,
    color: '#6293ff',
  }
  const dagreGraph = new graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))

  const getLayoutedElements = (nodes, edges) => {
    dagreGraph.setGraph({ rankdir: 'LR', ranksep: 100, nodesep: 50 }) // 自上而下的布局

    nodes.forEach((node) => {
      dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight })
    })

    edges.forEach((edge) => {
      dagreGraph.setEdge(edge.source, edge.target)
    })

    dagreLayout(dagreGraph)
    nodes.forEach((node) => {
      const nodeWithPosition = dagreGraph.node(node.id)
      node.targetPosition = 'left'
      node.sourcePosition = 'right'

      node.position = {
        x: nodeWithPosition.x - nodeWidth,
        y: nodeWithPosition.y - nodeHeight / 2,
      }
    })

    // Calculate offsets to center the graph
    const xMin = Math.min(...nodes.map((node) => node.position.x))
    const yMin = Math.min(...nodes.map((node) => node.position.y))

    nodes.forEach((node) => {
      node.position.x -= xMin - nodeWidth / 2
      node.position.y -= yMin - nodeHeight / 2
    })
    return { nodes, edges }
  }
  const clickNode = (e, node) => {
    if (node.type === 'moreNode') {
      setModalData({
        modalService: node.data.parentService,
        modalEndpoint: node.data.parentEndpoint,
        displayData: null,
      })
    }
  }

  useEffect(() => {
    const { initialNodes, initialEdges } = prepareData()
    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
      initialNodes,
      initialEdges,
    )
    setNodes([...layoutedNodes])
    setEdges([...layoutedEdges])
    requestAnimationFrame(() => {
      if (reactFlowInstance.current) {
        setTimeout(() => {
          fitView({
            padding: layoutedNodes.length > 1 ? 0.1 : 0.2,
            includeHiddenNodes: true,
          })
        }, 20)
      }
    })
  }, [data])
  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      // edgeTypes={edgeTypes}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      nodeTypes={nodeTypes}
      ref={reactFlowInstance}
      minZoom={0.1} // 设置最小缩放
      maxZoom={2} // 设置最大缩放
      onNodeClick={clickNode}
    />
  )
}
function FlowWithProvider(props) {
  const { modalDataUrl } = useSelector((state) => state.topologyReducer)
  const dispatch = useDispatch()

  const rollback = (value) => {
    dispatch({ type: 'rollback', payload: value })
  }
  console.log(modalDataUrl, modalDataUrl[modalDataUrl.length - 2])

  return (
    <>
      {modalDataUrl?.length > 1 && (
        <Tooltip
          title={'点击回退上级入口' + modalDataUrl[modalDataUrl.length - 2]?.modalService}
          placement="bottom"
        >
          <div
            className=" absolute top-0 right-3 h-10 flex items-center justify-center cursor-pointer"
            style={{ zIndex: 1 }}
            onClick={() => rollback()}
          >
            点击回退上级入口拓扑图
            <AiOutlineRollback size={28} />
          </div>
        </Tooltip>
      )}
      <ReactFlowProvider>
        <svg style={{ position: 'absolute', top: 0, left: 0 }}>
          <defs>
            <marker
              id="arrowhead"
              viewBox="0 0 10 10"
              refX="5"
              refY="5"
              markerWidth="6"
              markerHeight="6"
              orient="auto"
            >
              <path d="M0,0 L10,5 L0,10" fill="#6293ff" stroke="#6293ff" />
            </marker>
          </defs>
        </svg>

        <LayoutFlow {...props} />
      </ReactFlowProvider>
    </>
  )
}
export default FlowWithProvider
