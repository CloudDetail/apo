import React, { useState } from 'react'
import AlertEventTable from './AlertEvent'
import { Button, Card, Splitter } from 'antd'
import { PlusCircleFilled, UnorderedListOutlined } from '@ant-design/icons'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import LoadingSpinner from 'src/core/components/Spinner'
import ImpactList from './ImpactList'
import { AiOutlineFundView } from 'react-icons/ai'
import AlertRuleQuickSet from './AlertRuleQuickSet'
import AnomalyDetectionRecord from './AnomalyDetectionRecord'

function AlertAnalyzeCom() {
  const [alertEventTableHeight, setAlertEventTableHeight] = useState(300)
  const { loading, isAnalyze, setIsAnalyze } = useAlertAnalyzeContext()
  const [modalVisibility, setModalVisibility] = useState(false)
  const [recordModalVisibility, setRecordModalVisibility] = useState(false)
  const openModal = () => {
    setModalVisibility(true)
  }
  const closeModal = () => {
    setModalVisibility(false)
  }
  const openRecordModal = () => {
    setRecordModalVisibility(true)
  }
  const closeRecordModal = () => {
    setRecordModalVisibility(false)
  }
  const onResize = (sizes) => {
    setAlertEventTableHeight(sizes[0])
  }
  return (
    <Card
      style={{ height: 'calc(100vh - 120px)', position: 'relative' }}
      classNames={{
        body: 'p-0 ',
      }}
    >
      <Button
        className="anomaly-button"
        icon={<UnorderedListOutlined style={{ color: '#5286E8' }} />}
        color="default"
        onClick={openRecordModal}
        style={{
          position: 'absolute',
          right: '10px',
          top: '10px',
          color: '#5286E8',
          borderColor: '#5286E8',
        }}
      >
        异常检测告警记录
      </Button>
      <Button
        className="anomaly-button"
        icon={<PlusCircleFilled style={{ color: '#5286E8' }} />}
        onClick={openModal}
        style={{
          position: 'absolute',
          right: '190px',
          top: '10px',
          color: '#5286E8',
          borderColor: '#5286E8',
        }}
      >
        异常检测配置
      </Button>
      <LoadingSpinner loading={loading} />
      <Splitter
        style={{
          height: 'calc(100vh - 120px)',
          boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)',
          marginTop: '10px',
          // border: '1px solid rgb(229, 231, 235)',
        }}
        layout="vertical"
        onResize={onResize}
      >
        <Splitter.Panel collapsible defaultSize={'50%'}>
          <AlertEventTable scrolllHeight={alertEventTableHeight} />
        </Splitter.Panel>
        <Splitter.Panel collapsible>
          {isAnalyze ? (
            <ImpactList />
          ) : (
            <div className="text-center flex flex-col items-center justify-center w-full h-full">
              <AiOutlineFundView size={120} color="#1677ff" />
              <Button type="primary" size="large" onClick={() => setIsAnalyze(true)}>
                一键分析
              </Button>
              <div className="text-sm p-2">
                点击一键分析所有受告警影响的业务入口，分析时间可能较长，感谢您的耐心等待。
              </div>
            </div>
          )}
        </Splitter.Panel>
      </Splitter>
      <AlertRuleQuickSet modalVisibility={modalVisibility} closeModal={closeModal} />
      <AnomalyDetectionRecord
        modalVisibility={recordModalVisibility}
        closeListModal={closeRecordModal}
      />
    </Card>
  )
}
export default AlertAnalyzeCom
