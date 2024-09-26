import { CCard, CCardHeader } from '@coreui/react'
import React, { useState, useEffect } from 'react'
import Topology from 'components/ReactFlow/Topology'
import { serviceMock } from 'components/ReactFlow/mock'
import { useLocation } from 'react-router-dom'
import InfoUni from './component/infoUni'
import { PropsProvider } from 'src/contexts/PropsContext'
import LoadingSpinner from 'src/components/Spinner'
import DependentTabs from './component/dependent'
import { getServiceTopologyApi } from 'src/api/serviceInfo'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/store/reducers/timeRangeReducer'
import TopologyModal from './component/dependent/TopologyModal'
import { useDebounce } from 'react-use'
const ServiceInfo = () => {
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)

  const serviceName = searchParams.get('service-name')
  const endpoint = searchParams.get('endpoint')
  const [topologyData, setTopologyData] = useState({
    nodes: [],
    edges: [],
  })
  const [loading, setLoading] = useState(true)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  // current parent children的数据处理
  const prepareTopologyData = (data) => {
    if (!data) {
      return { nodes: [], edges: [] }
    }
    const current = data.current

    const nodes = [
      {
        id: 'current-' + current.service,
        data: {
          label: current.service,
          isTraced: current.isTraced,
          endpoint: current.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      },
    ]
    const edges = []
    data.children?.forEach((child) => {
      nodes.push({
        id: 'child-' + child.service,
        data: {
          label: child.service,
          isTraced: child.isTraced,
          endpoint: child.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      })
      edges.push({
        id: current.service + '-' + child.service,
        source: 'current-' + current.service,
        target: 'child-' + child.service,
      })
    })
    data.parents?.forEach((parent) => {
      nodes.push({
        id: 'parent-' + parent.service,
        data: {
          label: parent.service,
          isTraced: parent.isTraced,
          endpoint: parent.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      })
      edges.push({
        id: parent.service + '-' + current.service,
        source: 'parent-' + parent.service,
        target: 'current-' + current.service,
        // markerEnd: markerEnd,
        // style:{
        //   stroke: '#6293FF'
        // }
      })
    })
    return { nodes, edges }
  }
  const getServiceTopology = () => {
    setLoading(true)
    if (startTime && endTime) {
      getServiceTopologyApi({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
        endpoint: endpoint,
      })
        .then((res) => {
          const { nodes, edges } = prepareTopologyData(res)
          setTopologyData({ nodes, edges })
          setLoading(false)
        })
        .catch((error) => {
          setTopologyData({ nodes: [], edges: [] })
          setLoading(false)
        })
    }
  }

  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      getServiceTopology()
    },
    300, // 延迟时间 300ms
    [serviceName, startTime, endTime, endpoint],
  )

  return (
    <div className="h-full relative">
      <LoadingSpinner loading={loading} />
      <PropsProvider value={{ serviceName, endpoint }}>
        <>
          <div className="flex flex-row">
            <CCard className="mb-4 mr-1 h-[350px] p-2 w-2/5">
              <CCardHeader className="py-0 flex flex-row justify-between">
                <div>
                  {serviceName}的直接上下游依赖关系图
                  <div className="text-xs">服务端点: {endpoint}</div>
                </div>
                <TopologyModal startTime={startTime} endTime={endTime} />
              </CCardHeader>
              <Topology canZoom={false} data={topologyData} />
            </CCard>
            <DependentTabs />
          </div>
          <InfoUni />
        </>
      </PropsProvider>
    </div>
  )
}

export default ServiceInfo
