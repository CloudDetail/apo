import { Form, Input, Modal, Select, Tooltip } from 'antd'
import React, { useEffect, useState } from 'react'
import LogRouteRuleFormList from './component/LogRouteRuleFormList'
import {
  addLogRuleApi,
  getLogRuleApi,
  getLogRuleServiceRouteRuleApi,
  updateLogRuleApi,
} from 'src/api/logs'
import { showToast } from 'src/utils/toast'
import { useLogsContext } from 'src/contexts/LogsContext'
import { getServiceListApi } from 'src/api/service'
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
        title: '日志解析规则配置成功',
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
        title: '日志解析规则配置成功',
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
        // console.log(formState)
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
        if (result)
          result.push({
            key: {
              key: key,
              value: key,
              label: key,
            },
            value: value,
          })
      })
      result = result.filter((item) => item?.key && item.value)
      form.setFieldValue('routeRule', result)
    })
  }
  return (
    <Modal
      title={'日志解析规则配置'}
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
          label="规则名"
          name="parseName"
          required
          rules={[
            {
              pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, // 仅允许英文字符
              message: '规则名只能包含英文字符、数字、下划线',
            },
            {
              required: true,
              message: '请输入规则名',
            },
          ]}
        >
          <Input
            placeholder="请输入规则名，一经创建暂不支持修改"
            disabled={logRuleInfo?.parseName}
          />
        </Form.Item>
        <Form.Item label="规则描述" name="parseInfo">
          <Input placeholder="规则描述" />
        </Form.Item>
        <Form.Item label="执行应用" name="serviceName">
          <Select
            options={serviceList}
            placeholder="请选择执行规则的应用"
            onChange={(value) => getServiceRouteRule(value)}
          ></Select>
        </Form.Item>
        <LogRouteRuleFormList />

        <Form.Item
          label={
            <div className="flex items-center">
              解析规则 <AiOutlineInfoCircle size={16} className="ml-1" />
              <span className="text-xs text-gray-400">
                用于将日志文本解析为独立的字段，加快检索速度。请使用
                <a href="https://playground.vrl.dev/" className="underline" target="_blank">
                  VRL
                </a>
                语言，查看
                <Tooltip
                  title={
                    <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                      {`.msg, err = parse_regex(.content, r' \\[(?P<level>.*?)\\] \\[(?P<thread>.*?)\\] \\[(?P<method>.*?)\\(.*?\\)\\] - (?P<msg>.*)')
if err == null {
    .content = encode_json(.msg)
}
del(.msg)`}
                    </div>
                  }
                >
                  <span className="px-1 underline cursor-pointer">JAVA</span>
                </Tooltip>
                、
                <Tooltip
                  title={
                    <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                      {`.msg, err = parse_regex(.content, r' \\[(?P<level>.*?)\\] \\[(?P<thread>.*?)\\] \\[(?P<method>.*?)\\(.*?\\)\\] - (?P<msg>.*)')
if err == null {
    .content = encode_json(.msg)
}
del(.msg)`}
                    </div>
                  }
                >
                  <span className="px-1 underline cursor-pointer">Go</span>
                </Tooltip>
                、
                <Tooltip
                  title={
                    <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                      {`.msg, err = parse_regex(.content, r' \\[(?P<level>.*?)\\] \\[(?P<thread>.*?)\\] \\[(?P<method>.*?)\\(.*?\\)\\] - (?P<msg>.*)')
if err == null {
    .content = encode_json(.msg)
}
del(.msg)`}
                    </div>
                  }
                >
                  <span className="px-1 underline cursor-pointer">Nginx</span>
                </Tooltip>
                默认规则
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
