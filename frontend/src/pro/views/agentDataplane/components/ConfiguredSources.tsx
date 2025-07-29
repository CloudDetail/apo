/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Popconfirm, Table, TableProps, Typography } from 'antd'
import { MdOutlineEdit } from 'react-icons/md'
import { useEffect, useState } from 'react'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { notify } from 'src/core/utils/notify'
import { AlertInputBaseInfo } from 'src/core/types/alertIntegration'
import { Trans, useTranslation } from 'react-i18next'
import { dataplaneIconMap } from 'src/pro/constants'
import { deleteDataplaneApi, getDataplaneListApi } from 'src/pro/api/dataplane'
import Capability from 'src/pro/components/Dataplane/Capability'

const ConfiguredSources = ({ openDrawer, key }: { openDrawer: (any) => void; key: number }) => {
  const { t } = useTranslation('pro/agentDatasource')
  const [data, setData] = useState([])

  const getDataplaneList = () => {
    getDataplaneListApi()
      .then((res) => {
        setData(res?.dataPlanes || [])
      })
      .catch((error) => {
        console.error(error)
        setData([])
      })
  }

  const deleteDataplane = (sourceId: string) => {
    deleteDataplaneApi(sourceId)
      .then((res) => {
        notify({
          message: t('deleteSuccess'),
          type: 'success',
        })
      })
      .finally(() => {
        getDataplaneList()
      })
  }
  const columns: TableProps<AlertInputBaseInfo>['columns'] = [
    {
      title: t('typ'),
      dataIndex: 'typ',
      key: 'typ',
      width: 150,
      render: (text: string) => (
        <div className="flex">
          <img src={dataplaneIconMap[text]} height={30} width={20} className="mr-2"></img>
          {text}
        </div>
      ),
    },
    {
      title: t('name'),
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: t('desc'),
      dataIndex: 'desc',
      key: 'desc',
      width: 150,
    },
    {
      title: t('capabilitySpec'),
      dataIndex: 'capabilitySpec',
      key: 'capabilitySpec',
      render: (text: string[]) => <Capability capabilitySpec={text} />,
    },
    {
      title: t('operation'),
      dataIndex: 'operation',
      render: (_, record) => {
        return (
          <>
            <Button
              type="text"
              onClick={() => openDrawer(record)}
              icon={
                <MdOutlineEdit className="!text-[var(--ant-color-primary-text)] !hover:text-[var(--ant-color-primary-text-active)]" />
              }
            >
              <span className="text-[var(--ant-color-primary-text)] hover:text-[var(--ant-color-primary-text-active)]">
                {t('edit')}
              </span>
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
              onConfirm={() => deleteDataplane(record.id)}
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
    getDataplaneList()
  }, [key])
  return (
    <div className="overflow-hidden h-full">
      <Typography>
        <Typography.Title level={5}>{t('list')}</Typography.Title>
      </Typography>
      {/* <Search  className="mb-3" /> */}
      <Table columns={columns} dataSource={data} scroll={{ y: 'calc(100vh - 260px)' }} />
    </div>
  )
}
export default ConfiguredSources
