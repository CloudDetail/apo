/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Popover, Select } from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { MdNotificationsPaused } from 'react-icons/md'
import { getAlertSilentConfigApi, saveAlertSilentConfigApi } from 'src/core/api/alerts'
import { convertUTCToLocal } from 'src/core/utils/time'
const items = [
  {
    label: '30s',
    value: '30s',
  },
  {
    label: '1m',
    value: '1m',
  },
  {
    label: '5m',
    value: '5m',
  },
  {
    label: '15m',
    value: '15m',
  },
  {
    label: '30m',
    value: '30m',
  },
  {
    label: '1h',
    value: '1h',
  },
  {
    label: '3h',
    value: '3h',
  },
  {
    label: '6h',
    value: '6h',
  },
  {
    label: '12h',
    value: '12h',
  },
]
const SilentAlert = ({ alertId }: { alertId?: string | null }) => {
  const { t } = useTranslation('oss/alertEvents')

  const [slience, setSlience] = useState(null)
  const [forDuration, setForDuration] = useState(null)

  const getAlertSilentConfig = () => {
    getAlertSilentConfigApi({ alertId }).then((res) => {
      setForDuration(res?.slience?.for)
      setSlience(res?.slience)
    })
  }
  const saveAlertSilentConfig = (forDuration) => {
    saveAlertSilentConfigApi({ alertId, forDuration }).then((res) => {
      getAlertSilentConfig()
    })
  }
  useEffect(() => {
    getAlertSilentConfig()
  }, [alertId])

  const ConfigCom = useCallback(() => {
    return (
      <>
        {t('silentFor')}：
        <Select
          value={forDuration}
          style={{ width: 120 }}
          options={items}
          onChange={saveAlertSilentConfig}
        ></Select>
        <div className="text-xs mt-1">
          {slience ? (
            <span className="text-gray-400">
              {t('silentTimerange')}：{convertUTCToLocal(slience.startAt)} to{' '}
              {convertUTCToLocal(slience.endAt)}
            </span>
          ) : (
            <span className="text-gray-400">{t('silentNotify')}</span>
          )}
        </div>
      </>
    )
  }, [forDuration])
  return (
    <>
      <Popover content={<ConfigCom />}>
        <Button
          color="gold"
          variant="filled"
          className="ml-2"
          classNames={{ icon: 'flex items-center' }}
          icon={<MdNotificationsPaused size={20} />}
        >
          {forDuration ? t('silent') : t('onSilent')}
        </Button>
      </Popover>
    </>
  )
}
export default SilentAlert
