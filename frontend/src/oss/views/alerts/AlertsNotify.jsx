/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Input, Popconfirm, Space } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { deleteAlertNotifyApi, getAlertmanagerListApi } from 'core/api/alerts'
import LoadingSpinner from 'src/core/components/Spinner'
import BasicTable from 'src/core/components/Table/basicTable'
import { showToast } from 'src/core/utils/toast'
import { MdAdd, MdOutlineEdit } from 'react-icons/md'
import ModifyAlertNotifyModal from './modal/ModifyAlertNotifyModal'
import { useTranslation } from 'react-i18next' // 引入i18n
import CustomCard from 'src/core/components/Card/CustomCard'

export default function AlertsNotify() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalInfo, setModalInfo] = useState(null)
  const [searchName, setSearchName] = useState(null)
  const { t } = useTranslation('oss/alert')

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
        title: t('notify.deleteSuccess'),
        color: 'success',
      })
      refreshTable()
    })
  }

  const judgmentType = (type) => {
    switch (type) {
      case 'emailConfigs':
        return t('notify.type.email')
      case 'webhookConfigs':
        return t('notify.type.webhook')
      case 'dingTalkConfigs':
        return t('notify.type.dingtalk')
      case 'wechatConfigs':
        return t('notify.type.wechat')
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
      title: t('notify.alertNotifyName'),
      accessor: 'name',
      justifyContent: 'left',
      customWidth: '20%',
    },
    {
      title: t('notify.notifyType'),
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
      title: t('notify.notifyEmailOrWebhookUrl'),
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
      title: t('notify.operation'),
      accessor: 'action',
      Cell: (props) => {
        const row = props.row.original
        console.log('row', row)
        return (
          <div className="flex">
            <Button
              type="text"
              onClick={() => clickEditRule(row)}
              icon={<MdOutlineEdit className="text-blue-400 hover:text-blue-400" />}
            >
              <span className="text-blue-400 hover:text-blue-400">{t('notify.edit')}</span>
            </Button>
            <Popconfirm
              title={<>{t('notify.confirmDelete', { name: row.name })}</>}
              onConfirm={() => deleteAlertNotify(row)}
              okText={t('notify.confirm')}
              cancelText={t('notify.cancel')}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                {t('notify.delete')}
              </Button>
            </Popconfirm>
          </div>
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
  const handleTableChange = (pageIndex, pageSize) => {
    if (pageSize && pageIndex) {
      setPageSize(pageSize), setPageIndex(pageIndex)
    }
  }
  const refreshTable = () => {
    fetchData()
    setPageIndex(1)
  }
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        total: total,
      },
      loading: false,
    }
  }, [column, data, pageIndex, pageSize])
  return (
    <CustomCard styleType="alerts">
      <LoadingSpinner loading={loading} />
      <div className="flex items-center justify-betweeen text-sm ">
        <Space className="flex-grow">
          <Space className="flex-1">
            <span className="text-nowrap">{t('notify.alertNotifyName')}：</span>
            <Input
              value={searchName}
              onChange={(e) => {
                setSearchName(e.target.value)
                setPageIndex(1)
              }}
              placeholder={t('notify.search')}
            />
          </Space>
        </Space>

        <Button
          type="primary"
          icon={<MdAdd />}
          onClick={clickAddRule}
          className="flex-grow-0 flex-shrink-0"
        >
          <span className="text-xs">{t('notify.addAlertNotify')}</span>
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
    </CustomCard>
  )
}
