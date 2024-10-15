import React, { useEffect, useRef, useState } from 'react'
import { CTab, CTabList, CTabs, CRow, CCol, CCard, CToast, CToastBody } from '@coreui/react'
import { convertTime } from 'src/utils/time'
import { getLogContentApi, getLogPageListApi } from 'src/api/logs'
import { useSearchParams } from 'react-router-dom'
import { CustomSelect } from 'src/components/Select'
import LogContent from './component/LogContent'
import CustomPagination from 'src/components/Pagination/CustomPagination'
import Empty from 'src/components/Empty/Empty'
import LoadingSpinner from 'src/components/Spinner'
import LogsTraceFilter from 'src/components/Filter/LogsTraceFilter'
import { useSelector } from 'react-redux'
import { IoMdInformationCircleOutline } from 'react-icons/io'
function FaultSiteLogs(props) {
  const { startTime, endTime, service, instance, traceId, instanceOption } = useSelector(
    (state) => state.urlParamsReducer,
  )

  const [searchParams, setSearchParams] = useSearchParams()
  const [logsPageList, setLogsPageList] = useState([])
  const [activeItemKey, setActiveItemKey] = useState(0)
  const [logContent, setLogContent] = useState({})
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(true)
  const [logContentLoading, setLogContentLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const refs = useRef({})
  const previousValues = useRef({
    startTime: null,
    endTime: null,
    service: '',
    instance: '',
    traceId: '',
    pageIndex: 1,
    selectInstanceOption: {},
  })
  const [source, setSource] = useState('')
  const changeActiveItemKey = (key) => {
    setActiveItemKey(key)
    setSource('')
  }

  useEffect(() => {
    if (logContent?.logContents?.contents?.length > 0 && !source) {
      setSource(logContent?.sources[0])
    }
  }, [logContent])
  const getLogs = () => {
    setActiveItemKey(0)
    setLoading(true)
    const { containerId, nodeName, pid } = instanceOption[instance] ?? {}
    getLogPageListApi({
      startTime,
      endTime,
      service: service,
      // instance: instance,
      traceId: traceId,
      pageNum: pageIndex,
      pageSize: 10,
      containerId,
      nodeName,
      pid,
    })
      .then((res) => {
        setLoading(false)
        setLogsPageList(res?.list ?? [])
        setTotal(res?.pagination.total)

        //
      })
      .catch(() => {
        setLogsPageList([])
        setLoading(false)
      })
      .finally(() => {})
    // 你的getLogs实现
  }
  const getLogContent = () => {
    const params = logsPageList?.[activeItemKey]
    if (params) {
      setLogContentLoading(true)
      getLogContentApi({
        ...params,
        sourceFrom: source,
        startTime: params.startTime - 1000000,
        endTime: params.endTime + 1000000,
      })
        .then((res) => {
          setLogContentLoading(false)
          setLogContent(res)
        })
        .catch(() => {
          setLogContentLoading(false)
          setLogContent({})
        })
    }
  }
  // useEffect(() => {
  //   console.log(pageIndex)
  //   if (pageIndex > 0) {
  //     getLogs()
  //   }
  // }, [pageIndex])
  useEffect(() => {
    const prev = previousValues.current
    let paramsChange = false

    if (prev.startTime !== startTime) {
      console.log('startTime -> pre:', prev.startTime, 'now:', startTime)
      paramsChange = true
    }
    if (prev.endTime !== endTime) {
      console.log('endTime -> pre:', prev.endTime, 'now:', endTime)
      paramsChange = true
    }
    if (prev.service !== service) {
      console.log('service -> pre:', prev.service, 'now:', service)
      paramsChange = true
    }

    if (prev.traceId !== traceId) {
      console.log('traceId -> pre:', prev.traceId, 'now:', traceId)
      paramsChange = true
    }
    const selectInstanceOption = instanceOption[instance]
    if (JSON.stringify(prev.selectInstanceOption) !== JSON.stringify(selectInstanceOption)) {
      console.log(
        'selectInstanceOption -> pre:',
        prev.selectInstanceOption,
        'now:',
        selectInstanceOption,
      )
      paramsChange = true
    }
    if (instance && !selectInstanceOption) {
      paramsChange = false
    }
    console.log(
      '-----------',
      paramsChange,
      startTime,
      endTime,
      service,
      instance,
      traceId,
      pageIndex,
      instanceOption,
    )

    previousValues.current = {
      startTime,
      endTime,
      service,
      instance,
      traceId,
      pageIndex,
      selectInstanceOption,
    }
    if (startTime && endTime) {
      if (paramsChange) {
        if (pageIndex === 1) {
          getLogs()
        } else {
          setPageIndex(1)
        }
      } else if (prev.pageIndex !== pageIndex) {
        getLogs()
      }
    }
  }, [startTime, endTime, service, instance, traceId, pageIndex])
  useEffect(() => {
    getLogContent()
  }, [logsPageList, activeItemKey])
  useEffect(() => {
    if (source && source !== logContent.logContents.source) {
      getLogContent()
    }
  }, [source])
  return (
    <CCard
      className="h-full flex flex-col overflow-hidden text-xs px-2"
      style={{ height: 'calc(100vh - 120px)' }}
    >
      <LoadingSpinner loading={loading} />
      <CToast autohide={false} visible={true} className="align-items-center w-full my-2">
        <div className="d-flex">
          <CToastBody className=" flex flex-row items-center text-xs">
            <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
            默认采集控制台日志，从文件中采集日志的配置方式请参考
            <a
              className="underline text-sky-500"
              target="_blank"
              href="https://originx.kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/配置采集日志/"
            >
              文档
            </a>
          </CToastBody>
        </div>
      </CToast>
      <div className="flex-grow-0 flex-shrink-0">
        <LogsTraceFilter type="logs" />
      </div>
      <div className="flex-grow flex-shrink overflow-hidden flex-column-tab ">
        {logsPageList?.length > 0 && (
          <CTabs
            key={pageIndex + activeItemKey}
            activeItemKey={activeItemKey}
            className="flex flex-row h-full logs-tab"
            onChange={changeActiveItemKey}
          >
            <CTabList variant="tabs" className="flex-col w-[200px] shrink-0 flex-nowrap">
              <div className="overflow-y-auto w-full overflow-x-hidden flex-1">
                {logsPageList &&
                  logsPageList.map((logs, index) => {
                    return (
                      <CTab
                        itemKey={index}
                        key={index}
                        ref={(el) => (refs.current[index] = el)}
                        onClick={() => changeActiveItemKey(index)}
                      >
                        {convertTime(logs.startTime, 'yyyy-mm-dd hh:mm:ss.SSS')} 故障现场日志
                      </CTab>
                    )
                  })}
              </div>

              <div className="w-full overflow-hidden flex-grow-0 flex items-end justify-end">
                <CustomPagination
                  pageIndex={pageIndex}
                  pageSize={pageSize}
                  total={total}
                  previousPage={() => setPageIndex(pageIndex - 1)}
                  nextPage={() => setPageIndex(pageIndex + 1)}
                  gotoPage={(index) => setPageIndex(index)}
                  maxButtons={2}
                />
              </div>
            </CTabList>

            {logsPageList[activeItemKey] && (
              <div className="p-3 w-full h-full overflow-hidden flex flex-col relative">
                <LoadingSpinner loading={logContentLoading} />
                <div className="flex-grow-0 flex-shrink-0">
                  <div className="text-base font-bold">故障现场</div>

                  <CCard className="mx-4 my-2 p-2 font-bold">
                    <CRow className="my-1 ">
                      <CCol sm="2" className="text-gray-400 font-bold">
                        Trace Id
                      </CCol>
                      <CCol sm="auto">{logsPageList[activeItemKey]?.traceId}</CCol>
                    </CRow>
                    <CRow className="my-1">
                      <CCol sm="2" className="text-gray-400 font-bold">
                        服务端点
                      </CCol>
                      <CCol sm="auto">{logsPageList[activeItemKey]?.endpoint}</CCol>
                    </CRow>
                    <CRow className="my-1">
                      <CCol sm="2" className="text-gray-400 font-bold">
                        故障发生时间
                      </CCol>
                      <CCol sm="auto">
                        {convertTime(
                          logsPageList[activeItemKey]?.startTime,
                          'yyyy-mm-dd hh:mm:ss.SSS',
                        )}
                      </CCol>
                    </CRow>
                  </CCard>
                </div>
                <div className="text-base font-bold mb-2">具体日志信息</div>
                <div className="flex flex-row items-center">
                  <span className="text-nowrap">Source：</span>
                  <CustomSelect
                    options={logContent?.sources ?? []}
                    value={source}
                    onChange={(value) => setSource(value)}
                  />
                </div>
                <LogContent data={logContent} />
              </div>
            )}
          </CTabs>
        )}
        {(!logsPageList || logsPageList?.length === 0) && <Empty />}
      </div>
    </CCard>
  )
}
export default FaultSiteLogs
