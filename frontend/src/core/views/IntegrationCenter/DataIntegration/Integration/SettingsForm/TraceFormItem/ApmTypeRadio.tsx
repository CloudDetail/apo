/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Badge, Divider, Radio } from 'antd'
import DatasourceItem from 'src/core/views/IntegrationCenter/components/DatasourceItem'
import { traceItems } from 'src/core/views/IntegrationCenter/constant'
import apo from 'src/core/assets/images/logo.svg'
import { useTranslation } from 'react-i18next'
import { useMessageContext } from 'src/core/contexts/MessageContext'

interface ApmTypeRadioProps {
  value: string
  onChange: any
  id: string
}
const allowApmType = ['jaeger', 'opentelemetry', 'skywalking']
// const eeApmType = ['arms']
const ApmTypeRadio = ({ id, value, onChange }: ApmTypeRadioProps) => {
  const { t } = useTranslation('core/dataIntegration')
  const messageApi = useMessageContext()
  function clickRadio(key: string | undefined) {
    if (!allowApmType.includes(key)) {
      messageApi.warning(t('eeToast'))
    } else {
      onChange(key)
    }
  }
  return (
    <div id={id} className="flex overflow-hidden">
      <div className="flex-shrink-0 flex-grow-0">
        <div className="text-[var(--ant-color-text-secondary)]">{t('datasourceApo')}</div>
        <div className="relative w-[100px]" onClick={() => clickRadio('opentelemetry')}>
          <DatasourceItem
            size="small"
            src={apo}
            name={t('datasourceApo')}
            description=""
            key={'apo'}
          />
          <Radio checked={value === 'opentelemetry'} className="absolute right-0 top-0"></Radio>
        </div>
      </div>
      <Divider type="vertical" className="h-[60px] mt-4" />
      <div className="flex-1">
        <div className="text-[var(--ant-color-text-secondary)]">{t('datasourceExisted')}</div>
        <div className="flex flex-wrap ">
          {traceItems.map((item) => (
            <div className="relative mx-1 mb-2" onClick={() => clickRadio(item.apmType)}>
              {!allowApmType.includes(item.apmType) ? (
                <Badge.Ribbon text={t('ee')} style={{ top: 0, height: '20px', fontSize: 10 }}>
                  <DatasourceItem
                    size="small"
                    src={item.src}
                    name={item.name}
                    description=""
                    key={item.key}
                  />
                </Badge.Ribbon>
              ) : (
                <>
                  <DatasourceItem
                    size="small"
                    src={item.src}
                    name={item.name}
                    description=""
                    key={item.key}
                  />{' '}
                  <Radio
                    checked={value === item.apmType}
                    disabled={!allowApmType.includes(item.apmType)}
                    className="absolute right-0 top-0"
                  ></Radio>
                </>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
export default ApmTypeRadio
