import { CCard, CCardHeader } from '@coreui/react'
import React, { useState, useEffect } from 'react'
import Topology, { prepareTopologyData } from 'components/ReactFlow/Topology'
import { serviceMock } from 'components/ReactFlow/mock'
import { useLocation } from 'react-router-dom'
import InfoUni from './component/infoUni'
import { PropsProvider } from 'src/contexts/PropsContext'
import LoadingSpinner from 'src/components/Spinner'
import DependentTabs from './component/dependent'
import { getServiceTopologyApi } from 'src/api/serviceInfo'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/store/reducers/timeRangeReducer'
const ServiceInfo = () => {
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)

  const serviceName = searchParams.get('service-name')
  const endpoint = searchParams.get('endpoint')
  const [topologyData, setTopologyData] = useState({
    nodes:[],
    edges:[]
  })
  const [loading, setLoading] = useState(true)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  

  useEffect(() => {
    setLoading(true)

    getServiceTopologyApi({
      startTime:startTime,
      endTime: endTime,
      service:serviceName,
      endpoint: endpoint,
    }).then((res)=>{

      const {nodes,edges} = prepareTopologyData(res)
      setTopologyData({nodes,edges})
      setLoading(false)
    }).catch((error)=>{
      setTopologyData({nodes:[], edges:[]})
      setLoading(false)
    })
    
  }, [serviceName, startTime,endTime])
  return (
    <>
      <LoadingSpinner loading={loading} />
      <PropsProvider value={{ serviceName, endpoint }}>
        <>
          <div className="flex flex-row">
            <CCard className="mb-4 mr-1 h-[350px] p-2 w-2/5">
            <CCardHeader className="py-0">
              {serviceName}的直接上下游依赖关系图
              <div className="text-xs">服务端点: {endpoint}</div>
            </CCardHeader>
              <Topology canZoom={false} data={topologyData} />
            </CCard>
            <DependentTabs />
          </div>
          <InfoUni />
        </>
      </PropsProvider>
    </>
  )
}

export default ServiceInfo