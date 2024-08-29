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
import React, { useState } from 'react'
import DependentTable from './DependentTable'
import { usePropsContext } from 'src/contexts/PropsContext'
import TimelapseLineChart from './TimelapseLineChart'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'

function DependentTabs() {
  const { serviceName, endpoint } = usePropsContext()
  const [serviceList, setServiceList] = useState([])
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [activeItemKey, setActiveItemKey] = useState('timelapse')
  return (
    <CCard className="mb-4 ml-1 h-[350px] p-2  w-3/5">
      <CCardHeader>{serviceName}的所有依赖视图（包括所有递归依赖）</CCardHeader>
      <CCardBody className="text-xs overflow-hidden p-0">
        <CTabs
          activeItemKey={activeItemKey}
          className="w-full h-full overflow-hidden flex flex-col "
          onChange={(value) => setActiveItemKey(value)}
        >
          <CTabList variant="tabs" className="flex-grow-0 flex-shrink-0">
            <CTab itemKey="timelapse">依赖节点延时曲线全览对比图</CTab>
            <CTab itemKey="table">依赖节点延时曲线相似度排序</CTab>
          </CTabList>
          <CTabContent className="h-full overflow-hidden flex-grow">
            <CTabPanel itemKey="timelapse" className="overflow-hidden h-full">
              {activeItemKey === 'timelapse' && (
                <TimelapseLineChart
                  endpoint={endpoint}
                  startTime={startTime}
                  endTime={endTime}
                  serviceName={serviceName}
                />
              )}
            </CTabPanel>
            <CTabPanel itemKey="table" className="h-full overflow-hidden">
              <DependentTable
                endpoint={endpoint}
                startTime={startTime}
                endTime={endTime}
                serviceName={serviceName}
              />
            </CTabPanel>
          </CTabContent>
        </CTabs>
      </CCardBody>
    </CCard>
  )
}

export default DependentTabs
