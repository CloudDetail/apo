/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Input, Popconfirm, Select, Space, Tag, theme } from 'antd'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { deleteRuleApi, getAlertRulesApi, getAlertRulesStatusApi } from 'core/api/alerts'
import LoadingSpinner from 'src/core/components/Spinner'
import BasicTable from 'src/core/components/Table/basicTable'
import { notify } from 'src/core/utils/notify'
import { MdAdd, MdOutlineEdit } from 'react-icons/md'
import ModifyAlertRuleModal from './modal/ModifyAlertRuleModal'
import { useSelector } from 'react-redux'
import { useTranslation } from 'react-i18next'
import CustomCard from 'src/core/components/Card/CustomCard'

export default function AlertsRule() {
  const { t } = useTranslation('oss/alert')
  const { useToken } = theme
  const { token } = useToken()
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalInfo, setModalInfo] = useState(null)
  const [alertStateMap, setAlertStateMap] = useState(null)
  const { groupLabelSelectOptions } = useSelector((state) => state.groupLabelReducer)
  const [searchGroup, setSearchGroup] = useState(null)
  const [searchAlert, setSearchAlert] = useState(null)
  const changeSearchGroup = (value) => {
    setSearchGroup(value)
    setPageIndex(1)
  }
  const getStateTagItem = (state) => {
    switch (state) {
      case 'firing':
        return {
          type: 'error',
          context: t('rule.alertStatus.firing'),
        }
      case 'pending':
        return {
          type: 'warning',
          context: t('rule.alertStatus.pending'),
        }
      case 'inactive':
        return {
          type: 'success',
          context: t('rule.alertStatus.inactive'),
        }
      default:
        return {
          type: 'default',
          context: t('rule.alertStatus.unknown'),
        }
    }
  }
  const deleteAlertRule = (rule) => {
    setLoading(true)
    deleteRuleApi({
      group: rule.group,
      alert: rule.alert,
    })
      .then((res) => {
        notify({
          message: t('rule.deleteSuccess'),
          type: 'success',
        })
        refreshTable()
      })
      .catch((error) => {
        setLoading(false)
      })
  }
  const column = [
    {
      title: t('rule.groupName'),
      accessor: 'group',
      customWidth: 120,
      justifyContent: 'left',
    },
    {
      title: t('rule.alertRuleName'),
      accessor: 'alert',
      justifyContent: 'left',
      customWidth: 300,
    },

    {
      title: t('rule.duration'),
      accessor: 'for',
      customWidth: 100,
    },
    {
      title: t('rule.query'),
      accessor: 'expr',
      justifyContent: 'left',
      Cell: ({ value }) => {
        return <span className="text-[var(--ant-color-text)]">{value}</span>
      },
    },

    {
      title: t('rule.alertStatus.title'),
      accessor: 'state',
      customWidth: 150,
      Cell: (props) => {
        const row = props.row.original
        let state
        if (alertStateMap) {
          state = alertStateMap[row.group + '-' + row.alert]
        }
        const tagConfig = getStateTagItem(state)
        return <Tag color={tagConfig.type}>{tagConfig.context}</Tag>
      },
    },
    {
      title: t('rule.operation'),
      accessor: 'action',
      customWidth: 200,
      Cell: (props) => {
        const row = props.row.original
        return (
          <div className="flex">
            <Button
              type="text"
              onClick={() => clickEditRule(row)}
              icon={<MdOutlineEdit className="!text-[var(--ant-color-primary-text)] !hover:text-[var(--ant-color-primary-text-active)]" />}
            >
              <span style={{ color: token.colorPrimary }}>{t('rule.edit')}</span>
            </Button>
            <Popconfirm
              title={<>{t('rule.confirmDelete', { name: row.alert })}</>}
              onConfirm={() => deleteAlertRule(row)}
              okText={t('rule.confirm')}
              cancelText={t('rule.cancel')}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                {t('rule.delete')}
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
  const clickEditRule = (ruleInfo) => {
    setModalInfo(ruleInfo)
    setModalVisible(true)
  }
  const loadRulesData = useCallback(async () => {
    try {
      setLoading(true)
      const res = await getAlertRulesApi({
        currentPage: pageIndex,
        pageSize,
        alert: searchAlert,
        group: searchGroup?.label,
      })
      setData(res.alertRules)
      setTotal(res.pagination.total)
    } catch (err) {
      console.error('error:', err)
    } finally {
      setLoading(false)
    }
  }, [pageIndex, pageSize, searchAlert, searchGroup])

  const loadAlertStates = useCallback(async () => {
    try {
      const res = await getAlertRulesStatusApi({ type: 'alert', exclude_alerts: true })
      const map = {}

      res.data.groups.forEach((group) => {
        group.rules.forEach((rule) => {
          map[`${group.name}-${rule.name}`] = rule.state
        })
      })

      setAlertStateMap(map)
    } catch (err) {
      console.error('error:', err)
    }
  }, [])

  const init = useCallback(async () => {
    setLoading(true)
    await Promise.all([loadRulesData(), loadAlertStates()])
    setLoading(false)
  }, [loadRulesData, loadAlertStates])

  useEffect(() => {
    if (alertStateMap) {
      loadRulesData()
    } else {
      init()
    }
  }, [pageIndex, pageSize, searchAlert, searchGroup])
  async function fetchData() {
    try {
      setLoading(true)
      const [res1, res2] = await Promise.all([
        getAlertRulesApi({
          currentPage: pageIndex,
          pageSize: pageSize,
        }),
        getAlertRulesStatusApi({
          type: 'alert',
          exclude_alerts: true,
        }),
      ])
      setLoading(false)
      setData(res1.alertRules)
      setTotal(res1.pagination.total)
      let alertStateMap = {}
      res2.data.groups.forEach((group) => {
        group.rules.forEach((rule) => {
          // alertStateMap[rule.labels.group + '-' + rule.name] = rule.state
          alertStateMap[group.name + '-' + rule.name] = rule.state
        })
      })
      setAlertStateMap(alertStateMap)
      setLoading(false)
    } catch (error) {
      setLoading(false)
      console.error('error:', error)
    }
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
  }, [column, data, pageIndex, pageSize, searchAlert, searchGroup])
  return (
    <CustomCard styleType="alerts">
      <LoadingSpinner loading={loading} />
      <div className="flex items-center justify-between text-sm ">
        <Space className="flex-grow">
          <Space className="flex-1">
            <span className="text-nowrap">{t('rule.groupName')}：</span>
            <Select
              options={groupLabelSelectOptions}
              labelInValue
              placeholder={t('rule.groupName')}
              // mode="multiple"
              allowClear
              className=" min-w-[200px]"
              value={searchGroup}
              onChange={changeSearchGroup}
            />
          </Space>
          <div className="flex flex-row items-center mr-5 text-sm">
            <span className="text-nowrap">{t('rule.alertRuleName')}：</span>
            <Input
              value={searchAlert}
              placeholder={t('rule.alertRuleName')}
              onChange={(e) => {
                setSearchAlert(e.target.value)
                setPageIndex(1)
              }}
            />
          </div>
        </Space>

        <Button
          type="primary"
          icon={<MdAdd />}
          onClick={clickAddRule}
          className="flex-grow-0 flex-shrink-0"
        >
          <span className="text-xs">{t('rule.addAlertRule')}</span>
        </Button>
      </div>
      <div className="text-sm flex-1 overflow-auto">
        <div className="h-full text-xs justify-between">
          <BasicTable {...tableProps} />
        </div>
      </div>
      <ModifyAlertRuleModal
        modalVisible={modalVisible}
        ruleInfo={modalInfo}
        closeModal={() => setModalVisible(false)}
        refresh={refreshTable}
      />
    </CustomCard>
  )
}
