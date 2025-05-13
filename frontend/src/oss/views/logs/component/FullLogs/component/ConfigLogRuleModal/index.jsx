/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Modal, Select } from 'antd'
import React, { useEffect, useState } from 'react'
import LogRouteRuleFormList from './component/LogRouteRuleFormList'
import {
  addLogRuleApi,
  getLogRuleApi,
  getLogRuleServiceRouteRuleApi,
  updateLogRuleApi,
} from 'core/api/logs'
import { notify, notify } from 'src/core/utils/notify'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { getServiceListApi } from 'core/api/service'
import ParseRuleTabs from './component/ParseRuleTabs'
import { useTranslation } from 'react-i18next' // 引入i18n

const ConfigLogRuleModal = ({ modalVisible, closeModal, logRuleInfo }) => {
  const { t } = useTranslation('oss/fullLogs')
  const { getLogTableInfo, updateLoading } = useLogsContext()
  const [form] = Form.useForm()
  const [serviceList, setServiceList] = useState([])
  const getServiceListData = () => {
    // 获取7天前的开始时间（当天00:00:00）
    const startDate = new Date()
    startDate.setDate(new Date().getDate() - 7)
    startDate.setHours(0, 0, 0, 0)

    // 获取现在
    const endDate = new Date()

    // 转换为微秒级别的时间戳
    const startTime = startDate.getTime() * 1000
    const endTime = endDate.getTime() * 1000
    getServiceListApi({ startTime, endTime })
      .then((res) => {
        setServiceList(
          (res ?? []).map((service) => ({
            label: service,
            value: service,
          })),
        )
      })
      .catch((error) => {
        // console.log(error)
        setServiceList([])
      })
  }
  const getLogRule = () => {
    getLogRuleApi({
      dataBase: logRuleInfo.dataBase,
      tableName: logRuleInfo.tableName,
    }).then((res) => {
      const value = {
        parseName: res.parseName,
        parseRule: res.parseRule,
        routeRule: Object.entries(res.routeRule).map(([key, value]) => ({
          key: { key: key, value: key },
          value: value,
        })),
        serviceName: (res.serviceName || []).filter((item) => item !== ''),
        parseInfo: logRuleInfo.parseInfo,
        isStructured: res.isStructured,
        [res.isStructured ? 'structured' : 'unStructured']: res.tableFields?.map((item) => ({
          name: item.name,
          type: {
            key: item.type,
            label: item.type,
            value: item.type,
          },
        })),
      }
      form.setFieldsValue(value)
    })
  }
  useEffect(() => {
    if (logRuleInfo && modalVisible) {
      getLogRule()
    } else {
      form.resetFields()
    }
    if (modalVisible) getServiceListData()
  }, [modalVisible, logRuleInfo])

  function addLogRule(logRuleParams) {
    addLogRuleApi(logRuleParams).then((res) => {
      notify({
        message: t('configLogRuleModal.repositoryConfigSuccessToast'),
        type: 'success',
      })

      getLogTableInfo()
      closeModal()
    })
  }
  function updateLogRule(logRuleParams) {
    updateLogRuleApi({
      ...logRuleParams,
      dataBase: logRuleInfo.dataBase,
      tableName: logRuleInfo.tableName,
    }).then((res) => {
      notify({
        message: t('configLogRuleModal.repositoryConfigSuccessToast'),
        type: 'success',
      })
      closeModal()
      getLogTableInfo()
    })
  }
  function saveLogRule() {
    const recursive = [
      'parseName',
      'routeRule',

      ...(form.getFieldValue('isStructured')
        ? ['structuredRule', 'structured']
        : ['unStructured', 'parseRule']),
      ,
      ,
    ]
    form
      .validateFields(recursive, {
        recursive: true,
      })
      .then(() => {
        const formState = form.getFieldsValue(true)
        const logRuleParams = {
          ...formState,
        }
        //check routeRule
        let routeRule = {}
        formState?.routeRule?.forEach((route) => {
          if (route?.key?.value && route.value) {
            routeRule[route?.key?.value] = route.value
          }
        })
        logRuleParams.routeRule = routeRule
        if (Object.keys(routeRule).length === 0) {
          notify({
            message: t('configLogRuleModal.matchRuleErrorToast'),
            type: 'error',
          })
          return
        }
        // check tableFields
        const tableFields = []
        let tableFieldsList =
          (formState.isStructured ? formState?.structured : formState?.unStructured) || []
        tableFieldsList.forEach((item) => {
          if (item?.type?.value && item.name) {
            tableFields.push({
              name: item.name,
              type: item?.type?.value,
            })
          }
        })
        logRuleParams.tableFields = tableFields

        if (formState.isStructured) {
          delete logRuleParams.parseRule
          if (logRuleParams.tableFields.length === 0) {
            notify({
              message: t('configLogRuleModal.structureErrorToast'),
              type: 'error',
            })
            return
          }
        } else {
          delete logRuleParams.structuredRule
        }

        delete logRuleParams.structured
        delete logRuleParams.unStructured
        updateLoading(true)
        if (logRuleInfo) {
          updateLogRule(logRuleParams)
        } else {
          addLogRule(logRuleParams)
        }
      })
      .catch((error) => console.log(error))
  }
  const getServiceRouteRule = (serviceName) => {
    getLogRuleServiceRouteRuleApi({
      serviceName: serviceName,
    }).then((res) => {
      let result = form.getFieldValue('routeRule') || []

      Object.entries(res?.routeRule)?.forEach(([key, value]) => {
        // 检查当前 result 中是否已经存在该 key
        const existingIndex = result.findIndex((item) => item?.key?.key === key)

        if (existingIndex > -1) {
          // 如果已存在，替换对应的 value
          result[existingIndex].value = value
        } else {
          // 如果不存在，则新增项
          result.push({
            key: {
              key: key,
              value: key,
              label: key,
            },
            value: value,
          })
        }
      })

      // 过滤掉无效项
      result = result.filter((item) => item?.key && item.value)
      // 更新表单值
      form.setFieldValue('routeRule', result)
    })
  }
  return (
    <Modal
      title={t('configLogRuleModal.modalTitle')}
      open={modalVisible}
      onCancel={closeModal}
      destroyOnClose
      centered
      okText={t('configLogRuleModal.saveText')}
      cancelText={t('configLogRuleModal.cancelText')}
      maskClosable={false}
      onOk={saveLogRule}
      width={1000}
      bodyStyle={{ maxHeight: '80vh', overflowY: 'auto', overflowX: 'hidden' }}
    >
      <Form
        layout={'vertical'}
        form={form}
        preserve={false}
        initialValues={{
          routeRule: [{ key: null, value: '' }],
          unStructured: [
            {
              name: '',
              type: {
                key: 'String',
                label: 'String',
                value: 'String',
              },
            },
          ],
          structured: [
            {
              name: '',
              type: {
                key: 'String',
                label: 'String',
                value: 'String',
              },
            },
          ],
        }}
      >
        <Form.Item
          label={t('configLogRuleModal.repositoryNameLabel')}
          name="parseName"
          required
          rules={[
            {
              pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, // 仅允许英文字符
              message: t('configLogRuleModal.repositoryNameMessage1'),
            },
            {
              required: true,
              message: t('configLogRuleModal.repositoryNameMessage2'),
            },
          ]}
        >
          <Input
            placeholder={t('configLogRuleModal.repositoryNamePlaceholder')}
            disabled={logRuleInfo?.parseName}
          />
        </Form.Item>
        <Form.Item label={t('configLogRuleModal.repositoryDescriptionLabel')} name="parseInfo">
          <Input placeholder={t('configLogRuleModal.repositoryDescriptionPlaceholder')} />
        </Form.Item>
        <Form.Item label={t('configLogRuleModal.executiveApplicationLabel')} name="serviceName">
          <Select
            options={serviceList}
            placeholder={t('configLogRuleModal.executiveApplicationPlaceholder')}
            mode="multiple"
            onChange={(value) => getServiceRouteRule(value)}
          ></Select>
        </Form.Item>
        <LogRouteRuleFormList />
        <Form.Item label={t('configLogRuleModal.logConfigLabel')} required name="isStructured">
          <ParseRuleTabs />
        </Form.Item>
      </Form>
    </Modal>
  )
}
export default ConfigLogRuleModal
