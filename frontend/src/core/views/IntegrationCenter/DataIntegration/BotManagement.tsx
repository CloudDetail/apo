/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Input, Button, message } from 'antd'
import React, { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import { getSlackStatusApi, upsertAppConfigApi } from 'src/core/api/appConfig'
import './slackapp.css'
import slackPng from 'src/core/assets/images/slack.png'
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
    } catch {
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
      console.log('status', status)
      if (
        status &&
        status.status &&
        status.status !== 'connected' &&
        status.status !== 'disconnected'
      ) {
        startPolling()
      } else if (status && status.status === 'disconnected') {
        console.log('1212')
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
    } catch {
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
    <div className="content-section">
      <div className="section-header">
        <div className="section-title-wrapper">
          <h2 className="section-title">{t('botManagement.title')}</h2>
          <p className="section-description">{t('botManagement.description')}</p>
        </div>
      </div>

      <BasicCard>
        <div className="bot-management-content">
          {/* Slack App 设置卡片 */}
          <div className="bot-card border border-[var(--ant-color-border-secondary)]">
            <div className="bot-card-header bg-[var(--ant-color-fill-tertiary)]">
              <div className="bot-icon">
                <img src={slackPng} alt="Slack" className="slack-logo" />
              </div>
              <div className="bot-info">
                <h3 className="bot-name">{t('botManagement.slackBot.name')}</h3>
                <p className="bot-status">{t('botManagement.slackBot.description')}</p>
                <p className="bot-status">{t('botManagement.slackBot.descriptionSub')}</p>
              </div>
            </div>
            <div className="bot-card-body">
              {shouldShowConnectButton() ? (
                <div id="slackbot-setup-form">
                  <div className="slackbot-form-group">
                    <label className="font-bold" htmlFor="slackbot-team-id">
                      {t('botManagement.slackBot.teamIdLabel')}
                    </label>
                    <div className="slackbot-input-with-button flex gap-2 my-2">
                      <Input
                        type="text"
                        id="slackbot-team-id"
                        placeholder={t('botManagement.slackBot.teamIdPlaceholder')}
                        value={teamId}
                        onChange={(e) => setTeamId(e.target.value)}
                        disabled={connecting}
                      />
                      <a onClick={handleOpenSlackBotOAuth} className="slackbot get-teamid-btn">
                        {t('botManagement.slackBot.getTeamId')}
                      </a>
                    </div>
                  </div>
                  <Button
                    type="primary"
                    size="large"
                    className="slackbot create-robot-btn h-[50px]"
                    onClick={handleConnectSlackBot}
                    loading={connecting}
                    disabled={!teamId.trim()}
                  >
                    {t('botManagement.slackBot.createBot')}
                  </Button>
                  {slackStatus.status === 'disconnected' && countdown > 0 && (
                    <div className="waiting-message" style={{ color: '#ff4d4f' }}>
                      {t('botManagement.slackBot.disconnectedMessage', { countdown })}
                    </div>
                  )}
                </div>
              ) : (
                <div id="slackbot-bot-info">
                  <div className="bot-status-info">
                    <div className="status-row">
                      <label className="font-bold">{t('botManagement.slackBot.teamId')}</label>
                      <span>{slackStatus.teamId}</span>
                    </div>
                    {slackStatus.createDate > 0 && (
                      <div className="status-row">
                        <label className="font-bold">
                          {t('botManagement.slackBot.createTime')}
                        </label>
                        <span>{formatCreateDate(slackStatus.createDate)}</span>
                      </div>
                    )}
                    <div className="status-row">
                      <label className="font-bold">
                        {t('botManagement.slackBot.connectionStatus')}
                      </label>
                      <span style={{ color: getStatusDisplay().color, fontWeight: 'bold' }}>
                        {getStatusDisplay().text}
                      </span>
                    </div>
                    {slackStatus.status === 'connected' && (
                      <div className="successMessage">
                        {t('botManagement.slackBot.successMessage')}
                      </div>
                    )}
                    {slackStatus.status === 'waiting' && (
                      <div className="waiting-message">
                        {t('botManagement.slackBot.waitingMessage')}
                      </div>
                    )}
                    {slackStatus.status === 'connect retry' && (
                      <div className="waiting-message">
                        {t('botManagement.slackBot.connectRetryMessage')}
                      </div>
                    )}
                    {slackStatus.status === 'disconnected' && (
                      <div className="waiting-message" style={{ color: '#ff4d4f' }}>
                        {t('botManagement.slackBot.disconnectedMessage', { countdown })}
                      </div>
                    )}
                    {/* 添加重置按钮 */}
                    {(slackStatus.status === 'connected' ||
                      slackStatus.status === 'connect retry' ||
                      slackStatus.status === 'disconnected') && (
                      <div className="mt-4">
                        <Button
                          onClick={handleReset}
                          type="default"
                          size="large"
                          className="w-full h-[50px]"
                        >
                          {t('botManagement.slackBot.resetButton')}
                        </Button>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
            <div className="bot-card-footer bg-[var(--ant-color-fill-tertiary)]">
              <a
                href="https://syn-cause.com/docs/"
                target="_blank"
                rel="noopener noreferrer"
                className="docs-link"
              >
                {t('botManagement.slackBot.docsLink')}
              </a>
            </div>
          </div>
        </div>
      </BasicCard>
    </div>
  )
}

export default BotManagement
