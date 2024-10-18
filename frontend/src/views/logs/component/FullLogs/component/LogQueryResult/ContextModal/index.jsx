import { Modal, Skeleton } from 'antd'
import React, { useEffect, useState } from 'react'
import { getLogContextApi } from 'src/api/logs'
import QueryList from '../QueryList'

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
  useEffect(() => {
    if (modalVisible && logParams) getLogContext()
  }, [modalVisible, logParams])
  return (
    <Modal
      title={'上下文'}
      open={modalVisible}
      onCancel={closeModal}
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
      <Skeleton loading={loading}>
        <QueryList logs={context} />
      </Skeleton>
    </Modal>
  )
}
export default ContextModal
