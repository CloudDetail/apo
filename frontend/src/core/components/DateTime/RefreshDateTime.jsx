/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Dropdown, Tooltip } from 'antd'
import React, { useEffect, useState } from 'react'
import { LuRefreshCw } from 'react-icons/lu'
import { MdArrowDropDown } from 'react-icons/md'
import { useDispatch, useSelector } from 'react-redux'
import { useTranslation } from 'react-i18next'

export default function RefreshDateTime({ size = 'normal' }) {
  const { t } = useTranslation('core/dateTime')
  const [refreshKey, setRefreshKey] = useState(null)
  const { refreshTimestamp } = useSelector((state) => state.timeRange)
  const dispatch = useDispatch()
  const refreshTime = (value) => {
    dispatch({ type: 'REFRESH_TIMERANGE', payload: value })
  }
  const handleButtonClick = () => {
    refreshTime()
  }
  const handleMenuClick = (e) => {
    setRefreshKey(e.key)
    localStorage.setItem('refreshKey', e.key)
  }
  const items = [
    {
      key: 'grp',
      label: t('group'),
      type: 'group',
      children: [
        {
          label: t('refreshDateTime.refreshDateTimeOffText'),
          value: 0,
          key: '关',
        },
        {
          label: '15s',
          value: 15000,
          key: '15s',
        },
        {
          label: '1m',
          value: 60000,
          key: '1m',
        },
        {
          label: '5m',
          value: 300000,
          key: '5m',
        },
        {
          label: '15m',
          value: 900000,
          key: '15m',
        },
        {
          label: '30m',
          value: 1800000,
          key: '30m',
        },
        {
          label: '1h',
          value: 3600000,
          key: '1h',
        },
        {
          label: '1d',
          value: 86400000,
          key: '1d',
        },
      ],
    },
  ]
  const menuProps = {
    items,
    onClick: handleMenuClick,
    selectedKeys: [refreshKey],
  }

  useEffect(() => {
    let time = items[0].children.find((item) => item.key === refreshKey)?.value
    let intervalId
    if (time) {
      intervalId = setInterval(() => {
        handleRefresh()
      }, time)
    }
    return () => clearInterval(intervalId)
  }, [refreshKey, refreshTimestamp])

  // 刷新函数
  const handleRefresh = () => {
    refreshTime()
  }
  useEffect(() => {
    if (localStorage.getItem('refreshKey')) {
      setRefreshKey(localStorage.getItem('refreshKey'))
    }
  }, [])
  return (
    <div className={`mx-1 refresh-date-time ${size === 'small' ? 'text-xs' : ''}`}>
      <Dropdown.Button
        menu={menuProps}
        onClick={handleButtonClick}
        icon={
          <div className="flex w-full items-center">
            {refreshKey && refreshKey !== '关' && refreshKey}
            <MdArrowDropDown />
          </div>
        }
        size={size}
      >
        <Tooltip title={t('refresh')}>
          <LuRefreshCw />
        </Tooltip>
      </Dropdown.Button>
    </div>
  )
}
