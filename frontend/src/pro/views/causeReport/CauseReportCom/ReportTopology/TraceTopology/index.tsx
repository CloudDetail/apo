import { Button, Modal } from 'antd'
import { useState } from 'react'
import TraceTopologyBody from './TraceTopologyBody'
import { useTranslation } from 'react-i18next'

const TraceTopologyModal = ({ modalVisible, closeModal }) => {
  const { t } = useTranslation('pro/causeReport')
  const { t: ct } = useTranslation('common')
  return (
    <>
      <Modal
        title={t('topology.fullLinkFlamegraph')}
        open={modalVisible}
        onCancel={closeModal}
        destroyOnClose
        centered
        cancelText={ct('close')}
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
