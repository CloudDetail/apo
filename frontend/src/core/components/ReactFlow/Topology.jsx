/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useMemo } from 'react'
import ReactFlow, {
  ReactFlowProvider,
  useNodesState,
  useEdgesState,
  MarkerType,
  useReactFlow,
} from 'reactflow'
import { graphlib, layout as dagreLayout } from 'dagre'
import 'reactflow/dist/style.css'
import ServiceNode from './ServiceNode'
import { useDispatch, useSelector } from 'react-redux'
import MoreNode from './MoreNode'
import 'reactflow/dist/style.css'
import { AiOutlineRollback } from 'react-icons/ai'
import { theme, Tooltip } from 'antd'
import CustomSelfLoopEdge from './LoopEdges'
import './index.css'
import { useTranslation } from 'react-i18next'
const nodeWidth = 200
const defaultNodeTypes = {
  serviceNode: ServiceNode,
  moreNode: MoreNode,
} // 定义在组件外部
const edgeTypes = {
  // smart: SmartBezierEdge, // 或者使用 SmartBezierEdge 等
  loop: CustomSelfLoopEdge,
}
const LayoutFlow = (props) => {
  const { data, nodeHeight = 60, nodeTypes = {}, active = true } = props
  // 所有链路
  const reactFlowInstance = useRef(null)
  // const [initialNodes, setInitialNodes] = useState([])
  // const [initialEdges, setInitialEdges] = useState([])
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const dispatch = useDispatch()

  const setModalData = (value) => {
    dispatch({ type: 'setModalData', payload: value })
  }
  const { fitView } = useReactFlow()
  const { useToken } = theme
  const { token } = useToken()
  const prepareData = () => {
    const initialNodes = data?.nodes || []
    const initialEdges = []
    data.edges.forEach((edge) => {
      initialEdges.push({
        ...edge,
        markerEnd: markerEnd,
        style: {
          stroke: token.colorPrimaryText,
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
    color: token.colorPrimaryText,
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
    edges.map((edge) => {
      const sourceNode = nodes.find((node) => node.id === edge.source)
      const targetNode = nodes.find((node) => node.id === edge.target)
      if (
        sourceNode &&
        targetNode &&
        sourceNode.position.x > targetNode.position.x &&
        Math.abs(sourceNode.position.y - targetNode.position.y) < nodeHeight
      ) {
        edge.type = 'loop'
      }
    })
    return { nodes, edges }
  }
  const clickNode = (e, node) => {
    if (node.type === 'moreNode' && !node.data.disabled) {
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
            padding: layoutedNodes.length > 2 ? 0.1 : 0.2,
            includeHiddenNodes: true,
          })
        }, 20)
      }
    })
  }, [data])
  const onLoad = () => {
    console.log(1)
  }
  const memoNodeTypes = useMemo(() => ({ ...nodeTypes, ...defaultNodeTypes }), [])
  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      edgeTypes={edgeTypes}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      nodeTypes={memoNodeTypes}
      ref={reactFlowInstance}
      minZoom={0.1} // 设置最小缩放
      maxZoom={2} // 设置最大缩放
      onNodeClick={clickNode}
      onLoad={onLoad}
      nodesDraggable={active}
      elementsSelectable={active}
      panOnDrag={active}
      zoomOnScroll={active}
      zoomOnPinch={active}
    />
  )
}
function FlowWithProvider(props) {
  const { t } = useTranslation('oss/serviceInfo')
  const { useToken } = theme
  const { token } = useToken()
  const { modalDataUrl } = useSelector((state) => state.topologyReducer)
  const dispatch = useDispatch()

  const rollback = (value) => {
    dispatch({ type: 'rollback', payload: value })
  }

  return (
    <>
      {modalDataUrl?.length > 1 && (
        <Tooltip
          title={t('topology.clickToReturn') + modalDataUrl[modalDataUrl.length - 2]?.modalService}
          placement="bottom"
        >
          <div
            className=" absolute top-12 right-8 h-10 flex items-center justify-center cursor-pointer"
            style={{ zIndex: 1 }}
            onClick={() => rollback()}
          >
            {t('topology.clickToReturnUpperTopology')}
            <AiOutlineRollback size={28} />
          </div>
        </Tooltip>
      )}
      <ReactFlowProvider>
        <svg style={{ position: 'absolute', top: 0, left: 0 }}>
          <defs>
            <marker
              id="arrowhead"
              viewBox="0 0 74.4539794921875 67"
              refX="37.227"
              refY="33.5"
              markerWidth="16"
              markerHeight="16"
            >
              <path
                d="M45.4542 4.75L73.167 52.75C76.8236 59.0833 72.2529 67 64.9398 67L9.51418 67C2.20107 67 -2.36962 59.0833 1.28693 52.75L28.9997 4.75C32.6563 -1.58334 41.7977 -1.58334 45.4542 4.75Z"
                fill={token.colorPrimary}
              />
            </marker>
          </defs>
        </svg>

        <LayoutFlow {...props} />
      </ReactFlowProvider>
    </>
  )
}
export default FlowWithProvider
