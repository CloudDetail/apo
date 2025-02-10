import React, { useEffect, useRef, useState } from 'react'
import { CTab, CTabList, CTabs, CRow, CCol, CCard, CToast, CToastBody } from '@coreui/react'
import { convertTime } from 'src/core/utils/time'
import { getLogContentApi, getLogPageListApi } from 'core/api/logs'
import { CustomSelect } from 'src/core/components/Select'
import LogContent from './LogContent'
import Empty from 'src/core/components/Empty/Empty'
import LoadingSpinner from 'src/core/components/Spinner'
import { Button, Pagination } from 'antd'
import { useUpdateEffect } from 'react-use'
function LogsTabs(props) {
  const { logsPageList, changePagination, showFaultSite = true } = props

  const [activeItemKey, setActiveItemKey] = useState(0)
  const [logContent, setLogContent] = useState({})
  const [pageIndex, setPageIndex] = useState(1)
  const [logContentLoading, setLogContentLoading] = useState(false)
  const refs = useRef({})
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
  useEffect(() => {
    if (activeItemKey === 0) {
      getLogContent()
    } else {
      setActiveItemKey(0)
    }
  }, [logsPageList])
  useUpdateEffect(() => {
    getLogContent()
  }, [activeItemKey])
  useEffect(() => {
    if (source && source !== logContent.logContents.source) {
      getLogContent()
    }
  }, [source])
  const onChangePagination = (page) => {
    if (changePagination) {
      changePagination(page)
    } else {
      setPageIndex(page)
    }
  }
  return (
    <>
      {logsPageList?.length > 0 && (
        <CTabs
          key={pageIndex + activeItemKey}
          activeItemKey={activeItemKey}
          className="flex flex-row h-full logs-tab text-xs"
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
              {showFaultSite && (
                <Pagination
                  defaultCurrent={1}
                  total={props.total || logsPageList.length}
                  current={props.pageIndex || pageIndex}
                  className="flex-shrink-0 flex-grow-0 p-2"
                  align="end"
                  onChange={onChangePagination}
                  showSizeChanger={false}
                  size="small"
                  simple
                />
              )}
            </div>
          </CTabList>

          {logsPageList[activeItemKey] && (
            <div className="p-3 w-full h-full overflow-hidden flex flex-col relative">
              <LoadingSpinner loading={logContentLoading} />
              {showFaultSite && (
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
              )}
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
    </>
  )
}
export default LogsTabs
