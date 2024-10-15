import { Form, Input, Modal } from 'antd'
import React, { useEffect } from 'react'
import LogRouteRuleFormList from './component/LogRouteRuleFormList'
import { addLogRuleApi, getLogRuleApi, updateLogRuleApi } from 'src/api/logs'
import { showToast } from 'src/utils/toast'
import { useLogsContext } from 'src/contexts/LogsContext'

const ConfigLogRuleModal = ({ modalVisible, closeModal, logRuleInfo }) => {
  const { getLogTableInfo, updateLoading } = useLogsContext()
  const [form] = Form.useForm()
  const getLogRule = () => {
    getLogRuleApi({
      dataBase: logRuleInfo.dataBase,
      tableName: logRuleInfo.tableName,
    }).then((res) => {
      form.setFieldsValue({
        parseName: res.parseName,
        parseRule: res.parseRule,
        // routeRule: res.routeRule,
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
  }, [modalVisible, logRuleInfo])

  function addLogRule(logRuleParams) {
    addLogRuleApi(logRuleParams).then((res) => {
      showToast({
        title: '日志规则配置成功',
        color: 'success',
      })
    })
    getLogTableInfo()
    closeModal()
  }
  function updateLogRule(logRuleParams) {
    updateLogRuleApi({
      ...logRuleParams,
      dataBase: logRuleInfo.dataBase,
      tableName: logRuleInfo.tableName,
    }).then((res) => {
      showToast({
        title: '日志规则配置成功',
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
            title: '路由规则不可为空',
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

  return (
    <Modal
      title={'日志规则配置'}
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
      <Form layout={'vertical'} form={form} preserve={false}>
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
          <Input placeholder="规则名" />
        </Form.Item>
        <Form.Item
          label="规则描述"
          name="parseInfo"
          required
          rules={[
            {
              required: true,
              message: '请输入规则描述',
            },
          ]}
        >
          <Input placeholder="规则描述" />
        </Form.Item>
        <LogRouteRuleFormList />

        <Form.Item
          label="解析规则"
          name="parseRule"
          rules={[
            {
              required: true,
              message: '请输入解析规则',
            },
          ]}
        >
          <Input placeholder="解析规则" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
export default ConfigLogRuleModal
