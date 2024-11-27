import { Form, Input, Modal, Select, Tooltip, Tabs, Divider } from 'antd'
import React, { Children, useEffect, useRef, useState } from 'react'
import { flushSync } from 'react-dom'
import LogRouteRuleFormList from './component/LogRouteRuleFormList'
import LogStructRuleFormList from './component/LogStructRuleFormList'
import { IoMdAddCircleOutline } from 'react-icons/io'
import LoadingSpinner from 'src/core/components/Spinner'
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
  const [currentLogType, setCurrentLogType] = useState(0)
  const [parseRule, setParseRule] = useState("")
  const [loading, setLoading] = useState(false)
  const [jsonRule, setJsonRule] = useState("")
  const subFormRef = useRef(null)

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
        console.log(error)
        setServiceList([])
      })
  }
  const getLogRule = () => {
    getLogRuleApi({
      dataBase: logRuleInfo.dataBase,
      tableName: logRuleInfo.tableName,
    }).then((res) => {
      // 更新 currentLogType
      const newLogType = res.isStructured ? 0 : 1;
      setCurrentLogType(newLogType);
      setTimeout(() => {
        form.setFieldsValue({
          parseName: res.parseName,
          parseRule: res.parseRule,
          routeRule: Object.entries(res.routeRule).map(([key, value]) => ({
            key: { key: key, value: key },
            value: value,
          })),
          serviceName: (res.serviceName.length == 1 && res.serviceName[0] == "") ? null : res.serviceName,
          parseInfo: logRuleInfo.parseInfo,
        });
        setParseRule(res.parseRule);
        subFormRef.current.setStructuringObject(res.tableFields);
      }, 0);
    });
  };
  useEffect(() => {
    if (logRuleInfo && modalVisible) {
      getLogRule()
    } else {
      form.resetFields()
    }
    if (modalVisible) getServiceListData()
  }, [modalVisible, logRuleInfo])

  useEffect(() => {
    setLoading(false)
  }, [])

  function addLogRule(logRuleParams) {
    addLogRuleApi(logRuleParams).then((res) => {
      showToast({
        title: '日志库配置成功',
        color: 'success',
      })
      setLoading(false)
      getLogTableInfo()
      handleModalClose()
    }).catch(() => {
      setLoading(false)
      updateLoading(false)
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
      setLoading(false)
      handleModalClose()
      getLogTableInfo()
    }).catch((error) => {
      setLoading(false)
      updateLoading(false)
      console.log(error)
      showToast({
        title: error,
        color: "danger"
      })
    })
  }
  function saveLogRule() {
    if (loading) return
    setLoading(true)
    Promise.all(
      !currentLogType ? [
        form.validateFields({}),
        subFormRef.current.form.validateFields({})
      ] : [
        form.validateFields({})
      ]
    )
      .then(([formState, subFormState]) => {
        let logRuleParams
        if (!currentLogType) {
          logRuleParams = {
            ...formState,
            parseRule: ""
          }
        } else {
          logRuleParams = {
            ...formState,
          }
        }
        let routeRule = {}
        let tableFields = []
        formState?.routeRule?.forEach((route) => {
          if (route?.key?.value && route.value) {
            routeRule[route?.key?.value] = route.value
          }
        })
        if (Object.keys(routeRule).length === 0) {
          showToast({
            title: '匹配规则不可为空',
            color: 'danger',
          })
          updateLoading(false)
          setLoading(false)
          return
        }
        logRuleParams.routeRule = routeRule
        if (!currentLogType) {
          Object.keys(subFormState).forEach(key => {
            const match = key.match(/^(\w+)_Type$/);
            if (match) {
              const field = match[1];
              tableFields.push({
                type: subFormState[`${field}_Type`],
                name: subFormState[`${field}_Name`]
              });
            }
          });
          const names = new Set();
          for (const obj of tableFields) {
            if (names.has(obj.name)) {
              showToast({
                title: "结构化规则键重复",
                color: "danger"
              })
              updateLoading(false)
              setLoading(false)
              return
            }
            names.add(obj.name);
          }
          logRuleParams.tableFields = tableFields
          if (tableFields.length === 0) {
            showToast({
              title: "日志格式不能为空",
              color: "danger"
            })
            updateLoading(false)
            setLoading(false)
            return
          }
        }
        if (!currentLogType) {
          logRuleParams.isStructured = true
        } else {
          logRuleParams.isStructured = false
        }
        setJsonRule("")
        updateLoading(true)
        if (logRuleInfo) {
          updateLogRule(logRuleParams)
        } else {
          addLogRule(logRuleParams)
        }
      })
      .catch((error) => {
        updateLoading(false)
        setLoading(false)
        console.log(error)
      })
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
      .catch((error) => {
        updateLoading(false)
        setLoading(false)
        showToast({
          title: "错误",
          message: error,
          color: "danger"
        })
      })
  }
  const tabItems = [
    {
      key: 0,
      label: <span className='text-md select-none'>结构化JSON日志</span>,
      children: (
        <div className='min-h-36'>
          <Form.Item
            name="structuredRule"
          >
            <div className='flex mb-2'>
              <AiOutlineInfoCircle size={16} className="ml-1 mr-1" />
              <span className="text-xs text-gray-400">请输入JSON格式的日志样本自动生成日志格式（仅支持解析JSON最外层的键）</span>
            </div>
            <TextArea placeholder="日志样本" rows={3} onChange={(e) => { setJsonRule(e.target.value) }} />
          </Form.Item>
          <div className='flex flex-col items-start w-full'>
            <LogStructRuleFormList jsonRule={jsonRule} fForm={form} ref={subFormRef} />
          </div>
        </div>
      )
    },
    {
      key: 1,
      label: <span className='text-md select-none'>非结构化日志</span>,
      children: (
        <Form.Item
          name="parseRule"
          rules={[
            {
              validator: (_, value) => {
                if (currentLogType) {
                  if (!value) {
                    return Promise.reject(new Error("请输入解析规则"))
                  }
                }
                return Promise.resolve()
              }
            }
          ]}
        >
          <div className='flex mb-2'>
            <AiOutlineInfoCircle size={16} className="ml-1 mr-1" />
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
          <TextArea placeholder="解析规则" rows={3} value={parseRule} onChange={(e) => {
            setParseRule(e.target.value)
            form.setFieldValue('parseRule', e.target.value)
          }} />
        </Form.Item>
      )
    },
  ]
  const handleModalClose = () => {
    if (!loading) {
      closeModal()
      setCurrentLogType(0)
      setParseRule("")
      subFormRef.current.setStructuringObject(null)
      setJsonRule("")
    }
  }
  return (
    <Modal
      title={'日志库配置'}
      open={modalVisible}
      onCancel={handleModalClose}
      destroyOnClose
      centered
      okText={'保存'}

      cancelText="取消"
      maskClosable={false}
      onOk={saveLogRule}
      width={1000}
      bodyStyle={{ maxHeight: '80vh', overflowY: 'auto', overflowX: 'hidden', padding: "20px" }}
    >
      <LoadingSpinner loading={loading} />
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
        {/* <div className="flex items-center">
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
        </div> */}
        <div className='flex'>
          <span className='text-[#DC4446] text-[20px] mr-1'>*</span><p>日志格式配置</p>
        </div>
        <Tabs items={tabItems} activeKey={currentLogType} onChange={(key) => {
          setCurrentLogType(key)
        }} />
      </Form>
    </Modal>
  )
}
export default ConfigLogRuleModal
