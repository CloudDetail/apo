import {
  CCard,
  CCardBody,
  CCardHeader,
  CTab,
  CTabContent,
  CTabList,
  CTabPanel,
  CTabs,
} from '@coreui/react'
import React, { useEffect, useState } from 'react'
import DependentTable from './DependentTable'
import { usePropsContext } from 'src/contexts/PropsContext'
import TimelapseLineChart from './TimelapseLineChart'
import { serviceMock } from 'src/components/ReactFlow/mock'
import { getServiceDsecendantMetricsApi } from 'src/api/serviceInfo'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'
import { getStep } from 'src/utils/step'
import { showToast } from 'src/utils/toast'

function DependentTabs() {
  const { serviceName, endpoint } = usePropsContext()
  const [serviceList, setServiceList] = useState([])
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [chartData, setChartData] = useState()
  const [activeItemKey, setActiveItemKey] = useState('timelapse')
  const [loading, setLoading] = useState(false)

  const prepareMockData = (target) => {
    const result = new Set()

    function findNext(name) {
      const item = serviceMock.find((d) => d.name === name)
      if (item && item.next) {
        item.next.forEach((nextName) => {
          if (!result.has(nextName)) {
            result.add(nextName)
            findNext(nextName)
          }
        })
      }
    }

    findNext(target)
    return Array.from(result)
  }

  const getChartData = () => {
    getServiceDsecendantMetricsApi({
      startTime: startTime,
      endTime: endTime,
      service: serviceName,
      endpoint: endpoint,
      step: getStep(startTime, endTime),
    }).then((res) => {
      setChartData(res ?? [])
      setLoading(false)
    }).catch((error)=>{
      setChartData([])
      setLoading(false)
    })
  }
  useEffect(() => {
    setLoading(true)
    getChartData()
  }, [serviceName, startTime, endTime,endpoint])
  return (
    <CCard className="mb-4 ml-1 h-[350px] p-2  w-3/5">
      <CCardHeader>
        {serviceName}的依赖视图
        {/* <div className="text-xs">
          <span className="text-slate-300">展示所有多层依赖，且已按照延时曲线相似度排序，点击</span>
          <a className="cursor-pointer underline" onClick={() => setVisible(true)}>
            查看延时曲线全览对比图
          </a>
        </div> */}
      </CCardHeader>
      <CCardBody className="text-xs overflow-hidden p-0">
        <CTabs activeItemKey={activeItemKey} className="w-full h-full overflow-hidden flex flex-col " onChange={(value)=>setActiveItemKey(value)}>
          <CTabList variant="tabs" className="flex-grow-0 flex-shrink-0" >
            <CTab itemKey="timelapse">依赖节点延时曲线全览对比图</CTab>
            <CTab itemKey="table">依赖节点延时曲线相似度排序</CTab>
          </CTabList>
          <CTabContent className="h-full overflow-hidden flex-grow">
            <CTabPanel itemKey="timelapse" className="overflow-hidden h-full" >
              {chartData && activeItemKey==='timelapse' && (
                <TimelapseLineChart data={chartData} startTime={startTime} endTime={endTime} />
              )}
            </CTabPanel>
            <CTabPanel itemKey="table" className="h-full overflow-hidden">
              <DependentTable serviceList={serviceList} />
            </CTabPanel>
          </CTabContent>
        </CTabs>
      </CCardBody>
    </CCard>
  )
}

export default DependentTabs
