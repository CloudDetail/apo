/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'
import QueryList from './QueryList'
import { Pagination } from 'antd'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import Histogram from './Histogram'
import ContextModal from './ContextModal'
import { useTranslation } from 'react-i18next' // 引入i18n

const LogQueryResult = () => {
  const { t } = useTranslation('oss/fullLogs') // 使用i18n
  const { pagination, updateLogsPagination, logs, tableInfo, query, loading } = useLogsContext()

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
      <QueryList logs={logs} openContextModal={openContextModal} loading={loading} />
      <Pagination
        defaultCurrent={1}
        total={pagination.total}
        current={pagination.pageIndex}
        pageSize={pagination.pageSize}
        className="flex-shrink-0 flex-grow-0 p-2"
        align="end"
        onChange={changePagination}
        showTotal={(total) => `${t('logQueryResult.totalText')} ${total} `}
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
