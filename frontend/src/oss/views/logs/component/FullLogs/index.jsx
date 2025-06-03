/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef } from 'react'
import { CCard } from '@coreui/react'
import { getFullLogApi, getFullLogChartApi } from 'core/api/logs'
import { useSearchParams } from 'react-router-dom'
import { ISOToTimestamp } from 'src/core/utils/time'
import LoadingSpinner from 'src/core/components/Spinner'
import SearchBar from './component/SerarchBar'
import IndexList from './component/IndexList'
import LogQueryResult from './component/LogQueryResult'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { useDebounce, useUpdateEffect } from 'react-use'
import FullLogSider from './component/Sider'
import { Splitter } from 'antd'
import { Content } from 'antd/es/layout/layout'
import './index.css'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import { useState } from 'react'
import { AiOutlineCaretLeft, AiOutlineCaretRight } from 'react-icons/ai'
function FullLogs() {
  const {
    query,
    pagination,
    fetchData,
    loading,
    clearFieldIndexMap,
    updateLogsPagination,
    tableInfo,
  } = useLogsContext()

const [siderSize, setSiderSize] = useState(sessionStorage.getItem('fullLogs:siderCollapse') === "true" ? 0 : 300)

  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  useUpdateEffect(() => {
    if (startTime && endTime) {
      fetchData({ startTime, endTime })
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
      if (startTime && endTime) {
        if (pagination.pageIndex === 1) {
          fetchData({ startTime, endTime })
        } else {
          updateLogsPagination({
            pageIndex: 1,
          })
        }
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, tableInfo, query],
  )

  // Record the collapse state of the FullLogSider
  const handleResize = (sizes) => {
    if (sizes[0] === 0) {
      // setSideCollapse(true)
      setSiderSize(0)
      sessionStorage.setItem('fullLogs:siderCollapse', "true")
    } else {
      setSiderSize(sizes[0])
      sessionStorage.setItem('fullLogs:siderCollapse', "false")
    }
  }

  return (
    <BasicCard
      bodyStyle={{ padding: 0 }}
    >
      <LoadingSpinner loading={loading} />

      <Splitter onResize={handleResize}>
        <Splitter.Panel
          // collapsible
          defaultSize={
            sessionStorage.getItem('fullLogs:siderCollapse') === "true" ? 0 : 300
          }
          className='relative text-[var(--ant-color-primary)]'
          size={siderSize}
        >
          <FullLogSider />
          {siderSize && <div
            onClick={() => {setSiderSize(0); sessionStorage.setItem('fullLogs:siderCollapse', "true");}}
            className='logSiderButton closeButton'
          >
            <AiOutlineCaretLeft />
          </div>}
        </Splitter.Panel>
        {/* <Content className="h-full relative flex overflow-hidden px-2"> */}
        <Splitter.Panel
          defaultSize={300}
        >
          {!siderSize && <div
            onClick={() => {setSiderSize(300); sessionStorage.setItem('fullLogs:siderCollapse', "false");}}
            className={`relative text-[var(--ant-color-primary)] logSiderButton openButton`}
          >
            <AiOutlineCaretRight />
          </div>}
          <div className="h-full px-2">
            <IndexList />
          </div>
        </Splitter.Panel>
        <Splitter.Panel>
          <Content className="h-full relative flex overflow-hidden px-2">
            <div className="flex flex-col flex-1 overflow-hidden ">
              <div className="flex-grow-0 flex-shrink-0">
                <SearchBar />
              </div>
              <div className="flex-1 h-full overflow-hidden">
                <LogQueryResult />
              </div>
            </div>
          </Content>
        </Splitter.Panel>
      </Splitter>
    </BasicCard>
  )
}
export default FullLogs
