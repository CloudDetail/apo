/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'
import TimeSinceRefresh from './TimeSinceRefresh'
import DateTimeRangePicker from './DateTimeRangePicker'
import { Button, Input, Popover, Segmented, theme } from 'antd'
import { FiSend } from 'react-icons/fi'
import CopyButton from '../CopyButton'
import { useLocation, useSearchParams } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { convertTime, TimestampToISO } from 'src/core/utils/time'
import { timeRangeMap } from 'src/core/store/reducers/timeRangeReducer'
import RefreshDateTime from './RefreshDateTime'
import { useTranslation } from 'react-i18next'

function ShareLink() {
  const { t, i18n } = useTranslation('core/dateTime')
  const location = useLocation()
  const [shareType, setShareType] = useState(t('dataTimeCombine.absoluteTimeText'))
  const { rangeTypeKey, startTime, endTime } = useSelector((state) => state.timeRange)
  const [copyUrl, setCopyUrl] = useState(window.location.href)
  const [searchParams] = useSearchParams()

  useEffect(() => {
    // 在翻译加载完成后设置shareType
    if (i18n.isInitialized) {
      setShareType(t('dataTimeCombine.absoluteTimeText'))
    }
  }, [i18n.isInitialized, t])

  useEffect(() => {
    let url = window.location.href
    if (shareType === t('dataTimeCombine.absoluteTimeText') && rangeTypeKey) {
      const currentUrl = new URL(window.location.href)
      const params = new URLSearchParams(location.search)
      params.delete('relativeTime')
      params.set('from', TimestampToISO(startTime))
      params.set('to', TimestampToISO(endTime))
      url = currentUrl.origin + '/#' + location.pathname + '?' + params.toString()
    }
    setCopyUrl(url)
  }, [shareType, rangeTypeKey, location.search, startTime, endTime, t])
  const { useToken } = theme
  const { token } = useToken()
  return (
    <Popover
      content={
        <div className="w-[500px]">
          <Segmented
            value={shareType}
            options={[
              t('dataTimeCombine.absoluteTimeText'),
              {
                label: t('dataTimeCombine.relativeTimeText'),
                value: t('dataTimeCombine.relativeTimeText'),
                disabled: !rangeTypeKey,
              },
            ]}
            onChange={setShareType}
          />
          <div
            className="my-2 text-[var(--ant-color-text-secondary)] mx-1"
            style={{ color: token.colorTextSecondary }}
          >
            {shareType === t('dataTimeCombine.absoluteTimeText') ? (
              <>
                {convertTime(startTime, 'yyyy-mm-dd hh:mm:ss')} to{' '}
                {convertTime(endTime, 'yyyy-mm-dd hh:mm:ss')}
              </>
            ) : (
              timeRangeMap[rangeTypeKey]?.name
            )}
          </div>
          <Input value={copyUrl} addonAfter={<CopyButton value={copyUrl} />} />
        </div>
      }
    >
      <Button type="text" icon={<FiSend />}></Button>
    </Popover>
  )
}

export default function DateTimeCombine() {
  const { rangeTypeKey } = useSelector((state) => state.timeRange)

  return (
    <div className="flex items-center">
      {rangeTypeKey && <TimeSinceRefresh />}
      <DateTimeRangePicker />
      {rangeTypeKey && <RefreshDateTime />}
      <ShareLink />
    </div>
  )
}
