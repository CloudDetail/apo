/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CCard, CCardHeader } from '@coreui/react'
import React, { useState, useEffect } from 'react'
import Topology from 'src/core/components/ReactFlow/Topology'
import { useLocation } from 'react-router-dom'
import InfoUni from './component/infoUni'
import { PropsProvider } from 'src/core/contexts/PropsContext'
import LoadingSpinner from 'src/core/components/Spinner'
import DependentTabs from './component/dependent'
import { getServiceTopologyApi } from 'core/api/serviceInfo'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import TopologyModal from './component/dependent/TopologyModal'
import { useDebounce } from 'react-use'
import { useTranslation } from 'react-i18next'
function escapeId(id) {
  return id.replace(/[^a-zA-Z0-9-_]/g, '_')
}
function contactServiceEndpoint(service, endpoint) {
  return escapeId(service + '-' + endpoint)
}
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
  const { t } = useTranslation('oss/serviceInfo')
  // current parent children的数据处理
  const prepareTopologyData = (data) => {
    if (!data) {
      return { nodes: [], edges: [] }
    }
    const current = data.current
    const currentNodeId = 'current-' + contactServiceEndpoint(current.service, current.endpoint)
    const nodes = [
      {
        id: currentNodeId,
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
      const childNodeId = 'child-' + contactServiceEndpoint(child.service, child.endpoint)
      nodes.push({
        id: childNodeId,
        data: {
          label: child.service,
          isTraced: child.isTraced,
          endpoint: child.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      })
      edges.push({
        id: currentNodeId + '-' + childNodeId,
        source: currentNodeId,
        target: childNodeId,
      })
    })
    data.parents?.forEach((parent) => {
      const parentNodeId = 'parent-' + contactServiceEndpoint(parent.service, parent.endpoint)
      nodes.push({
        id: parentNodeId,
        data: {
          label: parent.service,
          isTraced: parent.isTraced,
          endpoint: parent.endpoint,
        },
        position: { x: 0, y: 0 },
        type: 'serviceNode',
      })
      edges.push({
        id: parentNodeId + '-' + currentNodeId,
        source: parentNodeId,
        target: currentNodeId,
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
                  {serviceName}
                  {t('index.directDependencies')}
                  <div className="text-xs">
                    {t('index.serviceEndpoint')}: {endpoint}
                  </div>
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
