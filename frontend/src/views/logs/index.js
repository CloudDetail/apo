import {
  CCard,
  CCol,
  CFormInput,
  CFormSelect,
  CRow,
  CTab,
  CTabContent,
  CTabList,
  CTabPanel,
  CTabs,
} from '@coreui/react'
import React, { useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router-dom'
import HighLightCode from '../serviceInfo/component/infoUni/HighLightCode'
import { serviceMock } from 'src/components/ReactFlow/mock'
import FaultSiteLogs from './FaultSiteLogs'
import FullLogs from './FullLogs'
import { useSelector } from 'react-redux'
import {
  getTimestampRange,
  selectProcessedTimeRange,
  timeRangeList,
} from 'src/store/reducers/timeRangeReducer'
import { ISOToTimestamp } from 'src/utils/time'
import { PropsProvider } from 'src/contexts/PropsContext'
import { InstanceProvider, useInstance } from 'src/contexts/InstanceContext'
import Empty from 'src/components/Empty/Empty'
function LogsPage() {
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)
  const [activeItemKey, setActiveItemKey] = useState('faultSite')
  const [startTime, setStartTime] = useState(null)
  const [endTime, setEndTime] = useState(null)
  const [service, setService] = useState(null)
  const [instance, setInstance] = useState(null)
  const [traceId, setTraceId] = useState(null)
  const [from, setFrom] = useState(null)
  const [to, setTo] = useState(null)
  // const { instanceState } = useInstance()
  // useEffect(() => {
  //   const urlService = searchParams.get('service')
  //   const urlInstance = searchParams.get('instance')
  //   const urlTraceId = searchParams.get('traceId')
  //   const urlFrom = searchParams.get('logs-from')
  //   const urlTo = searchParams.get('logs-to')
  //   console.log(urlService, urlInstance, urlTraceId, urlFrom, urlTo, instanceState)
  //   if (urlService !== service) {
  //     setService(urlService)
  //   }
  //   if (urlInstance !== instance) {
  //     setInstance(urlInstance)
  //   }
  //   if (urlTraceId !== traceId) {
  //     setTraceId(urlTraceId)
  //   }
  //   // if (urlFrom !== from) {
  //   //   setFrom(urlFrom)
  //   // }
  //   // if (urlTo !== to) {
  //   //   setTo(urlTo)
  //   // }
  //   if (urlFrom && urlTo && (urlFrom !== from || urlTo !== to)) {
  //     const urlTimeRange = timeRangeList.find((item) => item.from === urlFrom && item.to === urlTo)
  //     if (urlTimeRange) {
  //       //说明是快速范围，根据rangetype 获取当前开始结束时间戳
  //       const { startTime, endTime } = getTimestampRange(urlTimeRange.rangeType)
  //       setStartTime(startTime)
  //       setEndTime(endTime)
  //     } else {
  //       //说明可能是精确时间，先判断是不是可以转化成微妙时间戳
  //       const startTimestamp = ISOToTimestamp(urlFrom)
  //       const endTimestamp = ISOToTimestamp(urlTo)
  //       if (startTimestamp && endTimestamp) {
  //         setStartTime(startTimestamp)
  //         setEndTime(endTimestamp)
  //       }
  //     }
  //     setFrom(urlFrom)
  //     setTo(urlTo)
  //   }
  // }, [searchParams])
  return (
    // <PropsProvider
    //   value={{
    //     startTime,
    //     endTime,
    //     service: service ?? '',
    //     instance: instance ?? '',
    //     traceId: traceId ?? '',
    //   }}
    // >
    <div
      style={{ width: '100%', overflow: 'hidden', height: 'calc(100vh - 150px)' }}
      className="text-xs"
    >
      <CTabs
        activeItemKey={activeItemKey}
        className="border-tab h-full flex flex-col"
        onChange={(key) => setActiveItemKey(key)}
      >
        <CTabList variant="tabs" className="flex-grow-0 flex-shrink-0 text-base">
          <CTab itemKey="faultSite">故障现场日志</CTab>
          <CTab itemKey="full">全量日志</CTab>
        </CTabList>
        <CTabContent className="flex-grow flex-shrink overflow-hidden">
          <CTabPanel className="p-3 h-full " itemKey="faultSite">
            {activeItemKey === 'faultSite' && <FaultSiteLogs />}
          </CTabPanel>
          <CTabPanel className="p-3 h-full" itemKey="full">
            {activeItemKey === 'full' && <Empty context="敬请期待" />}
            {/* <FullLogs logsList={logsList} /> */}
          </CTabPanel>
        </CTabContent>
      </CTabs>
    </div>
    // </PropsProvider>
  )
}
export default LogsPage
