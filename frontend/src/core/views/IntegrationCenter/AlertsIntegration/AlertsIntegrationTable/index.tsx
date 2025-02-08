/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Popconfirm, Table, TableProps, Typography } from 'antd'
import { datasourceSrc } from '../../constant'
import { MdOutlineEdit } from 'react-icons/md'
import Search from 'antd/es/input/Search'
import { deleteAlertIntegrationApi, getAlertInputSourceListApi } from 'src/core/api/alertInput'
import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
import { showToast } from 'src/core/utils/toast'
import { AlertInputBaseInfo, AlertKey, SourceInfo } from 'src/core/types/alertIntegration'
import { Trans, useTranslation } from 'react-i18next'

const AlertsIntegrationTable = () => {
  const { t } = useTranslation('core/alertsIntegration')
  const [data, setData] = useState([])

  const [searchParams, setSearchParams] = useSearchParams()
  const configDrawerVisible = useAlertIntegrationContext((ctx) => ctx.configDrawerVisible)
  const openDrawer = (sourceId: string, sourceType: AlertKey) => {
    setSearchParams({ sourceId: sourceId, sourceType: sourceType })
  }
  const getAlertInputSourceList = () => {
    getAlertInputSourceListApi()
      .then((res) => {
        setData(res?.alertSources || [])
      })
      .catch((error) => {
        console.error(error)
        setData([])
      })
  }

  const deleteAlertIntegration = (sourceId: string) => {
    deleteAlertIntegrationApi(sourceId)
      .then((res) => {
        showToast({
          title: t('deleteSuccess'),
          color: 'success',
        })
      })
      .finally(() => {
        getAlertInputSourceList()
      })
  }
  const columns: TableProps<AlertInputBaseInfo>['columns'] = [
    {
      title: t('sourceType'),
      dataIndex: 'sourceType',
      key: 'sourceType',
      render: (text: AlertKey) => (
        <div className="flex">
          <img src={datasourceSrc[text]} height={30} width={20} className="mr-2"></img>
          {text}
        </div>
      ),
    },
    {
      title: t('sourceName'),
      dataIndex: 'sourceName',
      key: 'sourceName',
    },
    {
      title: t('operation'),
      dataIndex: 'operation',
      render: (_, record) => {
        return (
          <>
            <Button
              type="text"
              onClick={() => openDrawer(record.sourceId, record.sourceType)}
              icon={<MdOutlineEdit className="text-blue-400 hover:text-blue-400" />}
            >
              <span className="text-blue-400 hover:text-blue-400">{t('edit')}</span>
            </Button>
            <Popconfirm
              title={
                <>
                  <Trans
                    t={t}
                    i18nKey="confirmDelete"
                    values={{ sourceName: record.sourceName }}
                    components={{ 1: <span className="font-bold" /> }}
                  />
                </>
              }
              onConfirm={() => deleteAlertIntegration(record.sourceId)}
              okText={t('confirm')}
              cancelText={t('cancel')}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                {t('delete')}
              </Button>
            </Popconfirm>
          </>
        )
      },
    },
  ]
  useEffect(() => {
    if (!configDrawerVisible) {
      getAlertInputSourceList()
    }
  }, [configDrawerVisible])
  return (
    <div className="overflow-hidden h-full">
      <Typography>
        <Typography.Title level={5}>{t('list')}</Typography.Title>
      </Typography>
      {/* <Search  className="mb-3" /> */}
      <Table
        columns={columns}
        dataSource={data}
        pagination={false}
        scroll={{ y: 'calc(100vh - 220px)' }}
      />
    </div>
  )
}
export default AlertsIntegrationTable
