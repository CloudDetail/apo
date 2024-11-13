import { Modal, Skeleton } from 'antd'
import React, { useEffect, useState } from 'react'
import { getLogContextApi } from 'core/api/logs'
import QueryList from '../QueryList'
import LoadingSpinner from 'src/core/components/Spinner'

const ContextModal = ({ modalVisible, closeModal, logParams }) => {
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
      title={'上下文'}
      open={modalVisible}
      onCancel={closeContextModal}
      destroyOnClose
      centered
      cancelText="关闭"
      width={1000}
      bodyStyle={{ maxHeight: '80vh', overflowY: 'auto', overflowX: 'hidden' }}
      footer={(_, { CancelBtn }) => (
        <>
          <CancelBtn />
        </>
      )}
    >
      <div className="h-[650px]">
        <LoadingSpinner loading={loading} />
        <QueryList logs={context} loading={loading} />
      </div>
    </Modal>
  )
}
export default ContextModal
