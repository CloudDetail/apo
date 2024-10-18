import React, { useEffect, useState } from 'react'
import QueryList from './QueryList'
import { Pagination } from 'antd'
import { useLogsContext } from 'src/contexts/LogsContext'
import Histogram from './Histogram'
import ContextModal from './ContextModal'

const LogQueryResult = () => {
  const { pagination, updateLogsPagination, logs, tableInfo, query } = useLogsContext()

  const [modalVisible, setModalVisible] = useState(false)
  const [contextLogParams, setContextLogParams] = useState(null)
  const openContextModal = (logInfo) => {
    setModalVisible(true)
    setContextLogParams({
      dataBase: tableInfo.dataBase,
      tableName: tableInfo.tableName,
      tags: logInfo.tags,
      timestamp: logInfo.timestamp,
    })
  }
  const closeContextModal = () => {
    setModalVisible(false)
    setContextLogParams(null)
  }
  const changePagination = (page, pageSize) => {
    updateLogsPagination({
      pageSize: pageSize,
      pageIndex: page,
      total: pagination.total,
    })
  }
  useEffect(() => {
    if (modalVisible) closeContextModal()
  }, [query])
  return (
    <div className="overflow-hidden flex flex-col h-full">
      <Histogram />
      <QueryList logs={logs} openContextModal={openContextModal} />
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
      <ContextModal
        modalVisible={modalVisible}
        closeModal={closeContextModal}
        logParams={contextLogParams}
      />
    </div>
  )
}
export default LogQueryResult
