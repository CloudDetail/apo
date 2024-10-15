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
import FullLogSider from './component/Sider'
import { Button, Layout } from 'antd'
import Sider from 'antd/es/layout/Sider'
import { Content } from 'antd/es/layout/layout'
import { AiOutlineCaretLeft, AiOutlineCaretRight } from 'react-icons/ai'
import './index.css'
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
    [searchParams.get('log-from'), searchParams.get('log-to'), tableInfo, query],
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
              <AiOutlineCaretRight color="#1a83fe" />
            ) : (
              <AiOutlineCaretLeft color="#1a83fe" />
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
          <Content className="h-full relative p-2">
            <div className="flex flex-col flex-1 h-full">
              <div className="flex-grow-0 flex-shrink-0">
                <SearchBar />
              </div>
              <div className="flex-1 flex overflow-hidden">
                <div className="w-[220px] flex-shrink-0 flex-grow-0">
                  <IndexList />
                </div>
                <div className=" h-full flex-1 overflow-hidden">
                  <LogQueryResult />
                </div>
              </div>
            </div>
          </Content>
        </Layout>
      </CCard>
    </>
  )
}
export default FullLogs
