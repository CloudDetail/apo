import { Form, Input, Modal, Select, Tooltip } from 'antd'
import React, { useEffect, useState } from 'react'
import LogRouteRuleFormList from './component/LogRouteRuleFormList'
import {
  addLogRuleApi,
  getLogRuleApi,
  getLogRuleServiceRouteRuleApi,
  updateLogRuleApi,
} from 'core/api/logs'
import { showToast } from 'src/core/utils/toast'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { getServiceListApi } from 'core/api/service'
import TextArea from 'antd/es/input/TextArea'
import { AiOutlineInfoCircle } from 'react-icons/ai'

const ConfigLogRuleModal = ({ modalVisible, closeModal, logRuleInfo }) => {
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
      form.setFieldsValue({
        parseName: res.parseName,
        parseRule: res.parseRule,
        routeRule: Object.entries(res.routeRule).map(([key, value]) => ({
          key: { key: key, value: key },
          value: value,
        })),
        serviceName: res.serviceName,
        parseInfo: logRuleInfo.parseInfo,
      })
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
      showToast({
        title: '日志库配置成功',
        color: 'success',
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
      showToast({
        title: '日志库配置成功',
        color: 'success',
      })
      closeModal()
      getLogTableInfo()
    })
  }
  function saveLogRule() {
    form
      .validateFields({})
      .then(() => {
        const formState = form.getFieldsValue(true)
        const logRuleParams = {
          ...formState,
        }
        let routeRule = {}
        formState?.routeRule?.forEach((route) => {
          if (route?.key?.value && route.value) {
            routeRule[route?.key?.value] = route.value
          }
        })
        logRuleParams.routeRule = routeRule
        if (Object.keys(routeRule).length === 0) {
          showToast({
            title: '匹配规则不可为空',
            color: 'danger',
          })
          return
        }
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
      title={'日志库配置'}
      open={modalVisible}
      onCancel={closeModal}
      destroyOnClose
      centered
      okText={'保存'}
      cancelText="取消"
      maskClosable={false}
      onOk={saveLogRule}
      width={1000}
      bodyStyle={{ maxHeight: '80vh', overflowY: 'auto', overflowX: 'hidden' }}
    >
      <Form
        layout={'vertical'}
        form={form}
        preserve={false}
        initialValues={{ routeRule: [{ key: null, value: '' }] }}
      >
        <Form.Item
          label="日志库名"
          name="parseName"
          required
          rules={[
            {
              pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, // 仅允许英文字符
              message: '日志库名只能包含英文字符、数字、下划线',
            },
            {
              required: true,
              message: '请输入日志库名',
            },
          ]}
        >
          <Input
            placeholder="请输入日志库名，一经创建暂不支持修改"
            disabled={logRuleInfo?.parseName}
          />
        </Form.Item>
        <Form.Item label="日志库描述" name="parseInfo">
          <Input placeholder="日志库描述" />
        </Form.Item>
        <Form.Item label="执行应用" name="serviceName">
          <Select
            options={serviceList}
            placeholder="请选择执行规则的应用"
            mode="multiple"
            onChange={(value) => getServiceRouteRule(value)}
          ></Select>
        </Form.Item>
        <LogRouteRuleFormList />

        <Form.Item
          label={
            <div className="flex items-center">
              解析规则 <AiOutlineInfoCircle size={16} className="ml-1" />
              <span className="text-xs text-gray-400">
                将符合规则的日志进行结构化并加快查询速度，查看
                <a
                  href="https://originx.kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/日志解析规则配置/"
                  className="underline"
                  target="_blank"
                >
                  帮助文档
                </a>
              </span>
            </div>
          }
          name="parseRule"
          rules={[
            {
              required: true,
              message: '请输入解析规则',
            },
          ]}
        >
          <TextArea placeholder="解析规则" rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  )
}
export default ConfigLogRuleModal
