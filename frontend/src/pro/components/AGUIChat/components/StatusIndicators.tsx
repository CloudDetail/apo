import React from 'react'
import { Button } from 'antd'
import { WarningOutlined, CloseOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import styles from '../AGUIChat.module.scss'

interface StatusIndicatorsProps {
  error: string | null
  isConnected: boolean
  currentStep: string | null
  isRunning: boolean
  runId: string | null
  onClearError: () => void
}

const StatusIndicators: React.FC<StatusIndicatorsProps> = ({
  error,
  isConnected,
  currentStep,
  isRunning,
  runId,
  onClearError,
}) => {
  const { t } = useTranslation('pro/agent')

  // 渲染连接状态
  const renderConnectionStatus = () => {
    if (isConnected) return null

    return (
      <div className={styles.connectionStatus}>
        <div className={`${styles.connectionIndicator} ${styles.error}`}>
          <WarningOutlined className={styles.connectionIcon} />
          <span className={styles.connectionText}>{t('status.connectionDisconnected')}</span>
        </div>
      </div>
    )
  }

  // 渲染错误提示
  const renderErrorAlert = () => {
    if (!error) return null

    return (
      <div className={styles.errorAlert}>
        {error}
        <Button
          type="text"
          icon={<CloseOutlined />}
          onClick={onClearError}
          className={styles.errorClose}
        />
      </div>
    )
  }

  // 渲染步骤状态
  const renderStepStatus = () => {
    if (!currentStep) return null

    return (
      <div className={styles.stepStatus}>
        <div className={styles.stepIndicator}>
          <div className={styles.stepDot}></div>
          <div className={styles.stepText}>
            {t('status.currentStep')}: {currentStep}
          </div>
        </div>
      </div>
    )
  }

  // 渲染运行状态
  const renderRunStatus = () => {
    if (!isRunning) return null

    return (
      <div className={styles.runStatus}>
        <div className={styles.runIndicator}>
          <div className={styles.runSpinner}></div>
          <div className={styles.runText}>{t('status.running')} </div>
        </div>
      </div>
    )
  }

  return (
    <>
      {renderErrorAlert()}
      {renderConnectionStatus()}
      {renderStepStatus()}
      {renderRunStatus()}
    </>
  )
}

export default StatusIndicators
