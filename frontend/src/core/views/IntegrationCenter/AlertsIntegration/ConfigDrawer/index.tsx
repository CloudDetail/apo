/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Drawer, Tabs } from 'antd'
import React from 'react'
import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { alertItems, datasourceSrc } from '../../constant'
import BaseInfoContent from './BaseInfoContent'
import TagContent from './TagContent'
import { getAlertInputBaseInfoApi } from 'src/core/api/alertInput'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
import { AlertKey } from 'src/core/types/alertIntegration'
import SourceTypeInfo from './SourceTypeInfo'
import { useTranslation } from 'react-i18next'

interface DrawerTitleProps {
  sourceName: string | null
  sourceType: AlertKey
}
const DrawerTitle = React.memo((props: DrawerTitleProps) => {
  const { sourceName, sourceType } = props
  return (
    <div className="flex items-center">
      <img
        src={datasourceSrc[sourceType]}
        height={30}
        className="overflow-hidden h-[30px] mr-2 flex items-center justify-center"
      ></img>
      {sourceName ? sourceName : alertItems.find((item) => item.key === sourceType)?.name}
    </div>
  )
})
const IntegrationDrawer = () => {
  const { t } = useTranslation('core/alertsIntegration')
  const [searchParams, setSearchParams] = useSearchParams()
  const [sourceId, setSourceId] = useState<string | null>(null)
  const [sourceType, setSourceType] = useState<AlertKey>('json')
  const [sourceName, setSourceName] = useState<string | null>(null)
  const [clusters, setClusters] = useState([])
  const setConfigDrawerVisible = useAlertIntegrationContext((ctx) => ctx.setConfigDrawerVisible)
  const configDrawerVisible = useAlertIntegrationContext((ctx) => ctx.configDrawerVisible)

  const closeDrawer = () => {
    setConfigDrawerVisible(false)
    const newParams = new URLSearchParams(searchParams)
    newParams.delete('sourceId')
    newParams.delete('sourceType')
    setSearchParams(newParams, { replace: true })
  }
  useEffect(() => {
    const sourceId = searchParams.get('sourceId')
    const sourceType = searchParams.get('sourceType')
    if (!sourceId) {
      setConfigDrawerVisible(false)
      setSourceId(null)
      setSourceName(null)
    } else if (sourceId === 'add') {
      setSourceType(sourceType as AlertKey)
      setConfigDrawerVisible(true)
      setSourceId(null)
      setSourceName(null)
    } else {
      getAlertIntegrationBaseInfo(sourceId)
      setConfigDrawerVisible(true)
    }
  }, [searchParams])
  const getAlertIntegrationBaseInfo = (sourceId: string) => {
    setSourceId(sourceId)
    getAlertInputBaseInfoApi({ sourceId: sourceId }).then((res) => {
      setSourceName(res?.sourceName)
      setSourceType(res?.sourceType)
      setClusters(res?.clusters || [])
    })
  }
  return (
    <Drawer
      title={sourceType && <DrawerTitle sourceName={sourceName} sourceType={sourceType} />}
      onClose={closeDrawer}
      open={configDrawerVisible}
      width={'80%'}
      classNames={{
        // content: 'bg-[#101215]',
        body: 'pt-0',
      }}
      styles={{
        content: { background: 'var(--ant-color-bg-container)', border: '1px solid #343a46' },
      }}
    >
      {sourceId ? (
        <Tabs
          items={[
            {
              key: '1',
              label: t('setting'),
              children: (
                <>
                  <BaseInfoContent
                    sourceId={sourceId}
                    sourceType={sourceType}
                    sourceName={sourceName}
                    clusters={clusters}
                    refreshDrawer={() => getAlertIntegrationBaseInfo(sourceId)}
                  />
                  <TagContent sourceId={sourceId} />
                </>
              ),
            },
            {
              key: '2',
              label: t('documentation'),
              children: <SourceTypeInfo sourceType={sourceType} />,
            },
          ]}
        ></Tabs>
      ) : (
        <div className="pt-4">
          <BaseInfoContent
            sourceId={sourceId}
            sourceType={sourceType}
            sourceName={sourceName}
            clusters={clusters}
            refreshDrawer={() => getAlertIntegrationBaseInfo(sourceId)}
            closeDrawer={closeDrawer}
          />
          <SourceTypeInfo sourceType={sourceType} />
        </div>
      )}
    </Drawer>
  )
}

export default IntegrationDrawer
