/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback, useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { useTranslation } from 'react-i18next'
import { timeUtils } from 'src/core/utils/time'

export default function TimeSinceRefresh() {
  const { t } = useTranslation('core/dateTime')
  const { refreshTimestamp } = useSelector((state) => state.timeRange)
  const [timeDiff, setTimeDiff] = useState('')
  const [intervalTime, setIntervalTime] = useState(1000)

  const calculateTimeDifference = useCallback(() => {
    // Use timeUtils to get the current microsecond timestamp
    const now = timeUtils.nowMicro()
    // Use timeUtils to calculate the second-level difference
    const diffInSeconds = Math.floor(timeUtils.convertMicroTime(now - refreshTimestamp, 's', 0))
    const days = Math.floor(diffInSeconds / (3600 * 24))
    const hours = Math.floor((diffInSeconds % (3600 * 24)) / 3600)
    const minutes = Math.floor((diffInSeconds % 3600) / 60)
    const seconds = diffInSeconds % 60

    let timeString = ' '
    if (days > 0) {
      timeString += `${days} ${t('timeSinceRefresh.dayText')} `
    } else if (hours > 0) {
      timeString += `${hours} ${t('timeSinceRefresh.hourText')} `
    } else if (minutes > 0) {
      timeString += `${minutes} ${t('timeSinceRefresh.minuteText')} `
    } else {
      timeString += `${seconds} ${t('timeSinceRefresh.secondText')} `
    }

    // 动态调整时间更新的频率
    if (days > 0) {
      setIntervalTime(86400000) // 当超过1天，更新频率调整为1天
    } else if (hours > 0) {
      setIntervalTime(3600000) // 当超过1小时，更新频率调整为1小时
    } else if (minutes > 0) {
      setIntervalTime(60000) // 当超过1分钟，更新频率为1分钟
    } else {
      setIntervalTime(1000) // 默认每秒更新一次
    }

    setTimeDiff(timeString)
  }, [refreshTimestamp])

  useEffect(() => {
    calculateTimeDifference() // 初始计算
    const intervalId = setInterval(calculateTimeDifference, intervalTime)

    return () => clearInterval(intervalId)
  }, [calculateTimeDifference, intervalTime])

  return (
    <div className="text-xs">
      {t('timeSinceRefresh.refreshTipText')}
      {timeDiff}
      {t('timeSinceRefresh.refreshTipTextAgo')}
    </div>
  )
}
