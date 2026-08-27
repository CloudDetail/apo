/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Input, Typography, message } from 'antd'
import { ExternalLink, Link, RotateCcw } from 'lucide-react'
import React, { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { getSlackStatusApi, upsertAppConfigApi } from 'src/core/api/appConfig'
import slackPng from 'src/core/assets/images/slack.png'
import styles from './botManagement.module.scss'
// 类型定义
interface SlackStatus {
  teamId: string
  createDate: number
  status: string
}

interface UpsertResponse {
  success: boolean
  message: string
}

const BotManagement = () => {
  const { t } = useTranslation('core/dataIntegration')
  const [teamId, setTeamId] = useState('')
  const [slackStatus, setSlackStatus] = useState<SlackStatus>({
    teamId: '',
    createDate: 0,
    status: '',
  })
  const [connecting, setConnecting] = useState(false)
  const [countdown, setCountdown] = useState(5)
  const pollingRef = useRef<number | null>(null)
  const countdownRef = useRef<number | null>(null)

  // 获取Slack状态（用于页面初始化、upsert后和轮询检查）
  const fetchSlackStatus = async (): Promise<SlackStatus | undefined> => {
    try {
      const response = (await getSlackStatusApi()) as unknown as SlackStatus

      setSlackStatus(response)

      // 如果有teamId但输入框为空，则填充输入框
      if (response.teamId && !teamId) {
        setTeamId(response.teamId)
      }

      return response
    } catch (error) {
      console.error('[BotManagement] failed to load Slack status', error)
      message.error(t('botManagement.messages.getStatusFailed'))
      return undefined
    }
  }

  // 开始轮询（仅在waiting或connect retry状态下使用）
  const startPolling = () => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current)
    }

    pollingRef.current = setInterval(async () => {
      const status = await fetchSlackStatus()
      // 轮询直至状态变为connected或disconnected
      if (status?.status === 'connected' || status?.status === 'disconnected') {
        if (pollingRef.current) {
          clearInterval(pollingRef.current)
          pollingRef.current = null
        }
        // 如果是disconnected，开始5秒倒计时自动重置
        if (status?.status === 'disconnected') {
          setConnecting(false)
          startCountdown()
        }
      }
    }, 60000) // 每分钟轮询一次
  }

  // 停止轮询
  const stopPolling = () => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current)
      pollingRef.current = null
    }
  }

  // 停止倒计时
  const stopCountdown = () => {
    if (countdownRef.current) {
      clearInterval(countdownRef.current)
      countdownRef.current = null
    }
  }

  // 开始倒计时（disconnected状态5秒后自动重置）
  const startCountdown = () => {
    stopCountdown()
    setCountdown(5)

    countdownRef.current = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          stopCountdown()
          handleReset()
          return 0
        }
        return prev - 1
      })
    }, 1000)
  }

  // 重置状态并返回创建页面
  const handleReset = () => {
    stopPolling()
    stopCountdown()
    setSlackStatus({
      teamId: '',
      createDate: 0,
      status: '',
    })
    setTeamId('')
    setConnecting(false)
    setCountdown(5)
  }

  // 组件加载时获取状态
  useEffect(() => {
    fetchSlackStatus().then((status) => {
      // 只有在waiting或connect retry状态下才开始轮询
      // connected状态：只显示连接成功
      // disconnected：显示断开提示并开始5秒倒计时
      // 未连接：显示创建表单
      if (
        status &&
        status.status &&
        status.status !== 'connected' &&
        status.status !== 'disconnected'
      ) {
        startPolling()
      } else if (status && status.status === 'disconnected') {
        startCountdown()
      }
    })

    // 组件销毁时清理轮询和倒计时
    return () => {
      stopPolling()
      stopCountdown()
    }
  }, [])

  // 连接Slack Bot
  const handleConnectSlackBot = async () => {
    if (!teamId.trim()) {
      message.error(t('botManagement.messages.teamIdRequired'))
      return
    }

    try {
      setConnecting(true)
      const response = (await upsertAppConfigApi({
        teamId: teamId.trim(),
      })) as unknown as UpsertResponse

      if (response.success) {
        message.success(response.message || t('botManagement.messages.addSuccess'))
        // upsert 成功后，调用 getSlackStatus 获取状态
        const status = await fetchSlackStatus()

        if (status) {
          // 如果是waiting或connect retry，开始轮询直至connected或disconnected
          if (status.status !== 'connected' && status.status !== 'disconnected') {
            startPolling()
          } else if (status.status === 'disconnected') {
            // 如果返回 disconnected 状态，开始倒计时
            startCountdown()
          }
        }
      } else {
        message.error(response.message || t('botManagement.messages.connectFailed'))
      }
    } catch (error) {
      console.error('[BotManagement] failed to connect Slack bot', error)
      message.error(t('botManagement.messages.connectFailedRetry'))
    } finally {
      setConnecting(false)
    }
  }

  const handleOpenSlackBotOAuth = (e: React.MouseEvent) => {
    e?.preventDefault?.()
    e?.stopPropagation?.()
    const oauthUrl =
      'https://slack.com/oauth/v2/authorize?client_id=9516861752262.9516864716406&scope=app_mentions:read,channels:history,channels:read,chat:write,groups:read,mpim:history,mpim:read,groups:history&user_scope='
    const width = 600
    const height = 700

    const left = window.screenX + Math.max(0, (window.outerWidth - width) / 2)

    const top = window.screenY + Math.max(0, (window.outerHeight - height) / 2)

    window.open(
      oauthUrl,
      'slack_oauth_popup',
      `width=${width},height=${height},left=${left},top=${top},menubar=no,toolbar=no,location=no,status=no,resizable=yes,scrollbars=yes`,
    )
  }

  // 根据状态获取状态文本和颜色
  const getStatusDisplay = () => {
    switch (slackStatus.status) {
      case 'connected':
        return { text: t('botManagement.slackBot.status.connected'), color: '#52c41a' }
      case 'waiting':
        return { text: t('botManagement.slackBot.status.waiting'), color: '#1890ff' }
      case 'connect retry':
        return { text: t('botManagement.slackBot.status.connectRetry'), color: '#faad14' }
      case 'disconnected':
        return { text: t('botManagement.slackBot.status.disconnected'), color: '#ff4d4f' }
      default:
        return { text: t('botManagement.slackBot.status.notConnected'), color: '#d9d9d9' }
    }
  }

  // 是否显示连接按钮（未连接或disconnected状态显示创建表单）
  const shouldShowConnectButton = () => {
    // 1. 未连接：没有teamId或状态为空
    // 2. disconnected：连接失败，允许重新创建
    return !slackStatus.teamId || slackStatus.status === '' || slackStatus.status === 'disconnected'
  }

  // 格式化创建日期
  const formatCreateDate = (timestamp: number) => {
    if (!timestamp) return ''
    return new Date(timestamp).toLocaleString('zh-CN')
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <Typography.Title level={4}>{t('botManagement.title')}</Typography.Title>
        <Typography.Text type="secondary">{t('botManagement.description')}</Typography.Text>
      </div>

      <Card className={styles.botCard} styles={{ body: { padding: 0 } }}>
        <div className={styles.botCardHeader}>
          <div className={styles.botIcon}>
            <img src={slackPng} alt="Slack" className={styles.slackLogo} />
          </div>
          <div className={styles.botInfo}>
            <Typography.Title level={5}>{t('botManagement.slackBot.name')}</Typography.Title>
            <Typography.Text type="secondary">
              {t('botManagement.slackBot.description')}
            </Typography.Text>
            <Typography.Text type="secondary">
              {t('botManagement.slackBot.descriptionSub')}
            </Typography.Text>
          </div>
        </div>

        <div className={styles.botCardBody}>
          {shouldShowConnectButton() ? (
            <div id="slackbot-setup-form" className={styles.setupForm}>
              <label className={styles.fieldLabel} htmlFor="slackbot-team-id">
                {t('botManagement.slackBot.teamIdLabel')}
              </label>
              <div className={styles.inputRow}>
                <Input
                  type="text"
                  id="slackbot-team-id"
                  placeholder={t('botManagement.slackBot.teamIdPlaceholder')}
                  value={teamId}
                  onChange={(e) => setTeamId(e.target.value)}
                  disabled={connecting}
                />
                <Button icon={<ExternalLink size={15} />} onClick={handleOpenSlackBotOAuth}>
                  {t('botManagement.slackBot.getTeamId')}
                </Button>
              </div>
              <Button
                type="primary"
                className={styles.primaryAction}
                icon={<Link size={15} />}
                onClick={handleConnectSlackBot}
                loading={connecting}
                disabled={!teamId.trim()}
              >
                {t('botManagement.slackBot.createBot')}
              </Button>
              {slackStatus.status === 'disconnected' && countdown > 0 && (
                <div className={styles.errorMessage}>
                  {t('botManagement.slackBot.disconnectedMessage', { countdown })}
                </div>
              )}
            </div>
          ) : (
            <div id="slackbot-bot-info" className={styles.statusInfo}>
              <div className={styles.statusRow}>
                <span>{t('botManagement.slackBot.teamId')}</span>
                <strong>{slackStatus.teamId}</strong>
              </div>
              {slackStatus.createDate > 0 && (
                <div className={styles.statusRow}>
                  <span>{t('botManagement.slackBot.createTime')}</span>
                  <strong>{formatCreateDate(slackStatus.createDate)}</strong>
                </div>
              )}
              <div className={styles.statusRow}>
                <span>{t('botManagement.slackBot.connectionStatus')}</span>
                <strong style={{ color: getStatusDisplay().color }}>
                  {getStatusDisplay().text}
                </strong>
              </div>
              {slackStatus.status === 'connected' && (
                <div className={styles.successMessage}>
                  {t('botManagement.slackBot.successMessage')}
                </div>
              )}
              {slackStatus.status === 'waiting' && (
                <div className={styles.waitingMessage}>
                  {t('botManagement.slackBot.waitingMessage')}
                </div>
              )}
              {slackStatus.status === 'connect retry' && (
                <div className={styles.waitingMessage}>
                  {t('botManagement.slackBot.connectRetryMessage')}
                </div>
              )}
              {slackStatus.status === 'disconnected' && (
                <div className={styles.errorMessage}>
                  {t('botManagement.slackBot.disconnectedMessage', { countdown })}
                </div>
              )}
              {(slackStatus.status === 'connected' ||
                slackStatus.status === 'connect retry' ||
                slackStatus.status === 'disconnected') && (
                <Button
                  className={styles.resetAction}
                  icon={<RotateCcw size={15} />}
                  onClick={handleReset}
                >
                  {t('botManagement.slackBot.resetButton')}
                </Button>
              )}
            </div>
          )}
        </div>

        <div className={styles.botCardFooter}>
          <a href="https://syn-cause.com/docs/" target="_blank" rel="noopener noreferrer">
            {t('botManagement.slackBot.docsLink')}
            <ExternalLink size={13} />
          </a>
        </div>
      </Card>
    </div>
  )
}

export default BotManagement
