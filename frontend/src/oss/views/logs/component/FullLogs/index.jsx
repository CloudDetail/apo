/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

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
      sessionStorage.setItem('fullLogs:siderCollapse', "true")
    } else {
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
            collapsible
            defaultSize={
              sessionStorage.getItem('fullLogs:siderCollapse') === "true" ? 0 : 300
            }
          >
            <FullLogSider />
          </Splitter.Panel>
          <Splitter.Panel collapsible defaultSize={300}>
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
