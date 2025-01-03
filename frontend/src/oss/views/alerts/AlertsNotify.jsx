/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Input, Popconfirm, Select, Space } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { deleteAlertNotifyApi, getAlertmanagerListApi } from 'core/api/alerts'
import LoadingSpinner from 'src/core/components/Spinner'
import BasicTable from 'src/core/components/Table/basicTable'
import { showToast } from 'src/core/utils/toast'
import { MdAdd, MdOutlineEdit } from 'react-icons/md'
import { useSelector } from 'react-redux'
import ModifyAlertNotifyModal from './modal/ModifyAlertNotifyModal'

export default function AlertsNotify() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalInfo, setModalInfo] = useState(null)
  const [searchName, setSearchName] = useState(null)
  const deleteAlertNotify = (row) => {
    deleteAlertNotifyApi(
      row.dingTalkConfigs
        ? {
            name: row.name,
            type: 'dingtalk',
          }
        : {
            name: row.name,
          },
    ).then((res) => {
      showToast({
        title: '删除告警通知成功',
        color: 'success',
      })
      refreshTable()
    })
  }

  const judgmentType = (type) => {
    switch (type) {
      case 'emailConfigs':
        return '邮件'
      case 'webhookConfigs':
        return 'webhook'
      case 'dingTalkConfigs':
        return '钉钉'
      case 'wechatConfigs':
        return '微信'
    }
  }

  const getUrl = (type, row) => {
    switch (type) {
      case 'emailConfigs':
        return row.emailConfigs[0]?.to
      case 'webhookConfigs':
        return row.webhookConfigs[0]?.url
      case 'dingTalkConfigs':
        return row.dingTalkConfigs[0]?.url
      case 'wechatConfigs':
        return ''
      default:
        return 'N/A'
    }
  }

  const typeList = ['emailConfigs', 'webhookConfigs', 'dingTalkConfigs', 'wechatConfigs']

  const column = [
    {
      title: '告警通知规则名',
      accessor: 'name',
      justifyContent: 'left',
      customWidth: '20%',
    },
    {
      title: '通知类型',
      accessor: 'type',
      customWidth: 120,
      Cell: (props) => {
        const row = props.row.original
        let foundItem = typeList.find((item) => Object.hasOwn(row, item))
        foundItem = judgmentType(foundItem)
        return foundItem || null
      },
    },
    {
      title: '通知邮箱或WebhookUrl',
      accessor: 'to',
      customWidth: '50%',
      Cell: (props) => {
        const row = props.row.original
        let foundItem = typeList.find((item) => Object.hasOwn(row, item))
        foundItem = getUrl(foundItem, row)
        return foundItem
      },
    },
    {
      title: '操作',
      accessor: 'action',
      Cell: (props) => {
        const row = props.row.original
        return (
          <div className="flex">
            <Button
              type="text"
              onClick={() => clickEditRule(row)}
              icon={<MdOutlineEdit className="text-blue-400 hover:text-blue-400" />}
            >
              <span className="text-blue-400 hover:text-blue-400">编辑</span>
            </Button>
            <Popconfirm
              title={
                <>
                  是否确定删除名为“<span className="font-bold ">{row.alert}</span>
                  ”的告警规则
                </>
              }
              onConfirm={() => deleteAlertNotify(row)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                删除
              </Button>
            </Popconfirm>
          </div>
          // <div className=" cursor-pointer">
          //   <AiOutlineDelete color="#97242e" size={18} />
          //   删除
          // </div>
        )
      },
    },
  ]

  const clickAddRule = () => {
    setModalInfo(null)
    setModalVisible(true)
  }
  const clickEditRule = (notifyInfo) => {
    setModalInfo(notifyInfo)
    setModalVisible(true)
  }
  useEffect(() => {
    fetchData()
  }, [pageSize, pageIndex, searchName])
  const fetchData = () => {
    setLoading(true)

    getAlertmanagerListApi({
      currentPage: pageIndex,
      pageSize: pageSize,
      name: searchName,
      refreshCache: true,
    })
      .then((res) => {
        setLoading(false)
        setTotal(res.pagination.total)
        setData(res.amConfigReceivers)
      })
      .catch((error) => {
        setLoading(false)
      })
  }
  const handleTableChange = (props) => {
    if (props.pageSize && props.pageIndex) {
      setPageSize(props.pageSize), setPageIndex(props.pageIndex)
    }
  }
  const refreshTable = () => {
    fetchData()
    setPageIndex(1)
  }
  const tableProps = useMemo(() => {
    // 分页处理
    return {
      columns: column,
      data: data,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        pageCount: Math.ceil(total / pageSize),
      },
      loading: false,
    }
  }, [data, pageIndex, pageSize])
  return (
    <Card
      style={{ height: 'calc(100vh - 60px)' }}
      styles={{
        body: {
          height: '100%',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          padding: '12px 24px',
        },
      }}
    >
      <LoadingSpinner loading={loading} />
      <div className="flex items-center justify-betweeen text-sm ">
        <Space className="flex-grow">
          <Space className="flex-1">
            <span className="text-nowrap">通知规则名：</span>
            <Input
              value={searchName}
              onChange={(e) => {
                setSearchName(e.target.value)
                setPageIndex(1)
              }}
            />
          </Space>
        </Space>

        <Button
          type="primary"
          icon={<MdAdd size={20} />}
          onClick={clickAddRule}
          className="flex-grow-0 flex-shrink-0"
        >
          <span className="text-xs">新增告警通知</span>
        </Button>
      </div>
      <div className="text-sm flex-1 overflow-auto">
        <div className="h-full text-xs justify-between">
          <BasicTable {...tableProps} />
        </div>
      </div>
      <ModifyAlertNotifyModal
        modalVisible={modalVisible}
        notifyInfo={modalInfo}
        closeModal={() => setModalVisible(false)}
        refresh={refreshTable}
      />
    </Card>
  )
}
