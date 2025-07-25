/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Divider, Radio } from 'antd'
import DatasourceItem from 'src/core/views/IntegrationCenter/components/DatasourceItem'
import apo from 'src/core/assets/images/logo.svg'
import { useTranslation } from 'react-i18next'
import { useMessageContext } from 'src/core/contexts/MessageContext'

interface DbTypeRadioProps {
  value: string
  onChange: any
  id: string
}
const DbTypeRadio = ({ id, value, onChange }: DbTypeRadioProps) => {
  const { t } = useTranslation('core/dataIntegration')
  const messageApi = useMessageContext()
  function clickRadio(key: string | undefined) {
    if (key !== 'self-collector') {
      messageApi.warning(t('typeNotSupport'))
    } else {
      onChange(key)
    }
  }
  return (
    <div id={id} className="flex overflow-hidden">
      <div className="flex-shrink-0 flex-grow-0">
        <div className="text-[var(--ant-color-text-secondary)]">{t('datasourceApo')}</div>
        <div className="relative w-[100px]" onClick={() => clickRadio('self-collector')}>
          <DatasourceItem
            size="small"
            src={apo}
            name={t('datasourceApo')}
            description=""
            key={'apo'}
          />
          <Radio checked={value === 'self-collector'} className="absolute right-0 top-0"></Radio>
        </div>
      </div>
      <Divider type="vertical" className="h-[60px] mt-4" />
      {/* <div className="flex-1">
        <div className="text-[var(--ant-color-text-secondary)]">{t('datasourceExisted')}</div>
        <div className="flex flex-wrap ">
          {logsItems.map((item) => (
            <div className="relative mx-1 mb-2" onClick={() => clickRadio(item.key)}>
              <DatasourceItem
                size="small"
                src={item.src}
                name={item.name}
                description=""
                key={item.key}
              />
              <Radio
                disabled={item.key !== 'apo'}
                checked={value === item.key}
                className="absolute right-0 top-0"
              ></Radio>
            </div>
          ))}
        </div>
      </div> */}
    </div>
  )
}
export default DbTypeRadio
