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

const AlertsIntegrationTable = () => {
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
          title: '删除告警接入成功',
          color: 'success',
        })
      })
      .finally(() => {
        getAlertInputSourceList()
      })
  }
  const columns: TableProps<AlertInputBaseInfo>['columns'] = [
    {
      title: '告警源类型',
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
      title: '告警接入名称',
      dataIndex: 'sourceName',
      key: 'sourceName',
    },
    {
      title: '操作',
      dataIndex: 'operation',
      render: (_, record) => {
        return (
          <>
            <Button
              type="text"
              onClick={() => openDrawer(record.sourceId, record.sourceType)}
              icon={<MdOutlineEdit className="text-blue-400 hover:text-blue-400" />}
            >
              <span className="text-blue-400 hover:text-blue-400">编辑</span>
            </Button>
            <Popconfirm
              title={
                <>
                  是否确定删除名为“<span className="font-bold ">{record.sourceName}</span>
                  ”的告警接入
                </>
              }
              onConfirm={() => deleteAlertIntegration(record.sourceId)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                删除
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
        <Typography.Title level={5}>告警接入列表</Typography.Title>
      </Typography>
      {/* <Search placeholder="输入搜索告警接入名称" className="mb-3" /> */}
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
