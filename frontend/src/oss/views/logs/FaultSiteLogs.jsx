/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useRef, useState } from 'react'
import { CTab, CTabList, CTabs, CRow, CCol, CCard } from '@coreui/react'
import { convertTime } from 'src/core/utils/time'
import { getLogContentApi, getLogPageListApi } from 'core/api/logs'
import { CustomSelect } from 'src/core/components/Select'
import LogContent from './component/LogContent'
import CustomPagination from 'src/core/components/Pagination/CustomPagination'
import Empty from 'src/core/components/Empty/Empty'
import LoadingSpinner from 'src/core/components/Spinner'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import { useTranslation } from 'react-i18next'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import LogsTraceFilter from 'src/oss/components/Filter/LogsTraceFilter'
import { useLogsTraceFilterContext } from 'src/oss/contexts/LogsTraceFilterContext'
import { useDebounce } from 'react-use'
import { useSelector } from 'react-redux'
// import DataSourceFilter from 'src/core/components/Filter/DataSourceFilter'
function FaultSiteLogs(props) {
  const { t, i18n } = useTranslation('oss/faultSiteLogs')

  const [logsPageList, setLogsPageList] = useState([])
  const [activeItemKey, setActiveItemKey] = useState(0)
  const [logContent, setLogContent] = useState({})
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(false)
  const [source, setSource] = useState('')
  const [logContentLoading, setLogContentLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const refs = useRef({})
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)

  //params for api

  const { clusterIds, services, instance, traceId, namespaces, startTime, endTime, isFilterDone } =
    useLogsTraceFilterContext((ctx) => ctx)
  const changeActiveItemKey = (key) => {
    setActiveItemKey(key)
    setSource('')
  }
  const filterUndefinedOrEmpty = (obj) => {
    return Object.fromEntries(
      Object.entries(obj).filter(([_, value]) => {
        if (value === undefined || value === null) return false
        if (typeof value === 'string' && value.trim() === '') return false
        if (Array.isArray(value) && value.length === 0) return false
        return true
      }),
    )
  }

  const getLogs = () => {
    setActiveItemKey(0)
    setLoading(true)
    const { containerId, node: nodeName, pid, id: instanceId } = instance?.[0] ?? {}
    const queryParams = filterUndefinedOrEmpty({
      startTime,
      endTime,
      clusterIds,
      service: services,
      instance: instanceId,
      traceId,
      pageNum: pageIndex,
      pageSize: 10,
      containerId,
      namespaces,
      nodeName,
      pid,
      groupId: dataGroupId,
    })
    getLogPageListApi(queryParams)
      .then((res) => {
        setLoading(false)
        setLogsPageList(res?.list ?? [])
        setTotal(res?.pagination.total)
      })
      .catch(() => {
        setLogsPageList([])
        setLoading(false)
      })
      .finally(() => {})
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
  useDebounce(
    () => {
      if (startTime && endTime && isFilterDone) {
        if (pageIndex === 1) {
          getLogs()
        } else {
          setPageIndex(1)
        }
      }
    },
    300,
    [
      clusterIds,
      services,
      instance,
      traceId,
      namespaces,
      startTime,
      endTime,
      isFilterDone,
      dataGroupId,
    ],
  )
  useEffect(() => {
    if (pageIndex > 0) {
      getLogs()
    }
  }, [pageIndex])
  useEffect(() => {
    getLogContent()
  }, [logsPageList, activeItemKey])
  useEffect(() => {
    if (source && source !== logContent.logContents.source) {
      getLogContent()
    }
  }, [source])
  useEffect(() => {
    if (logContent?.logContents?.contents?.length > 0 && !source) {
      setSource(logContent?.sources[0])
    }
  }, [logContent])
  return (
    <BasicCard>
      <LoadingSpinner loading={loading} />

      <BasicCard.Header>
        <div className="w-full flex justify-start items-center text-sm font-normal">
          <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
          {i18n.language === 'zh' ? (
            <>
              {t('faultSiteLogs.faultLogTableToast')}
              <a
                className="underline text-[var(--ant-color-link)] hover:text-[var(--ant-color-link-hover)] active:text-[var(--ant-color-link-active)]"
                target="_blank"
                href="https://kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/配置日志采集/配置故障日志采集"
              >
                <span>{t('faultSiteLogs.documentText')}</span>
              </a>
            </>
          ) : (
            <p className="my-0">
              {t('faultSiteLogs.faultLogTableToast1')}
              <a
                className="underline text-[var(--ant-color-link)] hover:text-[var(--ant-color-link-hover)] active:text-[var(--ant-color-link-active)]"
                target="_blank"
                href="https://docs.autopilotobservability.com/Logs%20Monitoring/Fault%20Log%20Collection"
              >
                <span>{t('faultSiteLogs.documentText')}</span>
              </a>
              {t('faultSiteLogs.faultLogTableToast2')}
            </p>
          )}
        </div>
      </BasicCard.Header>

      <BasicCard.Header>
        <div className="w-full flex-none">
          <LogsTraceFilter type="logs" />
          {/* <DataSourceFilter category="log" /> */}
        </div>
      </BasicCard.Header>

      <BasicCard.Table>
        <div className="h-full flex-grow flex-shrink overflow-hidden flex-column-tab ">
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
                          {convertTime(logs.startTime, 'yyyy-mm-dd hh:mm:ss.SSS')}{' '}
                          {t('faultSiteLogs.faultLogsText')}
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
                    <CCard className="mx-4 my-2 p-2 font-bold">
                      <CRow className="my-1 ">
                        <CCol sm="2" className="text-gray-400 font-bold">
                          Trace Id
                        </CCol>
                        <CCol sm="auto">{logsPageList[activeItemKey]?.traceId}</CCol>
                      </CRow>
                      <CRow className="my-1">
                        <CCol sm="2" className="text-gray-400 font-bold">
                          {t('faultSiteLogs.endpoint')}
                        </CCol>
                        <CCol sm="auto">{logsPageList[activeItemKey]?.endpoint}</CCol>
                      </CRow>
                      <CRow className="my-1">
                        <CCol sm="2" className="text-gray-400 font-bold">
                          {t('faultSiteLogs.timeOfFailure')}
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
                  <div className="text-base font-bold mb-2">
                    {t('faultSiteLogs.specificLogInformation')}
                  </div>
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
      </BasicCard.Table>
    </BasicCard>
  )
}
export default FaultSiteLogs
