import React from 'react'
import QueryList from './QueryList'
import { Pagination } from 'antd'
import { useLogsContext } from 'src/contexts/LogsContext'
import Histogram from './Histogram'

const LogQueryResult = () => {
  const { pagination, updateLogsPagination } = useLogsContext()
  const changePagination = (page, pageSize) => {
    updateLogsPagination({
      pageSize: pageSize,
      pageIndex: page,
      total: pagination.total,
    })
  }
  return (
    <div className="overflow-hidden flex flex-col h-full">
      <Histogram />
      <QueryList />
      <Pagination
        defaultCurrent={1}
        total={pagination.total}
        current={pagination.pageIndex}
        pageSize={pagination.pageSize}
        className="flex-shrink-0 flex-grow-0 p-2"
        align="end"
        onChange={changePagination}
        showTotal={(total) => `日志总条数: ${total} `}
      />
    </div>
  )
}
export default LogQueryResult
