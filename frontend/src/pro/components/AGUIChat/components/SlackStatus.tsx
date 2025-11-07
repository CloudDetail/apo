import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Tooltip } from 'antd'
import { useTranslation } from 'react-i18next'
import { getSlackStatusApi } from 'src/core/api/appConfig'
import { useDeployMode } from 'src/core/contexts/DeployModeContext'
import slackIcon from 'src/core/assets/images/slack.png'
import styles from './DataStatus.module.scss'

interface SlackStatusResponse {
  status?: string
  teamId?: string
  createDate?: number
}

type NormalizedStatus = 'connected' | 'waiting' | 'connectRetry' | 'disconnected' | 'notConnected'

const indicatorClassMap: Record<NormalizedStatus, string> = {
  connected: styles.loader,
  waiting: styles.loader,
  connectRetry: styles.errorLoader,
  disconnected: styles.errorLoader,
  notConnected: styles.errorLoader,
}

const normalizeStatus = (value?: string): NormalizedStatus => {
  switch (value?.toLowerCase()) {
    case 'connected':
      return 'connected'
    case 'waiting':
      return 'waiting'
    case 'connect retry':
      return 'connectRetry'
    case 'disconnected':
      return 'disconnected'
    default:
      return 'notConnected'
  }
}

const SlackStatus: React.FC = () => {
  const { t } = useTranslation('pro/agent')
  const { deployMode } = useDeployMode()
  const [slackStatus, setSlackStatus] = useState<SlackStatusResponse | null>(null)
  const [loading, setLoading] = useState<boolean>(false)

  const fetchStatus = useCallback(async () => {
    try {
      setLoading(true)
      const response = (await getSlackStatusApi()) as unknown as SlackStatusResponse
      setSlackStatus(response)
    } catch (error) {
      console.error('Failed to fetch Slack status', error)
      setSlackStatus(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchStatus()
  }, [fetchStatus])

  const normalizedStatus = useMemo<NormalizedStatus>(
    () => normalizeStatus(slackStatus?.status),
    [slackStatus?.status],
  )

  const statusText = useMemo(() => {
    if (loading) {
      return t('slackStatus.loading')
    }
    return t(`slackStatus.states.${normalizedStatus}`)
  }, [loading, normalizedStatus, t])

  const indicatorClassName = loading
    ? styles.loader
    : indicatorClassMap[normalizedStatus] || styles.loader

  const handleNavigate = useCallback(() => {
    const target = deployMode === 'syncause' ? '/probe-management' : '/#/integration/data?'
    window.open(target, '_blank', 'noopener')
  }, [deployMode])

  const tooltipContent = (
    <div className="max-w-[220px]">
      <div className="text-xs font-semibold mb-1">{t('slackStatus.tooltip.title')}</div>
      <div className="text-xs mb-1">{statusText}</div>
      {slackStatus?.teamId && (
        <div className="text-xs text-[var(--ant-color-text-tertiary)] break-all">
          {t('slackStatus.tooltip.teamId')}: {slackStatus.teamId}
        </div>
      )}
      <div className="text-[10px] text-[var(--ant-color-text-tertiary)] mt-2">
        {t('slackStatus.tooltip.navigate')}
      </div>
    </div>
  )

  return (
    <Tooltip title={tooltipContent} placement="top">
      <div
        className="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-black/10 cursor-pointer transition-colors"
        role="button"
        tabIndex={0}
        onClick={handleNavigate}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault()
            handleNavigate()
          }
        }}
      >
        <span className={indicatorClassName} style={{ margin: 0 }} />
        <img src={slackIcon} alt="Slack" className="w-3 h-3" />
        <span className="text-xs text-[var(--ant-color-text-tertiary)]">
          {t('slackStatus.label')}
        </span>
      </div>
    </Tooltip>
  )
}

export default SlackStatus
