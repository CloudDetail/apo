import { CButton, CCard, CCardBody, CCardHeader } from '@coreui/react'
import React, { useEffect, useMemo } from 'react'
import { useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { getServiceRelationApi } from 'src/api/serviceInfo'
import Topology from 'src/components/ReactFlow/Topology'
import { usePropsContext } from 'src/contexts/PropsContext'
import TimelapseLineChart from './TimelapseLineChart'
import DependentTable from './DependentTable'
import { Modal } from 'antd'
import LoadingSpinner from 'src/components/Spinner'
function escapeId(id) {
  return id.replace(/[^a-zA-Z0-9-_]/g, '_')
}
function contactServiceEndpoint(service, endpoint) {
  return escapeId(service + '-' + endpoint)
}
export default function TopologyModal(props) {
  const [visible, setVisible] = useState(false)
  const { serviceName, endpoint } = usePropsContext()
  const { startTime, endTime } = props
  const [topologyData, setTopologyData] = useState({
    nodes: [],
    edges: [],
  })
  const [allTopologyData, setAllTopologyData] = useState(null)
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(false)
  const { displayData, modalService, modalEndpoint } = useSelector((state) => state.topologyReducer)
  const dispatch = useDispatch()
  const clearTopology = () => {
    dispatch({ type: 'clearTopology' })
  }
  const setModalData = (value) => {
    dispatch({ type: 'setModalData', payload: value })
  }

  const openModal = () => {
    setVisible(true)
    setModalData({
      modalService: serviceName ?? null,
      modalEndpoint: endpoint ?? null,
      displayData: null,
    })
  }
  // current parent children的数据处理
  const prepareAllTopologyData = (data) => {
    if (!data) {
      setAllTopologyData({ nodes: [], edges: [] })
      return
    }
    const nodeSet = new Set() // 用于存储已添加节点的 ID
    const edgeSet = new Set() // 用于存储已添加边的 ID
    const current = data.current

    const nodes = [
      {
        id: contactServiceEndpoint(current.service, current.endpoint),
        data: {
          label: current.service,
          isTraced: current.isTraced,
          service: current.service,
          endpoint: current.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      },
    ]
    nodeSet.add(contactServiceEndpoint(current.service, current.endpoint))
    const edges = []
    data.parents?.forEach((parent) => {
      const nodeId = contactServiceEndpoint(parent.service, parent.endpoint)
      if (!nodeSet.has(nodeId)) {
        nodes.push({
          id: nodeId,
          data: {
            label: parent.service,
            isTraced: parent.isTraced,
            service: parent.service,
            endpoint: parent.endpoint,
          },
          position: { x: 0, y: 0 },
          type: 'serviceNode',
        })
        nodeSet.add(nodeId)
      }
      const edgeId = nodeId + '-' + contactServiceEndpoint(current.service, current.endpoint)
      if (!edgeSet.has(edgeId)) {
        const targetId = contactServiceEndpoint(current.service, current.endpoint)
        edges.push({
          id: edgeId,
          source: nodeId,
          target: targetId,
          type: nodeId === targetId ? 'loop' : 'smart',
          markerEnd: 'url(#arrowhead)',
        })
        edgeSet.add(edgeId)
      }
    })
    data.childRelations?.forEach((child, index) => {
      const nodeId = contactServiceEndpoint(child.service, child.endpoint)
      if (!nodeSet.has(nodeId)) {
        nodes.push({
          id: nodeId,
          data: {
            label: child.service,
            isTraced: child.isTraced,
            service: child.service,
            endpoint: child.endpoint,
          },
          position: { x: 0, y: 0 },
          type: 'serviceNode',
        })
        nodeSet.add(nodeId)
      }

      const edgeId =
        contactServiceEndpoint(child.parentService, child.parentEndpoint) + '-' + nodeId
      if (!edgeSet.has(edgeId)) {
        const sourceId = contactServiceEndpoint(child.parentService, child.parentEndpoint)
        edges.push({
          id: edgeId,
          source: sourceId,
          target: nodeId,
          type: nodeId === sourceId ? 'loop' : 'smart',
        })
        edgeSet.add(edgeId)
      }
    })

    setAllTopologyData({ nodes: [...nodes], edges })
    // console.log({ nodes: [...nodes], edges: [...edges] })
    // console.log('全拓扑图整理完毕', new Date())

    // return { nodes: [...nodes], edges }
  }
  //准备传递的可见的props
  // 找出所有上游节点的函数
  function createEdgesMap(edges) {
    const edgesMap = {}
    edges.forEach((edge) => {
      if (!edgesMap[edge.target]) {
        edgesMap[edge.target] = []
      }
      edgesMap[edge.target].push(edge.source)
    })
    // console.log(edgesMap)
    return edgesMap
  }
  const edgesMap = useMemo(
    () => allTopologyData?.edges && createEdgesMap(allTopologyData?.edges),
    [allTopologyData?.edges],
  )
  function findUpstreamNodes(nodeId) {
    const upstreamNodes = new Set()
    function findParents(currentNodeId) {
      if (edgesMap[currentNodeId]) {
        edgesMap[currentNodeId].forEach((source) => {
          if (!upstreamNodes.has(source)) {
            upstreamNodes.add(source)
            if (source !== currentNodeId) findParents(source)
          }
        })
      }
    }

    findParents(nodeId)
    return Array.from(upstreamNodes)
  }

  // 找出所有需要展示的节点和边
  function getNodesAndEdgesToDisplay(selectedNodeIds, allNodes, allEdges) {
    selectedNodeIds = Array.from(new Set(selectedNodeIds))
    const nodesToDisplay = new Set()
    const edgesToDisplay = new Set()
    // console.log(selectedNodeIds, allNodes, allEdges)

    selectedNodeIds.forEach((nodeId) => {
      nodesToDisplay.add(nodeId)

      // 找到该节点的所有上游节点
      const upstreamNodes = findUpstreamNodes(nodeId)
      // console.log(upstreamNodes)
      upstreamNodes.forEach((upNodeId) => nodesToDisplay.add(upNodeId))

      // 添加相关的边
      allEdges.forEach((edge) => {
        if (nodesToDisplay.has(edge.target) && nodesToDisplay.has(edge.source)) {
          edgesToDisplay.add(edge.id)
        }
      })
    })
    // 根据 ID 找到实际的节点和边数据
    const filteredNodes = allNodes.filter((node) => nodesToDisplay.has(node.id))
    const filteredEdges = allEdges.filter((edge) => edgesToDisplay.has(edge.id))
    const moreNodes = []
    const moreEdges = []
    filteredNodes.forEach((node) => {
      const children = allEdges.filter(
        (edge) => node.id === edge.source && !edgesToDisplay.has(edge.id),
      )
      if (children?.length > 0) {
        moreNodes.push({
          id: node.id + '-child',
          data: {
            label: '更多',
            parentService: node.data.service,
            parentEndpoint: node.data.endpoint,
          },
          position: { x: 0, y: 0 },
          type: 'moreNode',
        })
        moreEdges.push({
          id: node.id + '-' + node.id + '-child',
          source: node.id,
          target: node.id + '-child',
          type: 'smart',
          markerEnd: 'url(#arrowhead)',
          // style:{
          //   stroke: '#6293FF'
          // }
        })
      }
    })
    // 在这里对 nodes 进行排序
    filteredNodes.sort((a, b) => a.id.localeCompare(b.id))
    moreNodes.sort((a, b) => a.id.localeCompare(b.id))
    filteredEdges.sort((a, b) => a.id.localeCompare(b.id))
    moreEdges.sort((a, b) => a.id.localeCompare(b.id))
    const nodes = filteredNodes.concat(moreNodes)
    const edges = filteredEdges.concat(moreEdges)
    //console.log(nodes, edges)

    setTopologyData({
      nodes: nodes,
      edges: edges,
    })
    setLoading(false)

    //console.log('可见数据整理完毕', new Date())

    return { filteredNodes, filteredEdges }
  }
  useEffect(() => {
    if (
      allTopologyData !== null &&
      allTopologyData !== undefined &&
      displayData !== null &&
      displayData !== undefined
    ) {
      const currentChild = data.childRelations.filter(
        (item) =>
          item.parentService === data.current.service &&
          item.parentEndpoint === data.current.endpoint,
      )

      const currentChildEndpoints = currentChild.map((item) =>
        contactServiceEndpoint(item.service, item.endpoint),
      )

      const selectedNodeIds = [
        ...displayData.map((data) => contactServiceEndpoint(data.serviceName, data.endpoint)),
        ...currentChildEndpoints,
        contactServiceEndpoint(data.current.service, data.current.endpoint),
      ]
      // console.log(allTopologyData, displayData)
      getNodesAndEdgesToDisplay(selectedNodeIds, allTopologyData.nodes, allTopologyData.edges)
    }
  }, [allTopologyData, displayData])

  useEffect(() => {
    if (modalService && modalEndpoint) {
      setLoading(true)
      setData(null)
      setAllTopologyData(null)
      getServiceRelationApi({
        startTime: startTime,
        endTime: endTime,
        service: modalService,
        endpoint: modalEndpoint,
      })
        .then((res) => {
          setData(res)
          prepareAllTopologyData(res)
        })
        .catch((error) => {
          // setTopologyData({ nodes: [], edges: [] })
          setLoading(false)
        })
    }
  }, [modalService, modalEndpoint, startTime, endTime])
  const closeModal = () => {
    clearTopology()
    setVisible(false)
    setAllTopologyData(null)
    setTopologyData({
      nodes: [],
      edges: [],
    })
  }
  useEffect(() => {
    closeModal()
  }, [serviceName, endpoint])
  const memoProps = useMemo(() => {
    return {
      serviceName: modalService,
      endpoint: modalEndpoint,
      startTime: startTime,
      endTime: endTime,
      storeDisplayData: true,
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }
  }, [modalService, modalEndpoint, startTime, endTime])
  return (
    <>
      <CButton color="primary" variant="ghost" size="sm" className="text-xs" onClick={openModal}>
        查看更多下游依赖
      </CButton>
      {visible && (
        <>
          {/* <div className=" fixed bg-black h-full w-full top-0" style={{ zIndex: 999999 }}>
          </div> */}
          <Modal
            title={modalService + '更多下游依赖'}
            open={visible}
            footer={null} // 如果你不需要默认的底部按钮
            style={{ top: 0, left: 0, right: 0, bottom: 0, width: '100vw', height: '100vh' }}
            bodyStyle={{ height: 'calc(100vh - 75px)', overflowY: 'auto' }}
            width="100vw"
            onCancel={closeModal}
            destroyOnClose
          >
            <CCard className="h-1/2">
              <LoadingSpinner loading={loading} />
              <Topology canZoom={false} data={topologyData} />
            </CCard>
            <div className="flex flex-row h-1/2 pt-2">
              <CCard className="w-1/2 mr-2 h-full">
                <CCardHeader>依赖节点延时曲线全览对比图</CCardHeader>
                <CCardBody className="text-xs overflow-hidden p-0">
                  <TimelapseLineChart {...memoProps} />
                </CCardBody>
              </CCard>
              <CCard className="w-1/2 ml-2  h-full">
                <CCardHeader>依赖节点延时曲线相似度排序</CCardHeader>
                <CCardBody className="text-xs overflow-hidden p-0">
                  <DependentTable {...memoProps} />
                </CCardBody>
              </CCard>
            </div>
          </Modal>
        </>
      )}
    </>
  )
}
