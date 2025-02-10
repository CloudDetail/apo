import React, { useState } from 'react'
import DetectionRecord from './component/DetectionRecord'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import AlertRules from './component/AlertRule'
import { getStep } from 'src/core/utils/step'
import { Modal, Splitter } from 'antd'

const AnomalyDetectionRecord = ({ modalVisibility, closeListModal }) => {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [refresh, setRefresh] = useState(false)
  const step = getStep(startTime, endTime)
  return (
    <Modal
      title="异常检测告警记录"
      open={modalVisibility}
      onCancel={closeListModal}
      maskClosable={false}
      centered
      width={'95vw'}
      destroyOnClose
      footer={null}
    >
      <Splitter
        layout="vertical"
        style={{
          height: '85vh',
          boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)',
        }}
      >
        <Splitter.Panel style={{ marginTop: '20px' }} defaultSize={'50%'} collapsible>
          <DetectionRecord refresh={refresh} />
        </Splitter.Panel>
        <Splitter.Panel style={{ marginTop: '20px' }} collapsible>
          <AlertRules
            startTime={startTime}
            endTime={endTime}
            step={step}
            setRefresh={setRefresh}
            refresh={refresh}
          />
        </Splitter.Panel>
      </Splitter>
    </Modal>
  )
}

export default AnomalyDetectionRecord
