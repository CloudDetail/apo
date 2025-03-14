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
import { useTranslation } from 'react-i18next'
import { ChartsProvider } from 'src/core/contexts/ChartsContext'

function AlertAnalyzeCom() {
  const { t } = useTranslation('pro/alertAnalyze')
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
      style={{ position: 'relative' }}
      classNames={{
        body: 'p-0 ',
      }}
      title={t(`alertEvent.alertEvents`)}
      extra={
        <>
          {' '}
          <Button
            className="anomaly-button"
            icon={<UnorderedListOutlined style={{ color: '#5286E8' }} />}
            color="default"
            onClick={openRecordModal}
            style={{
              color: '#5286E8',
              borderColor: '#5286E8',
            }}
          >
            {t('anomalyRecord')}
          </Button>
          <Button
            className="anomaly-button ml-2"
            icon={<PlusCircleFilled style={{ color: '#5286E8' }} />}
            onClick={openModal}
            style={{
              color: '#5286E8',
              borderColor: '#5286E8',
            }}
          >
            {t('anomalyConfig')}
          </Button>
        </>
      }
      styles={{ body: { height: 'calc(100vh - 180px)' } }}
    >
      <LoadingSpinner loading={loading} />
      <Splitter
        style={{
          height: 'calc(100vh - 180px)',
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
            <ChartsProvider>
              <ImpactList />
            </ChartsProvider>
          ) : (
            <div className="text-center flex flex-col items-center justify-center w-full h-full">
              <AiOutlineFundView size={120} color="#1677ff" />
              <Button type="primary" size="large" onClick={() => setIsAnalyze(true)}>
                {t('oneClickAnalyze')}
              </Button>
              <div className="text-sm p-2">{t('analyzeInfo')}</div>
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
