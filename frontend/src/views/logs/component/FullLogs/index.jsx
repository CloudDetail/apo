import React, { useEffect, useRef, useState } from 'react'
import { CCard } from '@coreui/react'
import { getFullLogApi, getFullLogChartApi } from 'src/api/logs'
import { useSearchParams } from 'react-router-dom'
import { ISOToTimestamp } from 'src/utils/time'
import LoadingSpinner from 'src/components/Spinner'
import SearchBar from './component/SerarchBar'
import IndexList from './component/IndexList'
import LogQueryResult from './component/LogQueryResult'
import { useLogsContext } from 'src/contexts/LogsContext'
import { useDebounce, useUpdateEffect } from 'react-use'
function FullLogs() {
  const { query, pagination, fetchData, loading, clearFieldIndexMap, updateLogsPagination } =
    useLogsContext()

  const [searchParams] = useSearchParams()

  useUpdateEffect(() => {
    if (searchParams.get('log-from') && searchParams.get('log-to')) {
      fetchData({
        startTime: ISOToTimestamp(searchParams.get('log-from')),
        endTime: ISOToTimestamp(searchParams.get('log-to')),
      })
    }
  }, [
    pagination.pageIndex,
    pagination.pageSize,
    //先隐藏 后续加上字段筛选了再放开，目前只支持搜索按钮和初始化
    // query,
  ])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      clearFieldIndexMap()
      if (searchParams.get('log-from') && searchParams.get('log-to')) {
        if (pagination.pageIndex === 1) {
          fetchData({
            startTime: ISOToTimestamp(searchParams.get('log-from')),
            endTime: ISOToTimestamp(searchParams.get('log-to')),
          })
        } else {
          updateLogsPagination({
            pageIndex: 1,
          })
        }
      }
    },
    300, // 延迟时间 300ms
    [searchParams.get('log-from'), searchParams.get('log-to')],
  )
  return (
    <>
      <LoadingSpinner loading={loading} />
      {/* 顶部筛选 */}
      <CCard style={{ height: 'calc(100vh - 120px)' }} className="flex flex-col ">
        <div className="flex-grow-0 flex-shrink-0">
          <SearchBar />
        </div>
        <div className="flex-1 flex overflow-hidden">
          <div className="w-[220px] flex-shrink-0 flex-grow-0">
            <IndexList />
          </div>
          <div className=" h-full flex-1">
            <LogQueryResult />
          </div>
        </div>
      </CCard>
    </>
  )
}
export default FullLogs
