import { Button, Modal } from 'antd'
import { useState } from 'react'
import TraceTopologyBody from './TraceTopologyBody'

const TraceTopologyModal = ({ modalVisible, closeModal }) => {
  return (
    <>
      <Modal
        title="全链路调用火焰图"
        open={modalVisible}
        onCancel={closeModal}
        destroyOnClose
        centered
        cancelText="关闭"
        width={1000}
        className="trace_topology_modal"
        footer={(_, { CancelBtn }) => (
          <>
            <CancelBtn />
          </>
        )}
      >
        <div className="h-[700px] pb-5 overflow-auto">
          <TraceTopologyBody />
        </div>
      </Modal>
    </>
  )
}
export default TraceTopologyModal
