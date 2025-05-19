/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useState } from 'react'
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
import { Button, Layout, theme } from 'antd'
import Sider from 'antd/es/layout/Sider'
import { Content } from 'antd/es/layout/layout'
import { AiOutlineCaretLeft, AiOutlineCaretRight } from 'react-icons/ai'
import './index.css'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
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

  const [searchParams] = useSearchParams()
  const [collapsed, setCollapsed] = useState(false)
  const { useToken } = theme
  const { token } = useToken()
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
  return (
    <>
      <LoadingSpinner loading={loading} />
      {/* 顶部筛选 */}
      <CCard style={{ height: 'calc(100vh - 120px)' }}>
        <Layout className="relative ">
          <div
            onClick={() => setCollapsed(!collapsed)}
            className={`logSiderButton ${collapsed ? ' closeButton ' : 'openButton'}`}
          >
            {collapsed ? (
              <AiOutlineCaretRight color={token.colorPrimary} />
            ) : (
              <AiOutlineCaretLeft color={token.colorPrimary} />
            )}
          </div>
          <Sider
            trigger={null}
            collapsible
            collapsed={collapsed}
            className="p-2 "
            collapsedWidth={0}
            width={300}
          >
            <FullLogSider />
          </Sider>
          <Content className="h-full relative flex overflow-hidden px-2">
            <div className="w-[250px] h-full">
              <IndexList />
            </div>
            <div className="flex flex-col flex-1 overflow-hidden">
              <div className="flex-grow-0 flex-shrink-0">
                <SearchBar />
              </div>
              <div className="flex-1 h-full overflow-hidden">
                <LogQueryResult />
              </div>
            </div>
          </Content>
        </Layout>
      </CCard>
    </>
  )
}
export default FullLogs
