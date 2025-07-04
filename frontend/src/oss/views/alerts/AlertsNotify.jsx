/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Input, Popconfirm, Space } from 'antd'
import React, { useMemo, useState } from 'react'
import { deleteAlertNotifyApi, getAlertmanagerListApi } from 'core/api/alerts'
import LoadingSpinner from 'src/core/components/Spinner'
import BasicTable from 'src/core/components/Table/basicTable'
import { notify } from 'src/core/utils/notify'
import { MdAdd, MdOutlineEdit } from 'react-icons/md'
import ModifyAlertNotifyModal from './modal/ModifyAlertNotifyModal'
import { useTranslation } from 'react-i18next' // 引入i18n
import { RiDeleteBin5Line } from 'react-icons/ri'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import { useSelector } from 'react-redux'
import { useDebounce } from 'react-use'

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
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
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
      notify({
        message: t('notify.deleteSuccess'),
        type: 'success',
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
              icon={
                <MdOutlineEdit className="!text-[var(--ant-color-primary-text)] !hover:text-[var(--ant-color-primary-text-active)]" />
              }
            >
              <span className="text-[var(--ant-color-primary-text)] hover:text-[var(--ant-color-primary-text-active)]">
                {t('notify.edit')}
              </span>
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
  useDebounce(
    () => {
      if (dataGroupId !== null) {
        fetchData()
      }
    },
    300,
    [searchName, dataGroupId],
  )
  const fetchData = () => {
    setLoading(true)

    getAlertmanagerListApi({
      currentPage: pageIndex,
      pageSize: pageSize,
      name: searchName,
      refreshCache: true,
      groupId: dataGroupId,
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
  }, [column, data, pageIndex, pageSize, dataGroupId])
  return (
    <BasicCard>
      <LoadingSpinner loading={loading} />

      <BasicCard.Header>
        <Space className="flex-grow my-2">
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
          className="flex-grow-0 flex-shrink-0 my-2"
        >
          <span className="text-xs">{t('notify.addAlertNotify')}</span>
        </Button>
      </BasicCard.Header>

      <BasicCard.Table>
        <BasicTable {...tableProps} />
      </BasicCard.Table>

      <ModifyAlertNotifyModal
        modalVisible={modalVisible}
        notifyInfo={modalInfo}
        closeModal={() => setModalVisible(false)}
        refresh={refreshTable}
      />
    </BasicCard>
  )
}
