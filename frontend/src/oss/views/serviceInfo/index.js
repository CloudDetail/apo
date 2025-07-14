/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useMemo } from 'react'
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
import { ServiceInfoProvider } from 'src/oss/contexts/ServiceInfoContext'
import { Card } from 'antd'
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
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const stringToArray = (value) => {
    if (!value) return null
    return value.split(',').filter(Boolean)
  }
  const clusterIdsStr = searchParams.get('clusterIds')
  const clusterIds = useMemo(() => {
    return stringToArray(clusterIdsStr) || null
  }, [clusterIdsStr])
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
          outOfGroup: current.outOfGroup,
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
          outOfGroup: child.outOfGroup,
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
          outOfGroup: parent.outOfGroup,
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
        clusterIds: clusterIds,
        groupId: dataGroupId,
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
      if (dataGroupId !== null && serviceName && endpoint) {
        getServiceTopology()
      }
    },
    300, // 延迟时间 300ms
    [serviceName, startTime, endTime, endpoint, dataGroupId, clusterIds],
  )

  return (
    <div className="h-full relative">
      <LoadingSpinner loading={loading} />
      <ServiceInfoProvider>
        <PropsProvider value={{ serviceName, endpoint, clusterIds }}>
          <>
            <div className="flex flex-row">
              <Card
                size="small"
                title={
                  <div>
                    {serviceName}
                    {t('index.directDependencies')}
                    <div className="text-xs">
                      {t('index.serviceEndpoint')}: {endpoint}
                    </div>
                  </div>
                }
                className="mb-4 mr-1 h-[350px] w-2/5 whitespace-normal flex flex-col"
                classNames={{
                  body: 'h-0 flex-1',
                  title: 'whitespace-normal',
                }}
                styles={{
                  title: {
                    whiteSpace: 'normal',
                  },
                }}
                extra={<TopologyModal startTime={startTime} endTime={endTime} />}
              >
                <div className="h-full w-full">
                  <Topology canZoom={false} data={topologyData} />
                </div>
              </Card>
              <DependentTabs />
            </div>
            <InfoUni />
          </>
        </PropsProvider>
      </ServiceInfoProvider>
    </div>
  )
}

export default ServiceInfo
