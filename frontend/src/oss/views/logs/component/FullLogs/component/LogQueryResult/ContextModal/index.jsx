/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Modal, Skeleton } from 'antd'
import React, { useEffect, useState } from 'react'
import { getLogContextApi } from 'core/api/logs'
import QueryList from '../QueryList'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next' // 引入i18n

const ContextModal = ({ modalVisible, closeModal, logParams }) => {
  const { t } = useTranslation('oss/fullLogs') // 使用i18n
  const [context, setContext] = useState([])
  const [loading, setLoading] = useState(false)
  const getLogContext = () => {
    setLoading(true)
    getLogContextApi(logParams)
      .then((res) => {
        const back = res?.back ?? []
        const front = res?.front ?? []
        setContext(front.concat(back))
      })
      .catch(() => {
        setContext([])
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const closeContextModal = () => {
    closeModal()
    setContext([])
  }
  useEffect(() => {
    if (modalVisible && logParams) getLogContext()
  }, [modalVisible, logParams])
  return (
    <Modal
      title={t('contextModal.contextText')}
      open={modalVisible}
      onCancel={closeContextModal}
      destroyOnClose
      centered
      cancelText={t('contextModal.cancelText')}
      width={1000}
      bodyStyle={{ height: '80vh', overflowY: 'auto', overflowX: 'hidden' }}
      footer={(_, { CancelBtn }) => (
        <>
          <CancelBtn />
        </>
      )}
    >
      <div className="h-full overflow-hidden ">
        <LoadingSpinner loading={loading} />
        <QueryList logs={context} loading={loading} />
      </div>
    </Modal>
  )
}
export default ContextModal
